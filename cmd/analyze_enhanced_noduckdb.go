//go:build !duckdb

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var analyzeEnhancedCmd = &cobra.Command{
	Use:   "analyze-enhanced [parquet-file]",
	Short: "Enhanced FOCUS data analysis (requires duckdb build tag)",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("this binary was built without DuckDB support; rebuild with -tags duckdb")
	},
}

func initAnalyzeEnhanced() {}
