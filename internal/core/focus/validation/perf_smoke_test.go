package validation

import (
	"os"
	"path/filepath"
	"testing"
)

// Smoke test: ensure performance validator participates and returns a sane score
func TestEngine_PerformanceSmoke(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "perf.parquet")
	if err := os.WriteFile(f, []byte(""), 0o600); err != nil {
		t.Fatalf("write temp: %v", err)
	}

	e := NewEngine()
	cfg := ValidationConfig{
		Spec:              SpecFOCUS12,
		EnablePerformance: true,
		Quiet:             true,
	}
	res, err := e.Validate(f, cfg)
	if err != nil {
		t.Fatalf("validate err: %v", err)
	}
	if res.PerformanceValidation.Score <= 0 {
		t.Fatalf("expected performance score > 0, got %.2f", res.PerformanceValidation.Score)
	}
}
