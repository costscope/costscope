//go:build duckdb && !race

package writers

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/marcboeker/go-duckdb"

	"local/costscope/internal/core/focus/types"
)

// Ensures that when rotation is disabled, the single output parquet file at base path is readable by DuckDB
// and record counts match what was written.
func TestParquetWriter_NoRotation_DuckDBReadable(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	base := filepath.Join(tmp, "focus.parquet")

	opts := &types.ParquetOptions{
		CompressionCodec:  "snappy",
		RotateSizeBytes:   -1, // disable rotation explicitly
		RowGroupSizeBytes: 64 * 1024,
		PageSizeBytes:     8 * 1024,
	}
	ctx := WithParquetOptions(context.Background(), opts)

	w, _, err := NewWriter(ctx, base, "parquet", types.GetFocusV12Schema())
	if err != nil {
		t.Fatalf("NewWriter: %v", err)
	}

	now := time.Now().UTC()
	total := 200
	rec := types.FocusRecord{BillingAccountId: "A", BillingCurrency: "USD", BillingPeriodStart: now, BillingPeriodEnd: now}
	for i := 0; i < total; i++ {
		if err := w.Write(context.Background(), []types.FocusRecord{rec}); err != nil {
			t.Fatalf("write: %v", err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	db, err := sql.Open("duckdb", ":memory:")
	if err != nil {
		t.Fatalf("duckdb open: %v", err)
	}
	defer func() { _ = db.Close() }()

	if _, err := db.Exec("CREATE TABLE t AS SELECT * FROM read_parquet(?)", base); err != nil {
		t.Fatalf("duckdb read_parquet: %v", err)
	}
	var cnt int
	if err := db.QueryRow("SELECT COUNT(*) FROM t").Scan(&cnt); err != nil {
		t.Fatalf("duckdb count: %v", err)
	}
	if cnt != total {
		t.Fatalf("row count mismatch: wrote %d, read %d", total, cnt)
	}
}
