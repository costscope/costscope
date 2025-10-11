package production

import (
	"context"
	"fmt"
	"testing"
	"time"

	"local/costscope/internal/core/logging"
)

// Error-inducing optimization engine variants to exercise distinct error branches.
type errAnalyzeEngine struct{ fakeOptimizationEngine }

func (errAnalyzeEngine) AnalyzeOptimizations(context.Context, *OptimizationOptions) (*OptimizationResults, error) {
	return nil, fmt.Errorf("analyze boom")
}

type errGenerateEngine struct{ fakeOptimizationEngine }

func (errGenerateEngine) GenerateRecommendations(context.Context, *ProductionSystemMetrics) ([]ProductionRecommendation, error) {
	return nil, fmt.Errorf("gen boom")
}

type errROIEngine struct{ fakeOptimizationEngine }

func (errROIEngine) CalculateROI(context.Context, []ProductionRecommendation) (*ROIAnalysis, error) {
	return nil, fmt.Errorf("roi boom")
}

type errRoadmapEngine struct{ fakeOptimizationEngine }

func (errRoadmapEngine) PlanRoadmap(context.Context, *ProductionSystemMetrics) ([]RoadmapItem, error) {
	return nil, fmt.Errorf("roadmap boom")
}

// Engine that forces ROI fallback by returning empty recommendations then nil ROI pointer.
type fallbackROIEngine struct{}

func (fallbackROIEngine) AnalyzeOptimizations(context.Context, *OptimizationOptions) (*OptimizationResults, error) {
	return &OptimizationResults{}, nil
}
func (fallbackROIEngine) GenerateRecommendations(context.Context, *ProductionSystemMetrics) ([]ProductionRecommendation, error) {
	return nil, nil
}
func (fallbackROIEngine) CalculateROI(context.Context, []ProductionRecommendation) (*ROIAnalysis, error) {
	return nil, fmt.Errorf("roi calc failed")
}
func (fallbackROIEngine) PlanRoadmap(context.Context, *ProductionSystemMetrics) ([]RoadmapItem, error) {
	return nil, nil
}

// Test AssessDeploymentReadiness happy path (previously uncovered service method).
func TestBasicProductionService_AssessDeploymentReadiness(t *testing.T) {
	logger := logging.NewLogger(logging.LevelError)
	svc := &BasicProductionService{logger: logger, cache: make(map[string]interface{})}
	svc.metricsCollector = fakeMetricsCollector{}
	svc.optimizationEngine = fakeOptimizationEngine{}
	svc.deploymentAssessor = fakeDeploymentAssessor{}
	assess, err := svc.AssessDeploymentReadiness(context.Background(), "production")
	if err != nil {
		t.Fatalf("assess readiness err: %v", err)
	}
	if assess.ReadinessScore == 0 || assess.ReadinessStatus == "" {
		t.Fatalf("expected readiness fields populated")
	}
	if assess.ProcessingTimeMs < 0 { // should never be negative
		t.Fatalf("invalid processing time")
	}
}

// Test RunOptimization error path: AnalyzeOptimizations fails.
func TestBasicProductionService_RunOptimization_ErrorAnalyze(t *testing.T) {
	logger := logging.NewLogger(logging.LevelError)
	svc := &BasicProductionService{logger: logger, cache: make(map[string]interface{})}
	svc.metricsCollector = fakeMetricsCollector{}
	svc.optimizationEngine = errAnalyzeEngine{}
	svc.deploymentAssessor = fakeDeploymentAssessor{}
	if _, err := svc.RunOptimization(context.Background(), &OptimizationOptions{Categories: []string{"performance"}}); err == nil {
		t.Fatalf("expected analyze error")
	}
}

// Test RunOptimization error path: GenerateRecommendations fails.
func TestBasicProductionService_RunOptimization_ErrorGenerate(t *testing.T) {
	logger := logging.NewLogger(logging.LevelError)
	svc := &BasicProductionService{logger: logger, cache: make(map[string]interface{})}
	svc.metricsCollector = fakeMetricsCollector{}
	svc.optimizationEngine = errGenerateEngine{}
	svc.deploymentAssessor = fakeDeploymentAssessor{}
	if _, err := svc.RunOptimization(context.Background(), &OptimizationOptions{Categories: []string{"performance"}}); err == nil {
		t.Fatalf("expected generate error")
	}
}

