package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/costscope/costscope/internal/core/focus/quality"
)

// BuildInvariantsCommand creates the root invariants command.
func BuildInvariantsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "invariants",
		Short: "Data quality invariants utilities (aggregates, distributions, drift)",
	}
	cmd.AddCommand(buildDiffCommand())
	cmd.AddCommand(buildRegenerateSubcommand())
	return cmd
}

func buildDiffCommand() *cobra.Command {
	var baseline string
	var tolerance float64
	var report string
	var artifactPointer string
	cmd := &cobra.Command{
		Use:   "diff <current.json>",
		Short: "Diff invariants JSON metrics against a baseline JSON",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			curPath := args[0]
			if baseline == "" {
				return fmt.Errorf("--baseline required")
			}
			cur, err := quality.LoadBaseline(curPath)
			if err != nil {
				return fmt.Errorf("load current: %w", err)
			}
			base, err := quality.LoadBaseline(baseline)
			if err != nil {
				return fmt.Errorf("load baseline: %w", err)
			}
			quality.CompareInvariants(&cur, base, tolerance)
			if report != "" {
				_ = quality.SaveReport(report, cur)
				if artifactPointer != "" {
					_ = os.MkdirAll(filepath.Dir(artifactPointer), 0o750)
					_ = os.WriteFile(artifactPointer, []byte(report), 0o600)
				}
			}
			b, _ := json.MarshalIndent(cur, "", "  ")
			fmt.Println(string(b))
			if len(cur.Violations) > 0 {
				return fmt.Errorf("invariants violations: %v", cur.Violations)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&baseline, "baseline", "", "Baseline invariants JSON path")
	cmd.Flags().Float64Var(&tolerance, "tolerance", 0.01, "Relative tolerance for aggregates and absolute percentage points for distributions")
	cmd.Flags().StringVar(&report, "report", "", "Optional path for comparison output JSON")
	cmd.Flags().StringVar(&artifactPointer, "artifact-pointer", "", "Optional file path to write the path of generated invariants report (for CI artifact discovery)")
	return cmd
}

// buildRegenerateCommand creates `costscope invariants regenerate <focus-file> --output baseline.json`.
// It computes invariants from a FOCUS dataset (parquet/csv/json) and writes a baseline JSON.
// buildRegenerateSubcommand is provided by tag-specific files:
// - invariants_regenerate_duckdb.go (//go:build duckdb)
// - invariants_regenerate_noduckdb.go (//go:build !duckdb)
// End of file
