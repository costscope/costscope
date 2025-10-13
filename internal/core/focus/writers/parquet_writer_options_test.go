//go:build !race
// +build !race

package writers

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/costscope/costscope/internal/core/focus/types"
)

func TestParquetWriter_OptionsPath(t *testing.T) {
	tmp := t.TempDir()
	out := filepath.Join(tmp, "out.parquet")
	// Use non-default settings to exercise code path
	opts := &types.ParquetOptions{
		CompressionCodec:  "gzip",
		RowGroupSizeBytes: 4 * 1024 * 1024,
		PageSizeBytes:     16 * 1024,
		// Disable rotation so Flush writes to the base path directly
		RotateSizeBytes: -1,
	}
	ctx := WithParquetOptions(context.Background(), opts)

	schema := types.GetFocusV12Schema()
	w, outFmt, err := NewWriter(ctx, out, "parquet", schema)
	if err != nil {
		t.Fatalf("NewWriter: %v", err)
	}
	if outFmt != "FOCUS_PARQUET" {
		t.Fatalf("unexpected outFmt: %s", outFmt)
	}
	defer func() { _ = w.Close() }()

	rec := types.FocusRecord{BillingAccountId: "a", BillingCurrency: "USD"}
	if err := w.Write(context.Background(), []types.FocusRecord{rec}); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := w.Flush(context.Background()); err != nil {
		t.Fatalf("flush: %v", err)
	}
	meta := w.GetMetadata()
	if meta == nil || meta.FilePath == "" {
		t.Fatalf("missing metadata")
	}
	// #nosec G304 - reading file we just wrote in temp dir
	if b, err := os.ReadFile(meta.FilePath); err != nil || len(b) == 0 {
		t.Fatalf("expected non-empty parquet file: %v, size=%d", err, len(b))
	}
}
