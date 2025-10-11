package api

import (
	"net/http"
	"strings"

	"local/costscope/internal/api/jobs"
	wsman "local/costscope/internal/api/websocket"
	"local/costscope/internal/core/logging"
)

// sharedWSManager is assigned during server setup; falls back to a local manager in tests.
var sharedWSManager *wsman.Manager

// wsJobsHandler streams real-time job updates using the internal websocket Manager.
// It extracts jobID from /ws/jobs/{id} and delegates the upgrade/streaming to the manager.
func wsJobsHandler(logger *logging.Logger) http.Handler {
	const prefix = "/ws/jobs/"
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, prefix) || len(r.URL.Path) == len(prefix) {
			http.NotFound(w, r)
			return
		}
		jobID := r.URL.Path[len(prefix):]
		// If the job exists, we can later enrich; for now we don't block on presence
		if jm := (*jobs.Manager)(nil); jm != nil {
			_, _ = jm.GetJob(jobID)
		}
		mgr := sharedWSManager
		if mgr == nil { // fallback (unit tests without full server)
			mgr = wsman.NewManager(logger)
		}
		mgr.HandleConnectionHTTP(w, r, jobID)
	})
}
