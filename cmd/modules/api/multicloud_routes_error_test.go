package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"local/costscope/internal/core/logging"
)

func TestMulticloudMigrationFeasibilityHandler_InvalidJSON(t *testing.T) {
	logger := logging.NewLogger(logging.LevelInfo)
	handler := multicloudMigrationFeasibilityHandler(logger)

	// malformed JSON
	req := httptest.NewRequest(http.MethodPost, "/multicloud/feasibility", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing body got %d", w.Code)
	}
}
