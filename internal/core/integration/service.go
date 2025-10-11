package integration

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Service implements the IntegrationService interface
type Service struct {
	mu                 sync.RWMutex
	connections        map[string]*SystemConnection
	alerts             map[string]*Alert
	workflows          map[string]*Workflow
	webhooks           map[string]*Webhook
	dashboardRunning   bool
	dashboardConfig    *DashboardConfig
	dashboardStartTime time.Time
	auditLogger        *AuditLogger
	metrics            *MetricsCollector
	ctx                context.Context
	cancel             context.CancelFunc
}

// SystemConnection represents an active connection to an external system
type SystemConnection struct {
	ID            string                      `json:"id"`
	Name          string                      `json:"name"`
	Type          SystemType                  `json:"type"`
	Status        IntegrationConnectionStatus `json:"status"`
	Config        *SystemConfig               `json:"config"`
	Metrics       *IntegrationMetrics         `json:"metrics"`
	EstablishedAt time.Time                   `json:"established_at"`
	LastActivity  time.Time                   `json:"last_activity"`
	HealthScore   float64                     `json:"health_score"`
}

// NewService creates a new integration service instance
func NewService() *Service {
	ctx, cancel := context.WithCancel(context.Background())

	service := &Service{
		connections:      make(map[string]*SystemConnection),
		alerts:           make(map[string]*Alert),
		workflows:        make(map[string]*Workflow),
		webhooks:         make(map[string]*Webhook),
		dashboardRunning: false,
		auditLogger:      NewAuditLogger(),
		metrics:          NewMetricsCollector(),
		ctx:              ctx,
		cancel:           cancel,
	}

	// Start background monitoring
	go service.startBackgroundMonitoring()

	return service
}

// Connection Management Methods

func (s *Service) ListIntegrations(filter *IntegrationFilter) (*IntegrationListResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	integrations := s.getAvailableIntegrations()

	// Apply filters
	if filter != nil {
		filtered := make([]Integration, 0)
		for _, integration := range integrations {
			if filter.Category != "" && integration.Category != filter.Category {
				continue
			}
			if filter.Status != "" && integration.Status != filter.Status {
				continue
			}
			filtered = append(filtered, integration)
		}
		integrations = filtered
	}

	categories := s.getCategories()

	s.auditLogger.Log("list_integrations", "", fmt.Sprintf("Filter: %+v", filter))

	return &IntegrationListResult{
		Integrations: integrations,
		Total:        len(integrations),
		Categories:   categories,
	}, nil
}

func (s *Service) ConnectToSystem(request *ConnectionRequest) (*ConnectionResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Validate request
	if request.SystemName == "" {
		return &ConnectionResult{
			Success: false,
			Error:   "system name is required",
		}, fmt.Errorf("system name is required")
	}

	// Check if already connected
	if conn, exists := s.connections[request.SystemName]; exists {
		if conn.Status == StatusConnected {
			return &ConnectionResult{
				Success: false,
				Error:   "system already connected",
			}, fmt.Errorf("system %s already connected", request.SystemName)
		}
	}

	// Create system configuration
	config := &SystemConfig{
		Name:       request.SystemName,
		TestMode:   request.TestMode,
		Timeout:    30 * time.Second,
		RetryCount: 3,
		Options:    request.Config,
	}

	// Determine system type
	systemType := s.detectSystemType(request.SystemName)
	config.Type = systemType

	// Test connection
	if !request.TestMode {
		success, err := s.testSystemConnection(config, request.Credentials)
		if !success {
			s.auditLogger.LogError("connect_system", request.SystemName, err.Error())
			return &ConnectionResult{
				Success: false,
				Error:   fmt.Sprintf("connection test failed: %v", err),
			}, err
		}
	}

	// Create connection
	connection := &SystemConnection{
		ID:            s.generateConnectionID(),
		Name:          request.SystemName,
		Type:          systemType,
		Status:        StatusConnected,
		Config:        config,
		EstablishedAt: time.Now(),
		LastActivity:  time.Now(),
		HealthScore:   100.0,
		Metrics: &IntegrationMetrics{
			SystemName:       request.SystemName,
			ConnectionHealth: 100.0,
			DataSyncRate:     0.0,
			ErrorRate:        0.0,
			LastSync:         time.Now(),
			TotalRequests:    0,
			FailedRequests:   0,
			AverageLatency:   "0ms",
			Uptime:           "0s",
		},
	}

	s.connections[request.SystemName] = connection

	// Get available features and metrics
	features := s.getSystemFeatures(systemType)
	metricsCount := s.getAvailableMetricsCount(systemType)

	s.auditLogger.Log("connect_system", request.SystemName, fmt.Sprintf("Connected to %s", systemType))
	s.metrics.RecordConnection(request.SystemName, true)

	return &ConnectionResult{
		Success:          true,
		SystemName:       request.SystemName,
		Status:           string(StatusConnected),
		ConnectionID:     connection.ID,
		DataSync:         "active",
		AvailableMetrics: metricsCount,
		Features:         features,
		EstablishedAt:    connection.EstablishedAt,
	}, nil
}

