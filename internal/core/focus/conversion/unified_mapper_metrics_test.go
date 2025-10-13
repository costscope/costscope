package conversion

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	awsp "github.com/costscope/costscope/internal/core/focus/conversion/aws"
	azp "github.com/costscope/costscope/internal/core/focus/conversion/azure"
	c "github.com/costscope/costscope/internal/core/focus/conversion/common"
	gcpp "github.com/costscope/costscope/internal/core/focus/conversion/gcp"
	"github.com/costscope/costscope/internal/core/focus/types"
	"github.com/costscope/costscope/internal/core/monitoring/telemetry"

	"github.com/prometheus/client_golang/prometheus/testutil"
)

// TestUnifiedMapper_NoErrors ensures unified mapper error counter does not
// increase during minimal successful conversions (delta == 0) for each provider.
func TestUnifiedMapper_NoErrors(t *testing.T) {
	cases := []struct{ provider, header, row string }{
		{"aws", strings.Join([]string{"bill/BillingAccountId", "bill/BillingAccountName", "bill/BillingCurrency", "lineItem/UnblendedCost", "lineItem/UsageAmount", "lineItem/UsageStartDate", "lineItem/UsageEndDate", "product/ProductName", "lineItem/UsageAccountId"}, ","), strings.Join([]string{"123", "Master", "USD", "1.00", "2", "2024-01-01 00:00:00", "2024-01-01 01:00:00", "AmazonEC2", "456"}, ",")},
		{"azure", strings.Join([]string{"BillingAccountId", "BillingAccountName", "BillingCurrency", "SubscriptionId", "SubscriptionName", "ServiceName", "ServiceFamily", "ResourceId", "Quantity", "UnitOfMeasure", "AmortizedCost", "RetailPrice", "UsageStart", "UsageEnd"}, ","), strings.Join([]string{"BA-1", "Main", "USD", "sub-1", "Dev", "Virtual Machines", "Compute", "/subs/sub-1/rg/rg/vm/vm1", "1", "Hours", "0.50", "0.50", "2024-01-01T00:00:00Z", "2024-01-01T01:00:00Z"}, ",")},
		{"gcp", strings.Join([]string{"billing_account_id", "billing_account_name", "currency", "project.id", "service.description", "sku.id", "usage.amount", "usage.unit", "usage_start_time", "usage_end_time", "cost"}, ","), strings.Join([]string{"BA-9", "Main", "USD", "proj-1", "Compute Engine", "SKU-1", "1", "Hrs", "2024-01-01T00:00:00Z", "2024-01-01T01:00:00Z", "0.10"}, ",")},
	}
	for _, c := range cases {
		runUnifiedNoError(t, c.provider, c.header, c.row)
	}
}

func runUnifiedNoError(t *testing.T, provider, header, row string) {
	t.Helper()
	tmp := t.TempDir()
	in := filepath.Join(tmp, provider+".csv")
	out := filepath.Join(tmp, provider+".ndjson")
	if err := os.WriteFile(in, []byte(strings.Join([]string{header, row}, "\n")), 0o600); err != nil {
		t.Fatalf("write input: %v", err)
	}
	// local interface matching converter methods we need
	type convIface interface {
		ValidateInput(context.Context, *types.ConversionConfig) error
		ConvertStream(context.Context, *types.ConversionConfig, types.ProgressCallback) (*types.ConversionResult, error)
	}
	var conv convIface
	switch provider {
	case "aws":
		conv = awsp.NewAWSConverter()
	case "azure":
		conv = azp.NewAzureConverter()
	case "gcp":
		conv = gcpp.NewGCPConverter()
	default:
		t.Fatalf("unknown provider %s", provider)
	}
	start := testutil.ToFloat64(telemetry.UnifiedMapperErrors.WithLabelValues(provider, "unified"))
	cfg := &types.ConversionConfig{Provider: provider, InputPath: in, OutputPath: out, Streaming: true, UseUnifiedMapper: true, ChunkSize: 1000, Workers: 1, ConversionId: "metrics-" + provider}
	if err := conv.ValidateInput(context.Background(), cfg); err != nil {
		t.Fatalf("validate %s: %v", provider, err)
	}
	if _, err := conv.ConvertStream(context.Background(), cfg, nil); err != nil {
		t.Fatalf("convert %s: %v", provider, err)
	}
	end := testutil.ToFloat64(telemetry.UnifiedMapperErrors.WithLabelValues(provider, "unified"))
	if end-start != 0 {
		t.Fatalf("unified mapper errors delta for %s expected 0, got %f (start=%f end=%f)", provider, end-start, start, end)
	}
	// reference shared classification hook (no-op here) to keep coverage on the helper path
	c.ApplyUnifiedClassification(provider, map[string]string{}, &types.FocusRecord{})
}
