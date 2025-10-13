//go:build !duckdb

package handlers

import (
	"net/http"

	"github.com/costscope/costscope/internal/api/response"
	"github.com/costscope/costscope/internal/core/logging"

	"github.com/gin-gonic/gin"
)

// AnalyticsReadHandler exposes lightweight GET analytics when DuckDB is not enabled (fallback).
type AnalyticsReadHandler struct{}

func NewAnalyticsReadHandler(_ *logging.Logger) *AnalyticsReadHandler { return &AnalyticsReadHandler{} }

func (h *AnalyticsReadHandler) Summary(c *gin.Context) {
	response.AutoFail(c, http.StatusNotImplemented, "duckdb build required for analytics; rebuild with -tags duckdb", "not_implemented")
}
func (h *AnalyticsReadHandler) TopServices(c *gin.Context) {
	response.AutoFail(c, http.StatusNotImplemented, "duckdb build required for analytics; rebuild with -tags duckdb", "not_implemented")
}
func (h *AnalyticsReadHandler) Trends(c *gin.Context) {
	response.AutoFail(c, http.StatusNotImplemented, "duckdb build required for analytics; rebuild with -tags duckdb", "not_implemented")
}
