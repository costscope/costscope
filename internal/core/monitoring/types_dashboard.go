package monitoring

import "time"

// Dashboard visualization types
type DashboardData struct {
	Timestamp         time.Time          `json:"timestamp"`
	SystemOverview    SystemOverview     `json:"system_overview"`
	RealTimeMetrics   RealTimeMetrics    `json:"real_time_metrics"`
	PerformanceCharts []ChartData        `json:"performance_charts"`
	HealthIndicators  []HealthIndicator  `json:"health_indicators"`
	AlertsSummary     AlertsSummary      `json:"alerts_summary"`
	TopComponents     []ComponentSummary `json:"top_components"`
	TrendData         []TrendIndicator   `json:"trend_data"`
	KPIMetrics        []KPIMetric        `json:"kpi_metrics"`
	QuickActions      []QuickAction      `json:"quick_actions"`
}

type SystemOverview struct {
	OverallHealth    string `json:"overall_health"`
	HealthScore      int    `json:"health_score"`
	ActiveComponents int    `json:"active_components"`
	TotalComponents  int    `json:"total_components"`
	CriticalAlerts   int    `json:"critical_alerts"`
	ActiveUsers      int    `json:"active_users"`
	Uptime           string `json:"uptime"`
	Version          string `json:"version"`
	Environment      string `json:"environment"`
}
type ChartData struct {
	ChartID         string        `json:"chart_id"`
	Title           string        `json:"title"`
	Type            string        `json:"type"`
	Data            []DataSeries  `json:"data"`
	TimeRange       time.Duration `json:"time_range"`
	RefreshInterval time.Duration `json:"refresh_interval"`
	Thresholds      []Threshold   `json:"thresholds"`
}

type DataSeries struct {
	Name  string      `json:"name"`
	Data  []DataPoint `json:"data"`
	Color string      `json:"color"`
	Unit  string      `json:"unit"`
}

type Threshold struct {
	Value float64 `json:"value"`
	Color string  `json:"color"`
	Label string  `json:"label"`
	Type  string  `json:"type"`
}

type HealthIndicator struct {
	Name        string    `json:"name"`
	Status      string    `json:"status"`
	Value       float64   `json:"value"`
	Threshold   float64   `json:"threshold"`
	Unit        string    `json:"unit"`
	Trend       string    `json:"trend"`
	LastUpdated time.Time `json:"last_updated"`
}

type AlertsSummary struct {
	TotalAlerts       int            `json:"total_alerts"`
	CriticalAlerts    int            `json:"critical_alerts"`
	WarningAlerts     int            `json:"warning_alerts"`
	InfoAlerts        int            `json:"info_alerts"`
	RecentAlerts      []Alert        `json:"recent_alerts"`
	AlertsByComponent map[string]int `json:"alerts_by_component"`
	AlertTrend        string         `json:"alert_trend"`
}

type ComponentSummary struct {
	Name        string    `json:"name"`
	Status      string    `json:"status"`
	Health      float64   `json:"health"`
	Performance float64   `json:"performance"`
	LastUpdated time.Time `json:"last_updated"`
	AlertCount  int       `json:"alert_count"`
}

type TrendIndicator struct {
	Metric        string  `json:"metric"`
	CurrentValue  float64 `json:"current_value"`
	PreviousValue float64 `json:"previous_value"`
	ChangePercent float64 `json:"change_percent"`
	Trend         string  `json:"trend"`
	Significance  string  `json:"significance"`
}

type KPIMetric struct {
	Name     string  `json:"name"`
	Value    float64 `json:"value"`
	Target   float64 `json:"target"`
	Unit     string  `json:"unit"`
	Trend    string  `json:"trend"`
	Variance float64 `json:"variance"`
	Status   string  `json:"status"`
}

type QuickAction struct {
	ID                   string `json:"id"`
	Title                string `json:"title"`
	Description          string `json:"description"`
	Action               string `json:"action"`
	Icon                 string `json:"icon"`
	Enabled              bool   `json:"enabled"`
	RequiresConfirmation bool   `json:"requires_confirmation"`
}
