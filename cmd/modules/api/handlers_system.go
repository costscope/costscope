package api

import (
	"fmt"
	"net/http"
	"time"

	"local/costscope/internal/core/logging"
)

func healthHandler(logger *logging.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{"status":"healthy","version":"1.0.0","timestamp":"` + time.Now().Format(time.RFC3339) + `"}`)); err != nil {
			logger.Error(fmt.Sprintf("Failed to write health response: %s", err.Error()))
		}
	})
}

func infoHandler(logger *logging.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		info := `{
			"name": "CostScope FOCUS API",
			"version": "1.0.0",
			"description": "Enterprise API for FOCUS cost data operations",
			"endpoints": {
				"focus": "/api/v1/focus/*",
				"jobs": "/api/v1/focus/jobs/*",
				"websocket": "/ws/jobs/{jobId}",
				"health": "/health",
				"docs": "/docs"
			},
			"authentication": "JWT Bearer token required",
			"rate_limit": "` + fmt.Sprintf("%d requests per %s", rateLimitRequests, rateLimitWindow) + `"
		}`
		if _, err := w.Write([]byte(info)); err != nil {
			logger.Error(fmt.Sprintf("Failed to write info response: %s", err.Error()))
		}
	})
}

func docsHandler(logger *logging.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		docs := `<!DOCTYPE html>
<html>
<head>
	<title>CostScope FOCUS API Documentation</title>
	<style>
		body { font-family: Arial, sans-serif; margin: 40px; line-height: 1.6; }
		h1, h2, h3 { color: #333; }
		code { background: #f4f4f4; padding: 2px 4px; border-radius: 3px; }
		pre { background: #f4f4f4; padding: 10px; border-radius: 5px; overflow-x: auto; }
		.endpoint { margin: 20px 0; padding: 15px; border: 1px solid #ddd; border-radius: 5px; }
		.method { font-weight: bold; padding: 2px 8px; border-radius: 3px; }
		.post { background: #d4edda; color: #155724; }
		.get { background: #cce7ff; color: #004085; }
	</style>
</head>
<body>
	<h1>CostScope FOCUS API Documentation</h1>
	<p>Enterprise REST API for FOCUS cost data operations</p>

	<h2>Authentication</h2>
	<p>All API endpoints may require JWT. Include the token in the Authorization header:</p>
	<pre>Authorization: Bearer &lt;your-jwt-token&gt;</pre>

	<h2>FOCUS Operations</h2>
	<div class="endpoint">
		<h3><span class="method post">POST</span> /api/v1/focus/convert</h3>
		<p>Convert billing data to FOCUS format (async)</p>
	</div>
	<div class="endpoint">
		<h3><span class="method post">POST</span> /api/v1/focus/analyze</h3>
		<p>Analyze FOCUS dataset (async)</p>
	</div>
	<div class="endpoint">
		<h3><span class="method post">POST</span> /api/v1/focus/validate</h3>
		<p>Validate a FOCUS dataset against the specification</p>
	</div>
	<div class="endpoint">
		<h3><span class="method get">GET</span> /api/v1/focus/datasets</h3>
		<p>List available FOCUS datasets</p>
	</div>

	<h2>Health Check</h2>
	<div class="endpoint">
		<h3><span class="method get">GET</span> /health</h3>
		<p>Check API server health</p>
	</div>

	<h2>Docs</h2>
	<div class="endpoint">
		<h3><span class="method get">GET</span> /docs</h3>
		<p>API overview</p>
	</div>
</body>
</html>`
		if _, err := w.Write([]byte(docs)); err != nil {
			logger.Error("Failed to write response")
		}
	})
}
