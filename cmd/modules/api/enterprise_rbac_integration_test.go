package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/costscope/costscope/internal/api/middleware"
	"github.com/costscope/costscope/internal/core/logging"
	"github.com/costscope/costscope/internal/core/monitoring/telemetry"
	"github.com/costscope/costscope/internal/core/security"
)

// buildTestRBAC constructs an RBAC service with a single role 'analyst' having the provided permissions.
func buildTestRBAC(t *testing.T, perms []security.Permission) *security.RBACService {
	t.Helper()
	store := security.NewFileRBACStore(t.TempDir())
	if err := store.Load(); err != nil {
		t.Fatalf("load store: %v", err)
	}
	svc := security.NewRBACService(store, logging.NewLogger(logging.LevelError))
	if _, err := svc.CreateRole("analyst", "", perms); err != nil {
		t.Fatalf("create role: %v", err)
	}
	return svc
}

// TestEnterpriseRBAC_RoutesAndAudit verifies allowed, denied, and audit soft-deny flows on the enterprise router
// for analytics + streaming endpoints now wrapped with RequirePermission.
func TestEnterpriseRBAC_RoutesAndAudit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	defer func() { _ = recover() }() // guard duplicate telemetry register
	telemetry.Register()

	// Grant subset of permissions (missing analytics train_model and streaming delete_job)
	perms := []security.Permission{
		{Resource: security.ResourceAnalytics, Action: security.ActionForecast},
		{Resource: security.ResourceStreaming, Action: security.ActionCreateJob},
		{Resource: security.ResourceStreaming, Action: security.ActionStartJob},
	}
	rbac := buildTestRBAC(t, perms)

	// Build a minimal gin engine reusing enterprise middlewares but registering only tested routes.
	router := gin.New()
	router.Use(middleware.Prometheus())

	// Metrics endpoint
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Register subset endpoints with permission middleware
	router.POST("/api/v1/analytics/forecast", middleware.RequirePermission(rbac, security.ResourceAnalytics, security.ActionForecast), func(c *gin.Context) { c.String(http.StatusOK, "forecast") })
	router.POST("/api/v1/analytics/train", middleware.RequirePermission(rbac, security.ResourceAnalytics, security.ActionTrainModel), func(c *gin.Context) { c.String(http.StatusOK, "train") })
	router.POST("/api/v1/streaming/jobs", middleware.RequirePermission(rbac, security.ResourceStreaming, security.ActionCreateJob), func(c *gin.Context) { c.String(http.StatusOK, "create") })
	router.POST("/api/v1/streaming/jobs/start", middleware.RequirePermission(rbac, security.ResourceStreaming, security.ActionStartJob), func(c *gin.Context) { c.String(http.StatusOK, "start") })
	router.DELETE("/api/v1/streaming/jobs/:id", middleware.RequirePermission(rbac, security.ResourceStreaming, security.ActionDeleteJob), func(c *gin.Context) { c.String(http.StatusOK, "deleted") })

	// Allowed: analytics forecast
	req1 := httptest.NewRequest(http.MethodPost, "/api/v1/analytics/forecast", nil)
	req1.Header.Set("X-User-Roles", "analyst")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	if w1.Code != http.StatusOK || strings.TrimSpace(w1.Body.String()) != "forecast" {
		t.Fatalf("expected forecast 200, got %d %s", w1.Code, w1.Body.String())
	}

	// Denied: analytics train_model (no permission, normal mode)
	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/analytics/train", nil)
	req2.Header.Set("X-User-Roles", "analyst")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	if w2.Code != http.StatusForbidden {
		t.Fatalf("expected 403 train model, got %d", w2.Code)
	}

	// Allowed: streaming create
	req3 := httptest.NewRequest(http.MethodPost, "/api/v1/streaming/jobs", nil)
	req3.Header.Set("X-User-Roles", "analyst")
	w3 := httptest.NewRecorder()
	router.ServeHTTP(w3, req3)
	if w3.Code != http.StatusOK {
		t.Fatalf("expected 200 create job, got %d", w3.Code)
	}

	// Denied: streaming delete (no permission), but enable audit mode to soft allow
	middleware.SetAuditModeForTests(true)
	defer middleware.SetAuditModeForTests(false)
	req4 := httptest.NewRequest(http.MethodDelete, "/api/v1/streaming/jobs/123", nil)
	req4.Header.Set("X-User-Roles", "analyst")
	w4 := httptest.NewRecorder()
	router.ServeHTTP(w4, req4)
	if w4.Code == http.StatusForbidden {
		t.Fatalf("audit mode delete_job should not hard deny")
	}
	if h := w4.Header().Get("X-RBAC-Audit"); h != "deny" {
		t.Fatalf("expected audit deny header, got %s", h)
	}

	// Metrics scrape to assert audit soft deny metric increment for delete_job
	metricsReq := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	metricsW := httptest.NewRecorder()
	router.ServeHTTP(metricsW, metricsReq)
	body := metricsW.Body.String()
	if !strings.Contains(body, "costscope_rbac_audit_soft_denies_total{action=\"delete_job\",resource=\"streaming\"}") {
		t.Fatalf("expected audit soft deny metric for streaming delete_job, metrics: %s", body)
	}
}
