package production

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/costscope/costscope/internal/core/logging"
)

// fakeMetricsCollector covers success + error branches for runStep + GetSystemStatus parallel fan-out.
type fakeMetricsCollector struct {
	failHealth bool
}

func (f fakeMetricsCollector) CollectSystemHealth(ctx context.Context) (*SystemHealthStatus, error) {
	if f.failHealth {
		return nil, fmt.Errorf("health failed")
	}
	return &SystemHealthStatus{Status: "healthy", ComponentHealth: map[string]string{"db": "ok"}, UptimeHours: 10, ErrorRate: 0.1, ResponseTimeMs: 12.3, HealthScore: 90}, nil
}
func (f fakeMetricsCollector) CollectPerformanceMetrics(context.Context) (*PerformanceMetrics, error) {
	return &PerformanceMetrics{ThroughputOpsPerSec: 100, MemoryUsageMB: 10, MemoryLimitMB: 100, CPUUsagePercent: 5, DiskUsagePercent: 10, NetworkLatencyMs: 1, OptimizationScore: 80, PerformanceGrade: "A"}, nil
}
func (f fakeMetricsCollector) CollectSecurityMetrics(context.Context) (*SecurityMetrics, error) {
	return &SecurityMetrics{SecurityScore: 85, VulnerabilitiesOpen: 0, VulnerabilitiesHigh: 0, ComplianceStatus: map[string]string{"CIS": "pass"}, EncryptionEnabled: true, AccessViolations: 0, AuditScore: 88, SecurityGrade: "A"}, nil
}
func (f fakeMetricsCollector) CollectIntegrationMetrics(context.Context) (*IntegrationMetrics, error) {
	return &IntegrationMetrics{ConnectedSystems: 1, ActiveWorkflows: 0, AlertChannels: 1, AutomationSavings: 0, IntegrationHealth: map[string]string{"core": "ok"}, DeploymentStatus: "stable", IntegrationScore: 80, OperationalMaturity: "advanced"}, nil
}
func (f fakeMetricsCollector) CollectAnalyticsMetrics(context.Context) (*AnalyticsMetrics, error) {
	return &AnalyticsMetrics{MLModelsActive: 0, PredictionAccuracy: 0, AnomaliesDetected: 0, ForecastReliability: 0, InsightsGenerated: 0, DataQualityScore: 70, AnalyticsMaturity: "basic"}, nil
}

// fakeOptimizationEngine exercises Analyze + Generate + ROI + Plan success paths.
type fakeOptimizationEngine struct{}

func (fakeOptimizationEngine) AnalyzeOptimizations(context.Context, *OptimizationOptions) (*OptimizationResults, error) {
	return &OptimizationResults{TotalImprovements: 1, PerformanceGains: 5, CostSavings: 100, SecurityEnhancements: 1, EfficiencyGains: 3, OptimizationScore: 80}, nil
}
func (fakeOptimizationEngine) GenerateRecommendations(context.Context, *ProductionSystemMetrics) ([]ProductionRecommendation, error) {
	return []ProductionRecommendation{{ID: "1", Type: "performance", Priority: PriorityHigh, Title: "T", Description: "D", Impact: ImpactHigh, Effort: EffortLow, Timeline: "Q1"}}, nil
}
func (fakeOptimizationEngine) CalculateROI(context.Context, []ProductionRecommendation) (*ROIAnalysis, error) {
	return &ROIAnalysis{TotalInvestment: 100, ProjectedSavings: 300, ROIPercentage: 200, PaybackPeriodDays: 30, NPV: 200, IRR: 50, SavingsBreakdown: map[string]float64{"perf": 100}, CostBenefitRatio: 3}, nil
}
func (fakeOptimizationEngine) PlanRoadmap(context.Context, *ProductionSystemMetrics) ([]RoadmapItem, error) {
	return []RoadmapItem{{ID: "R1", Title: "Improve", Description: "Desc", Quarter: "Q1", Priority: PriorityMedium, Category: "performance", Dependencies: nil, Resources: nil, Value: 10, Effort: 5}}, nil
}

// fakeDeploymentAssessor single success path.
type fakeDeploymentAssessor struct{}