func (s *Service) DisconnectFromSystem(systemName string) (*DisconnectionResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, exists := s.connections[systemName]
	if !exists {
		return &DisconnectionResult{
			Success: false,
			Error:   "system not connected",
		}, fmt.Errorf("system %s not connected", systemName)
	}

	// Remove connection
	delete(s.connections, systemName)

	s.auditLogger.Log("disconnect_system", systemName, "System disconnected")
	s.metrics.RecordDisconnection(systemName)

	return &DisconnectionResult{
		Success:        true,
		SystemName:     systemName,
		DisconnectedAt: time.Now(),
	}, nil
}

func (s *Service) GetConnectionStatus(systemName string) (*ConnectionStatus, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	connection, exists := s.connections[systemName]
	if !exists {
		return nil, fmt.Errorf("system %s not connected", systemName)
	}

	uptime := time.Since(connection.EstablishedAt)

	return &ConnectionStatus{
		SystemName:      systemName,
		Status:          string(connection.Status),
		LastSync:        connection.Metrics.LastSync,
		DataTransferred: connection.Metrics.TotalRequests * 1024, // Approximate
		Uptime:          uptime.String(),
		HealthScore:     connection.HealthScore,
	}, nil
}

func (s *Service) TestConnection(systemName string) (*ConnectionTestResult, error) {
	s.mu.RLock()
	connection, exists := s.connections[systemName]
	s.mu.RUnlock()

	if !exists {
		return &ConnectionTestResult{
			SystemName: systemName,
			Success:    false,
			Error:      "system not connected",
		}, fmt.Errorf("system %s not connected", systemName)
	}

	start := time.Now()

	// Simulate connection test
	time.Sleep(100 * time.Millisecond)

	responseTime := time.Since(start)
	features := s.getSystemFeatures(connection.Type)

	s.auditLogger.Log("test_connection", systemName, fmt.Sprintf("Response time: %v", responseTime))

	return &ConnectionTestResult{
		SystemName:        systemName,
		Success:           true,
		ResponseTime:      responseTime.String(),
		AvailableFeatures: features,
	}, nil
}

// Alert Management Methods

func (s *Service) CreateAlert(request *AlertCreateRequest) (*AlertCreateResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Validate request
	if request.Name == "" {
		return &AlertCreateResult{
			Success: false,
			Error:   "alert name is required",
		}, fmt.Errorf("alert name is required")
	}

	alertID := s.generateAlertID()
	alert := &Alert{
		ID:           alertID,
		Name:         request.Name,
		Type:         request.Type,
		Severity:     request.Severity,
		Threshold:    request.Threshold,
		Channels:     request.Channels,
		Status:       "active",
		Enabled:      true,
		Created:      time.Now(),
		TriggerCount: 0,
	}

	s.alerts[alertID] = alert

	s.auditLogger.Log("create_alert", "", fmt.Sprintf("Alert created: %s", request.Name))

	return &AlertCreateResult{
		AlertID:   alertID,
		Success:   true,
		CreatedAt: alert.Created,
	}, nil
}

func (s *Service) ListAlerts(filter *AlertFilter) (*AlertListResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	alerts := make([]Alert, 0)
	activeCount := 0

	for _, alert := range s.alerts {
		// Apply filters
		if filter != nil {
			if filter.Type != "" && alert.Type != filter.Type {
				continue
			}
			if filter.Severity != "" && alert.Severity != filter.Severity {
				continue
			}
			if filter.Status != "" && alert.Status != filter.Status {
				continue
			}
		}

		alerts = append(alerts, *alert)
		if alert.Status == "active" {
			activeCount++
		}
	}

	return &AlertListResult{
		Alerts: alerts,
		Total:  len(alerts),
		Active: activeCount,
	}, nil
}

func (s *Service) UpdateAlert(alertID string, request *AlertUpdateRequest) (*AlertUpdateResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	alert, exists := s.alerts[alertID]
	if !exists {
		return &AlertUpdateResult{
			Success: false,
			Error:   "alert not found",
		}, fmt.Errorf("alert %s not found", alertID)
	}

	// Update alert fields
	if request.Name != "" {
		alert.Name = request.Name
	}
	if request.Threshold > 0 {
		alert.Threshold = request.Threshold
	}
	if len(request.Channels) > 0 {
		alert.Channels = request.Channels
	}
	alert.Enabled = request.Enabled

	s.auditLogger.Log("update_alert", alertID, fmt.Sprintf("Alert updated: %s", alert.Name))

	return &AlertUpdateResult{
		AlertID:   alertID,
		Success:   true,
		UpdatedAt: time.Now(),
	}, nil
}

