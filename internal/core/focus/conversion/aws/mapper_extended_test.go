package aws

import (
	"testing"
	"time"
)

// Common test constants to satisfy goconst linter
const (
	liTypeSavingsPlanCovered = "SavingsPlanCoveredUsage"
	liTypeDiscountedUsage    = "DiscountedUsage"
	liTypeCredit             = "Credit"
	liTypeTax                = "Tax"
	usageUnitHours           = "Hours"
	regionInferred           = "us-east-1"
)

// buildTestHeaders returns a representative AWS CUR header slice for tests.
func buildTestHeadersExt() []string {
	return []string{
		"bill/BillingAccountId",
		"bill/BillingAccountName",
		"bill/BillingCurrency",
		"lineItem/UnblendedCost",
		"lineItem/UsageAmount",
		"lineItem/UsageStartDate",
		"lineItem/UsageEndDate",
		"lineItem/LineItemDescription",
		"lineItem/Operation",
		"lineItem/UsageType",
		"product/ProductName",
		"product/ProductFamily",
		"lineItem/LineItemType",
		"pricing/PriceId",
		"lineItem/UsageAccountId",
		"lineItem/ResourceId",
		"lineItem/AvailabilityZone",
		"product/Region",
		"savingsPlan/SavingsPlanArn",
		"savingsPlan/SavingsPlanId",
		"reservation/ReservationARN",
		"reservation/SubscriptionId",
	}
}

// helper to create a base row with defaults; callers mutate indices as needed.
func baseRowExt(h []string) []string {
	row := make([]string, len(h))
	set := func(name, val string) { // assign if exists
		for i, col := range h {
			if col == name {
				row[i] = val
				return
			}
		}
	}
	set("bill/BillingAccountId", "ba-1")
	set("bill/BillingAccountName", "Acct")
	set("bill/BillingCurrency", "usd")
	set("lineItem/UnblendedCost", "10")
	set("lineItem/UsageAmount", "5")
	set("lineItem/UsageStartDate", "2024-01-01 00:00:00")
	set("lineItem/UsageEndDate", "2024-01-01 01:00:00")
	set("lineItem/LineItemDescription", "desc")
	set("lineItem/Operation", "RunInstances")
	set("lineItem/UsageType", "Hours")
	set("product/ProductName", "AmazonEC2")
	set("product/ProductFamily", "Compute")
	set("lineItem/LineItemType", "OnDemand")
	set("pricing/PriceId", "p-1")
	set("lineItem/UsageAccountId", "123456789012")
	set("lineItem/ResourceId", "i-abc")
	set("lineItem/AvailabilityZone", "us-east-1a")
	set("product/Region", "") // blank to test inference
	return row
}

func newMapperExt(t *testing.T, unified bool) (RowMapper, *HeaderIndex, []string) {
	t.Helper()
	headers := buildTestHeadersExt()
	idx, err := NewHeaderIndex(headers)
	if err != nil {
		t.Fatalf("header index: %v", err)
	}
	mapper := NewRowMapper(idx, "test.csv", unified)
	return mapper, idx, headers
}

func TestAWSRowMapper_OnDemandBasic(t *testing.T) {
	mapper, _, headers := newMapperExt(t, false)
	row := baseRowExt(headers)
	fr, err := mapper.Map(row)
	if err != nil {
		t.Fatalf("map: %v", err)
	}
	if fr.EffectiveCost != 10 || fr.UsageQuantity != 5 {
		t.Errorf("unexpected cost/usage: %v/%v", fr.EffectiveCost, fr.UsageQuantity)
	}
	if fr.Region == nil || *fr.Region != regionInferred {
		t.Errorf("expected region inference %s got %v", regionInferred, fr.Region)
	}
}

func TestAWSRowMapper_SavingsPlanCoveredUsage(t *testing.T) {
	mapper, idx, headers := newMapperExt(t, false)
	row := baseRowExt(headers)
	row[idx.ILineItemType] = liTypeSavingsPlanCovered
	row[idx.ISPId] = "sp-123"
	fr, err := mapper.Map(row)
	if err != nil {
		t.Fatalf("map: %v", err)
	}
	if fr.CommitmentDiscountId == nil || *fr.CommitmentDiscountId != "sp-123" {
		t.Errorf("expected SP discount id sp-123 got %v", fr.CommitmentDiscountId)
	}
}

