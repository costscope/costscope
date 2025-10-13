package aws

import (
	"testing"
	"time"

	ftypes "github.com/costscope/costscope/internal/core/focus/types"
)

// Test MapRowFast for a minimal synthetic row covering core fields
func TestMapRowFast_CoreFields(t *testing.T) {
	idx := RowIndexes{
		BillingAccountId:   0,
		BillingAccountName: 1,
		BillingCurrency:    2,
		UnblendedCost:      3,
		UsageAmount:        4,
		UsageStartDate:     5,
		UsageEndDate:       6,
		LineItemDesc:       7,
		Operation:          8,
		UsageType:          9,
		ProductName:        10,
		ProductFamily:      11,
		LineItemType:       12,
		PriceId:            13,
		UsageAccountId:     14,
		ResourceId:         15,
		AvailabilityZone:   16,
		Region:             17,
	}
	row := []string{
		"ba-123",              // 0
		"BillingName",         // 1
		"usd",                 // 2
		"12.50",               // 3 UnblendedCost
		"2.5",                 // 4 UsageAmount
		"2024-01-01 00:00:00", // 5 UsageStart
		"2024-01-01 01:00:00", // 6 UsageEnd
		"EC2 usage",           // 7 Desc
		"RunInstances",        // 8 Operation
		"Hrs",                 // 9 UsageType
		"AmazonEC2",           // 10 ProductName
		"Compute",             // 11 ProductFamily
		"OnDemand",            // 12 LineItemType
		"price-1",             // 13 PriceId
		"111111111111",        // 14 UsageAccountId
		"i-abc",               // 15 ResourceId
		"us-east-1a",          // 16 AZ
		"us-east-1",           // 17 Region
	}

	fr, err := MapRowFast(idx, row)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if fr.BillingAccountId != "ba-123" {
		t.Errorf("BillingAccountId: want ba-123 got %s", fr.BillingAccountId)
	}
	if fr.BillingCurrency != "usd" {
		t.Errorf("BillingCurrency: want usd got %s", fr.BillingCurrency)
	}
	if fr.EffectiveCost != 12.50 {
		t.Errorf("EffectiveCost: want 12.50 got %v", fr.EffectiveCost)
	}
	if fr.UsageQuantity != 2.5 {
		t.Errorf("UsageQuantity: want 2.5 got %v", fr.UsageQuantity)
	}
	if fr.ProviderName != ftypes.ProviderNames.AWS {
		t.Errorf("ProviderName: want %s got %s", ftypes.ProviderNames.AWS, fr.ProviderName)
	}
	if fr.SourceProvider != "aws" {
		t.Errorf("SourceProvider: want aws got %s", fr.SourceProvider)
	}
	if fr.ServiceName != "AmazonEC2" {
		t.Errorf("ServiceName: want AmazonEC2 got %s", fr.ServiceName)
	}
	if fr.SubAccountId != "111111111111" {
		t.Errorf("SubAccountId: want 111111111111 got %s", fr.SubAccountId)
	}
	if fr.ResourceId != "i-abc" {
		t.Errorf("ResourceId: want i-abc got %s", fr.ResourceId)
	}
	if fr.Region == nil || *fr.Region != "us-east-1" {
		t.Errorf("Region: want us-east-1 got %v", fr.Region)
	}
	if fr.AvailabilityZone == nil || *fr.AvailabilityZone != "us-east-1a" {
		t.Errorf("AZ: want us-east-1a got %v", fr.AvailabilityZone)
	}

	// Date parse sanity
	if fr.BillingPeriodStart.IsZero() || fr.BillingPeriodEnd.IsZero() {
		t.Errorf("expected non-zero period start/end")
	}
	if fr.BillingPeriodEnd.Before(fr.BillingPeriodStart) {
		t.Errorf("end before start: %s < %s", fr.BillingPeriodEnd.Format(time.RFC3339), fr.BillingPeriodStart.Format(time.RFC3339))
	}
}
