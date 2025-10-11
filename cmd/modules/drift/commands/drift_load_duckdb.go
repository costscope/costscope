//go:build duckdb

package commands

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"strings"

	_ "github.com/marcboeker/go-duckdb"
)

func loadValues(path string) ([]float64, []float64, error) {
	fn, err := tableFnFor(path)
	if err != nil {
		return nil, nil, err
	}
	db, err := sql.Open("duckdb", "")
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = db.Close() }()
	q := fmt.Sprintf("CREATE TABLE drift_in AS SELECT effective_cost, usage_quantity FROM %s", fn) //nolint:gosec
	if _, err := db.Exec(q, path); err != nil {
		return nil, nil, err
	}
	rows, err := db.Query("SELECT effective_cost, usage_quantity FROM drift_in")
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = rows.Close() }()
	eff := make([]float64, 0, 1024)
	use := make([]float64, 0, 1024)
	for rows.Next() {
		var ec, uq float64
		if err := rows.Scan(&ec, &uq); err != nil {
			return nil, nil, err
		}
		eff = append(eff, ec)
		use = append(use, uq)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}
	return eff, use, nil
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
