package api

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMulticloudHandlers_MissingBodyAndInvalidJSON(t *testing.T) {
	// use the net/http handlers directly
	hRec := httptest.NewRecorder()

	// Missing body -> should return 400
	req1 := httptest.NewRequest(http.MethodPost, "/", nil)
	multicloudRecommendationsHandler(nil).ServeHTTP(hRec, req1)
	if hRec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing body, got %d", hRec.Code)
	}

	// Invalid JSON -> 400
	hRec2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte("not-json")))
	multicloudMigrationPlanHandler(nil).ServeHTTP(hRec2, req2)
	if hRec2.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid json, got %d", hRec2.Code)
	}
}
