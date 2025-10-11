package azure

import (
	"local/costscope/internal/core/focus/types"
	"strings"
)

// EnrichUnified fills fields not directly produced by generic field mapping so that
// unified path output matches the legacy fast-path semantics. It is intentionally
// idempotent; existing non-zero/non-empty fields are preserved unless clearly missing.
// csvPath indicates whether we're in the CSV ingestion path (affects some header nuances).
// Note: Generic provider-agnostic enrichment is applied by the root forwarder to avoid
// import cycles; this function performs only Azure-specific enrichment.
func EnrichUnified(rec map[string]string, fr *types.FocusRecord, csvPath bool) {
	if fr == nil {
		return
	}
	jsonParityChargeDescription(rec, fr, csvPath)
	applyDiscountNormalization(rec, fr)
	ensurePeriods(rec, fr)
	fillPricingFromRecord(rec, fr)
	fillDescriptions(rec, fr)
	defaultsIssuerService(rec, fr)
	pricingAndClass(rec, fr)
	categoryFallback(rec, fr)
	resourceAndSku(rec, fr)
	subAccount(rec, fr)
	region(rec, fr)
	billingAccount(rec, fr)
	effectiveAndBilledCost(rec, fr)
	applyBenefits(rec, fr)
	negativeUsageToCredit(fr)
}

// jsonParityChargeDescription enforces JSON path parity rules for ChargeDescription.
func jsonParityChargeDescription(rec map[string]string, fr *types.FocusRecord, csvPath bool) {
	if csvPath || fr == nil {
		return
	}
	if strings.TrimSpace(FirstNonEmptyValue(FirstNonEmptyField(rec, "MeterName"), FirstNonEmptyField(rec, "Product"), FirstNonEmptyField(rec, "ServiceName"))) == "" {
		fr.ChargeDescription = ""
	}
}

// applyDiscountNormalization ensures Discount classification parity with env override/metrics.
func applyDiscountNormalization(rec map[string]string, fr *types.FocusRecord) {
	AzureEnsureDiscount(FirstNonEmptyValue(FirstNonEmptyField(rec, "ChargeType")), FirstNonEmptyValue(FirstNonEmptyField(rec, "BillingType")), fr)
}

// ensurePeriods fills charge/billing periods from record if missing.
func ensurePeriods(rec map[string]string, fr *types.FocusRecord) {
	if fr == nil {
		return
	}
	if fr.ChargePeriodStart.IsZero() || fr.ChargePeriodEnd.IsZero() {
		start, end := DeriveDates(rec)
		if fr.ChargePeriodStart.IsZero() {
			fr.ChargePeriodStart = start
		}
		if fr.ChargePeriodEnd.IsZero() {
			fr.ChargePeriodEnd = end
		}
		if fr.BillingPeriodStart.IsZero() {
			fr.BillingPeriodStart = start
		}
		if fr.BillingPeriodEnd.IsZero() {
			fr.BillingPeriodEnd = end
		}
	}
}

// fillPricingFromRecord populates pricing unit, list unit price, and list cost.
func fillPricingFromRecord(rec map[string]string, fr *types.FocusRecord) {
	if fr == nil {
		return
	}
	if strings.TrimSpace(fr.PricingUnit) == "" {
		fr.PricingUnit = FirstNonEmptyValue(FirstNonEmptyField(rec, "UnitOfMeasure"), FirstNonEmptyField(rec, "MeterUnit"))
	}
	if fr.ListUnitPrice == 0 {
		fr.ListUnitPrice = ParseFloat(FirstNonEmptyValue(FirstNonEmptyField(rec, "RetailPrice"), FirstNonEmptyField(rec, "UnitPrice")))
	}
	if fr.ListCost == 0 && fr.ListUnitPrice != 0 && fr.PricingQuantity != 0 {
		fr.ListCost = fr.ListUnitPrice * fr.PricingQuantity
	}
}

// fillDescriptions sets ChargeDescription and ChargeSubcategory with legacy precedence.
func fillDescriptions(rec map[string]string, fr *types.FocusRecord) {
	if fr == nil {
		return
	}
	if strings.TrimSpace(fr.ChargeDescription) == "" {
		fr.ChargeDescription = FirstNonEmptyValue(FirstNonEmptyField(rec, "MeterName"), FirstNonEmptyField(rec, "Product"), FirstNonEmptyField(rec, "ServiceName"))
	}
	if strings.TrimSpace(fr.ChargeSubcategory) == "" {
		fr.ChargeSubcategory = FirstNonEmptyValue(FirstNonEmptyField(rec, "MeterSubCategory"), FirstNonEmptyField(rec, "ServiceName"), FirstNonEmptyField(rec, "Product"), FirstNonEmptyField(rec, "ServiceInfo2"), FirstNonEmptyField(rec, "Operation"))
	}
}

// defaultsIssuerService sets provider name and service/category defaults.
func defaultsIssuerService(rec map[string]string, fr *types.FocusRecord) {
	if fr == nil {
		return
	}
	if strings.TrimSpace(fr.InvoiceIssuerName) == "" {
		fr.InvoiceIssuerName = types.ProviderNames.Azure
	}
	if strings.TrimSpace(fr.ServiceName) == "" {
		fr.ServiceName = FirstNonEmptyValue(FirstNonEmptyField(rec, "ServiceName"), FirstNonEmptyField(rec, "Product"))
	}
	if strings.TrimSpace(fr.ServiceCategory) == "" {
		fr.ServiceCategory = FirstNonEmptyValue(FirstNonEmptyField(rec, "ServiceFamily"), FirstNonEmptyField(rec, "MeterCategory"))
	}
}

