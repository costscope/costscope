package integration

import (
	"time"
)

// IntegrationService defines the interface for integration operations
type IntegrationService interface {
	// Connection Management
	ListIntegrations(filter *IntegrationFilter) (*IntegrationListResult, error)
	ConnectToSystem(request *ConnectionRequest) (*ConnectionResult, error)
	DisconnectFromSystem(systemName string) (*DisconnectionResult, error)
	GetConnectionStatus(systemName string) (*ConnectionStatus, error)
	TestConnection(systemName string) (*ConnectionTestResult, error)

	// Alert Management
	CreateAlert(request *AlertCreateRequest) (*AlertCreateResult, error)
	ListAlerts(filter *AlertFilter) (*AlertListResult, error)
	UpdateAlert(alertID string, request *AlertUpdateRequest) (*AlertUpdateResult, error)
	DeleteAlert(alertID string) (*AlertDeleteResult, error)
	TestAlertChannels() (*AlertTestResult, error)

	// Workflow Management
	CreateWorkflow(request *WorkflowCreateRequest) (*WorkflowCreateResult, error)
	ListWorkflows(filter *WorkflowFilter) (*WorkflowListResult, error)
	ExecuteWorkflow(workflowID string) (*WorkflowExecutionResult, error)
	UpdateWorkflow(workflowID string, request *WorkflowUpdateRequest) (*WorkflowUpdateResult, error)
	DeleteWorkflow(workflowID string) (*WorkflowDeleteResult, error)

	// Dashboard Management
	StartDashboard(config *DashboardConfig) (*DashboardStartResult, error)
	StopDashboard() (*DashboardStopResult, error)
	GetDashboardStatus() (*DashboardStatusResult, error)
	GetDashboardMetrics() (*DashboardMetricsResult, error)

	// Webhook Management
	CreateWebhook(request *WebhookCreateRequest) (*WebhookCreateResult, error)
	ListWebhooks() (*WebhookListResult, error)
	TestWebhook(webhookID string) (*WebhookTestResult, error)
	DeleteWebhook(webhookID string) (*WebhookDeleteResult, error)
}

// Request types for Connection Management
type IntegrationFilter struct {
	Category string `json:"category"`
	Status   string `json:"status"`
}

type ConnectionRequest struct {
	SystemName  string                 `json:"system_name"`
	Config      map[string]interface{} `json:"config"`
	Credentials map[string]string      `json:"credentials"`
	TestMode    bool                   `json:"test_mode"`
}

// Request types for Alert Management
type AlertCreateRequest struct {
	Name       string   `json:"name"`
	Type       string   `json:"type"`
	Severity   string   `json:"severity"`
	Threshold  float64  `json:"threshold"`
	Channels   []string `json:"channels"`
	Schedule   string   `json:"schedule"`
	Conditions string   `json:"conditions"`
}

type AlertFilter struct {
	Type     string `json:"type"`
	Severity string `json:"severity"`
	Status   string `json:"status"`
}

type AlertUpdateRequest struct {
	Name      string   `json:"name"`
	Threshold float64  `json:"threshold"`
	Channels  []string `json:"channels"`
	Enabled   bool     `json:"enabled"`
}

// Request types for Workflow Management
type WorkflowCreateRequest struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Schedule    string                 `json:"schedule"`
	Steps       []WorkflowStep         `json:"steps"`
	Config      map[string]interface{} `json:"config"`
}

type WorkflowFilter struct {
	Status   string `json:"status"`
	Schedule string `json:"schedule"`
}

type WorkflowUpdateRequest struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Schedule    string         `json:"schedule"`
	Steps       []WorkflowStep `json:"steps"`
	Enabled     bool           `json:"enabled"`
}

// Request types for Dashboard Management
type DashboardConfig struct {
	Port        int    `json:"port"`
	Theme       string `json:"theme"`
	AutoOpen    bool   `json:"auto_open"`
	RefreshRate int    `json:"refresh_rate"`
	AuthEnabled bool   `json:"auth_enabled"`
}

// Request types for Webhook Management
type WebhookCreateRequest struct {
	Name    string            `json:"name"`
	URL     string            `json:"url"`
	Events  []string          `json:"events"`
	Headers map[string]string `json:"headers"`
	Secret  string            `json:"secret"`
	Enabled bool              `json:"enabled"`
}

