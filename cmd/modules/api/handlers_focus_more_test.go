package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/costscope/costscope/internal/core/logging"
)

func TestFocusValidateHandler_ReturnsJSON(t *testing.T) {
	logger := logging.NewLogger(logging.LevelInfo)
	h := focusValidateHandler(logger)
	req := httptest.NewRequest(http.MethodGet, "/focus/validate", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != expectedJSONContentType {
		t.Fatalf("unexpected content-type: %q", ct)
	}
	if !strings.Contains(w.Body.String(), "is_compliant") {
		t.Fatalf("body missing expected field: %q", w.Body.String())
	}
}

func TestFocusDatasetsHandler_ReturnsJSON(t *testing.T) {
	logger := logging.NewLogger(logging.LevelInfo)
	h := focusDatasetsHandler(logger)
	req := httptest.NewRequest(http.MethodGet, "/focus/datasets", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != expectedJSONContentType {
		t.Fatalf("unexpected content-type: %q", ct)
	}
	if !strings.Contains(w.Body.String(), "datasets") {
		t.Fatalf("body missing datasets: %q", w.Body.String())
	}
}

func TestFocusJobStatusHandler_Success(t *testing.T) {
	logger := logging.NewLogger(logging.LevelInfo)
	h := focusJobStatusHandler(logger)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/focus/jobs/job-42", nil)
	// set Host header so websocket_url builds a host
	req.Host = "example.local"
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "job_id") || !strings.Contains(w.Body.String(), "websocket_url") {
		t.Fatalf("unexpected body: %q", w.Body.String())
	}
}

func TestFocusSchemasHandler_ReturnsSchemas(t *testing.T) {
	logger := logging.NewLogger(logging.LevelInfo)
	h := focusSchemasHandler(logger)
	req := httptest.NewRequest(http.MethodGet, "/focus/schemas", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "schemas") {
		t.Fatalf("expected schemas json, got %q", w.Body.String())
	}
}
