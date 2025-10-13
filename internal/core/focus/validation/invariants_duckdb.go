//go:build duckdb

package validation

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"strings"

	_ "github.com/marcboeker/go-duckdb"

	"github.com/costscope/costscope/internal/core/focus/quality"
	focustypes "github.com/costscope/costscope/internal/core/focus/types"
)

// ComputeInvariantsFromFile dispatches to DuckDB for supported extensions.
func ComputeInvariantsFromFile(path string) (quality.InvariantMetrics, error) {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".parquet":
		return computeInvariantsDuckDB(path, "read_parquet(?)")
	case ".csv":
		return computeInvariantsDuckDB(path, "read_csv_auto(?, HEADER=TRUE)")
	case ".json":
		return computeInvariantsDuckDB(path, "read_json_auto(?)")
	default:
		return quality.InvariantMetrics{}, fmt.Errorf("unsupported invariants input format: %s", ext)
	}
}

func computeInvariantsDuckDB(path, tableFn string) (quality.InvariantMetrics, error) {
	db, err := sql.Open("duckdb", "")
	if err != nil {
		return quality.InvariantMetrics{}, err
	}
	defer func() { _ = db.Close() }()
	// Whitelist accepted function patterns to avoid arbitrary injection in tableFn.
	switch tableFn {
	case "read_parquet(?)", "read_csv_auto(?, HEADER=TRUE)", "read_json_auto(?)":
		// allowed
	default:
		return quality.InvariantMetrics{}, fmt.Errorf("unsupported table function: %s", tableFn)
	}
	var createStmt string
	switch tableFn { // exact statements, no concatenation of untrusted input beyond validated token
	case "read_parquet(?)":
		createStmt = "CREATE TABLE inv_focus AS SELECT * FROM read_parquet(?)"
	case "read_csv_auto(?, HEADER=TRUE)":
		createStmt = "CREATE TABLE inv_focus AS SELECT * FROM read_csv_auto(?, HEADER=TRUE)"
	case "read_json_auto(?)":
		createStmt = "CREATE TABLE inv_focus AS SELECT * FROM read_json_auto(?)"
	default:
		return quality.InvariantMetrics{}, fmt.Errorf("unsupported table function build: %s", tableFn)
	}
	if _, err := db.Exec(createStmt, path); err != nil {
		return quality.InvariantMetrics{}, fmt.Errorf("duckdb load: %w", err)
	}
	rows, err := db.Query(`SELECT effective_cost,list_cost,usage_quantity,charge_category,pricing_category,provider_name,resource_id FROM inv_focus`)
	if err != nil {
		return quality.InvariantMetrics{}, err
	}
	defer func() { _ = rows.Close() }()
	recs := make([]focustypes.FocusRecord, 0, 1024)
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
