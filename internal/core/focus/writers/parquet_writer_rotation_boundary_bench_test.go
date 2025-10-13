package writers

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/costscope/costscope/internal/core/focus/types"
)

// This micro-benchmark stresses the rotation boundary behavior with large batches.
// Motivation: unified mapping path showed a duration regression when both rotation
// is enabled and chunk size is large (e.g., 50k). While mapping is outside the
// writer, the writer’s rotate-per-row check and finalize/rename on boundary can
// dominate wall time when frequently triggered mid-batch. These benches isolate
// that cost by writing synthetic FOCUS records with size rotation enabled.
//
// Notes:
// - We keep RotateSizeBytes small to force multiple rotations quickly without
//   generating huge files on disk during the benchmark run.
// - Batch size is large (50k) to mimic unified mapper chunk sizes.
// - We use simple, minimally varying records; parquet-go still encodes per row.

func benchRotationBoundary(b *testing.B, codec string, rotateBytes int64, batchSize int) {
	b.Helper()
	tmp := b.TempDir()
	out := filepath.Join(tmp, "rotation-boundary.parquet")

	opts := &types.ParquetOptions{
		CompressionCodec:  codec,
		RowGroupSizeBytes: 128 * 1024 * 1024, // default 128MB
		PageSizeBytes:     8 * 1024,          // explicit default
		RotateSizeBytes:   rotateBytes,       // enable size rotation
		RotateInterval:    "",                // time rotation disabled
		FilePrefix:        "bench-rot",
	}
	ctx := WithParquetOptions(context.Background(), opts)

	w, _, err := NewWriter(ctx, out, "parquet", types.GetFocusV12Schema())
	if err != nil {
		b.Fatalf("NewWriter: %v", err)
	}

	now := time.Now().UTC()
	base := types.FocusRecord{
		BillingAccountId:    "BA",
		BillingAccountName:  "Bench",
		BillingCurrency:     "USD",
		BillingPeriodStart:  now,
		BillingPeriodEnd:    now,
		ChargeCategory:      "Usage",
		ChargeClass:         "Usage",
		ChargeDescription:   "rotation-boundary",
		EffectiveCost:       1.0,
		ListCost:            1.0,
		UsageQuantity:       1.0,
		UsageUnit:           "Hours",
		ProviderName:        "bench",
		ServiceName:         "writer",
		ResourceId:          "res",
		ConversionTimestamp: now,
	}

	batch := make([]types.FocusRecord, batchSize)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for j := 0; j < batchSize; j++ {
			r := base
			// Minimal mutation to avoid completely identical rows
			r.UsageQuantity = float64(i*batchSize + j + 1)
			batch[j] = r
		}
		if err := w.Write(ctx, batch); err != nil {
			b.Fatalf("write(batch): %v", err)
		}
	}

	b.StopTimer()
	if err := w.Flush(ctx); err != nil {
		b.Fatalf("flush: %v", err)
	}
	if err := w.Close(); err != nil {
		b.Fatalf("close: %v", err)
	}
}

// Rotate aggressively (64 KiB) to trigger multiple rotate events mid-batch.
func BenchmarkParquetWriter_RotationBoundary_Batch50000_Snappy(b *testing.B) {
	benchRotationBoundary(b, "snappy", 64*1024, 50000)
}

// ZSTD typically yields different row sizes; include for comparative profiling.
func BenchmarkParquetWriter_RotationBoundary_Batch50000_ZSTD(b *testing.B) {
	benchRotationBoundary(b, "zstd", 64*1024, 50000)
}
