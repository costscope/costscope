package azure_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	azure "github.com/costscope/costscope/internal/core/focus/conversion/azure"
	"github.com/costscope/costscope/internal/core/focus/types"
)

func TestAzureCSVStreaming_Minimal(t *testing.T) {
	// Minimal CSV with headers and one row
	csv := strings.Join([]string{
		"BillingAccountId,BillingAccountName,BillingCurrency,SubscriptionId,SubscriptionName,ServiceName,ServiceFamily,ResourceId,ResourceName,ResourceType,ResourceLocation,Quantity,UnitOfMeasure,AmortizedCost,RetailPrice,UsageStart,UsageEnd,Tags",
		"BA-1,Main,USD,sub-123,Dev,Virtual Machines,Compute,/subs/sub-123/rg/rg1/vm/vm1,vm1,Microsoft.Compute/virtualMachines,eastus,10,Hours,12.34,1.5,2024-01-01T00:00:00Z,2024-01-01T01:00:00Z,\"{\"env\":\"dev\"}\"",
	}, "\n")

	tmp := t.TempDir()
	in := filepath.Join(tmp, "in.csv")
	out := filepath.Join(tmp, "out.ndjson")
	if err := os.WriteFile(in, []byte(csv), 0o600); err != nil {
		t.Fatalf("write input: %v", err)
	}

	conv := azure.NewAzureConverter()
	cfg := &types.ConversionConfig{
		Provider:     "azure",
		InputPath:    in,
		OutputPath:   out,
		Streaming:    true,
		ChunkSize:    1000,
		Workers:      1,
		ConversionId: "test-azure-csv",
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

	// ensure output is non-empty
	// #nosec G304 - test reads file created in temp dir
	data, err := os.ReadFile(out)
	if err != nil || len(data) == 0 {
		t.Fatalf("expected output file: %v, size=%d", err, len(data))
	}
}
