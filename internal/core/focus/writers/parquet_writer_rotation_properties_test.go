//go:build duckdb && !race

package writers

import (
	"context"
	"database/sql"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"testing"
	"time"

	_ "github.com/marcboeker/go-duckdb"

	"local/costscope/internal/core/focus/types"
)

// TestParquetWriter_RotationProperties validates multiple formal properties of the
// size-based rotation implementation using randomized batch sizes:
//  1. Random batched writes produce multiple rotated files with expected naming pattern.
//  2. No temporary *.tmp files are left behind after Close (atomic rename correctness).
//  3. All rotated parquet files together contain exactly the number of rows written (DuckDB read_parquet pattern).
//  4. For size rotation, every completed (non-final) file has size >= threshold.
//     (The writer triggers rotation only after size check >= threshold per row.)
func TestParquetWriter_RotationProperties(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	base := filepath.Join(tmp, "focus.parquet")

	// Small threshold to force multiple rotations quickly.
	const rotateThreshold = int64(2 * 1024) // 2KB

	// Deterministic PRNG for reproducibility while still giving variability.
	// gosec: disable G404 (math/rand is acceptable for non-crypto test data)
	r := rand.New(rand.NewSource(42)) //nolint:gosec

	opts := &types.ParquetOptions{
		CompressionCodec:  "snappy",
		RotateSizeBytes:   rotateThreshold,
		RowGroupSizeBytes: 32 * 1024, // small row groups to avoid buffering large amounts before size check
		PageSizeBytes:     8 * 1024,
		FilePrefix:        "rotprop",
	}
	ctx := WithParquetOptions(context.Background(), opts)

	wIface, _, err := NewWriter(ctx, base, "parquet", types.GetFocusV12Schema())
	if err != nil {
		t.Fatalf("NewWriter: %v", err)
	}
	w := wIface // DataWriter interface

	// Generate random records whose serialized size will exceed threshold after a few writes.
	// Large ChargeDescription + ResourceId to inflate per-row size deterministically.
	totalRecords := 0
	targetRecords := 500 + r.Intn(300) // 500-799
	bigStr := func(n int) string {
		b := make([]byte, n)
		for i := range b {
			b[i] = byte('A' + r.Intn(26))
		}
		return string(b)
	}
	// Precompute a heavy payload to re-use (less allocation noise) plus slight variance per row.
	payload := bigStr(1500) // ~1.5KB

	makeRecord := func() types.FocusRecord {
		now := time.Now().UTC()
		desc := payload + bigStr(32) // ensure some per-row variability
		return types.FocusRecord{
			BillingAccountId:    "BA",
			BillingAccountName:  "Main",
			BillingCurrency:     "USD",
			BillingPeriodStart:  now,
			BillingPeriodEnd:    now,
			ChargeCategory:      types.ChargeCategories.Usage,
			ChargeClass:         types.ChargeClasses.OnDemand,
			ChargeDescription:   desc,
			ChargeFrequency:     "Hourly",
			ChargePeriodStart:   now,
			ChargePeriodEnd:     now,
			ChargeSubcategory:   "Compute",
			EffectiveCost:       1.0,
			InvoiceIssuerName:   "Issuer",
			ListCost:            1.0,
			ListUnitPrice:       1.0,
			PricingCategory:     types.PricingCategories.Standard,
			PricingQuantity:     1.0,
			PricingUnit:         "Hours",
			ProviderName:        "test",
			PublisherName:       "pub",
			ResourceId:          "res-" + bigStr(32),
			ResourceName:        "resname",
			ResourceType:        "rtype",
			ServiceCategory:     "Compute",
			ServiceName:         "svc",
			SkuId:               "sku",
			SkuPriceId:          "skuprice",
			SubAccountId:        "suba",
			SubAccountName:      "subaccount",
			UsageQuantity:       1,
			UsageUnit:           "Hours",
			SourceProvider:      "generator",
			SourceFileName:      "gen.csv",
			ConversionTimestamp: now,
		}
	}

	for totalRecords < targetRecords {
		// Random batch size 1..25, bounded by remaining.
		batchSize := 1 + r.Intn(25)
		if totalRecords+batchSize > targetRecords {
			batchSize = targetRecords - totalRecords
		}
		batch := make([]types.FocusRecord, 0, batchSize)
		for i := 0; i < batchSize; i++ {
			_ = i // kept for clarity; index not needed inside record generator
			batch = append(batch, makeRecord())
		}
		if err := w.Write(context.Background(), batch); err != nil {
			t.Fatalf("write batch: %v", err)
		}
		totalRecords += batchSize
	}
	if err := w.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	// Gather files and check naming pattern and sizes.
	entries, err := os.ReadDir(tmp)
	if err != nil {
		t.Fatalf("readdir: %v", err)
	}
	nameRe := regexp.MustCompile(`^rotprop-\d{8}-\d{4}-\d{3}\.parquet$`)
	var parquetFiles []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if nameRe.MatchString(e.Name()) {
			parquetFiles = append(parquetFiles, filepath.Join(tmp, e.Name()))
		}
		// Property 2: no lingering temp files.
		if filepath.Ext(e.Name()) == ".tmp" {
			t.Fatalf("found unexpected temp file after close: %s", e.Name())
		}
	}
	if len(parquetFiles) < 2 { // rotation should have happened with this payload
		t.Fatalf("expected >=2 rotated parquet files, got %d", len(parquetFiles))
	}

	// Sort by name (timestamp & seq ordering) for size threshold property.
	for i := 1; i < len(parquetFiles); i++ {
		j := i
		for j > 0 && parquetFiles[j] < parquetFiles[j-1] {
			parquetFiles[j], parquetFiles[j-1] = parquetFiles[j-1], parquetFiles[j]
			j--
		}
	}
	// Property 4: all but last should be >= threshold.
	for i, f := range parquetFiles {
		fi, err := os.Stat(f)
		if err != nil {
			t.Fatalf("stat %s: %v", f, err)
		}
		if fi.Size() == 0 {
			t.Fatalf("file empty: %s", f)
		}
		if i < len(parquetFiles)-1 && fi.Size() < rotateThreshold {
			t.Fatalf("rotated file below threshold: %s size=%d < %d", f, fi.Size(), rotateThreshold)
		}
	}

	// Property 3: DuckDB can read all rows (pattern glob).
	db, err := sql.Open("duckdb", ":memory:")
	if err != nil {
		t.Fatalf("duckdb open: %v", err)
	}
	defer func() { _ = db.Close() }()
	pattern := filepath.Join(tmp, "rotprop-*.parquet")
	if _, err := db.Exec("CREATE TABLE t AS SELECT * FROM read_parquet(?)", pattern); err != nil {
		t.Fatalf("duckdb read_parquet: %v", err)
	}
	var cnt int
	if err := db.QueryRow("SELECT COUNT(*) FROM t").Scan(&cnt); err != nil {
		t.Fatalf("duckdb count: %v", err)
	}
	if cnt != totalRecords {
		t.Fatalf("row count mismatch: wrote %d, read %d", totalRecords, cnt)
	}
}
