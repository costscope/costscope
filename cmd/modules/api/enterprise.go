package api

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"

	"github.com/costscope/costscope/internal/api/handlers"
	"github.com/costscope/costscope/internal/api/jobs"
	"github.com/costscope/costscope/internal/api/middleware"
	"github.com/costscope/costscope/internal/api/websocket"
	"github.com/costscope/costscope/internal/core/config"
	"github.com/costscope/costscope/internal/core/logging"
	"github.com/costscope/costscope/internal/core/monitoring/telemetry"
	"github.com/costscope/costscope/internal/core/multitenant"
	"github.com/costscope/costscope/internal/core/normalization"
	"github.com/costscope/costscope/internal/core/persistence"
	"github.com/costscope/costscope/internal/core/reports"
	"github.com/costscope/costscope/internal/core/security"
)

// buildTenantMiddleware returns the tenant enforcement middleware; extracting logic reduces
// cyclomatic complexity of runEnterpriseAPIServer.
func buildTenantMiddleware(cfg *config.ConsolidatedConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		if cfg == nil || !multitenant.IsEnabled(cfg) {
			c.Next()
			return
		}
		// Prefer JWT claim
		tenantClaim, claimExists := c.Get("tenant_id")
		var tenant string
		if claimExists {
			if s, ok := tenantClaim.(string); ok {
				tenant = s
			}
		}
		if tenant == "" {
			headTen := c.GetHeader("X-Tenant-ID")
			if headTen != "" {
				tenant = headTen
			}
		}
		if tenant == "" { // require for non-admin roles
			rolesVal, _ := c.Get("roles")
			roles, _ := rolesVal.([]string)
			isAdmin := false
			for _, r := range roles {
				if r == "admin" || r == "system_admin" {
					isAdmin = true
					break
				}
			}
			if !isAdmin {
				c.JSON(http.StatusBadRequest, gin.H{"error": "tenant required (X-Tenant-ID) in multi-tenant mode"})
				c.Abort()
				return
			}
		}
		if tenant != "" {
			c.Set("tenant_id", tenant)
			// Also propagate into request context so downstream code using multitenant.EffectiveTenant
			// with a ContextResolver can resolve the tenant deterministically.
			ctx := multitenant.WithTenantToContext(c.Request.Context(), tenant)
			c.Request = c.Request.WithContext(ctx)
		}
		c.Next()
	}
}

// registerDebugRoutes conditionally registers debug endpoints (e.g., cache stats) with auth & role gating.
func registerDebugRoutes(router *gin.Engine, reg *handlers.EnterpriseRegistry, auth gin.HandlerFunc) {
	if !enableCacheStats {
		return
	}
	debugGroup := router.Group("/debug")
	// Require authentication and admin role for debug endpoints
	debugGroup.Use(auth)
	debugGroup.Use(middleware.RBAC("admin", "system_admin"))
	debugGroup.GET("/cache-stats", func(c *gin.Context) {
		reg.HealthHandler.CacheStats(c)
	})
}

// =====================================================================================
// Enterprise API Server (Gin) - Table-driven routing with middleware registry
// =====================================================================================

// Server settings
var (
	enterpriseHost string
	enterprisePort int

	// Security settings
	enterpriseJwtSecret string
	enterpriseJwtIssuer string
	enterpriseApiKeys   []string

	// CORS settings
	enterpriseCorsOrigins []string

	// Rate limiting
	enterpriseRateLimitRequests int
	enterpriseRateLimitWindow   time.Duration

	// Job processing
	enterpriseJobWorkers int

	// TLS settings
	enterpriseTlsEnabled bool
	enterpriseTlsCert    string
	enterpriseTlsKey     string
	// TLS advanced settings
	tlsMinVersion          string
	tlsCipherSuites        []string
	tlsPreferServerCiphers bool

	// Health check settings
	healthPath string

	// Documentation
	docsEnabled bool
	docsPath    string

	// Cache / normalization observability
	cacheMetricsRefreshInterval time.Duration
	enableCacheStats            bool

	// Casbin RBAC (optional)
	enterpriseCasbinEnabled       bool
	enterpriseCasbinModelPath     string
	enterpriseCasbinPolicyPath    string
	enterpriseCasbinDefaultDomain string
)

