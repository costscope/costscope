package validation

import (
	"strings"
	"testing"
)

// Verify that runSchemaValidation returns a clear error when the schema validator is not registered
func Test_runSchemaValidation_missingSchemaValidator(t *testing.T) {
	e := NewEngine()
	// ensure schema validator is absent
	delete(e.validators, "schema")

	res := &ValidationResult{}
	err := e.runSchemaValidation("dummy.file", ValidationConfig{Quiet: true}, res)
	if err == nil {
		t.Fatalf("expected error when schema validator is missing, got nil")
	}
	if !strings.Contains(err.Error(), "schema validator not found") {
		t.Fatalf("unexpected error message: %v", err)
	}
}
