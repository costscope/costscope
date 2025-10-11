//go:build duckdb && !race
// +build duckdb,!race

package writers

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/marcboeker/go-duckdb"

	"local/costscope/internal/core/focus/types"
)

// Smoke test: write a tiny Parquet and ensure DuckDB can read it back.
func TestParquet_DuckDBSmoke_NoRace(t *testing.T) {
	tmp := t.TempDir()
	out := filepath.Join(tmp, "focus.parquet")

	// Disable rotation for this smoke test to write directly to the requested path
	opts := &types.ParquetOptions{CompressionCodec: "snappy", RotateSizeBytes: -1, RotateInterval: ""}
	ctx := WithParquetOptions(context.Background(), opts)

	// Build writer via factory to mimic production path
	w, _, err := NewWriter(ctx, out, "parquet", types.GetFocusV12Schema())
	if err != nil {
		t.Fatalf("open writer: %v", err)
	}

	now := time.Now().UTC()
	rec := types.FocusRecord{
		BillingAccountId:    "BA",
		BillingAccountName:  "Main",
		BillingCurrency:     "USD",
		BillingPeriodStart:  now,
		BillingPeriodEnd:    now,
		ChargeCategory:      "Usage",
		ChargeClass:         "Usage",
		ChargeDescription:   "test",
		EffectiveCost:       1.23,
		ListCost:            1.23,
		UsageQuantity:       1,
		UsageUnit:           "Hours",
		ProviderName:        "test",
		ServiceName:         "service",
		ResourceId:          "res-1",
		ConversionTimestamp: now,
	}
	if err := w.Write(ctx, []types.FocusRecord{rec}); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	// Ensure file exists and is non-empty
	fi, err := os.Stat(out)
	if err != nil || fi.Size() == 0 {
		if err == nil {
			t.Fatalf("expected non-empty parquet output: size=%d", fi.Size())
		}
		t.Fatalf("expected parquet output: %v", err)
	}

	// DuckDB read
	db, err := sql.Open("duckdb", ":memory:")
	if err != nil {
		t.Fatalf("duckdb open: %v", err)
	}
	defer func() { _ = db.Close() }()

	// Create table from parquet
	if _, err := db.Exec("CREATE TABLE t AS SELECT * FROM read_parquet(?)", out); err != nil {
		t.Fatalf("duckdb read_parquet: %v", err)
	}
	// Count rows
	var cnt int
	if err := db.QueryRow("SELECT COUNT(*) FROM t").Scan(&cnt); err != nil {
		t.Fatalf("duckdb count: %v", err)
	}
	if cnt != 1 {
		t.Fatalf("expected 1 row, got %d", cnt)
	}
}
