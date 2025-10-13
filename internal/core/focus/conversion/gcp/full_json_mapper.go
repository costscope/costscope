package gcp

import (
	"time"

	c "github.com/costscope/costscope/internal/core/focus/conversion/common"
	"github.com/costscope/costscope/internal/core/focus/types"
)

// FullJSONMapper maps a JSON object (map[string]any) to a FocusRecord fields-only.
// Caller is responsible for classification and metrics to preserve root behavior.
type FullJSONMapper struct{}

func NewFullJSONMapper() *FullJSONMapper { return &FullJSONMapper{} }

// Map returns (FocusRecord pointer, hasCredits, isSpot)
func (m *FullJSONMapper) Map(obj map[string]interface{}) (*types.FocusRecord, bool, bool) { //nolint:cyclop
	gs := func(path string) string { return c.GetStringPath(obj, path) }
	gf := func(path string) float64 { return c.GetFloatPath(obj, path) }

	cost := firstFloatNonZero(gf("cost"), gf("pricing.cost"))
	usageQty := firstFloatNonZero(gf("usage.amount"), gf("usage.pricing_quantity"))
	start := c.ParseTimeAny(gs("usage_start_time"))
	end := c.ParseTimeAny(gs("usage_end_time"))

	tags := map[string]string{}
	c.MergeLabels(tags, c.ExtractLabels(obj, "labels"))
	c.MergeLabels(tags, c.ExtractLabels(obj, "system_labels"))
	c.MergeLabels(tags, c.ExtractLabels(obj, "resource.labels"))

	fr := &types.FocusRecord{
		BillingAccountId:   firstNonEmpty(gs("billing_account_id"), gs("billing.account_id")),
		BillingAccountName: gs("billing_account_name"),
		BillingCurrency:    firstNonEmpty(gs("currency"), gs("billing_currency")),
		BillingPeriodStart: start,
		BillingPeriodEnd:   end,
		ChargeCategory:     types.ChargeCategories.Usage,
		ChargeClass:        types.ChargeClasses.OnDemand,
		ChargeDescription:  firstNonEmpty(gs("sku.description"), gs("sku.description_truncated")),
		ChargeFrequency:    "Daily",
		ChargePeriodStart:  start,
		ChargePeriodEnd:    end,
		ChargeSubcategory:  gs("service.description"),
		EffectiveCost:      cost,
		InvoiceIssuerName:  "Google",
		ListCost:           cost,
		ListUnitPrice:      unitPrice(cost, usageQty),
		PricingCategory:    types.PricingCategories.Standard,
		PricingQuantity:    usageQty,
		PricingUnit:        firstNonEmpty(gs("usage.unit"), gs("usage.pricing_unit")),
		ProviderName:       types.ProviderNames.GCP,
		PublisherName:      "Google",
		ResourceId:         firstNonEmpty(gs("resource.name"), gs("resource.global_name")),
		ResourceName:       firstNonEmpty(gs("resource.name"), gs("resource.global_name")),
		ResourceType:       gs("resource.name"),
		ServiceCategory:    gs("service.category"),
		ServiceName:        firstNonEmpty(gs("service.description"), gs("service.id")),
		SkuId:              firstNonEmpty(gs("sku.id"), gs("sku")),
		SkuPriceId:         firstNonEmpty(gs("pricing.price_id"), gs("sku.id")),
		SubAccountId:       firstNonEmpty(gs("project.id"), gs("project.number")),
		SubAccountName:     gs("project.name"),
		UsageQuantity:      usageQty,
		UsageUnit:          firstNonEmpty(gs("usage.unit"), gs("usage.pricing_unit")),
		Tags:               tags,

		SourceProvider:      "gcp",
		SourceFileName:      "",
		ConversionTimestamp: time.Now(),
	}

	if v := gs("location.region"); v != "" {
		fr.Region = &v
	}
	if v := gs("location.zone"); v != "" {
		fr.AvailabilityZone = &v
	}

	hasCredits := false
	isSpot := false
	if v, ok := c.GetPath(obj, "credits"); ok && v != nil {
		id, typ, name, isCredit, spot := c.ParseCreditsUnified(v)
		if id != "" {
			fr.CommitmentDiscountId = &id
		}
		if typ != "" {
			fr.CommitmentDiscountType = &typ
		}
		if name != "" {
			fr.CommitmentDiscountName = &name
		}
		hasCredits = isCredit
		isSpot = spot
		if isSpot {
			fr.PricingCategory = types.PricingCategories.Spot
		}
	}

	// Provider-local classification mirroring root wrapper behavior
	if fr.EffectiveCost < 0 {
		fr.ChargeCategory = types.ChargeCategories.Credit
	}
	if hasCredits {
		fr.ChargeCategory = types.ChargeCategories.Credit
	}

	// Shared unified classification tweaks
	c.ApplyUnifiedClassification("gcp", map[string]string{
		"credits.type": gs("credits.type"),
		"credits.name": gs("credits.name"),
		"credits.id":   gs("credits.id"),
	}, fr)

	return fr, hasCredits, isSpot
}

// helpers provided in helpers.go
