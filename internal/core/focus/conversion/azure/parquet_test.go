//go:build !race
// +build !race

package azure_test

import (
	"context"
	azure "local/costscope/internal/core/focus/conversion/azure"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"local/costscope/internal/core/focus/types"
)

// Ensures Azure converter can write Parquet when requested
func TestAzureCSVStreaming_ParquetOutput(t *testing.T) {
	csv := strings.Join([]string{
		"BillingAccountId,BillingAccountName,BillingCurrency,SubscriptionId,SubscriptionName,ServiceName,ServiceFamily,ResourceId,ResourceName,ResourceType,ResourceLocation,Quantity,UnitOfMeasure,AmortizedCost,RetailPrice,UsageStart,UsageEnd",
		"BA-1,Main,USD,sub-123,Dev,Virtual Machines,Compute,/subs/sub-123/rg/rg1/vm/vm1,vm1,Microsoft.Compute/virtualMachines,eastus,1,Hours,1.00,1.0,2024-01-01T00:00:00Z,2024-01-01T01:00:00Z",
	}, "\n")

	tmp := t.TempDir()
	in := filepath.Join(tmp, "in.csv")
	out := filepath.Join(tmp, "out.parquet")
	if err := os.WriteFile(in, []byte(csv), 0o600); err != nil {
		t.Fatalf("write input: %v", err)
	}

	conv := azure.NewAzureConverter()
	cfg := &types.ConversionConfig{
		Provider:     "azure",
		InputPath:    in,
		OutputPath:   out,
		OutputFormat: "parquet",
		Streaming:    true,
		ChunkSize:    1000,
		Workers:      1,
		ConversionId: "test-azure-parquet",
	}

	// Parquet-go exhibits data races during WriteStop/Close under -race.
	// Disable rotation to keep deterministic single-file output at the base path.
	cfg.Parquet.RotateSizeBytes = -1

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

	// ensure parquet output exists and is non-empty
	fi, err := os.Stat(out)
	if err != nil {
		t.Fatalf("expected parquet output: %v", err)
	}
	if fi.Size() == 0 {
		t.Fatalf("parquet output should not be empty")
	}
}
