package monitoring

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/costscope/costscope/internal/core/integration"
	"github.com/costscope/costscope/internal/core/logging"
	"github.com/costscope/costscope/internal/core/production"
)

// BasicMonitoringService implements MonitoringService interface
type BasicMonitoringService struct {
	config              *MonitoringConfig
	logger              *logging.Logger
	productionService   production.ProductionService
	integrationService  integration.IntegrationService
	metricsCollector    MetricsCollector
	alertManager        AlertManager
	notificationService NotificationService
	metricEmitter       MetricEmitter
	alertEvaluator      AlertEvaluator
	healthChecker       HealthChecker

	mu                 sync.RWMutex
	isRunning          bool
	realTimeMetrics    *RealTimeMetrics
	healthStatus       *SystemHealthStatus
	performanceMetrics *PerformanceMetrics
	activeAlerts       []*Alert
	dashboardData      *DashboardData

	stopChan      chan struct{}
	metricsTicker *time.Ticker
	healthTicker  *time.Ticker
}

// NewBasicMonitoringService creates a new monitoring service
func NewBasicMonitoringService(
	logger *logging.Logger,
	productionService production.ProductionService,
	integrationService integration.IntegrationService,
) *BasicMonitoringService {
	config := &MonitoringConfig{
		EnableRealTime:       true,
		MetricsInterval:      30 * time.Second,
		AlertingEnabled:      true,
		NotificationChannels: []string{"email", "slack"},
		DashboardPort:        8081,
		RetentionPeriod:      24 * time.Hour,
		HealthCheckInterval:  1 * time.Minute,
		PerformanceThresholds: PerformanceThresholds{
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

	service := &BasicMonitoringService{
		config:             config,
		logger:             logger.WithFields(map[string]interface{}{"component": "monitoring", "subcomponent": "service"}),
		productionService:  productionService,
		integrationService: integrationService,
		stopChan:           make(chan struct{}),
		activeAlerts:       make([]*Alert, 0),
	}

	service.metricsCollector = NewBasicMetricsCollector(logger.WithFields(map[string]interface{}{"component": "monitoring", "subcomponent": "metrics"}), productionService, integrationService)
	service.alertManager = NewBasicAlertManager(logger.WithFields(map[string]interface{}{"component": "monitoring", "subcomponent": "alerts"}), service)
	service.notificationService = NewBasicNotificationService(logger.WithFields(map[string]interface{}{"component": "monitoring", "subcomponent": "notifier"}))
	service.metricEmitter = NewLoggingMetricEmitter(logger)
	service.alertEvaluator = NewDefaultAlertEvaluator(logger)
	service.healthChecker = NewDefaultHealthChecker(logger)

	return service
}

// StartRealTimeMonitoring starts real-time monitoring
func (bms *BasicMonitoringService) StartRealTimeMonitoring(ctx context.Context, config *MonitoringConfig) error {
	bms.mu.Lock()
	defer bms.mu.Unlock()

	if bms.isRunning {
		return fmt.Errorf("monitoring is already running")
	}

	bms.logger.Info("Starting real-time monitoring")

	if config != nil {
		bms.config = config
	}

	// Start metrics collection ticker
	bms.metricsTicker = time.NewTicker(bms.config.MetricsInterval)

	// Start health check ticker
	bms.healthTicker = time.NewTicker(bms.config.HealthCheckInterval)

	stopCh := bms.stopChan
	metricsTicker := bms.metricsTicker
	healthTicker := bms.healthTicker

	go bms.metricsCollectionLoop(ctx, stopCh, metricsTicker)
	go bms.healthCheckLoop(ctx, stopCh, healthTicker)
	go bms.alertProcessingLoop(ctx, stopCh)

	bms.isRunning = true
	bms.logger.Info("Real-time monitoring started successfully")

	return nil
}

// StopRealTimeMonitoring stops real-time monitoring
func (bms *BasicMonitoringService) StopRealTimeMonitoring(ctx context.Context) error {
	bms.mu.Lock()
	defer bms.mu.Unlock()

	if !bms.isRunning {
		return fmt.Errorf("monitoring is not running")
	}

	bms.logger.Info("Stopping real-time monitoring")

	if bms.metricsTicker != nil {
		bms.metricsTicker.Stop()
	}
	if bms.healthTicker != nil {
		bms.healthTicker.Stop()
	}

	close(bms.stopChan)
	bms.stopChan = make(chan struct{})

	bms.isRunning = false
	bms.logger.Info("Real-time monitoring stopped")

	return nil
}

// GetRealTimeMetrics returns current real-time metrics
func (bms *BasicMonitoringService) GetRealTimeMetrics(ctx context.Context) (*RealTimeMetrics, error) {
	bms.mu.RLock()
	defer bms.mu.RUnlock()

	if bms.realTimeMetrics == nil {
		return bms.collectRealTimeMetrics(ctx)
	}

	return bms.realTimeMetrics, nil
}

// GetSystemHealth returns current system health status
func (bms *BasicMonitoringService) GetSystemHealth(ctx context.Context) (*SystemHealthStatus, error) {
	if bms.healthChecker != nil {
		return bms.healthChecker.System(ctx), nil
	}
	return bms.collectSystemHealth(ctx), nil
}

// GetComponentHealth returns health status for specific component
func (bms *BasicMonitoringService) GetComponentHealth(ctx context.Context, component string) (*ComponentHealth, error) {
	if bms.healthChecker != nil {
		return bms.healthChecker.Component(ctx, component)
	}
	return nil, fmt.Errorf("health checker not available")
}

// RunHealthChecks runs comprehensive health checks
func (bms *BasicMonitoringService) RunHealthChecks(ctx context.Context, components []string) (*HealthCheckResults, error) {
	if bms.healthChecker != nil {
		return bms.healthChecker.Run(ctx, components)
	}
	return nil, fmt.Errorf("health checker not available")
}

// GetPerformanceMetrics returns current performance metrics
func (bms *BasicMonitoringService) GetPerformanceMetrics(ctx context.Context) (*PerformanceMetrics, error) {
	bms.mu.RLock()
	defer bms.mu.RUnlock()

	if bms.performanceMetrics == nil {
		return bms.collectPerformanceMetrics(ctx)
	}

	return bms.performanceMetrics, nil
}

// GetPerformanceTrends returns performance trend analysis
func (bms *BasicMonitoringService) GetPerformanceTrends(ctx context.Context, timeRange time.Duration) (*PerformanceTrends, error) {
	bms.logger.Info(fmt.Sprintf("Analyzing performance trends for %v", timeRange))

	endTime := time.Now()
	startTime := endTime.Add(-timeRange)

	// Generate sample trend data
	trends := &PerformanceTrends{
		TimeRange: timeRange,
		StartTime: startTime,
		EndTime:   endTime,
		CPUTrend: TrendData{
			MetricName:    "CPU Usage",
			Trend:         "increasing",
			ChangePercent: 12.5,
			Slope:         0.8,
		},
		MemoryTrend: TrendData{
			MetricName:    "Memory Usage",
			Trend:         "stable",
			ChangePercent: -2.1,
			Slope:         -0.1,
		},
		LatencyTrend: TrendData{
			MetricName:    "Response Latency",
			Trend:         "decreasing",
			ChangePercent: -8.3,
			Slope:         -1.2,
		},
		PerformanceSummary: "Overall performance is stable with slight CPU increase",
		Recommendations: []string{
			"Monitor CPU utilization closely",
			"Consider scaling if trend continues",
			"Optimize resource-intensive operations",
		},
		Anomalies: []Anomaly{
			{
				Timestamp:     time.Now().Add(-2 * time.Hour),
				MetricName:    "CPU Usage",
				Value:         95.2,
				ExpectedValue: 45.0,
				Deviation:     50.2,
				Severity:      "high",
				Description:   "Unexpected CPU spike detected",
			},
		},
	}

	// Generate historical data points
	trends.CPUTrend.DataPoints = bms.generateTrendDataPoints(startTime, endTime, 45.0, 15.0)
	trends.MemoryTrend.DataPoints = bms.generateTrendDataPoints(startTime, endTime, 65.0, 10.0)
	trends.LatencyTrend.DataPoints = bms.generateTrendDataPoints(startTime, endTime, 75.0, 25.0)

	return trends, nil
}

// CreateAlert creates a new alert
func (bms *BasicMonitoringService) CreateAlert(ctx context.Context, alert *AlertDefinition) error {
	return bms.alertManager.CreateAlertRule(ctx, &AlertRule{
		ID:                   fmt.Sprintf("alert_%d", time.Now().Unix()),
		Name:                 alert.Name,
		Description:          alert.Description,
		Metric:               alert.MetricName,
		Operator:             alert.Condition,
		Threshold:            alert.Threshold,
		Severity:             alert.Severity,
		NotificationChannels: alert.NotificationChannels,
		Tags:                 alert.Tags,
		Enabled:              alert.Enabled,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	})
}

// GetActiveAlerts returns all active alerts
func (bms *BasicMonitoringService) GetActiveAlerts(ctx context.Context) ([]*Alert, error) {
	bms.mu.RLock()
	defer bms.mu.RUnlock()

	return bms.activeAlerts, nil
}

// ResolveAlert resolves an active alert
func (bms *BasicMonitoringService) ResolveAlert(ctx context.Context, alertID string) error {
	bms.mu.Lock()
	defer bms.mu.Unlock()

	for i, alert := range bms.activeAlerts {
		if alert.ID == alertID {
			now := time.Now()
			alert.Status = "resolved"
			alert.ResolvedAt = &now
			alert.UpdatedAt = now

			// Remove from active alerts
			bms.activeAlerts = append(bms.activeAlerts[:i], bms.activeAlerts[i+1:]...)

			bms.logger.Info(fmt.Sprintf("Alert %s resolved", alertID))
			return nil
		}
	}

	return fmt.Errorf("alert %s not found", alertID)
}

// GetDashboardData returns dashboard data
func (bms *BasicMonitoringService) GetDashboardData(ctx context.Context) (*DashboardData, error) {
	bms.mu.RLock()
	defer bms.mu.RUnlock()

	if bms.dashboardData == nil {
		return bms.generateDashboardData(ctx)
	}

	return bms.dashboardData, nil
}

// GetMonitoringConfig returns current monitoring configuration
func (bms *BasicMonitoringService) GetMonitoringConfig(ctx context.Context) (*MonitoringConfig, error) {
	bms.mu.RLock()
	defer bms.mu.RUnlock()

	return bms.config, nil
}

// UpdateMonitoringConfig updates monitoring configuration
func (bms *BasicMonitoringService) UpdateMonitoringConfig(ctx context.Context, config *MonitoringConfig) error {
	bms.mu.Lock()
	defer bms.mu.Unlock()

	bms.config = config
	bms.logger.Info("Monitoring configuration updated")

	if bms.isRunning {
		bms.logger.Info("Restarting monitoring with new configuration")
	}

	return nil
}

func (bms *BasicMonitoringService) metricsCollectionLoop(ctx context.Context, stopCh <-chan struct{}, metricsTicker *time.Ticker) {
	bms.logger.Info("Starting metrics collection loop")

	for {
		select {
		case <-ctx.Done():
			bms.logger.Info("Metrics collection loop stopped (context done)")
			return
		case <-stopCh:
			bms.logger.Info("Metrics collection loop stopped (stop signal)")
			return
		case <-metricsTicker.C:
			bms.collectAndUpdateMetrics(ctx)
		}
	}
}

func (bms *BasicMonitoringService) healthCheckLoop(ctx context.Context, stopCh <-chan struct{}, healthTicker *time.Ticker) {
	bms.logger.Info("Starting health check loop")

	for {
		select {
		case <-ctx.Done():
			bms.logger.Info("Health check loop stopped (context done)")
			return
		case <-stopCh:
			bms.logger.Info("Health check loop stopped (stop signal)")
			return
		case <-healthTicker.C:
			bms.updateSystemHealth(ctx)
		}
	}
}

func (bms *BasicMonitoringService) alertProcessingLoop(ctx context.Context, stopCh <-chan struct{}) {
	bms.logger.Info("Starting alert processing loop")

	alertTicker := time.NewTicker(10 * time.Second) // Check for alerts every 10 seconds
	defer alertTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			bms.logger.Info("Alert processing loop stopped (context done)")
			return
		case <-stopCh:
			bms.logger.Info("Alert processing loop stopped (stop signal)")
			return
		case <-alertTicker.C:
			bms.processAlerts(ctx)
		}
	}
}

