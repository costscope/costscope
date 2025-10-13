package conversion

// Performance guard benchmark (M11 – Perf guard unified vs fast)
// This benchmark is intentionally small enough (100k synthetic AWS rows by default)
// to run quickly in CI while still reflecting relative performance differences
// between the optimized (legacy/fast) mapper path and the unified mapper path.
//
// The accompanying script scripts/perf/perf-guard.sh parses the output of:
//   go test -run '^$' -bench PerfGuardUnified -benchmem ./internal/core/focus/conversion
// and enforces the SLA ratios:
//   unified_duration / legacy_duration <= 1.20
//   unified_allocs   / legacy_allocs   <= 1.25
// (defaults overridable via PERF_GUARD_DURATION_MAX / PERF_GUARD_ALLOCS_MAX env vars)
//
// Acceptance (from __TODO.md M11): Fail when artificial slow-down (e.g., sleep) is introduced
// in unified path. The script will exit non‑zero if thresholds are exceeded.

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	awsp "github.com/costscope/costscope/internal/core/focus/conversion/aws"
	"github.com/costscope/costscope/internal/core/focus/types"
)

// generatePerfGuardSyntheticCSV creates N synthetic AWS-like CUR rows (minimal subset of fields).
// Duplicated (lightly) from unified_benchmark_test to avoid test import coupling.
func generatePerfGuardSyntheticCSV(n int) string { // nolint:revive // concise benchmark helper
	headers := []string{"bill/BillingAccountId", "bill/BillingAccountName", "bill/BillingCurrency", "lineItem/UnblendedCost", "lineItem/UsageAmount", "lineItem/UsageStartDate", "lineItem/UsageEndDate", "lineItem/LineItemDescription", "lineItem/Operation", "lineItem/UsageType", "product/ProductName", "product/ProductFamily", "lineItem/ResourceId", "pricing/PriceId", "lineItem/UsageAccountId"}
	b := &strings.Builder{}
	b.Grow(n * 96)
	b.WriteString(strings.Join(headers, ","))
	b.WriteByte('\n')
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	for i := 0; i < n; i++ {
		us := start.Add(time.Duration(i) * time.Minute).Format("2006-01-02 15:04:05")
		ue := start.Add(time.Duration(i+1) * time.Minute).Format("2006-01-02 15:04:05")
		//nolint:gosec // math/rand acceptable for synthetic data
		row := fmt.Sprintf("123456789012,Master,USD,%.4f,1,%s,%s,desc,RunInstances,USW2-BoxUsage:t3.micro,AmazonEC2,Compute,i-perf%06d,price-%d,111111111111\n", rand.Float64()*5, us, ue, i, i)
		b.WriteString(row)
	}
	return b.String()
}

// BenchmarkPerfGuardUnified executes conversion for legacy vs unified on a fixed synthetic dataset.
// Name is stable for parsing; keep sub-benchmark names exactly 'legacy' and 'unified'.
func BenchmarkPerfGuardUnified(b *testing.B) {
	// Row count can be overridden (e.g. PERF_GUARD_ROWS=200000) if needed for local tuning.
	rows := 100_000
	if v := os.Getenv("PERF_GUARD_ROWS"); v != "" {
		// Parse user override; ignore errors silently.
		if _, err := fmt.Sscan(v, &rows); err != nil || rows <= 0 {
			rows = 100_000
		}
	}
	csv := generatePerfGuardSyntheticCSV(rows)
	dir := b.TempDir()
	input := filepath.Join(dir, "perf_input.csv")
	if err := os.WriteFile(input, []byte(csv), 0o600); err != nil {
		b.Fatalf("write synthetic: %v", err)
	}
	conv := awsp.NewAWSConverter()
	fastCfg := &types.ConversionConfig{Provider: "aws", InputPath: input, OutputPath: filepath.Join(dir, "out_fast.ndjson"), Streaming: true, ChunkSize: 10000, Workers: 1, ConversionId: "perf-guard-fast", UseUnifiedMapper: false}
	uniCfg := &types.ConversionConfig{Provider: "aws", InputPath: input, OutputPath: filepath.Join(dir, "out_unified.ndjson"), Streaming: true, ChunkSize: 10000, Workers: 1, ConversionId: "perf-guard-unified", UseUnifiedMapper: true}
	if err := conv.ValidateInput(context.Background(), fastCfg); err != nil {
		b.Fatalf("validate fast: %v", err)
	}
	if err := conv.ValidateInput(context.Background(), uniCfg); err != nil {
		b.Fatalf("validate unified: %v", err)
	}

	b.ResetTimer()
	b.Run("legacy", func(sb *testing.B) {
		for i := 0; i < sb.N; i++ {
			if _, err := conv.ConvertStream(context.Background(), fastCfg, nil); err != nil {
				sb.Fatalf("convert fast: %v", err)
			}
		}
	})
	b.Run("unified", func(sb *testing.B) {
		for i := 0; i < sb.N; i++ {
			if _, err := conv.ConvertStream(context.Background(), uniCfg, nil); err != nil {
				sb.Fatalf("convert unified: %v", err)
			}
		}
	})
}
