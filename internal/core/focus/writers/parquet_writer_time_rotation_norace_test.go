//go:build !race
// +build !race

package writers

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/costscope/costscope/internal/core/focus/types"
)

// Test rotation by time interval
func TestParquetWriter_RotateByTime_NoRace(t *testing.T) {
	tmp := t.TempDir()
	out := filepath.Join(tmp, "dataset.parquet")
	opts := &types.ParquetOptions{
		CompressionCodec: "snappy",
		RotateInterval:   "10ms",
		RotateSizeBytes:  0, // keep default size-based rotation
	}
	ctx := WithParquetOptions(context.Background(), opts)

	w, _, err := NewWriter(ctx, out, "parquet", types.GetFocusV12Schema())
	if err != nil {
		t.Fatalf("NewWriter: %v", err)
	}
	defer func() { _ = w.Close() }()

	// Write across multiple intervals
	for i := 0; i < 5; i++ {
		if err := w.Write(context.Background(), []types.FocusRecord{{BillingAccountId: "b", BillingCurrency: "USD"}}); err != nil {
			t.Fatalf("write: %v", err)
		}
		time.Sleep(12 * time.Millisecond)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
	entries, err := os.ReadDir(tmp)
	if err != nil {
		t.Fatalf("readdir: %v", err)
	}
	var parquetCount int
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".parquet") && strings.Contains(e.Name(), "dataset-") {
			parquetCount++
		}
	}
	if parquetCount < 2 {
		t.Fatalf("expected >=2 files due to time rotation, got %d", parquetCount)
	}
}
