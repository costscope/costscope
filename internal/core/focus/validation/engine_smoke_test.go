package validation

import (
	"os"
	"testing"
)

// Smoke test for validation engine: run Validate on an empty temp file (schema validator will complain but should return a result)
func TestValidationEngine_Smoke(t *testing.T) {
	eng := NewEngine()
	tmp, err := os.CreateTemp(t.TempDir(), "empty-*.csv")
	if err != nil {
		t.Fatalf("tmp: %v", err)
	}
	_ = tmp.Close()
	res, err := eng.Validate(tmp.Name(), ValidationConfig{Quiet: true})
	if err == nil && res == nil { // either error with nil result or valid result acceptable depending on validators
		t.Fatalf("expected either error or non-nil result")
	}
}
