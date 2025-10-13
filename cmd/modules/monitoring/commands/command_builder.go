package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/costscope/costscope/internal/core/integration"
	"github.com/costscope/costscope/internal/core/logging"
	"github.com/costscope/costscope/internal/core/monitoring"
	"github.com/costscope/costscope/internal/core/production"
)

// MonitoringCommands holds monitoring command functionality
type MonitoringCommands struct {
	monitoringService monitoring.MonitoringService
	logger            *logging.Logger
}

// NewMonitoringCommands creates monitoring commands
func NewMonitoringCommands(
	productionService production.ProductionService,
	integrationService integration.IntegrationService,
	logger *logging.Logger,
) *MonitoringCommands {
	return &MonitoringCommands{
		monitoringService: monitoring.NewBasicMonitoringService(logger, productionService, integrationService),
		logger:            logger,
	}
}

// BuildCommands creates all monitoring CLI commands
func (mc *MonitoringCommands) BuildCommands() *cobra.Command {
	monitoringCmd := &cobra.Command{
		Use:   "monitoring",
		Short: "Production monitoring and metrics system",
		Long: `Production monitoring and metrics system for real-time system health,
performance monitoring, alerting, and dashboard functionality.

This module provides comprehensive monitoring capabilities including:
- Real-time metrics collection
- System health monitoring  
- Performance trend analysis
- Alert management
- Dashboard data generation
- Notification services`,
	}

	// Add subcommands
	monitoringCmd.AddCommand(mc.buildStartCommand())
	monitoringCmd.AddCommand(mc.buildStopCommand())
	monitoringCmd.AddCommand(mc.buildStatusCommand())
	monitoringCmd.AddCommand(mc.buildMetricsCommand())
	monitoringCmd.AddCommand(mc.buildHealthCommand())
	monitoringCmd.AddCommand(mc.buildPerformanceCommand())
	monitoringCmd.AddCommand(mc.buildAlertsCommand())
	monitoringCmd.AddCommand(mc.buildDashboardCommand())
	monitoringCmd.AddCommand(mc.buildTrendsCommand())
	monitoringCmd.AddCommand(mc.buildConfigCommand())

	return monitoringCmd
}

