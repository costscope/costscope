package api

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/spf13/cobra"

	"local/costscope/internal/api/jobs"
	"local/costscope/internal/api/websocket"
	"local/costscope/internal/core/config"
	"local/costscope/internal/core/logging"
	"local/costscope/internal/core/security"
)

// =====================================================================================
// API Server Command - Enterprise REST API for FOCUS Operations
// =====================================================================================

var (
	// Server settings
	apiHost string

	// Rate limiting
	rateLimitRequests int
	rateLimitWindow   time.Duration

	// Network / port
	apiPort int

	// Auth / security
	jwtSecret   string
	jwtIssuer   string
	corsOrigins []string

	// Job processing
	jobWorkers int

	// TLS settings
	tlsEnabled bool
	tlsCert    string
	tlsKey     string
)

// BuildAPICommand creates the API server command
func BuildAPICommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the FOCUS API server",
		Long: `Start the enterprise API server for FOCUS operations.

The API server provides REST endpoints for:
- FOCUS data conversion (async)
- Cost analysis (async)  
- Dataset comparison (async)
- FOCUS validation (sync)
- Job status monitoring
- Real-time WebSocket updates

Features:
- JWT authentication and authorization
- Role-based access control (RBAC)
- API rate limiting
- CORS support
- Real-time progress updates via WebSocket
- OpenAPI/Swagger documentation

Authentication:
The API uses JWT tokens for authentication. You can generate tokens using:
  costscope auth generate --user-id user123 --roles focus:admin

Examples:
  # Start API server on default port 8080
  costscope serve

  # Start with custom host and port
  costscope serve --host 0.0.0.0 --port 9090

  # Start with TLS enabled
  costscope serve --tls --cert server.crt --key server.key

  # Start with custom JWT settings
	costscope serve --jwt-secret <YOUR_LONG_RANDOM_SECRET> --jwt-issuer costscope

  # Start with rate limiting
  costscope serve --rate-limit 100 --rate-window 1m`,
		RunE: runAPIServer,
	}

	// Server settings
	cmd.Flags().StringVar(&apiHost, "host", "localhost", "API server host")
	cmd.Flags().IntVar(&apiPort, "port", 8080, "API server port")

	// Security settings (no default secret; must be provided via flag/env/config)
	cmd.Flags().StringVar(&jwtSecret, "jwt-secret", "", "JWT signing secret (required; or set COSTSCOPE_JWT_SECRET / security.jwt_secret in config)")
	cmd.Flags().StringVar(&jwtIssuer, "jwt-issuer", "costscope", "JWT issuer")

	// CORS settings
	cmd.Flags().StringSliceVar(&corsOrigins, "cors-origins", []string{"*"}, "CORS allowed origins")

	// Rate limiting
	cmd.Flags().IntVar(&rateLimitRequests, "rate-limit", 100, "Rate limit requests per window")
	cmd.Flags().DurationVar(&rateLimitWindow, "rate-window", time.Minute, "Rate limit time window")

	// Job processing
	cmd.Flags().IntVar(&jobWorkers, "workers", 4, "Number of job worker goroutines")

	// TLS settings
	cmd.Flags().BoolVar(&tlsEnabled, "tls", false, "Enable TLS/HTTPS")
	cmd.Flags().StringVar(&tlsCert, "cert", "server.crt", "TLS certificate file")
	cmd.Flags().StringVar(&tlsKey, "key", "server.key", "TLS private key file")

	return cmd
}