// Response types for Connection Management
type IntegrationListResult struct {
	Integrations []Integration `json:"integrations"`
	Total        int           `json:"total"`
	Categories   []string      `json:"categories"`
}

type ConnectionResult struct {
	Success          bool      `json:"success"`
	SystemName       string    `json:"system_name"`
	Status           string    `json:"status"`
	ConnectionID     string    `json:"connection_id"`
	DataSync         string    `json:"data_sync"`
	AvailableMetrics int       `json:"available_metrics"`
	Features         []string  `json:"features"`
	Error            string    `json:"error,omitempty"`
	EstablishedAt    time.Time `json:"established_at"`
}

type DisconnectionResult struct {
	Success        bool      `json:"success"`
	SystemName     string    `json:"system_name"`
	DisconnectedAt time.Time `json:"disconnected_at"`
	Error          string    `json:"error,omitempty"`
}

type ConnectionStatus struct {
	SystemName      string    `json:"system_name"`
	Status          string    `json:"status"`
	LastSync        time.Time `json:"last_sync"`
	DataTransferred int64     `json:"data_transferred"`
	Uptime          string    `json:"uptime"`
	HealthScore     float64   `json:"health_score"`
}

type ConnectionTestResult struct {
	SystemName        string   `json:"system_name"`
	Success           bool     `json:"success"`
	ResponseTime      string   `json:"response_time"`
	AvailableFeatures []string `json:"available_features"`
	Error             string   `json:"error,omitempty"`
}

// Response types for Alert Management
type AlertCreateResult struct {
	AlertID   string    `json:"alert_id"`
	Success   bool      `json:"success"`
	CreatedAt time.Time `json:"created_at"`
	Error     string    `json:"error,omitempty"`
}

type AlertListResult struct {
	Alerts []Alert `json:"alerts"`
	Total  int     `json:"total"`
	Active int     `json:"active"`
}

type AlertUpdateResult struct {
	AlertID   string    `json:"alert_id"`
	Success   bool      `json:"success"`
	UpdatedAt time.Time `json:"updated_at"`
	Error     string    `json:"error,omitempty"`
}

type AlertDeleteResult struct {
	AlertID   string    `json:"alert_id"`
	Success   bool      `json:"success"`
	DeletedAt time.Time `json:"deleted_at"`
	Error     string    `json:"error,omitempty"`
}

type AlertTestResult struct {
	Email   string `json:"email"`
	Slack   string `json:"slack"`
	SMS     string `json:"sms"`
	Webhook string `json:"webhook"`
	Teams   string `json:"teams"`
	Discord string `json:"discord"`
}

// Response types for Workflow Management
type WorkflowCreateResult struct {
	WorkflowID string    `json:"workflow_id"`
	Success    bool      `json:"success"`
	CreatedAt  time.Time `json:"created_at"`
	Error      string    `json:"error,omitempty"`
}

type WorkflowListResult struct {
	Workflows []Workflow `json:"workflows"`
	Total     int        `json:"total"`
	Active    int        `json:"active"`
}

type WorkflowExecutionResult struct {
	WorkflowID     string    `json:"workflow_id"`
	ExecutionID    string    `json:"execution_id"`
	Success        bool      `json:"success"`
	Duration       string    `json:"duration"`
	TasksExecuted  int       `json:"tasks_executed"`
	TasksSucceeded int       `json:"tasks_succeeded"`
	TasksFailed    int       `json:"tasks_failed"`
	CostSavings    float64   `json:"cost_savings"`
	StartedAt      time.Time `json:"started_at"`
	CompletedAt    time.Time `json:"completed_at"`
	Error          string    `json:"error,omitempty"`
}

type WorkflowUpdateResult struct {
	WorkflowID string    `json:"workflow_id"`
	Success    bool      `json:"success"`
	UpdatedAt  time.Time `json:"updated_at"`
	Error      string    `json:"error,omitempty"`
}

type WorkflowDeleteResult struct {
	WorkflowID string    `json:"workflow_id"`
	Success    bool      `json:"success"`
	DeletedAt  time.Time `json:"deleted_at"`
	Error      string    `json:"error,omitempty"`
}

