package aws

import (
	"compress/gzip"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"local/costscope/internal/core/focus/types"
)

func TestAWSCSVStreaming_GzipInput(t *testing.T) {
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
	in := filepath.Join(tmp, "in.csv.gz")
	out := filepath.Join(tmp, "out.ndjson")

	// #nosec G304 - writing test input in temp dir
	f, err := os.Create(in)
	if err != nil {
		t.Fatalf("create input: %v", err)
	}
	gz := gzip.NewWriter(f)
	if _, err := gz.Write([]byte(csv)); err != nil {
		t.Fatalf("write gzip: %v", err)
	}
	_ = gz.Close()
	_ = f.Close()

	conv := NewAWSConverter()
	cfg := &types.ConversionConfig{
		Provider:     "aws",
		InputPath:    in,
		OutputPath:   out,
		Streaming:    true,
		ChunkSize:    1000,
		Workers:      1,
		ConversionId: "test-aws-gzip",
	}

	if err := conv.ValidateInput(context.Background(), cfg); err != nil {
		t.Fatalf("validate: %v", err)
	}

	res, err := conv.ConvertStream(context.Background(), cfg, nil)
	if err != nil {
		t.Fatalf("convert: %v", err)
	}
	if !res.Success || res.OutputRecords != 1 {
		t.Fatalf("unexpected result: %+v", res)
	}

	// #nosec G304 - reading file created in temp dir
	data, err := os.ReadFile(out)
	if err != nil || len(data) == 0 {
		t.Fatalf("expected non-empty output: %v, size=%d", err, len(data))
	}
}
