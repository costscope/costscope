//go:build experimental

package focus

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/costscope/costscope/internal/api/jobs"
	"github.com/costscope/costscope/internal/api/websocket"
	"github.com/costscope/costscope/internal/core/logging"

	"github.com/gin-gonic/gin"
)

// setupTestHandler returns a handler with minimal dependencies for synchronous engine tests.
func setupTestHandler(t *testing.T) *Handler {
	t.Helper()
	logger := logging.NewLogger(logging.LevelError)
	jm := jobs.NewManager(logger, 0) // in-memory job manager with zero workers (not started for sync path)
	ws := websocket.NewManager(logger)
	return NewHandler(logger, jm, ws, nil, nil, nil, nil)
}

func TestAnalyzeAsync_SynchronousFocusEngine(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := setupTestHandler(t)
	r := gin.New()
	api := r.Group("/api/v1")
	h.RegisterRoutes(api)

	tru := true
	body := AnalyzeRequest{InputPath: "test.parquet", AnalysisType: "cost_breakdown", UseFocusEngine: &tru}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/focus/analyze", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK { // synchronous path should return 200
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp["engine"] != "focus" {
		t.Fatalf("expected engine=focus, got %v", resp["engine"])
	}
	if _, ok := resp["result"].(map[string]any); !ok {
		t.Fatalf("expected result object in response: %v", resp)
	}
}

func TestDiffAsync_SynchronousFocusEngine(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := setupTestHandler(t)
	r := gin.New()
	api := r.Group("/api/v1")
	h.RegisterRoutes(api)

	tru := true
	body := DiffRequest{BaselinePath: "baseline.parquet", CurrentPath: "current.parquet", UseFocusEngine: &tru}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/focus/diff", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK { // synchronous path should return 200
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp["engine"] != "focus" {
		t.Fatalf("expected engine=focus, got %v", resp["engine"])
	}
	if _, ok := resp["result"].(map[string]any); !ok {
		t.Fatalf("expected result object in response: %v", resp)
	}
}
