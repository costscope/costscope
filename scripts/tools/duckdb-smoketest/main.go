//go:build duckdb

package main

import (
	"database/sql"
	"fmt"

	_ "github.com/marcboeker/go-duckdb"
)

func main() {
	fmt.Println("[duckdb-smoketest] opening in-memory duckdb instance")
	db, err := sql.Open("duckdb", "")
	if err != nil {
		panic(fmt.Errorf("open duckdb: %w", err))
	}
	defer db.Close()
	var v int
	if err := db.QueryRow("SELECT 42").Scan(&v); err != nil {
		panic(fmt.Errorf("query failed: %w", err))
	}
	fmt.Printf("[duckdb-smoketest] query result=%d (expected 42)\n", v)
	fmt.Println("[duckdb-smoketest] SUCCESS")
}
