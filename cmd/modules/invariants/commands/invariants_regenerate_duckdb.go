//go:build duckdb

package commands

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/marcboeker/go-duckdb" // register duckdb driver
	"github.com/spf13/cobra"

	"local/costscope/internal/core/focus/quality"
	focustypes "local/costscope/internal/core/focus/types"
)

// buildRegenerateSubcommand (duckdb) creates `costscope invariants regenerate` with DuckDB-backed loaders.
func buildRegenerateSubcommand() *cobra.Command { //nolint:funlen
	var output string
	var force bool
	var metadataPairs []string
	var tolerance float64
	cmd := &cobra.Command{
		Use:   "regenerate <focus-file>",
		Short: "Recompute invariants baseline JSON from a FOCUS dataset (parquet/csv/json)",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			if output == "" {
				return fmt.Errorf("--output required")
			}
			if _, err := os.Stat(output); err == nil && !force {
				return fmt.Errorf("output exists: %s (use --force to overwrite)", output)
			}
			inv, err := computeInvariantsFromAny(args[0])
			if err != nil {
				return err
			}
			if inv.Metadata == nil {
				inv.Metadata = map[string]string{}
			}
			inv.Metadata["baseline_source"] = "regenerated_cli"
			inv.Metadata["baseline_placeholder"] = "false"
			inv.Metadata["generated_by"] = "costscope"
			inv.Metadata["generated_at_rfc3339"] = time.Now().UTC().Format(time.RFC3339)
			if tolerance > 0 {
				inv.Metadata["recommended_tolerance"] = fmt.Sprintf("%g", tolerance)
			}
			for _, kv := range metadataPairs {
				// accept key=value
				var k, v string
				for i, ch := range kv {
					if ch == '=' {
						k = kv[:i]
						v = kv[i+1:]
						break
					}
				}
				if k != "" {
					inv.Metadata[k] = v
				}
			}
			if err := quality.SaveReport(output, inv); err != nil {
				return fmt.Errorf("save baseline: %w", err)
			}
			fmt.Printf(" Invariants baseline written: %s\n", output)
			return nil
		},
	}
	cmd.Flags().StringVarP(&output, "output", "o", "", "Output baseline invariants JSON path")
	cmd.Flags().BoolVar(&force, "force", false, "Overwrite existing output file")
	cmd.Flags().StringSliceVar(&metadataPairs, "meta", nil, "Additional metadata key=value pairs to embed in baseline JSON")
	cmd.Flags().Float64Var(&tolerance, "tolerance", 0.0, "Optional recommended tolerance to embed (e.g. 0.01 for ±1%)")
	return cmd
}

// computeInvariantsFromAny minimal helper (parquet/csv/json).
func computeInvariantsFromAny(path string) (quality.InvariantMetrics, error) { //nolint:dupword
	ext := filepath.Ext(path)
	switch ext {
	case ".parquet":
		return computeInvariantsDuckDB(path, "read_parquet(?)")
	case ".csv":
		return computeInvariantsDuckDB(path, "read_csv_auto(?, HEADER=TRUE)")
	case ".json":
		return computeInvariantsDuckDB(path, "read_json_auto(?)")
	default:
		return quality.InvariantMetrics{}, fmt.Errorf("unsupported file extension: %s", ext)
	}
}

// computeInvariantsDuckDB loads records via DuckDB and computes invariants.
func computeInvariantsDuckDB(path string, tableFn string) (quality.InvariantMetrics, error) { //nolint:dupl
	db, err := sql.Open("duckdb", "")
	if err != nil {
		return quality.InvariantMetrics{}, err
	}
	defer func() { _ = db.Close() }()
	query := fmt.Sprintf("CREATE TABLE inv_focus AS SELECT * FROM %s", tableFn) //nolint:gosec
	if _, err := db.Exec(query, path); err != nil {
		return quality.InvariantMetrics{}, fmt.Errorf("duckdb load: %w", err)
	}
	rows, err := db.Query(`SELECT effective_cost,list_cost,usage_quantity,charge_category,pricing_category,provider_name,resource_id FROM inv_focus`)
	if err != nil {
		return quality.InvariantMetrics{}, err
	}
	defer func() { _ = rows.Close() }()
	var recs []focustypes.FocusRecord
	for rows.Next() {
		var r focustypes.FocusRecord
		if err := rows.Scan(&r.EffectiveCost, &r.ListCost, &r.UsageQuantity, &r.ChargeCategory, &r.PricingCategory, &r.ProviderName, &r.ResourceId); err != nil {
			return quality.InvariantMetrics{}, err
		}
		recs = append(recs, r)
	}
	if err := rows.Err(); err != nil {
		return quality.InvariantMetrics{}, err
	}
	return quality.ComputeInvariants(recs), nil
}