// buildStartCommand creates start monitoring command
func (mc *MonitoringCommands) buildStartCommand() *cobra.Command {
	var (
		metricsInterval      string
		healthInterval       string
		enableAlerting       bool
		dashboardPort        int
		notificationChannels []string
		outputFormat         string
	)

	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start real-time monitoring",
		Long: `Start real-time monitoring with configurable intervals and options.

This command starts the monitoring service with real-time metrics collection,
health checks, alerting, and dashboard functionality.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Parse intervals
			metricsIntv, err := time.ParseDuration(metricsInterval)
			if err != nil {
				return fmt.Errorf("invalid metrics interval: %w", err)
			}

			healthIntv, err := time.ParseDuration(healthInterval)
			if err != nil {
				return fmt.Errorf("invalid health interval: %w", err)
			}

			// Create monitoring configuration
			config := &monitoring.MonitoringConfig{
				EnableRealTime:       true,
				MetricsInterval:      metricsIntv,
				AlertingEnabled:      enableAlerting,
				NotificationChannels: notificationChannels,
				DashboardPort:        dashboardPort,
				RetentionPeriod:      24 * time.Hour,
				HealthCheckInterval:  healthIntv,
				PerformanceThresholds: monitoring.PerformanceThresholds{
					CPUWarning:        70.0,
					CPUCritical:       90.0,
					MemoryWarning:     80.0,
					MemoryCritical:    95.0,
					DiskWarning:       85.0,
					DiskCritical:      95.0,
					LatencyWarning:    100.0,
					LatencyCritical:   500.0,
					ErrorRateWarning:  1.0,
					ErrorRateCritical: 5.0,
				},
			}

			// Start monitoring
			ctx := context.Background()
			err = mc.monitoringService.StartRealTimeMonitoring(ctx, config)
			if err != nil {
				return fmt.Errorf("failed to start monitoring: %w", err)
			}

			result := map[string]interface{}{
				"status":                "started",
				"metrics_interval":      metricsInterval,
				"health_interval":       healthInterval,
				"alerting_enabled":      enableAlerting,
				"dashboard_port":        dashboardPort,
				"notification_channels": notificationChannels,
				"started_at":            time.Now(),
			}

			return mc.outputResult(result, outputFormat)
		},
	}

	cmd.Flags().StringVar(&metricsInterval, "metrics-interval", "30s", "Metrics collection interval")
	cmd.Flags().StringVar(&healthInterval, "health-interval", "1m", "Health check interval")
	cmd.Flags().BoolVar(&enableAlerting, "enable-alerting", true, "Enable alerting")
	cmd.Flags().IntVar(&dashboardPort, "dashboard-port", 8081, "Dashboard server port")
	cmd.Flags().StringSliceVar(&notificationChannels, "notification-channels", []string{"email", "slack"}, "Notification channels")
	cmd.Flags().StringVar(&outputFormat, "output", monitoring.FormatTable, "Output format: table, json")

	return cmd
}

// buildStopCommand creates stop monitoring command
func (mc *MonitoringCommands) buildStopCommand() *cobra.Command {
	var outputFormat string

	cmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop real-time monitoring",
		Long:  `Stop the real-time monitoring service and cleanup resources.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			err := mc.monitoringService.StopRealTimeMonitoring(ctx)
			if err != nil {
				return fmt.Errorf("failed to stop monitoring: %w", err)
			}

			result := map[string]interface{}{
				"status":     "stopped",
				"stopped_at": time.Now(),
			}

			return mc.outputResult(result, outputFormat)
		},
	}

	cmd.Flags().StringVar(&outputFormat, "output", monitoring.FormatTable, "Output format: table, json")

	return cmd
}

// buildStatusCommand creates monitoring status command
func (mc *MonitoringCommands) buildStatusCommand() *cobra.Command {
	var (
		outputFormat string
		detailed     bool
	)

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Get monitoring system status",
		Long: `Get current status of the monitoring system including configuration,
active alerts, and real-time metrics.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			// Get monitoring configuration
			config, err := mc.monitoringService.GetMonitoringConfig(ctx)
			if err != nil {
				return fmt.Errorf("failed to get monitoring config: %w", err)
			}

			// Get active alerts
			alerts, err := mc.monitoringService.GetActiveAlerts(ctx)
			if err != nil {
				return fmt.Errorf("failed to get active alerts: %w", err)
			}

			// Get real-time metrics if detailed
			var metrics *monitoring.RealTimeMetrics
			if detailed {
				metrics, err = mc.monitoringService.GetRealTimeMetrics(ctx)
				if err != nil {
					return fmt.Errorf("failed to get real-time metrics: %w", err)
				}
			}

			result := map[string]interface{}{
				"monitoring_enabled":    config.EnableRealTime,
				"metrics_interval":      config.MetricsInterval.String(),
				"health_check_interval": config.HealthCheckInterval.String(),
				"alerting_enabled":      config.AlertingEnabled,
				"dashboard_port":        config.DashboardPort,
				"notification_channels": config.NotificationChannels,
				"active_alerts_count":   len(alerts),
				"status_timestamp":      time.Now(),
			}

			if detailed && metrics != nil {
				result["real_time_metrics"] = metrics
			}

			if outputFormat == monitoring.FormatTable {
				return mc.printMonitoringStatusTable(result)
			}

			return mc.outputResult(result, outputFormat)
		},
	}

	cmd.Flags().StringVar(&outputFormat, "output", monitoring.FormatTable, "Output format: table, json")
	cmd.Flags().BoolVar(&detailed, "detailed", false, "Show detailed metrics")

	return cmd
}

// buildMetricsCommand creates metrics command
func (mc *MonitoringCommands) buildMetricsCommand() *cobra.Command {
	var (
		outputFormat string
		metricsType  string
	)

	cmd := &cobra.Command{
		Use:   "metrics",
		Short: "Get system metrics",
		Long: `Get current system metrics including system, resource, application,
business, integration, and provider metrics.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			switch metricsType {
			case "all", "":
				metrics, err := mc.monitoringService.GetRealTimeMetrics(ctx)
				if err != nil {
					return fmt.Errorf("failed to get real-time metrics: %w", err)
				}
				return mc.outputResult(metrics, outputFormat)

			case "performance":
				metrics, err := mc.monitoringService.GetPerformanceMetrics(ctx)
				if err != nil {
					return fmt.Errorf("failed to get performance metrics: %w", err)
				}
				return mc.outputResult(metrics, outputFormat)

			default:
				return fmt.Errorf("unsupported metrics type: %s", metricsType)
			}
		},
	}

	cmd.Flags().StringVar(&outputFormat, "output", monitoring.FormatTable, "Output format: table, json")
	cmd.Flags().StringVar(&metricsType, "type", "all", "Metrics type: all, performance")

	return cmd
}

