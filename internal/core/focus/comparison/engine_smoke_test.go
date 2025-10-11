package comparison_test

import (
	"testing"
	"time"

	comparison "local/costscope/internal/core/focus/comparison"
	"local/costscope/internal/core/logging"
)

func newCompLogger() *logging.Logger { return logging.NewLogger(logging.LevelInfo) }

// synth creates a minimal slice of FOCUSRecord for grouping by service/region
func synth(n int, service, region, acct string, start time.Time, stepDays int, baseCost float64) []comparison.FOCUSRecord {
	out := make([]comparison.FOCUSRecord, 0, n)
	for i := 0; i < n; i++ {
		d := start.AddDate(0, 0, i*stepDays)
		out = append(out, comparison.FOCUSRecord{
			BillingAccountID:   acct,
			BillingAccountName: "acct",
			BillingCurrency:    "USD",
			BillingPeriodStart: d,
			BillingPeriodEnd:   d.Add(24 * time.Hour),
			ServiceName:        service,
			Region:             region,
			UsageQuantity:      1,
			UsageUnit:          "Hours",
			BilledCost:         baseCost + float64(i),
		})
	}
	return out
}

// Smoke test: CompareFOCUSDatasets runs on synthetic inputs (mock loader path uses filename only)
func TestComparisonEngine_CompareFOCUSDatasets_Smoke(t *testing.T) {
	eng := comparison.NewEngine(newCompLogger(), nil)

	// The current implementation of loadFOCUSDataset returns empty slice when filename is non-empty.
	// To still exercise downstream methods with data, call lower-level helpers directly below.
	if _, err := eng.CompareFOCUSDatasets("baseline.parquet", "current.parquet", comparison.DiffOptions{Dimensions: []string{"service", "region"}}); err != nil {
		// CompareFOCUSDatasets may fail only if filenames are empty; ensure it does not when provided.
		t.Fatalf("CompareFOCUSDatasets error: %v", err)
	}
}

// Smoke test: exercise DetectCostChanges/IdentifyServiceChanges/AnalyzeTrends/DetectAnomalies/GenerateForecast
func TestComparisonEngine_Methods_Smoke(t *testing.T) {
	eng := comparison.NewEngine(newCompLogger(), nil)

	start := time.Now().AddDate(0, 0, -14)
	// Use different lengths to exercise code paths and avoid unparam lint on synth(n)
	baseline := append(
		synth(5, "EC2", "us-east-1", "123", start, 2, 100),
		synth(7, "S3", "eu-west-1", "456", start, 1, 50)...,
	)
	current := append(
		synth(7, "EC2", "us-east-1", "123", start, 1, 130),      // increase
		synth(7, "Lambda", "us-east-1", "456", start, 1, 10)..., // new service
	)

	opts := comparison.DiffOptions{Dimensions: []string{"service", "region"}, Threshold: 1.0, SignificanceLevel: 0.1, ShowAnomalies: true, ShowTrends: true}

	if _, err := eng.DetectCostChanges(baseline, current, opts); err != nil {
		t.Fatalf("DetectCostChanges error: %v", err)
	}
	if _, _, err := eng.IdentifyServiceChanges(baseline, current); err != nil {
		t.Fatalf("IdentifyServiceChanges error: %v", err)
	}
	if _, err := eng.AnalyzeTrends(baseline, current, opts); err != nil {
		t.Fatalf("AnalyzeTrends error: %v", err)
	}
	if _, err := eng.DetectAnomalies(current, opts); err != nil {
		t.Fatalf("DetectAnomalies error: %v", err)
	}
	if _, err := eng.GenerateForecast(current, 3); err != nil {
		t.Fatalf("GenerateForecast error: %v", err)
	}
}
