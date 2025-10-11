package aws

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	dto "github.com/prometheus/client_model/go"

	"local/costscope/internal/core/focus/types"
	"local/costscope/internal/core/monitoring/telemetry"
)

func getAWSCount(path, decision string) float64 {
	m := &dto.Metric{}
	_ = telemetry.ClassifierDecisions.WithLabelValues("aws", path, decision).Write(m)
	if m.GetCounter() == nil {
		return 0
	}
	return m.GetCounter().GetValue()
}

// TestAWS_ClassifierDecisionMetric_Legacy ensures metrics increment on the legacy fast path.
func TestAWS_ClassifierDecisionMetric_Legacy(t *testing.T) {
	header := strings.Join([]string{
		"bill/BillingAccountId",
		"bill/BillingAccountName",
		"bill/BillingCurrency",
		"lineItem/UnblendedCost",
		"lineItem/UsageAmount",
		"lineItem/UsageStartDate",
		"lineItem/UsageEndDate",
		"lineItem/LineItemDescription",
		"lineItem/Operation",
		"lineItem/UsageType",
		"product/ProductName",
		"product/ProductFamily",
		"lineItem/LineItemType",
		"lineItem/ResourceId",
		"pricing/PriceId",
		"lineItem/UsageAccountId",
	}, ",")

	// Two rows: positive cost (Usage) and negative cost (Credit)
	rowUsage := strings.Join([]string{
		"123456789012", "Master", "USD", "1.00", "1", "2024-01-01 00:00:00", "2024-01-01 01:00:00", "Desc", "RunInstances", "USW2-BoxUsage:t3.micro", "AmazonEC2", "Compute", "Usage", "i-1", "price-1", "111111111111",
	}, ",")
	rowCredit := strings.Join([]string{
		"123456789012", "Master", "USD", "-0.50", "0", "2024-01-01 00:00:00", "2024-01-01 01:00:00", "Desc", "RunInstances", "USW2-BoxUsage:t3.micro", "AmazonEC2", "Compute", "Credit", "i-1", "price-1", "111111111111",
	}, ",")
	csv := strings.Join([]string{header, rowUsage, rowCredit}, "\n")

	tmp := t.TempDir()
	in := filepath.Join(tmp, "in.csv")
	out := filepath.Join(tmp, "out.ndjson")
	if err := os.WriteFile(in, []byte(csv), 0o600); err != nil {
		t.Fatalf("write input: %v", err)
	}

	preUsage := getAWSCount("legacy", "Usage")
	preCredit := getAWSCount("legacy", "Credit")

	conv := NewAWSConverter()
	cfg := &types.ConversionConfig{Provider: "aws", InputPath: in, OutputPath: out, Streaming: true}
	if err := conv.ValidateInput(context.Background(), cfg); err != nil {
		t.Fatalf("validate: %v", err)
	}
	if _, err := conv.ConvertStream(context.Background(), cfg, nil); err != nil {
		t.Fatalf("convert: %v", err)
	}

	postUsage := getAWSCount("legacy", "Usage")
	postCredit := getAWSCount("legacy", "Credit")
	if int(postUsage-preUsage) < 1 {
		t.Fatalf("expected legacy Usage increment, delta=%.0f", postUsage-preUsage)
	}
	if int(postCredit-preCredit) < 1 {
		t.Fatalf("expected legacy Credit increment, delta=%.0f", postCredit-preCredit)
	}
}

// TestAWS_ClassifierDecisionMetric_Unified ensures metrics increment on the unified path too.
func TestAWS_ClassifierDecisionMetric_Unified(t *testing.T) {
	header := strings.Join([]string{
		"bill/BillingAccountId",
		"bill/BillingAccountName",
		"bill/BillingCurrency",
		"lineItem/UnblendedCost",
		"lineItem/UsageAmount",
		"lineItem/UsageStartDate",
		"lineItem/UsageEndDate",
		"lineItem/LineItemDescription",
		"lineItem/Operation",
		"lineItem/UsageType",
		"product/ProductName",
		"product/ProductFamily",
		"lineItem/LineItemType",
		"lineItem/ResourceId",
		"pricing/PriceId",
		"lineItem/UsageAccountId",
	}, ",")

	rowUsage := strings.Join([]string{
		"123456789012", "Master", "USD", "1.00", "1", "2024-01-01 00:00:00", "2024-01-01 01:00:00", "Desc", "RunInstances", "USW2-BoxUsage:t3.micro", "AmazonEC2", "Compute", "Usage", "i-1", "price-1", "111111111111",
	}, ",")
	rowCredit := strings.Join([]string{
		"123456789012", "Master", "USD", "-0.50", "0", "2024-01-01 00:00:00", "2024-01-01 01:00:00", "Desc", "RunInstances", "USW2-BoxUsage:t3.micro", "AmazonEC2", "Compute", "Credit", "i-1", "price-1", "111111111111",
	}, ",")
	csv := strings.Join([]string{header, rowUsage, rowCredit}, "\n")

	tmp := t.TempDir()
	in := filepath.Join(tmp, "in.csv")
	out := filepath.Join(tmp, "out.ndjson")
	if err := os.WriteFile(in, []byte(csv), 0o600); err != nil {
		t.Fatalf("write input: %v", err)
	}

	preUsage := getAWSCount("unified", "Usage")
	preCredit := getAWSCount("unified", "Credit")

	conv := NewAWSConverter()
	cfg := &types.ConversionConfig{Provider: "aws", InputPath: in, OutputPath: out, Streaming: true, UseUnifiedMapper: true}
	if err := conv.ValidateInput(context.Background(), cfg); err != nil {
		t.Fatalf("validate: %v", err)
	}
	if _, err := conv.ConvertStream(context.Background(), cfg, nil); err != nil {
		t.Fatalf("convert: %v", err)
	}

	postUsage := getAWSCount("unified", "Usage")
	postCredit := getAWSCount("unified", "Credit")
	if int(postUsage-preUsage) < 1 {
		t.Fatalf("expected unified Usage increment, delta=%.0f", postUsage-preUsage)
	}
	if int(postCredit-preCredit) < 1 {
		t.Fatalf("expected unified Credit increment, delta=%.0f", postCredit-preCredit)
	}
}
