package azure_test

import (
	"compress/gzip"
	"context"
	azure "local/costscope/internal/core/focus/conversion/azure"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"local/costscope/internal/core/focus/types"
)

// Ensures gzipped CSV input is supported end-to-end
func TestAzureCSVStreaming_GzipInput(t *testing.T) {
	csv := strings.Join([]string{
		"BillingAccountId,BillingAccountName,BillingCurrency,SubscriptionId,SubscriptionName,ServiceName,ServiceFamily,ResourceId,ResourceName,ResourceType,ResourceLocation,Quantity,UnitOfMeasure,AmortizedCost,RetailPrice,UsageStart,UsageEnd",
		"BA-1,Main,USD,sub-123,Dev,Virtual Machines,Compute,/subs/sub-123/rg/rg1/vm/vm1,vm1,Microsoft.Compute/virtualMachines,eastus,1,Hours,1.00,1.0,2024-01-01T00:00:00Z,2024-01-01T01:00:00Z",
	}, "\n")

	tmp := t.TempDir()
	in := filepath.Join(tmp, "in.csv.gz")
	out := filepath.Join(tmp, "out.ndjson")

	// #nosec G304 - writing test input inside temp dir
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

	conv := azure.NewAzureConverter()
	cfg := &types.ConversionConfig{
		Provider:     "azure",
		InputPath:    in,
		OutputPath:   out,
		Streaming:    true,
		ChunkSize:    1000,
		Workers:      1,
		ConversionId: "test-azure-csv-gzip",
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

	// ensure output is non-empty
	// #nosec G304 - reading file created in temp dir
	data, err := os.ReadFile(out)
	if err != nil || len(data) == 0 {
		t.Fatalf("expected output file: %v, size=%d", err, len(data))
	}
}

// Ensures gzipped JSON array input is supported
func TestAzureJSONStreaming_Array_Gzip(t *testing.T) {
	jsonBody := `[
        {
            "BillingAccountId": "BA-1",
            "BillingAccountName": "Main",
            "BillingCurrency": "USD",
            "SubscriptionId": "sub-1",
            "SubscriptionName": "Dev",
            "ServiceName": "Virtual Machines",
            "ServiceFamily": "Compute",
            "ResourceId": "/subs/sub-1/rg/rg1/vm/vm1",
            "ResourceName": "vm1",
            "ResourceType": "Microsoft.Compute/virtualMachines",
            "ResourceLocation": "eastus",
            "Quantity": 5,
            "UnitOfMeasure": "Hours",
            "AmortizedCost": 7.5,
            "RetailPrice": 1.5,
            "UsageStart": "2024-01-01T00:00:00Z",
            "UsageEnd": "2024-01-01T01:00:00Z"
        }
    ]`

	tmp := t.TempDir()
	in := filepath.Join(tmp, "in.json.gz")
	out := filepath.Join(tmp, "out.ndjson")

	// #nosec G304 - writing test input inside temp dir
	f, err := os.Create(in)
	if err != nil {
		t.Fatalf("create input: %v", err)
	}
	gz := gzip.NewWriter(f)
	if _, err := gz.Write([]byte(jsonBody)); err != nil {
		t.Fatalf("write gzip: %v", err)
	}
	_ = gz.Close()
	_ = f.Close()

	conv := azure.NewAzureConverter()
	cfg := &types.ConversionConfig{
		Provider:     "azure",
		InputPath:    in,
		OutputPath:   out,
		Streaming:    true,
		ChunkSize:    1000,
		Workers:      1,
		ConversionId: "test-azure-json-gzip",
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

	// ensure output is non-empty
	// #nosec G304 - reading file created in temp dir
	data, err := os.ReadFile(out)
	if err != nil || len(data) == 0 {
		t.Fatalf("expected output file: %v, size=%d", err, len(data))
	}
}
