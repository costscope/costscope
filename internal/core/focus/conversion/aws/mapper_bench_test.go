package aws

import (
	"fmt"
	"strconv"
	"testing"
)

const benchRegion = "us-east-1"

// buildDemoHeaders returns a minimal realistic AWS CUR header set used by the mapper
func buildDemoHeaders() []string {
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
	}
}

// buildDemoRecords synthesizes N rows matching headers
func buildDemoRecords(n int) [][]string {
	headers := buildDemoHeaders()
	recs := make([][]string, n)
	for i := 0; i < n; i++ {
		r := make([]string, len(headers))
		r[0] = "123456789012"
		r[1] = "DemoAccount"
		r[2] = "USD"
		r[3] = fmt.Sprintf("%d", 100+i%10) // UnblendedCost
		r[4] = fmt.Sprintf("%d", 50+(i%5)) // UsageAmount
		r[5] = "2025-07-31 00:00:00"       // UsageStartDate
		r[6] = "2025-07-31 01:00:00"       // UsageEndDate
		r[7] = "LineItemDescription"
		r[8] = "RunInstances"
		r[9] = "BoxUsage:t3.micro"
		r[10] = "Amazon Elastic Compute Cloud"
		r[11] = "Compute Instance"
		r[12] = "Usage"
		r[13] = "PRICE123"
		r[14] = "123456789012"
		r[15] = "i-abcdef" + strconv.Itoa(i)
		r[16] = benchRegion + "a"
		r[17] = benchRegion
		recs[i] = r
	}
	return recs
}

func BenchmarkAWSMapper_LegacyFastPath(b *testing.B) {
	headers := buildDemoHeaders()
	records := buildDemoRecords(10_000)
	idx, err := NewHeaderIndex(headers)
	if err != nil {
		b.Fatalf("index error: %v", err)
	}
	mapper := NewRowMapper(idx, "bench.csv", false)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var mapped int
		for _, r := range records {
			if _, err := mapper.Map(r); err == nil {
				mapped++
			}
		}
		if mapped != len(records) {
			b.Fatalf("mapped mismatch: %d", mapped)
		}
	}
}

func BenchmarkAWSMapper_UnifiedPath(b *testing.B) {
	headers := buildDemoHeaders()
	records := buildDemoRecords(10_000)
	idx, err := NewHeaderIndex(headers)
	if err != nil {
		b.Fatalf("index error: %v", err)
	}
	mapper := NewRowMapper(idx, "bench.csv", true)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var mapped int
		for _, r := range records {
			if _, err := mapper.Map(r); err == nil {
				mapped++
			}
		}
		if mapped != len(records) {
			b.Fatalf("mapped mismatch: %d", mapped)
		}
	}
}
