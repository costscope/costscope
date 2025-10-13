package azure_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	azure "github.com/costscope/costscope/internal/core/focus/conversion/azure"

	"github.com/costscope/costscope/internal/core/focus/types"
)

// TestAzureDiscountJSONSubstring validates substring discount normalization for JSON input parity.
func TestAzureDiscountJSONSubstring(t *testing.T) {
	records := []map[string]interface{}{
		{"ChargeType": "promo-discount", "UsageStart": "2025-01-01T00:00:00Z", "UsageEnd": "2025-01-01T01:00:00Z", "Quantity": 1, "CostInBillingCurrency": 1, "Cost": 1, "UnitOfMeasure": "Hours", "SubscriptionId": "s1", "ServiceName": "VM", "MeterCategory": "Compute", "MeterName": "Standard"},
		{"BillingType": "reservation-discount", "UsageStart": "2025-01-01T01:00:00Z", "UsageEnd": "2025-01-01T02:00:00Z", "Quantity": 1, "CostInBillingCurrency": 1, "Cost": 1, "UnitOfMeasure": "Hours", "SubscriptionId": "s1", "ServiceName": "VM", "MeterCategory": "Compute", "MeterName": "Standard"},
	}
	tmp := t.TempDir()
	input := filepath.Join(tmp, "azure_discount_array.json")
	data, _ := json.Marshal(records)
	if err := os.WriteFile(input, data, 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	outLegacy := filepath.Join(tmp, "legacy.ndjson")
	outUnified := filepath.Join(tmp, "unified.ndjson")
	convr := azure.NewAzureConverter()
	cfgL := &types.ConversionConfig{Provider: "azure", InputPath: input, OutputPath: outLegacy, Streaming: true, ConversionId: "azure-json-substring-legacy"}
	if err := convr.ValidateInput(context.Background(), cfgL); err != nil {
		t.Fatalf("validate legacy: %v", err)
	}
	if _, err := convr.ConvertStream(context.Background(), cfgL, nil); err != nil {
		t.Fatalf("convert legacy: %v", err)
	}
	cfgU := &types.ConversionConfig{Provider: "azure", InputPath: input, OutputPath: outUnified, Streaming: true, ConversionId: "azure-json-substring-unified", UseUnifiedMapper: true}
	if err := convr.ValidateInput(context.Background(), cfgU); err != nil {
		t.Fatalf("validate unified: %v", err)
	}
	if _, err := convr.ConvertStream(context.Background(), cfgU, nil); err != nil {
		t.Fatalf("convert unified: %v", err)
	}
	legacy := readAllFocusRecordsFromNDJSONLocal(t, outLegacy)
	unified := readAllFocusRecordsFromNDJSONLocal(t, outUnified)
	if len(legacy) != len(records) || len(unified) != len(records) {
		t.Fatalf("unexpected counts legacy=%d unified=%d", len(legacy), len(unified))
	}
	for i := range legacy {
		if legacy[i].ChargeCategory != azure.CategoryDiscount || unified[i].ChargeCategory != azure.CategoryDiscount {
			// Provide diagnostic info
			b, _ := json.MarshalIndent(legacy[i], "", "  ")
			ub, _ := json.MarshalIndent(unified[i], "", "  ")
			t.Fatalf("row %d ChargeCategory mismatch legacy=%s unified=%s\nlegacy=%s\nunified=%s", i, legacy[i].ChargeCategory, unified[i].ChargeCategory, string(b), string(ub))
		}
	}
}
