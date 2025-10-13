//go:build duckdb && !race

package writers

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/marcboeker/go-duckdb"

	"github.com/costscope/costscope/internal/core/focus/types"
)

// Validate TIMESTAMP_MILLIS round-trip via DuckDB
func TestParquetWriter_Timestamps_NoRace(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	out := filepath.Join(tmp, "ts.parquet")

	// Disable rotation for single-file output
	opts := &types.ParquetOptions{CompressionCodec: "snappy", RotateSizeBytes: -1}
	ctx := WithParquetOptions(context.Background(), opts)

	w, _, err := NewWriter(ctx, out, "parquet", types.GetFocusV12Schema())
	if err != nil {
		t.Fatalf("NewWriter: %v", err)
	}

	// Use a fixed UTC time truncated to millis so round-trip compares cleanly
	base := time.Date(2024, 10, 11, 12, 34, 56, 789*int(time.Millisecond), time.UTC)
	rec := types.FocusRecord{
		BillingAccountId:    "BA",
		BillingAccountName:  "TS",
		BillingCurrency:     "USD",
		BillingPeriodStart:  base,
		BillingPeriodEnd:    base.Add(1 * time.Hour),
		ChargeCategory:      "Usage",
		ChargeClass:         "Usage",
		ChargeDescription:   "ts",
		ChargePeriodStart:   base.Add(-24 * time.Hour),
		ChargePeriodEnd:     base.Add(24 * time.Hour),
		EffectiveCost:       1,
		ListCost:            1,
		UsageQuantity:       1,
		UsageUnit:           "Hours",
		ProviderName:        "test",
		ServiceName:         "svc",
		ResourceId:          "res",
		ConversionTimestamp: base,
	}

	if err := w.Write(ctx, []types.FocusRecord{rec}); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	db, err := sql.Open("duckdb", ":memory:")
	if err != nil {
		t.Fatalf("duckdb open: %v", err)
	}
	defer func() { _ = db.Close() }()

	if _, err := db.Exec("CREATE TABLE t AS SELECT * FROM read_parquet(?)", out); err != nil {
		t.Fatalf("duckdb read_parquet: %v", err)
	}

	var bps, bpe, cps, cpe, ct time.Time
	row := db.QueryRow("SELECT billing_period_start, billing_period_end, charge_period_start, charge_period_end, conversion_timestamp FROM t")
	if err := row.Scan(&bps, &bpe, &cps, &cpe, &ct); err != nil {
		t.Fatalf("scan: %v", err)
	}

	// DuckDB should read Parquet TIMESTAMP_MILLIS as timestamps in UTC.
	// Allow exact equality (we truncated to millis), but keep <= 1ms tolerance just in case.
	mustClose := func(name string, got, want time.Time) {
		if d := got.Sub(want); d > time.Millisecond || d < -time.Millisecond {
			t.Fatalf("%s mismatch: got=%s want=%s diff=%s", name, got.UTC(), want.UTC(), d)
		}
	}

	mustClose("billing_period_start", bps.UTC(), base)
	mustClose("billing_period_end", bpe.UTC(), base.Add(1*time.Hour))
	mustClose("charge_period_start", cps.UTC(), base.Add(-24*time.Hour))
	mustClose("charge_period_end", cpe.UTC(), base.Add(24*time.Hour))
	mustClose("conversion_timestamp", ct.UTC(), base)
}
