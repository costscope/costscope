package monitoring

import (
	"time"
)

// MonitoringConfig defines monitoring service configuration
type MonitoringConfig struct {
	EnableRealTime        bool                  `json:"enable_real_time"`
	MetricsInterval       time.Duration         `json:"metrics_interval"`
	AlertingEnabled       bool                  `json:"alerting_enabled"`
	NotificationChannels  []string              `json:"notification_channels"`
	DashboardPort         int                   `json:"dashboard_port"`
	RetentionPeriod       time.Duration         `json:"retention_period"`
	HealthCheckInterval   time.Duration         `json:"health_check_interval"`
	PerformanceThresholds PerformanceThresholds `json:"performance_thresholds"`
	AlertRules            []AlertRule           `json:"alert_rules"`
	CustomMetrics         []CustomMetricConfig  `json:"custom_metrics"`
}

// Additional supporting types (shared across modules)
type PerformanceThresholds struct {
	CPUWarning        float64 `json:"cpu_warning"`
	CPUCritical       float64 `json:"cpu_critical"`
	MemoryWarning     float64 `json:"memory_warning"`
	MemoryCritical    float64 `json:"memory_critical"`
	DiskWarning       float64 `json:"disk_warning"`
	DiskCritical      float64 `json:"disk_critical"`
	LatencyWarning    float64 `json:"latency_warning"`
	LatencyCritical   float64 `json:"latency_critical"`
	ErrorRateWarning  float64 `json:"error_rate_warning"`
	ErrorRateCritical float64 `json:"error_rate_critical"`
}

type CustomMetricConfig struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Source  string `json:"source"`
	Query   string `json:"query"`
	Unit    string `json:"unit"`
	Enabled bool   `json:"enabled"`
}

type TrendData struct {
	MetricName    string      `json:"metric_name"`
	DataPoints    []DataPoint `json:"data_points"`
	Trend         string      `json:"trend"`
	ChangePercent float64     `json:"change_percent"`
	Slope         float64     `json:"slope"`
	Correlation   float64     `json:"correlation"`
}

type DataPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}

type Anomaly struct {
	Timestamp     time.Time `json:"timestamp"`
	MetricName    string    `json:"metric_name"`
	Value         float64   `json:"value"`
	ExpectedValue float64   `json:"expected_value"`
	Deviation     float64   `json:"deviation"`
	Severity      string    `json:"severity"`
	Description   string    `json:"description"`
}

type Notification struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Channel     string                 `json:"channel"`
	Recipient   string                 `json:"recipient"`
	Subject     string                 `json:"subject"`
	Message     string                 `json:"message"`
	Severity    string                 `json:"severity"`
	CreatedAt   time.Time              `json:"created_at"`
	SentAt      *time.Time             `json:"sent_at,omitempty"`
	DeliveredAt *time.Time             `json:"delivered_at,omitempty"`
	Status      string                 `json:"status"`
	RetryCount  int                    `json:"retry_count"`
	Metadata    map[string]interface{} `json:"metadata"`
}
