package api

import (
	"encoding/json"
	"net/http"

	"local/costscope/internal/core/logging"
)

// routesSummaryHandler returns a lightweight JSON summary of all declared routes.
// It reuses BuildRouteSpecs to ensure the summary reflects the declarative source of truth.
func routesSummaryHandler(logger *logging.Logger) http.Handler {
	type route struct {
		Method string   `json:"method"`
		Path   string   `json:"path"`
		Tags   []string `json:"tags,omitempty"`
		Gate   string   `json:"feature_gate,omitempty"`
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		specs := BuildRouteSpecs(logger)
		out := make([]route, 0, len(specs))
		for _, s := range specs {
			out = append(out, route{Method: s.Method, Path: s.Path, Tags: s.Tags, Gate: s.FeatureGate})
		}
		w.Header().Set("Content-Type", "application/json")
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		if err := enc.Encode(map[string]interface{}{"routes": out}); err != nil {
			http.Error(w, "failed to encode", http.StatusInternalServerError)
			logger.Error("routes summary encode failed: " + err.Error())
			return
		}
	})
}
