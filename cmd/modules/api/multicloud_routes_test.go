package api

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"local/costscope/internal/core/logging"
)

func TestMulticloudRecommendationsHandler_SuccessAndErrors(t *testing.T) {
	logger := logging.NewLogger(logging.LevelInfo)
	h := multicloudRecommendationsHandler(logger)

	// Missing body -> expect 400
	req0 := httptest.NewRequest(http.MethodPost, "/multicloud/recommendations", nil)
	rr0 := httptest.NewRecorder()
	h.ServeHTTP(rr0, req0)
	if rr0.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing body; got %d", rr0.Code)
	}

	// Invalid JSON -> expect 400
	req1 := httptest.NewRequest(http.MethodPost, "/multicloud/recommendations", strings.NewReader("{bad"))
	rr1 := httptest.NewRecorder()
	h.ServeHTTP(rr1, req1)
	if rr1.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid json; got %d", rr1.Code)
	}

	// Valid (empty) JSON -> expect 200 and recommendations field
	req2 := httptest.NewRequest(http.MethodPost, "/multicloud/recommendations", bytes.NewBufferString("{}"))
	rr2 := httptest.NewRecorder()
	h.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusOK {
		t.Fatalf("expected 200 for valid json; got %d", rr2.Code)
	}
	if ct := rr2.Header().Get("Content-Type"); ct != expectedJSONContentType {
		t.Fatalf("expected %s content type; got %q", expectedJSONContentType, ct)
	}
	if !strings.Contains(rr2.Body.String(), "recommendations") {
		t.Fatalf("expected recommendations key in body; got %q", rr2.Body.String())
	}
}

func TestMulticloudMigrationPlanHandler_SuccessAndErrors(t *testing.T) {
	logger := logging.NewLogger(logging.LevelInfo)
	h := multicloudMigrationPlanHandler(logger)

	// Missing body
	req0 := httptest.NewRequest(http.MethodPost, "/multicloud/migration/plan", nil)
	rr0 := httptest.NewRecorder()
	h.ServeHTTP(rr0, req0)
	if rr0.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing body; got %d", rr0.Code)
	}

	// Invalid JSON
	req1 := httptest.NewRequest(http.MethodPost, "/multicloud/migration/plan", strings.NewReader("[bad"))
	rr1 := httptest.NewRecorder()
	h.ServeHTTP(rr1, req1)
	if rr1.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid json; got %d", rr1.Code)
	}

	// Valid JSON
	req2 := httptest.NewRequest(http.MethodPost, "/multicloud/migration/plan", bytes.NewBufferString("{}"))
	rr2 := httptest.NewRecorder()
	h.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusOK {
		t.Fatalf("expected 200 for valid json; got %d", rr2.Code)
	}
	if !strings.Contains(rr2.Body.String(), "plan") {
		t.Fatalf("expected plan key in body; got %q", rr2.Body.String())
	}
}
