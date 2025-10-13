//go:build duckdb

package api

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/costscope/costscope/internal/core/logging"
	"github.com/costscope/costscope/internal/database"
	"github.com/costscope/costscope/internal/database/analytics"
	"github.com/costscope/costscope/internal/database/duckdb"
)

// analyticsBreakdownHandler (duckdb): load a parquet path then compute summary + top 5 services
// Query params:
//   - input: absolute path to a FOCUS parquet file to read into DuckDB
//   - limit: optional, top N services (default 5)
func analyticsBreakdownHandler(logger *logging.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		path := r.URL.Query().Get("input")
		if path == "" {
			// Minimal shim: allow env var for convenience in dev/tests
			path = os.Getenv("COSTSCOPE_FOCUS_PARQUET")
		}
		if path == "" {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "missing input parquet path (query input= or COSTSCOPE_FOCUS_PARQUET)"})
			return
		}

		config := duckdb.DefaultConfig()
		engine, err := duckdb.NewDuckDBEngine(config)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "init duckdb: " + err.Error()})
			return
		}
		defer func() { _ = engine.Close() }()
		if err := engine.Connect(); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "connect duckdb: " + err.Error()})
			return
		}
		if err := engine.Health(ctx); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "duckdb health: " + err.Error()})
			return
		}
		if err := engine.LoadFOCUSData(ctx, path); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "load parquet: " + err.Error()})
			return
		}

		facade := analytics.NewFacade(engine)
		summary, err := facade.CostSummary(ctx, nil)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "summary: " + err.Error()})
			return
		}
		services, err := facade.TopServices(ctx, nil, 5)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "top services: " + err.Error()})
			return
		}

		resp := map[string]any{
			"summary":      summary,
			"top_services": services,
			"generated_at": time.Now().UTC().Format(time.RFC3339),
		}
		writeJSON(w, http.StatusOK, resp)
	})
}

// analyticsSummaryHandler (duckdb): compute cost summary for a parquet dataset
// Query params:
//   - input: absolute path to a FOCUS parquet file to read into DuckDB
func analyticsSummaryHandler(logger *logging.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		path := r.URL.Query().Get("input")
		if path == "" {
			path = os.Getenv("COSTSCOPE_FOCUS_PARQUET")
		}
		if path == "" {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "missing input parquet path (query input= or COSTSCOPE_FOCUS_PARQUET)"})
			return
		}
		config := duckdb.DefaultConfig()
		engine, err := duckdb.NewDuckDBEngine(config)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "init duckdb: " + err.Error()})
			return
		}
		defer func() { _ = engine.Close() }()
		if err := engine.Connect(); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "connect duckdb: " + err.Error()})
			return
		}
		if err := engine.Health(ctx); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "duckdb health: " + err.Error()})
			return
		}
		if err := engine.LoadFOCUSData(ctx, path); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "load parquet: " + err.Error()})
			return
		}
		facade := analytics.NewFacade(engine)
		summary, err := facade.CostSummary(ctx, nil)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "summary: " + err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"summary": summary, "generated_at": time.Now().UTC().Format(time.RFC3339)})
	})
}

// analyticsTopServicesHandler (duckdb): list top N services by cost
// Query params:
//   - input: parquet path
//   - limit: optional integer (default 5)
func analyticsTopServicesHandler(logger *logging.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		path := r.URL.Query().Get("input")
		if path == "" {
			path = os.Getenv("COSTSCOPE_FOCUS_PARQUET")
		}
		if path == "" {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "missing input parquet path (query input= or COSTSCOPE_FOCUS_PARQUET)"})
			return
		}
		limit := 5
		if s := strings.TrimSpace(r.URL.Query().Get("limit")); s != "" {
			if n, err := strconv.Atoi(s); err == nil && n > 0 {
				limit = n
			}
		}
		config := duckdb.DefaultConfig()
		engine, err := duckdb.NewDuckDBEngine(config)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "init duckdb: " + err.Error()})
			return
		}
		defer func() { _ = engine.Close() }()
		if err := engine.Connect(); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "connect duckdb: " + err.Error()})
			return
		}
		if err := engine.Health(ctx); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "duckdb health: " + err.Error()})
			return
		}
		if err := engine.LoadFOCUSData(ctx, path); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "load parquet: " + err.Error()})
			return
		}
		facade := analytics.NewFacade(engine)
		services, err := facade.TopServices(ctx, nil, limit)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "top services: " + err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"top_services": services, "generated_at": time.Now().UTC().Format(time.RFC3339)})
	})
}

// analyticsTrendsHandler (duckdb): time series of cost by period granularity
// Query params:
//   - input: parquet path
//   - granularity: hour|day|week|month|year (default day)
func analyticsTrendsHandler(logger *logging.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		path := r.URL.Query().Get("input")
		if path == "" {
			path = os.Getenv("COSTSCOPE_FOCUS_PARQUET")
		}
		if path == "" {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "missing input parquet path (query input= or COSTSCOPE_FOCUS_PARQUET)"})
			return
		}
		granStr := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("granularity")))
		if granStr == "" {
			granStr = "day"
		}
		var gran database.TimeGranularity
		switch granStr {
		case string(database.TimeGranularityHour):
			gran = database.TimeGranularityHour
		case string(database.TimeGranularityDay):
			gran = database.TimeGranularityDay
		case string(database.TimeGranularityWeek):
			gran = database.TimeGranularityWeek
		case string(database.TimeGranularityMonth):
			gran = database.TimeGranularityMonth
		case string(database.TimeGranularityYear):
			gran = database.TimeGranularityYear
		default:
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid granularity; expected one of hour,day,week,month,year"})
			return
		}
		config := duckdb.DefaultConfig()
		engine, err := duckdb.NewDuckDBEngine(config)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "init duckdb: " + err.Error()})
			return
		}
		defer func() { _ = engine.Close() }()
		if err := engine.Connect(); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "connect duckdb: " + err.Error()})
			return
		}
		if err := engine.Health(ctx); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "duckdb health: " + err.Error()})
			return
		}
		if err := engine.LoadFOCUSData(ctx, path); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "load parquet: " + err.Error()})
			return
		}
		facade := analytics.NewFacade(engine)
		trends, err := facade.CostTrends(ctx, nil, gran)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "trends: " + err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"trends": trends, "granularity": granStr, "generated_at": time.Now().UTC().Format(time.RFC3339)})
	})
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
