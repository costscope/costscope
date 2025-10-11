package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"local/costscope/internal/core/monitoring/telemetry"
)

// TestEnterpriseRouterMetricsEndpoint verifies that the enterprise Gin router exposes /metrics
// and that our key metric series are present. This mirrors production wiring (newEnterpriseGinRouter).
func TestEnterpriseRouterMetricsEndpoint(t *testing.T) {
	// Register Prometheus metrics; guard against double registration in other tests.
	defer func() { _ = recover() }()
	telemetry.Register()

	router := newEnterpriseGinRouter()

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 from /metrics, got %d", rr.Code)
	}

	body := rr.Body.String()

	// Presence checks for a small, stable set of core metrics. Avoid strict names that
	// may vary across instrumentation changes to reduce flakiness.
	required := []string{
		"costscope_conversion_active_jobs",
		"costscope_health_readiness",
		"costscope_unified_mapper_rows_total",
	}

	for _, name := range required {
		present := strings.Contains(body, name+"_bucket") || strings.Contains(body, name+" ") || strings.Contains(body, name+"{")
		if !present {
			t.Logf("/metrics body (truncated 2k):\n%s", truncate(body, 2048))
			t.Fatalf("expected metric %s to be exposed by /metrics", name)
		}
	}
}

// truncate returns s limited to max bytes (roughly) for logging convenience.
func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "...<truncated>"
}
