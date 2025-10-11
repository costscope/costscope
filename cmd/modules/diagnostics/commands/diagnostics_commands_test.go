package commands

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"

	"local/costscope/internal/core/logging"
	"local/costscope/internal/core/production"
	"local/costscope/internal/providers"
)

// mockProductionService provides deterministic data for golden tests
type mockProductionService struct{}

func (m *mockProductionService) GetSystemStatus(_ context.Context) (*production.ProductionSystemMetrics, error) {
	return &production.ProductionSystemMetrics{
		Timestamp:       time.Unix(1700000000, 0),
		ReadinessScore:  88,
		ProductionReady: true,
		SystemHealth: production.SystemHealthStatus{
			Status:          "healthy",
			HealthScore:     92,
			ComponentHealth: map[string]string{"providers": "healthy", "analytics": "healthy", "reports": "healthy"},
		},
		Performance: production.PerformanceMetrics{PerformanceGrade: "A-", OptimizationScore: 90},
		Security:    production.SecurityMetrics{SecurityGrade: "A-", SecurityScore: 90, ComplianceStatus: map[string]string{"data_encryption": "compliant"}},
		Integration: production.IntegrationMetrics{OperationalMaturity: "intermediate", IntegrationScore: 88, IntegrationHealth: map[string]string{"aws_provider": "healthy"}},
	}, nil
}

func (m *mockProductionService) RunOptimization(_ context.Context, _ *production.OptimizationOptions) (*production.ProductionOptimizationReport, error) {
	return nil, nil
}
func (m *mockProductionService) AssessDeploymentReadiness(_ context.Context, _ string) (*production.DeploymentReadinessAssessment, error) {
	return nil, nil
}
func (m *mockProductionService) GenerateExecutiveReport(_ context.Context, _ *production.ReportOptions) (*production.ExecutiveReport, error) {
	return nil, nil
}
func (m *mockProductionService) ValidateProductionConfiguration(_ context.Context) (*production.ValidationResult, error) {
	return nil, nil
}
func (m *mockProductionService) GetHealthChecks(_ context.Context) (map[string]production.CheckResult, error) {
	return nil, nil
}

// lowMock returns a lower readiness score to test thresholds
type lowMock struct{}

func (l *lowMock) GetSystemStatus(_ context.Context) (*production.ProductionSystemMetrics, error) {
	return &production.ProductionSystemMetrics{
		Timestamp:       time.Unix(1700000000, 0),
		ReadinessScore:  50,
		ProductionReady: false,
		SystemHealth:    production.SystemHealthStatus{Status: "degraded", HealthScore: 70, ComponentHealth: map[string]string{"providers": "warning"}},
		Performance:     production.PerformanceMetrics{PerformanceGrade: "B", OptimizationScore: 70},
		Security:        production.SecurityMetrics{SecurityGrade: "B", SecurityScore: 70},
		Integration:     production.IntegrationMetrics{OperationalMaturity: "basic", IntegrationScore: 60},
	}, nil
}
func (l *lowMock) RunOptimization(_ context.Context, _ *production.OptimizationOptions) (*production.ProductionOptimizationReport, error) {
	return nil, nil
}
func (l *lowMock) AssessDeploymentReadiness(_ context.Context, _ string) (*production.DeploymentReadinessAssessment, error) {
	return nil, nil
}
func (l *lowMock) GenerateExecutiveReport(_ context.Context, _ *production.ReportOptions) (*production.ExecutiveReport, error) {
	return nil, nil
}
func (l *lowMock) ValidateProductionConfiguration(_ context.Context) (*production.ValidationResult, error) {
	return nil, nil
}
func (l *lowMock) GetHealthChecks(_ context.Context) (map[string]production.CheckResult, error) {
	return nil, nil
}

// helper to execute command and capture output
func execCmd(t *testing.T, cmd *cobra.Command, args ...string) (string, error) {
	t.Helper()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return buf.String(), err
}

func TestDiagnostics_Status_FlagsValidation(t *testing.T) {
	logger := logging.NewLogger("test")
	pm := &providers.ProviderManager{}
	// invalid output uses a fresh command instance
	dc1 := NewDiagnosticsCommandsWithService(logger, pm, &mockProductionService{})
	root1 := dc1.BuildDiagnosticsCommand()
	_, err := execCmd(t, root1, "status", "--output", "xml")
	if err == nil || !strings.Contains(err.Error(), "invalid --output") {
		t.Fatalf("expected output flag validation error, got %v", err)
	}

	// invalid min-score uses a fresh command instance to avoid flag carryover
	dc2 := NewDiagnosticsCommandsWithService(logger, pm, &mockProductionService{})
	root2 := dc2.BuildDiagnosticsCommand()
	_, err = execCmd(t, root2, "status", "--min-score", "101")
	if err == nil || !strings.Contains(err.Error(), "invalid --min-score") {
		t.Fatalf("expected min-score validation error, got %v", err)
	}
}

func TestDiagnostics_Status_E2E_JSON(t *testing.T) {
	// e2e: real production service, ensure JSON includes readiness_score
	logger := logging.NewLogger("test")
	pm := providers.NewProviderManager()
	dc := NewDiagnosticsCommands(logger, pm) // real service wiring
	root := dc.BuildDiagnosticsCommand()

	out, err := execCmd(t, root, "status", "--output", "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "\"readiness_score\"") {
		t.Fatalf("expected readiness_score in JSON, got: %s", out)
	}
}

func TestDiagnostics_Status_Golden_Table(t *testing.T) {
	logger := logging.NewLogger("test")
	pm := &providers.ProviderManager{}
	dc := NewDiagnosticsCommandsWithService(logger, pm, &mockProductionService{})
	root := dc.BuildDiagnosticsCommand()

	out, err := execCmd(t, root, "status", "--output", "table", "--detailed")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	golden := "Diagnostics Status\n===================\nReadiness Score: 88\nProduction Ready: true\nHealth: healthy (score: 92)\nPerformance: grade A- (score: 90)\nSecurity: grade A- (score: 90)\nIntegration: intermediate (score: 88)\n\nComponents:\n  - providers: healthy\n  - analytics: healthy\n  - reports: healthy\n"

	if out != golden {
		t.Fatalf("golden mismatch\nexpected:\n%s\n---\nactual:\n%s", golden, out)
	}
}

func TestDiagnostics_Status_Threshold(t *testing.T) {
	logger := logging.NewLogger("test")
	pm := &providers.ProviderManager{}
	dc := NewDiagnosticsCommandsWithService(logger, pm, &lowMock{})
	root := dc.BuildDiagnosticsCommand()

	_, err := execCmd(t, root, "status", "--min-score", "80")
	if err == nil || !strings.Contains(err.Error(), "below required minimum") {
		t.Fatalf("expected threshold error, got %v", err)
	}
}
