package commands

import (
	"os"
	"testing"
)

// TestAnalyzeFocusEngineFlag ensures the --use-focus-engine path executes without error.
func TestAnalyzeFocusEngineFlag(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "focus-analyze-*.parquet")
	if err != nil {
		t.Fatalf("temp file: %v", err)
	}
	_ = f.Close()

	analysisUseFocusEngine = true
	analysisOutputFile = "" // stdout path
	if err := runAnalyze(f.Name()); err != nil {
		t.Fatalf("runAnalyze focus engine path failed: %v", err)
	}
}

// TestDiffFocusEngineFlag ensures the --use-focus-engine path for diff executes.
func TestDiffFocusEngineFlag(t *testing.T) {
	b, err := os.CreateTemp(t.TempDir(), "baseline-*.parquet")
	if err != nil {
		t.Fatalf("temp baseline: %v", err)
	}
	c, err := os.CreateTemp(t.TempDir(), "current-*.parquet")
	if err != nil {
		t.Fatalf("temp current: %v", err)
	}
	_ = b.Close()
	_ = c.Close()

	diffUseFocusEngine = true
	diffOutputFile = ""
	if err := runDiff(b.Name(), c.Name()); err != nil {
		t.Fatalf("runDiff focus engine path failed: %v", err)
	}
}
