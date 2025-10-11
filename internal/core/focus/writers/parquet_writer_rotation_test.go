//go:build !race
// +build !race

package writers

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"local/costscope/internal/core/focus/types"
)

// Test rotation by small size threshold triggers multiple files
func TestParquetWriter_RotateBySize(t *testing.T) {
	tmp := t.TempDir()
	out := filepath.Join(tmp, "focus.parquet")
	opts := &types.ParquetOptions{
		CompressionCodec:  "snappy",
		RowGroupSizeBytes: 8 * 1024, // tiny row group
		RotateSizeBytes:   2 * 1024, // low threshold to ensure rotation
		PageSizeBytes:     4 * 1024,
	}
	ctx := WithParquetOptions(context.Background(), opts)

	w, _, err := NewWriter(ctx, out, "parquet", types.GetFocusV12Schema())
	if err != nil {
		t.Fatalf("NewWriter: %v", err)
	}
	defer func() { _ = w.Close() }()

	// write enough records to exceed rotate size several times
	rec := types.FocusRecord{BillingAccountId: "a", BillingCurrency: "USD", ProviderName: "aws", ChargeDescription: strings.Repeat("X", 4096)}
	batch := make([]types.FocusRecord, 0, 2000)
	for i := 0; i < 2000; i++ {
		batch = append(batch, rec)
	}
	if err := w.Write(context.Background(), batch); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	// Verify multiple parquet files exist with expected pattern
	entries, err := os.ReadDir(tmp)
	if err != nil {
		t.Fatalf("readdir: %v", err)
	}
	var parquetCount int
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".parquet") && strings.Contains(e.Name(), "focus-") {
			parquetCount++
		}
	}
	if parquetCount < 2 {
		t.Fatalf("expected rotated parquet files >=2, got %d", parquetCount)
	}
}
