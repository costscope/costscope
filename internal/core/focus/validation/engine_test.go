package validation

import (
	"os"
	"path/filepath"
	"testing"
)

// TestEngineValidate_Basic ensures the validation engine runs end-to-end on a temp file
func TestEngineValidate_Basic(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "sample.parquet")
	if err := os.WriteFile(f, []byte(""), 0600); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	e := NewEngine()
	cfg := ValidationConfig{
		Level:             ValidationLevelStandard,
		Spec:              SpecFOCUS12,
		Format:            "parquet",
		EnableCompliance:  true,
		EnableQuality:     true,
		EnablePerformance: true,
		Quiet:             true,
	}

	res, err := e.Validate(f, cfg)
	if err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}
	if res == nil {
		t.Fatalf("expected non-nil result")
	}
	if res.FilePath != f {
		t.Errorf("expected FilePath %q, got %q", f, res.FilePath)
	}
}

// TestSchemaValidator_SupportsAndValidate covers schema validator paths
func TestSchemaValidator_SupportsAndValidate(t *testing.T) {
	v := NewSchemaValidator(nil)
	for _, format := range []string{"parquet", "csv", "json", "orc", "avro"} {
		if !v.SupportsFormat(format) {
			t.Errorf("expected SupportsFormat(%s)=true", format)
		}
	}
	// Use a file path that triggers the full demo/test column set
	data := filepath.Join("demo", "demo-focus.parquet")
	cfg := ValidationConfig{Spec: SpecFOCUS12}
	out, err := v.Validate(data, cfg)
	if err != nil {
		t.Fatalf("schema Validate error: %v", err)
	}
	if out == nil {
		t.Fatalf("schema Validate returned nil")
	}
}
