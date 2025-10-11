package aws

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"local/costscope/internal/core/focus/types"
)

func TestAWSConverterStreaming_TruncatedRowSkipped(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	input := filepath.Join(dir, "truncated.csv")
	// Minimal required headers for header index + one optional classification field.
	csvData := "bill/BillingAccountId,lineItem/UnblendedCost,lineItem/UsageStartDate,lineItem/UsageEndDate,product/ProductName,lineItem/UsageAccountId,lineItem/LineItemType\n" +
		"123,1.23,2024-01-01T00:00:00Z,2024-01-01T01:00:00Z,AmazonEC2,999,Usage\n" +
		"123,4.56,2024-01-01T00:00:00Z\n" // truncated row (too few columns)
	if err := os.WriteFile(input, []byte(csvData), 0o600); err != nil {
		t.Fatalf("write input: %v", err)
	}

	out := filepath.Join(dir, "out.parquet")
	conv := NewAWSConverter()
	cfg := &types.ConversionConfig{InputPath: input, OutputPath: out, Streaming: true, ChunkSize: 10, OutputFormat: "parquet"}
	res, err := conv.ConvertStream(context.Background(), cfg, nil)
	if err != nil {
		t.Fatalf("ConvertStream error: %v", err)
	}
	if res.InputRecords != 2 {
		t.Fatalf("expected InputRecords=2 got %d", res.InputRecords)
	}
	if res.OutputRecords != 1 {
		t.Fatalf("expected OutputRecords=1 got %d", res.OutputRecords)
	}
	if res.ErrorRecords < 1 {
		t.Fatalf("expected at least 1 ErrorRecord got %d", res.ErrorRecords)
	}
}