// buildHealthCommand creates health monitoring command
func (mc *MonitoringCommands) buildHealthCommand() *cobra.Command {
	var (
		outputFormat string
		component    string
		components   []string
	)

	cmd := &cobra.Command{
		Use:   "health",
		Short: "Check system health",
		Long: `Check system health status for all components or specific components.
Run comprehensive health checks and get detailed status information.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			if component != "" {
				// Get health for specific component
				health, err := mc.monitoringService.GetComponentHealth(ctx, component)
				if err != nil {
					return fmt.Errorf("failed to get component health: %w", err)
				}
				return mc.outputResult(health, outputFormat)
			}

			if len(components) > 0 {
				// Run health checks for specified components
				results, err := mc.monitoringService.RunHealthChecks(ctx, components)
				if err != nil {
					return fmt.Errorf("failed to run health checks: %w", err)
				}
				return mc.outputResult(results, outputFormat)
			}

			// Get overall system health
			health, err := mc.monitoringService.GetSystemHealth(ctx)
			if err != nil {
				return fmt.Errorf("failed to get system health: %w", err)
			}

			if outputFormat == monitoring.FormatTable {
				return mc.printHealthTable(health)
			}

			return mc.outputResult(health, outputFormat)
		},
	}

	cmd.Flags().StringVar(&outputFormat, "output", monitoring.FormatTable, "Output format: table, json")
	cmd.Flags().StringVar(&component, "component", "", "Specific component to check")
	cmd.Flags().StringSliceVar(&components, "components", []string{}, "List of components to check")

	return cmd
}

// buildPerformanceCommand creates performance monitoring command
func (mc *MonitoringCommands) buildPerformanceCommand() *cobra.Command {
	var (
		outputFormat string
		timeRange    string
		trends       bool
	)

	cmd := &cobra.Command{
		Use:   "performance",
		Short: "Get performance metrics and trends",
		Long: `Get current performance metrics and optionally analyze performance trends
over a specified time range.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			if trends {
				// Parse time range
				duration, err := time.ParseDuration(timeRange)
				if err != nil {
					return fmt.Errorf("invalid time range: %w", err)
				}

				trends, err := mc.monitoringService.GetPerformanceTrends(ctx, duration)
				if err != nil {
					return fmt.Errorf("failed to get performance trends: %w", err)
				}
				return mc.outputResult(trends, outputFormat)
			}

			// Get current performance metrics
			metrics, err := mc.monitoringService.GetPerformanceMetrics(ctx)
			if err != nil {
				return fmt.Errorf("failed to get performance metrics: %w", err)
			}

			if outputFormat == monitoring.FormatTable {
				return mc.printPerformanceTable(metrics)
			}

			return mc.outputResult(metrics, outputFormat)
		},
	}

	cmd.Flags().StringVar(&outputFormat, "output", monitoring.FormatTable, "Output format: table, json")
	cmd.Flags().StringVar(&timeRange, "time-range", "1h", "Time range for trends analysis")
	cmd.Flags().BoolVar(&trends, "trends", false, "Show performance trends")

	return cmd
}

