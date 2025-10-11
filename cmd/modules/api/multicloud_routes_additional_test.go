package api

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"local/costscope/internal/core/logging"
)

func TestMulticloudRecommendationsHandler_EdgeAndHappy(t *testing.T) {
	logger := logging.NewLogger(logging.LevelInfo)
	h := multicloudRecommendationsHandler(logger)

	// missing body -> 400
	req := httptest.NewRequest(http.MethodPost, "/multicloud/recommendations", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing body got %d", rr.Code)
	}

	// invalid JSON -> 400
	req2 := httptest.NewRequest(http.MethodPost, "/multicloud/recommendations", bytes.NewBufferString("{invalid"))
	rr2 := httptest.NewRecorder()
	h.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid JSON got %d", rr2.Code)
	}

	// valid JSON -> 200 + content-type + non-empty body
	req3 := httptest.NewRequest(http.MethodPost, "/multicloud/recommendations", bytes.NewBufferString("{}"))
	rr3 := httptest.NewRecorder()
	h.ServeHTTP(rr3, req3)
	if rr3.Code != http.StatusOK {
		t.Fatalf("expected 200 for valid JSON got %d", rr3.Code)
	}
	if ct := rr3.Header().Get("Content-Type"); ct != expectedJSONContentType {
		t.Fatalf("expected %s content type, got %q", expectedJSONContentType, ct)
	}
	if rr3.Body.Len() == 0 {
		t.Fatalf("expected non-empty body for recommendations")
	}
}

func TestMulticloudMigrationPlanHandler_EdgeAndHappy(t *testing.T) {
	logger := logging.NewLogger(logging.LevelInfo)
	h := multicloudMigrationPlanHandler(logger)

	// missing body -> 400
	req := httptest.NewRequest(http.MethodPost, "/multicloud/migration-plan", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing body got %d", rr.Code)
	}

	// invalid JSON -> 400
	req2 := httptest.NewRequest(http.MethodPost, "/multicloud/migration-plan", bytes.NewBufferString("notjson"))
	rr2 := httptest.NewRecorder()
	h.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid JSON got %d", rr2.Code)
	}

	// valid JSON -> 200 + content-type + body
	req3 := httptest.NewRequest(http.MethodPost, "/multicloud/migration-plan", bytes.NewBufferString("{}"))
	rr3 := httptest.NewRecorder()
	h.ServeHTTP(rr3, req3)
	if rr3.Code != http.StatusOK {
		t.Fatalf("expected 200 for valid JSON got %d", rr3.Code)
	}
	if ct := rr3.Header().Get("Content-Type"); ct != expectedJSONContentType {
		t.Fatalf("expected %s content type, got %q", expectedJSONContentType, ct)
	}
	if rr3.Body.Len() == 0 {
		t.Fatalf("expected non-empty body for migration plan")
	}
}
