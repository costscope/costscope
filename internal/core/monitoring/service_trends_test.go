package monitoring

import (
	"context"
	"testing"
	"time"

	"github.com/costscope/costscope/internal/core/integration"
	"github.com/costscope/costscope/internal/core/logging"
	"github.com/costscope/costscope/internal/core/production"
)

// stubProduction implements production.ProductionService minimally
type stubProduction struct{}

func (s *stubProduction) GetSystemStatus(ctx context.Context) (*production.ProductionSystemMetrics, error) {
	return nil, nil
}
func (s *stubProduction) RunOptimization(ctx context.Context, options *production.OptimizationOptions) (*production.ProductionOptimizationReport, error) {
	return nil, nil
}
func (s *stubProduction) AssessDeploymentReadiness(ctx context.Context, environment string) (*production.DeploymentReadinessAssessment, error) {
	return nil, nil
}
func (s *stubProduction) GenerateExecutiveReport(ctx context.Context, options *production.ReportOptions) (*production.ExecutiveReport, error) {
	return nil, nil
}
func (s *stubProduction) ValidateProductionConfiguration(ctx context.Context) (*production.ValidationResult, error) {
	return nil, nil
}
func (s *stubProduction) GetHealthChecks(ctx context.Context) (map[string]production.CheckResult, error) {
	return map[string]production.CheckResult{}, nil
}

type stubIntegration struct{}

func (s *stubIntegration) ListIntegrations(filter *integration.IntegrationFilter) (*integration.IntegrationListResult, error) {
	return nil, nil
}
func (s *stubIntegration) ConnectToSystem(request *integration.ConnectionRequest) (*integration.ConnectionResult, error) {
	return nil, nil
}
func (s *stubIntegration) DisconnectFromSystem(systemName string) (*integration.DisconnectionResult, error) {
	return nil, nil
}
func (s *stubIntegration) GetConnectionStatus(systemName string) (*integration.ConnectionStatus, error) {
	return nil, nil
}
func (s *stubIntegration) TestConnection(systemName string) (*integration.ConnectionTestResult, error) {
	return nil, nil
}
func (s *stubIntegration) CreateAlert(request *integration.AlertCreateRequest) (*integration.AlertCreateResult, error) {
	return nil, nil
}
func (s *stubIntegration) ListAlerts(filter *integration.AlertFilter) (*integration.AlertListResult, error) {
	return nil, nil
}
func (s *stubIntegration) UpdateAlert(alertID string, request *integration.AlertUpdateRequest) (*integration.AlertUpdateResult, error) {
	return nil, nil
}
func (s *stubIntegration) DeleteAlert(alertID string) (*integration.AlertDeleteResult, error) {
	return nil, nil
}
func (s *stubIntegration) TestAlertChannels() (*integration.AlertTestResult, error) { return nil, nil }
func (s *stubIntegration) CreateWorkflow(request *integration.WorkflowCreateRequest) (*integration.WorkflowCreateResult, error) {
	return nil, nil
}
func (s *stubIntegration) ListWorkflows(filter *integration.WorkflowFilter) (*integration.WorkflowListResult, error) {
	return nil, nil
}
func (s *stubIntegration) ExecuteWorkflow(workflowID string) (*integration.WorkflowExecutionResult, error) {
	return nil, nil
}
func (s *stubIntegration) UpdateWorkflow(workflowID string, request *integration.WorkflowUpdateRequest) (*integration.WorkflowUpdateResult, error) {
	return nil, nil
}
func (s *stubIntegration) DeleteWorkflow(workflowID string) (*integration.WorkflowDeleteResult, error) {
	return nil, nil
}
func (s *stubIntegration) StartDashboard(config *integration.DashboardConfig) (*integration.DashboardStartResult, error) {
	return nil, nil
}
func (s *stubIntegration) StopDashboard() (*integration.DashboardStopResult, error) { return nil, nil }
func (s *stubIntegration) GetDashboardStatus() (*integration.DashboardStatusResult, error) {
	return nil, nil
}
func (s *stubIntegration) GetDashboardMetrics() (*integration.DashboardMetricsResult, error) {
	return nil, nil
}
func (s *stubIntegration) CreateWebhook(request *integration.WebhookCreateRequest) (*integration.WebhookCreateResult, error) {
	return nil, nil
}
func (s *stubIntegration) ListWebhooks() (*integration.WebhookListResult, error) { return nil, nil }
func (s *stubIntegration) TestWebhook(webhookID string) (*integration.WebhookTestResult, error) {
	return nil, nil
}
func (s *stubIntegration) DeleteWebhook(webhookID string) (*integration.WebhookDeleteResult, error) {
	return nil, nil
}

func TestBasicMonitoringService_PerformanceTrends(t *testing.T) {
	logger := logging.NewLogger(logging.LevelInfo)
	svc := NewBasicMonitoringService(logger, &stubProduction{}, &stubIntegration{})
	trends, err := svc.GetPerformanceTrends(context.Background(), 1*time.Hour)
	if err != nil {
		t.Fatalf("GetPerformanceTrends error: %v", err)
	}
	if trends == nil || len(trends.CPUTrend.DataPoints) == 0 {
		t.Fatalf("expected non-empty CPU trend data")
	}
	if trends.TimeRange != 1*time.Hour {
		t.Fatalf("unexpected time range: %v", trends.TimeRange)
	}
}
