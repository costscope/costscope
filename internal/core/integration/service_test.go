package integration

import (
	"testing"
	"time"
)

func TestNewService(t *testing.T) {
	service := NewService()
	defer func() { _ = service.Close() }()

	if service == nil {
		t.Fatal("NewService() returned nil")
	}

	if service.connections == nil {
		t.Error("Service connections map not initialized")
	}

	if service.alerts == nil {
		t.Error("Service alerts map not initialized")
	}

	if service.workflows == nil {
		t.Error("Service workflows map not initialized")
	}

	if service.webhooks == nil {
		t.Error("Service webhooks map not initialized")
	}
}

func TestListIntegrations(t *testing.T) {
	service := NewService()
	defer func() { _ = service.Close() }()

	// Test listing all integrations
	result, err := service.ListIntegrations(nil)
	if err != nil {
		t.Fatalf("ListIntegrations() failed: %v", err)
	}

	if result.Total == 0 {
		t.Error("Expected some integrations to be available")
	}

	if len(result.Categories) == 0 {
		t.Error("Expected some categories to be available")
	}

	// Test filtering by category
	filter := &IntegrationFilter{Category: "billing"}
	result, err = service.ListIntegrations(filter)
	if err != nil {
		t.Fatalf("ListIntegrations() with filter failed: %v", err)
	}

	for _, integration := range result.Integrations {
		if integration.Category != "billing" {
			t.Errorf("Expected category 'billing', got '%s'", integration.Category)
		}
	}
}

func TestConnectToSystem(t *testing.T) {
	service := NewService()
	defer func() { _ = service.Close() }()

	// Test successful connection
	request := &ConnectionRequest{
		SystemName: "test-system",
		Config: map[string]interface{}{
			"test_key": "test_value",
		},
		Credentials: map[string]string{
			"api_key": "test_api_key",
		},
		TestMode: true,
	}

	result, err := service.ConnectToSystem(request)
	if err != nil {
		t.Fatalf("ConnectToSystem() failed: %v", err)
	}

	if !result.Success {
		t.Errorf("Expected success, got failure: %s", result.Error)
	}

	if result.SystemName != "test-system" {
		t.Errorf("Expected system name 'test-system', got '%s'", result.SystemName)
	}

	// Test duplicate connection
	_, err = service.ConnectToSystem(request)
	if err == nil {
		t.Error("Expected error for duplicate connection, got nil")
	}

	// Test invalid request
	invalidRequest := &ConnectionRequest{
		SystemName: "",
	}

	result, err = service.ConnectToSystem(invalidRequest)
	if err == nil {
		t.Error("Expected error for empty system name, got nil")
	}

	if result.Success {
		t.Error("Expected failure for invalid request")
	}
}

func TestDisconnectFromSystem(t *testing.T) {
	service := NewService()
	defer func() { _ = service.Close() }()

	// Test disconnecting non-existent system
	result, err := service.DisconnectFromSystem("non-existent")
	if err == nil {
		t.Error("Expected error for non-existent system, got nil")
	}

	if result.Success {
		t.Error("Expected failure for non-existent system")
	}

	// Connect a system first
	connectRequest := &ConnectionRequest{
		SystemName: "test-system",
		TestMode:   true,
	}

	_, err = service.ConnectToSystem(connectRequest)
	if err != nil {
		t.Fatalf("ConnectToSystem() failed: %v", err)
	}

	// Test successful disconnection
	result, err = service.DisconnectFromSystem("test-system")
	if err != nil {
		t.Fatalf("DisconnectFromSystem() failed: %v", err)
	}

	if !result.Success {
		t.Errorf("Expected success, got failure: %s", result.Error)
	}
}

func TestGetConnectionStatus(t *testing.T) {
	service := NewService()
	defer func() { _ = service.Close() }()

	// Test getting status for non-existent connection
	_, err := service.GetConnectionStatus("non-existent")
	if err == nil {
		t.Error("Expected error for non-existent connection, got nil")
	}

	// Connect a system first
	connectRequest := &ConnectionRequest{
		SystemName: "test-system",
		TestMode:   true,
	}

	_, err = service.ConnectToSystem(connectRequest)
	if err != nil {
		t.Fatalf("ConnectToSystem() failed: %v", err)
	}

	// Test getting status for existing connection
	status, err := service.GetConnectionStatus("test-system")
	if err != nil {
		t.Fatalf("GetConnectionStatus() failed: %v", err)
	}

	if status.SystemName != "test-system" {
		t.Errorf("Expected system name 'test-system', got '%s'", status.SystemName)
	}

	if status.HealthScore < 0 || status.HealthScore > 100 {
		t.Errorf("Invalid health score: %f", status.HealthScore)
	}
}

