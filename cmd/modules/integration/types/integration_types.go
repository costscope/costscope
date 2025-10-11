package types

import "time"

// Integration represents a third-party system integration with enhanced metadata
type Integration struct {
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Category    string              `json:"category"`
	Status      string              `json:"status"`
	Version     string              `json:"version"`
	LastUpdated time.Time           `json:"last_updated"`
	Features    []string            `json:"features"`
	Config      map[string]string   `json:"config"`
	Health      *HealthStatus       `json:"health,omitempty"`
	Metrics     *IntegrationMetrics `json:"metrics,omitempty"`
}

// HealthStatus represents the health status of an integration
type HealthStatus struct {
	Status       string    `json:"status"`
	LastCheck    time.Time `json:"last_check"`
	ResponseTime float64   `json:"response_time_ms"`
	ErrorCount   int       `json:"error_count"`
	Uptime       float64   `json:"uptime_percent"`
}

// IntegrationMetrics represents performance metrics for an integration
type IntegrationMetrics struct {
	RequestsPerMinute float64   `json:"requests_per_minute"`
	AverageLatency    float64   `json:"average_latency_ms"`
	ErrorRate         float64   `json:"error_rate_percent"`
	DataTransferred   int64     `json:"data_transferred_bytes"`
	LastDataSync      time.Time `json:"last_data_sync"`
	SyncFrequency     string    `json:"sync_frequency"`
}

// ConnectionResult represents the result of connecting to a third-party system
type ConnectionResult struct {
	Success          bool                   `json:"success"`
	Status           string                 `json:"status"`
	DataSync         string                 `json:"data_sync"`
	AvailableMetrics int                    `json:"available_metrics"`
	Features         []string               `json:"features"`
	Error            string                 `json:"error,omitempty"`
	ConnectionID     string                 `json:"connection_id,omitempty"`
	EstablishedAt    time.Time              `json:"established_at"`
	Configuration    map[string]interface{} `json:"configuration"`
	HealthCheck      *HealthStatus          `json:"health_check,omitempty"`
}

// Enhanced Workflow types
type Workflow struct {
	ID            string               `json:"id"`
	Name          string               `json:"name"`
	Description   string               `json:"description"`
	Schedule      string               `json:"schedule"`
	Status        string               `json:"status"`
	LastRun       time.Time            `json:"last_run"`
	NextRun       time.Time            `json:"next_run"`
	Steps         []WorkflowStep       `json:"steps"`
	Triggers      []WorkflowTrigger    `json:"triggers"`
	Notifications []NotificationConfig `json:"notifications"`
	Metrics       *WorkflowMetrics     `json:"metrics,omitempty"`
}

// WorkflowStep represents a single step in a workflow
type WorkflowStep struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Config      map[string]interface{} `json:"config"`
	DependsOn   []string               `json:"depends_on"`
	Timeout     time.Duration          `json:"timeout"`
	RetryPolicy *RetryPolicy           `json:"retry_policy,omitempty"`
}

// WorkflowTrigger represents a trigger for workflow execution
type WorkflowTrigger struct {
	Type      string                 `json:"type"`
	Condition string                 `json:"condition"`
	Config    map[string]interface{} `json:"config"`
}

// RetryPolicy defines retry behavior for workflow steps
type RetryPolicy struct {
	MaxRetries  int           `json:"max_retries"`
	BackoffType string        `json:"backoff_type"`
	BaseDelay   time.Duration `json:"base_delay"`
	MaxDelay    time.Duration `json:"max_delay"`
}

// WorkflowMetrics represents workflow execution metrics
type WorkflowMetrics struct {
	TotalExecutions      int           `json:"total_executions"`
	SuccessfulRuns       int           `json:"successful_runs"`
	FailedRuns           int           `json:"failed_runs"`
	AverageExecutionTime time.Duration `json:"average_execution_time"`
	LastExecutionTime    time.Duration `json:"last_execution_time"`
	CostSavings          float64       `json:"cost_savings"`
}

// WorkflowResult represents the result of workflow execution
type WorkflowResult struct {
	WorkflowID    string               `json:"workflow_id"`
	ExecutionID   string               `json:"execution_id"`
	StartTime     time.Time            `json:"start_time"`
	EndTime       time.Time            `json:"end_time"`
	Duration      time.Duration        `json:"duration"`
	Status        string               `json:"status"`
	TasksExecuted int                  `json:"tasks_executed"`
	CostSavings   float64              `json:"cost_savings"`
	StepResults   []WorkflowStepResult `json:"step_results"`
	Error         string               `json:"error,omitempty"`
}

