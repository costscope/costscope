package azure_test

import (
	"bufio"
	"context"
	"encoding/json"
	azure "local/costscope/internal/core/focus/conversion/azure"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"local/costscope/internal/core/focus/types"
)

func TestAzureJSONStreaming_Array(t *testing.T) {
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
	in := filepath.Join(tmp, "in.json")
	out := filepath.Join(tmp, "out.ndjson")
	if err := os.WriteFile(in, []byte(jsonBody), 0o600); err != nil {
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
		ConversionId: "test-azure-json-array",
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

	// #nosec G304 - test reads file created in temp dir
	data, err := os.ReadFile(out)
	if err != nil || len(data) == 0 {
		t.Fatalf("expected output file: %v, size=%d", err, len(data))
	}
	if !strings.Contains(string(data), "Microsoft Azure") {
		t.Fatalf("expected provider name in output")
	}
}

func TestAzureCSV_NegativeCost_DefaultsToCredit(t *testing.T) {
	csv := strings.Join([]string{
		"SubscriptionId,Quantity,UnitOfMeasure,ServiceName,ResourceLocation,CostInBillingCurrency,UsageStart,UsageEnd",
		"sub-123,1,Hours,Virtual Machines,eastus,-3.00,2024-01-01T00:00:00Z,2024-01-01T01:00:00Z",
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
		ConversionId: "test-azure-negative-cost",
	}

	if err := conv.ValidateInput(context.Background(), cfg); err != nil {
		t.Fatalf("validate: %v", err)
	}
	if _, err := conv.ConvertStream(context.Background(), cfg, nil); err != nil {
		t.Fatalf("convert: %v", err)
	}

	// #nosec G304 - test opens file created in temp dir
	f, err := os.Open(out)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	r := bufio.NewReader(f)
	line, _, err := r.ReadLine()
	if err != nil {
		t.Fatalf("read output: %v", err)
	}

	var obj map[string]interface{}
	if err := json.Unmarshal(line, &obj); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	eff := obj["effective_cost"]
	billed := obj["billed_cost"]
	if got := obj["charge_category"]; got != types.ChargeCategories.Credit {
		t.Fatalf("expected charge_category=Credit, got %v (effective_cost=%v, billed_cost=%v)", got, eff, billed)
	}
}

func TestAzureMapRecordToFocus_NegativeBilled(t *testing.T) {
	// End-to-end check using public API ensuring negative billed cost stays negative and classified as Credit
	header := strings.Join([]string{
		"SubscriptionId",
		"Quantity",
		"UnitOfMeasure",
		"ServiceName",
		"ResourceLocation",
		"CostInBillingCurrency",
		"UsageStart",
		"UsageEnd",
	}, ",")
	row := strings.Join([]string{
		"sub-xyz",
		"1",
		"Hours",
		"Virtual Machines",
		"eastus",
		"-3.00",
		"2024-01-01T00:00:00Z",
		"2024-01-01T01:00:00Z",
	}, ",")
	csv := strings.Join([]string{header, row}, "\n")

	tmp := t.TempDir()
	in := filepath.Join(tmp, "in.csv")
	out := filepath.Join(tmp, "out.ndjson")
	if err := os.WriteFile(in, []byte(csv), 0o600); err != nil {
		t.Fatalf("write input: %v", err)
	}

	conv := azure.NewAzureConverter()
	cfg := &types.ConversionConfig{Provider: "azure", InputPath: in, OutputPath: out, Streaming: true, ChunkSize: 1000, Workers: 1, ConversionId: "neg-billed"}
	if err := conv.ValidateInput(context.Background(), cfg); err != nil {
		t.Fatalf("validate: %v", err)
	}
	if _, err := conv.ConvertStream(context.Background(), cfg, nil); err != nil {
		t.Fatalf("convert: %v", err)
	}

	// #nosec G304 - open file created in temp dir
	f, err := os.Open(out)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	r := bufio.NewReader(f)
	line, _, err := r.ReadLine()
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	var obj map[string]any
	if err := json.Unmarshal(line, &obj); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if eff, _ := obj["effective_cost"].(float64); eff >= 0 {
		t.Fatalf("expected negative effective_cost, got %v", eff)
	}
	if bc, ok := obj["billed_cost"].(float64); !ok || bc >= 0 {
		t.Fatalf("expected negative billed_cost, got %v", obj["billed_cost"])
	}
	if cc, _ := obj["charge_category"].(string); cc != types.ChargeCategories.Credit {
		t.Fatalf("expected charge_category Credit, got %v", cc)
	}
}
