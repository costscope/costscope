package standalone

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	focuscmd "github.com/costscope/costscope/cmd/modules/focus/commands"
	"github.com/costscope/costscope/internal/testutil"
)

// invariantsReport holds just the fields we assert for the smoke test.
type invariantsReport struct {
	RowCount                    int     `json:"row_count"`
	SumEffectiveCost            float64 `json:"sum_effective_cost"`
	NegativeUsageViolationCount int     `json:"negative_usage_violation_count"`
}

// TestConvert_InvariantsSmoke performs a very fast end‑to‑end invariants collection run
// over tiny CSV fixtures for aws / azure / gcp. Acceptance criteria (M6):
//   - row_count > 0
//   - sum_effective_cost >= 0
//   - negative_usage_violation_count == 0
//   - total runtime < ~2s (guarded implicitly by keeping dataset tiny and avoiding heavy flags)
func TestConvert_InvariantsSmoke(t *testing.T) {
	if testing.Short() { // allow skipping in short mode
		t.Skip("short")
	}
	start := time.Now()

	// Use the centralized test helper to derive the repository root.
	repoRoot := testutil.FindRepoRoot(t)

	cases := []struct {
		name     string
		provider string
		inputRel string
	}{
		{name: "aws", provider: "aws", inputRel: "tests/fixtures/aws/cur_smoke.csv"},
		{name: "azure", provider: "azure", inputRel: "tests/fixtures/azure/usage_smoke.csv"},
		{name: "gcp", provider: "gcp", inputRel: "tests/fixtures/gcp/usage_smoke.csv"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			out := filepath.Join(tmpDir, "focus.parquet")
			rep := filepath.Join(tmpDir, "inv.json")
			cmd := focuscmd.BuildConvertCommand()
			// minimal required flags: provider, input, output, invariants + report path, quiet to reduce noise
			inputPath := filepath.Join(repoRoot, c.inputRel)
			if _, err := os.Stat(inputPath); err != nil {
				t.Fatalf("fixture missing: %s (%v)", inputPath, err)
			}
			args := []string{"--provider", c.provider, "--input", inputPath, "--output", out, "--streaming", "--invariants", "--invariants-report", rep, "--quiet"}
			cmd.SetArgs(args)
			cmd.SetOut(os.Stdout)
			cmd.SetErr(os.Stderr)
			if err := cmd.Execute(); err != nil {
				t.Fatalf("convert failed (%s): %v", c.name, err)
			}
			// The report path is generated within t.TempDir() (unique per test run) and not influenced by user input.
			// Validate it resides under the temp directory root before reading to satisfy gosec (G304) intent.
			if !strings.HasPrefix(rep, tmpDir+string(os.PathSeparator)) {
				t.Fatalf("unexpected report path outside tempdir: %s", rep)
			}
			b, err := os.ReadFile(rep) //nolint:gosec // rep is a deterministic path under t.TempDir(), not user-controlled
			if err != nil {
				t.Fatalf("read invariants report: %v", err)
			}
			var inv invariantsReport
			if err := json.Unmarshal(b, &inv); err != nil {
				t.Fatalf("unmarshal invariants report: %v", err)
			}
			if inv.RowCount <= 0 {
				t.Fatalf("expected row_count>0, got %d", inv.RowCount)
			}
			if inv.SumEffectiveCost < 0 { // should never be negative for these samples
				t.Fatalf("expected sum_effective_cost>=0, got %f", inv.SumEffectiveCost)
			}
			if inv.NegativeUsageViolationCount != 0 {
				t.Fatalf("expected negative_usage_violation_count==0, got %d", inv.NegativeUsageViolationCount)
			}
		})
	}

	if dur := time.Since(start); dur > 2*time.Second {
		t.Fatalf("invariants smoke tests exceeded 2s budget: %v", dur)
	}
}
