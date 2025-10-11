package monitoring

import (
	"context"
	"fmt"
	"time"
)

// generateDashboardData was extracted from service.go to reduce LOC without changing behavior.
func (bms *BasicMonitoringService) generateDashboardData(ctx context.Context) (*DashboardData, error) {
	realTimeMetrics, err := bms.GetRealTimeMetrics(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get real-time metrics: %w", err)
	}

	systemHealth, err := bms.GetSystemHealth(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get system health: %w", err)
	}

	dashboard := &DashboardData{
		Timestamp: time.Now(),
		SystemOverview: SystemOverview{
			OverallHealth:    systemHealth.OverallHealth,
			HealthScore:      systemHealth.HealthScore,
			ActiveComponents: len(systemHealth.HealthyComponents) + len(systemHealth.DegradedComponents),
			TotalComponents:  len(systemHealth.ComponentHealth),
			CriticalAlerts:   len(bms.activeAlerts),
			ActiveUsers:      realTimeMetrics.Applications.ConcurrentUsers,
			Uptime:           fmt.Sprintf("%.1fh", systemHealth.UptimeHours),
			Version:          "1.0.0",
			Environment:      "production",
		},
		RealTimeMetrics:   *realTimeMetrics,
		PerformanceCharts: bms.generatePerformanceCharts(),
		HealthIndicators:  bms.generateHealthIndicators(systemHealth, realTimeMetrics),
		AlertsSummary: AlertsSummary{
			TotalAlerts:    len(bms.activeAlerts),
			CriticalAlerts: bms.countAlertsBySeverity("critical"),
			WarningAlerts:  bms.countAlertsBySeverity("warning"),
			InfoAlerts:     bms.countAlertsBySeverity("info"),
			RecentAlerts:   bms.getRecentAlerts(5),
			AlertTrend:     "stable",
		},
		TopComponents: bms.generateTopComponents(systemHealth),
		TrendData:     bms.generateTrendData(),
		KPIMetrics:    bms.generateKPIMetrics(realTimeMetrics),
		QuickActions:  bms.generateQuickActions(),
	}

	return dashboard, nil
}
