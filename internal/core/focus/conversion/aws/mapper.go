package aws

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	c "github.com/costscope/costscope/internal/core/focus/conversion/common"
	"github.com/costscope/costscope/internal/core/focus/types"
	"github.com/costscope/costscope/internal/interfaces"
)

// RowMapper maps CSV rows to FocusRecord.
// Map returns a fully classified & normalized FocusRecord or error.
type RowMapper interface {
	Map(row []string) (types.FocusRecord, error)
}

// rowMapper is the concrete implementation used by ConvertStream.
type rowMapper struct {
	idx         *HeaderIndex
	sourceFile  string
	unified     bool
	fastIndexes RowIndexes
	// meta is reused only in unified mode to avoid per-row map allocations
	meta map[string]string
}

// NewRowMapper constructs a mapper given a header index.
func NewRowMapper(idx *HeaderIndex, sourceFile string, unified bool) RowMapper {
	fx := RowIndexes{
		BillingAccountId:   idx.IBillingAccountId,
		BillingAccountName: idx.IBillingAccountName,
		BillingCurrency:    idx.IBillingCurrency,
		UnblendedCost:      idx.IUnblendedCost,
		UsageAmount:        idx.IUsageAmount,
		UsageStartDate:     idx.IUsageStartDate,
		UsageEndDate:       idx.IUsageEndDate,
		LineItemDesc:       idx.ILineItemDesc,
		Operation:          idx.IOperation,
		UsageType:          idx.IUsageType,
		ProductName:        idx.IProductName,
		ProductFamily:      idx.IProductFamily,
		LineItemType:       idx.ILineItemType,
		PriceId:            idx.IPriceId,
		UsageAccountId:     idx.IUsageAccountId,
		ResourceId:         idx.IResourceId,
		AvailabilityZone:   idx.IAvailabilityZone,
		Region:             idx.IRegion,
		SPArn:              idx.ISPArn,
		SPId:               idx.ISPId,
		RIArn:              idx.IRIArn,
		RISubscriptionId:   idx.IRISubscriptionId,
	}
	rm := &rowMapper{idx: idx, sourceFile: sourceFile, unified: unified, fastIndexes: fx}
	if unified {
		// Preallocate and seed keys to stabilize map hash buckets; values are overwritten per row
		rm.meta = map[string]string{
			"lineItem/LineItemType":      "",
			"lineItem/UsageType":         "",
			"lineItem/Operation":         "",
			"savingsPlan/SavingsPlanId":  "",
			"savingsPlan/SavingsPlanArn": "",
			"reservation/ReservationARN": "",
			"reservation/SubscriptionId": "",
		}
	}
	return rm
}

// Map implements RowMapper.
func (m *rowMapper) Map(row []string) (types.FocusRecord, error) { //nolint:cyclop
	if len(row) < m.idx.Cols {
		return types.FocusRecord{}, fmt.Errorf("short row: have %d need %d", len(row), m.idx.Cols)
	}
	fr, _ := MapRowFast(m.fastIndexes, row)
	at := func(i int) string {
		if i >= 0 && i < len(row) {
			return row[i]
		}
		return ""
	}
	lineItemType := at(m.idx.ILineItemType)
	usageType := at(m.idx.IUsageType)
	operation := at(m.idx.IOperation)
	spId := at(m.idx.ISPId)
	spArn := at(m.idx.ISPArn)
	riArn := at(m.idx.IRIArn)
	riSub := at(m.idx.IRISubscriptionId)
	classifyRecord(&fr, lineItemType, usageType, operation, spId, spArn, riArn, riSub)
	// Region inference from AZ when region empty
	if (fr.Region == nil || *fr.Region == "") && fr.AvailabilityZone != nil {
		az := *fr.AvailabilityZone
		if az != "" && strings.Count(az, "-") >= 2 {
			last := az[len(az)-1]
			if (last >= 'a' && last <= 'z') || (last >= 'A' && last <= 'Z') {
				reg := az[:len(az)-1]
				fr.Region = &reg
			}
		}
	}
	fr.SourceFileName = m.sourceFile
	// Unified vs legacy normalization paths
	if m.unified {
		// Reuse a preallocated map to minimize per-row allocations
		if m.meta != nil {
			m.meta["lineItem/LineItemType"] = lineItemType
			m.meta["lineItem/UsageType"] = usageType
			m.meta["lineItem/Operation"] = operation
			m.meta["savingsPlan/SavingsPlanId"] = spId
			m.meta["savingsPlan/SavingsPlanArn"] = spArn
			m.meta["reservation/ReservationARN"] = riArn
			m.meta["reservation/SubscriptionId"] = riSub
			c.ApplyUnifiedClassification("aws", m.meta, &fr)
		} else {
			// Fallback (shouldn't happen): allocate on the fly
			c.ApplyUnifiedClassification("aws", map[string]string{
				"lineItem/LineItemType":      lineItemType,
				"lineItem/UsageType":         usageType,
				"lineItem/Operation":         operation,
				"savingsPlan/SavingsPlanId":  spId,
				"savingsPlan/SavingsPlanArn": spArn,
				"reservation/ReservationARN": riArn,
				"reservation/SubscriptionId": riSub,
			}, &fr)
		}
		c.ApplyUnifiedNormalization("aws", &fr)
	} else {
		c.NormalizeFocusRecord(&fr)
	}
	return fr, nil
}

