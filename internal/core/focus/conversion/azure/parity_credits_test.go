package azure_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	convutil "github.com/costscope/costscope/internal/core/focus/conversion"
	azure "github.com/costscope/costscope/internal/core/focus/conversion/azure"

	"github.com/costscope/costscope/internal/core/focus/types"
)

// TestAzure_CreditsAndChargesParity validates deterministic lite hash parity between legacy and unified paths plus semantic checks.
func TestAzure_FullParity_CreditsAndCharges(t *testing.T) {
	header := "BillingAccountId,CostInBillingCurrency,Cost,AmortizedCost,Quantity,UsageStart,UsageEnd,MeterCategory,ServiceName,MeterName,ChargeType,RetailPrice,UnitOfMeasure,SubscriptionId,SubscriptionName"
	rows := []string{
		// Usage positive cost
		"BA-1,1.20,1.20,1.10,2,2024-02-01T00:00:00Z,2024-02-01T01:00:00Z,Compute,VM,Standard_D2s_v3,Usage,0.60,Hours,sub-1,SubOne",
		// Credit negative cost
		"BA-1,-0.50,-0.50,-0.50,1,2024-02-01T00:00:00Z,2024-02-01T01:00:00Z,Storage,Blob,Hot,Credit,0.50,GB,sub-1,SubOne",
		// Discount negative cost (ChargeType includes 'usage-discount' to trigger classification case-insensitively)
		"BA-1,-0.25,-0.25,-0.25,1,2024-02-01T01:00:00Z,2024-02-01T02:00:00Z,Compute,VM,Standard_D2s_v3,usage-discount,0.55,Hours,sub-1,SubOne",
		// Tax positive cost (classification should reflect Tax, not Credit)
		"BA-1,0.10,0.10,0.10,1,2024-02-01T02:00:00Z,2024-02-01T03:00:00Z,General,Service,Generic,Tax,0.10,Hours,sub-1,SubOne",
		// Reservation usage positive cost (tests benefit classification remains Usage)
		"BA-1,0.80,0.80,0.80,1,2024-02-01T03:00:00Z,2024-02-01T04:00:00Z,Compute,VM,Standard_D2s_v3,Usage,0.80,Hours,sub-1,SubOne",
	}
	csvData := header + "\n" + rows[0] + "\n" + rows[1] + "\n" + rows[2] + "\n" + rows[3] + "\n" + rows[4] + "\n"
	tmp := t.TempDir()
	in := filepath.Join(tmp, "in_azure_credits.csv")
	if err := os.WriteFile(in, []byte(csvData), 0o600); err != nil {
		t.Fatalf("write csv: %v", err)
	}
	outLegacy := filepath.Join(tmp, "legacy.ndjson")
	outUnified := filepath.Join(tmp, "unified.ndjson")

	convr := azure.NewAzureConverter()
	cfgLegacy := &types.ConversionConfig{Provider: "azure", InputPath: in, OutputPath: outLegacy, Streaming: true, ChunkSize: 1000, Workers: 1, ConversionId: "azure-parity-credits-legacy"}
	if err := convr.ValidateInput(context.Background(), cfgLegacy); err != nil {
		t.Fatalf("validate legacy: %v", err)
	}
	if _, err := convr.ConvertStream(context.Background(), cfgLegacy, nil); err != nil {
		t.Fatalf("convert legacy: %v", err)
	}

	time.Sleep(5 * time.Millisecond)
	cfgUnified := &types.ConversionConfig{Provider: "azure", InputPath: in, OutputPath: outUnified, Streaming: true, ChunkSize: 1000, Workers: 1, ConversionId: "azure-parity-credits-unified", UseUnifiedMapper: true}
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

	// Use lite parity hash (full hash removed as part of deadcode reduction)
	liteLegacy := make([]convutil.FocusRecordLite, 0, len(legacy))
	liteUnified := make([]convutil.FocusRecordLite, 0, len(unified))
	for i := range legacy {
		liteLegacy = append(liteLegacy, convutil.FocusRecordLite{EffectiveCost: legacy[i].EffectiveCost, UsageQuantity: legacy[i].UsageQuantity, ProviderName: legacy[i].ProviderName, ServiceName: legacy[i].ServiceName, ChargeCategory: legacy[i].ChargeCategory})
		liteUnified = append(liteUnified, convutil.FocusRecordLite{EffectiveCost: unified[i].EffectiveCost, UsageQuantity: unified[i].UsageQuantity, ProviderName: unified[i].ProviderName, ServiceName: unified[i].ServiceName, ChargeCategory: unified[i].ChargeCategory})
	}
	hLegacy := convutil.HashFocusLite(liteLegacy)
	hUnified := convutil.HashFocusLite(liteUnified)
	if hLegacy != hUnified {
		// Build minimal diff for first differing record
		var diffIdx = -1
		max := len(legacy)
		for i := 0; i < max; i++ {
			c1 := canonicalizeFR(legacy[i])
			c2 := canonicalizeFR(unified[i])
			if c1 != c2 {
				diffIdx = i
				break
			}
		}
		if diffIdx >= 0 {
			l := canonicalizeFR(legacy[diffIdx])
			u := canonicalizeFR(unified[diffIdx])
			fmt.Printf("HASH_MISMATCH idx=%d\nLEGACY=%s\nUNIFIED=%s\n", diffIdx, l, u)
		}
		t.Fatalf("azure credits/charges parity hash mismatch: %s vs %s", hLegacy, hUnified)
	}

	// Semantic checks: ensure negative cost credit row classified Credit; discount row classified Discount; tax row not misclassified as Credit.
	for i := range legacy {
		if legacy[i].ChargeCategory != unified[i].ChargeCategory {
			t.Fatalf("classification parity mismatch at row %d: legacy=%s unified=%s", i, legacy[i].ChargeCategory, unified[i].ChargeCategory)
		}
	}
	foundCredit := false
	foundDiscount := false
	foundTax := false
	for idx, r := range legacy {
		switch r.ChargeCategory {
		case types.ChargeCategories.Credit:
			if r.EffectiveCost >= 0 {
				t.Fatalf("credit row has non-negative cost: %f", r.EffectiveCost)
			}
			foundCredit = true
		case azure.CategoryDiscount:
			if r.EffectiveCost >= 0 {
				t.Fatalf("discount row has non-negative cost: %f", r.EffectiveCost)
			}
			foundDiscount = true
		case "Tax":
			if r.EffectiveCost <= 0 {
				t.Fatalf("tax row has non-positive cost: %f", r.EffectiveCost)
			}
			foundTax = true
		default:
			// record debug
			fmt.Printf("row %d category=%s eff=%f rawMeter=%s\n", idx, r.ChargeCategory, r.EffectiveCost, r.ServiceName)
		}
	}
	if !foundCredit {
		t.Fatalf("expected a Credit row not found")
	}
	if !foundDiscount {
		t.Fatalf("expected a Discount row not found")
	}
	if !foundTax {
		t.Fatalf("expected a Tax row not found")
	}
}
