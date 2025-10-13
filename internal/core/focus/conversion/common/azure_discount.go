package common

import (
	"os"
	"strings"

	"github.com/costscope/costscope/internal/core/focus/types"
	"github.com/costscope/costscope/internal/core/logging"
	"github.com/costscope/costscope/internal/core/monitoring/telemetry"
)

// hasDiscountToken performs a zero-allocation, case-insensitive search for the ASCII
// substring "discount" within s. It avoids creating lower-cased copies or builders.
func hasDiscountToken(s string) bool {
	const token = "discount"
	n := len(s)
	if n < len(token) {
		return false
	}
	// Manual case-insensitive scan (ASCII) to avoid allocations from ToLower.
	for i := 0; i <= n-len(token); i++ {
		// Fast first/last char pre-check (d / t) to skip most positions quickly.
		c := s[i]
		if c != 'd' && c != 'D' {
			continue
		}
		// Compare remaining characters case-insensitively.
		match := true
		for j := 1; j < len(token); j++ {
			sc := s[i+j]
			tc := token[j]
			// Uppercase A-Z differs by 32 from lowercase a-z in ASCII; fold if needed.
			if sc == tc || (sc|0x20) == tc { // bitwise lower
				continue
			}
			match = false
			break
		}
		if match {
			return true
		}
	}
	return false
}

// CategoryDiscount is the canonical Discount category label used across Azure normalization.
const CategoryDiscount = "Discount"

// AzureEnsureDiscount applies discount normalization while preserving non-usage categories.
// It returns true if the ChargeCategory was changed to Discount. Optimized to minimize
// allocations (no concatenation + ToLower) while keeping prior semantics.
//
// This helper is shared across legacy and unified mapping paths for Azure and is placed
// in the common package to enable provider subpackage refactors without behavior changes.
//
//nolint:unparam // return bool kept for future callers and symmetry with other normalizers
func AzureEnsureDiscount(chargeType, billingType string, fr *types.FocusRecord) bool {
	if fr == nil {
		return false
	}
	// Diagnostic escape hatch: allow disabling normalization via env for parity / troubleshooting.
	if disable := os.Getenv("COSTSCOPE_DISABLE_AZURE_DISCOUNT_NORMALIZATION"); disable != "" {
		if disable == "1" || strings.EqualFold(disable, "true") || strings.EqualFold(disable, "yes") || strings.EqualFold(disable, "on") || strings.EqualFold(disable, "enable") || strings.EqualFold(disable, "enabled") {
			telemetry.AzureDiscountNormalizationSkips.WithLabelValues("azure").Inc()
			return false
		}
	}
	cat := fr.ChargeCategory
	if cat != "" && cat != types.ChargeCategories.Usage && !hasDiscountToken(cat) {
		return false
	}

	changed := false
	if cat == types.ChargeCategories.Usage && (hasDiscountToken(chargeType) || hasDiscountToken(billingType)) {
		fr.ChargeCategory = CategoryDiscount
		changed = true
		cat = fr.ChargeCategory
	}
	// Canonicalize any existing variant like usage-discount or mixed-case Discount.
	if hasDiscountToken(cat) && cat != CategoryDiscount {
		fr.ChargeCategory = CategoryDiscount
		changed = true
	}
	if changed {
		pathLabel := "legacy"
		if strings.Contains(strings.ToLower(fr.SourceFileName), "unified") { // heuristic path inference
			pathLabel = "unified"
		}
		telemetry.AzureDiscountNormalizations.WithLabelValues("azure", pathLabel).Inc()
		logging.GetLogger().DebugWithFields("azure discount normalization applied", map[string]interface{}{
			"charge_type":    chargeType,
			"billing_type":   billingType,
			"final_category": fr.ChargeCategory,
		})
	}
	return changed
}
