package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/costscope/costscope/internal/core/logging"
)

func TestRoutesSummaryHandler_ReturnsJSON_New(t *testing.T) {
	logger := logging.NewLogger(logging.LevelInfo)
	h := routesSummaryHandler(logger)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/routes", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK; got %d", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); ct != expectedJSONContentType {
		t.Fatalf("expected %s content type, got %q", expectedJSONContentType, ct)
	}
	if !strings.Contains(rr.Body.String(), "routes") {
		t.Fatalf("expected routes key in body; got %q", rr.Body.String())
	}
}

func TestCORSMiddleware_AllowsOriginAndHandlesOptions(t *testing.T) {
	prev := corsOrigins
	defer func() { corsOrigins = prev }()

	// Set a specific allowed origin
	corsOrigins = []string{"https://example.com"}

	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	// Matching Origin should be echoed
	h := corsMiddleware(next)
	req := httptest.NewRequest(http.MethodGet, "/foo", nil)
	req.Header.Set("Origin", "https://example.com")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Header().Get("Access-Control-Allow-Origin") != "https://example.com" {
		t.Fatalf("expected Allow-Origin to be echoed, got %q", rr.Header().Get("Access-Control-Allow-Origin"))
	}
	if !called {
		t.Fatalf("expected next handler to be called for non-OPTIONS request")
	}

	// OPTIONS preflight should not call next and return 204
	called = false
	req2 := httptest.NewRequest(http.MethodOptions, "/foo", nil)
	req2.Header.Set("Origin", "https://example.com")
	rr2 := httptest.NewRecorder()
	h.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusNoContent {
		t.Fatalf("expected 204 for OPTIONS preflight; got %d", rr2.Code)
	}
	if called {
		t.Fatalf("expected next not to be called for OPTIONS")
	}
}