// WorkflowStepResult represents the result of a single workflow step
type WorkflowStepResult struct {
	StepID   string        `json:"step_id"`
	Status   string        `json:"status"`
	Duration time.Duration `json:"duration"`
	Output   interface{}   `json:"output,omitempty"`
	Error    string        `json:"error,omitempty"`
}

// Enhanced Alert types
type Alert struct {
	ID         string                `json:"id"`
	Name       string                `json:"name"`
	Type       string                `json:"type"`
	Severity   string                `json:"severity"`
	Threshold  float64               `json:"threshold"`
	Created    time.Time             `json:"created"`
	Updated    time.Time             `json:"updated"`
	Status     string                `json:"status"`
	Conditions []AlertCondition      `json:"conditions"`
	Actions    []AlertAction         `json:"actions"`
	Channels   []NotificationChannel `json:"channels"`
	Metrics    *AlertMetrics         `json:"metrics,omitempty"`
	Schedule   *AlertSchedule        `json:"schedule,omitempty"`
}

// AlertCondition represents a condition that triggers an alert
type AlertCondition struct {
	Metric    string  `json:"metric"`
	Operator  string  `json:"operator"`
	Value     float64 `json:"value"`
	Period    string  `json:"period"`
	Aggregate string  `json:"aggregate"`
}

// AlertAction represents an action to take when an alert is triggered
type AlertAction struct {
	Type   string                 `json:"type"`
	Config map[string]interface{} `json:"config"`
}

// AlertSchedule defines when alerts should be active
type AlertSchedule struct {
	Enabled   bool     `json:"enabled"`
	TimeZone  string   `json:"timezone"`
	Days      []string `json:"days"`
	StartTime string   `json:"start_time"`
	EndTime   string   `json:"end_time"`
}

// AlertMetrics represents alert performance metrics
type AlertMetrics struct {
	TriggeredCount    int           `json:"triggered_count"`
	LastTriggered     time.Time     `json:"last_triggered"`
	AverageResolution time.Duration `json:"average_resolution"`
	FalsePositives    int           `json:"false_positives"`
}

// NotificationChannel represents a channel for sending notifications
type NotificationChannel struct {
	ID     string                 `json:"id"`
	Type   string                 `json:"type"`
	Name   string                 `json:"name"`
	Config map[string]interface{} `json:"config"`
	Status string                 `json:"status"`
}

// NotificationConfig represents notification configuration
type NotificationConfig struct {
	ChannelID string                 `json:"channel_id"`
	Template  string                 `json:"template"`
	Config    map[string]interface{} `json:"config"`
}

// AlertTestResults represents the results of testing alert systems
type AlertTestResults struct {
	Email     AlertTestResult `json:"email"`
	Slack     AlertTestResult `json:"slack"`
	SMS       AlertTestResult `json:"sms"`
	Webhook   AlertTestResult `json:"webhook"`
	Dashboard AlertTestResult `json:"dashboard"`
	Overall   string          `json:"overall"`
}

// AlertTestResult represents the result of testing a single alert channel
type AlertTestResult struct {
	Status       string        `json:"status"`
	ResponseTime time.Duration `json:"response_time"`
	Error        string        `json:"error,omitempty"`
}

// Enhanced Dashboard types
type DashboardConfig struct {
	Port     int                `json:"port"`
	Theme    string             `json:"theme"`
	AutoOpen bool               `json:"auto_open"`
	Security *DashboardSecurity `json:"security,omitempty"`
	Features []string           `json:"features"`
	Plugins  []DashboardPlugin  `json:"plugins"`
	Layout   *DashboardLayout   `json:"layout,omitempty"`
}

// DashboardSecurity represents security configuration for the dashboard
type DashboardSecurity struct {
	Enabled      bool     `json:"enabled"`
	AuthType     string   `json:"auth_type"`
	AllowedIPs   []string `json:"allowed_ips"`
	RequireHTTPS bool     `json:"require_https"`
}

// DashboardPlugin represents a dashboard plugin
type DashboardPlugin struct {
	Name    string                 `json:"name"`
	Version string                 `json:"version"`
	Config  map[string]interface{} `json:"config"`
	Enabled bool                   `json:"enabled"`
}

