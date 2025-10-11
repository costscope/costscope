package api

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"local/costscope/internal/core/logging"
)

func TestRoutesSummaryHandler_ReturnsJSON(t *testing.T) {
	logger := logging.NewLogger(logging.LevelInfo)
	h := routesSummaryHandler(logger)
	req := httptest.NewRequest(http.MethodGet, "/routes", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected application/json content type, got %q", ct)
	}
	if body := w.Body.String(); body == "" || len(body) < 20 {
		t.Fatalf("unexpected empty or short body: %q", body)
	}
}

func TestMulticloudRecommendationsHandler_BadRequests(t *testing.T) {
	logger := logging.NewLogger(logging.LevelInfo)
	h := multicloudRecommendationsHandler(logger)

	// missing body
	req := httptest.NewRequest(http.MethodPost, "/multicloud/recommendations", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing body got %d", w.Code)
	}

	// invalid JSON
	req2 := httptest.NewRequest(http.MethodPost, "/multicloud/recommendations", bytes.NewBufferString("{invalid"))
	w2 := httptest.NewRecorder()
	h.ServeHTTP(w2, req2)
	if w2.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid JSON got %d", w2.Code)
	}
}
