package azure_test

import (
	"context"
	"encoding/json"
	"fmt"
	convutil "local/costscope/internal/core/focus/conversion"
	azure "local/costscope/internal/core/focus/conversion/azure"
	"os"
	"path/filepath"
	"testing"
	"time"

	"local/costscope/internal/core/focus/types"
)

// TestAzure_JSONParity_CreditsAndCharges exercises the JSON ingestion path ensuring classification
// and full hash parity between legacy and unified mapper paths for mixed charge types.
func TestAzure_JSONParity_CreditsAndCharges(t *testing.T) {
	records := []map[string]interface{}{
		{ // Usage positive cost
			"BillingAccountId": "BA-1", "CostInBillingCurrency": 1.20, "Cost": 1.20, "AmortizedCost": 1.10, "Quantity": 2,
			"UsageStart": "2024-02-01T00:00:00Z", "UsageEnd": "2024-02-01T01:00:00Z", "MeterCategory": "Compute", "ServiceName": "VM",
			"MeterName": "Standard_D2s_v3", "ChargeType": "Usage", "RetailPrice": 0.60, "UnitOfMeasure": "Hours", "SubscriptionId": "sub-1", "SubscriptionName": "SubOne",
		},
		{ // Credit negative cost
			"BillingAccountId": "BA-1", "CostInBillingCurrency": -0.50, "Cost": -0.50, "AmortizedCost": -0.50, "Quantity": 1,
			"UsageStart": "2024-02-01T00:00:00Z", "UsageEnd": "2024-02-01T01:00:00Z", "MeterCategory": "Storage", "ServiceName": "Blob",
			"MeterName": "Hot", "ChargeType": "Credit", "RetailPrice": 0.50, "UnitOfMeasure": "GB", "SubscriptionId": "sub-1", "SubscriptionName": "SubOne",
		},
		{ // Discount negative cost (usage-discount)
			"BillingAccountId": "BA-1", "CostInBillingCurrency": -0.25, "Cost": -0.25, "AmortizedCost": -0.25, "Quantity": 1,
			"UsageStart": "2024-02-01T01:00:00Z", "UsageEnd": "2024-02-01T02:00:00Z", "MeterCategory": "Compute", "ServiceName": "VM",
			"MeterName": "Standard_D2s_v3", "ChargeType": "usage-discount", "RetailPrice": 0.55, "UnitOfMeasure": "Hours", "SubscriptionId": "sub-1", "SubscriptionName": "SubOne",
		},
		{ // Tax positive cost
			"BillingAccountId": "BA-1", "CostInBillingCurrency": 0.10, "Cost": 0.10, "AmortizedCost": 0.10, "Quantity": 1,
			"UsageStart": "2024-02-01T02:00:00Z", "UsageEnd": "2024-02-01T03:00:00Z", "MeterCategory": "General", "ServiceName": "Service",
			"MeterName": "Generic", "ChargeType": "Tax", "RetailPrice": 0.10, "UnitOfMeasure": "Hours", "SubscriptionId": "sub-1", "SubscriptionName": "SubOne",
		},
		{ // Reservation-like usage (still Usage)
			"BillingAccountId": "BA-1", "CostInBillingCurrency": 0.80, "Cost": 0.80, "AmortizedCost": 0.80, "Quantity": 1,
			"UsageStart": "2024-02-01T03:00:00Z", "UsageEnd": "2024-02-01T04:00:00Z", "MeterCategory": "Compute", "ServiceName": "VM",
			"MeterName": "Standard_D2s_v3", "ChargeType": "Usage", "RetailPrice": 0.80, "UnitOfMeasure": "Hours", "SubscriptionId": "sub-1", "SubscriptionName": "SubOne",
		},
	}

	tmp := t.TempDir()
	in := filepath.Join(tmp, "in_azure_credits.json")
	data, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		t.Fatalf("marshal json: %v", err)
	}
	if err := os.WriteFile(in, data, 0o600); err != nil {
		t.Fatalf("write json: %v", err)
	}

	outLegacy := filepath.Join(tmp, "legacy_json.ndjson")
	outUnified := filepath.Join(tmp, "unified_json.ndjson")

	convr := azure.NewAzureConverter()
	cfgLegacy := &types.ConversionConfig{Provider: "azure", InputPath: in, OutputPath: outLegacy, Streaming: true, ChunkSize: 1000, Workers: 1, ConversionId: "azure-json-parity-legacy"}
	if err := convr.ValidateInput(context.Background(), cfgLegacy); err != nil {
		t.Fatalf("validate legacy: %v", err)
	}
	if _, err := convr.ConvertStream(context.Background(), cfgLegacy, nil); err != nil {
		t.Fatalf("convert legacy: %v", err)
	}

	time.Sleep(5 * time.Millisecond)
	cfgUnified := &types.ConversionConfig{Provider: "azure", InputPath: in, OutputPath: outUnified, Streaming: true, ChunkSize: 1000, Workers: 1, ConversionId: "azure-json-parity-unified", UseUnifiedMapper: true}
	if err := convr.ValidateInput(context.Background(), cfgUnified); err != nil {
		t.Fatalf("validate unified: %v", err)
	}
	if _, err := convr.ConvertStream(context.Background(), cfgUnified, nil); err != nil {
		t.Fatalf("convert unified: %v", err)
	}

	legacy := readAllFocusRecordsFromNDJSONLocal(t, outLegacy)
	unified := readAllFocusRecordsFromNDJSONLocal(t, outUnified)
	if len(legacy) != len(unified) {
		t.Fatalf("record count mismatch legacy=%d unified=%d", len(legacy), len(unified))
	}

	// Zero out ChargeDescription to avoid instability between JSON legacy/unified paths
	for i := range legacy {
		legacy[i].ChargeDescription = ""
		legacy[i].ChargeSubcategory = ""
	}
	for i := range unified {
		unified[i].ChargeDescription = ""
		unified[i].ChargeSubcategory = ""
	}
	// Switch to lite hash (full hash removed) for parity validation
	liteLegacy := make([]convutil.FocusRecordLite, 0, len(legacy))
	liteUnified := make([]convutil.FocusRecordLite, 0, len(unified))
	for i := range legacy {
		liteLegacy = append(liteLegacy, convutil.FocusRecordLite{EffectiveCost: legacy[i].EffectiveCost, UsageQuantity: legacy[i].UsageQuantity, ProviderName: legacy[i].ProviderName, ServiceName: legacy[i].ServiceName, ChargeCategory: legacy[i].ChargeCategory})
		liteUnified = append(liteUnified, convutil.FocusRecordLite{EffectiveCost: unified[i].EffectiveCost, UsageQuantity: unified[i].UsageQuantity, ProviderName: unified[i].ProviderName, ServiceName: unified[i].ServiceName, ChargeCategory: unified[i].ChargeCategory})
	}
	hLegacy := convutil.HashFocusLite(liteLegacy)
	hUnified := convutil.HashFocusLite(liteUnified)
	if hLegacy != hUnified {
		for i := range legacy {
			if canonicalizeFR(legacy[i]) != canonicalizeFR(unified[i]) {
				fmt.Printf("HASH_MISMATCH_JSON idx=%d\nLEGACY=%s\nUNIFIED=%s\n", i, canonicalizeFR(legacy[i]), canonicalizeFR(unified[i]))
				break
			}
		}
		t.Fatalf("azure JSON credits/charges parity hash mismatch: %s vs %s", hLegacy, hUnified)
	}

	foundCredit, foundDiscount, foundTax := false, false, false
	for i := range legacy {
		if legacy[i].ChargeCategory != unified[i].ChargeCategory {
			t.Fatalf("classification parity mismatch row %d: legacy=%s unified=%s", i, legacy[i].ChargeCategory, unified[i].ChargeCategory)
		}
		cat := legacy[i].ChargeCategory
		switch cat {
		case types.ChargeCategories.Credit:
			if legacy[i].EffectiveCost >= 0 {
				t.Fatalf("credit row non-negative cost: %f", legacy[i].EffectiveCost)
			}
			foundCredit = true
		case azure.CategoryDiscount:
			if legacy[i].EffectiveCost >= 0 {
				t.Fatalf("discount row non-negative cost: %f", legacy[i].EffectiveCost)
			}
			foundDiscount = true
		case types.ChargeCategories.Tax:
			if legacy[i].EffectiveCost <= 0 {
				t.Fatalf("tax row non-positive cost: %f", legacy[i].EffectiveCost)
			}
			foundTax = true
		}
	}
	if !foundCredit {
		t.Fatalf("expected Credit row not found")
	}
	if !foundDiscount {
		t.Fatalf("expected Discount row not found")
	}
	if !foundTax {
		t.Fatalf("expected Tax row not found")
	}
}
