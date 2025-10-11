package api

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Test that multicloud handlers accept a JSON body and return 200 for valid payloads
func TestMulticloudHandlers_ValidAndInvalid(t *testing.T) {
	logger := testLogger()

	// Valid JSON payload
	body := bytes.NewBufferString(`{"example":1}`)
	req := httptest.NewRequest(http.MethodPost, "/", body)
	w := httptest.NewRecorder()
	h := multicloudRecommendationsHandler(logger)
	h.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for valid JSON, got %d body=%s", w.Code, w.Body.String())
	}

	// Invalid JSON
	req2 := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("notjson"))
	w2 := httptest.NewRecorder()
	h.ServeHTTP(w2, req2)
	if w2.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 BadRequest for invalid JSON, got %d body=%s", w2.Code, w2.Body.String())
	}

	// Missing body (nil) should be handled as bad request by handlers
	req3 := httptest.NewRequest(http.MethodPost, "/", nil)
	w3 := httptest.NewRecorder()
	h.ServeHTTP(w3, req3)
	if w3.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 BadRequest for missing body, got %d", w3.Code)
	}
}
