package standalone

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	focuscmd "github.com/costscope/costscope/cmd/modules/focus/commands"
	"github.com/costscope/costscope/internal/testutil"
)

// TestConvert_InvariantsNegativeDrift simulates a drift violation by using a baseline
// with a deliberately different aggregate (row_count) causing a violation under the
// default tolerance (1%). It asserts that --fail-on-invariants yields a conversion
// error and that violations are surfaced in the report.
func TestConvert_InvariantsNegativeDrift(t *testing.T) {
	if testing.Short() {
		t.Skip("short")
	}
	start := time.Now()

	// Use centralized helper to locate repo root
	repoRoot := testutil.FindRepoRoot(t)
	inputPath := filepath.Join(repoRoot, "tests/fixtures/aws/cur_smoke.csv")
	if _, err := os.Stat(inputPath); err != nil {
		t.Fatalf("fixture missing: %s (%v)", inputPath, err)
	}

	// Prepare baseline with exaggerated row_count so actual run drifts.
	baseline := map[string]any{
		"row_count":                     9999, // far from real (~3) thus > tolerance
		"sum_effective_cost":            0,    // not used for drift here
		"sum_list_cost":                 0,
		"sum_usage_quantity":            0,
		"charge_category_distribution":  map[string]float64{"Usage": 100},
		"pricing_category_distribution": map[string]float64{"OnDemand": 100},
		"provider_distribution":         map[string]float64{"aws": 100},
	}

	bdir := t.TempDir()
	baselinePath := filepath.Join(bdir, "baseline.json")
	b, _ := json.MarshalIndent(baseline, "", "  ")
	if err := os.WriteFile(baselinePath, b, 0o600); err != nil {
		t.Fatalf("write baseline: %v", err)
	}

	rep := filepath.Join(bdir, "inv.json")
	out := filepath.Join(bdir, "focus.parquet")
	cmd := focuscmd.BuildConvertCommand()
	args := []string{"--provider", "aws", "--input", inputPath, "--output", out, "--streaming", "--invariants", "--invariants-report", rep, "--invariants-baseline", baselinePath, "--fail-on-invariants", "--quiet"}
	cmd.SetArgs(args)
	cmd.SetOut(os.Stdout)
	cmd.SetErr(os.Stderr)
	if err := cmd.Execute(); err == nil { // expect failure
		t.Fatalf("expected conversion failure due to invariants drift, got success")
	}
	// read report to confirm violations recorded
	bb, err := os.ReadFile(rep) //nolint:gosec // path is under TempDir
	if err != nil {
		t.Fatalf("read invariants report: %v", err)
	}
	var rpt map[string]any
	if err := json.Unmarshal(bb, &rpt); err != nil {
		t.Fatalf("unmarshal invariants report: %v", err)
	}
	violations, ok := rpt["violations"].([]any)
	if !ok || len(violations) == 0 {
		t.Fatalf("expected violations present in report, got: %v", rpt["violations"])
	}

	if time.Since(start) > 2*time.Second { // keep test fast
		t.Fatalf("negative drift test exceeded 2s budget")
	}
}
