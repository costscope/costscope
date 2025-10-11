//go:build !duckdb

package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

// buildRegenerateSubcommand (!duckdb) provides a stub that informs users that DuckDB is required.
func buildRegenerateSubcommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "regenerate <focus-file>",
		Short: "Recompute invariants baseline JSON from a FOCUS dataset (parquet/csv/json)",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return fmt.Errorf("invariants regenerate requires a binary built with -tags duckdb")
		},
	}
	// Define flags for parity with duckdb build (no-ops here)
	cmd.Flags().StringP("output", "o", "", "Output baseline invariants JSON path")
	cmd.Flags().Bool("force", false, "Overwrite existing output file")
	cmd.Flags().StringSlice("meta", nil, "Additional metadata key=value pairs to embed in baseline JSON")
	cmd.Flags().Float64("tolerance", 0.0, "Optional recommended tolerance to embed (e.g. 0.01 for ±1%)")
	return cmd
}
