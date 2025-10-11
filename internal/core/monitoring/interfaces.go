package monitoring

import (
	"context"
	"time"
)

// MonitoringService defines the main monitoring service interface
type MonitoringService interface {
	// Real-time monitoring
	StartRealTimeMonitoring(ctx context.Context, config *MonitoringConfig) error
	StopRealTimeMonitoring(ctx context.Context) error
	GetRealTimeMetrics(ctx context.Context) (*RealTimeMetrics, error)

	// Health monitoring
	GetSystemHealth(ctx context.Context) (*SystemHealthStatus, error)
	GetComponentHealth(ctx context.Context, component string) (*ComponentHealth, error)
	RunHealthChecks(ctx context.Context, components []string) (*HealthCheckResults, error)

	// Performance monitoring
	GetPerformanceMetrics(ctx context.Context) (*PerformanceMetrics, error)
	GetPerformanceTrends(ctx context.Context, timeRange time.Duration) (*PerformanceTrends, error)

	// Alerting
	CreateAlert(ctx context.Context, alert *AlertDefinition) error
	GetActiveAlerts(ctx context.Context) ([]*Alert, error)
	ResolveAlert(ctx context.Context, alertID string) error

	// Dashboard
	GetDashboardData(ctx context.Context) (*DashboardData, error)
	GetMonitoringConfig(ctx context.Context) (*MonitoringConfig, error)
	UpdateMonitoringConfig(ctx context.Context, config *MonitoringConfig) error
}

// MetricsCollector collects various types of metrics
type MetricsCollector interface {
	// System metrics
	CollectSystemMetrics(ctx context.Context) (*SystemMetrics, error)
	CollectResourceMetrics(ctx context.Context) (*ResourceMetrics, error)

	// Application metrics
	CollectApplicationMetrics(ctx context.Context) (*ApplicationMetrics, error)
	CollectBusinessMetrics(ctx context.Context) (*BusinessMetrics, error)

	// External integration metrics
	CollectIntegrationMetrics(ctx context.Context) (*IntegrationMetrics, error)
	CollectProviderMetrics(ctx context.Context) (*ProviderMetrics, error)
}

// AlertManager manages alerting functionality
type AlertManager interface {
	// Alert lifecycle
	TriggerAlert(ctx context.Context, alert *Alert) error
	EscalateAlert(ctx context.Context, alertID string) error
	AcknowledgeAlert(ctx context.Context, alertID string, user string) error

	// Alert configuration
	CreateAlertRule(ctx context.Context, rule *AlertRule) error
	UpdateAlertRule(ctx context.Context, ruleID string, rule *AlertRule) error
	DeleteAlertRule(ctx context.Context, ruleID string) error
	ListAlertRules(ctx context.Context) ([]*AlertRule, error)
}

// DashboardRenderer renders monitoring dashboards
type DashboardRenderer interface {
	RenderSystemDashboard(ctx context.Context, data *DashboardData) (string, error)
	RenderMetricsDashboard(ctx context.Context, metrics *RealTimeMetrics) (string, error)
	RenderAlertsDashboard(ctx context.Context, alerts []*Alert) (string, error)
	RenderPerformanceDashboard(ctx context.Context, performance *PerformanceMetrics) (string, error)
}

// NotificationService sends notifications for alerts
type NotificationService interface {
	SendNotification(ctx context.Context, notification *Notification) error
	SendAlert(ctx context.Context, alert *Alert, channels []string) error
	ValidateChannel(ctx context.Context, channel string) error
	GetSupportedChannels() []string
}

// MetricEmitter emits metrics to an external sink (e.g., Prometheus, OTEL) or logs them.
// Implementations must be non-blocking and respect context cancellation.
type MetricEmitter interface {
	Emit(ctx context.Context, metrics *RealTimeMetrics) error
}

// AlertEvaluator converts current metrics + thresholds into alert candidates (no side-effects).
type AlertEvaluator interface {
	Evaluate(ctx context.Context, metrics *RealTimeMetrics, thresholds PerformanceThresholds) []*Alert
}

// HealthChecker abstracts health evaluation for system and components.
type HealthChecker interface {
	System(ctx context.Context) *SystemHealthStatus
	Component(ctx context.Context, component string) (*ComponentHealth, error)
	Run(ctx context.Context, components []string) (*HealthCheckResults, error)
}