// BuildEnterpriseAPICommand creates the complete enterprise API server command
func BuildEnterpriseAPICommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "enterprise",
		Short: "Start the complete enterprise API server",
		Long: `Start the complete enterprise API server with all modules integrated.

This command starts a comprehensive REST API server that provides access to all
CostScope modules through a unified enterprise-grade interface.

Endpoints and behavior are stable; authentication is enforced on /api/v1/* while /health, /metrics and /docs are public.`,
		RunE: runEnterpriseAPIServer,
	}

	// Flags
	cmd.Flags().StringVar(&enterpriseHost, "host", "0.0.0.0", "Server host")
	cmd.Flags().IntVar(&enterprisePort, "port", 8080, "Server port")

	cmd.Flags().StringVar(&enterpriseJwtSecret, "jwt-secret", os.Getenv("COSTSCOPE_JWT_SECRET"), "JWT secret")
	cmd.Flags().StringVar(&enterpriseJwtIssuer, "jwt-issuer", "costscope", "JWT issuer")
	cmd.Flags().StringSliceVar(&enterpriseApiKeys, "api-keys", []string{}, "Allowed API keys")

	cmd.Flags().StringSliceVar(&enterpriseCorsOrigins, "cors-origins", []string{"*"}, "CORS allowed origins")

	cmd.Flags().IntVar(&enterpriseRateLimitRequests, "rate-limit-requests", 1000, "Rate limit: requests per window")
	cmd.Flags().DurationVar(&enterpriseRateLimitWindow, "rate-limit-window", time.Minute, "Rate limit window")

	cmd.Flags().IntVar(&enterpriseJobWorkers, "job-workers", 4, "Number of job workers")

	cmd.Flags().BoolVar(&enterpriseTlsEnabled, "tls-enabled", false, "Enable TLS")
	cmd.Flags().StringVar(&enterpriseTlsCert, "tls-cert", "server.crt", "TLS certificate path")
	cmd.Flags().StringVar(&enterpriseTlsKey, "tls-key", "server.key", "TLS key path")
	// TLS hardening (defaults: TLS 1.2 minimum; strong TLS 1.2 ciphers; TLS 1.3 preferred when available)
	cmd.Flags().StringVar(&tlsMinVersion, "tls-min-version", "1.2", "Minimum TLS version to allow (1.2 or 1.3)")
	cmd.Flags().StringSliceVar(&tlsCipherSuites, "tls-cipher-suites", []string{}, "Allowed TLS 1.2 cipher suites (names, empty = secure defaults). Ignored for TLS 1.3")
	cmd.Flags().BoolVar(&tlsPreferServerCiphers, "tls-prefer-server-ciphers", true, "Prefer server cipher suites order")

	cmd.Flags().StringVar(&healthPath, "health-path", "/health", "Health endpoint path")
	cmd.Flags().BoolVar(&docsEnabled, "docs", true, "Enable docs endpoint")
	cmd.Flags().StringVar(&docsPath, "docs-path", "/docs", "Docs endpoint base path")

	// Normalization cache metrics / debug
	cmd.Flags().DurationVar(&cacheMetricsRefreshInterval, "cache-metrics-refresh-interval", 0, "Interval for background cache metrics gauge refresh (0=disabled)")
	cmd.Flags().BoolVar(&enableCacheStats, "enable-cache-stats", false, "Enable authenticated /debug/cache-stats endpoint (admin role required)")

	cmd.Flags().BoolVar(&enterpriseCasbinEnabled, "casbin", false, "Enable Casbin RBAC")
	cmd.Flags().StringVar(&enterpriseCasbinModelPath, "casbin-model", "", "Casbin model path (optional)")
	cmd.Flags().StringVar(&enterpriseCasbinPolicyPath, "casbin-policy", "", "Casbin policy path (optional)")
	cmd.Flags().StringVar(&enterpriseCasbinDefaultDomain, "casbin-domain", "", "Casbin default domain (optional)")

	return cmd
}

