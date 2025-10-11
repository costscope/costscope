//go:build !enterprise

package production

import (
	"context"
	"testing"

	"local/costscope/internal/core/enterprise"
	"local/costscope/internal/core/integration"
	"local/costscope/internal/core/logging"
)

// minimal fake integration service for constructor parity
type fakeIntegrationService struct{}

// Implement required IntegrationService methods with no-op / zero returns.
func (f fakeIntegrationService) ListIntegrations(*integration.IntegrationFilter) (*integration.IntegrationListResult, error) {
	return &integration.IntegrationListResult{}, nil
}
func (f fakeIntegrationService) ConnectToSystem(*integration.ConnectionRequest) (*integration.ConnectionResult, error) {
	return &integration.ConnectionResult{}, nil
}
func (f fakeIntegrationService) DisconnectFromSystem(string) (*integration.DisconnectionResult, error) {
	return &integration.DisconnectionResult{}, nil
}
func (f fakeIntegrationService) GetConnectionStatus(string) (*integration.ConnectionStatus, error) {
	return &integration.ConnectionStatus{}, nil
}
func (f fakeIntegrationService) TestConnection(string) (*integration.ConnectionTestResult, error) {
	return &integration.ConnectionTestResult{}, nil
}
func (f fakeIntegrationService) CreateAlert(*integration.AlertCreateRequest) (*integration.AlertCreateResult, error) {
	return &integration.AlertCreateResult{}, nil
}
func (f fakeIntegrationService) ListAlerts(*integration.AlertFilter) (*integration.AlertListResult, error) {
	return &integration.AlertListResult{}, nil
}
func (f fakeIntegrationService) UpdateAlert(string, *integration.AlertUpdateRequest) (*integration.AlertUpdateResult, error) {
	return &integration.AlertUpdateResult{}, nil
}
func (f fakeIntegrationService) DeleteAlert(string) (*integration.AlertDeleteResult, error) {
	return &integration.AlertDeleteResult{}, nil
}
func (f fakeIntegrationService) TestAlertChannels() (*integration.AlertTestResult, error) {
	return &integration.AlertTestResult{}, nil
}
func (f fakeIntegrationService) CreateWorkflow(*integration.WorkflowCreateRequest) (*integration.WorkflowCreateResult, error) {
	return &integration.WorkflowCreateResult{}, nil
}
func (f fakeIntegrationService) ListWorkflows(*integration.WorkflowFilter) (*integration.WorkflowListResult, error) {
	return &integration.WorkflowListResult{}, nil
}
func (f fakeIntegrationService) ExecuteWorkflow(string) (*integration.WorkflowExecutionResult, error) {
	return &integration.WorkflowExecutionResult{}, nil
}
func (f fakeIntegrationService) UpdateWorkflow(string, *integration.WorkflowUpdateRequest) (*integration.WorkflowUpdateResult, error) {
	return &integration.WorkflowUpdateResult{}, nil
}
func (f fakeIntegrationService) DeleteWorkflow(string) (*integration.WorkflowDeleteResult, error) {
	return &integration.WorkflowDeleteResult{}, nil
}
func (f fakeIntegrationService) StartDashboard(*integration.DashboardConfig) (*integration.DashboardStartResult, error) {
	return &integration.DashboardStartResult{}, nil
}
func (f fakeIntegrationService) StopDashboard() (*integration.DashboardStopResult, error) {
	return &integration.DashboardStopResult{}, nil
}
func (f fakeIntegrationService) GetDashboardStatus() (*integration.DashboardStatusResult, error) {
	return &integration.DashboardStatusResult{}, nil
}
func (f fakeIntegrationService) GetDashboardMetrics() (*integration.DashboardMetricsResult, error) {
	return &integration.DashboardMetricsResult{}, nil
}
func (f fakeIntegrationService) CreateWebhook(*integration.WebhookCreateRequest) (*integration.WebhookCreateResult, error) {
	return &integration.WebhookCreateResult{}, nil
}
func (f fakeIntegrationService) ListWebhooks() (*integration.WebhookListResult, error) {
	return &integration.WebhookListResult{}, nil
}
func (f fakeIntegrationService) TestWebhook(string) (*integration.WebhookTestResult, error) {
	return &integration.WebhookTestResult{}, nil
}
func (f fakeIntegrationService) DeleteWebhook(string) (*integration.WebhookDeleteResult, error) {
	return &integration.WebhookDeleteResult{}, nil
}

func TestEnhancedProductionServiceStub_Disabled(t *testing.T) {
	logger := logging.NewLogger(logging.LevelError)
	basic := NewBasicProductionService(nil, logger) // providerManager nil acceptable for stub test
	svc := NewEnhancedProductionService(basic, fakeIntegrationService{}, logger)
	if _, err := svc.GetSystemStatusWithIntegrations(context.Background()); !enterprise.IsDisabled(err) {
		t.Fatalf("expected disabled enterprise error, got %v", err)
	}
}
