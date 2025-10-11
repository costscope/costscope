package integration

import (
	"time"
)

// IntegrationCategory represents different integration categories
type IntegrationCategory string

const (
	CategoryBilling      IntegrationCategory = "billing"
	CategoryITSM         IntegrationCategory = "itsm"
	CategoryBI           IntegrationCategory = "bi"
	CategoryMonitoring   IntegrationCategory = "monitoring"
	CategoryAutomation   IntegrationCategory = "automation"
	CategoryNotification IntegrationCategory = "notification"
	CategorySecurity     IntegrationCategory = "security"
	CategoryDevOps       IntegrationCategory = "devops"
)

// IntegrationConnectionStatus represents the status of a system connection
type IntegrationConnectionStatus string

const (
	StatusConnected    IntegrationConnectionStatus = "connected"
	StatusDisconnected IntegrationConnectionStatus = "disconnected"
	StatusConnecting   IntegrationConnectionStatus = "connecting"
	StatusError        IntegrationConnectionStatus = "error"
	StatusTesting      IntegrationConnectionStatus = "testing"
)

// AlertType represents different types of cost alerts
type AlertType string

const (
	AlertTypeBudget       AlertType = "budget"
	AlertTypeThreshold    AlertType = "threshold"
	AlertTypeAnomaly      AlertType = "anomaly"
	AlertTypeForecast     AlertType = "forecast"
	AlertTypeUsage        AlertType = "usage"
	AlertTypeCompliance   AlertType = "compliance"
	AlertTypeOptimization AlertType = "optimization"
)

// AlertSeverity represents alert severity levels
type AlertSeverity string

const (
	SeverityLow      AlertSeverity = "low"
	SeverityMedium   AlertSeverity = "medium"
	SeverityHigh     AlertSeverity = "high"
	SeverityCritical AlertSeverity = "critical"
)

// AlertChannel represents different notification channels
type AlertChannel string

const (
	ChannelEmail     AlertChannel = "email"
	ChannelSlack     AlertChannel = "slack"
	ChannelSMS       AlertChannel = "sms"
	ChannelWebhook   AlertChannel = "webhook"
	ChannelTeams     AlertChannel = "teams"
	ChannelDiscord   AlertChannel = "discord"
	ChannelDashboard AlertChannel = "dashboard"
	ChannelPagerDuty AlertChannel = "pagerduty"
)

// WorkflowStatus represents workflow execution status
type WorkflowStatus string

const (
	WorkflowStatusActive    WorkflowStatus = "active"
	WorkflowStatusInactive  WorkflowStatus = "inactive"
	WorkflowStatusRunning   WorkflowStatus = "running"
	WorkflowStatusCompleted WorkflowStatus = "completed"
	WorkflowStatusFailed    WorkflowStatus = "failed"
	WorkflowStatusScheduled WorkflowStatus = "scheduled"
	WorkflowStatusPaused    WorkflowStatus = "paused"
)

// WorkflowStepType represents different types of workflow steps
type WorkflowStepType string

const (
	StepTypeAnalysis     WorkflowStepType = "analysis"
	StepTypeOptimization WorkflowStepType = "optimization"
	StepTypeAlert        WorkflowStepType = "alert"
	StepTypeReport       WorkflowStepType = "report"
	StepTypeAction       WorkflowStepType = "action"
	StepTypeCondition    WorkflowStepType = "condition"
	StepTypeIntegration  WorkflowStepType = "integration"
	StepTypeNotification WorkflowStepType = "notification"
)

// WebhookEvent represents different webhook events
type WebhookEvent string

const (
	EventCostChange        WebhookEvent = "cost_change"
	EventBudgetExceeded    WebhookEvent = "budget_exceeded"
	EventAnomalyDetected   WebhookEvent = "anomaly_detected"
	EventReportGenerated   WebhookEvent = "report_generated"
	EventAlertTriggered    WebhookEvent = "alert_triggered"
	EventWorkflowCompleted WebhookEvent = "workflow_completed"
	EventSystemConnected   WebhookEvent = "system_connected"
	EventDashboardAccessed WebhookEvent = "dashboard_accessed"
)

// Supported third-party systems
type SystemType string

const (
	// Billing Systems
	SystemAWS         SystemType = "aws"
	SystemAzure       SystemType = "azure"
	SystemGCP         SystemType = "gcp"
	SystemCloudHealth SystemType = "cloudhealth"
	SystemCloudCheckr SystemType = "cloudcheckr"

	// ITSM Systems
	SystemServiceNow   SystemType = "servicenow"
	SystemJIRA         SystemType = "jira"
	SystemFreshService SystemType = "freshservice"
	SystemZendesk      SystemType = "zendesk"

	// BI Systems
	SystemTableau   SystemType = "tableau"
	SystemPowerBI   SystemType = "powerbi"
	SystemLooker    SystemType = "looker"
	SystemQlikSense SystemType = "qliksense"
	SystemGrafana   SystemType = "grafana"

	// Monitoring Systems
	SystemDatadog       SystemType = "datadog"
	SystemNewRelic      SystemType = "newrelic"
	SystemSplunk        SystemType = "splunk"
	SystemElasticSearch SystemType = "elasticsearch"
	SystemPrometheus    SystemType = "prometheus"

	// Notification Systems
	SystemSlack     SystemType = "slack"
	SystemTeams     SystemType = "teams"
	SystemDiscord   SystemType = "discord"
	SystemPagerDuty SystemType = "pagerduty"
	SystemTwilio    SystemType = "twilio"

	// Automation Systems
	SystemJenkins       SystemType = "jenkins"
	SystemGitHubActions SystemType = "github_actions"
	SystemGitLab        SystemType = "gitlab"
	SystemAnsible       SystemType = "ansible"
	SystemTerraform     SystemType = "terraform"
)