// Implement required interface methods (only minimal behavior used in tests).
func (fakeDeploymentAssessor) AssessReadiness(ctx context.Context, reqs *DeploymentRequirements) (*DeploymentReadinessAssessment, error) {
	if reqs == nil {
		return nil, fmt.Errorf("nil requirements")
	}
	return &DeploymentReadinessAssessment{ReadinessScore: 85, ReadinessStatus: "ready", AssessmentTimestamp: time.Now()}, nil
}
func (fakeDeploymentAssessor) ValidateEnvironment(context.Context, string) (*EnvironmentValidation, error) {
	return &EnvironmentValidation{Environment: "test", ResourcesAvailable: true, ConfigurationValid: true, ConnectivityOK: true, DependenciesReady: true, ValidationScore: 90, ValidationTimestamp: time.Now()}, nil
}
func (fakeDeploymentAssessor) RunHealthChecks(context.Context, []string) (*HealthCheckResults, error) {
	return &HealthCheckResults{OverallHealthScore: 90, ComponentResults: map[string]bool{"api": true}, FailedChecks: nil, CheckTimestamp: time.Now()}, nil
}
func (fakeDeploymentAssessor) GenerateDeploymentPlan(context.Context, string, *DeploymentRequirements) (*DeploymentPlan, error) {
	return &DeploymentPlan{Strategy: "blue_green", Steps: []DeploymentStep{{Order: 1, Name: "step", Duration: time.Minute, Type: "validation"}}, EstimatedDuration: time.Minute, RiskLevel: "low", ApprovalRequired: true, PlanCreatedAt: time.Now()}, nil
}

// TestBasicProductionService_GetSystemStatus_Success validates happy path aggregation.
func TestBasicProductionService_GetSystemStatus_Success(t *testing.T) {
	logger := logging.NewLogger(logging.LevelError)
	svc := &BasicProductionService{logger: logger, cache: make(map[string]interface{})}
	svc.metricsCollector = fakeMetricsCollector{}
	svc.optimizationEngine = fakeOptimizationEngine{}
	svc.deploymentAssessor = fakeDeploymentAssessor{}

	ctx := context.Background()
	metrics, err := svc.GetSystemStatus(ctx)
	if err != nil {
		t.Fatalf("expected success: %v", err)
	}
	if metrics.ReadinessScore == 0 {
		t.Fatalf("expected non-zero readiness score")
	}
}

// TestBasicProductionService_GetSystemStatus_Error ensures aggregated error surfaces when one collector fails.
func TestBasicProductionService_GetSystemStatus_Error(t *testing.T) {
	logger := logging.NewLogger(logging.LevelError)
	svc := &BasicProductionService{logger: logger, cache: make(map[string]interface{})}
	svc.metricsCollector = fakeMetricsCollector{failHealth: true}
	svc.optimizationEngine = fakeOptimizationEngine{}
	svc.deploymentAssessor = fakeDeploymentAssessor{}
	if _, err := svc.GetSystemStatus(context.Background()); err == nil {
		t.Fatalf("expected error from failed collector")
	}
}

// Test_runStep_ErrorWrap verifies error wrapping format and logging resilience.
func Test_runStep_ErrorWrap(t *testing.T) {
	logger := logging.NewLogger(logging.LevelError)
	_, err := runStep(context.Background(), logger, StepSpec[int]{Name: "x", Run: func(context.Context) (int, error) { return 0, errors.New("boom") }, ErrWrap: "problem: %w"})
	if err == nil || err.Error() != "problem: boom" {
		t.Fatalf("unexpected wrapped error: %v", err)
	}
}

// Test_runStep_PanicInFormat ensures panic in fmt wrapping is handled (simulate malformed ErrWrap missing %w).
func Test_runStep_PanicInFormat(t *testing.T) {
	logger := logging.NewLogger(logging.LevelError)
	// ErrWrap without %w triggers malformed fmt output; ensure error returned (content may differ by Go version) and original message substring present.
	_, err := runStep(context.Background(), logger, StepSpec[int]{Name: "x", Run: func(context.Context) (int, error) { return 0, errors.New("fail") }, ErrWrap: "noverb"})
	if err == nil || !strings.Contains(err.Error(), "fail") {
		t.Fatalf("expected error containing original message, got %v", err)
	}
}

// TestBasicProductionService_RunOptimization uses fake engine to exercise full path including roadmap.
func TestBasicProductionService_RunOptimization(t *testing.T) {
	logger := logging.NewLogger(logging.LevelError)
	svc := &BasicProductionService{logger: logger, cache: make(map[string]interface{})}
	svc.metricsCollector = fakeMetricsCollector{}
	svc.optimizationEngine = fakeOptimizationEngine{}
	svc.deploymentAssessor = fakeDeploymentAssessor{}
	rep, err := svc.RunOptimization(context.Background(), &OptimizationOptions{IncludeRoadmap: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rep.OptimizationResults.TotalImprovements == 0 || len(rep.Recommendations) == 0 {
		t.Fatalf("expected optimization content")
	}
}
