package drift

import "testing"

func TestSeverityAndCustomSchema(t *testing.T) {
	baseCharge := Distribution{"A": 1000, "B": 1000}
	curCharge := Distribution{"A": 2000, "B": 0}
	// pricing identical to isolate effect
	basePricing := Distribution{"X": 100}
	curPricing := Distribution{"X": 100}
	// buckets with custom schema
	effBase := []float64{0.05, 0.5, 5, 50, 500}
	effCur := []float64{0.05, 0.5, 5, 50, 500, 5000}
	baseBuckets, _ := BuildCostBuckets(effBase, effBase, []float64{0.1, 1, 10, 100, 1000})
	curBuckets, _ := BuildCostBuckets(effCur, effCur, []float64{0.1, 1, 10, 100, 1000})
	rep, err := Run(Config{Alpha: 0.05, BucketSchema: []float64{0.1, 1, 10, 100, 1000}}, baseCharge, curCharge, basePricing, curPricing, baseBuckets, curBuckets, nil, Snapshot{RowCount: 1}, nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("run error: %v", err)
	}
	if rep.ChiSquare.Severity != "high" {
		t.Fatalf("expected high severity, got %s (p=%f)", rep.ChiSquare.Severity, rep.ChiSquare.PValue)
	}
	if len(rep.CostBucketDeltas.Effective) == 0 {
		t.Fatalf("expected bucket deltas with custom schema")
	}
}