func (bms *BasicMonitoringService) collectAndUpdateMetrics(ctx context.Context) {
	metrics, err := bms.collectRealTimeMetrics(ctx)
	if err != nil {
		bms.logger.Error(fmt.Sprintf("Failed to collect real-time metrics: %v", err))
		return
	}

	bms.mu.Lock()
	bms.realTimeMetrics = metrics
	bms.mu.Unlock()

	bms.logger.Debug("Real-time metrics updated")

	// Emit metrics via emitter (non-blocking best-effort)
	if bms.metricEmitter != nil {
		_ = bms.metricEmitter.Emit(ctx, metrics)
	}
}

func (bms *BasicMonitoringService) updateSystemHealth(ctx context.Context) {
	var health *SystemHealthStatus
	if bms.healthChecker != nil {
		health = bms.healthChecker.System(ctx)
	} else {
		health = bms.collectSystemHealth(ctx)
	}
	if health == nil {
		bms.logger.Error("Failed to collect system health")
		return
	}

	bms.mu.Lock()
	bms.healthStatus = health
	bms.mu.Unlock()

	bms.logger.Debug("System health updated")
}

func (bms *BasicMonitoringService) processAlerts(ctx context.Context) {
	select {
	case <-ctx.Done():
		bms.logger.Warn("Alert processing cancelled")
		return
	default:
	}

	if bms.realTimeMetrics == nil {
		return
	}
	if bms.alertEvaluator != nil {
		candidates := bms.alertEvaluator.Evaluate(ctx, bms.realTimeMetrics, bms.config.PerformanceThresholds)
		for _, a := range candidates {
			// ensure no duplicates
			dup := false
			for _, ex := range bms.activeAlerts {
				if ex.Type == a.Type && ex.Status == AlertStatusActive {
					dup = true
					break
				}
			}
			if !dup {
				bms.mu.Lock()
				bms.activeAlerts = append(bms.activeAlerts, a)
				bms.mu.Unlock()
				bms.logger.Warn(fmt.Sprintf("Alert triggered: %s - %s", a.Type, a.Description))
				if bms.config.AlertingEnabled {
					n := &Notification{ID: fmt.Sprintf("notif_%d", time.Now().Unix()), Type: "alert", Channel: "default", Subject: a.Title, Message: a.Description, Severity: a.Severity, CreatedAt: time.Now(), Status: "pending"}
					if err := bms.notificationService.SendNotification(context.Background(), n); err != nil {
						bms.logger.Error(fmt.Sprintf("Failed to send notification: %v", err))
					}
				}
			}
		}
	}
}

