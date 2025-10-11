package writers

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"local/costscope/internal/core/focus/types"
)

// Benchmark the Parquet writer throughput and allocations.
// We disable rotation to avoid rename overhead and to keep a single output path.
// Each benchmark opens one writer, writes b.N records (or batches), and then closes it.

func benchmarkParquetWriter(b *testing.B, codec string, batchSize int) {
	b.Helper()
	tmp := b.TempDir()
	out := filepath.Join(tmp, "bench.parquet")

	opts := &types.ParquetOptions{
		CompressionCodec:  codec,
		RowGroupSizeBytes: 128 * 1024 * 1024, // default 128MB
		PageSizeBytes:     8 * 1024,          // explicit default
		RotateSizeBytes:   -1,                // disable rotation for deterministic single file
		RotateInterval:    "",
	}
	ctx := WithParquetOptions(context.Background(), opts)

	w, _, err := NewWriter(ctx, out, "parquet", types.GetFocusV12Schema())
	if err != nil {
		b.Fatalf("NewWriter: %v", err)
	}

	// Prepare a base record and mutate a few fields to avoid identical rows.
	now := time.Now().UTC()
	rec := types.FocusRecord{
		BillingAccountId:    "BA",
		BillingAccountName:  "Bench",
		BillingCurrency:     "USD",
		BillingPeriodStart:  now,
		BillingPeriodEnd:    now,
		ChargeCategory:      "Usage",
		ChargeClass:         "Usage",
		ChargeDescription:   "benchmark",
		EffectiveCost:       1.0,
		ListCost:            1.0,
		UsageQuantity:       1.0,
		UsageUnit:           "Hours",
		ProviderName:        "bench",
		ServiceName:         "writer",
		ResourceId:          "res",
		ConversionTimestamp: now,
	}

	b.ReportAllocs()
	b.ResetTimer()

	if batchSize <= 1 {
		for i := 0; i < b.N; i++ {
			rec.UsageQuantity = float64(i + 1)
			if err := w.Write(ctx, []types.FocusRecord{rec}); err != nil {
				b.Fatalf("write: %v", err)
			}
		}
	} else {
		// Write in chunks of batchSize per iteration to model production batching.
		batch := make([]types.FocusRecord, batchSize)
		for i := 0; i < b.N; i++ {
			for j := 0; j < batchSize; j++ {
				rec.UsageQuantity = float64(i*batchSize + j + 1)
				batch[j] = rec
			}
			if err := w.Write(ctx, batch); err != nil {
				b.Fatalf("write(batch): %v", err)
			}
		}
		// Account for bytes per iteration if desired by callers using -benchmem.
		// We intentionally skip SetBytes here due to variability of parquet encoding size.
	}

	b.StopTimer()
	// Ensure any buffered data is flushed before closing.
	if err := w.Flush(ctx); err != nil {
		b.Fatalf("flush: %v", err)
	}
	if err := w.Close(); err != nil {
		b.Fatalf("close: %v", err)
	}
}

func BenchmarkParquetWriter_SingleRecord_Snappy(b *testing.B) {
	benchmarkParquetWriter(b, "snappy", 1)
}

func BenchmarkParquetWriter_Batch100_Snappy(b *testing.B) {
	benchmarkParquetWriter(b, "snappy", 100)
}

func BenchmarkParquetWriter_Batch100_ZSTD(b *testing.B) {
	benchmarkParquetWriter(b, "zstd", 100)
}
