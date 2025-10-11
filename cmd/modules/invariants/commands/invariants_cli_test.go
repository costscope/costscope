package commands

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// helper executes a cobra command produced by builder with given args
func execCmd(t *testing.T, cmd *cobra.Command, args ...string) error {
	t.Helper()
	cmd.SetArgs(args)
	// Discard command output to keep test logs clean
	cmd.SetOut(os.Stdout)
	cmd.SetErr(os.Stderr)
	return cmd.Execute()
}

func TestInvariantsDiff_NoViolationsAndReportArtifact(t *testing.T) {
	cmd := buildDiffCommand()
	dir := t.TempDir()
	// current equals baseline
	current := map[string]any{
		"row_count":                     2,
		"sum_effective_cost":            3.0,
		"sum_list_cost":                 3.0,
		"sum_usage_quantity":            15.0,
		"charge_category_distribution":  map[string]float64{"Usage": 100},
		"pricing_category_distribution": map[string]float64{"Standard": 100},
		"provider_distribution":         map[string]float64{"aws": 100},
		"generated_at":                  "2024-01-01T00:00:00Z",
	}
	// write current and baseline files
	curPath := filepath.Join(dir, "current.json")
	basePath := filepath.Join(dir, "baseline.json")
	bcur, _ := json.Marshal(current)
	if err := os.WriteFile(curPath, bcur, 0o600); err != nil {
		t.Fatalf("write cur: %v", err)
	}
	if err := os.WriteFile(basePath, bcur, 0o600); err != nil {
		t.Fatalf("write base: %v", err)
	}

	reportPath := filepath.Join(dir, "report.json")
	artifactPtr := filepath.Join(dir, "artifact_path.txt")

	// expect no error
	if err := execCmd(t, cmd, curPath, "--baseline", basePath, "--tolerance", "0.01", "--report", reportPath, "--artifact-pointer", artifactPtr); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// report file should exist
	if _, err := os.Stat(reportPath); err != nil {
		t.Fatalf("report not created: %v", err)
	}
	// artifact pointer should contain report path
	// Ensure the path is confined to the test temp dir to satisfy gosec (G304)
	cleaned := filepath.Clean(artifactPtr)
	if !strings.HasPrefix(cleaned, filepath.Clean(dir)+string(os.PathSeparator)) {
		t.Fatalf("artifact pointer path escaped temp dir: %q not under %q", cleaned, dir)
	}
	data, err := os.ReadFile(cleaned) // #nosec G304 -- path is generated within t.TempDir and validated above
	if err != nil {
		t.Fatalf("artifact pointer missing: %v", err)
	}
	if string(data) != reportPath {
		t.Fatalf("artifact pointer content mismatch: got %q want %q", string(data), reportPath)
	}
}

func TestInvariantsDiff_ViolationsFail(t *testing.T) {
	cmd := buildDiffCommand()
	dir := t.TempDir()
	// baseline metrics
	base := map[string]any{
		"row_count":                     2,
		"sum_effective_cost":            1.0,
		"sum_list_cost":                 1.0,
		"sum_usage_quantity":            1.0,
		"charge_category_distribution":  map[string]float64{"Usage": 100},
		"pricing_category_distribution": map[string]float64{"Standard": 100},
		"provider_distribution":         map[string]float64{"aws": 100},
		"generated_at":                  "2024-01-01T00:00:00Z",
	}
	current := map[string]any{
		"row_count":                     2,
		"sum_effective_cost":            10.0, // > 1% drift
		"sum_list_cost":                 10.0,
		"sum_usage_quantity":            5.0,
		"charge_category_distribution":  map[string]float64{"Usage": 100},
		"pricing_category_distribution": map[string]float64{"Standard": 100},
		"provider_distribution":         map[string]float64{"aws": 100},
		"generated_at":                  "2024-01-01T00:00:00Z",
	}
	curPath := filepath.Join(dir, "current.json")
	basePath := filepath.Join(dir, "baseline.json")
	if b, err := json.Marshal(base); err == nil {
		_ = os.WriteFile(basePath, b, 0o600)
	} else {
		t.Fatalf("marshal base: %v", err)
	}
	if b, err := json.Marshal(current); err == nil {
		_ = os.WriteFile(curPath, b, 0o600)
	} else {
		t.Fatalf("marshal current: %v", err)
	}

	// expect error due to violations
	if err := execCmd(t, cmd, curPath, "--baseline", basePath, "--tolerance", "0.01"); err == nil {
		t.Fatalf("expected drift error, got nil")
	}
}
