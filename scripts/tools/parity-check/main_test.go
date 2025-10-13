//go:build cgo
// +build cgo

package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"sort"
	"testing"
	"time"

	_ "github.com/marcboeker/go-duckdb"
)

// TestComputeLiteHash ensures computeLiteHash is deterministic and order-independent.
func TestComputeLiteHash(t *testing.T) {
	db, err := sql.Open("duckdb", ":memory:")
	if err != nil {
		t.Fatalf("open duckdb: %v", err)
	}
	defer func() { _ = db.Close() }()

	stmts := []string{
		"CREATE TABLE t (effective_cost DOUBLE, usage_quantity DOUBLE, provider_name VARCHAR, service_name VARCHAR, charge_category VARCHAR)",
		// Intentionally unsorted insert order
		"INSERT INTO t VALUES (2.000000, 8, 'aws', 'EC2', 'Usage')",
		"INSERT INTO t VALUES (0.500000, 5, 'aws', 'S3', 'Credit')",
		"INSERT INTO t VALUES (1.230000, 10, 'aws', 'EC2', 'Usage')",
	}
	for _, s := range stmts {
		if _, err := db.Exec(s); err != nil {
			t.Fatalf("exec %q: %v", s, err)
		}
	}
	dir := t.TempDir()
	parquetPath := filepath.Join(dir, "sample.parquet")
	if _, err := db.Exec(fmt.Sprintf("COPY t TO '%s' (FORMAT 'parquet')", parquetPath)); err != nil { //nolint:gosec
		t.Fatalf("copy to parquet: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	got, err := computeLiteHash(ctx, db, parquetPath)
	if err != nil {
		t.Fatalf("computeLiteHash error: %v", err)
	}
	if got == "" {
		t.Fatalf("empty hash")
	}

	// Expected hash using same formatting logic (sorted serializations)
	parts := []string{
		"2.000000|8.000000|aws|EC2|Usage",
		"0.500000|5.000000|aws|S3|Credit",
		"1.230000|10.000000|aws|EC2|Usage",
	}
	sort.Strings(parts)
	h := sha256.New()
	for _, p := range parts {
		_, _ = h.Write([]byte(p))
	}
	want := hex.EncodeToString(h.Sum(nil))
	if want != got {
		t.Fatalf("hash mismatch\nwant=%s\ngot=%s", want, got)
	}

	// Second run determinism
	got2, err := computeLiteHash(ctx, db, parquetPath)
	if err != nil {
		t.Fatalf("computeLiteHash second: %v", err)
	}
	if got2 != got {
		t.Fatalf("non-deterministic hash: %s vs %s", got, got2)
	}
}
