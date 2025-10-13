//go:build duckdb

package handlers

import (
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/costscope/costscope/internal/api/response"
	"github.com/costscope/costscope/internal/core/logging"
	"github.com/costscope/costscope/internal/database"
	"github.com/costscope/costscope/internal/database/analytics"
	"github.com/costscope/costscope/internal/database/duckdb"

	"github.com/gin-gonic/gin"
)

// AnalyticsReadHandler exposes lightweight GET analytics backed by the facade and DuckDB.
type AnalyticsReadHandler struct {
	logger *logging.Logger
}

func NewAnalyticsReadHandler(logger *logging.Logger) *AnalyticsReadHandler {
	return &AnalyticsReadHandler{logger: logger}
}

// Summary: GET /api/v1/analytics/summary?input=path
func (h *AnalyticsReadHandler) Summary(c *gin.Context) {
	ctx := c.Request.Context()
	path := strings.TrimSpace(c.Query("input"))
	if path == "" {
		path = os.Getenv("COSTSCOPE_FOCUS_PARQUET")
	}
	if path == "" {
		response.AutoBadRequestCode(c, "missing input parquet path (query input= or COSTSCOPE_FOCUS_PARQUET)", response.ErrCodeMissingInput)
		return
	}
	eng, err := duckdb.NewDuckDBEngine(duckdb.DefaultConfig())
	if err != nil {
		response.AutoFail(c, http.StatusInternalServerError, "init duckdb: "+err.Error(), "duckdb_init")
		return
	}
	defer func() { _ = eng.Close() }()
	if err := eng.Connect(); err != nil {
		response.AutoFail(c, http.StatusInternalServerError, "connect duckdb: "+err.Error(), "duckdb_connect")
		return
	}
	if err := eng.Health(ctx); err != nil {
		response.AutoFail(c, http.StatusInternalServerError, "duckdb health: "+err.Error(), "duckdb_health")
		return
	}
	if err := eng.LoadFOCUSData(ctx, path); err != nil {
		response.AutoBadRequestCode(c, "load parquet: "+err.Error(), response.ErrCodeLoadParquet)
		return
	}
	fac := analytics.NewFacade(eng)
	sum, err := fac.CostSummary(ctx, nil)
	if err != nil {
		response.AutoFail(c, http.StatusInternalServerError, "summary: "+err.Error(), "summary_error")
		return
	}
	response.AutoOK200(c, gin.H{"summary": sum, "generated_at": time.Now().UTC().Format(time.RFC3339)})
}

// TopServices: GET /api/v1/analytics/top-services?input=path&limit=5
func (h *AnalyticsReadHandler) TopServices(c *gin.Context) {
	ctx := c.Request.Context()
	path := strings.TrimSpace(c.Query("input"))
	if path == "" {
		path = os.Getenv("COSTSCOPE_FOCUS_PARQUET")
	}
	if path == "" {
		response.AutoBadRequestCode(c, "missing input parquet path (query input= or COSTSCOPE_FOCUS_PARQUET)", response.ErrCodeMissingInput)
		return
	}
	limit := 5
	if s := strings.TrimSpace(c.Query("limit")); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 {
			limit = n
		}
	}
	eng, err := duckdb.NewDuckDBEngine(duckdb.DefaultConfig())
	if err != nil {
		response.AutoFail(c, http.StatusInternalServerError, "init duckdb: "+err.Error(), "duckdb_init")
		return
	}
	defer func() { _ = eng.Close() }()
	if err := eng.Connect(); err != nil {
		response.AutoFail(c, http.StatusInternalServerError, "connect duckdb: "+err.Error(), "duckdb_connect")
		return
	}
	if err := eng.Health(ctx); err != nil {
		response.AutoFail(c, http.StatusInternalServerError, "duckdb health: "+err.Error(), "duckdb_health")
		return
	}
	if err := eng.LoadFOCUSData(ctx, path); err != nil {
		response.AutoBadRequestCode(c, "load parquet: "+err.Error(), response.ErrCodeLoadParquet)
		return
	}
	fac := analytics.NewFacade(eng)
	items, err := fac.TopServices(ctx, nil, limit)
	if err != nil {
		response.AutoFail(c, http.StatusInternalServerError, "top services: "+err.Error(), "top_services_error")
		return
	}
	response.AutoOK200(c, gin.H{"top_services": items, "generated_at": time.Now().UTC().Format(time.RFC3339)})
}

// Trends: GET /api/v1/analytics/trends?input=path&granularity=day
func (h *AnalyticsReadHandler) Trends(c *gin.Context) {
	ctx := c.Request.Context()
	path := strings.TrimSpace(c.Query("input"))
	if path == "" {
		path = os.Getenv("COSTSCOPE_FOCUS_PARQUET")
	}
	if path == "" {
		response.AutoBadRequestCode(c, "missing input parquet path (query input= or COSTSCOPE_FOCUS_PARQUET)", response.ErrCodeMissingInput)
		return
	}
	granStr := strings.ToLower(strings.TrimSpace(c.Query("granularity")))
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
		response.AutoBadRequestCode(c, "invalid granularity; expected one of hour,day,week,month,year", response.ErrCodeInvalidGranularity)
		return
	}
	eng, err := duckdb.NewDuckDBEngine(duckdb.DefaultConfig())
	if err != nil {
		response.AutoFail(c, http.StatusInternalServerError, "init duckdb: "+err.Error(), "duckdb_init")
		return
	}
	defer func() { _ = eng.Close() }()
	if err := eng.Connect(); err != nil {
		response.AutoFail(c, http.StatusInternalServerError, "connect duckdb: "+err.Error(), "duckdb_connect")
		return
	}
	if err := eng.Health(ctx); err != nil {
		response.AutoFail(c, http.StatusInternalServerError, "duckdb health: "+err.Error(), "duckdb_health")
		return
	}
	if err := eng.LoadFOCUSData(ctx, path); err != nil {
		response.AutoBadRequestCode(c, "load parquet: "+err.Error(), response.ErrCodeLoadParquet)
		return
	}
	fac := analytics.NewFacade(eng)
	items, err := fac.CostTrends(ctx, nil, gran)
	if err != nil {
		response.AutoFail(c, http.StatusInternalServerError, "trends: "+err.Error(), "trends_error")
		return
	}
	response.AutoOK200(c, gin.H{"trends": items, "granularity": granStr, "generated_at": time.Now().UTC().Format(time.RFC3339)})
}