// Compile-time assertion: rowMapper implements the provider-agnostic interface.
var _ interfaces.ProviderRowMapper = (*rowMapper)(nil)

// RowIndexes holds precomputed positions used by mappers (subset of HeaderIndex)
type RowIndexes struct {
	BillingAccountId   int
	BillingAccountName int
	BillingCurrency    int
	UnblendedCost      int
	UsageAmount        int
	UsageStartDate     int
	UsageEndDate       int
	LineItemDesc       int
	Operation          int
	UsageType          int
	ProductName        int
	ProductFamily      int
	LineItemType       int
	PriceId            int
	UsageAccountId     int
	ResourceId         int
	AvailabilityZone   int
	Region             int
	// Commitment related
	SPArn            int
	SPId             int
	RIArn            int
	RISubscriptionId int
}

// MapRowFast maps a normalized CUR row slice into a FocusRecord using prebuilt indexes.
func MapRowFast(idx RowIndexes, r []string) (types.FocusRecord, error) { //nolint:cyclop
	get := func(i int) string {
		if i >= 0 && i < len(r) {
			return r[i]
		}
		return ""
	}
	unblendedCost, _ := strconv.ParseFloat(get(idx.UnblendedCost), 64)
	usageAmount, _ := strconv.ParseFloat(get(idx.UsageAmount), 64)
	usageStartDate, _ := time.Parse("2006-01-02 15:04:05", get(idx.UsageStartDate))
	usageEndDate, _ := time.Parse("2006-01-02 15:04:05", get(idx.UsageEndDate))

	var azPtr *string
	if idx.AvailabilityZone >= 0 {
		az := get(idx.AvailabilityZone)
		if az != "" {
			azPtr = &az
		}
	}
	var regionPtr *string
	if idx.Region >= 0 {
		rg := get(idx.Region)
		if rg != "" {
			regionPtr = &rg
		}
	}

	fr := types.FocusRecord{
		BillingAccountId:   get(idx.BillingAccountId),
		BillingAccountName: get(idx.BillingAccountName),
		BillingCurrency:    get(idx.BillingCurrency),
		BillingPeriodStart: usageStartDate,
		BillingPeriodEnd:   usageEndDate,
		ChargeCategory:     types.ChargeCategories.Usage,
		ChargeClass:        types.ChargeClasses.OnDemand,
		ChargeDescription:  get(idx.LineItemDesc),
		ChargeFrequency:    "Daily",
		ChargePeriodStart:  usageStartDate,
		ChargePeriodEnd:    usageEndDate,
		ChargeSubcategory:  get(idx.Operation),
		EffectiveCost:      unblendedCost,
		InvoiceIssuerName:  types.ProviderNames.AWS,
		ListCost:           unblendedCost,
		ListUnitPrice: func() float64 {
			if usageAmount != 0 {
				return unblendedCost / usageAmount
			}
			return 0
		}(),
		PricingCategory:  types.PricingCategories.Standard,
		PricingQuantity:  usageAmount,
		PricingUnit:      get(idx.UsageType),
		ProviderName:     types.ProviderNames.AWS,
		PublisherName:    types.ProviderNames.AWS,
		ResourceId:       get(idx.ResourceId),
		ResourceName:     get(idx.ResourceId),
		ResourceType:     get(idx.ProductName),
		ServiceCategory:  get(idx.ProductFamily),
		ServiceName:      get(idx.ProductName),
		SkuId:            get(idx.LineItemType),
		SkuPriceId:       get(idx.PriceId),
		SubAccountId:     get(idx.UsageAccountId),
		SubAccountName:   get(idx.UsageAccountId),
		UsageQuantity:    usageAmount,
		UsageUnit:        get(idx.UsageType),
		AvailabilityZone: azPtr,
		Region:           regionPtr,
		SourceProvider:   "aws",
	}
	return fr, nil
}

