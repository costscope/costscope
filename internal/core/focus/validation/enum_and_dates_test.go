package validation

import "testing"

func TestQuality_EnumInvalidValues(t *testing.T) {
	v := NewQualityValidator()
	res := QualityAssessmentResult{
		Valid:      true,
		Score:      100,
		Statistics: make(map[string]ColumnStatistics),
		ValueSamples: map[string][]string{
			"PricingCategory": {"Standard", "UnknownTier"},
			"ProviderName":    {"Amazon Web Services", "SomeCloud"},
			"ChargeCategory":  {"Usage", "Weird"},
		},
	}
	v.checkEnumValues(&res)
	if res.Valid {
		t.Fatalf("expected invalid result when enums contain bad values")
	}
	found := 0
	for _, is := range res.Issues {
		if is.Type == "enum_invalid_value" {
			found++
		}
	}
	if found < 1 {
		t.Fatalf("expected enum_invalid_value issues, found %d", found)
	}
}

func TestQuality_DateRangeViolation(t *testing.T) {
	v := NewQualityValidator()
	res := QualityAssessmentResult{
		Valid:      true,
		Score:      100,
		Statistics: make(map[string]ColumnStatistics),
	}
	res.Statistics["BillingPeriodStart"] = ColumnStatistics{Mean: "2024-12-31T00:00:00Z"}
	res.Statistics["BillingPeriodEnd"] = ColumnStatistics{Mean: "2024-01-01T00:00:00Z"}
	v.checkDateRanges(&res)
	if res.Valid {
		t.Fatalf("expected res.Valid=false due to date range violation")
	}
}