// pricingAndClass derives pricing category and charge class.
func pricingAndClass(rec map[string]string, fr *types.FocusRecord) {
	if fr == nil {
		return
	}
	if strings.TrimSpace(fr.PricingCategory) == "" || strings.TrimSpace(fr.ChargeClass) == "" {
		pc, cc := DerivePricing(rec)
		if strings.TrimSpace(fr.PricingCategory) == "" {
			fr.PricingCategory = pc
		}
		if strings.TrimSpace(fr.ChargeClass) == "" {
			fr.ChargeClass = cc
		}
	}
}

// categoryFallback assigns a category if still empty.
func categoryFallback(rec map[string]string, fr *types.FocusRecord) {
	if fr == nil {
		return
	}
	if strings.TrimSpace(fr.ChargeCategory) == "" {
		fr.ChargeCategory = DeriveChargeCategory(rec, fr.EffectiveCost)
	}
}

// resourceAndSku fills resource, type, and SKU identifiers.
func resourceAndSku(rec map[string]string, fr *types.FocusRecord) {
	if fr == nil {
		return
	}
	if strings.TrimSpace(fr.ResourceName) == "" {
		fr.ResourceName = FirstNonEmptyValue(FirstNonEmptyField(rec, "ResourceName"), FirstNonEmptyField(rec, "InstanceName"))
	}
	if strings.TrimSpace(fr.ResourceType) == "" {
		fr.ResourceType = FirstNonEmptyValue(FirstNonEmptyField(rec, "ResourceType"), FirstNonEmptyField(rec, "ServiceTier"))
	}
	if strings.TrimSpace(fr.SkuId) == "" {
		fr.SkuId = FirstNonEmptyValue(FirstNonEmptyField(rec, "MeterId"), FirstNonEmptyField(rec, "SkuId"))
	}
	if strings.TrimSpace(fr.SkuPriceId) == "" {
		fr.SkuPriceId = FirstNonEmptyValue(FirstNonEmptyField(rec, "PartNumber"), FirstNonEmptyField(rec, "ProductOrderNumber"))
	}
}

// subAccount populates subscription fields.
func subAccount(rec map[string]string, fr *types.FocusRecord) {
	if fr == nil {
		return
	}
	if strings.TrimSpace(fr.SubAccountId) == "" {
		fr.SubAccountId = FirstNonEmptyField(rec, "SubscriptionId")
	}
	if strings.TrimSpace(fr.SubAccountName) == "" {
		fr.SubAccountName = FirstNonEmptyField(rec, "SubscriptionName")
	}
}

// region ensures region pointer is set when present.
func region(rec map[string]string, fr *types.FocusRecord) {
	if fr == nil || fr.Region != nil {
		return
	}
	if reg := NormalizeRegion(FirstNonEmptyValue(FirstNonEmptyField(rec, "ResourceLocation"), FirstNonEmptyField(rec, "Location"))); strings.TrimSpace(reg) != "" {
		r := reg
		fr.Region = &r
	}
}

// billingAccount fills billing account id/name.
func billingAccount(rec map[string]string, fr *types.FocusRecord) {
	if fr == nil {
		return
	}
	if strings.TrimSpace(fr.BillingAccountId) == "" {
		fr.BillingAccountId = FirstNonEmptyValue(FirstNonEmptyField(rec, "BillingAccountId"), FirstNonEmptyField(rec, "EnrollmentNumber"), FirstNonEmptyField(rec, "BillingProfileId"))
	}
	if strings.TrimSpace(fr.BillingAccountName) == "" {
		fr.BillingAccountName = FirstNonEmptyValue(FirstNonEmptyField(rec, "BillingAccountName"), FirstNonEmptyField(rec, "BillingProfileName"))
	}
}

// effectiveAndBilledCost computes effective and billed cost when missing.
func effectiveAndBilledCost(rec map[string]string, fr *types.FocusRecord) {
	if fr == nil {
		return
	}
	if fr.EffectiveCost == 0 {
		fr.EffectiveCost = DeriveEffectiveCost(rec)
	}
	if fr.BilledCost == nil {
		fr.BilledCost = GetBilledCostPtr(rec)
	}
}

// applyBenefits applies commitments/benefits and pricing classification.
func applyBenefits(rec map[string]string, fr *types.FocusRecord) { ApplyBenefitsMap(rec, fr) }

// negativeUsageToCredit converts negative usage to Credit.
func negativeUsageToCredit(fr *types.FocusRecord) {
	if fr == nil {
		return
	}
	if fr.ChargeCategory == types.ChargeCategories.Usage {
		if fr.EffectiveCost < 0 || (fr.BilledCost != nil && *fr.BilledCost < 0) {
			fr.ChargeCategory = types.ChargeCategories.Credit
		}
	}
}

// no local stringPtr to avoid duplicate definitions across provider files
