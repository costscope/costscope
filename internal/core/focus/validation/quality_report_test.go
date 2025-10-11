package validation

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// Ensures manual JSON export of validation result works (ExportJSON removed)
func TestEngine_JSONReportManual(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "sample.parquet")
	if err := os.WriteFile(f, []byte(""), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	out := filepath.Join(dir, "report.json")

	e := NewEngine()
	cfg := ValidationConfig{Spec: SpecFOCUS12, EnableQuality: true, EnableCompliance: true, Quiet: true}
	res, err := e.Validate(f, cfg)
	if err != nil {
		t.Fatalf("validate: %v", err)
	}
	data, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := os.WriteFile(out, data, 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	// #nosec G304 - reading file we just wrote
	b, err := os.ReadFile(out)
	if err != nil || len(b) == 0 {
		t.Fatalf("expected non-empty report: %v", err)
	}
}
