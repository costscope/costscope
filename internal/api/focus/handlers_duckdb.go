//go:build experimental && duckdb

package focus

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"strings"

	_ "github.com/marcboeker/go-duckdb"
)

// loadValuesOptional uses DuckDB to load optional effective_cost and usage_quantity arrays
// from a dataset when provided. Returns nil slices on any error.
func loadValuesOptional(path string) ([]float64, []float64) {
	if path == "" {
		return nil, nil
	}
	fn, err := tableFnFor(path)
	if err != nil {
		return nil, nil
	}
	db, err := sql.Open("duckdb", "")
	if err != nil {
		return nil, nil
	}
	defer func() { _ = db.Close() }()
	q := fmt.Sprintf("CREATE TABLE drift_api AS SELECT effective_cost, usage_quantity FROM %s", fn) //nolint:gosec
	if _, err := db.Exec(q, path); err != nil {
		return nil, nil
	}
	rows, err := db.Query("SELECT effective_cost, usage_quantity FROM drift_api")
	if err != nil {
		return nil, nil
	}
	defer func() { _ = rows.Close() }()
	eff := make([]float64, 0, 1024)
	use := make([]float64, 0, 1024)
	for rows.Next() {
		var e, u float64
		if err := rows.Scan(&e, &u); err != nil {
			return nil, nil
		}
		eff = append(eff, e)
		use = append(use, u)
	}
	return eff, use
}

func tableFnFor(path string) (string, error) {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".parquet":
		return "read_parquet(?)", nil
	case ".csv":
		return "read_csv_auto(?, HEADER=TRUE)", nil
	case ".json":
		return "read_json_auto(?)", nil
	default:
		return "", fmt.Errorf("unsupported dataset format: %s", path)
	}
}
