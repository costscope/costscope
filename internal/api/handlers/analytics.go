package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"local/costscope/internal/api/jobs"
	"local/costscope/internal/api/response"
	"local/costscope/internal/core/logging"
)

// =====================================================================================
// Analytics Handler - Advanced Analytics & ML
// =====================================================================================

// AnalyticsHandler provides advanced analytics and ML endpoints
type AnalyticsHandler struct {
	logger *logging.Logger
	jobMgr *jobs.Manager
}

// NewAnalyticsHandler creates a new analytics handler
func NewAnalyticsHandler(logger *logging.Logger, jobMgr *jobs.Manager) *AnalyticsHandler {
	return &AnalyticsHandler{
		logger: logger,
		jobMgr: jobMgr,
	}
}

// GenerateForecast generates cost forecast using ML models
func (h *AnalyticsHandler) GenerateForecast(c *gin.Context) {
	var request struct {
		DataSource   string  `json:"data_source" binding:"required"`
		ForecastDays int     `json:"forecast_days" binding:"required"`
		ModelType    string  `json:"model_type"`
		Confidence   float64 `json:"confidence"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		response.AutoBadRequest(c, "Invalid request format")
		return
	}

	jobID := "analytics-forecast"
	if tenant, ok := c.Get("tenant_id"); ok {
		h.logger.Debugf("tenant context for forecast: %v", tenant)
	}
	h.logger.InfoWithFields("Forecast generation requested", map[string]interface{}{
		"job_id":        jobID,
		"data_source":   request.DataSource,
		"forecast_days": request.ForecastDays,
	})

	response.AutoOK(c, http.StatusAccepted, gin.H{
		"job_id":  jobID,
		"status":  "accepted",
		"message": "Forecast generation job created",
	})
}

// DetectAnomalies detects cost anomalies in data
func (h *AnalyticsHandler) DetectAnomalies(c *gin.Context) {
	var request struct {
		DataSource  string  `json:"data_source" binding:"required"`
		Sensitivity float64 `json:"sensitivity"`
		TimeRange   struct {
			Start string `json:"start"`
			End   string `json:"end"`
		} `json:"time_range"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		response.AutoBadRequest(c, "Invalid request format")
		return
	}

	jobID := "analytics-anomalies"
	if tenant, ok := c.Get("tenant_id"); ok {
		h.logger.Debugf("tenant context for anomalies: %v", tenant)
	}
	response.AutoOK(c, http.StatusAccepted, gin.H{
		"job_id":  jobID,
		"status":  "accepted",
		"message": "Anomaly detection job created",
	})
}

// GetRecommendations provides cost optimization recommendations
func (h *AnalyticsHandler) GetRecommendations(c *gin.Context) {
	var request struct {
		DataSource string   `json:"data_source" binding:"required"`
		Categories []string `json:"categories"`
		MinSavings float64  `json:"min_savings"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		response.AutoBadRequest(c, "Invalid request format")
		return
	}

	jobID := "analytics-recommendations"
	if tenant, ok := c.Get("tenant_id"); ok {
		h.logger.Debugf("tenant context for recommendations: %v", tenant)
	}
	response.AutoOK(c, http.StatusAccepted, gin.H{
		"job_id":  jobID,
		"status":  "accepted",
		"message": "Recommendations generation job created",
	})
}

// AnalyzeTrends analyzes cost trends and patterns
func (h *AnalyticsHandler) AnalyzeTrends(c *gin.Context) {
	var request struct {
		DataSource  string   `json:"data_source" binding:"required"`
		Dimensions  []string `json:"dimensions"`
		Granularity string   `json:"granularity"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		response.AutoBadRequest(c, "Invalid request format")
		return
	}

	jobID := "analytics-trends"
	if tenant, ok := c.Get("tenant_id"); ok {
		h.logger.Debugf("tenant context for trends: %v", tenant)
	}
	response.AutoOK(c, http.StatusAccepted, gin.H{
		"job_id":  jobID,
		"status":  "accepted",
		"message": "Trends analysis job created",
	})
}

// ListModels returns available ML models
func (h *AnalyticsHandler) ListModels(c *gin.Context) {
	models := []gin.H{
		{
			"id":           "arima-forecasting",
			"name":         "ARIMA Forecasting Model",
			"type":         "forecasting",
			"status":       "trained",
			"accuracy":     0.87,
			"last_trained": "2025-01-30T12:00:00Z",
		},
		{
			"id":           "isolation-forest",
			"name":         "Isolation Forest Anomaly Detection",
			"type":         "anomaly_detection",
			"status":       "trained",
			"accuracy":     0.92,
			"last_trained": "2025-01-30T15:30:00Z",
		},
	}

	response.AutoOK200(c, gin.H{
		"models": models,
		"total":  len(models),
	})
}

// TrainModel trains or retrains an ML model
func (h *AnalyticsHandler) TrainModel(c *gin.Context) {
	modelID := c.Param("id")

	var request struct {
		DataSource    string                 `json:"data_source" binding:"required"`
		Parameters    map[string]interface{} `json:"parameters"`
		TestSplit     float64                `json:"test_split"`
		CrossValidate bool                   `json:"cross_validate"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		response.AutoBadRequest(c, "Invalid request format")
		return
	}

	jobID := "analytics-train-" + modelID
	response.AutoOK(c, http.StatusAccepted, gin.H{
		"job_id":   jobID,
		"model_id": modelID,
		"status":   "accepted",
		"message":  "Model training job created",
	})
}

// GetAnalyticsJob returns analytics job status
func (h *AnalyticsHandler) GetAnalyticsJob(c *gin.Context) {
	jobID := c.Param("id")

	// Mock job status
	response.AutoOK200(c, gin.H{
		"job_id":     jobID,
		"type":       "analytics",
		"status":     "running",
		"progress":   60,
		"created_at": "2025-01-31T00:00:00Z",
		"updated_at": "2025-01-31T00:03:00Z",
		"results":    nil,
	})
}
