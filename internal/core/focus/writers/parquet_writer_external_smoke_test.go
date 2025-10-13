package writers

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/costscope/costscope/internal/core/focus/types"
)

// NOTE: This is a scaffold for external engine smoke tests.
// It writes a tiny Parquet file and relies on environment-provided validators
// (e.g., TRINO_JDBC_URL, ATHENA_CATALOG/DB) to verify read-back. The test is
// skipped unless the relevant env variables are present. Implementors can
// extend with actual queries using their preferred client libraries.
func TestParquetWriter_External_Engines_Smoke(t *testing.T) {
	ctx := context.Background()

	// Gate via env; if neither is present, skip.
	hasTrino := os.Getenv("TRINO_JDBC_URL") != ""
	hasAthena := os.Getenv("ATHENA_CATALOG") != "" && os.Getenv("ATHENA_DB") != ""
	if !hasTrino && !hasAthena {
		t.Skip("no external engine configuration provided; set TRINO_JDBC_URL or ATHENA_CATALOG/ATHENA_DB to enable")
	}

	tmp := t.TempDir()
	out := tmp + "/ext.parquet"

	opts := &types.ParquetOptions{CompressionCodec: "snappy", RotateSizeBytes: -1}
	ctx = WithParquetOptions(ctx, opts)

	w, _, err := NewWriter(ctx, out, "parquet", types.GetFocusV12Schema())
	if err != nil {
		t.Fatalf("NewWriter: %v", err)
	}

	now := time.UnixMilli(time.Now().UnixMilli()).UTC()

	rec := types.FocusRecord{
		BillingAccountId:    "BA",
		BillingAccountName:  "EXT",
		BillingCurrency:     "USD",
		BillingPeriodStart:  now,
		BillingPeriodEnd:    now,
		ChargeCategory:      "Usage",
		ChargeClass:         "Usage",
		ChargeDescription:   "smoke",
		ChargePeriodStart:   now,
		ChargePeriodEnd:     now,
		EffectiveCost:       0.01,
		ListCost:            0.01,
		UsageQuantity:       1,
		UsageUnit:           "Hours",
		ProviderName:        "test",
		ServiceName:         "svc",
		ResourceId:          "res",
		ConversionTimestamp: now,
	}

	if err := w.Write(ctx, []types.FocusRecord{rec}); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	// Minimal assertion: file exists and non-zero size; external CI job can
	// pick up and run a query using Trino/Athena to validate types.
	st, err := os.Stat(out)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if st.Size() <= 0 {
		t.Fatalf("expected non-empty parquet file")
	}
}
