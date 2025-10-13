//go:build duckdb && !race

package writers

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"testing"
	"time"

	_ "github.com/marcboeker/go-duckdb"
	"pgregory.net/rapid"

	"github.com/costscope/costscope/internal/core/focus/types"
)

// TestParquetWriter_RotationRapidProperties performs property-based randomized testing of size rotation.
// Properties (probabilistic):
//  1. Sum of row counts across rotated files == total rows written (DuckDB read_parquet glob).
//  2. No lingering temporary .tmp files after Close (atomic rotation correctness).
//  3. Every non-final rotated file size >= threshold (rotation triggers only after >= threshold).
//  4. (Conditional) When the computed row group size <= threshold, every non-final rotated file size
//     is <= threshold + rowGroupSize*upperMult (bounds overshoot caused by row group flush).
//  5. File naming matches pattern <prefix>-YYYYMMDD-HHMM-<seq>.parquet.
//
// The test randomizes: rotation threshold (4–64KB), row group size (derived <= threshold), number of batches (5–25),
// batch sizes (1–30), and payload sizes. All failure messages include the RNG seed for reproduction.
func TestParquetWriter_RotationRapidProperties(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		seed := time.Now().UnixNano()
		r := rand.New(rand.NewSource(seed)) //nolint:gosec // test RNG

		tmp, err := os.MkdirTemp("", "rot-prop-")
		if err != nil {
			t.Fatalf("[seed=%d] tmpdir: %v", seed, err)
		}
		t.Cleanup(func() { _ = os.RemoveAll(tmp) })
		base := filepath.Join(tmp, "focus.parquet")

		rotateThreshold := rapid.Int64Range(4*1024, 64*1024).Draw(rt, "rotate_threshold") // 4–64KB
		// Row group size chosen so it does not exceed threshold (to bound overshoot) but at least 4KB.
		divisor := rapid.Int64Range(1, 2).Draw(rt, "rowgroup_div") // 1 or 2
		rowGroupSize := rotateThreshold / divisor
		if rowGroupSize < 4*1024 {
			rowGroupSize = 4 * 1024
		}
		if rowGroupSize > rotateThreshold {
			rowGroupSize = rotateThreshold
		}
		batches := rapid.IntRange(5, 25).Draw(rt, "batches")
		prefix := fmt.Sprintf("prop%02d", batches)

		opts := &types.ParquetOptions{
			CompressionCodec:  "snappy",
			RotateSizeBytes:   rotateThreshold,
			RowGroupSizeBytes: rowGroupSize,
			PageSizeBytes:     8 * 1024,
			FilePrefix:        prefix,
		}
		ctx := WithParquetOptions(context.Background(), opts)
		wIface, _, err := NewWriter(ctx, base, "parquet", types.GetFocusV12Schema())
		if err != nil {
			t.Fatalf("[seed=%d] NewWriter: %v", seed, err)
		}
		w := wIface

		total := 0
		for b := 0; b < batches; b++ {
			batchSize := rapid.IntRange(1, 30).Draw(rt, fmt.Sprintf("batch_size_%d", b))
			recs := make([]types.FocusRecord, 0, batchSize)
			for i := 0; i < batchSize; i++ {
				payLen := 400 + r.Intn(1601) // 400..2000
				payload := randAlpha(r, payLen)
				now := time.UnixMilli(time.Now().UnixMilli()).UTC()
				recs = append(recs, types.FocusRecord{
					BillingAccountId: "A", BillingAccountName: "Acct", BillingCurrency: "USD",
					BillingPeriodStart: now, BillingPeriodEnd: now,
					ChargeCategory: types.ChargeCategories.Usage, ChargeClass: types.ChargeClasses.OnDemand,
					ChargeDescription: payload, ChargeFrequency: "Hourly", ChargePeriodStart: now, ChargePeriodEnd: now,
					ChargeSubcategory: "General", EffectiveCost: 1.0, InvoiceIssuerName: "Issuer", ListCost: 1.0, ListUnitPrice: 1.0,
					PricingCategory: types.PricingCategories.Standard, PricingQuantity: 1.0, PricingUnit: "Hours",
					ProviderName: "test", PublisherName: "pub", ResourceId: "res-" + randAlpha(r, 16), ResourceName: "rname", ResourceType: "rtype",
					ServiceCategory: "Compute", ServiceName: "svc", SkuId: "sku", SkuPriceId: "spid",
					SubAccountId: "sub", SubAccountName: "subname", UsageQuantity: 1, UsageUnit: "Hours",
					SourceProvider: "gen", SourceFileName: "gen.csv", ConversionTimestamp: now,
				})
			}
			if err := w.Write(context.Background(), recs); err != nil {
				// include seed and parameters
				t.Fatalf("[seed=%d thr=%d rg=%d] write batch %d: %v", seed, rotateThreshold, rowGroupSize, b, err)
			}
			total += batchSize
		}
		if err := w.Close(); err != nil {
			t.Fatalf("[seed=%d thr=%d rg=%d] close: %v", seed, rotateThreshold, rowGroupSize, err)
		}

		entries, err := os.ReadDir(tmp)
		if err != nil {
			t.Fatalf("[seed=%d] readdir: %v", seed, err)
		}
		nameRe := regexp.MustCompile(fmt.Sprintf("^%s-\\d{8}-\\d{4}-\\d{3}\\.parquet$", regexp.QuoteMeta(prefix)))
		var files []string
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			if filepath.Ext(e.Name()) == ".tmp" {
				t.Fatalf("[seed=%d thr=%d rg=%d] unexpected lingering temp file: %s", seed, rotateThreshold, rowGroupSize, e.Name())
			}
			if nameRe.MatchString(e.Name()) {
				files = append(files, filepath.Join(tmp, e.Name()))
			}
		}
		if len(files) < 1 {
			t.Fatalf("[seed=%d thr=%d rg=%d] expected at least 1 parquet file, got 0", seed, rotateThreshold, rowGroupSize)
		}
		for i := 1; i < len(files); i++ {
			j := i
			for j > 0 && files[j] < files[j-1] {
				files[j], files[j-1] = files[j-1], files[j]
				j--
			}
		}

		upperMult := int64(4)
		enforceUpper := rowGroupSize <= rotateThreshold
		for i, f := range files {
			st, err := os.Stat(f)
			if err != nil {
				t.Fatalf("[seed=%d thr=%d rg=%d] stat %s: %v", seed, rotateThreshold, rowGroupSize, f, err)
			}
			if st.Size() == 0 {
				t.Fatalf("[seed=%d thr=%d rg=%d] file empty: %s", seed, rotateThreshold, rowGroupSize, f)
			}
			if i < len(files)-1 { // non-final
				if st.Size() < rotateThreshold {
					t.Fatalf("[seed=%d thr=%d rg=%d] rotated file below threshold: %s size=%d < %d", seed, rotateThreshold, rowGroupSize, f, st.Size(), rotateThreshold)
				}
				if enforceUpper {
					upper := rotateThreshold + rowGroupSize*upperMult
					if st.Size() > upper {
						t.Fatalf("[seed=%d thr=%d rg=%d] rotated file overshoot: %s size=%d > upper=%d (thr+rg*%d)", seed, rotateThreshold, rowGroupSize, f, st.Size(), upper, upperMult)
					}
				}
			}
		}

		db, err := sql.Open("duckdb", ":memory:")
		if err != nil {
			t.Fatalf("[seed=%d] duckdb open: %v", seed, err)
		}
		defer func() { _ = db.Close() }()
		glob := filepath.Join(tmp, prefix+"-*.parquet")
		if _, err := db.Exec("CREATE TABLE t AS SELECT * FROM read_parquet(?)", glob); err != nil {
			t.Fatalf("[seed=%d thr=%d rg=%d] duckdb read_parquet glob=%s: %v", seed, rotateThreshold, rowGroupSize, glob, err)
		}
		var cnt int
		if err := db.QueryRow("SELECT COUNT(*) FROM t").Scan(&cnt); err != nil {
			t.Fatalf("[seed=%d thr=%d rg=%d] duckdb count: %v", seed, rotateThreshold, rowGroupSize, err)
		}
		if cnt != total {
			t.Fatalf("[seed=%d thr=%d rg=%d] row count mismatch: wrote %d, read %d", seed, rotateThreshold, rowGroupSize, total, cnt)
		}
	})
}

// randAlpha returns a random alphabetic string of length n using provided RNG.
func randAlpha(r *rand.Rand, n int) string { //nolint:revive // helper fine for tests
	b := make([]byte, n)
	for i := range b {
		b[i] = byte('a' + r.Intn(26))
	}
	return string(b)
}