func (s *Service) DeleteAlert(alertID string) (*AlertDeleteResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, exists := s.alerts[alertID]
	if !exists {
		return &AlertDeleteResult{
			Success: false,
			Error:   "alert not found",
		}, fmt.Errorf("alert %s not found", alertID)
	}

	delete(s.alerts, alertID)

	s.auditLogger.Log("delete_alert", alertID, "Alert deleted")

	return &AlertDeleteResult{
		AlertID:   alertID,
		Success:   true,
		DeletedAt: time.Now(),
	}, nil
}

func (s *Service) TestAlertChannels() (*AlertTestResult, error) {
	// Simulate testing all alert channels
	time.Sleep(500 * time.Millisecond)

	s.auditLogger.Log("test_alert_channels", "", "Alert channels tested")

	return &AlertTestResult{
		Email:   " Connected",
		Slack:   " Connected",
		SMS:     " Connected",
		Webhook: " Connected",
		Teams:   " Connected",
		Discord: " Connected",
	}, nil
}

// Helper methods

func (s *Service) generateConnectionID() string {
	return fmt.Sprintf("conn_%d", time.Now().UnixNano())
}

func (s *Service) generateAlertID() string {
	return fmt.Sprintf("alert_%d", time.Now().UnixNano())
}

func (s *Service) detectSystemType(systemName string) SystemType {
	// Simple detection based on system name
	switch systemName {
	case "aws", "amazon":
		return SystemAWS
	case "azure", "microsoft":
		return SystemAzure
	case "gcp", "google":
		return SystemGCP
	case "slack":
		return SystemSlack
	case "teams":
		return SystemTeams
	case "jira":
		return SystemJIRA
	case "servicenow":
		return SystemServiceNow
	case "tableau":
		return SystemTableau
	case "datadog":
		return SystemDatadog
	default:
		return SystemType("custom")
	}
}

func (s *Service) testSystemConnection(config *SystemConfig, _ map[string]string) (bool, error) {
	// Simulate connection test
	time.Sleep(200 * time.Millisecond)

	// In a real implementation, this would test actual connectivity
	if config.TestMode {
		return true, nil
	}

	// Simulate different connection scenarios
	if config.Name == "fail-test" {
		return false, fmt.Errorf("connection failed")
	}

	return true, nil
}

func (s *Service) getSystemFeatures(systemType SystemType) []string {
	switch systemType {
	case SystemAWS:
		return []string{"cost_analysis", "budget_alerts", "resource_optimization", "rightsizing"}
	case SystemSlack:
		return []string{"notifications", "interactive_alerts", "cost_reports"}
	case SystemTableau:
		return []string{"data_visualization", "cost_dashboards", "custom_reports"}
	case SystemJIRA:
		return []string{"ticket_creation", "cost_tracking", "project_budgets"}
	default:
		return []string{"basic_integration", "data_sync"}
	}
}

func (s *Service) getAvailableMetricsCount(systemType SystemType) int {
	switch systemType {
	case SystemAWS:
		return 25
	case SystemAzure:
		return 20
	case SystemGCP:
		return 18
	default:
		return 5
	}
}

func (s *Service) getAvailableIntegrations() []Integration {
	return []Integration{
		{
			Name:           "aws",
			DisplayName:    "Amazon Web Services",
			Description:    "Connect to AWS for cost analysis and optimization",
			Category:       string(CategoryBilling),
			Status:         "available",
			Version:        "1.0.0",
			Features:       []string{"cost_analysis", "budget_alerts", "rightsizing"},
			RequiredConfig: []string{"access_key", "secret_key", "region"},
		},
		{
			Name:           "slack",
			DisplayName:    "Slack",
			Description:    "Send cost alerts and reports to Slack channels",
			Category:       string(CategoryNotification),
			Status:         "available",
			Version:        "1.0.0",
			Features:       []string{"notifications", "interactive_alerts"},
			RequiredConfig: []string{"webhook_url", "channel"},
		},
		{
			Name:           "tableau",
			DisplayName:    "Tableau",
			Description:    "Create cost visualization dashboards in Tableau",
			Category:       string(CategoryBI),
			Status:         "available",
			Version:        "1.0.0",
			Features:       []string{"data_visualization", "dashboards"},
			RequiredConfig: []string{"server_url", "username", "password"},
		},
	}
}

func (s *Service) getCategories() []string {
	return []string{
		string(CategoryBilling),
		string(CategoryITSM),
		string(CategoryBI),
		string(CategoryMonitoring),
		string(CategoryAutomation),
		string(CategoryNotification),
		string(CategorySecurity),
		string(CategoryDevOps),
	}
}

func (s *Service) startBackgroundMonitoring() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.updateConnectionMetrics()
		}
	}
}

func (s *Service) updateConnectionMetrics() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, connection := range s.connections {
		connection.LastActivity = time.Now()
		connection.Metrics.LastSync = time.Now()
		connection.Metrics.TotalRequests++
	}
}

// Close gracefully shuts down the service
func (s *Service) Close() error {
	s.cancel()
	return nil
}
