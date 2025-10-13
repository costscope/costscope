package handlers

import (
	"context"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/costscope/costscope/internal/api/jobs"
	"github.com/costscope/costscope/internal/api/response"
	"github.com/costscope/costscope/internal/core/logging"
	"github.com/costscope/costscope/internal/core/monitoring/telemetry"
	"github.com/costscope/costscope/internal/core/normalization"
	"github.com/costscope/costscope/internal/core/persistence"
	duckdb "github.com/costscope/costscope/internal/database/duckdb"
	"github.com/costscope/costscope/internal/providers/registry"
)

// =====================================================================================
// Health Handler - System Health and Readiness Checks
// =====================================================================================

// HealthHandler provides system health check endpoints
type HealthHandler struct {
	logger *logging.Logger
	start  time.Time
	jobs   *jobs.Manager
	repo   persistence.Repository
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(logger *logging.Logger) *HealthHandler {
	return &HealthHandler{
		logger: logger,
		start:  time.Now(),
	}
}

// WithJobs attaches the job manager to the health handler for readiness checks.
func (h *HealthHandler) WithJobs(m *jobs.Manager) *HealthHandler {
	h.jobs = m
	return h
}

// WithRepository attaches a persistence repository for DB health checks.
func (h *HealthHandler) WithRepository(r persistence.Repository) *HealthHandler {
	h.repo = r
	return h
}

// HealthCheck returns basic server health status
func (h *HealthHandler) HealthCheck(c *gin.Context) {
	payload := gin.H{
		"status":    "healthy",
		"server":    "CostScope Enterprise API",
		"version":   "1.0.0",
		"timestamp": "2025-01-31T00:00:00Z",
	}

	response.AutoOK200(c, payload)
}

// ReadinessCheck returns server readiness status
func (h *HealthHandler) ReadinessCheck(c *gin.Context) {
	// Minimal readiness: process is up and core subsystems are initialized.
	// Today we can safely assert job manager state when available.
	checks := gin.H{}
	ready := true

	// Background jobs subsystem
	if h.jobs != nil && h.jobs.IsRunning() {
		checks["background_jobs"] = "ok"
	} else if h.jobs != nil {
		checks["background_jobs"] = "not_running"
		// Log for diagnostics visibility
		if h.logger != nil {
			h.logger.Warn("readiness: background jobs manager not running")
		}
		ready = false
	} else {
		checks["background_jobs"] = "unknown"
	}

	const statusError = "error"
	// Database repository (optional)
	if h.repo != nil {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 500*time.Millisecond)
		defer cancel()
		if err := h.repo.Health(ctx); err != nil {
			checks["database"] = statusError
			checks["database_error"] = err.Error()
			if h.logger != nil {
				h.logger.Error("readiness database health error: " + err.Error())
			}
			ready = false
		} else {
			checks["database"] = "ok"
		}
	} else {
		checks["database"] = "unknown"
	}

	// DuckDB embedding (optional, build-tag aware)
	if duckdb.Linked() {
		if err := duckdb.QuickPing(); err != nil {
			checks["duckdb"] = statusError
			checks["duckdb_error"] = err.Error()
			if h.logger != nil {
				h.logger.Warn("readiness duckdb quick ping error: " + err.Error())
			}
			// non-fatal: only mark not ready when the build expects DuckDB-backed features
			// Keep as soft signal to avoid blocking slim deployments unintentionally.
		} else {
			checks["duckdb"] = "ok"
		}
	} else {
		checks["duckdb"] = "not_linked"
	}

	// Provider registry sanity check (lightweight). We treat the absence of any registered
	// providers as a potential wiring fault (should have aws/azure/gcp side-effect regs).
	// Additionally, an env flag allows simulating a provider registry failure for
	// integration / chaos testing without altering build tags.
	// Env: COSTSCOPE_SIMULATE_PROVIDER_REGISTRY_FAILURE (truthy => force failure)
	if simulate := isTruthy(os.Getenv("COSTSCOPE_SIMULATE_PROVIDER_REGISTRY_FAILURE")); simulate {
		checks["provider_registry"] = statusError
		checks["provider_registry_error"] = "simulated failure"
		ready = false
	} else {
		// Defer import to avoid cycle – small helper inline referencing registry via fully qualified path.
		// (We keep it inside the else to avoid needless work on simulated path.)
		if list := registry.List(); len(list) == 0 {
			checks["provider_registry"] = "empty"
			ready = false
		} else {
			checks["provider_registry"] = "ok"
			checks["provider_registry_count"] = len(list)
		}
	}

	status := "ready"
	code := http.StatusOK
	if !ready {
		status = "not_ready"
		code = http.StatusServiceUnavailable
	}

	// Emit a single structured readiness diagnostics log entry with clear fields
	if h.logger != nil {
		h.logger.InfoWithFields("readiness_check", map[string]interface{}{
			"status": status,
			"ready":  ready,
			"checks": checks,
		})
	}

	// Update readiness gauge metric (1 = ready, 0 = not ready). Non-blocking optional signal.
	if ready {
		telemetry.HealthReadiness.Set(1)
	} else {
		telemetry.HealthReadiness.Set(0)
	}

	response.AutoOK(c, code, gin.H{
		"status": status,
		"checks": checks,
	})
}

// isTruthy returns true for common truthy strings.
func isTruthy(v string) bool {
	if v == "" {
		return false
	}
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "1", "true", "t", "yes", "y", "on", "enable", "enabled":
		return true
	}
	return false
}

// (No additional helpers for registry; direct call used above.)

// LivenessCheck returns server liveness status
func (h *HealthHandler) LivenessCheck(c *gin.Context) {
	// Calculate process uptime based on handler initialization time
	uptime := time.Since(h.start).Truncate(time.Second).String()
	payload := gin.H{
		"status": "alive",
		"uptime": uptime,
	}

	response.AutoOK200(c, payload)
}

// CacheStats returns runtime normalization cache stats (debug/admin usage).
func (h *HealthHandler) CacheStats(c *gin.Context) {
	// Note: We only expose normalizer region/unit caches for now; unified enum caches
	// are intentionally omitted to limit surface unless needed.
	regionHits, regionMiss, regionEvict, regionSize := normalization.RegionCacheStats()
	unitHits, unitMiss, unitEvict, unitSize := normalization.UnitCacheStats()
	response.AutoOK200(c, gin.H{
		"region_cache": gin.H{
			"hits":      regionHits,
			"misses":    regionMiss,
			"evictions": regionEvict,
			"size":      regionSize,
		},
		"unit_cache": gin.H{
			"hits":      unitHits,
			"misses":    unitMiss,
			"evictions": unitEvict,
			"size":      unitSize,
		},
	})
}
