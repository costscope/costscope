package api

import (
	"encoding/json"
	"net/http"

	"github.com/costscope/costscope/internal/core/logging"
)

// Basic (non-enterprise) multicloud HTTP handlers.
// These are lightweight stubs that expose the same route surface exercised in enterprise mode
// so that the generic API server tests can validate presence and response shape.
// When the full multicloud service wiring is needed in basic mode these can be replaced
// with thin adapters into the internal service layer while preserving route contract.

func multicloudRecommendationsHandler(logger *logging.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Minimal input validation: require JSON body (can be empty object)
		if r.Body == nil {
			http.Error(w, "missing body", http.StatusBadRequest)
			return
		}
		// Drain body to ensure tests sending payload are accepted; reject malformed JSON
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}
		writeJSONMulti(w, map[string]any{
			"recommendations": []map[string]any{
				{"id": "rec-1", "summary": "stub recommendation", "savings": 0.0},
			},
		})
	})
}

func multicloudInventoryHandler(logger *logging.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSONMulti(w, map[string]any{
			"inventory": map[string]any{
				"providers": []string{"aws", "azure", "gcp"},
				"resources": 0,
			},
		})
	})
}

func multicloudMigrationPlanHandler(logger *logging.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Body == nil {
			http.Error(w, "missing body", http.StatusBadRequest)
			return
		}
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}
		writeJSONMulti(w, map[string]any{
			"plan": map[string]any{
				"steps":  []map[string]string{{"action": "analyze", "status": "pending"}},
				"status": "draft",
			},
		})
	})
}

func multicloudMigrationFeasibilityHandler(logger *logging.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Body == nil {
			http.Error(w, "missing body", http.StatusBadRequest)
			return
		}
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}
		writeJSONMulti(w, map[string]any{
			"feasibility": map[string]any{
				"score": 0.0,
				"risk":  "unknown",
			},
		})
	})
}

// writeJSON is a tiny helper local to these preview handlers to serialize JSON responses.
func writeJSONMulti(w http.ResponseWriter, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(payload)
}
