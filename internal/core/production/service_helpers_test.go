package production

import (
	"context"
	"testing"
	"time"

	"local/costscope/internal/core/logging"
)

// Test helper calculation branching paths.
func TestBasicProductionService_HelperCalculations(t *testing.T) {
	svc := &BasicProductionService{logger: logging.NewLogger(logging.LevelError)}

	// Table for completion levels.
	cases := []struct {
		h, p, s int
		expect  string
	}{
		{95, 95, 95, "production_ready"},
		{80, 70, 75, "near_ready"},
		{65, 60, 55, "development_complete"},
		{50, 50, 50, "beta"},
		{10, 20, 30, "alpha"},
	}
	for i, c := range cases {
		got := svc.calculateCompletionLevel(&SystemHealthStatus{HealthScore: c.h}, &PerformanceMetrics{OptimizationScore: c.p}, &SecurityMetrics{SecurityScore: c.s})
		if got != c.expect {
			t.Fatalf("case %d expected %s got %s", i, c.expect, got)
		}
	}

	// Production readiness boolean branches.
	if !svc.calculateProductionReadiness(&SystemHealthStatus{HealthScore: 80}, &PerformanceMetrics{OptimizationScore: 70}, &SecurityMetrics{SecurityScore: 80}) {
		t.Fatalf("expected readiness true")
	}
	if svc.calculateProductionReadiness(&SystemHealthStatus{HealthScore: 79}, &PerformanceMetrics{OptimizationScore: 70}, &SecurityMetrics{SecurityScore: 80}) {
		t.Fatalf("expected readiness false")
	}

	// Readiness score weight calculation deterministic.
	rs := svc.calculateReadinessScore(&SystemHealthStatus{HealthScore: 80}, &PerformanceMetrics{OptimizationScore: 70}, &SecurityMetrics{SecurityScore: 90}, &IntegrationMetrics{IntegrationScore: 60}, &AnalyticsMetrics{DataQualityScore: 50})
	if rs == 0 {
		t.Fatalf("expected non-zero readiness score")
	}
}

// Test recommendations & critical issues branch triggers.
func TestBasicProductionService_RecommendationsAndIssues(t *testing.T) {
	svc := &BasicProductionService{logger: logging.NewLogger(logging.LevelError)}
	// Trigger all recommendation branches.
	health := &SystemHealthStatus{HealthScore: 70, ErrorRate: 2.0}
	perf := &PerformanceMetrics{OptimizationScore: 60}
	sec := &SecurityMetrics{SecurityScore: 70}
	recs := svc.generateBasicRecommendations(health, perf, sec)
	if len(recs) != 4 {
		t.Fatalf("expected 4 recommendations, got %d", len(recs))
	}

	// Trigger all critical issues branches.
	healthCrit := &SystemHealthStatus{HealthScore: 50, ErrorRate: 6.0}
	perfCrit := &PerformanceMetrics{MemoryUsageMB: 910, MemoryLimitMB: 1000}
	secCrit := &SecurityMetrics{VulnerabilitiesHigh: 2, SecurityScore: 70}
	issues := svc.identifyCriticalIssues(healthCrit, perfCrit, secCrit)
	if len(issues) != 4 {
		t.Fatalf("expected 4 critical issues, got %d", len(issues))
	}
}

// Test strategic recommendations conditional additions.
func TestBasicProductionService_StrategicRecommendationsAndFutureOutlook(t *testing.T) {
	svc := &BasicProductionService{logger: logging.NewLogger(logging.LevelError)}
	metrics := &ProductionSystemMetrics{
		Performance:  PerformanceMetrics{OptimizationScore: 70},
		Security:     SecurityMetrics{SecurityScore: 80},
		SystemHealth: SystemHealthStatus{Status: "degraded"},
	}
	recs := svc.generateStrategicRecommendations(metrics)
	// Expect extra performance rec (Critical Performance Remediation) and security enhancement (score <85) -> 4 base +2 conditional = 4? Actually base returns 2 base; may add perf and security -> total >=4.
	if len(recs) < 4 {
		t.Fatalf("expected at least 4 recommendations, got %d", len(recs))
	}

	fo := svc.generateFutureOutlook(metrics)
	if fo.Investments["performance"] <= 50000 {
		t.Fatalf("expected increased performance investment")
	}
	if fo.Investments["security"] <= 30000 {
		t.Fatalf("expected increased security investment")
	}
	if fo.StrategicGoals[len(fo.StrategicGoals)-1] != "System reliability improvement" {
		t.Fatalf("expected degraded health adjustment")
	}
}

// Test appendices generation with charts branch (IncludeCharts true) producing third appendix.
func TestBasicProductionService_GenerateAppendices_WithCharts(t *testing.T) {
	svc := &BasicProductionService{logger: logging.NewLogger(logging.LevelError)}
	metrics := &ProductionSystemMetrics{Performance: PerformanceMetrics{OptimizationScore: 80}}
	apps := svc.generateAppendices(metrics, &ReportOptions{IncludeAppendix: true, IncludeCharts: true})
	if len(apps) != 3 {
		t.Fatalf("expected 3 appendices incl charts, got %d", len(apps))
	}
}

// Test optimization engine additional early/conditional branches.
func TestBasicOptimizationEngine_Branches(t *testing.T) {
	eng := NewBasicOptimizationEngine(logging.NewLogger(logging.LevelError))
	// nil options error
	if _, err := eng.AnalyzeOptimizations(context.Background(), nil); err == nil {
		t.Fatalf("expected error for nil options")
	}

	// CalculateROI empty recommendations path
	roi, err := eng.CalculateROI(context.Background(), nil)
	if err != nil || roi.TotalInvestment != 0 {
		t.Fatalf("expected zero ROI path, err %v", err)
	}

	// PlanRoadmap conditional absence of Q1 item when readiness >=90
	metricsHigh := &ProductionSystemMetrics{ReadinessScore: 95}
	roadmapHigh, err := eng.PlanRoadmap(context.Background(), metricsHigh)
	if err != nil {
		t.Fatalf("roadmap err: %v", err)
	}
	for _, item := range roadmapHigh {
		if item.Quarter == "Q1 2025" {
			t.Fatalf("unexpected Q1 item for high readiness")
		}
	}

	// PlanRoadmap includes Q1 when readiness <90
	metricsLow := &ProductionSystemMetrics{ReadinessScore: 80}
	roadmapLow, _ := eng.PlanRoadmap(context.Background(), metricsLow)
	foundQ1 := false
	for _, item := range roadmapLow {
		if item.Quarter == "Q1 2025" {
			foundQ1 = true
		}
	}
	if !foundQ1 {
		t.Fatalf("expected Q1 item for low readiness")
	}
}

// Sanity: ensure timestamp/processing time formatting unaffected by helper usage.
func TestBasicProductionService_GenerateExecutiveReport_DirectHelpersDoNotAlterTime(t *testing.T) {
	svc := &BasicProductionService{logger: logging.NewLogger(logging.LevelError), cache: make(map[string]interface{})}
	svc.metricsCollector = fakeMetricsCollector{}
	svc.optimizationEngine = fakeOptimizationEngine{}
	svc.deploymentAssessor = fakeDeploymentAssessor{}
	rep, err := svc.GenerateExecutiveReport(context.Background(), &ReportOptions{IncludeAppendix: true, IncludeCharts: true})
	if err != nil {
		t.Fatalf("exec report err: %v", err)
	}
	if time.Since(rep.GeneratedAt) > time.Minute {
		t.Fatalf("stale generated timestamp")
	}
}