func runEnterpriseAPIServer(_ *cobra.Command, _ []string) error {
	logger := logging.NewLogger(logging.LevelInfo)
	logger.Info("Starting Enterprise API server...")

	// Metrics registry (idempotent within process if invoked once)
	// Ensures all CostScope metrics are registered before /metrics is scraped.
	// Note: Register may panic on double registration; server runs once per process.
	telemetry.Register()

	// Tracing (optional)
	shutdownTrace, err := telemetry.InitTracing(context.Background())
	if err != nil {
		logger.WarnWithFields("Tracing init failed", map[string]interface{}{"error": err.Error()})
	}
	defer func() {
		if shutdownTrace != nil {
			_ = shutdownTrace(context.Background())
		}
	}()

	// Config (optional): use centralized resolver to load COSTSCOPE_CONFIG or ~/.costscope/config.yaml
	loadedCfg := config.LoadOptionalYAML(logger)

	// Enforce JWT secret policy (fail fast)
	if err := validateEnterpriseJWTSecret(enterpriseJwtSecret); err != nil {
		logger.FatalWithFields("Invalid JWT secret", map[string]interface{}{
			"error":  err.Error(),
			"length": len(enterpriseJwtSecret),
		})
	}

	// Managers
	jobManager := jobs.NewManager(logger, enterpriseJobWorkers)
	// Configure job quotas if multi-tenant enabled
	if loadedCfg != nil && multitenant.IsEnabled(loadedCfg) {
		jobManager.ConfigureQuotas(loadedCfg.MultiTenant.MaxJobsPerTenant, loadedCfg.MultiTenant.MaxActiveJobsPerTenant)
	}
	if err := jobManager.Start(); err != nil {
		return fmt.Errorf("failed to start job manager: %w", err)
	}
	defer func() { _ = jobManager.Stop() }()
	wsManager := websocket.NewManager(logger)

	// Optional: initialize a repository for readiness DB checks (stubbed when built without 'sqlite').
	// Using a local file path aligns with default persistence expectations; stub builds ignore the file.
	var repo persistence.Repository
	{
		cfg := persistence.DefaultDatabaseConfig()
		// Prefer a dedicated data dir; if it doesn't exist and build uses stub, it's harmless.
		cfg.FilePath = "costscope-data/costscope.db"
		if r, err := persistence.NewSQLiteRepository(cfg); err != nil {
			logger.WarnWithFields("Repository init failed; DB readiness check disabled", map[string]interface{}{"error": err.Error()})
		} else {
			repo = r
			defer func() { _ = repo.Close() }()
		}
	}

	// Handlers registry
	handlersRegistry := handlers.NewEnterpriseRegistry(logger, jobManager, wsManager)
	// Ensure handlers registry can clean up persistent resources on shutdown (e.g., BoltJobStore)
	defer func() {
		if handlersRegistry != nil {
			_ = handlersRegistry.Close()
		}
	}()

	// Optional report metadata persistence wiring (precedence: YAML reports.metadata_store_path > ENV COSTSCOPE_REPORTS_METADATA_PATH).
	var metadataStorePath string
	if loadedCfg != nil && strings.TrimSpace(loadedCfg.Reports.MetadataStorePath) != "" {
		metadataStorePath = strings.TrimSpace(loadedCfg.Reports.MetadataStorePath)
	} else if v := os.Getenv("COSTSCOPE_REPORTS_METADATA_PATH"); strings.TrimSpace(v) != "" {
		metadataStorePath = strings.TrimSpace(v)
	}
	// Emit unified precedence audit log (explicit flag currently unsupported; YAML > ENV > default)
	source := "default"
	if loadedCfg != nil && strings.TrimSpace(loadedCfg.Reports.MetadataStorePath) != "" {
		source = "yaml"
	} else if strings.TrimSpace(os.Getenv("COSTSCOPE_REPORTS_METADATA_PATH")) != "" {
		source = "env"
	}
	logger.InfoWithFields("config_precedence_resolved", map[string]interface{}{
		"field":  "reports.metadata_store_path",
		"value":  metadataStorePath,
		"source": source,
	})

	if metadataStorePath != "" {
		backend := "file"
		// Retention config (optional)
		maxRec := 0
		maxAge := time.Duration(0)
		if loadedCfg != nil {
			maxRec = loadedCfg.Reports.MetadataRetentionMaxRecords
			maxAge = loadedCfg.Reports.MetadataRetentionMaxAge
		}
		// Heuristic: if path has sqlite:// prefix strip and use sqlite backend (when tag enabled). Also allow .db extension.
		pathVal := metadataStorePath
		useSQLite := false
		if strings.HasPrefix(pathVal, "sqlite://") {
			pathVal = strings.TrimPrefix(pathVal, "sqlite://")
			useSQLite = true
		} else if strings.HasSuffix(strings.ToLower(pathVal), ".db") || strings.HasSuffix(strings.ToLower(pathVal), ".sqlite") || strings.HasSuffix(strings.ToLower(pathVal), ".sqlite3") {
			useSQLite = true
		}
		if useSQLite {
			if ms, err := reports.NewSQLiteMetadataStore(pathVal, logger, maxRec, maxAge); err != nil {
				logger.WarnWithFields("Failed to initialize sqlite metadata store; falling back to file store", map[string]interface{}{"error": err.Error()})
				msFile := reports.NewFileMetadataStore(metadataStorePath, logger, maxRec, maxAge)
				handlersRegistry.WithReportMetadataStore(msFile)
			} else {
				backend = "sqlite"
				handlersRegistry.WithReportMetadataStore(ms)
			}
		} else {
			ms := reports.NewFileMetadataStore(metadataStorePath, logger, maxRec, maxAge)
			handlersRegistry.WithReportMetadataStore(ms)
		}
		logger.InfoWithFields("Report metadata persistence enabled", map[string]interface{}{"path": metadataStorePath, "backend": backend, "retention_max_records": maxRec, "retention_max_age": maxAge.String()})
	} else {
		logger.Info("Report metadata persistence disabled (in-memory only)")
	}
	if repo != nil {
		handlersRegistry.HealthHandler.WithRepository(repo)
	}

	// Auth
	authMiddleware := middleware.JWTAuth(enterpriseJwtSecret, enterpriseJwtIssuer)
	apiKeyMiddleware := middleware.APIKeyAuth(enterpriseApiKeys)
	combinedAuth := middleware.CombinedAuth(authMiddleware, apiKeyMiddleware)

	tenantMiddleware := buildTenantMiddleware(loadedCfg)

	// CORS production guidance: warn if wildcard is used in production
	checkAndLogCorsWarning(logger, loadedCfg, enterpriseCorsOrigins)

	// RBAC service (file-backed store); failure is non-fatal (routes may still rely on simple role lists)
	rbacStore := security.NewFileRBACStore("data/security")
	_ = rbacStore.Load()
	rbacService := security.NewRBACService(rbacStore, logger)

	// Router
	router := newEnterpriseGinRouter()

	// Health and docs (no auth); debug cache stats route conditionally added below with auth
	registerHealthAndDocs(router, handlersRegistry)

	registerDebugRoutes(router, handlersRegistry, combinedAuth)

	// API v1 routes with authentication
	v1 := router.Group("/api/v1")
	v1.Use(combinedAuth, tenantMiddleware)
	RegisterGinRouteGroups(v1, buildModuleRouteGroups(handlersRegistry, logger, rbacService))

	// Enterprise-only structured AuthMiddleware route (no-op on slim builds)
	registerEnterpriseStructuredAuthRoutes(router, logger)

	// WebSocket endpoints (with authentication)
	registerWebSocketRoutes(router, combinedAuth, handlersRegistry)

	// HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", enterpriseHost, enterprisePort),
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Apply hardened TLS configuration when enabled
	if enterpriseTlsEnabled {
		server.TLSConfig = buildTLSConfig(tlsMinVersion, tlsCipherSuites, tlsPreferServerCiphers)
		logger.InfoWithFields("TLS configuration applied", map[string]interface{}{
			"min_version":           tlsMinVersion,
			"cipher_suites_count":   len(tlsCipherSuites),
			"prefer_server_ciphers": tlsPreferServerCiphers,
		})
	}

	// Optional Casbin wrapping
	wrapServerWithCasbinIfEnabled(&server.Handler, logger)

	// Background cache metrics refresher (optional)
	stopCh := make(chan struct{})
	// Pre-warm normalization caches so cache gauges are visible early and cache hit paths are hot.
	normalization.PreWarm()
	if cacheMetricsRefreshInterval > 0 {
		logger.InfoWithFields("Starting cache metrics refresher", map[string]interface{}{"interval": cacheMetricsRefreshInterval.String()})
		normalization.StartCacheMetricsRefresher(cacheMetricsRefreshInterval, stopCh)
	}

	// Start server
	serverErrors := make(chan error, 1)
	go func() {
		logger.InfoWithFields("Enterprise API server listening", map[string]interface{}{
			"address": server.Addr,
			"tls":     enterpriseTlsEnabled,
		})
		if enterpriseTlsEnabled {
			serverErrors <- server.ListenAndServeTLS(enterpriseTlsCert, enterpriseTlsKey)
		} else {
			serverErrors <- server.ListenAndServe()
		}
	}()

	// Wait for shutdown signal
	shutdownCh := make(chan os.Signal, 1)
	signal.Notify(shutdownCh, os.Interrupt)

	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %v", err)
	case <-shutdownCh:
		logger.Info("Starting graceful shutdown...")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		close(stopCh)
		if err := server.Shutdown(ctx); err != nil {
			logger.ErrorWithFields("Server shutdown error", map[string]interface{}{"error": err.Error()})
			return err
		}
		logger.Info("Enterprise API server stopped")
		return nil
	}
}

