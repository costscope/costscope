package azure

import (
	"strings"

	"github.com/costscope/costscope/internal/core/focus/types"
)

// Classifier applies provider-aware category classification to a partially-mapped record.
// It mirrors the logic previously embedded in FullRowMapper (initial classify + negative-cost fallback
// and discount normalization cooperation).
type Classifier interface {
	Classify(row []string, fr *types.FocusRecord) // mutates fr.ChargeCategory
}

type classifier struct {
	idx      *HeaderIndex
	classify func(chargeType, billingType string, effectiveCost float64, billedCost *float64, candidateValues []string, provider string) string
}

func newClassifier(idx *HeaderIndex, classify func(chargeType, billingType string, effectiveCost float64, billedCost *float64, candidateValues []string, provider string) string) Classifier {
	if classify == nil {
		classify = classifyChargeCategoryAzure
	}
	return &classifier{idx: idx, classify: classify}
}

func (c *classifier) Classify(row []string, fr *types.FocusRecord) { //nolint:cyclop
	if c == nil || c.idx == nil || fr == nil {
		return
	}
	idx := c.idx
	// Initial provider-aware classification (injected to preserve behavior)
	billedPtr := fr.BilledCost
	fr.ChargeCategory = c.classify(
		FirstNonEmptyValue(rowValue(row, idx.ChargeType), ""),
		FirstNonEmptyValue(rowValue(row, idx.BillingType), ""),
		fr.EffectiveCost,
		billedPtr,
		collectCandidateValuesAzure(row, idx.AmortizedCost, idx.CostInBillingCurrency, idx.Cost, idx.CostInUSD, idx.CostLikeIdx),
		providerAzure,
	)

	// Fallback: Usage but negative cost -> Credit unless discount token present
	if fr.ChargeCategory == types.ChargeCategories.Usage && (fr.EffectiveCost < 0 || (fr.BilledCost != nil && *fr.BilledCost < 0)) {
		ct := strings.ToLower(strings.TrimSpace(FirstNonEmptyValue(rowValue(row, idx.ChargeType), "")))
		bt := strings.ToLower(strings.TrimSpace(FirstNonEmptyValue(rowValue(row, idx.BillingType), "")))
		token := ct
		if token == "" {
			token = bt
		}
		if token != tokenDiscount && token != tokenUsageDisc && !strings.Contains(token, tokenDiscount) {
			fr.ChargeCategory = types.ChargeCategories.Credit
		}
	}
}