// Response types for Dashboard Management
type DashboardStartResult struct {
	Success   bool      `json:"success"`
	URL       string    `json:"url"`
	Port      int       `json:"port"`
	StartedAt time.Time `json:"started_at"`
	Error     string    `json:"error,omitempty"`
}

type DashboardStopResult struct {
	Success   bool      `json:"success"`
	StoppedAt time.Time `json:"stopped_at"`
	Error     string    `json:"error,omitempty"`
}

type DashboardStatusResult struct {
	Running      bool      `json:"running"`
	URL          string    `json:"url"`
	Port         int       `json:"port"`
	StartedAt    time.Time `json:"started_at"`
	Uptime       string    `json:"uptime"`
	ActiveUsers  int       `json:"active_users"`
	RequestCount int64     `json:"request_count"`
}

type DashboardMetricsResult struct {
	TotalCost        float64   `json:"total_cost"`
	MonthlyCost      float64   `json:"monthly_cost"`
	CostTrend        string    `json:"cost_trend"`
	TopServices      []string  `json:"top_services"`
	LastUpdated      time.Time `json:"last_updated"`
	ActiveAlerts     int       `json:"active_alerts"`
	ActiveWorkflows  int       `json:"active_workflows"`
	ConnectedSystems int       `json:"connected_systems"`
}

// Response types for Webhook Management
type WebhookCreateResult struct {
	WebhookID string    `json:"webhook_id"`
	Success   bool      `json:"success"`
	CreatedAt time.Time `json:"created_at"`
	Error     string    `json:"error,omitempty"`
}

type WebhookListResult struct {
	Webhooks []Webhook `json:"webhooks"`
	Total    int       `json:"total"`
	Active   int       `json:"active"`
}

type WebhookTestResult struct {
	WebhookID    string `json:"webhook_id"`
	Success      bool   `json:"success"`
	ResponseCode int    `json:"response_code"`
	ResponseTime string `json:"response_time"`
	Error        string `json:"error,omitempty"`
}

type WebhookDeleteResult struct {
	WebhookID string    `json:"webhook_id"`
	Success   bool      `json:"success"`
	DeletedAt time.Time `json:"deleted_at"`
	Error     string    `json:"error,omitempty"`
}

// Core types
type Integration struct {
	Name           string   `json:"name"`
	DisplayName    string   `json:"display_name"`
	Description    string   `json:"description"`
	Category       string   `json:"category"`
	Status         string   `json:"status"`
	Version        string   `json:"version"`
	Features       []string `json:"features"`
	RequiredConfig []string `json:"required_config"`
}

type Alert struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Type          string    `json:"type"`
	Severity      string    `json:"severity"`
	Threshold     float64   `json:"threshold"`
	Channels      []string  `json:"channels"`
	Status        string    `json:"status"`
	Enabled       bool      `json:"enabled"`
	Created       time.Time `json:"created"`
	LastTriggered time.Time `json:"last_triggered,omitempty"`
	TriggerCount  int       `json:"trigger_count"`
}

type Workflow struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Schedule    string         `json:"schedule"`
	Status      string         `json:"status"`
	Enabled     bool           `json:"enabled"`
	Steps       []WorkflowStep `json:"steps"`
	Created     time.Time      `json:"created"`
	LastRun     time.Time      `json:"last_run,omitempty"`
	NextRun     time.Time      `json:"next_run,omitempty"`
	RunCount    int            `json:"run_count"`
}

type WorkflowStep struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Type      string                 `json:"type"`
	Config    map[string]interface{} `json:"config"`
	Order     int                    `json:"order"`
	DependsOn []string               `json:"depends_on"`
}

type Webhook struct {
	ID            string            `json:"id"`
	Name          string            `json:"name"`
	URL           string            `json:"url"`
	Events        []string          `json:"events"`
	Headers       map[string]string `json:"headers"`
	Enabled       bool              `json:"enabled"`
	Created       time.Time         `json:"created"`
	LastTriggered time.Time         `json:"last_triggered,omitempty"`
	TriggerCount  int               `json:"trigger_count"`
}
