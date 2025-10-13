package azure

import (
	c "github.com/costscope/costscope/internal/core/focus/conversion/common"
	"github.com/costscope/costscope/internal/core/focus/types"
)

// Normalizer applies provider-specific and shared normalization to a FocusRecord.
// Behavior must remain identical to previous inlined logic.
type Normalizer interface {
	Normalize(row []string, fr *types.FocusRecord)
}

type normalizer struct {
	idx *HeaderIndex
}

func newNormalizer(idx *HeaderIndex) Normalizer { return &normalizer{idx: idx} }

func (n *normalizer) Normalize(row []string, fr *types.FocusRecord) {
	if n == nil || n.idx == nil || fr == nil {
		return
	}
	// Discount normalization using shared helper with metrics/env behavior
	AzureEnsureDiscount(FirstNonEmptyValue(rowValue(row, n.idx.ChargeType), ""), FirstNonEmptyValue(rowValue(row, n.idx.BillingType), ""), fr)

	// Minimal shared normalized fields parity for unified path is handled elsewhere
	// (applyUnifiedNormalizationAzure). Keep this Normalizer focused on provider-specific step.
	_ = c.NormalizeCurrency // reference to keep import if future extensions add usage
}