// newEnterpriseGinRouter creates a Gin engine with common middlewares configured
func newEnterpriseGinRouter() *gin.Engine {
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.New()

	// CORS
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = enterpriseCorsOrigins
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization", "X-API-Key", "X-Tenant-ID"}
	corsConfig.AllowCredentials = true
	router.Use(cors.New(corsConfig))

	// Global middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.RequestID())
	router.Use(middleware.Tracing("costscope-enterprise-api"))
	router.Use(middleware.RequestLogging())
	router.Use(middleware.SecurityHeaders())
	router.Use(middleware.RateLimit(enterpriseRateLimitRequests, enterpriseRateLimitWindow))
	router.Use(middleware.Prometheus())

	// Metrics (no auth)
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))
	return router
}

func registerHealthAndDocs(router *gin.Engine, reg *handlers.EnterpriseRegistry) {
	router.GET(healthPath, reg.HealthHandler.HealthCheck)
	router.GET("/health/ready", reg.HealthHandler.ReadinessCheck)
	router.GET("/health/live", reg.HealthHandler.LivenessCheck)

	if docsEnabled {
		router.GET(docsPath, reg.DocsHandler.GetDocumentation)
		router.GET(docsPath+"/*filepath", reg.DocsHandler.ServeSwaggerUI)
	}
}

