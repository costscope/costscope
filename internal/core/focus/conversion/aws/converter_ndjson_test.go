package aws

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"local/costscope/internal/core/focus/types"
)

func TestAWSCSVStreaming_NDJSONOutput(t *testing.T) {
	// Minimal AWS CUR-like CSV with required fields for mapping
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
		"lineItem/ResourceId",
		"pricing/PriceId",
		"lineItem/UsageAccountId",
	}, ",")

	row := strings.Join([]string{
		"123456789012",
		"Master",
		"USD",
		"1.23",
		"10",
		"2024-01-01 00:00:00",
		"2024-01-01 01:00:00",
		"EC2 usage",
		"RunInstances",
		"USW2-BoxUsage:t3.micro",
		"AmazonEC2",
		"Compute",
		"i-abc123",
		"price-1",
		"111111111111",
	}, ",")

	csv := strings.Join([]string{header, row}, "\n")

	tmp := t.TempDir()
	in := filepath.Join(tmp, "in.csv")
	out := filepath.Join(tmp, "out.ndjson")
	if err := os.WriteFile(in, []byte(csv), 0o600); err != nil {
		t.Fatalf("write input: %v", err)
	}

	conv := NewAWSConverter()
	cfg := &types.ConversionConfig{
		Provider:     "aws",
		InputPath:    in,
		OutputPath:   out,
		Streaming:    true,
		ChunkSize:    1000,
		Workers:      1,
		ConversionId: "test-aws-ndjson",
	}

	if err := conv.ValidateInput(context.Background(), cfg); err != nil {
		t.Fatalf("validate: %v", err)
	}

	res, err := conv.ConvertStream(context.Background(), cfg, nil)
	if err != nil {
		t.Fatalf("convert: %v", err)
	}

	if !res.Success {
		t.Fatalf("expected success: %+v", res)
	}
	if res.OutputRecords != 1 {
		t.Fatalf("expected 1 output record, got %d", res.OutputRecords)
	}

	// #nosec G304 - reading file created in temp dir
	data, err := os.ReadFile(out)
	if err != nil || len(data) == 0 {
		t.Fatalf("expected non-empty output: %v, size=%d", err, len(data))
	}
}
