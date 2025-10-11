package api

import (
	"bytes"
	"net/http/httptest"
	"testing"

	"local/costscope/internal/core/logging"
)

const expectedJSONContentType = "application/json"

func TestAnalyticsForecastHandler_WritesAccepted(t *testing.T) {
	logger := logging.NewLogger(logging.LevelInfo)
	h := analyticsForecastHandler(logger)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/analytics/forecast", nil)
	h.ServeHTTP(rr, req)
	if rr.Code != 202 {
		t.Fatalf("expected 202 Accepted, got %d", rr.Code)
	}
}

func TestAnalyticsAnomaliesHandler_ReturnsJSON(t *testing.T) {
	logger := logging.NewLogger(logging.LevelInfo)
	h := analyticsAnomaliesHandler(logger)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/analytics/anomalies", nil)
	h.ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("expected 200 OK, got %d", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); ct != expectedJSONContentType {
		t.Fatalf("expected %s content type, got %q", expectedJSONContentType, ct)
	}
	if body := rr.Body.String(); body == "" || len(body) < 20 {
		t.Fatalf("unexpected empty or short body: %q", body)
	}
}

func TestAnalyticsOptimizeHandler_ReturnsJSON(t *testing.T) {
	logger := logging.NewLogger(logging.LevelInfo)
	h := analyticsOptimizeHandler(logger)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/analytics/optimize", nil)
	h.ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("expected 200 OK, got %d", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); ct != expectedJSONContentType {
		t.Fatalf("expected %s content type, got %q", expectedJSONContentType, ct)
	}
}

func TestAnalyticsMetricsHandler_ReturnsMetrics(t *testing.T) {
	logger := logging.NewLogger(logging.LevelInfo)
	h := analyticsMetricsHandler(logger)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/analytics/metrics", nil)
	h.ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("expected 200 OK, got %d", rr.Code)
	}
	if body := rr.Body.String(); body == "" || len(body) < 20 {
		t.Fatalf("unexpected empty or short body: %q", body)
	}
}

func TestAnalyticsAnalyzeHandler_AcceptsAndReturnsJobID(t *testing.T) {
	logger := logging.NewLogger(logging.LevelInfo)
	h := analyticsAnalyzeHandler(logger)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "http://example.local/analytics/analyze", nil)
	req.Host = "example.local"
	h.ServeHTTP(rr, req)
	if rr.Code != 202 {
		t.Fatalf("expected 202 Accepted, got %d", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); ct != expectedJSONContentType {
		t.Fatalf("expected %s content type, got %q", expectedJSONContentType, ct)
	}
	if body := rr.Body.String(); body == "" || len(body) < 20 {
		t.Fatalf("unexpected empty or short body: %q", body)
	}
}

func TestMulticloudRecommendationsHandler_AcceptsJSON(t *testing.T) {
	logger := logging.NewLogger(logging.LevelInfo)
	h := multicloudRecommendationsHandler(logger)
	rr := httptest.NewRecorder()
	body := `{}`
	req := httptest.NewRequest("POST", "/multicloud/recommendations", bytes.NewBufferString(body))
	h.ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("expected 200 OK, got %d", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); ct != expectedJSONContentType {
		t.Fatalf("expected %s content type, got %q", expectedJSONContentType, ct)
	}
}

func TestMulticloudMigrationPlanHandler_ParsesBody(t *testing.T) {
	logger := logging.NewLogger(logging.LevelInfo)
	h := multicloudMigrationPlanHandler(logger)
	rr := httptest.NewRecorder()
	body := `{"from":"aws","to":"gcp"}`
	req := httptest.NewRequest("POST", "/multicloud/migration/plan", bytes.NewBufferString(body))
	h.ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("expected 200 OK, got %d", rr.Code)
	}
}
