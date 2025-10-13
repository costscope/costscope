package gcp

import (
	"strings"
	"time"

	c "github.com/costscope/costscope/internal/core/focus/conversion/common"
	"github.com/costscope/costscope/internal/core/focus/types"
)

// ApplyUnifiedPostMapGCP finalizes a FocusRecord using fields from a JSON object map.
// Mirrors root behavior; exported so root shims can delegate here.
func ApplyUnifiedPostMapGCP(fr *types.FocusRecord, obj map[string]interface{}) {
	if fr == nil {
		return
	}
	if strings.TrimSpace(fr.ChargeCategory) == "" {
		fr.ChargeCategory = types.ChargeCategories.Usage
	}
	if strings.TrimSpace(fr.PricingCategory) == "" {
		fr.PricingCategory = types.PricingCategories.Standard
	}
	if strings.TrimSpace(fr.ChargeClass) == "" {
		fr.ChargeClass = types.ChargeClasses.OnDemand
	}
	if v, ok := c.GetPath(obj, "currency"); ok {
		if s, ok := v.(string); ok && s != "" {
			fr.BillingCurrency = strings.ToUpper(strings.TrimSpace(s))
		}
	}
	if v := c.GetStringPath(obj, "location.region"); v != "" {
		vv := strings.ToLower(strings.TrimSpace(v))
		fr.Region = &vv
	}
	if v := c.GetStringPath(obj, "location.zone"); v != "" {
		vv := strings.ToLower(strings.TrimSpace(v))
		fr.AvailabilityZone = &vv
	}
	if v, ok := c.GetPath(obj, "credits"); ok {
		id, typ, name, isCredit, isSpot := c.ParseCreditsUnified(v)
		if isCredit {
			fr.ChargeCategory = types.ChargeCategories.Credit
		}
		if id != "" {
			fr.CommitmentDiscountId = &id
		}
		if typ != "" {
			fr.CommitmentDiscountType = &typ
		}
		if name != "" {
			fr.CommitmentDiscountName = &name
		}
		if isSpot {
			fr.PricingCategory = types.PricingCategories.Spot
		}
	}
	if cost := firstFloatNonZero(c.GetFloatPath(obj, "cost"), c.GetFloatPath(obj, "pricing.cost")); cost < 0 {
		fr.ChargeCategory = types.ChargeCategories.Credit
	}
}

// ApplyUnifiedPostMapGCPRow applies the same logic as ApplyUnifiedPostMapGCP, but for CSV row maps.
// Removed ApplyUnifiedPostMapGCPRow: CSV path builds obj map and reuses ApplyUnifiedPostMapGCP.

// EnsureUsageUnit sets UsageUnit based on raw row hints or PricingUnit when empty.
func EnsureUsageUnit(fr *types.FocusRecord, obj map[string]interface{}) {
	if fr == nil || fr.UsageUnit != "" {
		return
	}
	if v, ok := obj["usage.unit"]; ok {
		if s, sok := v.(string); sok && s != "" {
			fr.UsageUnit = c.CanonicalUnit(s)
		}
	}
	if fr.UsageUnit == "" {
		if v, ok := obj["usage.pricing_unit"]; ok {
			if s, sok := v.(string); sok && s != "" {
				fr.UsageUnit = c.CanonicalUnit(s)
			}
		}
	}
	if fr.UsageUnit == "" && fr.PricingUnit != "" {
		fr.UsageUnit = fr.PricingUnit
	}
}

// EnrichUnified fills in derived fields the minimal mapping rules omit so unified output matches legacy semantics.
// Mirrors the former root enrichGCPUnified behavior without changing semantics or metrics.
func EnrichUnified(fr *types.FocusRecord, raw map[string]interface{}) {
	if fr == nil || raw == nil {
		return
	}
	enrichCurrency(fr, raw)
	enrichIssuerPublisher(fr)
	enrichPeriods(fr, raw)
	enrichUnitsAndPricing(fr, raw)
	enrichTextFields(fr, raw)
	enrichProjectResource(fr, raw)
	setClassificationDefaults(fr)
	overrideCreditIfNegative(fr)
	ensureConversionTimestamp(fr)
}

// --- helpers split from EnrichUnified to reduce cyclomatic complexity (no behavior change) ---

func enrichCurrency(fr *types.FocusRecord, raw map[string]interface{}) {
	if fr.BillingCurrency != "" {
		return
	}
	if v, ok := raw["currency"]; ok {
		if s, ok2 := v.(string); ok2 {
			fr.BillingCurrency = strings.ToUpper(strings.TrimSpace(s))
		}
	}
}

func enrichIssuerPublisher(fr *types.FocusRecord) {
	if fr.InvoiceIssuerName == "" {
		fr.InvoiceIssuerName = "Google"
	}
	if fr.PublisherName == "" {
		fr.PublisherName = "Google"
	}
}

