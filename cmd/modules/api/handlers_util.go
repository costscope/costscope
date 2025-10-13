package api

import (
	"net/http"
	"time"

	"github.com/costscope/costscope/internal/core/logging"
)

// writeAcceptedJob writes a standard asynchronous job acceptance JSON payload.
// Shape: {"job_id":"<id>","status":"pending","type":"<type>","created_at":"<ts>","websocket_url":"ws://<host>/ws/jobs/<id>"}
func writeAcceptedJob(w http.ResponseWriter, r *http.Request, logger *logging.Logger, id, jobType string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	payload := `{"job_id":"` + id + `","status":"pending","type":"` + jobType + `","created_at":"` + time.Now().Format(time.RFC3339) + `","websocket_url":"ws://` + r.Host + `/ws/jobs/` + id + `"}`
	if _, err := w.Write([]byte(payload)); err != nil {
		logger.Error("Failed to write job response")
	}
}
