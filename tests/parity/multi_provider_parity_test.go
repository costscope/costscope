package parity

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"

	_ "github.com/marcboeker/go-duckdb"

	conversion "github.com/costscope/costscope/internal/core/focus/conversion"
	"github.com/costscope/costscope/internal/core/focus/types"
)

// TestMultiProviderLegacyUnifiedParity: lightweight parity check between default fast path
// and unified mapper (enabled via env variable) using only the public Convert() API.
// It writes tiny synthetic CSV inputs for each provider, runs conversion twice (fast/unified)
// and compares aggregate invariants: record count (stubbed), hashes (lite), and placeholder sums.
// The test now opens the produced Parquet via an in‑memory DuckDB connection to compute
// deterministic parity invariants (record count, sums, order‑independent lite hash).
func TestMultiProviderLegacyUnifiedParity(t *testing.T) { //nolint:tparallel
	t.Parallel()

	providerData := map[string]struct {
		header []string
		rows   [][]string
	}{
		"aws": {[]string{
			"bill/BillingAccountId",
			"lineItem/UnblendedCost",
			"lineItem/UsageAmount",
			"product/ProductName",
			"lineItem/ResourceId",
			"lineItem/UsageStartDate",
			"lineItem/UsageEndDate",
			"lineItem/UsageAccountId",
		}, [][]string{{
			"123456789012",
			"1.23",
			"10",
			"AmazonEC2",
			"i-abc",
			"2025-08-01T00:00:00Z",
			"2025-08-01T01:00:00Z",
			"123456789012",
		}, {
			"123456789012",
			"0.77",
			"5",
			"AmazonS3",
			"bucket/foo",
			"2025-08-01T00:00:00Z",
			"2025-08-01T01:00:00Z",
			"123456789012",
		}}},
		"azure": {[]string{"BillingAccountId", "Cost", "UsageQuantity", "ServiceName", "ResourceId"}, [][]string{{"az-acc", "2.50", "3", "Azure VM", "/subscriptions/xxx/resourceGroups/rg/providers/Microsoft.Compute/virtualMachines/vm1"}}},
		"gcp":   {[]string{"billing_account_id", "cost", "usage_amount", "service_description", "resource_name"}, [][]string{{"AAAA-BBBB-CCCC", "5.00", "100", "Compute Engine", "projects/p/instances/i1"}}},
	}

	tmpDir := t.TempDir()
	cm := conversion.NewConversionManager(2)
	uc := cm.GetConverter()

	for provider, data := range providerData {
		provider, data := provider, data
		t.Run(provider, func(t *testing.T) {
			t.Parallel()
			inputPath := filepath.Join(tmpDir, fmt.Sprintf("%s.csv", provider))
			f, err := os.Create(inputPath) // #nosec G304 -- path derived from t.TempDir(), not user input
			if err != nil {
				t.Fatalf("create csv: %v", err)
			}
			w := csv.NewWriter(f)
			if err := w.Write(data.header); err != nil {
				t.Fatalf("write header: %v", err)
			}
			for _, r := range data.rows {
				if err := w.Write(r); err != nil {
					t.Fatalf("write row: %v", err)
				}
			}
			w.Flush()
			if err := f.Close(); err != nil {
				t.Fatalf("close csv file: %v", err)
			}

			run := func(unified bool) (*types.ConversionResult, parityStats) {
				outPath := filepath.Join(tmpDir, fmt.Sprintf("%s_%v_%d.parquet", provider, unified, time.Now().UnixNano()))
				if unified {
					if err := os.Setenv("COSTSCOPE_USE_UNIFIED_MAPPER", "true"); err != nil {
						t.Fatalf("set unified env: %v", err)
					}
				} else {
					if err := os.Unsetenv("COSTSCOPE_USE_UNIFIED_MAPPER"); err != nil {
						t.Fatalf("unset unified env: %v", err)
					}
				}
				cfg := &types.ConversionConfig{Provider: provider, InputPath: inputPath, OutputPath: outPath, Streaming: true, ValidateOutput: false,
					Parquet: types.ParquetOptions{RotateSizeBytes: -1, CompressionCodec: "snappy"}, OutputFormat: "parquet"}
				res, err := uc.Convert(context.Background(), cfg)
				if err != nil {
					t.Fatalf("convert unified=%v: %v", unified, err)
				}
				ps, err := extractParityStats(res)
				if err != nil {
					t.Fatalf("stats: %v", err)
				}
				return res, ps
			}

			fastRes, fastStats := run(false)
			unifiedRes, unifiedStats := run(true)

			if fastStats.count != unifiedStats.count {
				t.Errorf("record count mismatch fast=%d unified=%d", fastStats.count, unifiedStats.count)
			}
			if fastStats.hash != unifiedStats.hash {
				// Include provider context for easier triage.
				t.Errorf("lite hash mismatch provider=%s fast=%s unified=%s", provider, fastStats.hash, unifiedStats.hash)
			}
			if fastStats.effectiveCostSum != unifiedStats.effectiveCostSum {
				t.Errorf("effective_cost sum mismatch fast=%f unified=%f", fastStats.effectiveCostSum, unifiedStats.effectiveCostSum)
			}
			if fastStats.usageQtySum != unifiedStats.usageQtySum {
				t.Errorf("usage_quantity sum mismatch fast=%f unified=%f", fastStats.usageQtySum, unifiedStats.usageQtySum)
			}
			// Service distribution parity (per service_name)
			fastSvc, err := loadServiceDistributions(fastRes.OutputFile)
			if err != nil {
				t.Fatalf("service dist fast: %v", err)
			}
			unifiedSvc, err := loadServiceDistributions(unifiedRes.OutputFile)
			if err != nil {
				t.Fatalf("service dist unified: %v", err)
			}
			if len(fastSvc) != len(unifiedSvc) {
				t.Errorf("service set size mismatch fast=%d unified=%d", len(fastSvc), len(unifiedSvc))
			}
			for svc, fvals := range fastSvc {
				uvals, ok := unifiedSvc[svc]
				if !ok {
					t.Errorf("missing service in unified: %s", svc)
					continue
				}
				if fvals.count != uvals.count {
					t.Errorf("svc=%s count mismatch fast=%d unified=%d", svc, fvals.count, uvals.count)
				}
				if fvals.costSum != uvals.costSum {
					t.Errorf("svc=%s cost sum mismatch fast=%f unified=%f", svc, fvals.costSum, uvals.costSum)
				}
				if fvals.usageSum != uvals.usageSum {
					t.Errorf("svc=%s usage sum mismatch fast=%f unified=%f", svc, fvals.usageSum, uvals.usageSum)
				}
			}
		})
	}
}

