package conversion

import (
	c "local/costscope/internal/core/focus/conversion/common"
	"local/costscope/internal/core/focus/types"
	"testing"
)

// This lightweight test keeps coverage around the shared classification path
// and verifies the canonical helper in conversion/common is callable.
func TestCommonClassificationHelperCompiles(t *testing.T) {
	fr := &types.FocusRecord{}
	// Should not panic or error; classification result is provider-specific
	// and opaque here. We only assert the call succeeds.
	c.ApplyUnifiedClassification("azure", map[string]string{"ChargeType": "discount"}, fr)
}