func runAPIServer(cmd *cobra.Command, args []string) error {
	// Initialize logger
	logger := logging.NewLogger(logging.LevelInfo)
	logger.Info("Starting FOCUS API server...")

	// Resolve JWT secret with precedence: flag > env > YAML config
	resolvedSecret, err := resolveJWTSecret(cmd, logger)
	if err != nil {
		logger.Error(err.Error())
		return err
	}
	jwtSecret = resolvedSecret
	if len(jwtSecret) < 32 { // weak secret warning (not fatal)
		logger.Warn("JWT secret length < 32 bytes; consider using a longer random value (e.g. openssl rand -base64 48)")
	}

	// In test mode, skip starting network server to allow unit tests to validate secret resolution logic quickly.
	if os.Getenv("COSTSCOPE_TEST_MODE") == "1" {
		logger.Info("Test mode detected - skipping server startup")
		return nil
	}

	// Initialize mock managers
	conversionMgr := NewMockConversionManager(logger)
	analysisMgr := NewMockAnalysisManager(logger)
	comparisonMgr := NewMockComparisonManager(logger)
	validationMgr := NewMockValidationManager(logger)

	// Initialize job manager
	jobManager := jobs.NewManager(logger, jobWorkers)
	if err := jobManager.Start(); err != nil {
		return fmt.Errorf("failed to start job manager: %w", err)
	}
	defer func() {
		if err := jobManager.Stop(); err != nil {
			logger.Error(fmt.Sprintf("Failed to stop job manager: %s", err.Error()))
		}
	}()

	// Initialize WebSocket manager
	wsManager := websocket.NewManager(logger)
	// Wire job manager to broadcaster for real-time updates
	jobManager.SetBroadcaster(wsManager)

	// Create router
	router := setupRouter(
		logger,
		jobManager,
		wsManager,
		conversionMgr,
		analysisMgr,
		comparisonMgr,
		validationMgr,
	)

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", apiHost, apiPort),
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Optional: Wrap with Casbin RBAC if configured via env
	modelPath := os.Getenv("CASBIN_MODEL_PATH")
	policyPath := os.Getenv("CASBIN_POLICY_PATH")
	if modelPath != "" && policyPath != "" {
		enf, err := security.NewEnforcerFromFiles(modelPath, policyPath)
		if err != nil {
			logger.Warn(fmt.Sprintf("Casbin disabled: %v", err))
		} else {
			logger.Info("Casbin RBAC enabled")
			server.Handler = security.CasbinHTTPMiddleware(
				server.Handler,
				enf,
				security.JWTSubjectExtractor(jwtSecret, jwtIssuer),
			)
		}
	}

	// Start server in goroutine
	go func() {
		var err error
		if tlsEnabled {
			logger.Info(fmt.Sprintf("Starting HTTPS server on %s:%d", apiHost, apiPort))
			err = server.ListenAndServeTLS(tlsCert, tlsKey)
		} else {
			logger.Info(fmt.Sprintf("Starting HTTP server on %s:%d", apiHost, apiPort))
			err = server.ListenAndServe()
		}

		if err != nil && err != http.ErrServerClosed {
			logger.Error(fmt.Sprintf("Server error: %s", err.Error()))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	logger.Info("Shutting down API server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error(fmt.Sprintf("Server shutdown error: %s", err.Error()))
		return err
	}

	logger.Info("API server stopped")
	return nil
}

// resolveJWTSecret determines the JWT secret using precedence: flag > env > YAML config (security.jwt_secret)
func resolveJWTSecret(cmd *cobra.Command, logger *logging.Logger) (string, error) {
	// Use centralized resolver: explicit (flag) > YAML > ENV > default("")
	var explicit *string
	if cmd.Flags().Changed("jwt-secret") && jwtSecret != "" {
		explicit = &jwtSecret
	}
	res := config.ResolveStringField(logger, "security.jwt_secret", explicit, func(cc *config.ConsolidatedConfig) *string {
		if cc == nil || cc.Security.JWTSecret == "" {
			return nil
		}
		return &cc.Security.JWTSecret
	}, "COSTSCOPE_JWT_SECRET", "")
	if res.Value != "" {
		return res.Value, nil
	}
	return "", fmt.Errorf("jwt secret not provided: set --jwt-secret, COSTSCOPE_JWT_SECRET env, or security.jwt_secret in config YAML")
}

// cfgShimLogger adapts *logging.Logger to config.Logger interface
// cfgShimLogger removed; unified Resolve*Field helpers perform YAML loading + logging.

func setupRouter(
	logger *logging.Logger,
	_ *jobs.Manager,
	wsMgr *websocket.Manager,
	_ *MockConversionManager,
	_ *MockAnalysisManager,
	_ *MockComparisonManager,
	_ *MockValidationManager,
) http.Handler {
	mux := http.NewServeMux()
	specs := BuildRouteSpecs(logger)
	common := buildCommonMiddleware()
	// expose the shared websocket manager to handlers
	sharedWSManager = wsMgr
	registerRouteSpecs(mux, specs, common, logger)
	return mux
}
