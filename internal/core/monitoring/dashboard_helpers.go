package monitoring

import (
	"time"
)

func (bms *BasicMonitoringService) generateTrendDataPoints(start, end time.Time, baseValue, variance float64) []DataPoint {
	points := make([]DataPoint, 0)
	duration := end.Sub(start)
	interval := duration / 50 // Generate 50 data points
	for i := 0; i < 50; i++ {
		timestamp := start.Add(time.Duration(i) * interval)
		value := baseValue + (float64(i%10-5) * variance / 10)
		points = append(points, DataPoint{Timestamp: timestamp, Value: value})
	}
	return points
}

func (bms *BasicMonitoringService) generatePerformanceCharts() []ChartData {
	return []ChartData{
		{ChartID: "cpu_usage", Title: "CPU Usage", Type: "line", TimeRange: 1 * time.Hour, RefreshInterval: 30 * time.Second},
		{ChartID: "memory_usage", Title: "Memory Usage", Type: "line", TimeRange: 1 * time.Hour, RefreshInterval: 30 * time.Second},
		{ChartID: "response_time", Title: "Response Time", Type: "line", TimeRange: 1 * time.Hour, RefreshInterval: 30 * time.Second},
	}
}

func (bms *BasicMonitoringService) generateHealthIndicators(health *SystemHealthStatus, metrics *RealTimeMetrics) []HealthIndicator {
	return []HealthIndicator{
		{Name: "System Health", Status: health.OverallHealth, Value: float64(health.HealthScore), Threshold: 80.0, Unit: "score", Trend: "stable", LastUpdated: time.Now()},
		{Name: "CPU Usage", Status: bms.getStatusFromValue(metrics.Performance.CPU.UsagePercent, 70, 90), Value: metrics.Performance.CPU.UsagePercent, Threshold: 70.0, Unit: "%", Trend: "increasing", LastUpdated: time.Now()},
		{Name: "Memory Usage", Status: bms.getStatusFromValue(metrics.Performance.Memory.UsagePercent, 80, 95), Value: metrics.Performance.Memory.UsagePercent, Threshold: 80.0, Unit: "%", Trend: "stable", LastUpdated: time.Now()},
	}
}

func (bms *BasicMonitoringService) generateTopComponents(health *SystemHealthStatus) []ComponentSummary {
	components := make([]ComponentSummary, 0)
	for name, status := range health.ComponentHealth {
		var score float64
		switch status {
		case DegradedStatus:
			score = 65.0
		case "critical":
			score = 35.0
		default:
			score = 85.0
		}
		components = append(components, ComponentSummary{Name: name, Status: status, Health: score, Performance: score, LastUpdated: time.Now(), AlertCount: 0})
	}
	return components
}

func (bms *BasicMonitoringService) generateTrendData() []TrendIndicator {
	return []TrendIndicator{
		{Metric: "CPU Usage", CurrentValue: 45.2, PreviousValue: 42.1, ChangePercent: 7.4, Trend: "increasing", Significance: "moderate"},
		{Metric: "Memory Usage", CurrentValue: 68.5, PreviousValue: 69.1, ChangePercent: -0.9, Trend: "stable", Significance: "low"},
		{Metric: "Response Time", CurrentValue: 45.2, PreviousValue: 52.8, ChangePercent: -14.4, Trend: "improving", Significance: "high"},
	}
}

func (bms *BasicMonitoringService) generateKPIMetrics(metrics *RealTimeMetrics) []KPIMetric {
	return []KPIMetric{
		{Name: "System Availability", Value: 99.85, Target: 99.9, Unit: "%", Trend: "stable", Variance: -0.05, Status: "warning"},
		{Name: "Error Rate", Value: metrics.Performance.Application.ErrorRate, Target: 0.5, Unit: "%", Trend: "stable", Variance: metrics.Performance.Application.ErrorRate - 0.5, Status: bms.getStatusFromValue(metrics.Performance.Application.ErrorRate, 1.0, 5.0)},
		{Name: "Response Time", Value: metrics.Performance.Application.ResponseTime.Average, Target: 50.0, Unit: "ms", Trend: "improving", Variance: metrics.Performance.Application.ResponseTime.Average - 50.0, Status: bms.getStatusFromValue(metrics.Performance.Application.ResponseTime.Average, 100, 500)},
	}
}

func (bms *BasicMonitoringService) generateQuickActions() []QuickAction {
	return []QuickAction{
		{ID: "restart_service", Title: "Restart Service", Description: "Restart a specific service", Action: "restart", Icon: "restart", Enabled: true, RequiresConfirmation: true},
		{ID: "clear_cache", Title: "Clear Cache", Description: "Clear system cache", Action: "clear_cache", Icon: "clear", Enabled: true, RequiresConfirmation: false},
		{ID: "scale_up", Title: "Scale Up", Description: "Scale up resources", Action: "scale_up", Icon: "scale", Enabled: true, RequiresConfirmation: true},
	}
}
