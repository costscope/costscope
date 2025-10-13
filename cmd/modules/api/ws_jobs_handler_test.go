package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/costscope/costscope/internal/core/logging"
)

func TestWSJobsHandler_NotFound(t *testing.T) {
	logger := logging.NewLogger(logging.LevelInfo)
	handler := wsJobsHandler(logger)

	// Request missing job id
	req := httptest.NewRequest(http.MethodGet, "/ws/jobs/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 got %d", w.Code)
	}
}

func TestWSJobsHandler_FallbackManager_NoPanic(t *testing.T) {
	logger := logging.NewLogger(logging.LevelInfo)
	handler := wsJobsHandler(logger)
	// Ensure sharedWSManager is nil to hit fallback path
	sharedWSManager = nil

	req := httptest.NewRequest(http.MethodGet, "/ws/jobs/job-123", nil)
	w := httptest.NewRecorder()
	// Should not panic and should return something (the fallback manager writes headers)
	handler.ServeHTTP(w, req)
	// We accept any non-500 response as success (ensures no panic)
	if w.Code == http.StatusInternalServerError {
		t.Fatalf("unexpected 500 from fallback handler")
	}
}
