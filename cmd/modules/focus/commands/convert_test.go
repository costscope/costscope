package commands

import (
	"os"
	"path/filepath"
	"testing"
)

func TestHelpers_ProviderAndPaths(t *testing.T) {
	// isProviderSupported
	if !isProviderSupported("aws") || !isProviderSupported("GCP") || isProviderSupported("ibm") {
		t.Fatalf("provider support check failed")
	}

	// generateOutputPath
	inDir := "/tmp/in"
	outDir := "/tmp/out"
	p := generateOutputPath(filepath.Join(inDir, "a/b/c.csv"), inDir, outDir)
	if filepath.Ext(p) != ".parquet" {
		t.Fatalf("expected .parquet, got %s", p)
	}
}

func TestValidateBatchDirectories(t *testing.T) {
	in := t.TempDir()
	out := filepath.Join(t.TempDir(), "out")
	convertInputDir = in
	convertOutputDir = out
	if err := validateBatchDirectories(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := os.Stat(out); err != nil {
		t.Fatalf("expected output dir to be created: %v", err)
	}
}
