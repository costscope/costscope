//go:build !duckdb

package api

import (
	"net/http"
	"time"

	"github.com/costscope/costscope/internal/core/logging"
)

// analyticsBreakdownHandler (slim): static illustrative payload
func analyticsBreakdownHandler(logger *logging.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := `{
			"summary": {
				"total_cost": 156892.47,
				"currency": "USD",
				"period": {"start": "2025-01-01T00:00:00Z", "end": "2025-01-31T23:59:59Z"},
				"record_count": 123456
			},
			"top_services": [
				{"service_name":"EC2","provider":"aws","total_cost":68423.12,"currency":"USD","record_count":23456},
				{"service_name":"S3","provider":"aws","total_cost":31246.89,"currency":"USD","record_count":14567}
			],
			"generated_at": "` + time.Now().Format(time.RFC3339) + `"
		}`
		_, _ = w.Write([]byte(response))
	})
}

// analyticsSummaryHandler (slim): static payload
func analyticsSummaryHandler(logger *logging.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := `{
			"summary": {
				"total_cost": 156892.47,
				"currency": "USD",
				"period": {"start": "2025-01-01T00:00:00Z", "end": "2025-01-31T23:59:59Z"},
				"record_count": 123456
			},
			"generated_at": "` + time.Now().Format(time.RFC3339) + `"
		}`
		_, _ = w.Write([]byte(response))
	})
}

// analyticsTopServicesHandler (slim): static payload
func analyticsTopServicesHandler(logger *logging.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := `{
			"top_services": [
				{"service_name":"EC2","provider":"aws","total_cost":68423.12,"currency":"USD","record_count":23456},
				{"service_name":"S3","provider":"aws","total_cost":31246.89,"currency":"USD","record_count":14567}
			],
			"generated_at": "` + time.Now().Format(time.RFC3339) + `"
		}`
		_, _ = w.Write([]byte(response))
	})
}

// analyticsTrendsHandler (slim): static payload
func analyticsTrendsHandler(logger *logging.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := `{
			"granularity": "day",
			"trends": [
				{"period":"2025-01-01","effective_cost": 5200.12},
				{"period":"2025-01-02","effective_cost": 4890.45}
			],
			"generated_at": "` + time.Now().Format(time.RFC3339) + `"
		}`
		_, _ = w.Write([]byte(response))
	})
}
