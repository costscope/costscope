package api

import (
	"net/http"
	"time"

	"local/costscope/internal/core/logging"
)

// analyticsAnalyzeHandler - placeholder async analyze job submission (mirrors other job style)
func analyticsAnalyzeHandler(logger *logging.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		id := "analyze-" + time.Now().Format("20060102-150405")
		resp := `{"job_id":"` + id + `","status":"pending","type":"analytics_analyze","created_at":"` + time.Now().Format(time.RFC3339) + `","websocket_url":"ws://` + r.Host + `/ws/jobs/` + id + `"}`
		_, _ = w.Write([]byte(resp))
	})
}
