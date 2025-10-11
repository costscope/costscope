package azure

import (
	"local/costscope/internal/core/focus/conversion/common"
	"local/costscope/internal/core/focus/types"
)

// CategoryDiscount mirrors the canonical label from the shared common package.
// Keeping a local constant avoids leaking the common package at call sites that
// prefer the azure namespace without changing semantics.
const CategoryDiscount = common.CategoryDiscount

// AzureEnsureDiscount delegates to the shared normalization helper in the common
// package. This preserves identical behavior/metrics/env semantics while allowing
// callers to use the azure namespace today. When/if the implementation moves
// into this package, callers remain unchanged.
func AzureEnsureDiscount(chargeType, billingType string, fr *types.FocusRecord) bool {
	return common.AzureEnsureDiscount(chargeType, billingType, fr)
}
