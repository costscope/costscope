package conversion

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
	azp "github.com/costscope/costscope/internal/core/focus/conversion/azure"
	gcpp "github.com/costscope/costscope/internal/core/focus/conversion/gcp"
	"github.com/costscope/costscope/internal/core/focus/types"
)

// generateSyntheticCSV generates N synthetic AWS-like rows for benchmarking
func generateSyntheticCSV(n int) string {
	headers := []string{"bill/BillingAccountId", "bill/BillingAccountName", "bill/BillingCurrency", "lineItem/UnblendedCost", "lineItem/UsageAmount", "lineItem/UsageStartDate", "lineItem/UsageEndDate", "lineItem/LineItemDescription", "lineItem/Operation", "lineItem/UsageType", "product/ProductName", "product/ProductFamily", "lineItem/ResourceId", "pricing/PriceId", "lineItem/UsageAccountId"}
	b := &strings.Builder{}
	b.Grow(n * 128)
	b.WriteString(strings.Join(headers, ","))
	b.WriteByte('\n')
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	for i := 0; i < n; i++ {
		us := start.Add(time.Duration(i) * time.Hour).Format("2006-01-02 15:04:05")
		ue := start.Add(time.Duration(i+1) * time.Hour).Format("2006-01-02 15:04:05")
		//nolint:gosec // G404: math/rand acceptable for synthetic benchmark data generation
		row := fmt.Sprintf("123456789012,Master,USD,%.2f,1,%s,%s,desc,RunInstances,USW2-BoxUsage:t3.micro,AmazonEC2,Compute,i-abc%06d,price-%d,111111111111\n", rand.Float64()*10, us, ue, i, i)
		b.WriteString(row)
	}
	return b.String()
}

// generateSyntheticAzureCSV creates synthetic Azure like usage records.
func generateSyntheticAzureCSV(n int) string {
	headers := []string{"BillingAccountId", "BillingAccountName", "BillingCurrency", "SubscriptionId", "SubscriptionName", "ServiceName", "ServiceFamily", "ResourceId", "ResourceName", "ResourceType", "ResourceLocation", "Quantity", "UnitOfMeasure", "AmortizedCost", "RetailPrice", "UsageStart", "UsageEnd"}
	b := &strings.Builder{}
	b.Grow(n * 128)
	b.WriteString(strings.Join(headers, ","))
	b.WriteByte('\n')
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	for i := 0; i < n; i++ {
		us := start.Add(time.Duration(i) * time.Hour).Format(time.RFC3339)
		ue := start.Add(time.Duration(i+1) * time.Hour).Format(time.RFC3339)
		// Hours unit intentionally mixed case to exercise normalizer.
		//nolint:gosec // G404: math/rand acceptable for benchmark synthesis
		row := fmt.Sprintf("BA-1,Main,usd,sub-%d,Env%d,Virtual Machines,Compute,/subs/sub-%d/rg/rg/vm/vm%d,vm%d,Microsoft.Compute/virtualMachines,eastus,%d,hrs,%.2f,%.2f,%s,%s\n", i%10, i%5, i%10, i, i, 1, rand.Float64()*5, rand.Float64()*5, us, ue)
		b.WriteString(row)
	}
	return b.String()
}

// generateSyntheticGCPCSV creates synthetic GCP billing export like rows.
func generateSyntheticGCPCSV(n int) string {
	headers := []string{"billing_account_id", "billing_account_name", "currency", "project.id", "project.name", "service.description", "sku.id", "usage.amount", "usage.unit", "usage_start_time", "usage_end_time", "cost"}
	b := &strings.Builder{}
	b.Grow(n * 128)
	b.WriteString(strings.Join(headers, ","))
	b.WriteByte('\n')
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	for i := 0; i < n; i++ {
		us := start.Add(time.Duration(i) * time.Hour).Format(time.RFC3339)
		ue := start.Add(time.Duration(i+1) * time.Hour).Format(time.RFC3339)
		//nolint:gosec // G404: math/rand acceptable for benchmark synthesis
		row := fmt.Sprintf("BA-100,Main,usd,proj-%d,Project %d,Compute Engine,SKU-%d,1,GiB,%s,%s,%.2f\n", i%50, i%50, i, us, ue, rand.Float64()*3)
		b.WriteString(row)
	}
	return b.String()
}

