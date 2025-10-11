package drift

import "testing"

func TestRunBasic(t *testing.T) {
	baseCharge := Distribution{"Compute": 90, "Storage": 10}
	curCharge := Distribution{"Compute": 80, "Storage": 20}
	basePricing := Distribution{"OnDemand": 100}
	curPricing := Distribution{"OnDemand": 100}
	baseBuckets := CostBuckets{EffectiveCounts: map[string]int64{"1-10": 1}, UsageCounts: map[string]int64{"1-10": 1}}
	curBuckets := CostBuckets{EffectiveCounts: map[string]int64{"1-10": 1}, UsageCounts: map[string]int64{"1-10": 1}}
	rep, err := Run(Config{Alpha: 0.05}, baseCharge, curCharge, basePricing, curPricing, baseBuckets, curBuckets, nil, Snapshot{RowCount: 100, SumEffective: 123.4, SumUsage: 55}, []float64{1, 2, 3}, []float64{1.1, 2.2, 3.3}, []float64{10, 20, 30}, []float64{11, 19, 31})
	if err != nil {
		t.Fatalf("run error: %v", err)
	}
	if rep.ChiSquare.DegreesOfFreedom == 0 {
		t.Fatalf("expected dof >0")
	}
	if rep.Trend.EffectiveSlope != 0 {
		t.Fatalf("expected zero slope with single point")
	}
}

func TestPercentileThresholdExceeded(t *testing.T) {
	baseCharge := Distribution{"A": 100}
	curCharge := Distribution{"A": 100}
	basePricing := Distribution{"X": 100}
	curPricing := Distribution{"X": 100}
	baseBuckets := CostBuckets{EffectiveCounts: map[string]int64{"<1": 1}, UsageCounts: map[string]int64{"<1": 1}}
	curBuckets := baseBuckets
	baseEff := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	curEff := []float64{10, 11, 12, 13, 14, 15, 16, 17, 18, 19}
	baseUse := []float64{1, 1, 1, 1, 1, 1, 1, 1, 1, 1}
	curUse := []float64{2, 2, 2, 2, 2, 2, 2, 2, 2, 2}
	rep, err := Run(Config{Alpha: 0.05, Percentiles: []float64{0.5, 0.9}, PercentileDriftThreshold: 0.01}, baseCharge, curCharge, basePricing, curPricing, baseBuckets, curBuckets, nil, Snapshot{RowCount: 10}, baseEff, curEff, baseUse, curUse)
	if err != nil {
		t.Fatalf("run error: %v", err)
	}
	if !rep.Percentiles.ThresholdExceeded {
		t.Fatalf("expected percentile threshold exceeded")
	}
	if rep.Percentiles.MaxEffectiveDelta <= 0 {
		t.Fatalf("expected positive max effective delta")
	}
}