// Test RunOptimization error path: CalculateROI fails.
func TestBasicProductionService_RunOptimization_ErrorROI(t *testing.T) {
	logger := logging.NewLogger(logging.LevelError)
	svc := &BasicProductionService{logger: logger, cache: make(map[string]interface{})}
	svc.metricsCollector = fakeMetricsCollector{}
	svc.optimizationEngine = errROIEngine{}
	svc.deploymentAssessor = fakeDeploymentAssessor{}
	if _, err := svc.RunOptimization(context.Background(), &OptimizationOptions{Categories: []string{"performance"}}); err == nil {
		t.Fatalf("expected roi error")
	}
}

// Test RunOptimization roadmap error branch still returns success with empty roadmap.
func TestBasicProductionService_RunOptimization_RoadmapErrorFallback(t *testing.T) {
	logger := logging.NewLogger(logging.LevelError)
	svc := &BasicProductionService{logger: logger, cache: make(map[string]interface{})}
	svc.metricsCollector = fakeMetricsCollector{}
	svc.optimizationEngine = errRoadmapEngine{}
	svc.deploymentAssessor = fakeDeploymentAssessor{}
	rep, err := svc.RunOptimization(context.Background(), &OptimizationOptions{Categories: []string{"performance"}, IncludeRoadmap: true})
	if err != nil {
		t.Fatalf("unexpected error despite roadmap failure: %v", err)
	}
	if rep.FutureRoadmap == nil || len(rep.FutureRoadmap) != 0 {
		t.Fatalf("expected empty roadmap due to error fallback")
	}
}

// Test GenerateExecutiveReport options toggling of appendices/charts.
func TestBasicProductionService_GenerateExecutiveReport_OptionsVariants(t *testing.T) {
	logger := logging.NewLogger(logging.LevelError)
	baseSvc := func() *BasicProductionService {
		svc := &BasicProductionService{logger: logger, cache: make(map[string]interface{})}
		svc.metricsCollector = fakeMetricsCollector{}
		svc.optimizationEngine = fakeOptimizationEngine{}
		svc.deploymentAssessor = fakeDeploymentAssessor{}
		return svc
	}

	// No appendix
	svc1 := baseSvc()
	rep1, err := svc1.GenerateExecutiveReport(context.Background(), &ReportOptions{IncludeAppendix: false})
	if err != nil {
		t.Fatalf("report err: %v", err)
	}
	if len(rep1.Appendices) != 0 {
		t.Fatalf("expected zero appendices when disabled")
	}

	// Appendix without charts (should include base 2 only)
	svc2 := baseSvc()
	rep2, err := svc2.GenerateExecutiveReport(context.Background(), &ReportOptions{IncludeAppendix: true, IncludeCharts: false})
	if err != nil {
		t.Fatalf("report err: %v", err)
	}
	if len(rep2.Appendices) != 2 {
		t.Fatalf("expected 2 base appendices without charts, got %d", len(rep2.Appendices))
	}
	for _, a := range rep2.Appendices {
		if a.Title == "Performance Charts" {
			t.Fatalf("charts appendix should be absent")
		}
	}
}

// Test executive report ROI fallback path when optimization engine returns nil ROI (ensure zeroed ROIAnalysis fields set).
func TestBasicProductionService_GenerateExecutiveReport_ROIFallback(t *testing.T) {
	logger := logging.NewLogger(logging.LevelError)
	svc := &BasicProductionService{logger: logger, cache: make(map[string]interface{})}
	svc.metricsCollector = fakeMetricsCollector{}
	svc.optimizationEngine = fallbackROIEngine{}
	svc.deploymentAssessor = fakeDeploymentAssessor{}
	rep, err := svc.GenerateExecutiveReport(context.Background(), &ReportOptions{IncludeAppendix: false})
	if err != nil {
		t.Fatalf("report err: %v", err)
	}
	if rep.ROIAnalysis.TotalInvestment != 0 || rep.ROIAnalysis.ProjectedSavings != 0 || rep.ROIAnalysis.ROIPercentage != 0 {
		t.Fatalf("expected zeroed ROI fallback")
	}
}

// Light timing assertion to ensure processing time recorded for readiness assessment.
func TestBasicProductionService_AssessDeploymentReadiness_Timing(t *testing.T) {
	logger := logging.NewLogger(logging.LevelError)
	svc := &BasicProductionService{logger: logger, cache: make(map[string]interface{})}
	svc.metricsCollector = fakeMetricsCollector{}
	svc.optimizationEngine = fakeOptimizationEngine{}
	svc.deploymentAssessor = fakeDeploymentAssessor{}
	assess, err := svc.AssessDeploymentReadiness(context.Background(), "production")
	if err != nil {
		t.Fatalf("readiness err: %v", err)
	}
	if assess.ProcessingTimeMs < 0 {
		t.Fatalf("processing time negative")
	}
	if time.Since(assess.AssessmentTimestamp) > time.Minute {
		t.Fatalf("timestamp too old")
	}
}
