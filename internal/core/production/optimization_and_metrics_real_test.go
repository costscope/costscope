package production

import (
	"context"
	"local/costscope/internal/core/logging"
	"testing"
)

// Test real optimization engine Analyze with all categories and aggressive flag.
func TestBasicOptimizationEngine_Analyze_AllCategoriesAggressive(t *testing.T) {
	eng := NewBasicOptimizationEngine(logging.NewLogger(logging.LevelError))
	opts := &OptimizationOptions{Categories: []string{"performance", "security", "cost", "integration"}, Aggressive: true}
	res, err := eng.AnalyzeOptimizations(context.Background(), opts)
	if err != nil {
		t.Fatalf("analyze err: %v", err)
	}
	if res.TotalImprovements == 0 || res.OptimizationScore == 0 {
		t.Fatalf("expected improvements and score")
	}
}

// Test GenerateRecommendations branches with metrics triggering all conditional additions (performance<80, security<85, readiness >70, integration<80).
func TestBasicOptimizationEngine_GenerateRecommendations_AllBranches(t *testing.T) {
	eng := NewBasicOptimizationEngine(logging.NewLogger(logging.LevelError))
	metrics := &ProductionSystemMetrics{
		Performance:    PerformanceMetrics{OptimizationScore: 70},
		Security:       SecurityMetrics{SecurityScore: 80},
		Integration:    IntegrationMetrics{IntegrationScore: 70},
		ReadinessScore: 75,
		CriticalIssues: nil,
	}
	recs, err := eng.GenerateRecommendations(context.Background(), metrics)
	if err != nil {
		t.Fatalf("gen recs err: %v", err)
	}
	// Expect at least 4 recommendations given all branch triggers.
	if len(recs) < 4 {
		t.Fatalf("expected >=4 recommendations, got %d", len(recs))
	}
}

// Test CalculateROI with non-empty recommendations to cover savings calculations.
func TestBasicOptimizationEngine_CalculateROI_NonEmpty(t *testing.T) {
	eng := NewBasicOptimizationEngine(logging.NewLogger(logging.LevelError))
	recs := []ProductionRecommendation{
		{ID: "r1", Type: "performance", Cost: 1000, ROI: 150},
		{ID: "r2", Type: "security", Cost: 500, ROI: 200},
	}
	roi, err := eng.CalculateROI(context.Background(), recs)
	if err != nil {
		t.Fatalf("roi err: %v", err)
	}
	if roi.TotalInvestment <= 0 || roi.ProjectedSavings <= 0 || roi.ROIPercentage == 0 {
		t.Fatalf("expected populated ROI analysis")
	}
	if len(roi.SavingsBreakdown) < 2 {
		t.Fatalf("expected breakdown entries")
	}
}

// Test metrics collector real implementations to raise coverage (each returns non-nil).
func TestBasicMetricsCollector_All(t *testing.T) {
	mc := NewBasicMetricsCollector(nil, logging.NewLogger(logging.LevelError))
	if h, err := mc.CollectSystemHealth(context.Background()); err != nil || h == nil {
		t.Fatalf("health err %v", err)
	}
	if p, err := mc.CollectPerformanceMetrics(context.Background()); err != nil || p == nil {
		t.Fatalf("perf err %v", err)
	}
	if s, err := mc.CollectSecurityMetrics(context.Background()); err != nil || s == nil {
		t.Fatalf("sec err %v", err)
	}
	if i, err := mc.CollectIntegrationMetrics(context.Background()); err != nil || i == nil {
		t.Fatalf("integ err %v", err)
	}
	if a, err := mc.CollectAnalyticsMetrics(context.Background()); err != nil || a == nil {
		t.Fatalf("analytics err %v", err)
	}
}