func TestCreateAlert(t *testing.T) {
	service := NewService()
	defer func() { _ = service.Close() }()

	// Test successful alert creation
	request := &AlertCreateRequest{
		Name:      "Test Alert",
		Type:      "budget",
		Severity:  "high",
		Threshold: 1000.0,
		Channels:  []string{"email", "slack"},
		Schedule:  "daily",
	}

	result, err := service.CreateAlert(request)
	if err != nil {
		t.Fatalf("CreateAlert() failed: %v", err)
	}

	if !result.Success {
		t.Errorf("Expected success, got failure: %s", result.Error)
	}

	if result.AlertID == "" {
		t.Error("Expected alert ID, got empty string")
	}

	// Test invalid request
	invalidRequest := &AlertCreateRequest{
		Name: "",
	}

	result, err = service.CreateAlert(invalidRequest)
	if err == nil {
		t.Error("Expected error for empty alert name, got nil")
	}

	if result.Success {
		t.Error("Expected failure for invalid request")
	}
}

func TestListAlerts(t *testing.T) {
	service := NewService()
	defer func() { _ = service.Close() }()

	// Create a test alert first
	createRequest := &AlertCreateRequest{
		Name:     "Test Alert",
		Type:     "budget",
		Severity: "high",
	}

	_, err := service.CreateAlert(createRequest)
	if err != nil {
		t.Fatalf("CreateAlert() failed: %v", err)
	}

	// Test listing all alerts
	result, err := service.ListAlerts(nil)
	if err != nil {
		t.Fatalf("ListAlerts() failed: %v", err)
	}

	if result.Total == 0 {
		t.Error("Expected at least one alert")
	}

	// Test filtering by type
	filter := &AlertFilter{Type: "budget"}
	result, err = service.ListAlerts(filter)
	if err != nil {
		t.Fatalf("ListAlerts() with filter failed: %v", err)
	}

	for _, alert := range result.Alerts {
		if alert.Type != "budget" {
			t.Errorf("Expected type 'budget', got '%s'", alert.Type)
		}
	}
}

func TestCreateWorkflow(t *testing.T) {
	service := NewService()
	defer func() { _ = service.Close() }()

	// Test successful workflow creation
	steps := []WorkflowStep{
		{
			ID:   "step1",
			Name: "Analysis Step",
			Type: "analysis",
		},
	}

	request := &WorkflowCreateRequest{
		Name:        "Test Workflow",
		Description: "Test workflow description",
		Schedule:    "daily",
		Steps:       steps,
	}

	result, err := service.CreateWorkflow(request)
	if err != nil {
		t.Fatalf("CreateWorkflow() failed: %v", err)
	}

	if !result.Success {
		t.Errorf("Expected success, got failure: %s", result.Error)
	}

	if result.WorkflowID == "" {
		t.Error("Expected workflow ID, got empty string")
	}

	// Test invalid request
	invalidRequest := &WorkflowCreateRequest{
		Name: "",
	}

	result, err = service.CreateWorkflow(invalidRequest)
	if err == nil {
		t.Error("Expected error for empty workflow name, got nil")
	}

	if result.Success {
		t.Error("Expected failure for invalid request")
	}
}

func TestStartStopDashboard(t *testing.T) {
	service := NewService()
	defer func() { _ = service.Close() }()

	// Test starting dashboard
	config := &DashboardConfig{
		Port:        8080,
		Theme:       "light",
		RefreshRate: 30,
	}

	result, err := service.StartDashboard(config)
	if err != nil {
		t.Fatalf("StartDashboard() failed: %v", err)
	}

	if !result.Success {
		t.Errorf("Expected success, got failure: %s", result.Error)
	}

	if result.Port != 8080 {
		t.Errorf("Expected port 8080, got %d", result.Port)
	}

	// Test getting dashboard status
	status, err := service.GetDashboardStatus()
	if err != nil {
		t.Fatalf("GetDashboardStatus() failed: %v", err)
	}

	if !status.Running {
		t.Error("Expected dashboard to be running")
	}

	// Test starting dashboard when already running
	_, err = service.StartDashboard(config)
	if err == nil {
		t.Error("Expected error when starting dashboard that's already running")
	}

	// Test stopping dashboard
	stopResult, err := service.StopDashboard()
	if err != nil {
		t.Fatalf("StopDashboard() failed: %v", err)
	}

	if !stopResult.Success {
		t.Errorf("Expected success, got failure: %s", stopResult.Error)
	}

	// Test getting status after stopping
	status, err = service.GetDashboardStatus()
	if err != nil {
		t.Fatalf("GetDashboardStatus() failed: %v", err)
	}

	if status.Running {
		t.Error("Expected dashboard to be stopped")
	}
}