func enrichPeriods(fr *types.FocusRecord, raw map[string]interface{}) {
	if fr.ChargePeriodStart.IsZero() || fr.BillingPeriodStart.IsZero() {
		if v, ok := raw["usage_start_time"]; ok {
			if s, ok2 := v.(string); ok2 && s != "" {
				if t := c.ParseTimeAny(s); !t.IsZero() {
					fr.ChargePeriodStart = t
					if fr.BillingPeriodStart.IsZero() {
						fr.BillingPeriodStart = t
					}
				}
			}
		}
	}
	if fr.ChargePeriodEnd.IsZero() || fr.BillingPeriodEnd.IsZero() {
		if v, ok := raw["usage_end_time"]; ok {
			if s, ok2 := v.(string); ok2 && s != "" {
				if t := c.ParseTimeAny(s); !t.IsZero() {
					fr.ChargePeriodEnd = t
					if fr.BillingPeriodEnd.IsZero() {
						fr.BillingPeriodEnd = t
					}
				}
			}
		}
	}
}

func enrichUnitsAndPricing(fr *types.FocusRecord, raw map[string]interface{}) {
	var unitCand string
	if v, ok := raw["usage.unit"]; ok {
		if s, ok2 := v.(string); ok2 {
			unitCand = s
		}
	}
	if strings.TrimSpace(fr.ChargeFrequency) == "" {
		fr.ChargeFrequency = "Daily"
	}
	if fr.PricingQuantity == 0 && fr.UsageQuantity != 0 {
		fr.PricingQuantity = fr.UsageQuantity
	}
	if fr.UsageUnit == "" && unitCand != "" {
		fr.UsageUnit = strings.TrimSpace(unitCand)
	}
	if fr.UsageUnit != "" {
		fr.UsageUnit = strings.ToLower(fr.UsageUnit)
	}
	if fr.PricingUnit == "" {
		fr.PricingUnit = fr.UsageUnit
	}
	if fr.PricingUnit != "" {
		fr.PricingUnit = strings.ToLower(fr.PricingUnit)
	}
	if fr.ListCost == 0 {
		fr.ListCost = fr.EffectiveCost
	}
	if fr.ListUnitPrice == 0 && fr.UsageQuantity != 0 {
		fr.ListUnitPrice = fr.EffectiveCost / fr.UsageQuantity
	}
}

func enrichTextFields(fr *types.FocusRecord, raw map[string]interface{}) {
	if fr.ChargeSubcategory == "" {
		if v, ok := raw["service.description"]; ok {
			if s, ok2 := v.(string); ok2 {
				fr.ChargeSubcategory = s
			}
		}
	}
	if fr.ServiceName == "" {
		if v, ok := raw["service.description"]; ok {
			if s, ok2 := v.(string); ok2 {
				fr.ServiceName = s
			}
		}
	}
	if fr.ChargeDescription == "" {
		if v, ok := raw["sku.description"]; ok {
			if s, ok2 := v.(string); ok2 {
				fr.ChargeDescription = s
			}
		}
	}
	if fr.SkuId == "" {
		if v, ok := raw["sku.id"]; ok {
			if s, ok2 := v.(string); ok2 {
				fr.SkuId = s
			}
		}
	}
	if fr.SkuPriceId == "" {
		if v, ok := raw["pricing.price_id"]; ok {
			if s, ok2 := v.(string); ok2 {
				fr.SkuPriceId = s
			}
		} else if fr.SkuId != "" {
			fr.SkuPriceId = fr.SkuId
		}
	}
}

func enrichProjectResource(fr *types.FocusRecord, raw map[string]interface{}) {
	if fr.SubAccountId == "" {
		if v, ok := raw["project.id"]; ok {
			if s, ok2 := v.(string); ok2 {
				fr.SubAccountId = s
			}
		} else if v, ok := raw["project.number"]; ok {
			if s, ok2 := v.(string); ok2 {
				fr.SubAccountId = s
			}
		}
	}
	if fr.SubAccountName == "" {
		if v, ok := raw["project.name"]; ok {
			if s, ok2 := v.(string); ok2 {
				fr.SubAccountName = s
			}
		}
	}
	if fr.ResourceId == "" {
		if v, ok := raw["resource.name"]; ok {
			if s, ok2 := v.(string); ok2 {
				fr.ResourceId = s
			}
		}
	}
	if fr.ResourceName == "" {
		fr.ResourceName = fr.ResourceId
	}
}

func setClassificationDefaults(fr *types.FocusRecord) {
	if fr.ChargeCategory == "" {
		fr.ChargeCategory = types.ChargeCategories.Usage
	}
	if fr.ChargeClass == "" {
		fr.ChargeClass = types.ChargeClasses.OnDemand
	}
	if fr.PricingCategory == "" {
		fr.PricingCategory = types.PricingCategories.Standard
	}
}

func overrideCreditIfNegative(fr *types.FocusRecord) {
	if fr.EffectiveCost < 0 || (fr.BilledCost != nil && *fr.BilledCost < 0) {
		fr.ChargeCategory = types.ChargeCategories.Credit
	}
}

func ensureConversionTimestamp(fr *types.FocusRecord) {
	if fr.ConversionTimestamp.IsZero() {
		fr.ConversionTimestamp = time.Now()
	}
}