func registerWebSocketRoutes(router *gin.Engine, auth gin.HandlerFunc, reg *handlers.EnterpriseRegistry) {
	ws := router.Group("/ws")
	ws.Use(auth)
	ws.GET("/jobs/:jobID", reg.WebSocketHandler.Connect)
}

// validateEnterpriseJWTSecret enforces a minimum length policy for the JWT secret.
// Returns an error when the secret is empty or shorter than 32 bytes.
func validateEnterpriseJWTSecret(secret string) error {
	if len(secret) < 32 {
		return fmt.Errorf("jwt secret must be at least 32 bytes; got %d", len(secret))
	}
	return nil
}

// isProductionEnv determines if the server is running in production.
// Prefers loaded YAML config; falls back to COSTSCOPE_ENVIRONMENT or ENV env vars.
func isProductionEnv(cfg *config.ConsolidatedConfig) bool {
	if cfg != nil {
		return cfg.Environment == config.Production
	}
	if v := os.Getenv("COSTSCOPE_ENVIRONMENT"); strings.EqualFold(v, "production") {
		return true
	}
	if v := os.Getenv("ENV"); strings.EqualFold(v, "production") {
		return true
	}
	return false
}

// containsWildcardOrigin returns true when "*" is present in origins.
func containsWildcardOrigin(origins []string) bool {
	for _, o := range origins {
		if strings.TrimSpace(o) == "*" {
			return true
		}
	}
	return false
}

// checkAndLogCorsWarning emits a warning when running in production with wildcard CORS origins.
// Separated for testability.
func checkAndLogCorsWarning(logger *logging.Logger, cfg *config.ConsolidatedConfig, origins []string) {
	if isProductionEnv(cfg) && containsWildcardOrigin(origins) {
		logger.WarnWithFields("CORS wildcard '*' is configured in production; restrict origins for security", map[string]interface{}{
			"origins": origins,
		})
	}
}

// buildTLSConfig constructs a hardened *tls.Config using the provided settings.
// - minVersion: "1.2" or "1.3" (default 1.2)
// - cipherSuites: optional list of TLS 1.2 cipher names; empty => secure defaults
// - preferServerCiphers: whether to prefer server cipher order
func buildTLSConfig(minVersion string, cipherSuites []string, preferServerCiphers bool) *tls.Config {
	cfg := &tls.Config{
		MinVersion:               tls.VersionTLS12,
		PreferServerCipherSuites: preferServerCiphers,
		CurvePreferences: []tls.CurveID{
			tls.X25519, tls.CurveP256, tls.CurveP384,
		},
	}
	if strings.TrimSpace(minVersion) == "1.3" {
		cfg.MinVersion = tls.VersionTLS13
		// Cipher suites are fixed for TLS 1.3 in Go; CipherSuites field is ignored.
		return cfg
	}
	// TLS 1.2: apply secure defaults unless user provided suites
	var suites []uint16
	if len(cipherSuites) == 0 {
		suites = []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
		}
	} else {
		for _, name := range cipherSuites {
			if id, ok := tlsCipherNameToID[strings.TrimSpace(strings.ToUpper(name))]; ok {
				suites = append(suites, id)
			}
		}
		if len(suites) == 0 {
			// Fallback to defaults if none recognized
			suites = []uint16{
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
				tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			}
		}
	}
	cfg.CipherSuites = suites
	return cfg
}

// tlsCipherNameToID maps string names to Go's uint16 cipher IDs (TLS 1.2 only).
var tlsCipherNameToID = map[string]uint16{
	"TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256": tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
	"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256":   tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
	"TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384": tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
	"TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384":   tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
	"TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305":  tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
	"TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305":    tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
}

