package validation

import "testing"

// TestNormalizationRecommendations ensures the quality validator emits normalization
// recommendations for currency, region, and unit when those columns are present.
func TestNormalizationRecommendations(t *testing.T) {
	v := NewQualityValidator()
	res := QualityAssessmentResult{
		Valid:          true,
		Score:          100,
		MissingValues:  map[string]int64{},
		DataTypes:      map[string]string{"RegionName": "string", "BillingCurrency": "string", "UsageUnit": "string"},
		UniqueValues:   map[string]int64{},
		NullPercentage: map[string]float64{},
		Statistics:     map[string]ColumnStatistics{},
		Issues:         []QualityIssue{},
	}
	v.checkAndNormalizeDictionaries(&res)
	found := map[string]bool{"RegionName": false, "BillingCurrency": false, "UsageUnit": false}
	for _, is := range res.Issues {
		if is.Type == "normalization_recommendation" {
			if _, ok := found[is.Column]; ok {
				found[is.Column] = true
			}
		}
	}
	for col, ok := range found {
		if !ok {
			t.Fatalf("expected normalization recommendation for %s", col)
		}
	}
}
