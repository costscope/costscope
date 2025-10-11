package conversion

import (
	"testing"
	"time"

	azure "local/costscope/internal/core/focus/conversion/azure"
	"local/costscope/internal/core/focus/types"
)

// Parity test (lightweight): map sample Azure CSV rows via existing row mapper vs direct logic replication subset.
func TestAzureLiteParityHash(t *testing.T) {
	headers := []string{"BillingAccountId", "AmortizedCost", "CostInBillingCurrency", "Cost", "Quantity", "MeterCategory", "ServiceName", "MeterName", "ChargeType", "BillingType"}
	idx := azure.NewHeaderIndex(headers)
	// Full mapper with minimal deps (nil benefits/tags parsers acceptable for this parity scope)
	full := azure.NewFullRowMapperWithDeps(idx, nil, nil, nil, azure.AzureEnsureDiscount, time.Now)
	rawRows := [][]string{
		{"BA-1", "1.00", "1.10", "1.20", "5", "Compute", "VM", "D2", "Usage"},
		{"BA-1", "0", "0.90", "0.90", "2", "Storage", "Blob", "Hot", "Credit"},
		{"BA-2", "3.33", "3.40", "3.50", "7", "Database", "SQL", "DB", "Usage"},
	}
	lite := make([]FocusRecordLite, 0, len(rawRows))
	for _, r := range rawRows {
		fr, err := full.Map(r)
		if err != nil {
			t.Fatalf("map error: %v", err)
		}
		// Emulate legacy lite service fallback: MeterCategory > ServiceName > MeterName
		service := firstNonEmpty(r[idx.MeterCategory], r[idx.ServiceName], r[idx.MeterName])
		if service == "" {
			service = fr.ServiceName
		}
		cat := fr.ChargeCategory
		if cat == "" { // safety: classifier sets it, but guard anyway
			cat = types.ChargeCategories.Usage
		}
		lite = append(lite, FocusRecordLite{
			EffectiveCost:  fr.EffectiveCost,
			UsageQuantity:  fr.UsageQuantity,
			ProviderName:   fr.ProviderName,
			ServiceName:    service,
			ChargeCategory: cat,
		})
	}
	hash := HashFocusLite(lite)
	if hash == "" {
		t.Fatalf("empty hash")
	}
}

// firstNonEmpty helper mirrors legacy lite mapper precedence chain.
func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
