package writers

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"local/costscope/internal/core/focus/types"
)

// Test that fsync is only invoked at rotation finalization and final Close (once per output file),
// and not per-row. Also verifies naming and absence of lingering temp files implicitly.
func TestParquetWriter_FsyncCounts_WithRotation(t *testing.T) {
	tmp := t.TempDir()
	base := filepath.Join(tmp, "focus.parquet")

	// Small threshold to force multiple rotated files
	const rotateThreshold = int64(2 * 1024) // 2KB
	opts := &types.ParquetOptions{
		CompressionCodec:  "snappy",
		RotateSizeBytes:   rotateThreshold,
		RowGroupSizeBytes: 8 * 1024,
		PageSizeBytes:     4 * 1024,
		FilePrefix:        "fsynctest",
	}
	ctx := WithParquetOptions(context.Background(), opts)

	// Hook syncFile to count invocations while preserving real behavior
	var syncCalls int32
	orig := syncFile
	syncFile = func(f *os.File) error {
		atomic.AddInt32(&syncCalls, 1)
		return f.Sync()
	}
	t.Cleanup(func() { syncFile = orig })

	w, _, err := NewWriter(ctx, base, "parquet", types.GetFocusV12Schema())
	if err != nil {
		t.Fatalf("NewWriter: %v", err)
	}

	// Write enough rows to trigger multiple rotations
	now := time.Now().UTC()
	payload := strings.Repeat("X", 1500)
	rec := types.FocusRecord{
		BillingAccountId:    "BA",
		BillingAccountName:  "Main",
		BillingCurrency:     "USD",
		BillingPeriodStart:  now,
		BillingPeriodEnd:    now,
		ChargeCategory:      types.ChargeCategories.Usage,
		ChargeClass:         types.ChargeClasses.OnDemand,
		ChargeDescription:   payload,
		ChargeFrequency:     "Hourly",
		ChargePeriodStart:   now,
		ChargePeriodEnd:     now,
		ChargeSubcategory:   "General",
		EffectiveCost:       1.0,
		InvoiceIssuerName:   "Issuer",
		ListCost:            1.0,
		ListUnitPrice:       1.0,
		PricingCategory:     types.PricingCategories.Standard,
		PricingQuantity:     1.0,
		PricingUnit:         "Hours",
		ProviderName:        "test",
		PublisherName:       "pub",
		ResourceId:          "rid",
		ResourceName:        "rname",
		ResourceType:        "rtype",
		ServiceCategory:     "Compute",
		ServiceName:         "svc",
		SkuId:               "sku",
		SkuPriceId:          "sp",
		SubAccountId:        "sub",
		SubAccountName:      "subn",
		UsageQuantity:       1,
		UsageUnit:           "Hours",
		SourceProvider:      "gen",
		SourceFileName:      "gen.csv",
		ConversionTimestamp: now,
	}
	// Write batches of small size to cross the threshold several times
	for i := 0; i < 300; i++ { // 300 rows ~ 450KB payload; should produce multiple files even with compression
		if err := w.Write(context.Background(), []types.FocusRecord{rec}); err != nil {
			t.Fatalf("write: %v", err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	// Count produced parquet files and ensure no temp files remain
	re := regexp.MustCompile(`^fsynctest-\d{8}-\d{4}-\d{3}\.parquet$`)
	var files []string
	entries, err := os.ReadDir(tmp)
	if err != nil {
		t.Fatalf("readdir: %v", err)
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if filepath.Ext(e.Name()) == ".tmp" {
			t.Fatalf("unexpected lingering temp file: %s", e.Name())
		}
		if re.MatchString(e.Name()) {
			files = append(files, e.Name())
		}
	}
	if len(files) == 0 {
		t.Fatalf("expected rotated parquet files, found none")
	}
	// Expect exactly one fsync per output file (each finalize+rename and final Close)
	if got, want := int(atomic.LoadInt32(&syncCalls)), len(files); got != want {
		t.Fatalf("fsync count mismatch: got %d, want %d (one per output file)", got, want)
	}
}

// Test that when rotation is disabled, output is a single file at base path and
// our fsync hook is never invoked (no per-row fsync; no rename path).
func TestParquetWriter_FsyncCounts_NoRotation(t *testing.T) {
	tmp := t.TempDir()
	base := filepath.Join(tmp, "focus.parquet")

	opts := &types.ParquetOptions{
		CompressionCodec:  "snappy",
		RotateSizeBytes:   -1, // disable rotation explicitly
		RowGroupSizeBytes: 8 * 1024,
		PageSizeBytes:     4 * 1024,
	}
	ctx := WithParquetOptions(context.Background(), opts)

	var syncCalls int32
	orig := syncFile
	syncFile = func(f *os.File) error {
		atomic.AddInt32(&syncCalls, 1)
		return f.Sync()
	}
	t.Cleanup(func() { syncFile = orig })

	w, _, err := NewWriter(ctx, base, "parquet", types.GetFocusV12Schema())
	if err != nil {
		t.Fatalf("NewWriter: %v", err)
	}

	now := time.Now().UTC()
	rec := types.FocusRecord{BillingAccountId: "A", BillingCurrency: "USD", BillingPeriodStart: now, BillingPeriodEnd: now}
	for i := 0; i < 50; i++ {
		if err := w.Write(context.Background(), []types.FocusRecord{rec}); err != nil {
			t.Fatalf("write: %v", err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	// Ensure exactly one parquet file at base path and no rotated naming
	entries, err := os.ReadDir(tmp)
	if err != nil {
		t.Fatalf("readdir: %v", err)
	}
	var parquetCount int
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if strings.HasSuffix(e.Name(), ".parquet") {
			parquetCount++
			if e.Name() != filepath.Base(base) {
				t.Fatalf("unexpected parquet name with rotation disabled: %s", e.Name())
			}
		}
		if filepath.Ext(e.Name()) == ".tmp" {
			t.Fatalf("unexpected lingering temp file: %s", e.Name())
		}
	}
	if parquetCount != 1 {
		t.Fatalf("expected single output parquet file, found %d", parquetCount)
	}
	if c := atomic.LoadInt32(&syncCalls); c != 0 {
		t.Fatalf("unexpected fsync calls with rotation disabled: %d", c)
	}
}
