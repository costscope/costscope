//go:build duckdb
// +build duckdb

package pipeline

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// TestEndToEndPipeline ensures aggregate drift ≤0.1% and no invariant violations.
func TestEndToEndPipeline(t *testing.T) {
	// Derive fixture absolute paths based on test file location for portability.
	_, thisFile, _, _ := runtime.Caller(0)
	base := filepath.Clean(filepath.Join(filepath.Dir(thisFile), "..", "..", "fixtures", "aws"))
	fixtures := []string{filepath.Join(base, "cur_baseline_sample.csv")} // matches baseline invariants
	tempDir, err := os.MkdirTemp("", "e2e-pipeline-*")
	if err != nil {
		t.Fatalf("temp dir: %v", err)
	}
	baseline := filepath.Join(filepath.Dir(base), "quality", "baseline_invariants.json")
	rep, err := Run(context.Background(), RunConfig{Provider: "aws", InputFiles: fixtures, WorkDir: tempDir, DriftTolerance: 0.001, ValidateOutput: true, BaselinePath: baseline, InvariantTolerance: 0.01})
	if err != nil {
		t.Fatalf("pipeline run error: %v", err)
	}
	reportPath := filepath.Join(tempDir, "e2e_report.json")
	if err := rep.Save(reportPath); err != nil {
		t.Fatalf("failed to save report: %v", err)
	}
	t.Logf("E2E_REPORT_PATH=%s", reportPath) // CI parses this key
	if !rep.Passed {
		t.Fatalf("pipeline integrity failed: notes=%v drift=%v violations=%v", rep.Notes, rep.RelativeDrift, rep.Invariants.Violations)
	}
}

// TestMultiProviderEndToEnd covers Azure and GCP minimal fixtures for regression & schema sanity.
func TestMultiProviderEndToEnd(t *testing.T) {
	_, thisFile, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(thisFile), "..", ".."))
	azureBase := filepath.Join(root, "fixtures", "azure")
	gcpBase := filepath.Join(root, "fixtures", "gcp")
	// Baseline not used for multi-provider test (provider distributions differ)
	azureFiles := []string{filepath.Join(azureBase, "usage.csv"), filepath.Join(azureBase, "tax_refund.csv"), filepath.Join(azureBase, "reservation_credit.csv")}
	gcpFiles := []string{filepath.Join(gcpBase, "usage_minimal.csv"), filepath.Join(gcpBase, "credit_cud.csv"), filepath.Join(gcpBase, "credit_spot.csv"), filepath.Join(gcpBase, "credit_sustained_promo.csv")}
	tempDir, err := os.MkdirTemp("", "e2e-multi-*")
	if err != nil {
		t.Fatalf("temp dir: %v", err)
	}
	azureBaseline := filepath.Join(root, "fixtures", "quality", "baseline_azure_invariants.json")
	azRep, err := Run(context.Background(), RunConfig{Provider: "azure", InputFiles: azureFiles, WorkDir: tempDir, DriftTolerance: 0.001, ValidateOutput: true, BaselinePath: azureBaseline, InvariantTolerance: 0.01})
	if err != nil {
		t.Fatalf("azure pipeline error: %v", err)
	}
	if !azRep.Passed {
		t.Fatalf("azure pipeline failed: notes=%v drift=%v inv=%v", azRep.Notes, azRep.RelativeDrift, azRep.Invariants.Violations)
	}
	gcpBaseline := filepath.Join(root, "fixtures", "quality", "baseline_gcp_invariants.json")
	gcpRep, err := Run(context.Background(), RunConfig{Provider: "gcp", InputFiles: gcpFiles, WorkDir: tempDir, DriftTolerance: 0.001, ValidateOutput: true, BaselinePath: gcpBaseline, InvariantTolerance: 0.01})
	if err != nil {
		t.Fatalf("gcp pipeline error: %v", err)
	}
	if !gcpRep.Passed {
		t.Fatalf("gcp pipeline failed: notes=%v drift=%v inv=%v", gcpRep.Notes, gcpRep.RelativeDrift, gcpRep.Invariants.Violations)
	}
	// Save combined indicator files for CI artifact (optional)
	if err := azRep.Save(filepath.Join(tempDir, "azure_report.json")); err == nil {
		t.Logf("E2E_AZURE_REPORT=%s", filepath.Join(tempDir, "azure_report.json"))
	}
	if err := gcpRep.Save(filepath.Join(tempDir, "gcp_report.json")); err == nil {
		t.Logf("E2E_GCP_REPORT=%s", filepath.Join(tempDir, "gcp_report.json"))
	}
}