func TestAWSRowMapper_RI_DiscountedUsage(t *testing.T) {
	mapper, idx, headers := newMapperExt(t, false)
	row := baseRowExt(headers)
	row[idx.ILineItemType] = liTypeDiscountedUsage
	row[idx.IRIArn] = "ri-arn"
	fr, err := mapper.Map(row)
	if err != nil {
		t.Fatalf("map: %v", err)
	}
	if fr.PricingCategory != "Reserved" {
		t.Errorf("expected Reserved got %s", fr.PricingCategory)
	}
	if fr.CommitmentDiscountType == nil {
		t.Errorf("expected RI commitment type")
	}
}

func TestAWSRowMapper_CreditClassification(t *testing.T) {
	mapper, idx, headers := newMapperExt(t, false)
	row := baseRowExt(headers)
	row[idx.ILineItemType] = liTypeCredit
	fr, err := mapper.Map(row)
	if err != nil {
		t.Fatalf("map: %v", err)
	}
	if fr.ChargeCategory != liTypeCredit {
		t.Errorf("expected Credit got %s", fr.ChargeCategory)
	}
}

func TestAWSRowMapper_TaxClassification(t *testing.T) {
	mapper, idx, headers := newMapperExt(t, false)
	row := baseRowExt(headers)
	row[idx.ILineItemType] = liTypeTax
	fr, err := mapper.Map(row)
	if err != nil {
		t.Fatalf("map: %v", err)
	}
	if fr.ChargeCategory != liTypeTax {
		t.Errorf("expected Tax got %s", fr.ChargeCategory)
	}
}

func TestAWSRowMapper_SpotDetection(t *testing.T) {
	mapper, idx, headers := newMapperExt(t, false)
	row := baseRowExt(headers)
	row[idx.IUsageType] = "SpotUsage" + usageUnitHours
	fr, err := mapper.Map(row)
	if err != nil {
		t.Fatalf("map: %v", err)
	}
	if fr.PricingCategory != "Spot" {
		t.Errorf("expected Spot got %s", fr.PricingCategory)
	}
}

func TestAWSRowMapper_DivideByZeroUsage(t *testing.T) {
	mapper, idx, headers := newMapperExt(t, false)
	row := baseRowExt(headers)
	row[idx.IUsageAmount] = "0" // zero usage
	fr, err := mapper.Map(row)
	if err != nil {
		t.Fatalf("map: %v", err)
	}
	if fr.ListUnitPrice != 0 {
		t.Errorf("expected 0 unit price got %v", fr.ListUnitPrice)
	}
}

func TestAWSRowMapper_TooShortRowError(t *testing.T) {
	mapper, _, _ := newMapperExt(t, false)
	if _, err := mapper.Map([]string{"only", "two"}); err == nil {
		t.Fatalf("expected error for short row")
	}
}

func TestAWSRowMapper_UnifiedNormalization(t *testing.T) {
	mapper, idx, headers := newMapperExt(t, true)
	row := baseRowExt(headers)
	row[idx.IBillingCurrency] = "usd" // lower, unified path should upper
	fr, err := mapper.Map(row)
	if err != nil {
		t.Fatalf("map: %v", err)
	}
	if fr.BillingCurrency != "USD" {
		t.Errorf("expected USD got %s", fr.BillingCurrency)
	}
	if fr.UsageUnit != usageUnitHours {
		t.Errorf("expected canonical %s got %s", usageUnitHours, fr.UsageUnit)
	}
	if fr.BillingPeriodEnd.Before(fr.BillingPeriodStart) {
		t.Errorf("invalid period ordering")
	}
	if time.Since(fr.BillingPeriodStart) <= 0 {
		t.Logf("period start is in future which is unexpected for test clock")
	}
}
