package api

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"local/costscope/internal/core/logging"
)

func TestRateLimitMiddleware_MaxZero_AllowsNext(t *testing.T) {
	mw := rateLimitMiddleware(0, time.Minute)
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})
	h := mw(next)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	// ensure RemoteAddr present
	req.RemoteAddr = "127.0.0.1:12345"
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if !called {
		t.Fatalf("expected next to be called when max=0")
	}
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", rr.Code)
	}
}

func TestRateLimitMiddleware_ExceedsLimit(t *testing.T) {
	// small window and max=1 to trigger quickly
	mw := rateLimitMiddleware(1, time.Minute)
	called := 0
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called++
		w.WriteHeader(http.StatusOK)
	})
	h := mw(next)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "127.0.0.1:54321"

	rr1 := httptest.NewRecorder()
	h.ServeHTTP(rr1, req)
	if rr1.Code != http.StatusOK {
		t.Fatalf("expected first request 200, got %d", rr1.Code)
	}

	// immediate second request should be rate-limited
	rr2 := httptest.NewRecorder()
	h.ServeHTTP(rr2, req)
	if rr2.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 on second request, got %d", rr2.Code)
	}
	if rr2.Header().Get("X-RateLimit-Limit") == "" {
		t.Fatalf("expected rate limit headers to be present")
	}
	if called != 1 {
		t.Fatalf("expected next called once, got %d", called)
	}
}

func TestRegisterRouteSpecs_GateDisabled_SkipsRoute(t *testing.T) {
	prev := os.Getenv("MY_TEST_GATE")
	if err := os.Setenv("MY_TEST_GATE", "0"); err != nil {
		t.Fatalf("failed to set env: %v", err)
	}
	defer func() { _ = os.Setenv("MY_TEST_GATE", prev) }()

	logger := logging.NewLogger(logging.LevelInfo)
	mux := http.NewServeMux()
	specs := []RouteSpec{
		{Method: http.MethodGet, Path: "/skip-me", Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}), FeatureGate: "MY_TEST_GATE"},
	}
	registerRouteSpecs(mux, specs, nil, logger)

	// request to skipped route should return 404
	req := httptest.NewRequest(http.MethodGet, "/skip-me", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for skipped route, got %d", rr.Code)
	}
}
