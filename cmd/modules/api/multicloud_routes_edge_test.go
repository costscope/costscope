package api

import (
	"bytes"
	"io"
	"net/http/httptest"
	"testing"

	"local/costscope/internal/core/logging"
)

func TestMulticloudHandlers_MissingOrInvalidBody(t *testing.T) {
	logger := logging.NewLogger(logging.LevelInfo)

	// Missing body for recommendations
	hRec := multicloudRecommendationsHandler(logger)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/recommendations", nil)
	hRec.ServeHTTP(w, req)
	if w.Code != 400 {
		t.Fatalf("expected 400 for missing body recommendations; got %d", w.Code)
	}

	// Invalid JSON for recommendations
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest("POST", "/recommendations", bytes.NewBufferString("not-json"))
	hRec.ServeHTTP(w2, req2)
	if w2.Code != 400 {
		t.Fatalf("expected 400 for invalid json recommendations; got %d", w2.Code)
	}

	// Missing body for migration plan
	hPlan := multicloudMigrationPlanHandler(logger)
	w3 := httptest.NewRecorder()
	req3 := httptest.NewRequest("POST", "/migration/plan", nil)
	hPlan.ServeHTTP(w3, req3)
	if w3.Code != 400 {
		t.Fatalf("expected 400 for missing body migration plan; got %d", w3.Code)
	}

	// Invalid JSON for migration plan
	w4 := httptest.NewRecorder()
	req4 := httptest.NewRequest("POST", "/migration/plan", bytes.NewBufferString("{bad:json}"))
	hPlan.ServeHTTP(w4, req4)
	if w4.Code != 400 {
		t.Fatalf("expected 400 for invalid json migration plan; got %d", w4.Code)
	}

	// Valid payload path should return 200 and JSON body
	w5 := httptest.NewRecorder()
	req5 := httptest.NewRequest("POST", "/migration/plan", bytes.NewBufferString("{}"))
	hPlan.ServeHTTP(w5, req5)
	if w5.Code != 200 {
		t.Fatalf("expected 200 for valid migration plan; got %d", w5.Code)
	}
	body, _ := io.ReadAll(w5.Body)
	if len(body) == 0 {
		t.Fatalf("expected non-empty body for valid migration plan")
	}
}
