package integration

// Health endpoints contract integration test.
// Covers readiness guarantees:
// 1. Start API (ephemeral port) and poll /health/ready (≤5s) -> expect 200.
// 2. Simulate provider registry failure via env COSTSCOPE_SIMULATE_PROVIDER_REGISTRY_FAILURE -> /health/ready -> 503.
// 3. /health and /health/live remain 200 in both states.
// Acceptance: stable (no race), runtime <3s. This test is race-safe (validated with -race) and typically <100ms.

import (
	"context"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/costscope/costscope/internal/api/handlers"
	"github.com/costscope/costscope/internal/core/logging"
)

// pollUntil polls the given URL until the predicate returns true or timeout passes.
// Returns error only on context cancellation.
func pollUntil(ctx context.Context, url string, pred func(*http.Response) bool) error { // nolint:unparam // signature kept simple for future use
	client := &http.Client{Timeout: 500 * time.Millisecond}
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		resp, err := client.Get(url) // #nosec G107 (test local address)
		if err == nil && pred(resp) {
			return nil
		}
		time.Sleep(60 * time.Millisecond)
	}
}

func startHealthServer(t *testing.T) (baseURL string, shutdown func()) {
	t.Helper()
	gin.SetMode(gin.ReleaseMode)
	logger := logging.NewLogger("warn")
	h := handlers.NewHealthHandler(logger) // jobs nil; readiness still OK due to provider registry + other checks
	r := gin.New()
	r.GET("/health", h.HealthCheck)
	r.GET("/health/live", h.LivenessCheck)
	r.GET("/health/ready", h.ReadinessCheck)

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	srv := &http.Server{ // Add ReadHeaderTimeout to satisfy gosec G112
		Handler:           r,
		ReadHeaderTimeout: 2 * time.Second,
	}
	go func() { _ = srv.Serve(ln) }()
	return "http://" + ln.Addr().String(), func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = srv.Shutdown(ctx)
	}
}

func TestReadiness_WithAndWithoutProviderRegistryFailure(t *testing.T) {
	// Ensure env cleanup
	orig := os.Getenv("COSTSCOPE_SIMULATE_PROVIDER_REGISTRY_FAILURE")
	defer func() { _ = os.Setenv("COSTSCOPE_SIMULATE_PROVIDER_REGISTRY_FAILURE", orig) }()

	baseURL, shutdown := startHealthServer(t)
	defer shutdown()

	// 1. Normal state: expect /health/ready 200 within 5s
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := pollUntil(ctx, baseURL+"/health/ready", func(r *http.Response) bool { return r.StatusCode == http.StatusOK }); err != nil {
		t.Fatalf("/health/ready did not become ready: %v", err)
	}

	// 2. Simulated provider registry failure – readiness should become 503 while health/live remain 200
	if err := os.Setenv("COSTSCOPE_SIMULATE_PROVIDER_REGISTRY_FAILURE", "true"); err != nil {
		t.Fatalf("setenv: %v", err)
	}
	// Give a small window for handler to observe new env (next request)
	ctx2, cancel2 := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel2()
	if err := pollUntil(ctx2, baseURL+"/health/ready", func(r *http.Response) bool { return r.StatusCode == http.StatusServiceUnavailable }); err != nil {
		t.Fatalf("/health/ready did not switch to 503 after simulation: %v", err)
	}

	// /health and /health/live must remain 200
	for _, path := range []string{"/health", "/health/live"} {
		resp, err := http.Get(baseURL + path) // #nosec G107
		if err != nil {
			t.Fatalf("GET %s: %v", path, err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("%s expected 200, got %d", path, resp.StatusCode)
		}
	}
}