// buildAlertsCommand creates alerts management command
func (mc *MonitoringCommands) buildAlertsCommand() *cobra.Command {
	var outputFormat string

	cmd := &cobra.Command{
		Use:   "alerts",
		Short: "Manage alerts",
		Long:  `Manage monitoring alerts including viewing active alerts and resolving them.`,
	}

	// List alerts subcommand
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List active alerts",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			alerts, err := mc.monitoringService.GetActiveAlerts(ctx)
			if err != nil {
				return fmt.Errorf("failed to get active alerts: %w", err)
			}

			if outputFormat == monitoring.FormatTable {
				return mc.printAlertsTable(alerts)
			}

			return mc.outputResult(alerts, outputFormat)
		},
	}

	// Resolve alert subcommand
	resolveCmd := &cobra.Command{
		Use:   "resolve [alert-id]",
		Short: "Resolve an alert",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			alertID := args[0]

			err := mc.monitoringService.ResolveAlert(ctx, alertID)
			if err != nil {
				return fmt.Errorf("failed to resolve alert: %w", err)
			}

			result := map[string]interface{}{
				"alert_id":    alertID,
				"status":      "resolved",
				"resolved_at": time.Now(),
			}

			return mc.outputResult(result, outputFormat)
		},
	}

	listCmd.Flags().StringVar(&outputFormat, "output", monitoring.FormatTable, "Output format: table, json")
	resolveCmd.Flags().StringVar(&outputFormat, "output", monitoring.FormatTable, "Output format: table, json")

	cmd.AddCommand(listCmd)
	cmd.AddCommand(resolveCmd)

	return cmd
}

// buildDashboardCommand creates dashboard command
func (mc *MonitoringCommands) buildDashboardCommand() *cobra.Command {
	var outputFormat string

	cmd := &cobra.Command{
		Use:   "dashboard",
		Short: "Get dashboard data",
		Long:  `Get comprehensive dashboard data including system overview, metrics, and alerts.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			dashboard, err := mc.monitoringService.GetDashboardData(ctx)
			if err != nil {
				return fmt.Errorf("failed to get dashboard data: %w", err)
			}

			if outputFormat == monitoring.FormatTable {
				return mc.printDashboardTable(dashboard)
			}

			return mc.outputResult(dashboard, outputFormat)
		},
	}

	cmd.Flags().StringVar(&outputFormat, "output", "table", "Output format: table, json")

	return cmd
}

// buildTrendsCommand creates trends analysis command
func (mc *MonitoringCommands) buildTrendsCommand() *cobra.Command {
	var (
		outputFormat string
		timeRange    string
	)

	cmd := &cobra.Command{
		Use:   "trends",
		Short: "Analyze performance trends",
		Long:  `Analyze performance trends over a specified time range with anomaly detection.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			// Parse time range
			duration, err := time.ParseDuration(timeRange)
			if err != nil {
				return fmt.Errorf("invalid time range: %w", err)
			}

			trends, err := mc.monitoringService.GetPerformanceTrends(ctx, duration)
			if err != nil {
				return fmt.Errorf("failed to get performance trends: %w", err)
			}

			if outputFormat == monitoring.FormatTable {
				return mc.printTrendsTable(trends)
			}

			return mc.outputResult(trends, outputFormat)
		},
	}

	cmd.Flags().StringVar(&outputFormat, "output", monitoring.FormatTable, "Output format: table, json")
	cmd.Flags().StringVar(&timeRange, "time-range", "1h", "Time range for analysis")

	return cmd
}

