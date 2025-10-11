package monitoring

import "time"

// SystemHealthStatus represents overall system health
type SystemHealthStatus struct {
	OverallHealth      string            `json:"overall_health"`
	HealthScore        int               `json:"health_score"`
	ComponentHealth    map[string]string `json:"component_health"`
	CriticalComponents []string          `json:"critical_components"`
	HealthyComponents  []string          `json:"healthy_components"`
	DegradedComponents []string          `json:"degraded_components"`
	FailedComponents   []string          `json:"failed_components"`
	LastHealthCheck    time.Time         `json:"last_health_check"`
	UptimeHours        float64           `json:"uptime_hours"`
	SystemLoad         float64           `json:"system_load"`
	MemoryPressure     float64           `json:"memory_pressure"`
	DiskPressure       float64           `json:"disk_pressure"`
}

// ComponentHealth represents health of individual component
type ComponentHealth struct {
	ComponentName   string                 `json:"component_name"`
	Status          string                 `json:"status"`
	HealthScore     int                    `json:"health_score"`
	LastChecked     time.Time              `json:"last_checked"`
	ResponseTime    float64                `json:"response_time"`
	ErrorRate       float64                `json:"error_rate"`
	Dependencies    []string               `json:"dependencies"`
	Metrics         map[string]interface{} `json:"metrics"`
	Issues          []string               `json:"issues"`
	Recommendations []string               `json:"recommendations"`
}

// HealthCheckResults contains results of health checks
type HealthCheckResults struct {
	OverallHealth    string          `json:"overall_health"`
	OverallScore     int             `json:"overall_score"`
	ComponentResults map[string]bool `json:"component_results"`
	FailedChecks     []string        `json:"failed_checks"`
	CheckTimestamp   time.Time       `json:"check_timestamp"`
	CheckDuration    time.Duration   `json:"check_duration"`
	TotalChecks      int             `json:"total_checks"`
	PassedChecks     int             `json:"passed_checks"`
	FailedCheckCount int             `json:"failed_check_count"`
	WarningChecks    []string        `json:"warning_checks"`
}
