//go:build duckdb
// +build duckdb

package duckdb

import (
	"database/sql"

	_ "github.com/marcboeker/go-duckdb"
)

// Linked reports whether this binary was built with DuckDB support.
func Linked() bool { return true }

// QuickPing opens an in-memory DuckDB connection and pings it to ensure the
// embedded engine is functional. It is intentionally lightweight.
func QuickPing() error {
	db, err := sql.Open("duckdb", ":memory:")
	if err != nil {
		return err
	}
	defer db.Close()
	return db.Ping()
}