// SystemConfig represents configuration for different systems
type SystemConfig struct {
	Type       SystemType             `json:"type"`
	Name       string                 `json:"name"`
	BaseURL    string                 `json:"base_url,omitempty"`
	APIKey     string                 `json:"api_key,omitempty"`
	Username   string                 `json:"username,omitempty"`
	Password   string                 `json:"password,omitempty"`
	Token      string                 `json:"token,omitempty"`
	Region     string                 `json:"region,omitempty"`
	Options    map[string]interface{} `json:"options,omitempty"`
	TestMode   bool                   `json:"test_mode"`
	Timeout    time.Duration          `json:"timeout"`
	RetryCount int                    `json:"retry_count"`
}

// IntegrationMetrics represents metrics for integration monitoring
type IntegrationMetrics struct {
	SystemName       string    `json:"system_name"`
	ConnectionHealth float64   `json:"connection_health"`
	DataSyncRate     float64   `json:"data_sync_rate"`
	ErrorRate        float64   `json:"error_rate"`
	LastSync         time.Time `json:"last_sync"`
	TotalRequests    int64     `json:"total_requests"`
	FailedRequests   int64     `json:"failed_requests"`
	AverageLatency   string    `json:"average_latency"`
	Uptime           string    `json:"uptime"`
}

// DashboardTheme represents dashboard themes
type DashboardTheme string

const (
	ThemeLight DashboardTheme = "light"
	ThemeDark  DashboardTheme = "dark"
	ThemeAuto  DashboardTheme = "auto"
)

// WebhookConfig represents webhook configuration
type WebhookConfig struct {
	URL         string            `json:"url"`
	Secret      string            `json:"secret"`
	Headers     map[string]string `json:"headers"`
	Timeout     time.Duration     `json:"timeout"`
	RetryCount  int               `json:"retry_count"`
	RetryDelay  time.Duration     `json:"retry_delay"`
	VerifySSL   bool              `json:"verify_ssl"`
	Method      string            `json:"method"`
	ContentType string            `json:"content_type"`
}

// WorkflowSchedule represents different schedule types
type WorkflowSchedule struct {
	Type     string `json:"type"` // cron, interval, manual
	Value    string `json:"value"`
	Timezone string `json:"timezone"`
	Enabled  bool   `json:"enabled"`
}

// AlertCondition represents alert trigger conditions
type AlertCondition struct {
	Metric    string      `json:"metric"`
	Operator  string      `json:"operator"` // >, <, >=, <=, ==, !=
	Value     interface{} `json:"value"`
	Duration  string      `json:"duration"`
	Aggregate string      `json:"aggregate"` // avg, sum, min, max, count
}

// NotificationTemplate represents notification message templates
type NotificationTemplate struct {
	Type     string `json:"type"`
	Subject  string `json:"subject"`
	Body     string `json:"body"`
	Format   string `json:"format"` // text, html, markdown
	Language string `json:"language"`
}

// IntegrationError represents integration-specific errors
type IntegrationError struct {
	Code      string                 `json:"code"`
	Message   string                 `json:"message"`
	System    string                 `json:"system"`
	Timestamp time.Time              `json:"timestamp"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

// APIResponse represents a generic API response structure
type APIResponse struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Error     string      `json:"error,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
	RequestID string      `json:"request_id"`
}

// ConnectionPool represents a pool of system connections
type ConnectionPool struct {
	MaxConnections      int           `json:"max_connections"`
	IdleTimeout         time.Duration `json:"idle_timeout"`
	ConnectionTimeout   time.Duration `json:"connection_timeout"`
	HealthCheckInterval time.Duration `json:"health_check_interval"`
	RetryPolicy         RetryPolicy   `json:"retry_policy"`
}

// RetryPolicy represents retry configuration for failed operations
type RetryPolicy struct {
	MaxRetries    int           `json:"max_retries"`
	InitialDelay  time.Duration `json:"initial_delay"`
	MaxDelay      time.Duration `json:"max_delay"`
	BackoffFactor float64       `json:"backoff_factor"`
	Jitter        bool          `json:"jitter"`
}

// AuditLog represents audit logging for integration activities
type AuditLog struct {
	ID        string                 `json:"id"`
	Action    string                 `json:"action"`
	System    string                 `json:"system"`
	User      string                 `json:"user"`
	Timestamp time.Time              `json:"timestamp"`
	Details   map[string]interface{} `json:"details"`
	Success   bool                   `json:"success"`
	Error     string                 `json:"error,omitempty"`
}

// SecurityConfig represents security configuration for integrations
type SecurityConfig struct {
	EncryptionEnabled bool          `json:"encryption_enabled"`
	TLSVersion        string        `json:"tls_version"`
	CipherSuites      []string      `json:"cipher_suites"`
	CertificatePath   string        `json:"certificate_path"`
	KeyPath           string        `json:"key_path"`
	CAPath            string        `json:"ca_path"`
	TokenExpiration   time.Duration `json:"token_expiration"`
	MaxRequestSize    int64         `json:"max_request_size"`
}

// Performance metrics for integration monitoring
type PerformanceMetrics struct {
	RequestsPerSecond float64 `json:"requests_per_second"`
	AverageLatency    string  `json:"average_latency"`
	P95Latency        string  `json:"p95_latency"`
	P99Latency        string  `json:"p99_latency"`
	ErrorRate         float64 `json:"error_rate"`
	ThroughputMB      float64 `json:"throughput_mb"`
	ConcurrentUsers   int     `json:"concurrent_users"`
	MemoryUsage       string  `json:"memory_usage"`
	CPUUsage          float64 `json:"cpu_usage"`
}
