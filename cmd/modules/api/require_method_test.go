package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// Test requireMethod middleware allows correct method and rejects others
func TestRequireMethod_AllowsAndRejects(t *testing.T) {
	handler := requireMethod(http.MethodPost)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Correct method
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for POST, got %d", w.Code)
	}

	// Wrong method
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req2)
	if w2.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405 for GET, got %d", w2.Code)
	}
}