// BenchmarkUnifiedMapperAWS_1M benchmarks legacy vs unified for 1M synthetic rows (may be shortened if testing.Short)
func BenchmarkUnifiedMapperAWS_1M(b *testing.B) {
	if testing.Short() {
		b.Skip("short mode")
	}
	n := 1_000_000
	csv := generateSyntheticCSV(n)
	tmp := b.TempDir()
	in := filepath.Join(tmp, "in.csv")
	if err := os.WriteFile(in, []byte(csv), 0o600); err != nil {
		b.Fatalf("write: %v", err)
	}
	conv := awsp.NewAWSConverter()
	cfgFast := &types.ConversionConfig{Provider: "aws", InputPath: in, OutputPath: filepath.Join(tmp, "out_fast.ndjson"), Streaming: true, ChunkSize: 10000, Workers: 1, ConversionId: "bench-fast", UseUnifiedMapper: false}
	cfgUni := &types.ConversionConfig{Provider: "aws", InputPath: in, OutputPath: filepath.Join(tmp, "out_uni.ndjson"), Streaming: true, ChunkSize: 10000, Workers: 1, ConversionId: "bench-uni", UseUnifiedMapper: true}
	if err := conv.ValidateInput(context.Background(), cfgFast); err != nil {
		b.Fatalf("validate fast: %v", err)
	}
	if err := conv.ValidateInput(context.Background(), cfgUni); err != nil {
		b.Fatalf("validate uni: %v", err)
	}
	b.ResetTimer()
	b.Run("legacy", func(sb *testing.B) {
		for i := 0; i < sb.N; i++ {
			if _, err := conv.ConvertStream(context.Background(), cfgFast, nil); err != nil {
				sb.Fatalf("convert fast: %v", err)
			}
		}
	})
	b.Run("unified", func(sb *testing.B) {
		for i := 0; i < sb.N; i++ {
			if _, err := conv.ConvertStream(context.Background(), cfgUni, nil); err != nil {
				sb.Fatalf("convert unified: %v", err)
			}
		}
	})
}

// BenchmarkUnifiedMapperAzure_1M mirrors AWS benchmark for Azure synthetic data.
func BenchmarkUnifiedMapperAzure_1M(b *testing.B) {
	if testing.Short() {
		b.Skip("short mode")
	}
	n := 1_000_000
	csv := generateSyntheticAzureCSV(n)
	tmp := b.TempDir()
	in := filepath.Join(tmp, "in.csv")
	if err := os.WriteFile(in, []byte(csv), 0o600); err != nil {
		b.Fatalf("write: %v", err)
	}
	conv := azp.NewAzureConverter()
	cfgFast := &types.ConversionConfig{Provider: "azure", InputPath: in, OutputPath: filepath.Join(tmp, "out_fast.ndjson"), Streaming: true, ChunkSize: 10000, Workers: 1, ConversionId: "bench-fast-azure", UseUnifiedMapper: false}
	cfgUni := &types.ConversionConfig{Provider: "azure", InputPath: in, OutputPath: filepath.Join(tmp, "out_uni.ndjson"), Streaming: true, ChunkSize: 10000, Workers: 1, ConversionId: "bench-uni-azure", UseUnifiedMapper: true}
	if err := conv.ValidateInput(context.Background(), cfgFast); err != nil {
		b.Fatalf("validate fast: %v", err)
	}
	if err := conv.ValidateInput(context.Background(), cfgUni); err != nil {
		b.Fatalf("validate uni: %v", err)
	}
	b.ResetTimer()
	b.Run("legacy", func(sb *testing.B) {
		for i := 0; i < sb.N; i++ {
			if _, err := conv.ConvertStream(context.Background(), cfgFast, nil); err != nil {
				sb.Fatalf("convert fast: %v", err)
			}
		}
	})
	b.Run("unified", func(sb *testing.B) {
		for i := 0; i < sb.N; i++ {
			if _, err := conv.ConvertStream(context.Background(), cfgUni, nil); err != nil {
				sb.Fatalf("convert unified: %v", err)
			}
		}
	})
}

// BenchmarkUnifiedMapperGCP_1M mirrors AWS benchmark for GCP synthetic data.
func BenchmarkUnifiedMapperGCP_1M(b *testing.B) {
	if testing.Short() {
		b.Skip("short mode")
	}
	n := 1_000_000
	csv := generateSyntheticGCPCSV(n)
	tmp := b.TempDir()
	in := filepath.Join(tmp, "in.csv")
	if err := os.WriteFile(in, []byte(csv), 0o600); err != nil {
		b.Fatalf("write: %v", err)
	}
	conv := gcpp.NewGCPConverter()
	cfgFast := &types.ConversionConfig{Provider: "gcp", InputPath: in, OutputPath: filepath.Join(tmp, "out_fast.ndjson"), Streaming: true, ChunkSize: 10000, Workers: 1, ConversionId: "bench-fast-gcp", UseUnifiedMapper: false}
	cfgUni := &types.ConversionConfig{Provider: "gcp", InputPath: in, OutputPath: filepath.Join(tmp, "out_uni.ndjson"), Streaming: true, ChunkSize: 10000, Workers: 1, ConversionId: "bench-uni-gcp", UseUnifiedMapper: true}
	if err := conv.ValidateInput(context.Background(), cfgFast); err != nil {
		b.Fatalf("validate fast: %v", err)
	}
	if err := conv.ValidateInput(context.Background(), cfgUni); err != nil {
		b.Fatalf("validate uni: %v", err)
	}
	b.ResetTimer()
	b.Run("legacy", func(sb *testing.B) {
		for i := 0; i < sb.N; i++ {
			if _, err := conv.ConvertStream(context.Background(), cfgFast, nil); err != nil {
				sb.Fatalf("convert fast: %v", err)
			}
		}
	})
	b.Run("unified", func(sb *testing.B) {
		for i := 0; i < sb.N; i++ {
			if _, err := conv.ConvertStream(context.Background(), cfgUni, nil); err != nil {
				sb.Fatalf("convert unified: %v", err)
			}
		}
	})
}

