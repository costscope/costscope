package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
)

func TestMulticloudHTTP_Inventory(t *testing.T) {
	srv := newTestServer()
	rr := doReq(t, srv, http.MethodGet, "/api/v1/multicloud/inventory", nil)
	if rr.Code != http.StatusOK {
		// Accept 404 only if route omission occurs unexpectedly (should not after registration)
		if rr.Code == http.StatusNotFound {
			t.Fatalf("multicloud inventory route not found (registration regression): %d", rr.Code)
		}
		t.Fatalf("inventory expected 200, got %d body=%s", rr.Code, rr.Body.String())
	}
	var body map[string]any
	_ = json.Unmarshal(rr.Body.Bytes(), &body)
	if _, ok := body["inventory"]; !ok {
		t.Fatalf("expected 'inventory' field in response")
	}
}

func TestMulticloudHTTP_Recommendations(t *testing.T) {
	srv := newTestServer()
	payload := []byte(`{"providers":["aws","gcp"],"savings_threshold":0.05,"max_recommendations":3}`)
	rr := doReq(t, srv, http.MethodPost, "/api/v1/multicloud/recommendations", bytes.NewReader(payload))
	if rr.Code != http.StatusOK {
		t.Fatalf("recommendations expected 200, got %d body=%s", rr.Code, rr.Body.String())
	}
	var body map[string]any
	_ = json.Unmarshal(rr.Body.Bytes(), &body)
	if _, ok := body["recommendations"]; !ok {
		t.Fatalf("expected 'recommendations' field in response")
	}
}

func TestMulticloudHTTP_MigrationPlanAndFeasibility(t *testing.T) {
	srv := newTestServer()
	planReq := []byte(`{"source":"aws","target":"gcp"}`)
	planRR := doReq(t, srv, http.MethodPost, "/api/v1/multicloud/migration/plan", bytes.NewReader(planReq))
	if planRR.Code != http.StatusOK {
		t.Fatalf("migration plan expected 200, got %d body=%s", planRR.Code, planRR.Body.String())
	}
	feasReq := []byte(`{"source":"aws","target":"gcp"}`)
	feasRR := doReq(t, srv, http.MethodPost, "/api/v1/multicloud/migration/feasibility", bytes.NewReader(feasReq))
	if feasRR.Code != http.StatusOK {
		t.Fatalf("migration feasibility expected 200, got %d body=%s", feasRR.Code, feasRR.Body.String())
	}
}