// buildConfigCommand creates configuration command
func (mc *MonitoringCommands) buildConfigCommand() *cobra.Command {
	var outputFormat string

	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage monitoring configuration",
		Long:  `View and manage monitoring configuration including thresholds and settings.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			config, err := mc.monitoringService.GetMonitoringConfig(ctx)
			if err != nil {
				return fmt.Errorf("failed to get monitoring config: %w", err)
			}

			if outputFormat == "table" {
				return mc.printConfigTable(config)
			}

			return mc.outputResult(config, outputFormat)
		},
	}

	cmd.Flags().StringVar(&outputFormat, "output", "table", "Output format: table, json")

	return cmd
}

// Output formatting methods

func (mc *MonitoringCommands) outputResult(result interface{}, format string) error {
	switch format {
	case "json":
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(result)
	case "table":
		// Default table output for generic results
		return mc.printGenericTable(result)
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}
}

func (mc *MonitoringCommands) printGenericTable(result interface{}) error {
	// Simple key-value table for generic results
	fmt.Println("═══════════════════════════════════════════════════")
	fmt.Println("                   RESULT")
	fmt.Println("═══════════════════════════════════════════════════")

	// Use JSON marshal/unmarshal to convert to map
	jsonData, err := json.Marshal(result)
	if err != nil {
		return err
	}

	var data map[string]interface{}
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return err
	}

	for key, value := range data {
		fmt.Printf("%-25s: %v\n", key, value)
	}

	return nil
}

func (mc *MonitoringCommands) printMonitoringStatusTable(result map[string]interface{}) error {
	fmt.Println("═══════════════════════════════════════════════════")
	fmt.Println("             MONITORING STATUS")
	fmt.Println("═══════════════════════════════════════════════════")

	fmt.Printf("Monitoring Enabled: %v\n", result["monitoring_enabled"])
	fmt.Printf("Metrics Interval: %v\n", result["metrics_interval"])
	fmt.Printf("Health Check Interval: %v\n", result["health_check_interval"])
	fmt.Printf("Alerting Enabled: %v\n", result["alerting_enabled"])
	fmt.Printf("Dashboard Port: %v\n", result["dashboard_port"])
	fmt.Printf("Active Alerts: %v\n", result["active_alerts_count"])

	if channels, ok := result["notification_channels"].([]string); ok {
		fmt.Printf("Notification Channels: %s\n", strings.Join(channels, ", "))
	}

	fmt.Printf("Status Timestamp: %v\n", result["status_timestamp"])

	return nil
}

func (mc *MonitoringCommands) printHealthTable(health *monitoring.SystemHealthStatus) error {
	fmt.Println("═══════════════════════════════════════════════════")
	fmt.Println("              SYSTEM HEALTH")
	fmt.Println("═══════════════════════════════════════════════════")

	fmt.Printf("Overall Health: %s\n", health.OverallHealth)
	fmt.Printf("Health Score: %d/100\n", health.HealthScore)
	fmt.Printf("Uptime: %.1f hours\n", health.UptimeHours)
	fmt.Printf("System Load: %.2f\n", health.SystemLoad)
	fmt.Printf("Memory Pressure: %.2f\n", health.MemoryPressure)
	fmt.Printf("Disk Pressure: %.2f\n", health.DiskPressure)

	fmt.Println("\n=== Component Health ===")
	for component, status := range health.ComponentHealth {
		statusSymbol := monitoring.SymbolHealthy
		if status != monitoring.HealthyStatus {
			statusSymbol = monitoring.SymbolUnhealthy
		}
		fmt.Printf("%s %-20s: %s\n", statusSymbol, component, status)
	}

	if len(health.CriticalComponents) > 0 {
		fmt.Printf("\nCritical Components: %s\n", strings.Join(health.CriticalComponents, ", "))
	}

	if len(health.DegradedComponents) > 0 {
		fmt.Printf("Degraded Components: %s\n", strings.Join(health.DegradedComponents, ", "))
	}

	if len(health.FailedComponents) > 0 {
		fmt.Printf("Failed Components: %s\n", strings.Join(health.FailedComponents, ", "))
	}

	return nil
}

func (mc *MonitoringCommands) printPerformanceTable(metrics *monitoring.PerformanceMetrics) error {
	fmt.Println("═══════════════════════════════════════════════════")
	fmt.Println("            PERFORMANCE METRICS")
	fmt.Println("═══════════════════════════════════════════════════")

	fmt.Printf("Performance Score: %d/100 (Grade: %s)\n", metrics.PerformanceScore, metrics.Grade)

	fmt.Println("\n=== CPU Metrics ===")
	fmt.Printf("Usage: %.1f%% (%d cores)\n", metrics.CPU.UsagePercent, metrics.CPU.Cores)
	fmt.Printf("Load Average: %.2f (1m), %.2f (5m), %.2f (15m)\n",
		metrics.CPU.LoadAverage1m, metrics.CPU.LoadAverage5m, metrics.CPU.LoadAverage15m)

	fmt.Println("\n=== Memory Metrics ===")
	fmt.Printf("Usage: %.1f%% (%.1f GB / %.1f GB)\n",
		metrics.Memory.UsagePercent, metrics.Memory.UsedGB, metrics.Memory.TotalGB)
	fmt.Printf("Free: %.1f GB, Cached: %.1f GB\n", metrics.Memory.FreeGB, metrics.Memory.CachedGB)

	fmt.Println("\n=== Disk Metrics ===")
	fmt.Printf("Usage: %.1f%% (%.1f GB / %.1f GB)\n",
		metrics.Disk.UsagePercent, metrics.Disk.UsedGB, metrics.Disk.TotalGB)
	fmt.Printf("I/O: %.1f read ops/s, %.1f write ops/s\n",
		metrics.Disk.ReadOpsPerSec, metrics.Disk.WriteOpsPerSec)

	fmt.Println("\n=== Network Metrics ===")
	fmt.Printf("Latency: %.1f ms\n", metrics.Network.NetworkLatency)
	fmt.Printf("Bandwidth: %.1f Mbps\n", metrics.Network.Bandwidth)
	fmt.Printf("Connections: %d\n", metrics.Network.Connections)

	fmt.Println("\n=== Application Metrics ===")
	fmt.Printf("Request Count: %d\n", metrics.Application.RequestCount)
	fmt.Printf("Success Rate: %.1f%%\n", metrics.Application.SuccessRate)
	fmt.Printf("Error Rate: %.1f%%\n", metrics.Application.ErrorRate)
	fmt.Printf("Throughput: %.1f RPS\n", metrics.Application.ThroughputRPS)
	fmt.Printf("Response Time: %.1f ms (avg), %.1f ms (p95)\n",
		metrics.Application.ResponseTime.Average, metrics.Application.ResponseTime.P95)

	return nil
}

func (mc *MonitoringCommands) printAlertsTable(alerts []*monitoring.Alert) error {
	fmt.Println("═══════════════════════════════════════════════════")
	fmt.Println("              ACTIVE ALERTS")
	fmt.Println("═══════════════════════════════════════════════════")

	if len(alerts) == 0 {
		fmt.Println("No active alerts")
		return nil
	}

	// Sort alerts by severity and creation time
	sort.Slice(alerts, func(i, j int) bool {
		severityOrder := map[string]int{monitoring.SeverityCritical: 3, monitoring.SeverityWarning: 2, "info": 1}
		if severityOrder[alerts[i].Severity] != severityOrder[alerts[j].Severity] {
			return severityOrder[alerts[i].Severity] > severityOrder[alerts[j].Severity]
		}
		return alerts[i].CreatedAt.After(alerts[j].CreatedAt)
	})

	for _, alert := range alerts {
		severitySymbol := "ℹ"
		switch alert.Severity {
		case monitoring.SeverityCritical:
			severitySymbol = ""
		case monitoring.SeverityWarning:
			severitySymbol = "🟡"
		}

		fmt.Printf("%s [%s] %s\n", severitySymbol, alert.Severity, alert.Title)
		fmt.Printf("   Component: %s | Source: %s\n", alert.Component, alert.Source)
		fmt.Printf("   Description: %s\n", alert.Description)
		fmt.Printf("   Created: %s | ID: %s\n",
			alert.CreatedAt.Format("2006-01-02 15:04:05"), alert.ID)

		if alert.AcknowledgedAt != nil {
			fmt.Printf("   Acknowledged: %s by %s\n",
				alert.AcknowledgedAt.Format("2006-01-02 15:04:05"), alert.AcknowledgedBy)
		}

		fmt.Println()
	}

	return nil
}

func (mc *MonitoringCommands) printDashboardTable(dashboard *monitoring.DashboardData) error {
	fmt.Println("═══════════════════════════════════════════════════")
	fmt.Println("              DASHBOARD OVERVIEW")
	fmt.Println("═══════════════════════════════════════════════════")

	fmt.Printf("Overall Health: %s (Score: %d/100)\n",
		dashboard.SystemOverview.OverallHealth, dashboard.SystemOverview.HealthScore)
	fmt.Printf("Active Components: %d/%d\n",
		dashboard.SystemOverview.ActiveComponents, dashboard.SystemOverview.TotalComponents)
	fmt.Printf("Critical Alerts: %d\n", dashboard.SystemOverview.CriticalAlerts)
	fmt.Printf("Active Users: %d\n", dashboard.SystemOverview.ActiveUsers)
	fmt.Printf("Uptime: %s\n", dashboard.SystemOverview.Uptime)
	fmt.Printf("Environment: %s | Version: %s\n",
		dashboard.SystemOverview.Environment, dashboard.SystemOverview.Version)

	fmt.Println("\n=== Real-Time Metrics ===")
	fmt.Printf("Health Score: %d/100\n", dashboard.RealTimeMetrics.HealthScore)
	fmt.Printf("Active Alerts: %d\n", dashboard.RealTimeMetrics.ActiveAlerts)
	fmt.Printf("Collection Time: %d ms\n", dashboard.RealTimeMetrics.CollectionTimeMs)

	fmt.Println("\n=== Health Indicators ===")
	for _, indicator := range dashboard.HealthIndicators {
		statusSymbol := monitoring.SymbolHealthy
		if indicator.Status != monitoring.HealthyStatus {
			statusSymbol = monitoring.SymbolUnhealthy
		}
		fmt.Printf("%s %-20s: %.1f %s (%s)\n",
			statusSymbol, indicator.Name, indicator.Value, indicator.Unit, indicator.Trend)
	}

	fmt.Println("\n=== KPI Metrics ===")
	for _, kpi := range dashboard.KPIMetrics {
		statusSymbol := monitoring.SymbolHealthy
		if kpi.Status != monitoring.HealthyStatus {
			statusSymbol = monitoring.SymbolUnhealthy
		}
		fmt.Printf("%s %-25s: %.2f %s (Target: %.2f, %s)\n",
			statusSymbol, kpi.Name, kpi.Value, kpi.Unit, kpi.Target, kpi.Trend)
	}

	return nil
}

func (mc *MonitoringCommands) printTrendsTable(trends *monitoring.PerformanceTrends) error {
	fmt.Println("═══════════════════════════════════════════════════")
	fmt.Println("            PERFORMANCE TRENDS")
	fmt.Println("═══════════════════════════════════════════════════")

	fmt.Printf("Time Range: %v (%s to %s)\n",
		trends.TimeRange,
		trends.StartTime.Format("2006-01-02 15:04:05"),
		trends.EndTime.Format("2006-01-02 15:04:05"))

	fmt.Println("\n=== Trend Analysis ===")
	trends_list := []monitoring.TrendData{
		trends.CPUTrend,
		trends.MemoryTrend,
		trends.LatencyTrend,
		trends.ThroughputTrend,
		trends.ErrorRateTrend,
	}

	for _, trend := range trends_list {
		trendSymbol := "→"
		switch trend.Trend {
		case "increasing":
			trendSymbol = "↗"
		case "decreasing":
			trendSymbol = "↘"
		case "stable":
			trendSymbol = "→"
		}

		fmt.Printf("%s %-20s: %s (%.1f%% change)\n",
			trendSymbol, trend.MetricName, trend.Trend, trend.ChangePercent)
	}

	fmt.Printf("\nSummary: %s\n", trends.PerformanceSummary)

	if len(trends.Recommendations) > 0 {
		fmt.Println("\n=== Recommendations ===")
		for i, rec := range trends.Recommendations {
			fmt.Printf("%d. %s\n", i+1, rec)
		}
	}

	if len(trends.Anomalies) > 0 {
		fmt.Println("\n=== Anomalies Detected ===")
		for _, anomaly := range trends.Anomalies {
			fmt.Printf(" %s: %s (%.1f deviation)\n",
				anomaly.MetricName, anomaly.Description, anomaly.Deviation)
			fmt.Printf("   Time: %s | Severity: %s\n",
				anomaly.Timestamp.Format("2006-01-02 15:04:05"), anomaly.Severity)
		}
	}

	return nil
}

func (mc *MonitoringCommands) printConfigTable(config *monitoring.MonitoringConfig) error {
	fmt.Println("═══════════════════════════════════════════════════")
	fmt.Println("           MONITORING CONFIGURATION")
	fmt.Println("═══════════════════════════════════════════════════")

	fmt.Printf("Real-time Monitoring: %v\n", config.EnableRealTime)
	fmt.Printf("Metrics Interval: %v\n", config.MetricsInterval)
	fmt.Printf("Health Check Interval: %v\n", config.HealthCheckInterval)
	fmt.Printf("Alerting Enabled: %v\n", config.AlertingEnabled)
	fmt.Printf("Dashboard Port: %d\n", config.DashboardPort)
	fmt.Printf("Retention Period: %v\n", config.RetentionPeriod)

	if len(config.NotificationChannels) > 0 {
		fmt.Printf("Notification Channels: %s\n", strings.Join(config.NotificationChannels, ", "))
	}

	fmt.Println("\n=== Performance Thresholds ===")
	fmt.Printf("CPU Warning: %.1f%% | Critical: %.1f%%\n",
		config.PerformanceThresholds.CPUWarning, config.PerformanceThresholds.CPUCritical)
	fmt.Printf("Memory Warning: %.1f%% | Critical: %.1f%%\n",
		config.PerformanceThresholds.MemoryWarning, config.PerformanceThresholds.MemoryCritical)
	fmt.Printf("Disk Warning: %.1f%% | Critical: %.1f%%\n",
		config.PerformanceThresholds.DiskWarning, config.PerformanceThresholds.DiskCritical)
	fmt.Printf("Latency Warning: %.1f ms | Critical: %.1f ms\n",
		config.PerformanceThresholds.LatencyWarning, config.PerformanceThresholds.LatencyCritical)
	fmt.Printf("Error Rate Warning: %.1f%% | Critical: %.1f%%\n",
		config.PerformanceThresholds.ErrorRateWarning, config.PerformanceThresholds.ErrorRateCritical)

	if len(config.AlertRules) > 0 {
		fmt.Printf("\nAlert Rules: %d configured\n", len(config.AlertRules))
	}

	if len(config.CustomMetrics) > 0 {
		fmt.Printf("Custom Metrics: %d configured\n", len(config.CustomMetrics))
	}

	return nil
}