type parityStats struct {
	count            int
	effectiveCostSum float64
	usageQtySum      float64
	hash             string
}

func extractParityStats(res *types.ConversionResult) (parityStats, error) {
	// Open ephemeral DuckDB (in-memory). We rely on embedded duckdb go driver already in go.mod.
	db, err := sql.Open("duckdb", "")
	if err != nil {
		return parityStats{}, fmt.Errorf("open duckdb: %w", err)
	}
	defer func() { _ = db.Close() }()

	// Aggregate sums + count.
	row := db.QueryRow("SELECT coalesce(sum(effective_cost),0), coalesce(sum(usage_quantity),0), count(*) FROM read_parquet(?)", res.OutputFile)
	var effSum, usageSum float64
	var cnt int
	if err := row.Scan(&effSum, &usageSum, &cnt); err != nil {
		return parityStats{}, fmt.Errorf("scan aggregates: %w", err)
	}

	// Load lite hash input columns; order independent hash to avoid false negatives if row order diverges.
	rows, err := db.Query("SELECT effective_cost, usage_quantity, coalesce(provider_name,''), coalesce(service_name,''), coalesce(charge_category,'') FROM read_parquet(?)", res.OutputFile)
	if err != nil {
		return parityStats{}, fmt.Errorf("query rows: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()
	type rec struct{ s string }
	var recs []rec
	for rows.Next() {
		var ec, uq float64
		var provider, service, category string
		if err := rows.Scan(&ec, &uq, &provider, &service, &category); err != nil {
			return parityStats{}, fmt.Errorf("scan row: %w", err)
		}
		recs = append(recs, rec{fmt.Sprintf("%.6f|%.6f|%s|%s|%s", ec, uq, provider, service, category)})
	}
	if err := rows.Err(); err != nil {
		return parityStats{}, fmt.Errorf("iterate rows: %w", err)
	}
	sort.Slice(recs, func(i, j int) bool { return recs[i].s < recs[j].s })
	h := sha256.New()
	for _, r := range recs {
		_, _ = h.Write([]byte(r.s))
		_, _ = h.Write([]byte{'\n'})
	}
	sum := h.Sum(nil)
	lite := fmt.Sprintf("%x", sum[:8])

	return parityStats{count: cnt, effectiveCostSum: effSum, usageQtySum: usageSum, hash: lite}, nil
}

// (blank import above ensures duckdb driver registration)

type serviceDist struct {
	count    int
	costSum  float64
	usageSum float64
}

func loadServiceDistributions(parquetPath string) (map[string]serviceDist, error) {
	db, err := sql.Open("duckdb", "")
	if err != nil {
		return nil, fmt.Errorf("open duckdb: %w", err)
	}
	defer func() { _ = db.Close() }()
	rows, err := db.Query(`SELECT coalesce(service_name,''), count(*), coalesce(sum(effective_cost),0), coalesce(sum(usage_quantity),0) FROM read_parquet(?) GROUP BY 1`, parquetPath)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	defer func() { _ = rows.Close() }()
	res := make(map[string]serviceDist)
	for rows.Next() {
		var svc string
		var cnt int
		var csum, usum float64
		if err := rows.Scan(&svc, &cnt, &csum, &usum); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		res[svc] = serviceDist{count: cnt, costSum: csum, usageSum: usum}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate: %w", err)
	}
	return res, nil
}
