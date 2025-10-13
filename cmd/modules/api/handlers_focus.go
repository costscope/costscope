package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/costscope/costscope/internal/core/focus/validation"
	"github.com/costscope/costscope/internal/core/logging"
)

// computeAPIInvariantsFromFile (duckdb/noduckdb variants) and the invariants route were removed
// alongside unused exported paths.

func focusConvertHandler(logger *logging.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeAcceptedJob(w, r, logger, "550e8400-e29b-41d4-a716-446655440000", "focus_convert")
	})
}
func focusAnalyzeHandler(logger *logging.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeAcceptedJob(w, r, logger, "550e8400-e29b-41d4-a716-446655440001", "focus_analyze")
	})
}

func focusValidateHandler(logger *logging.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := `{
			"is_compliant": true,
			"overall_score": 95.5,
			"spec_version": "1.2",
			"validation_timestamp": "` + time.Now().Format(time.RFC3339) + `",
			"summary": {
				"total_records": 100000,
				"valid_records": 99500,
				"invalid_records": 500,
				"error_rate": 0.5
			},
			"compliance_results": {
				"focus": {"passed": true, "score": 98.0},
				"schema": {"passed": true, "score": 100.0},
				"quality": {"passed": true, "score": 92.0}
			}
		}`
		if _, err := w.Write([]byte(response)); err != nil {
			logger.Error("Failed to write response")
		}
	})
}

func focusDatasetsHandler(logger *logging.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := `{
			"datasets": [
				{
					"path": "/data/focus/aws-2023-01.parquet",
					"name": "AWS January 2023",
					"provider": "aws",
					"spec_version": "1.2",
					"record_count": 1000000,
					"file_size_mb": 125.5,
					"date_range": {
						"start_date": "2023-01-01",
						"end_date": "2023-01-31"
					},
					"created_at": "2023-02-01T00:00:00Z",
					"last_modified": "2023-02-01T12:00:00Z",
					"tags": ["production", "monthly"]
				}
			],
			"total": 1,
			"limit": 50,
			"offset": 0,
			"has_more": false
		}`
		if _, err := w.Write([]byte(response)); err != nil {
			logger.Error("Failed to write response")
		}
	})
}

func focusJobStatusHandler(logger *logging.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jobID := r.URL.Path[len("/api/v1/focus/jobs/"):]
		if jobID == "" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := `{
			"job_id": "` + jobID + `",
			"status": "completed",
			"type": "focus_convert",
			"created_at": "` + time.Now().Add(-5*time.Minute).Format(time.RFC3339) + `",
			"started_at": "` + time.Now().Add(-4*time.Minute).Format(time.RFC3339) + `",
			"completed_at": "` + time.Now().Add(-1*time.Minute).Format(time.RFC3339) + `",
			"progress": {
				"current": 100,
				"total": 100,
				"percentage": 100.0,
				"message": "Conversion completed",
				"stage": "completed",
				"updated_at": "` + time.Now().Add(-1*time.Minute).Format(time.RFC3339) + `"
			},
			"result": {
				"records_processed": 100000,
				"file_size_mb": 25.5,
				"processing_time": "3m45s"
			},
			"websocket_url": "ws://` + r.Host + `/ws/jobs/` + jobID + `"
		}`
		if _, err := w.Write([]byte(response)); err != nil {
			logger.Error("Failed to write response")
		}
	})
}

// focusSchemasHandler returns the list of supported FOCUS schema versions in descending semantic version order.
// It reuses the validation engine's GetSupportedSpecs to ensure consistent ordering and future extensibility.
// Response shape (stable): { "schemas": ["focus-1.2","focus-1.1","focus-1.0"], "latest": "focus-1.2", "count": 3 }
func focusSchemasHandler(logger *logging.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		eng := validation.NewEngine()
		specs := eng.GetSupportedSpecs()
		latest := ""
		if len(specs) > 0 {
			latest = string(specs[0])
		}
		// simple JSON build (avoid importing extra libs beyond std)
		w.WriteHeader(http.StatusOK)
		// Build array
		// Pre-calculate size for minor perf (not critical)
		json := "{\n  \"schemas\": ["
		for i, s := range specs {
			if i > 0 {
				json += ","
			}
			json += "\"" + string(s) + "\""
		}
		json += "],\n  \"latest\": \"" + latest + "\",\n  \"count\": " + strconv.Itoa(len(specs)) + "\n}"
		if _, err := w.Write([]byte(json)); err != nil {
			logger.Error("Failed to write schemas response")
		}
	})
}
