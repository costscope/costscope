package framework

import (
	"context"
	"fmt"

	"local/costscope/internal/core/logging"
)

// Plugin version constant
const DefaultPluginVersion = "1.0.0"

// AnalyticsPlugin provides analytics functionality
type AnalyticsPlugin struct {
	container *Container
	running   bool
}

// Name returns the plugin name
func (p *AnalyticsPlugin) Name() string {
	return "analytics"
}

// Version returns the plugin version
func (p *AnalyticsPlugin) Version() string {
	return DefaultPluginVersion
}

// Initialize initializes the analytics plugin
func (p *AnalyticsPlugin) Initialize(ctx context.Context, container *Container) error {
	p.container = container

	// Register analytics services
	container.RegisterSingleton("analytics.processor", NewAnalyticsProcessor())
	container.RegisterSingleton("analytics.aggregator", NewAnalyticsAggregator())

	return nil
}

// Start starts the analytics plugin
func (p *AnalyticsPlugin) Start(ctx context.Context) error {
	p.running = true
	return nil
}

// Stop stops the analytics plugin
func (p *AnalyticsPlugin) Stop(ctx context.Context) error {
	p.running = false
	return nil
}

// Health returns the health status
func (p *AnalyticsPlugin) Health() string {
	return pluginHealth(p.running)
}

// ReportingPlugin provides reporting functionality
type ReportingPlugin struct {
	container *Container
	running   bool
}

// Name returns the plugin name
func (p *ReportingPlugin) Name() string {
	return "reporting"
}

// Version returns the plugin version
func (p *ReportingPlugin) Version() string {
	return DefaultPluginVersion
}

// Initialize initializes the reporting plugin
func (p *ReportingPlugin) Initialize(ctx context.Context, container *Container) error {
	p.container = container

	// Register reporting services
	container.RegisterSingleton("reporting.generator", NewReportGenerator())
	container.RegisterSingleton("reporting.formatter", NewReportFormatter())

	return nil
}

// Start starts the reporting plugin
func (p *ReportingPlugin) Start(ctx context.Context) error {
	p.running = true
	return nil
}

// Stop stops the reporting plugin
func (p *ReportingPlugin) Stop(ctx context.Context) error {
	p.running = false
	return nil
}

// Health returns the health status
func (p *ReportingPlugin) Health() string {
	return pluginHealth(p.running)
}

// MonitoringPlugin provides monitoring functionality
type MonitoringPlugin struct {
	container *Container
	running   bool
}

// Name returns the plugin name
func (p *MonitoringPlugin) Name() string {
	return "monitoring"
}

// Version returns the plugin version
func (p *MonitoringPlugin) Version() string {
	return DefaultPluginVersion
}

// Initialize initializes the monitoring plugin
func (p *MonitoringPlugin) Initialize(ctx context.Context, container *Container) error {
	p.container = container

	// Register monitoring services
	container.RegisterSingleton("monitoring.collector", NewMetricsCollector())
	container.RegisterSingleton("monitoring.alerter", NewAlerter())

	return nil
}

// Start starts the monitoring plugin
func (p *MonitoringPlugin) Start(ctx context.Context) error {
	p.running = true
	return nil
}

// Stop stops the monitoring plugin
func (p *MonitoringPlugin) Stop(ctx context.Context) error {
	p.running = false
	return nil
}

// Health returns the health status
func (p *MonitoringPlugin) Health() string {
	return pluginHealth(p.running)
}

// Placeholder service implementations
type AnalyticsProcessor struct{}
type AnalyticsAggregator struct{}
type ReportGenerator struct{}
type ReportFormatter struct{}
type MetricsCollector struct{}
type Alerter struct{}

func NewAnalyticsProcessor() *AnalyticsProcessor   { return &AnalyticsProcessor{} }
func NewAnalyticsAggregator() *AnalyticsAggregator { return &AnalyticsAggregator{} }
func NewReportGenerator() *ReportGenerator         { return &ReportGenerator{} }
func NewReportFormatter() *ReportFormatter         { return &ReportFormatter{} }
func NewMetricsCollector() *MetricsCollector       { return &MetricsCollector{} }
func NewAlerter() *Alerter                         { return &Alerter{} }

// Framework configuration and logging
type FrameworkConfig struct {
	Debug        bool   `json:"debug"`
	LogLevel     string `json:"log_level"`
	PluginDir    string `json:"plugin_dir"`
	EventWorkers int    `json:"event_workers"`
}

func NewFrameworkConfig() *FrameworkConfig {
	return &FrameworkConfig{
		Debug:        false,
		LogLevel:     "info",
		PluginDir:    "./plugins",
		EventWorkers: 4,
	}
}

type FrameworkLogger struct {
	level  string
	logger *logging.Logger
}

func NewFrameworkLogger() *FrameworkLogger {
	return &FrameworkLogger{
		level:  "info",
		logger: logging.GetLogger().WithFields(map[string]interface{}{"component": "framework"}),
	}
}

func (l *FrameworkLogger) Info(msg string, args ...interface{}) {
	if len(args) > 0 {
		l.logger.Info(fmt.Sprintf(msg, args...))
		return
	}
	l.logger.Info(msg)
}

func (l *FrameworkLogger) Debug(msg string, args ...interface{}) {
	if len(args) > 0 {
		l.logger.Debug(fmt.Sprintf(msg, args...))
		return
	}
	l.logger.Debug(msg)
}

func (l *FrameworkLogger) Error(msg string, args ...interface{}) {
	if len(args) > 0 {
		l.logger.Error(fmt.Sprintf(msg, args...))
		return
	}
	l.logger.Error(msg)
}

func (l *FrameworkLogger) Warn(msg string, args ...interface{}) {
	if len(args) > 0 {
		l.logger.Warn(fmt.Sprintf(msg, args...))
		return
	}
	l.logger.Warn(msg)
}

// pluginHealth centralizes the simple running->health mapping used by plugins.
func pluginHealth(running bool) string {
	if running {
		return HealthStatusHealthy
	}
	return HealthStatusStopped
}
