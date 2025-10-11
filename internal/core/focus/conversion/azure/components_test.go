package azure

import (
	"testing"
	"time"

	"local/costscope/internal/core/focus/types"
)

func basicHeaders() []string {
	return []string{
		"UsageStart", "UsageEnd", "AmortizedCost", "CostInBillingCurrency", "Cost", "CostInUSD",
		"Quantity", "UnitOfMeasure",
		"BillingAccountId", "BillingAccountName", "BillingCurrency", "Currency",
		"ServiceName", "Product", "MeterName", "MeterCategory", "MeterSubCategory",
		"SubscriptionId", "SubscriptionName", "ChargeType", "BillingType",
		"PricingModel", "PricingModelName", "RetailPrice", "UnitPrice",
		"ResourceLocation", "Location", "Tags",
	}
}

func basicRow(headers []string) []string {
	m := map[string]int{}
	for i, h := range headers {
		m[h] = i
	}
	r := make([]string, len(headers))
	now := time.Now().UTC().Format(time.RFC3339)
	r[m["UsageStart"]] = now
	r[m["UsageEnd"]] = now
	r[m["AmortizedCost"]] = "1.00"
	r[m["Quantity"]] = "2"
	r[m["UnitOfMeasure"]] = "Hours"
	r[m["BillingAccountId"]] = "A"
	r[m["BillingCurrency"]] = "usd"
	r[m["ServiceName"]] = "Compute"
	r[m["MeterName"]] = "vCPU"
	r[m["SubscriptionId"]] = "sub1"
	r[m["ChargeType"]] = "Usage"
	r[m["PricingModel"]] = "standard"
	r[m["ResourceLocation"]] = "eastus"
	return r
}

func TestFieldMapper_Basic(t *testing.T) {
	h := basicHeaders()
	idx := NewHeaderIndex(h)
	fm := newFieldMapper(idx, ApplyBenefitsRow, func(s string) types.Tags { return nil }, time.Now)
	row := basicRow(h)
	fr, err := fm.MapFields(row)
	if err != nil {
		t.Fatalf("MapFields: %v", err)
	}
	if fr.EffectiveCost != 1.00 {
		t.Fatalf("effective cost got %v", fr.EffectiveCost)
	}
	if fr.BillingCurrency != "USD" {
		t.Fatalf("billing currency expected USD got %s", fr.BillingCurrency)
	}
	if fr.PricingCategory == "" || fr.ChargeClass == "" {
		t.Fatalf("pricing or charge class empty")
	}
	if fr.Region == nil || *fr.Region == "" {
		t.Fatalf("region not set")
	}
}

func TestClassifier_Decisions(t *testing.T) {
	h := basicHeaders()
	idx := NewHeaderIndex(h)
	fm := newFieldMapper(idx, nil, nil, time.Now)
	cf := newClassifier(idx, classifyChargeCategoryAzure)
	cases := []struct {
		name   string
		mutate func([]string)
		want   string
	}{
		{"usage-positive", func(r []string) {}, "Usage"},
		{"tax", func(r []string) { r[colIndex(h, "ChargeType")] = "Tax" }, "Tax"},
		{"credit-token", func(r []string) { r[colIndex(h, "ChargeType")] = "Credit" }, "Credit"},
		{"purchase", func(r []string) { r[colIndex(h, "ChargeType")] = "Purchase" }, "Purchase"},
		{"negative-nondiscount", func(r []string) { r[colIndex(h, "AmortizedCost")] = "-1.0"; r[colIndex(h, "ChargeType")] = "Usage" }, "Credit"},
		{"usage-discount-token", func(r []string) {
			r[colIndex(h, "ChargeType")] = tokenUsageDisc
			r[colIndex(h, "AmortizedCost")] = "-1.0"
		}, "Usage"},
	}
	for _, tc := range cases {
		row := basicRow(h)
		if tc.mutate != nil {
			tc.mutate(row)
		}
		fr, err := fm.MapFields(row)
		if err != nil {
			t.Fatalf("map fields: %v", err)
		}
		cf.Classify(row, &fr)
		if fr.ChargeCategory != tc.want {
			t.Fatalf("%s: got %s want %s", tc.name, fr.ChargeCategory, tc.want)
		}
	}
}

func TestNormalizer_DiscountPromotion(t *testing.T) {
	h := basicHeaders()
	idx := NewHeaderIndex(h)
	fm := newFieldMapper(idx, nil, nil, time.Now)
	cf := newClassifier(idx, classifyChargeCategoryAzure)
	nz := newNormalizer(idx)
	row := basicRow(h)
	row[colIndex(h, "ChargeType")] = tokenUsageDisc
	row[colIndex(h, "AmortizedCost")] = "-0.5"
	fr, err := fm.MapFields(row)
	if err != nil {
		t.Fatalf("map fields: %v", err)
	}
	cf.Classify(row, &fr)
	nz.Normalize(row, &fr)
	if fr.ChargeCategory != CategoryDiscount {
		t.Fatalf("expected Discount, got %s", fr.ChargeCategory)
	}
}
