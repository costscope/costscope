package azure

import (
	"strconv"
	"strings"
	"time"

	"github.com/costscope/costscope/internal/core/focus/types"
)

// FieldMapper maps a raw Azure CSV row to a FocusRecord with fields populated,
// but without applying provider/category classification or post-map normalization.
// It preserves legacy semantics used by FullRowMapper prior to this split.
type FieldMapper interface {
	MapFields(row []string) (types.FocusRecord, error)
}

type fieldMapper struct {
	idx           *HeaderIndex
	applyBenefits func(idx *HeaderIndex, row []string, fr *types.FocusRecord)
	parseTags     func(s string) types.Tags
	now           func() time.Time
}

func newFieldMapper(idx *HeaderIndex, applyBenefits func(idx *HeaderIndex, row []string, fr *types.FocusRecord), parseTags func(s string) types.Tags, now func() time.Time) FieldMapper {
	if now == nil {
		now = time.Now
	}
	return &fieldMapper{idx: idx, applyBenefits: applyBenefits, parseTags: parseTags, now: now}
}

func (m *fieldMapper) MapFields(row []string) (types.FocusRecord, error) { //nolint:cyclop
	if m.idx == nil {
		return types.FocusRecord{}, ErrNilHeaderIndex
	}
	idx := m.idx

	quantity := ParseFloat(Get(row, idx.Quantity))

	start, end := func() (time.Time, time.Time) {
		if s := strings.TrimSpace(Get(row, idx.UsageStart)); s != "" {
			st := ParseTimeFlexible(s)
			if e := strings.TrimSpace(Get(row, idx.UsageEnd)); e != "" {
				return st, ParseTimeFlexible(e)
			}
			return st, st.Add(24 * time.Hour)
		}
		d := ParseDateOnlyFlexible(FirstNonEmptyValue(rowValue(row, idx.Date), rowValue(row, idx.UsageDate)))
		if d.IsZero() {
			return time.Time{}, time.Time{}
		}
		return d, d.Add(24 * time.Hour)
	}()

	pricingCategory, chargeClass := func() (string, string) {
		pm := strings.ToLower(strings.TrimSpace(FirstNonEmptyValue(rowValue(row, idx.PricingModel), rowValue(row, idx.PricingModelNm))))
		pc := types.PricingCategories.Standard
		cc := types.ChargeClasses.OnDemand
		switch pm {
		case spotStr:
			pc = types.PricingCategories.Spot
		case kwReservation, kwReserved, kwSavingsPlan1, kwSavingsPlan2, kwCommitment, kwAmortized:
			pc = types.PricingCategories.Reserved
			cc = types.ChargeClasses.Commitment
		}
		return pc, cc
	}()

	effectiveCost := func() float64 {
		if v := FirstNonEmptyValue(rowValue(row, idx.AmortizedCost), rowValue(row, idx.CostInBillingCurrency), rowValue(row, idx.Cost), rowValue(row, idx.CostInUSD)); strings.TrimSpace(v) != "" {
			if n, err := strconv.ParseFloat(strings.TrimSpace(v), 64); err == nil {
				return n
			}
		}
		for _, i := range append([]int{idx.AmortizedCost, idx.CostInBillingCurrency, idx.Cost, idx.CostInUSD}, idx.CostLikeIdx...) {
			if n, err := strconv.ParseFloat(strings.TrimSpace(Get(row, i)), 64); err == nil && n != 0 {
				return n
			}
		}
		return 0
	}()

	billedPtr := func() *float64 {
		s := FirstNonEmptyValue(rowValue(row, idx.CostInBillingCurrency), rowValue(row, idx.Cost))
		if strings.TrimSpace(s) == "" {
			return nil
		}
		b := ParseFloat(s)
		return &b
	}()

	fr := types.FocusRecord{
		BillingAccountId:    FirstNonEmptyValue(rowValue(row, idx.BillingAccountId), rowValue(row, idx.EnrollmentNumber), rowValue(row, idx.BillingProfileId)),
		BillingAccountName:  FirstNonEmptyValue(rowValue(row, idx.BillingAccountName), rowValue(row, idx.BillingProfileName)),
		BillingCurrency:     strings.ToUpper(FirstNonEmptyValue(rowValue(row, idx.BillingCurrency), rowValue(row, idx.Currency))),
		BillingPeriodStart:  start,
		BillingPeriodEnd:    end,
		ChargeCategory:      "", // set by classifier
		ChargeClass:         chargeClass,
		ChargeDescription:   FirstNonEmptyValue(rowValue(row, idx.MeterName), rowValue(row, idx.Product), rowValue(row, idx.ServiceName)),
		ChargeFrequency:     "Daily",
		ChargePeriodStart:   start,
		ChargePeriodEnd:     end,
		ChargeSubcategory:   FirstNonEmptyValue(rowValue(row, idx.MeterSubCat), rowValue(row, idx.ServiceName), rowValue(row, idx.Product)),
		EffectiveCost:       effectiveCost,
		InvoiceIssuerName:   types.ProviderNames.Azure,
		ListCost:            ParseFloat(rowValue(row, idx.RetailPrice)) * quantity,
		ListUnitPrice:       ParseFloat(FirstNonEmptyValue(rowValue(row, idx.RetailPrice), rowValue(row, idx.UnitPrice))),
		PricingCategory:     pricingCategory,
		PricingQuantity:     quantity,
		PricingUnit:         FirstNonEmptyValue(rowValue(row, idx.UnitOfMeasure), rowValue(row, idx.MeterUnit)),
		ProviderName:        types.ProviderNames.Azure,
		PublisherName:       "Microsoft",
		ResourceId:          Get(row, idx.ResourceId),
		ResourceName:        FirstNonEmptyValue(rowValue(row, idx.ResourceName), rowValue(row, idx.InstanceName)),
		ResourceType:        FirstNonEmptyValue(rowValue(row, idx.ResourceType), rowValue(row, idx.ServiceTier)),
		ServiceCategory:     FirstNonEmptyValue(rowValue(row, idx.ServiceFamily), rowValue(row, idx.MeterCategory)),
		ServiceName:         FirstNonEmptyValue(rowValue(row, idx.ServiceName), rowValue(row, idx.Product)),
		SkuId:               FirstNonEmptyValue(rowValue(row, idx.MeterId), rowValue(row, idx.SkuId)),
		SkuPriceId:          FirstNonEmptyValue(rowValue(row, idx.PartNumber), rowValue(row, idx.ProductOrderNo)),
		SubAccountId:        Get(row, idx.SubscriptionId),
		SubAccountName:      Get(row, idx.SubscriptionName),
		UsageQuantity:       quantity,
		UsageUnit:           FirstNonEmptyValue(rowValue(row, idx.UnitOfMeasure), rowValue(row, idx.MeterUnit)),
		AvailabilityZone:    nil,
		BilledCost:          billedPtr,
		Region:              stringPtr(NormalizeRegion(FirstNonEmptyValue(rowValue(row, idx.ResourceLocation), rowValue(row, idx.Location)))),
		SourceProvider:      providerAzure,
		ConversionTimestamp: m.now(),
	}

	// Tags and benefits (pure enrichment, no classification)
	if m.parseTags != nil {
		if tagsStr := strings.TrimSpace(Get(row, idx.Tags)); tagsStr != "" {
			fr.Tags = m.parseTags(tagsStr)
		}
	}
	if m.applyBenefits != nil {
		m.applyBenefits(idx, row, &fr)
	}

	return fr, nil
}