func TestCreateWebhook(t *testing.T) {
	service := NewService()
	defer func() { _ = service.Close() }()

	// Test successful webhook creation
	request := &WebhookCreateRequest{
		Name:    "Test Webhook",
		URL:     "https://example.com/webhook",
		Events:  []string{"cost_change", "budget_exceeded"},
		Enabled: true,
	}

	result, err := service.CreateWebhook(request)
	if err != nil {
		t.Fatalf("CreateWebhook() failed: %v", err)
	}

	if !result.Success {
		t.Errorf("Expected success, got failure: %s", result.Error)
	}

	if result.WebhookID == "" {
		t.Error("Expected webhook ID, got empty string")
	}

	// Test invalid request - empty name
	invalidRequest := &WebhookCreateRequest{
		Name: "",
		URL:  "https://example.com/webhook",
	}

	_, err = service.CreateWebhook(invalidRequest)
	if err == nil {
		t.Error("Expected error for empty webhook name, got nil")
	}

	// Test invalid request - empty URL
	invalidRequest2 := &WebhookCreateRequest{
		Name: "Test",
		URL:  "",
	}

	_, err = service.CreateWebhook(invalidRequest2)
	if err == nil {
		t.Error("Expected error for empty webhook URL, got nil")
	}
}

func TestAuditLogger(t *testing.T) {
	logger := NewAuditLogger()

	// Test logging
	logger.Log("test_action", "test_system", "test details")

	logs := logger.GetLogs(10)
	if len(logs) == 0 {
		t.Error("Expected at least one audit log")
	}

	log := logs[0]
	if log.Action != "test_action" {
		t.Errorf("Expected action 'test_action', got '%s'", log.Action)
	}

	if log.System != "test_system" {
		t.Errorf("Expected system 'test_system', got '%s'", log.System)
	}

	if !log.Success {
		t.Error("Expected success to be true")
	}

	// Test error logging
	logger.LogError("test_error", "test_system", "error message")

	logs = logger.GetLogs(10)
	if len(logs) < 2 {
		t.Error("Expected at least two audit logs")
	}

	errorLog := logs[1]
	if errorLog.Success {
		t.Error("Expected success to be false for error log")
	}

	if errorLog.Error != "error message" {
		t.Errorf("Expected error 'error message', got '%s'", errorLog.Error)
	}
}

func TestMetricsCollector(t *testing.T) {
	collector := NewMetricsCollector()

	// Test recording connection
	collector.RecordConnection("test-system", true)

	metrics := collector.GetMetrics("test-system")
	if metrics == nil {
		t.Error("Expected metrics for test-system, got nil")
		return
	}

	if metrics.ConcurrentUsers != 1 {
		t.Errorf("Expected 1 concurrent user, got %d", metrics.ConcurrentUsers)
	}

	// Test recording failed connection
	collector.RecordConnection("test-system", false)

	metrics = collector.GetMetrics("test-system")
	if metrics.ErrorRate == 0 {
		t.Error("Expected error rate to increase after failed connection")
	}

	// Test recording disconnection
	collector.RecordDisconnection("test-system")

	metrics = collector.GetMetrics("test-system")
	if metrics.ConcurrentUsers != 0 {
		t.Errorf("Expected 0 concurrent users after disconnection, got %d", metrics.ConcurrentUsers)
	}

	// Test getting all metrics
	allMetrics := collector.GetAllMetrics()
	if len(allMetrics) == 0 {
		t.Error("Expected at least one system in all metrics")
	}

	if _, exists := allMetrics["test-system"]; !exists {
		t.Error("Expected test-system in all metrics")
	}
}

func TestBackgroundMonitoring(t *testing.T) {
	service := NewService()

	// Connect a test system
	connectRequest := &ConnectionRequest{
		SystemName: "test-system",
		TestMode:   true,
	}

	_, err := service.ConnectToSystem(connectRequest)
	if err != nil {
		t.Fatalf("ConnectToSystem() failed: %v", err)
	}

	// Get initial metrics
	status1, err := service.GetConnectionStatus("test-system")
	if err != nil {
		t.Fatalf("GetConnectionStatus() failed: %v", err)
	}

	// Wait a short time to allow background monitoring to run
	time.Sleep(100 * time.Millisecond)

	// Get updated metrics
	status2, err := service.GetConnectionStatus("test-system")
	if err != nil {
		t.Fatalf("GetConnectionStatus() failed: %v", err)
	}

	// The last sync time should be updated by background monitoring
	if !status2.LastSync.After(status1.LastSync) && !status2.LastSync.Equal(status1.LastSync) {
		t.Error("Expected background monitoring to update connection metrics")
	}

	_ = service.Close()
}
