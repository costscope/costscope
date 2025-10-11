package azure_test

import (
	"context"
	azure "local/costscope/internal/core/focus/conversion/azure"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	focustypes "local/costscope/internal/core/focus/types"
)

const (
	cCurUSD        = "USD"
	cRetail09      = "0.09"
	cRetail01      = "0.1"
	cUsageStart    = "2024-08-01T00:00:00Z"
	cUsageEnd      = "2024-08-01T01:00:00Z"
	cDateIso       = "2024-08-01"
	cDateCompact   = "20240801"
	cFamilyCompute = "Compute"
	cOnDemandName  = "OnDemand"
)

// buildAzureDemoHeaders returns a representative subset of Azure export headers.
func buildAzureDemoHeaders() []string {
	return []string{
		"BillingAccountId", "BillingAccountName", "BillingCurrency",
		"SubscriptionId", "SubscriptionName",
		"ResourceId", "ResourceName", "ResourceType",
		"ServiceFamily", "ServiceName", "Product",
		"MeterName", "MeterSubCategory", "MeterId", "SkuId",
		"PartNumber", "ProductOrderNumber",
		"Quantity", "UnitOfMeasure", "MeterUnit",
		"RetailPrice", "UnitPrice",
		"PricingModel", "PricingModelName",
		"AmortizedCost", "CostInBillingCurrency", "Cost", "CostInUSD",
		"UsageStart", "UsageEnd", "Date", "UsageDate",
		"ChargeType", "BillingType",
		"Tags", "ResourceLocation", "Location",
	}
}

func buildAzureDemoRecords(n int) [][]string {
	headers := buildAzureDemoHeaders()
	idx := func(name string) int {
		for i, h := range headers {
			if h == name {
				return i
			}
		}
		return -1
	}
	rec := make([][]string, n)
	for i := 0; i < n; i++ {
		row := make([]string, len(headers))
		row[idx("BillingAccountId")] = "BA-123"
		row[idx("BillingAccountName")] = "Enterprise"
		row[idx("BillingCurrency")] = cCurUSD
		row[idx("SubscriptionId")] = "sub-" + strconv.Itoa(i%3)
		row[idx("SubscriptionName")] = "dev"
		row[idx("ResourceId")] = "/subscriptions/xxx/resourceGroups/rg/providers/Microsoft.Compute/virtualMachines/vm" + strconv.Itoa(i)
		row[idx("ResourceName")] = "vm" + strconv.Itoa(i)
		row[idx("ResourceType")] = "Microsoft.Compute/virtualMachines"
		row[idx("ServiceFamily")] = cFamilyCompute
		row[idx("ServiceName")] = "Virtual Machines"
		row[idx("Product")] = "D2s v5"
		row[idx("MeterName")] = "Compute Hours"
		row[idx("MeterSubCategory")] = "D2s"
		row[idx("MeterId")] = "m-123"
		row[idx("SkuId")] = "sku-123"
		row[idx("PartNumber")] = "PN-123"
		row[idx("ProductOrderNumber")] = "PO-123"
		row[idx("Quantity")] = "1"
		row[idx("UnitOfMeasure")] = "Hours"
		row[idx("MeterUnit")] = "Hours"
		row[idx("RetailPrice")] = cRetail01
		row[idx("UnitPrice")] = cRetail01
		row[idx("PricingModel")] = "on-demand"
		row[idx("PricingModelName")] = cOnDemandName
		row[idx("AmortizedCost")] = "0.08"
		row[idx("CostInBillingCurrency")] = cRetail09
		row[idx("Cost")] = cRetail09
		row[idx("CostInUSD")] = cRetail09
		row[idx("UsageStart")] = cUsageStart
		row[idx("UsageEnd")] = cUsageEnd
		row[idx("Date")] = cDateIso
		row[idx("UsageDate")] = cDateCompact
		row[idx("ChargeType")] = focustypes.ChargeCategories.Usage
		row[idx("BillingType")] = ""
		row[idx("Tags")] = "{\"env\":\"dev\"}"
		row[idx("ResourceLocation")] = "eastus"
		row[idx("Location")] = "eastus"
		rec[i] = row
	}
	return rec
}

// helper to materialize CSV content from demo headers and records
func azureDemoCSV(headers []string, records [][]string) string {
	var sb strings.Builder
	sb.WriteString(strings.Join(headers, ","))
	sb.WriteByte('\n')
	for _, r := range records {
		sb.WriteString(strings.Join(r, ","))
		sb.WriteByte('\n')
	}
	return sb.String()
}

// End-to-end benchmark via ConvertStream using a temp CSV file (table-driven)
func BenchmarkAzureConvert_CSV_EndToEnd(b *testing.B) {
	cases := []struct {
		name         string
		useUnified   bool
		conversionID string
	}{
		{name: "Legacy", useUnified: false, conversionID: "bench-azure-legacy"},
		{name: "Unified", useUnified: true, conversionID: "bench-azure-unified"},
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			az := azure.NewAzureConverter()
			headers := buildAzureDemoHeaders()
			records := buildAzureDemoRecords(10_000)
			toCSV := func() string { return azureDemoCSV(headers, records) }

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				dir := b.TempDir()
				in := filepath.Join(dir, "in.csv")
				out := filepath.Join(dir, "out.ndjson")
				if err := os.WriteFile(in, []byte(toCSV()), 0o600); err != nil {
					b.Fatalf("write csv: %v", err)
				}
				cfg := &focustypes.ConversionConfig{
					Provider:         "azure",
					InputPath:        in,
					OutputPath:       out,
					Streaming:        true,
					ChunkSize:        10000,
					Workers:          1,
					ConversionId:     tc.conversionID,
					UseUnifiedMapper: tc.useUnified,
				}
				if _, err := az.ConvertStream(context.Background(), cfg, nil); err != nil {
					b.Fatalf("convert: %v", err)
				}
			}
		})
	}
}
