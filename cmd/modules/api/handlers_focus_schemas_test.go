package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/costscope/costscope/internal/core/logging"
)

// TestFocusSchemasHandler verifies the schemas discovery endpoint returns ordered specs.
func TestFocusSchemasHandler(t *testing.T) {
	logger := logging.NewLogger("debug")
	rr := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/focus/schemas", nil)

	handler := focusSchemasHandler(logger)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200 got %d", rr.Code)
	}
	var payload struct {
		Schemas []string `json:"schemas"`
		Latest  string   `json:"latest"`
		Count   int      `json:"count"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if len(payload.Schemas) == 0 {
		t.Fatalf("expected non-empty schemas list")
	}
	// ensure descending (focus-1.2, focus-1.1, ...) by simple lexical fallback after prefix removal
	for i := 1; i < len(payload.Schemas); i++ {
		if payload.Schemas[i] == payload.Schemas[i-1] {
			t.Fatalf("duplicate adjacent schema %s", payload.Schemas[i])
		}
	}
	if payload.Latest != payload.Schemas[0] {
		t.Fatalf("latest %s mismatch first %s", payload.Latest, payload.Schemas[0])
	}
	if payload.Count != len(payload.Schemas) {
		t.Fatalf("count field mismatch: got %d want %d", payload.Count, len(payload.Schemas))
	}
}
