package azure_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	azure "github.com/costscope/costscope/internal/core/focus/conversion/azure"

	"github.com/costscope/costscope/internal/core/focus/types"
)

// TestAzureDiscountClassificationCSV ensures discount variants across legacy and unified CSV paths
// classify identically using the shared helper.
func TestAzureDiscountClassificationCSV(t *testing.T) {
	tmp := t.TempDir()
	csvPath := filepath.Join(tmp, "azure_discounts.csv")
	var b strings.Builder
	b.WriteString(strings.Join([]string{
		"BillingAccountId", "CostInBillingCurrency", "Cost", "AmortizedCost", "Quantity", "UsageStart", "UsageEnd", "MeterCategory", "ServiceName", "MeterName", "ChargeType", "BillingType", "RetailPrice", "UnitOfMeasure", "SubscriptionId", "SubscriptionName",
	}, ",") + "\n")
	rows := [][]string{
		{"BA-1", "1", "1", "1", "1", "2025-01-01T00:00:00Z", "2025-01-01T01:00:00Z", "Compute", "VM", "Standard_DS1", "discount", "", "0.5", "Hours", "sub-1", "SubOne"},
		{"BA-1", "1", "1", "1", "1", "2025-01-01T01:00:00Z", "2025-01-01T02:00:00Z", "Compute", "VM", "Standard_DS1", "usage-discount", "", "0.5", "Hours", "sub-1", "SubOne"},
		{"BA-1", "1", "1", "1", "1", "2025-01-01T02:00:00Z", "2025-01-01T03:00:00Z", "Compute", "VM", "Standard_DS1", "UsAgE-DisCoUnT", "", "0.5", "Hours", "sub-1", "SubOne"},
		{"BA-1", "1", "1", "1", "1", "2025-01-01T03:00:00Z", "2025-01-01T04:00:00Z", "Compute", "VM", "Standard_DS1", "promo-discount", "", "0.5", "Hours", "sub-1", "SubOne"},
		{"BA-1", "1", "1", "1", "1", "2025-01-01T04:00:00Z", "2025-01-01T05:00:00Z", "Compute", "VM", "Standard_DS1", "", "reservation-discount", "0.5", "Hours", "sub-1", "SubOne"},
		{"BA-1", "-0.25", "-0.25", "-0.25", "1", "2025-01-01T05:00:00Z", "2025-01-01T06:00:00Z", "Compute", "VM", "Standard_DS1", "discount", "", "0.5", "Hours", "sub-1", "SubOne"}, // negative cost discount stays Discount (Credit reserved for generic negative usage w/out discount token)
	}
	for _, r := range rows {
		b.WriteString(strings.Join(r, ",") + "\n")
	}
	if err := os.WriteFile(csvPath, []byte(b.String()), 0o600); err != nil {
		t.Fatalf("write csv: %v", err)
	}

	legacyOut := filepath.Join(tmp, "legacy_discounts.ndjson")
	unifiedOut := filepath.Join(tmp, "unified_discounts.ndjson")

	conv := azure.NewAzureConverter()
	legacyCfg := &types.ConversionConfig{Provider: "azure", InputPath: csvPath, OutputPath: legacyOut, Streaming: true, UseUnifiedMapper: false, ChunkSize: 1000, Workers: 1, ConversionId: "azure-discount-csv-legacy"}
	if err := conv.ValidateInput(context.Background(), legacyCfg); err != nil {
		t.Fatalf("validate legacy: %v", err)
	}
	if _, err := conv.ConvertStream(context.Background(), legacyCfg, nil); err != nil {
		t.Fatalf("convert legacy: %v", err)
	}

	unifiedCfg := &types.ConversionConfig{Provider: "azure", InputPath: csvPath, OutputPath: unifiedOut, Streaming: true, UseUnifiedMapper: true, ChunkSize: 1000, Workers: 1, ConversionId: "azure-discount-csv-unified"}
	if err := conv.ValidateInput(context.Background(), unifiedCfg); err != nil {
		t.Fatalf("validate unified: %v", err)
	}
	if _, err := conv.ConvertStream(context.Background(), unifiedCfg, nil); err != nil {
		t.Fatalf("convert unified: %v", err)
	}

	legacy := readAllFocusRecordsFromNDJSONLocal(t, legacyOut)
	unified := readAllFocusRecordsFromNDJSONLocal(t, unifiedOut)
	if len(legacy) != len(unified) {
		t.Fatalf("record count mismatch %d vs %d", len(legacy), len(unified))
	}

	for i := range legacy {
		if legacy[i].ChargeCategory != unified[i].ChargeCategory {
			fmt.Printf("MISMATCH idx=%d legacy=%s unified=%s\n", i, legacy[i].ChargeCategory, unified[i].ChargeCategory)
			// fallthrough to fail
		}
	}

	expect := []string{"Discount", "Discount", "Discount", "Discount", "Discount", "Discount"}
	for i, cat := range expect {
		if legacy[i].ChargeCategory != cat || unified[i].ChargeCategory != cat {
			if legacy[i].ChargeCategory == unified[i].ChargeCategory {
				t.Fatalf("row %d category unexpected got=%s want=%s", i, legacy[i].ChargeCategory, cat)
			}
			t.Fatalf("row %d categories differ legacy=%s unified=%s want=%s", i, legacy[i].ChargeCategory, unified[i].ChargeCategory, cat)
		}
	}
}
