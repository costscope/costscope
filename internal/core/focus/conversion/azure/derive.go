package azure

import (
	"strconv"
	"strings"
	"time"

	"github.com/costscope/costscope/internal/core/focus/types"
)

// DerivePricing determines pricing category and charge class from record fields.
func DerivePricing(rec map[string]string) (string, string) {
	pricingModel := strings.ToLower(strings.TrimSpace(FirstNonEmptyValue(FirstNonEmptyField(rec, "PricingModel"), FirstNonEmptyField(rec, "PricingModelName"))))
	pricingCategory := types.PricingCategories.Standard
	chargeClass := types.ChargeClasses.OnDemand
	switch pricingModel {
	case "spot":
		pricingCategory = types.PricingCategories.Spot
	case "reservation", "reserved", "savingsplan", "savings plan", "commitment", "amortized":
		pricingCategory = types.PricingCategories.Reserved
		chargeClass = types.ChargeClasses.Commitment
	}
	return pricingCategory, chargeClass
}

// DeriveDates returns charge period start/end with fallbacks to date-only.
func DeriveDates(rec map[string]string) (time.Time, time.Time) {
	var start, end time.Time
	if s := strings.TrimSpace(FirstNonEmptyField(rec, "UsageStart")); s != "" {
		start = ParseTimeFlexible(s)
		if e := strings.TrimSpace(FirstNonEmptyField(rec, "UsageEnd")); e != "" {
			end = ParseTimeFlexible(e)
		}
	}
	if start.IsZero() {
		d := ParseDateOnlyFlexible(FirstNonEmptyValue(FirstNonEmptyField(rec, "Date"), FirstNonEmptyField(rec, "UsageDate")))
		start = d
		end = d.Add(24 * time.Hour)
	} else if end.IsZero() {
		end = start.Add(24 * time.Hour)
	}
	return start, end
}

// DeriveEffectiveCost computes effective cost with broad fallbacks.
func DeriveEffectiveCost(rec map[string]string) float64 {
	cost := ParseFloat(FirstNonEmptyField(rec, "AmortizedCost", "CostInBillingCurrency", "Cost", "CostInUSD"))
	if cost == 0 {
		if bs := FirstNonEmptyValue(FirstNonEmptyField(rec, "CostInBillingCurrency"), FirstNonEmptyField(rec, "Cost")); strings.TrimSpace(bs) != "" {
			if n, err := strconv.ParseFloat(strings.TrimSpace(bs), 64); err == nil {
				cost = n
			}
		}
	}
	if cost == 0 {
		for k, v := range rec {
			if strings.Contains(strings.ToLower(k), "cost") {
				if n, err := strconv.ParseFloat(strings.TrimSpace(v), 64); err == nil {
					cost = n
					break
				}
			}
		}
	}
	return cost
}

// GetBilledCostPtr returns billed cost pointer if present, nil otherwise.
func GetBilledCostPtr(rec map[string]string) *float64 {
	billedStr := FirstNonEmptyValue(FirstNonEmptyField(rec, "CostInBillingCurrency"), FirstNonEmptyField(rec, "Cost"))
	if strings.TrimSpace(billedStr) == "" {
		return nil
	}
	b := ParseFloat(billedStr)
	return &b
}

// DeriveChargeCategory decides charge category from explicit types/negatives.
func DeriveChargeCategory(rec map[string]string, cost float64) string { //nolint:cyclop
	if ct := strings.ToLower(strings.TrimSpace(FirstNonEmptyValue(FirstNonEmptyField(rec, "ChargeType"), FirstNonEmptyField(rec, "BillingType")))); ct != "" {
		switch ct {
		case "usage":
			return types.ChargeCategories.Usage
		case "purchase", "buy", "reservation", "savingsplan":
			return types.ChargeCategories.Purchase
		case "credit", "refund":
			return types.ChargeCategories.Credit
		case "tax":
			return types.ChargeCategories.Tax
		case "adjustment":
			return types.ChargeCategories.Adjustment
		case "discount", "usage-discount":
			return CategoryDiscount
		}
	}
	billed := ParseFloat(FirstNonEmptyValue(FirstNonEmptyField(rec, "CostInBillingCurrency"), FirstNonEmptyField(rec, "Cost")))
	if cost < 0 || billed < 0 {
		return types.ChargeCategories.Credit
	}
	rawCosts := []string{FirstNonEmptyField(rec, "AmortizedCost"), FirstNonEmptyField(rec, "CostInBillingCurrency"), FirstNonEmptyField(rec, "Cost"), FirstNonEmptyField(rec, "CostInUSD")}
	for _, rc := range rawCosts {
		if s := strings.TrimSpace(rc); s != "" && (strings.HasPrefix(s, "-") || (strings.HasPrefix(s, "(") && strings.HasSuffix(s, ")"))) {
			return types.ChargeCategories.Credit
		}
	}
	for k, v := range rec {
		if strings.Contains(strings.ToLower(k), "cost") {
			s := strings.TrimSpace(v)
			if s != "" && (strings.HasPrefix(s, "-") || (strings.HasPrefix(s, "(") && strings.HasSuffix(s, ")"))) {
				return types.ChargeCategories.Credit
			}
		}
	}
	for _, v := range rec {
		s := strings.TrimSpace(v)
		if s != "" && strings.HasPrefix(s, "-") {
			return types.ChargeCategories.Credit
		}
	}
	return types.ChargeCategories.Usage
}
