//go:build duckdb

package api

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"strings"

	_ "github.com/marcboeker/go-duckdb"

	"github.com/costscope/costscope/internal/core/focus/quality"
	focustypes "github.com/costscope/costscope/internal/core/focus/types"
)

func computeAPIInvariantsFromFile(path string) (quality.InvariantMetrics, error) {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".parquet":
		return computeAPIInvariantsDuckDB(path, "read_parquet(?)")
	case ".csv":
		return computeAPIInvariantsDuckDB(path, "read_csv_auto(?, HEADER=TRUE)")
	case ".json":
		return computeAPIInvariantsDuckDB(path, "read_json_auto(?)")
	default:
		return quality.InvariantMetrics{}, fmt.Errorf("unsupported format: %s", ext)
	}
}

func computeAPIInvariantsDuckDB(path string, tableFn string) (quality.InvariantMetrics, error) {
	db, err := sql.Open("duckdb", "")
	if err != nil {
		return quality.InvariantMetrics{}, err
	}
	defer func() { _ = db.Close() }()
	q := fmt.Sprintf("CREATE TABLE inv_focus AS SELECT * FROM %s", tableFn) //nolint:gosec
	if _, err := db.Exec(q, path); err != nil {
		return quality.InvariantMetrics{}, err
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