// classifyRecord applies CUR-based classification and commitment enrichment to a FocusRecord.
// mini-helpers to enrich commitment metadata consistently
func enrichSavingsPlan(fr *types.FocusRecord, spId, spArn string) {
	if spId == "" && spArn == "" {
		return
	}
	ctype := ctypeSavingsPlan
	fr.CommitmentDiscountType = &ctype
	if id := firstNonEmpty(spId, spArn); id != "" {
		fr.CommitmentDiscountId = &id
	}
	if name := firstNonEmpty(spId, spArn); name != "" {
		fr.CommitmentDiscountName = &name
	}
}

func enrichReservedInstance(fr *types.FocusRecord, riArn, riSub string) {
	if riArn == "" && riSub == "" {
		return
	}
	ctype := ctypeReservedInstance
	fr.CommitmentDiscountType = &ctype
	if id := firstNonEmpty(riSub, riArn); id != "" {
		fr.CommitmentDiscountId = &id
	}
	if name := firstNonEmpty(riArn, riSub); name != "" {
		fr.CommitmentDiscountName = &name
	}
}

type classCtx struct {
	lt, ut, op   string
	spId, spArn  string
	riArn, riSub string
}

// registry of table-driven lineItemType handlers
var lineItemHandlers = map[string]func(fr *types.FocusRecord, ctx classCtx){
	"SavingsPlanCoveredUsage": func(fr *types.FocusRecord, ctx classCtx) {
		fr.ChargeClass = types.ChargeClasses.Commitment
		fr.PricingCategory = types.PricingCategories.Standard
		enrichSavingsPlan(fr, ctx.spId, ctx.spArn)
	},
	"SavingsPlanRecurringFee": func(fr *types.FocusRecord, ctx classCtx) {
		fr.ChargeCategory = types.ChargeCategories.Purchase
		fr.ChargeClass = types.ChargeClasses.Commitment
		enrichSavingsPlan(fr, ctx.spId, ctx.spArn)
	},
	"SavingsPlanNegation": func(fr *types.FocusRecord, ctx classCtx) {
		fr.ChargeCategory = types.ChargeCategories.Adjustment
		fr.ChargeClass = types.ChargeClasses.Commitment
		enrichSavingsPlan(fr, ctx.spId, ctx.spArn)
	},
	"DiscountedUsage": func(fr *types.FocusRecord, ctx classCtx) {
		fr.ChargeClass = types.ChargeClasses.Commitment
		fr.PricingCategory = types.PricingCategories.Reserved
		enrichReservedInstance(fr, ctx.riArn, ctx.riSub)
	},
	"RIFee": func(fr *types.FocusRecord, ctx classCtx) {
		fr.ChargeCategory = types.ChargeCategories.Purchase
		fr.ChargeClass = types.ChargeClasses.Commitment
		enrichReservedInstance(fr, ctx.riArn, ctx.riSub)
	},
	"Tax": func(fr *types.FocusRecord, _ classCtx) {
		fr.ChargeCategory = types.ChargeCategories.Tax
	},
	"Refund": func(fr *types.FocusRecord, _ classCtx) {
		fr.ChargeCategory = types.ChargeCategories.Adjustment
		fr.ChargeClass = types.ChargeClasses.Correction
	},
	curOpCredit: func(fr *types.FocusRecord, _ classCtx) {
		fr.ChargeCategory = types.ChargeCategories.Credit
	},
	"Fee": func(fr *types.FocusRecord, _ classCtx) {
		fr.ChargeCategory = types.ChargeCategories.Adjustment
	},
}

func classifyRecord(fr *types.FocusRecord, lineItemType, usageType, operation, spId, spArn, riArn, riSub string) { //nolint:cyclop,gocyclo
	lt := strings.TrimSpace(lineItemType)
	ut := strings.ToLower(strings.TrimSpace(usageType))
	op := strings.ToLower(strings.TrimSpace(operation))

	// Unified enrichment hook (preserve metrics/labels behavior)
	c.ApplyUnifiedClassification("aws", map[string]string{
		"lineItem/LineItemType":      lineItemType,
		"lineItem/UsageType":         usageType,
		"lineItem/Operation":         operation,
		"savingsPlan/SavingsPlanId":  spId,
		"savingsPlan/SavingsPlanArn": spArn,
		"reservation/ReservationARN": riArn,
		"reservation/SubscriptionId": riSub,
	}, fr)

	if handler, ok := lineItemHandlers[lt]; ok {
		handler(fr, classCtx{lt: lt, ut: ut, op: op, spId: spId, spArn: spArn, riArn: riArn, riSub: riSub})
		return
	}

	// default path: classify Spot via usageType/operation tokens
	if strings.Contains(ut, "spot") || strings.Contains(op, "spot") {
		fr.PricingCategory = types.PricingCategories.Spot
	}
}

// firstNonEmpty is deprecated; use common.FirstNonEmpty
func firstNonEmpty(ss ...string) string { return c.FirstNonEmpty(ss...) }