func (bms *BasicMonitoringService) collectRealTimeMetrics(ctx context.Context) (*RealTimeMetrics, error) {
	startTime := time.Now()

	systemMetrics, err := bms.metricsCollector.CollectSystemMetrics(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to collect system metrics: %w", err)
	}

	resourceMetrics, err := bms.metricsCollector.CollectResourceMetrics(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to collect resource metrics: %w", err)
	}

	appMetrics, err := bms.metricsCollector.CollectApplicationMetrics(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to collect application metrics: %w", err)
	}

	businessMetrics, err := bms.metricsCollector.CollectBusinessMetrics(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to collect business metrics: %w", err)
	}

	integrationMetrics, err := bms.metricsCollector.CollectIntegrationMetrics(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to collect integration metrics: %w", err)
	}

	providerMetrics, err := bms.metricsCollector.CollectProviderMetrics(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to collect provider metrics: %w", err)
	}

	performance := &PerformanceMetrics{
		CPU:     resourceMetrics.CPU,
		Memory:  resourceMetrics.Memory,
		Disk:    resourceMetrics.Disk,
		Network: resourceMetrics.Network,
		Application: AppPerformanceMetrics{
			RequestCount:   int64(appMetrics.RequestsPerSecond * 60), // Convert to per minute
			SuccessRate:    100.0 - appMetrics.ErrorRate,
			ErrorRate:      appMetrics.ErrorRate,
			ResponseTime:   appMetrics.ResponseTime,
			ThroughputRPS:  appMetrics.RequestsPerSecond,
			MemoryUsage:    resourceMetrics.Memory.UsagePercent,
			GoroutineCount: 150,
			GCPauseTime:    2.5,
		},
		PerformanceScore: bms.calculatePerformanceScore(resourceMetrics),
		Grade:            bms.calculatePerformanceGrade(resourceMetrics),
	}

	metrics := &RealTimeMetrics{
		Timestamp:        time.Now(),
		System:           *systemMetrics,
		Performance:      *performance,
		Resources:        *resourceMetrics,
		Applications:     *appMetrics,
		Business:         *businessMetrics,
		Integrations:     *integrationMetrics,
		Providers:        *providerMetrics,
		ActiveAlerts:     len(bms.activeAlerts),
		HealthScore:      bms.calculateOverallHealthScore(resourceMetrics, appMetrics),
		TrendIndicators:  bms.generateTrendIndicators(),
		CollectionTimeMs: time.Since(startTime).Milliseconds(),
	}

	return metrics, nil
}