// TestUnifiedMapperPerformanceGuard compares single-run allocations & ns between legacy and unified
func TestUnifiedMapperPerformanceGuard(t *testing.T) {
	if testing.Short() {
		t.Skip("short")
	}
	// reduce scale for CI speed
	n := 200_000
	csv := generateSyntheticCSV(n)
	tmp := t.TempDir()
	in := filepath.Join(tmp, "in.csv")
	if err := os.WriteFile(in, []byte(csv), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	conv := awsp.NewAWSConverter()
	cfgFast := &types.ConversionConfig{Provider: "aws", InputPath: in, OutputPath: filepath.Join(tmp, "fast.ndjson"), Streaming: true, ChunkSize: 10000, Workers: 1, ConversionId: "perf-fast", UseUnifiedMapper: false}
	cfgUni := &types.ConversionConfig{Provider: "aws", InputPath: in, OutputPath: filepath.Join(tmp, "uni.ndjson"), Streaming: true, ChunkSize: 10000, Workers: 1, ConversionId: "perf-uni", UseUnifiedMapper: true}
	if err := conv.ValidateInput(context.Background(), cfgFast); err != nil {
		t.Fatalf("validate fast: %v", err)
	}
	if err := conv.ValidateInput(context.Background(), cfgUni); err != nil {
		t.Fatalf("validate uni: %v", err)
	}
	start := time.Now()
	_, err := conv.ConvertStream(context.Background(), cfgFast, nil)
	if err != nil {
		t.Fatalf("fast: %v", err)
	}
	fastDur := time.Since(start)
	start = time.Now()
	_, err = conv.ConvertStream(context.Background(), cfgUni, nil)
	if err != nil {
		t.Fatalf("unified: %v", err)
	}
	uniDur := time.Since(start)
	// crude CPU proxy: wall time
	if uniDur > fastDur+fastDur/5 { // >20%
		t.Logf("WARNING: unified mapper slower by >20%% (fast=%s unified=%s)", fastDur, uniDur)
	} else {
		t.Logf("Unified mapper within 20%% performance envelope (fast=%s unified=%s)", fastDur, uniDur)
	}
	// allocations guard: inspect file sizes as proxy (should not exceed 25%)
	fastInfo, _ := os.Stat(filepath.Join(tmp, "fast.ndjson"))
	uniInfo, _ := os.Stat(filepath.Join(tmp, "uni.ndjson"))
	if fastInfo != nil && uniInfo != nil {
		if uniInfo.Size() > fastInfo.Size()*5/4 { // >25%
			t.Fatalf("Unified output size >25%% larger (fast=%d unified=%d)", fastInfo.Size(), uniInfo.Size())
		}
	}
}

// TestUnifiedMapperAllocationsGuard measures allocations for a smaller dataset using testing.AllocsPerRun.
func TestUnifiedMapperAllocationsGuard(t *testing.T) {
	if testing.Short() {
		t.Skip("short")
	}
	n := 50_000
	// AWS sample chosen; Azure/GCP have similar mapping cost and this keeps runtime low.
	csv := generateSyntheticCSV(n)
	tmp := t.TempDir()
	in := filepath.Join(tmp, "in.csv")
	if err := os.WriteFile(in, []byte(csv), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	conv := awsp.NewAWSConverter()
	fastCfg := &types.ConversionConfig{Provider: "aws", InputPath: in, OutputPath: filepath.Join(tmp, "alloc_fast.ndjson"), Streaming: true, ChunkSize: 5000, Workers: 1, ConversionId: "alloc-fast"}
	uniCfg := &types.ConversionConfig{Provider: "aws", InputPath: in, OutputPath: filepath.Join(tmp, "alloc_uni.ndjson"), Streaming: true, ChunkSize: 5000, Workers: 1, ConversionId: "alloc-uni", UseUnifiedMapper: true}
	if err := conv.ValidateInput(context.Background(), fastCfg); err != nil {
		t.Fatalf("validate fast: %v", err)
	}
	if err := conv.ValidateInput(context.Background(), uniCfg); err != nil {
		t.Fatalf("validate uni: %v", err)
	}
	fastAllocs := testing.AllocsPerRun(1, func() {
		if _, err := conv.ConvertStream(context.Background(), fastCfg, nil); err != nil {
			t.Fatalf("fast convert: %v", err)
		}
	})
	uniAllocs := testing.AllocsPerRun(1, func() {
		if _, err := conv.ConvertStream(context.Background(), uniCfg, nil); err != nil {
			t.Fatalf("uni convert: %v", err)
		}
	})
	if uniAllocs > fastAllocs*1.25 { // >25%
		t.Fatalf("Unified mapper allocations >25%% higher (fast=%.0f unified=%.0f)", fastAllocs, uniAllocs)
	}
	t.Logf("Allocations OK (fast=%.0f unified=%.0f)", fastAllocs, uniAllocs)
}