// buildModuleRouteGroups describes all module routes in a table-driven form
func buildModuleRouteGroups(registry *handlers.EnterpriseRegistry, logger *logging.Logger, rbac *security.RBACService) []GinRouteGroup {
	groups := []GinRouteGroup{
		{
			BasePath: "/focus",
			Routes: []GinRoute{
				{Method: http.MethodPost, Path: "/convert", Handler: func(c *gin.Context) {
					middleware.RequirePermission(rbac, security.ResourceFocus, security.ActionConvert)(c)
					if c.IsAborted() {
						return
					}
					registry.FocusHandler.ConvertData(c)
				}},
				// Async conversion job endpoints (new)
				{Method: http.MethodPost, Path: "/conversions", Handler: func(c *gin.Context) {
					middleware.RequirePermission(rbac, security.ResourceFocus, security.ActionConvert)(c)
					if c.IsAborted() {
						return
					}
					registry.FocusHandler.ConvertData(c)
				}},
				{Method: http.MethodGet, Path: "/conversions/:id", Handler: registry.FocusHandler.GetJob},
				{Method: http.MethodGet, Path: "/conversions", Handler: registry.FocusHandler.ListJobs},
				{Method: http.MethodDelete, Path: "/conversions/:id", Handler: registry.FocusHandler.CancelJob},
				{Method: http.MethodGet, Path: "/conversions/history", Handler: registry.FocusHandler.ListJobHistory},
				{Method: http.MethodPost, Path: "/analyze", Handler: registry.FocusHandler.AnalyzeData},
				{Method: http.MethodPost, Path: "/compare", Handler: registry.FocusHandler.CompareData},
				{Method: http.MethodPost, Path: "/validate", Handler: func(c *gin.Context) {
					middleware.RequirePermission(rbac, security.ResourceFocus, security.ActionValidate)(c)
					if c.IsAborted() {
						return
					}
					registry.FocusHandler.ValidateData(c)
				}},
				{Method: http.MethodGet, Path: "/jobs/:id", Handler: registry.FocusHandler.GetJob},
				{Method: http.MethodGet, Path: "/jobs", Handler: registry.FocusHandler.ListJobs},
				{Method: http.MethodDelete, Path: "/jobs/:id", Handler: registry.FocusHandler.CancelJob},
			},
		},
		{
			BasePath: "/providers",
			Routes: []GinRoute{
				{Method: http.MethodGet, Path: "", Handler: registry.ProvidersHandler.ListProviders},
				{Method: http.MethodGet, Path: "/:provider", Handler: registry.ProvidersHandler.GetProvider},
				{Method: http.MethodPost, Path: "/:provider/connect", Handler: func(c *gin.Context) {
					middleware.RequirePermission(rbac, security.ResourceProviders, security.ActionConnect)(c)
					if c.IsAborted() {
						return
					}
					registry.ProvidersHandler.ConnectProvider(c)
				}},
				{Method: http.MethodPost, Path: "/:provider/test", Handler: func(c *gin.Context) {
					middleware.RequirePermission(rbac, security.ResourceProviders, security.ActionConnect)(c)
					if c.IsAborted() {
						return
					}
					registry.ProvidersHandler.TestConnection(c)
				}},
				{Method: http.MethodGet, Path: "/:provider/accounts", Handler: registry.ProvidersHandler.ListAccounts},
				{Method: http.MethodGet, Path: "/:provider/regions", Handler: registry.ProvidersHandler.ListRegions},
				{Method: http.MethodGet, Path: "/:provider/services", Handler: registry.ProvidersHandler.ListServices},
			},
		},
		{
			BasePath: "/analytics",
			Routes: []GinRoute{
				{Method: http.MethodPost, Path: "/forecast", Handler: func(c *gin.Context) {
					middleware.RequirePermission(rbac, security.ResourceAnalytics, security.ActionForecast)(c)
					if c.IsAborted() {
						return
					}
					registry.AnalyticsHandler.GenerateForecast(c)
				}},
				{Method: http.MethodPost, Path: "/anomalies", Handler: func(c *gin.Context) {
					middleware.RequirePermission(rbac, security.ResourceAnalytics, security.ActionDetectAnomalies)(c)
					if c.IsAborted() {
						return
					}
					registry.AnalyticsHandler.DetectAnomalies(c)
				}},
				{Method: http.MethodPost, Path: "/recommendations", Handler: func(c *gin.Context) {
					middleware.RequirePermission(rbac, security.ResourceAnalytics, security.ActionRecommendations)(c)
					if c.IsAborted() {
						return
					}
					registry.AnalyticsHandler.GetRecommendations(c)
				}},
				{Method: http.MethodPost, Path: "/trends", Handler: func(c *gin.Context) {
					middleware.RequirePermission(rbac, security.ResourceAnalytics, security.ActionTrends)(c)
					if c.IsAborted() {
						return
					}
					registry.AnalyticsHandler.AnalyzeTrends(c)
				}},
				{Method: http.MethodGet, Path: "/models", Handler: registry.AnalyticsHandler.ListModels},
				{Method: http.MethodPost, Path: "/models/:id/train", Handler: func(c *gin.Context) {
					middleware.RequirePermission(rbac, security.ResourceAnalytics, security.ActionTrainModel)(c)
					if c.IsAborted() {
						return
					}
					registry.AnalyticsHandler.TrainModel(c)
				}},
				// Facade-backed lightweight GET analytics (DuckDB-enabled builds)
				{Method: http.MethodGet, Path: "/summary", Handler: registry.AnalyticsReadHandler.Summary},
				{Method: http.MethodGet, Path: "/top-services", Handler: registry.AnalyticsReadHandler.TopServices},
				{Method: http.MethodGet, Path: "/trends", Handler: registry.AnalyticsReadHandler.Trends},
				{Method: http.MethodGet, Path: "/jobs/:id", Handler: registry.AnalyticsHandler.GetAnalyticsJob},
			},
		},
		{
			BasePath: "/reports",
			Routes: []GinRoute{
				{Method: http.MethodPost, Path: "/generate", Handler: func(c *gin.Context) {
					middleware.RequirePermission(rbac, security.ResourceReports, security.ActionGenerate)(c)
					if c.IsAborted() {
						return
					}
					registry.ReportsHandler.GenerateReport(c)
				}},
				{Method: http.MethodGet, Path: "", Handler: registry.ReportsHandler.ListReports},
				// Export metadata listing & integrity verification (additive endpoints)
				{Method: http.MethodGet, Path: "/exports", Handler: registry.ReportsHandler.ListExports},
				{Method: http.MethodGet, Path: "/exports/:id/verify", Handler: registry.ReportsHandler.VerifyExport},
				{Method: http.MethodGet, Path: "/:id", Handler: registry.ReportsHandler.GetReport},
				{Method: http.MethodDelete, Path: "/:id", Handler: registry.ReportsHandler.DeleteReport},
				{Method: http.MethodGet, Path: "/:id/download", Handler: registry.ReportsHandler.DownloadReport},
				{Method: http.MethodGet, Path: "/templates", Handler: registry.ReportsHandler.ListTemplates},
				{Method: http.MethodPost, Path: "/schedule", Handler: registry.ReportsHandler.ScheduleReport},
			},
		},
		{
			BasePath: "/streaming",
			Routes: []GinRoute{
				{Method: http.MethodPost, Path: "/jobs", Handler: func(c *gin.Context) {
					middleware.RequirePermission(rbac, security.ResourceStreaming, security.ActionCreateJob)(c)
					if c.IsAborted() {
						return
					}
					registry.StreamingHandler.CreateStreamingJob(c)
				}},
				{Method: http.MethodGet, Path: "/jobs", Handler: registry.StreamingHandler.ListStreamingJobs},
				{Method: http.MethodGet, Path: "/jobs/:id", Handler: registry.StreamingHandler.GetStreamingJob},
				{Method: http.MethodPost, Path: "/jobs/:id/start", Handler: func(c *gin.Context) {
					middleware.RequirePermission(rbac, security.ResourceStreaming, security.ActionStartJob)(c)
					if c.IsAborted() {
						return
					}
					registry.StreamingHandler.StartJob(c)
				}},
				{Method: http.MethodPost, Path: "/jobs/:id/stop", Handler: func(c *gin.Context) {
					middleware.RequirePermission(rbac, security.ResourceStreaming, security.ActionStopJob)(c)
					if c.IsAborted() {
						return
					}
					registry.StreamingHandler.StopJob(c)
				}},
				{Method: http.MethodDelete, Path: "/jobs/:id", Handler: func(c *gin.Context) {
					middleware.RequirePermission(rbac, security.ResourceStreaming, security.ActionDeleteJob)(c)
					if c.IsAborted() {
						return
					}
					registry.StreamingHandler.DeleteJob(c)
				}},
				{Method: http.MethodGet, Path: "/sources", Handler: registry.StreamingHandler.ListSources},
			},
		},
		{
			BasePath: "/monitoring",
			Routes: []GinRoute{
				{Method: http.MethodGet, Path: "/metrics", Handler: registry.MonitoringHandler.GetMetrics},
				{Method: http.MethodGet, Path: "/alerts", Handler: registry.MonitoringHandler.ListAlerts},
				{Method: http.MethodPost, Path: "/alerts", Handler: registry.MonitoringHandler.CreateAlert},
				{Method: http.MethodGet, Path: "/alerts/:id", Handler: registry.MonitoringHandler.GetAlert},
				{Method: http.MethodPut, Path: "/alerts/:id", Handler: registry.MonitoringHandler.UpdateAlert},
				{Method: http.MethodDelete, Path: "/alerts/:id", Handler: registry.MonitoringHandler.DeleteAlert},
				{Method: http.MethodGet, Path: "/dashboards", Handler: registry.MonitoringHandler.ListDashboards},
			},
		},
		{
			BasePath: "/integration",
			Routes: []GinRoute{
				{Method: http.MethodGet, Path: "/connectors", Handler: registry.IntegrationHandler.ListConnectors},
				{Method: http.MethodPost, Path: "/connectors", Handler: registry.IntegrationHandler.CreateConnector},
				{Method: http.MethodGet, Path: "/connectors/:id", Handler: registry.IntegrationHandler.GetConnector},
				{Method: http.MethodPut, Path: "/connectors/:id", Handler: registry.IntegrationHandler.UpdateConnector},
				{Method: http.MethodDelete, Path: "/connectors/:id", Handler: registry.IntegrationHandler.DeleteConnector},
				{Method: http.MethodPost, Path: "/connectors/:id/test", Handler: registry.IntegrationHandler.TestConnector},
				{Method: http.MethodPost, Path: "/sync", Handler: registry.IntegrationHandler.SyncData},
			},
		},
		{
			BasePath: "/production",
			Routes: []GinRoute{
				{Method: http.MethodPost, Path: "/assess", Handler: registry.ProductionHandler.AssessReadiness},
				{Method: http.MethodGet, Path: "/assessments", Handler: registry.ProductionHandler.ListAssessments},
				{Method: http.MethodGet, Path: "/assessments/:id", Handler: registry.ProductionHandler.GetAssessment},
				{Method: http.MethodGet, Path: "/benchmarks", Handler: registry.ProductionHandler.ListBenchmarks},
				{Method: http.MethodPost, Path: "/validate", Handler: registry.ProductionHandler.ValidateConfiguration},
			},
		},
		{
			BasePath: "/config",
			Routes: []GinRoute{
				{Method: http.MethodGet, Path: "/profiles", Handler: registry.ConfigHandler.ListProfiles},
				{Method: http.MethodGet, Path: "/profiles/:name", Handler: registry.ConfigHandler.GetProfile},
				{Method: http.MethodPost, Path: "/profiles", Handler: registry.ConfigHandler.CreateProfile},
				{Method: http.MethodPut, Path: "/profiles/:name", Handler: registry.ConfigHandler.UpdateProfile},
				{Method: http.MethodDelete, Path: "/profiles/:name", Handler: registry.ConfigHandler.DeleteProfile},
				{Method: http.MethodPost, Path: "/validate", Handler: registry.ConfigHandler.ValidateConfig},
				{Method: http.MethodGet, Path: "/schema", Handler: registry.ConfigHandler.GetSchema},
			},
		},
		// Preview multicloud endpoints (mock/aggregate data). Additive & optional.
		{
			BasePath: "/multicloud",
			Routes: []GinRoute{
				{Method: http.MethodPost, Path: "/recommendations", Handler: registry.MulticloudHandler.Recommendations},
				{Method: http.MethodGet, Path: "/inventory", Handler: registry.MulticloudHandler.Inventory},
				{Method: http.MethodPost, Path: "/migration/plan", Handler: registry.MulticloudHandler.MigrationPlan},
				{Method: http.MethodPost, Path: "/migration/feasibility", Handler: registry.MulticloudHandler.MigrationFeasibility},
			},
		},
	}

	// Append DSL-driven integration action routes under /integration base path (POST endpoints)
	// This keeps the addition purely additive and does not interfere with existing routes.
	dslRoutes := buildIntegrationActionRoutes(logger)
	if len(dslRoutes) > 0 {
		groups = append(groups, GinRouteGroup{BasePath: "/integration", Routes: dslRoutes})
	}
	return groups
}

// wrapServerWithCasbinIfEnabled optionally wraps the server handler with Casbin RBAC
func wrapServerWithCasbinIfEnabled(handler *http.Handler, logger *logging.Logger) {
	if !enterpriseCasbinEnabled {
		return
	}

	modelPath := enterpriseCasbinModelPath
	policyPath := enterpriseCasbinPolicyPath
	if modelPath == "" || policyPath == "" {
		logger.Warn("Casbin enabled but model or policy path not provided; skipping")
		return
	}

	enforcer, err := security.NewEnforcerFromFiles(modelPath, policyPath)
	if err != nil {
		logger.WarnWithFields("Casbin init failed", map[string]interface{}{"error": err.Error()})
		return
	}
	logger.Info("Casbin RBAC enabled for Enterprise API")
	*handler = security.CasbinHTTPMiddleware(*handler, enforcer, security.JWTSubjectExtractor(enterpriseJwtSecret, enterpriseJwtIssuer))
}
