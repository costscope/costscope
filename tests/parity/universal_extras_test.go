package parity

import "testing"

// Deprecated: this test previously covered removed UniversalConverter helper methods
// (GetSupportedFormats, GetSchema, EstimateConversion, ValidateInput, ConvertBatch).
// With the interface slimmed to only Convert(), these helpers were deleted.
// A focused parity/feature test will be reintroduced separately.
func TestUniversalExtras(t *testing.T) { //nolint:tparallel
	t.Skip("deprecated helpers removed; placeholder test")
}
