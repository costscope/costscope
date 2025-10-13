package production

import (
	"context"
	"testing"

	"github.com/costscope/costscope/internal/core/logging"
)

// minimal fake deps reused from service_basic_test.go (could refactor but keep local for isolation)

func TestBasicProductionService_GenerateExecutiveReport(t *testing.T) {
	logger := logging.NewLogger(logging.LevelError)
	svc := &BasicProductionService{logger: logger, cache: make(map[string]interface{})}
	svc.metricsCollector = fakeMetricsCollector{} // already returns deterministic metrics
	svc.optimizationEngine = fakeOptimizationEngine{}
	svc.deploymentAssessor = fakeDeploymentAssessor{}
	rep, err := svc.GenerateExecutiveReport(context.Background(), nil)
	if err != nil {
		t.Fatalf("exec report err: %v", err)
	}
	if rep.Executive.OverallHealth == "" || rep.SystemOverview.TotalCapabilities == 0 {
		t.Fatalf("expected populated executive report fields")
	}
	if rep.ROIAnalysis.ROIPercentage == 0 && rep.ROIAnalysis.TotalInvestment != 0 {
		t.Fatalf("expected ROI fields coherence")
	}
}

func TestBasicProductionService_ValidateProductionConfiguration(t *testing.T) {
	logger := logging.NewLogger(logging.LevelError)
	svc := &BasicProductionService{logger: logger, cache: make(map[string]interface{})}
	svc.metricsCollector = fakeMetricsCollector{}
	svc.optimizationEngine = fakeOptimizationEngine{}
	svc.deploymentAssessor = fakeDeploymentAssessor{}
	vr, err := svc.ValidateProductionConfiguration(context.Background())
	if err != nil {
		t.Fatalf("validate config err: %v", err)
	}
	if vr.Score == 0 {
		t.Fatalf("expected non-zero score")
	}
}

func TestBasicProductionService_GetHealthChecks(t *testing.T) {
	logger := logging.NewLogger(logging.LevelError)
	svc := &BasicProductionService{logger: logger, cache: make(map[string]interface{})}
	svc.metricsCollector = fakeMetricsCollector{}
	svc.optimizationEngine = fakeOptimizationEngine{}
	svc.deploymentAssessor = fakeDeploymentAssessor{}
	checks, err := svc.GetHealthChecks(context.Background())
	if err != nil {
		t.Fatalf("health checks err: %v", err)
	}
	if len(checks) == 0 {
		t.Fatalf("expected health check results")
	}
	for k, v := range checks {
		if v.Status == "" {
			t.Fatalf("empty status for %s", k)
		}
	}
}
