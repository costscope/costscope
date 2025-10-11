package types

import (
	"testing"
	"time"
)

// Smoke test constructing representative structs to mark package covered
func TestFocusTypes_Construct(t *testing.T) {
	fr := &FocusRecord{BillingAccountId: "acc", BillingCurrency: "USD", BillingPeriodStart: nowUTC(), BillingPeriodEnd: nowUTC(), ChargeCategory: ChargeCategories.Usage, ChargeClass: ChargeClasses.OnDemand, PricingCategory: PricingCategories.Standard, ProviderName: ProviderNames.AWS, UsageQuantity: 1.0, UsageUnit: "Hrs"}
	if fr.BillingCurrency != "USD" || fr.ProviderName == "" {
		t.Fatalf("unexpected focus record: %+v", fr)
	}
	schema := GetFocusV12Schema()
	if schema.Version != "1.2" || len(schema.Fields) == 0 {
		t.Fatalf("unexpected schema: %+v", schema)
	}
}

// nowUTC helper (avoid importing time multiple times per test clarity)
func nowUTC() time.Time { return time.Now().UTC() }
