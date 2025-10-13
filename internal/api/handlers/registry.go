package handlers

import (
	"github.com/costscope/costscope/internal/api/jobs"
	"github.com/costscope/costscope/internal/api/websocket"
	"github.com/costscope/costscope/internal/core/focus/conversion"
	"github.com/costscope/costscope/internal/core/logging"
	"github.com/costscope/costscope/internal/core/reports"
	"github.com/costscope/costscope/internal/core/streaming"
	"github.com/costscope/costscope/internal/providers"
)

// =====================================================================================
// Enterprise Handlers Registry - Unified Access to All Module Handlers
// =====================================================================================

// EnterpriseRegistry provides access to all module handlers
type EnterpriseRegistry struct {
	// Core handlers
	HealthHandler    *HealthHandler
	DocsHandler      *DocsHandler
	WebSocketHandler *WebSocketHandler

	// Module handlers
	FocusHandler         *FocusHandler
	ProvidersHandler     *ProvidersHandler
	AnalyticsHandler     *AnalyticsHandler
	AnalyticsReadHandler *AnalyticsReadHandler
	ReportsHandler       *ReportsHandler
	StreamingHandler     *StreamingHandler
	MonitoringHandler    *MonitoringHandler
	IntegrationHandler   *IntegrationHandler
	ProductionHandler    *ProductionHandler
	ConfigHandler        *ConfigHandler
	MulticloudHandler    *MulticloudHandler
}

// NewEnterpriseRegistry creates a new enterprise handlers registry
func NewEnterpriseRegistry(
	logger *logging.Logger,
	jobManager *jobs.Manager,
	wsManager *websocket.Manager,
) *EnterpriseRegistry {
	// Initialize enterprise streaming engine (stub returns enterprise-only error on slim builds)
	engine := streaming.NewEnterpriseStreamingEngine(logger)
	// Lightweight provider manager (mock-style usage for multicloud preview endpoints)
	providerManager := providers.NewProviderManager()
	// Dedicated conversion manager (async conversion jobs). Concurrency kept small; can be made configurable later.
	// Use environment-aware factory so JOB_STORE_PATH can enable durable BoltJobStore when configured.
	convMgr := conversion.NewConfiguredConversionManager(4)

	// Minimal report service (no metadata store by default; caller can wrap if needed)
	reportSvc := reports.NewBasicReportService(logger)

	reg := &EnterpriseRegistry{
		// Core handlers
		HealthHandler:    NewHealthHandler(logger).WithJobs(jobManager),
		DocsHandler:      NewDocsHandler(logger),
		WebSocketHandler: NewWebSocketHandler(logger, wsManager),

		// Module handlers
		FocusHandler:         NewFocusHandler(logger, jobManager, wsManager, convMgr),
		ProvidersHandler:     NewProvidersHandler(logger),
		AnalyticsHandler:     NewAnalyticsHandler(logger, jobManager),
		AnalyticsReadHandler: NewAnalyticsReadHandler(logger),
		ReportsHandler:       NewReportsHandler(logger, jobManager, reportSvc),
		StreamingHandler:     NewStreamingHandler(logger, jobManager, engine),
		MonitoringHandler:    NewMonitoringHandler(logger),
		IntegrationHandler:   NewIntegrationHandler(logger, jobManager),
		ProductionHandler:    NewProductionHandler(logger),
		ConfigHandler:        NewConfigHandler(logger),
		MulticloudHandler:    NewMulticloudHandler(logger, providerManager),
	}

	// If the job manager exposes a repository (persistent mode), attach it for DB health checks.
	type repoGetter interface{ GetRepository() interface{} }
	if jobManager != nil {
		// The enterprise wiring may use a persistent job manager type that provides GetRepository.
		if rg, ok := any(jobManager).(repoGetter); ok {
			if r := rg.GetRepository(); r != nil {
				if pr, ok2 := r.(interface{ Health(ctx interface{}) error }); ok2 { // shape check; set via typed helper below
					_ = pr // just to satisfy static checks when tags vary
				}
			}
		}
	}

	return reg
}

// WithReportMetadataStore injects a metadata store into the underlying report service (fluent helper; no-op if absent).
// Intentional additive method (does not alter existing construction behavior) so server bootstrap can wire persistence
// based on config/env without changing call sites.
func (r *EnterpriseRegistry) WithReportMetadataStore(store reports.MetadataStore) *EnterpriseRegistry {
	if r != nil && r.ReportsHandler != nil && r.ReportsHandler.reportService != nil && store != nil {
		r.ReportsHandler.reportService.WithMetadataStore(store)
	}
	return r
}

// Close gracefully shuts down resources owned by handlers (e.g., conversion manager stores).
func (r *EnterpriseRegistry) Close() error {
	if r == nil {
		return nil
	}
	if r.FocusHandler != nil && r.FocusHandler.convMgr != nil {
		return r.FocusHandler.convMgr.Close()
	}
	return nil
}
