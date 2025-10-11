package api

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"local/costscope/internal/core/logging"
)

func TestRouteSpec_CriticalEndpointsPresent(t *testing.T) {
	logger := logging.NewLogger(logging.LevelError)
	specs := BuildRouteSpecs(logger)
	found := make(map[string]bool)
	required := map[string]string{
		"GET /health":                    "health",
		"GET /metrics":                   "metrics",
		"POST /api/v1/focus/convert":     "focus convert",
		"POST /api/v1/analytics/analyze": "analytics analyze",
	}
	for _, s := range specs {
		key := s.Method + " " + s.Path
		if _, ok := required[key]; ok {
			found[key] = true
		}
	}
	for k := range required {
		if !found[k] {
			t.Fatalf("required endpoint missing: %s", k)
		}
	}
}

func TestRoutesSummaryEndpoint(t *testing.T) {
	srv := newTestServer()
	rr := doReq(t, srv, "GET", "/api/v1/routes", nil)
	if rr.Code != 200 {
		t.Fatalf("routes summary expected 200, got %d", rr.Code)
	}
	body := rr.Body.String()
	for _, expect := range []string{"\"method\": \"GET\"", "\"path\": \"/health\""} {
		if !strings.Contains(body, expect) {
			t.Fatalf("routes summary missing %s; body prefix=%s", expect, body[:min(120, len(body))])
		}
	}
}

func TestAnalyticsBreakdownEndpoint(t *testing.T) {
	maybeSetFocusParquetEnv()
	srv := newTestServer()
	rr := doReq(t, srv, "GET", "/api/v1/analytics/breakdown", nil)
	if rr.Code != http.StatusOK && rr.Code != http.StatusBadRequest { // duckdb build returns 400 when missing/invalid parquet
		t.Fatalf("breakdown expected 200 or 400, got %d", rr.Code)
	}
	if rr.Code == http.StatusOK { // only assert payload structure on 200
		body := rr.Body.String()
		for _, expect := range []string{"summary", "top_services"} {
			if !strings.Contains(body, expect) {
				t.Fatalf("breakdown response missing %s; body prefix=%s", expect, body[:min(160, len(body))])
			}
		}
	}
}

func TestAnalyticsSummaryTopServicesTrendsEndpoints(t *testing.T) {
	maybeSetFocusParquetEnv()
	srv := newTestServer()
	// Summary
	rr := doReq(t, srv, "GET", "/api/v1/analytics/summary", nil)
	if rr.Code != http.StatusOK && rr.Code != http.StatusBadRequest { // duckdb may 400 if missing parquet
		t.Fatalf("summary expected 200 or 400, got %d", rr.Code)
	}
	if rr.Code == http.StatusOK {
		body := rr.Body.String()
		if !strings.Contains(body, "summary") {
			t.Fatalf("summary response missing 'summary'; body prefix=%s", body[:min(160, len(body))])
		}
	}
	// Top services
	rr = doReq(t, srv, "GET", "/api/v1/analytics/top-services", nil)
	if rr.Code != http.StatusOK && rr.Code != http.StatusBadRequest { // 400 when dataset missing/invalid on duckdb
		t.Fatalf("top-services expected 200 or 400, got %d", rr.Code)
	}
	if rr.Code == http.StatusOK {
		body := rr.Body.String()
		if !strings.Contains(body, "top_services") {
			t.Fatalf("top-services response missing 'top_services'; body prefix=%s", body[:min(160, len(body))])
		}
	}
	// Trends
	rr = doReq(t, srv, "GET", "/api/v1/analytics/trends", nil)
	if rr.Code != http.StatusOK && rr.Code != http.StatusBadRequest { // 400 when dataset missing/invalid on duckdb
		t.Fatalf("trends expected 200 or 400, got %d", rr.Code)
	}
	if rr.Code == http.StatusOK {
		body := rr.Body.String()
		for _, expect := range []string{"trends", "generated_at"} {
			if !strings.Contains(body, expect) {
				t.Fatalf("trends response missing %s; body prefix=%s", expect, body[:min(160, len(body))])
			}
		}
	}
}

// maybeSetFocusParquetEnv sets COSTSCOPE_FOCUS_PARQUET to the demo file if present so duckdb handlers can return 200.
func maybeSetFocusParquetEnv() {
	// relative path from repo root
	candidates := []string{"demo/focus-conversion/demo-focus.parquet"}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			if abs, err := filepath.Abs(c); err == nil {
				_ = os.Setenv("COSTSCOPE_FOCUS_PARQUET", abs)
			}
			return
		}
	}
}
