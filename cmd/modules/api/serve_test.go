package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"local/costscope/internal/api/jobs"
	"local/costscope/internal/api/websocket"
	"local/costscope/internal/core/logging"
)

func newTestServer() http.Handler {
	logger := logging.NewLogger(logging.LevelError)
	jm := jobs.NewManager(logger, 1)
	_ = jm.Start()
	ws := websocket.NewManager(logger)
	cm := NewMockConversionManager(logger)
	am := NewMockAnalysisManager(logger)
	cpm := NewMockComparisonManager(logger)
	vm := NewMockValidationManager(logger)
	return setupRouter(logger, jm, ws, cm, am, cpm, vm)
}

func doReq(t *testing.T, srv http.Handler, method, path string, body io.Reader) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, body)
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, req)
	return rr
}

func TestHealth(t *testing.T) {
	srv := newTestServer()
	rr := doReq(t, srv, http.MethodGet, "/health", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); !strings.Contains(ct, "application/json") {
		t.Fatalf("unexpected content-type: %s", ct)
	}
}

func TestDocs(t *testing.T) {
	srv := newTestServer()
	rr := doReq(t, srv, http.MethodGet, "/docs", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); !strings.Contains(ct, "text/html") {
		t.Fatalf("unexpected content-type: %s", ct)
	}
}

func TestInfo(t *testing.T) {
	srv := newTestServer()
	rr := doReq(t, srv, http.MethodGet, "/api/v1/info", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestFocusMethodEnforcement(t *testing.T) {
	srv := newTestServer()
	// POST endpoints should reject GET
	for _, p := range []string{"/api/v1/focus/convert", "/api/v1/focus/analyze", "/api/v1/focus/validate"} {
		rr := doReq(t, srv, http.MethodGet, p, nil)
		if rr.Code != http.StatusMethodNotAllowed {
			t.Fatalf("%s expected 405, got %d", p, rr.Code)
		}
		rr = doReq(t, srv, http.MethodPost, p, strings.NewReader("{}"))
		if rr.Code != http.StatusAccepted && p != "/api/v1/focus/validate" {
			t.Fatalf("%s expected 202, got %d", p, rr.Code)
		}
		if p == "/api/v1/focus/validate" && rr.Code != http.StatusOK {
			t.Fatalf("%s expected 200, got %d", p, rr.Code)
		}
	}
}

// Verify CORS preflight and headers are applied by the basic API server wiring
func TestServeRouter_CORS(t *testing.T) {
	srv := newTestServer()
	// Preflight on a POST endpoint
	req := httptest.NewRequest(http.MethodOptions, "/api/v1/focus/convert", nil)
	req.Header.Set("Origin", "https://example.com")
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, req)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("preflight expected 204, got %d", rr.Code)
	}
	if rr.Header().Get("Access-Control-Allow-Methods") == "" {
		t.Fatalf("missing Access-Control-Allow-Methods header")
	}
}

// Minimal rate limit smoke test for basic API server (per-IP limiter)
func TestServeRouter_RateLimit(t *testing.T) {
	// Override global limit knobs for this test
	oldMax, oldWin := rateLimitRequests, rateLimitWindow
	rateLimitRequests, rateLimitWindow = 1, time.Hour
	defer func() { rateLimitRequests, rateLimitWindow = oldMax, oldWin }()

	srv := newTestServer()
	// First request allowed
	req1 := httptest.NewRequest(http.MethodGet, "/api/v1/focus/datasets", nil)
	rr1 := httptest.NewRecorder()
	srv.ServeHTTP(rr1, req1)
	if rr1.Code != http.StatusOK {
		t.Fatalf("1st request expected 200, got %d", rr1.Code)
	}
	// Second immediately should be limited
	req2 := httptest.NewRequest(http.MethodGet, "/api/v1/focus/datasets", nil)
	rr2 := httptest.NewRecorder()
	srv.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusTooManyRequests {
		t.Fatalf("2nd request expected 429, got %d", rr2.Code)
	}
}

func TestFocusDatasetsAndJobStatus(t *testing.T) {
	srv := newTestServer()
	rr := doReq(t, srv, http.MethodGet, "/api/v1/focus/datasets", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("datasets expected 200, got %d", rr.Code)
	}

	// jobs prefix handler extracts id from path
	rr = doReq(t, srv, http.MethodGet, "/api/v1/focus/jobs/abc123", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("job status expected 200, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), `"job_id": "abc123"`) {
		t.Fatalf("job response missing id, body=%s", rr.Body.String())
	}
}

// ================= Additional tests for JWT secret resolution on serve command =================
func TestServeCommandMissingJWTSecret(t *testing.T) {
	// Ensure neither env nor flag provides secret
	t.Setenv("COSTSCOPE_TEST_MODE", "1")
	if err := os.Unsetenv("COSTSCOPE_JWT_SECRET"); err != nil {
		t.Fatalf("failed to unset COSTSCOPE_JWT_SECRET: %v", err)
	}
	cmd := BuildAPICommand()
	// Provide subcommand path (api serve)
	cmd.SetArgs([]string{"serve"})
	// Execute should error due to missing secret
	if err := cmd.Execute(); err == nil {
		t.Fatalf("expected error when jwt secret is not provided")
	}
}

func TestServeCommandWithJWTSecretFlag(t *testing.T) {
	t.Setenv("COSTSCOPE_TEST_MODE", "1")
	cmd := BuildAPICommand()
	cmd.SetArgs([]string{"serve", "--jwt-secret", "this-is-a-very-long-test-secret-value-1234567890"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected success when jwt secret flag provided, got error: %v", err)
	}
}
