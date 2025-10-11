package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"local/costscope/internal/core/monitoring/telemetry"
	"local/costscope/internal/core/security"
)

const auditDenyHeaderValue = "deny"

// Test analytics & streaming endpoints permission enforcement and audit soft-deny metric.
func TestRequirePermission_AnalyticsAndStreaming_WithAuditMetrics(t *testing.T) {
	gin.SetMode(gin.TestMode)
	// Register metrics (guard duplicate panics)
	func() { defer func() { _ = recover() }(); telemetry.Register() }()

	// Role with partial permissions: allow analytics forecast & streaming create_job only.
	perms := []security.Permission{
		{Resource: security.ResourceAnalytics, Action: security.ActionForecast},
		{Resource: security.ResourceStreaming, Action: security.ActionCreateJob},
	}
	rbac := newTestRBAC(t, perms)

	r := gin.New()
	// Analytics endpoints
	r.GET("/analytics/forecast", RequirePermission(rbac, security.ResourceAnalytics, security.ActionForecast), func(c *gin.Context) { c.String(http.StatusOK, "forecast") })
	r.GET("/analytics/anomalies", RequirePermission(rbac, security.ResourceAnalytics, security.ActionDetectAnomalies), func(c *gin.Context) { c.String(http.StatusOK, "anomalies") })
	r.GET("/analytics/train", RequirePermission(rbac, security.ResourceAnalytics, security.ActionTrainModel), func(c *gin.Context) { c.String(http.StatusOK, "train") })
	// Streaming endpoints
	r.POST("/streaming/jobs", RequirePermission(rbac, security.ResourceStreaming, security.ActionCreateJob), func(c *gin.Context) { c.String(http.StatusOK, "create") })
	r.POST("/streaming/jobs/start", RequirePermission(rbac, security.ResourceStreaming, security.ActionStartJob), func(c *gin.Context) { c.String(http.StatusOK, "start") })

	// Allowed analytics forecast
	req1 := httptest.NewRequest(http.MethodGet, "/analytics/forecast", nil)
	req1.Header.Set("X-User-Roles", "analyst")
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)
	if w1.Code != http.StatusOK || w1.Body.String() != "forecast" {
		t.Fatalf("expected forecast 200, got %d %s", w1.Code, w1.Body.String())
	}

	// Denied anomalies (no permission)
	req2 := httptest.NewRequest(http.MethodGet, "/analytics/anomalies", nil)
	req2.Header.Set("X-User-Roles", "analyst")
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w2.Code)
	}

	// Allowed streaming create_job (still normal mode)
	req4 := httptest.NewRequest(http.MethodPost, "/streaming/jobs", nil)
	req4.Header.Set("X-User-Roles", "analyst")
	w4 := httptest.NewRecorder()
	r.ServeHTTP(w4, req4)
	if w4.Code != http.StatusOK {
		t.Fatalf("expected 200 create job, got %d", w4.Code)
	}

	// Denied streaming start_job (no permission) normal mode
	req5 := httptest.NewRequest(http.MethodPost, "/streaming/jobs/start", nil)
	req5.Header.Set("X-User-Roles", "analyst")
	w5 := httptest.NewRecorder()
	r.ServeHTTP(w5, req5)
	if w5.Code != http.StatusForbidden {
		t.Fatalf("expected 403 start job, got %d", w5.Code)
	}

	// Enable audit mode only for train_model soft deny scenario
	SetAuditModeForTests(true)
	defer SetAuditModeForTests(false)
	req3 := httptest.NewRequest(http.MethodGet, "/analytics/train", nil)
	req3.Header.Set("X-User-Roles", "analyst")
	w3 := httptest.NewRecorder()
	r.ServeHTTP(w3, req3)
	if w3.Code == http.StatusForbidden {
		t.Fatalf("audit mode should not block train model")
	}
	if h := w3.Header().Get("X-RBAC-Audit"); h != auditDenyHeaderValue {
		t.Fatalf("expected audit deny header, got %s", h)
	}

	// Verify audit metric was incremented at least once for train_model soft deny
	metricsReq := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	metricsW := httptest.NewRecorder()
	promhttp.Handler().ServeHTTP(metricsW, metricsReq)
	body := metricsW.Body.String()
	if !strings.Contains(body, "costscope_rbac_audit_soft_denies_total{action=\"train_model\",resource=\"analytics\"}") {
		t.Fatalf("expected audit soft deny metric for train_model, metrics: %s", body)
	}
}
