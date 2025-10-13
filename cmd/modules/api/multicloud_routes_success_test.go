package api

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/costscope/costscope/internal/core/logging"
)

func TestMulticloudMigrationFeasibilityHandler_Success(t *testing.T) {
	logger := logging.NewLogger(logging.LevelInfo)
	handler := multicloudMigrationFeasibilityHandler(logger)

	// valid JSON body
	body := bytes.NewBufferString(`{"sources": [], "targets": []}`)
	req := httptest.NewRequest(http.MethodPost, "/multicloud/feasibility", body)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d body=%q", w.Code, w.Body.String())
	}
	if w.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("expected JSON content-type got %q", w.Header().Get("Content-Type"))
	}
}