// DashboardLayout represents dashboard layout configuration
type DashboardLayout struct {
	Columns int               `json:"columns"`
	Widgets []DashboardWidget `json:"widgets"`
}

// DashboardWidget represents a dashboard widget
type DashboardWidget struct {
	ID       string                 `json:"id"`
	Type     string                 `json:"type"`
	Title    string                 `json:"title"`
	Position WidgetPosition         `json:"position"`
	Config   map[string]interface{} `json:"config"`
}

// WidgetPosition represents widget position on dashboard
type WidgetPosition struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

// DashboardMetrics represents enhanced dashboard metrics and statistics
type DashboardMetrics struct {
	TotalCost    float64               `json:"total_cost"`
	MonthlyCost  float64               `json:"monthly_cost"`
	CostTrend    string                `json:"cost_trend"`
	CostChange   float64               `json:"cost_change_percent"`
	TopServices  []ServiceCost         `json:"top_services"`
	Predictions  *CostPredictions      `json:"predictions,omitempty"`
	LastUpdated  time.Time             `json:"last_updated"`
	ActiveAlerts int                   `json:"active_alerts"`
	ActiveUsers  int                   `json:"active_users"`
	Performance  *DashboardPerformance `json:"performance,omitempty"`
}

// ServiceCost represents cost information for a service
type ServiceCost struct {
	Service string  `json:"service"`
	Cost    float64 `json:"cost"`
	Change  float64 `json:"change_percent"`
}

// CostPredictions represents cost prediction data
type CostPredictions struct {
	NextMonth   float64 `json:"next_month"`
	NextQuarter float64 `json:"next_quarter"`
	YearEnd     float64 `json:"year_end"`
	Confidence  float64 `json:"confidence"`
}

// DashboardPerformance represents dashboard performance metrics
type DashboardPerformance struct {
	LoadTime       time.Duration `json:"load_time"`
	DataFreshness  time.Duration `json:"data_freshness"`
	CacheHitRate   float64       `json:"cache_hit_rate"`
	ActiveSessions int           `json:"active_sessions"`
}

// Enhanced Webhook types
type WebhookConfig struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	URL         string            `json:"url"`
	Method      string            `json:"method"`
	Headers     map[string]string `json:"headers"`
	Timeout     time.Duration     `json:"timeout"`
	RetryPolicy *RetryPolicy      `json:"retry_policy,omitempty"`
	Security    *WebhookSecurity  `json:"security,omitempty"`
	Events      []string          `json:"events"`
	Status      string            `json:"status"`
	Created     time.Time         `json:"created"`
	LastUsed    time.Time         `json:"last_used"`
}

// WebhookSecurity represents webhook security configuration
type WebhookSecurity struct {
	SignatureHeader string `json:"signature_header"`
	Secret          string `json:"secret"`
	Algorithm       string `json:"algorithm"`
}

// Constants for integration categories
const (
	CategoryBilling      = "billing"
	CategoryITSM         = "itsm"
	CategoryBI           = "bi"
	CategoryMonitoring   = "monitoring"
	CategoryAutomation   = "automation"
	CategoryNotification = "notification"
	CategorySecurity     = "security"
	CategoryCompliance   = "compliance"
)

// Constants for integration status
const (
	StatusAvailable    = "available"
	StatusConnected    = "connected"
	StatusDisconnected = "disconnected"
	StatusError        = "error"
	StatusMaintenance  = "maintenance"
)

// Constants for workflow status
const (
	WorkflowStatusActive    = "active"
	WorkflowStatusPaused    = "paused"
	WorkflowStatusCompleted = "completed"
	WorkflowStatusFailed    = "failed"
	WorkflowStatusScheduled = "scheduled"
)

// Constants for alert severity
const (
	AlertSeverityLow      = "low"
	AlertSeverityMedium   = "medium"
	AlertSeverityHigh     = "high"
	AlertSeverityCritical = "critical"
)

// Constants for alert status
const (
	AlertStatusActive     = "active"
	AlertStatusTriggered  = "triggered"
	AlertStatusResolved   = "resolved"
	AlertStatusSuppressed = "suppressed"
)

// Constants for health status
const (
	HealthStatusHealthy   = "healthy"
	HealthStatusDegraded  = "degraded"
	HealthStatusUnhealthy = "unhealthy"
	HealthStatusUnknown   = "unknown"
)
