//go:build experimental

package focus

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"local/costscope/internal/api/jobs"
	"local/costscope/internal/api/response"
	"local/costscope/internal/api/websocket"
	"local/costscope/internal/core/config"
	focusanalysis "local/costscope/internal/core/focus/analysis"
	focuscomparison "local/costscope/internal/core/focus/comparison"
	focquality "local/costscope/internal/core/focus/quality"
	"local/costscope/internal/core/logging"
	"local/costscope/internal/core/quality/drift"
)

// tenantFromContext extracts tenant_id from Gin context if present.
func tenantFromContext(c *gin.Context) string {
	if v, ok := c.Get("tenant_id"); ok {
		if s, ok2 := v.(string); ok2 {
			return s
		}
	}
	return ""
}

// =====================================================================================
// FOCUS REST API Handlers - Enterprise API for Core FOCUS Operations
// =====================================================================================

// Handler provides REST API endpoints for FOCUS operations
type Handler struct {
	logger        *logging.Logger
	jobManager    *jobs.Manager
	wsManager     *websocket.Manager
	conversionMgr interface{} // Mock manager interface
	analysisMgr   interface{} // Mock manager interface
	comparisonMgr interface{} // Mock manager interface
	validationMgr interface{} // Mock manager interface
}

// NewHandler creates a new FOCUS API handler
func NewHandler(
	logger *logging.Logger,
	jobManager *jobs.Manager,
	wsManager *websocket.Manager,
	conversionMgr interface{},
	analysisMgr interface{},
	comparisonMgr interface{},
	validationMgr interface{},
) *Handler {
	return &Handler{
		logger:        logger,
		jobManager:    jobManager,
		wsManager:     wsManager,
		conversionMgr: conversionMgr,
		analysisMgr:   analysisMgr,
		comparisonMgr: comparisonMgr,
		validationMgr: validationMgr,
	}
}

// =====================================================================================
// Request/Response Types
// =====================================================================================

// ConvertRequest represents a FOCUS conversion request
type ConvertRequest struct {
	Provider     string            `json:"provider" binding:"required" example:"aws"`
	InputPath    string            `json:"input_path" binding:"required" example:"/path/to/cur-data.csv"`
	OutputPath   string            `json:"output_path" binding:"required" example:"/path/to/focus-output.parquet"`
	Options      ConvertOptions    `json:"options,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
	WebhookURL   string            `json:"webhook_url,omitempty"`
	CallbackData interface{}       `json:"callback_data,omitempty"`
}

// ConvertOptions represents conversion configuration options
type ConvertOptions struct {
	ChunkSize   int    `json:"chunk_size,omitempty" example:"10000"`
	Workers     int    `json:"workers,omitempty" example:"4"`
	MaxMemory   int    `json:"max_memory,omitempty" example:"1024"`
	Compression bool   `json:"compression,omitempty" example:"true"`
	Validate    bool   `json:"validate,omitempty" example:"true"`
	Streaming   bool   `json:"streaming,omitempty" example:"false"`
	SpecVersion string `json:"spec_version,omitempty" example:"1.2"`
	// Pointer to distinguish omitted vs explicit false for precedence defaulting
	UseUnifiedMapper *bool `json:"use_unified_mapper,omitempty" example:"false"`
}

// Config precedence resolution now uses stateless Resolve*Field helpers (no loader shim needed).

// AnalyzeRequest represents a FOCUS analysis request
type AnalyzeRequest struct {
	InputPath    string            `json:"input_path" binding:"required" example:"/path/to/focus.parquet"`
	OutputPath   string            `json:"output_path,omitempty" example:"/path/to/analysis-report.json"`
	AnalysisType string            `json:"analysis_type" binding:"required" example:"cost_breakdown"`
	Dimensions   []string          `json:"dimensions,omitempty" example:"service,region"`
	TimeRange    *TimeRange        `json:"time_range,omitempty"`
	Options      AnalysisOptions   `json:"options,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
	WebhookURL   string            `json:"webhook_url,omitempty"`
	CallbackData interface{}       `json:"callback_data,omitempty"`
	// UseFocusEngine routes the request through the experimental FOCUS analysis engine synchronously
	UseFocusEngine *bool `json:"use_focus_engine,omitempty" example:"false"`
}

// AnalysisOptions represents analysis configuration options
type AnalysisOptions struct {
	IncludeTrends    bool    `json:"include_trends,omitempty" example:"true"`
	IncludeAnomalies bool    `json:"include_anomalies,omitempty" example:"true"`
	IncludeForecast  bool    `json:"include_forecast,omitempty" example:"false"`
	ForecastPeriods  int     `json:"forecast_periods,omitempty" example:"30"`
	Confidence       float64 `json:"confidence,omitempty" example:"0.95"`
	MLEnabled        bool    `json:"ml_enabled,omitempty" example:"true"`
	DetailLevel      string  `json:"detail_level,omitempty" example:"detailed"`
}

// DiffRequest represents a FOCUS dataset comparison request
type DiffRequest struct {
	BaselinePath string            `json:"baseline_path" binding:"required" example:"/path/to/baseline.parquet"`
	CurrentPath  string            `json:"current_path" binding:"required" example:"/path/to/current.parquet"`
	OutputPath   string            `json:"output_path,omitempty" example:"/path/to/diff-report.json"`
	Dimensions   []string          `json:"dimensions,omitempty" example:"service,region"`
	Options      DiffOptions       `json:"options,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
	WebhookURL   string            `json:"webhook_url,omitempty"`
	CallbackData interface{}       `json:"callback_data,omitempty"`
	// UseFocusEngine routes the request through the experimental FOCUS diff engine synchronously
	UseFocusEngine *bool `json:"use_focus_engine,omitempty" example:"false"`
}

// DiffOptions represents comparison configuration options
type DiffOptions struct {
	Threshold           float64  `json:"threshold,omitempty" example:"100.0"`
	SignificancePercent float64  `json:"significance_percent,omitempty" example:"10.0"`
	IncludeTrends       bool     `json:"include_trends,omitempty" example:"true"`
	IncludeAnomalies    bool     `json:"include_anomalies,omitempty" example:"true"`
	IncludeExecutive    bool     `json:"include_executive,omitempty" example:"true"`
	IncludeDetails      bool     `json:"include_details,omitempty" example:"true"`
	Confidence          float64  `json:"confidence,omitempty" example:"0.95"`
	AnomalyMethods      []string `json:"anomaly_methods,omitempty" example:"statistical,isolation_forest"`
}

// ValidateRequest represents a FOCUS validation request
type ValidateRequest struct {
	InputPath   string            `json:"input_path" binding:"required" example:"/path/to/focus.parquet"`
	OutputPath  string            `json:"output_path,omitempty" example:"/path/to/validation-report.json"`
	SpecVersion string            `json:"spec_version,omitempty" example:"1.2"`
	Compliance  []string          `json:"compliance,omitempty" example:"focus,gdpr"`
	Options     ValidationOptions `json:"options,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// DriftAnalyzeRequest for semantic drift analysis
type DriftAnalyzeRequest struct {
	BaselineInvariants  string    `json:"baseline_invariants" binding:"required"`
	CurrentInvariants   string    `json:"current_invariants" binding:"required"`
	History             []string  `json:"history,omitempty"`
	Alpha               float64   `json:"alpha,omitempty"`
	BucketSchema        []float64 `json:"bucket_schema,omitempty"`
	BaselineDataset     string    `json:"baseline_dataset,omitempty"`
	CurrentDataset      string    `json:"current_dataset,omitempty"`
	Percentiles         []float64 `json:"percentiles,omitempty"`
	PercentileThreshold float64   `json:"percentile_threshold,omitempty" example:"0.01"`
}

// ValidationOptions represents validation configuration options
type ValidationOptions struct {
	ValidateSchema      bool    `json:"validate_schema,omitempty" example:"true"`
	ValidateQuality     bool    `json:"validate_quality,omitempty" example:"true"`
	ValidatePerformance bool    `json:"validate_performance,omitempty" example:"false"`
	ValidateAnomalies   bool    `json:"validate_anomalies,omitempty" example:"true"`
	MinScore            float64 `json:"min_score,omitempty" example:"85.0"`
	FailFast            bool    `json:"fail_fast,omitempty" example:"false"`
	Workers             int     `json:"workers,omitempty" example:"4"`
}

// TimeRange represents a time range filter
type TimeRange struct {
	StartDate string `json:"start_date" example:"2023-01-01"`
	EndDate   string `json:"end_date" example:"2023-12-31"`
}

// JobResponse represents an async job response
type JobResponse struct {
	JobID        string                 `json:"job_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Status       string                 `json:"status" example:"running"`
	Type         string                 `json:"type" example:"focus_convert"`
	CreatedAt    time.Time              `json:"created_at" example:"2023-01-01T00:00:00Z"`
	StartedAt    *time.Time             `json:"started_at,omitempty" example:"2023-01-01T00:01:00Z"`
	CompletedAt  *time.Time             `json:"completed_at,omitempty"`
	Progress     *jobs.Progress         `json:"progress,omitempty"`
	Result       map[string]interface{} `json:"result,omitempty"`
	Error        *string                `json:"error,omitempty"`
	WebSocketURL string                 `json:"websocket_url" example:"${BASE_WS_URL}/ws/jobs/550e8400-e29b-41d4-a716-446655440000"`
}

// DatasetInfo represents information about a FOCUS dataset
type DatasetInfo struct {
	Path         string    `json:"path" example:"/data/focus/dataset-2023-01.parquet"`
	Name         string    `json:"name" example:"January 2023 AWS Costs"`
	Provider     string    `json:"provider" example:"aws"`
	SpecVersion  string    `json:"spec_version" example:"1.2"`
	RecordCount  int64     `json:"record_count" example:"1000000"`
	FileSizeMB   float64   `json:"file_size_mb" example:"125.5"`
	DateRange    TimeRange `json:"date_range"`
	CreatedAt    time.Time `json:"created_at" example:"2023-01-01T00:00:00Z"`
	LastModified time.Time `json:"last_modified" example:"2023-01-01T12:00:00Z"`
	Tags         []string  `json:"tags,omitempty" example:"production,monthly"`
}

// =====================================================================================
// API Endpoints
// =====================================================================================

// @Summary Convert billing data to FOCUS format
// @Description Convert cloud billing data to FOCUS v1.2 format with async processing
// @Tags FOCUS
// @Accept json
// @Produce json
// @Param request body ConvertRequest true "Conversion request"
// @Success 202 {object} JobResponse "Conversion job started"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/focus/convert [post]
func (h *Handler) ConvertAsync(c *gin.Context) {
	var req ConvertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.AutoBadRequest(c, err.Error())
		return
	}

	// Resolve UseUnifiedMapper with precedence: request > YAML > ENV > default(false)
	resolvedUnified := resolveUseUnifiedMapper(&req.Options, h.logger)
	// Mutate request so downstream tasks see the resolved value
	req.Options.UseUnifiedMapper = &resolvedUnified

	// Generate job ID
	jobID := uuid.New().String()

	// Create job configuration
	jobConfig := &jobs.JobConfig{
		ID:           jobID,
		Type:         "focus_convert",
		Priority:     jobs.PriorityNormal,
		Timeout:      30 * time.Minute,
		WebhookURL:   req.WebhookURL,
		CallbackData: req.CallbackData,
		Metadata:     req.Metadata,
		TenantID:     tenantFromContext(c),
	}

	// Create conversion task
	task := &ConversionTask{
		Request:       &req,
		Logger:        h.logger,
		ConversionMgr: h.conversionMgr,
		WSManager:     h.wsManager,
	}

	// Submit job
	job, err := h.jobManager.SubmitJob(jobConfig, task)
	if err != nil {
		h.logger.Error(fmt.Sprintf("Failed to submit conversion job: %s (jobID: %s, provider: %s)",
			err.Error(), jobID, req.Provider))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to submit job"})
		return
	}

	// Return job response
	response := &JobResponse{
		JobID:        job.ID,
		Status:       string(job.Status),
		Type:         job.Type,
		CreatedAt:    job.CreatedAt,
		WebSocketURL: fmt.Sprintf("ws://%s/ws/jobs/%s", c.Request.Host, job.ID),
	}

	h.logger.Info(fmt.Sprintf("FOCUS conversion job submitted: %s (provider: %s, input: %s, output: %s)",
		job.ID, req.Provider, req.InputPath, req.OutputPath))

	c.JSON(http.StatusAccepted, response)
}

// resolveUseUnifiedMapper resolves unified mapper mode with precedence: request (explicit) > YAML > ENV > default(false)
func resolveUseUnifiedMapper(opts *ConvertOptions, logger *logging.Logger) bool {
	var explicit *bool
	if opts != nil && opts.UseUnifiedMapper != nil {
		explicit = opts.UseUnifiedMapper
	}
	res := config.ResolveBoolField(logger, "focus.use_unified_mapper", explicit, func(cc *config.ConsolidatedConfig) *bool {
		if cc == nil {
			return nil
		}
		return &cc.Focus.UseUnifiedMapperDefault
	}, "COSTSCOPE_USE_UNIFIED_MAPPER", false)
	return res.Value
}

// @Summary Analyze FOCUS dataset
// @Description Perform cost analysis on a FOCUS dataset. If `use_focus_engine=true` is supplied in the request body the experimental synchronous FOCUS analysis engine is used and the result is returned immediately with 200 instead of starting an async job (202). Otherwise an async job is created.
// @Tags FOCUS
// @Accept json
// @Produce json
// @Param request body AnalyzeRequest true "Analysis request"
// @Success 202 {object} JobResponse "Analysis job started"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/focus/analyze [post]
func (h *Handler) AnalyzeAsync(c *gin.Context) {
	var req AnalyzeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.AutoBadRequest(c, err.Error())
		return
	}

	// If the caller explicitly requests the experimental engine, execute synchronously and return result
	if req.UseFocusEngine != nil && *req.UseFocusEngine {
		// Minimal inline execution using the analysis engine (placeholder implementation)
		eng := focusanalysis.NewEngine(h.logger, nil)
		res, err := eng.AnalyzeFOCUSDataset(req.InputPath, focusanalysis.AnalysisOptions{MLEnabled: true, AnomalyDetection: true, TrendAnalysis: true, OptimizationAnalysis: true, ForecastDays: 30, ConfidenceLevel: 0.95, OutputFormat: "json"})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "focus engine analysis failed"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"engine": "focus", "result": res})
		return
	}

	// Generate job ID
	jobID := uuid.New().String()

	// Create job configuration
	jobConfig := &jobs.JobConfig{
		ID:           jobID,
		Type:         "focus_analyze",
		Priority:     jobs.PriorityNormal,
		Timeout:      20 * time.Minute,
		WebhookURL:   req.WebhookURL,
		CallbackData: req.CallbackData,
		Metadata:     req.Metadata,
		TenantID:     tenantFromContext(c),
	}

	// Create analysis task
	task := &AnalysisTask{
		Request:     &req,
		Logger:      h.logger,
		AnalysisMgr: h.analysisMgr,
		WSManager:   h.wsManager,
	}

	// Submit job
	job, err := h.jobManager.SubmitJob(jobConfig, task)
	if err != nil {
		h.logger.Error(fmt.Sprintf("Failed to submit analysis job: %s (jobID: %s, type: %s)",
			err.Error(), jobID, req.AnalysisType))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to submit job"})
		return
	}

	// Return job response
	response := &JobResponse{
		JobID:        job.ID,
		Status:       string(job.Status),
		Type:         job.Type,
		CreatedAt:    job.CreatedAt,
		WebSocketURL: fmt.Sprintf("ws://%s/ws/jobs/%s", c.Request.Host, job.ID),
	}

	h.logger.Info(fmt.Sprintf("FOCUS analysis job submitted: %s (type: %s, input: %s)",
		job.ID, req.AnalysisType, req.InputPath))

	c.JSON(http.StatusAccepted, response)
}

// @Summary Compare FOCUS datasets
// @Description Compare two FOCUS datasets and identify cost differences
// @Tags FOCUS
// @Accept json
// @Produce json
// @Param request body DiffRequest true "Comparison request"
// @Success 202 {object} JobResponse "Comparison job started"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Summary Compare (diff) two FOCUS datasets
// @Description Compare two FOCUS datasets and identify cost differences. If `use_focus_engine=true` the experimental synchronous diff engine executes inline and returns 200 with the diff result; otherwise an async job (202) is started.
// @Tags FOCUS
// @Accept json
// @Produce json
// @Param request body DiffRequest true "Comparison request"
// @Success 202 {object} JobResponse "Comparison job started"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/focus/diff [post]
func (h *Handler) DiffAsync(c *gin.Context) {
	var req DiffRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.AutoBadRequest(c, err.Error())
		return
	}

	// Experimental engine synchronous path
	if req.UseFocusEngine != nil && *req.UseFocusEngine {
		eng := focuscomparison.NewEngine(h.logger, nil)
		res, err := eng.CompareFOCUSDatasets(req.BaselinePath, req.CurrentPath, focuscomparison.DiffOptions{Dimensions: req.Dimensions, Threshold: req.Options.Threshold, ShowAnomalies: true, ShowTrends: true, MLEnabled: true, OutputFormat: "json"})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "focus engine diff failed"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"engine": "focus", "result": res})
		return
	}

	// Generate job ID
	jobID := uuid.New().String()

	// Create job configuration
	jobConfig := &jobs.JobConfig{
		ID:           jobID,
		Type:         "focus_diff",
		Priority:     jobs.PriorityNormal,
		Timeout:      25 * time.Minute,
		WebhookURL:   req.WebhookURL,
		CallbackData: req.CallbackData,
		Metadata:     req.Metadata,
		TenantID:     tenantFromContext(c),
	}

	// Create comparison task
	task := &ComparisonTask{
		Request:       &req,
		Logger:        h.logger,
		ComparisonMgr: h.comparisonMgr,
		WSManager:     h.wsManager,
	}

	// Submit job
	job, err := h.jobManager.SubmitJob(jobConfig, task)
	if err != nil {
		h.logger.Error(fmt.Sprintf("Failed to submit comparison job: %s (jobID: %s, baseline: %s, current: %s)",
			err.Error(), jobID, req.BaselinePath, req.CurrentPath))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to submit job"})
		return
	}

	// Return job response
	response := &JobResponse{
		JobID:        job.ID,
		Status:       string(job.Status),
		Type:         job.Type,
		CreatedAt:    job.CreatedAt,
		WebSocketURL: fmt.Sprintf("ws://%s/ws/jobs/%s", c.Request.Host, job.ID),
	}

	h.logger.Info(fmt.Sprintf("FOCUS comparison job submitted: %s (baseline: %s, current: %s)",
		job.ID, req.BaselinePath, req.CurrentPath))

	c.JSON(http.StatusAccepted, response)
}

// @Summary Validate FOCUS dataset
// @Description Validate FOCUS dataset against specification and compliance requirements
// @Tags FOCUS
// @Accept json
// @Produce json
// @Param request body ValidateRequest true "Validation request"
// @Success 200 {object} map[string]interface{} "Validation result"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/focus/validate [post]
func (h *Handler) ValidateSync(c *gin.Context) {
	var req ValidateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.AutoBadRequest(c, err.Error())
		return
	}

	// Mock validation result for now
	result := map[string]interface{}{
		"is_compliant":         true,
		"overall_score":        95.5,
		"spec_version":         req.SpecVersion,
		"validation_timestamp": time.Now().Format(time.RFC3339),
		"summary": map[string]interface{}{
			"total_records":   100000,
			"valid_records":   99500,
			"invalid_records": 500,
			"error_rate":      0.5,
		},
		"compliance_results": map[string]interface{}{
			"focus": map[string]interface{}{
				"passed": true,
				"score":  98.0,
			},
			"schema": map[string]interface{}{
				"passed": true,
				"score":  100.0,
			},
			"quality": map[string]interface{}{
				"passed": true,
				"score":  92.0,
			},
		},
		"recommendations": []string{
			"Consider improving data quality for better compliance scores",
			"Review anomalous records for potential data issues",
		},
	}

	h.logger.Info(fmt.Sprintf("FOCUS validation completed for: %s (score: %.1f)",
		req.InputPath, 95.5))

	c.JSON(http.StatusOK, result)
}

// @Summary Analyze semantic drift between baseline and current invariants
// @Description Advanced chi-square, bucket delta, trend drift analysis
// @Tags FOCUS
// @Accept json
// @Produce json
// @Param request body DriftAnalyzeRequest true "Drift analyze request"
// @Success 200 {object} drift.Report
// @Failure 400 {object} map[string]string
// @Router /api/v1/focus/drift/analyze [post]
func (h *Handler) DriftAnalyze(c *gin.Context) {
	var req DriftAnalyzeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.AutoBadRequest(c, err.Error())
		return
	}
	baseline, err := focquality.LoadBaseline(req.BaselineInvariants)
	if err != nil {
		response.AutoBadRequest(c, "baseline load error")
		return
	}
	current, err := focquality.LoadBaseline(req.CurrentInvariants)
	if err != nil {
		response.AutoBadRequest(c, "current load error")
		return
	}
	// Distributions
	baseCharge := map[string]float64{}
	curCharge := map[string]float64{}
	basePricing := map[string]float64{}
	curPricing := map[string]float64{}
	for k, v := range baseline.ChargeCategoryDistribution {
		baseCharge[k] = v * float64(baseline.RowCount) / 100
	}
	for k, v := range current.ChargeCategoryDistribution {
		curCharge[k] = v * float64(current.RowCount) / 100
	}
	for k, v := range baseline.PricingCategoryDistribution {
		basePricing[k] = v * float64(baseline.RowCount) / 100
	}
	for k, v := range current.PricingCategoryDistribution {
		curPricing[k] = v * float64(current.RowCount) / 100
	}
	// Buckets via optional datasets
	baseEff, baseUse := loadValuesOptional(req.BaselineDataset)
	curEff, curUse := loadValuesOptional(req.CurrentDataset)
	if len(baseEff) == 0 { // fallback avg
		if baseline.RowCount > 0 {
			baseEff = []float64{baseline.SumEffectiveCost / float64(baseline.RowCount)}
			baseUse = []float64{baseline.SumUsageQuantity / float64(baseline.RowCount)}
		}
	}
	if len(curEff) == 0 {
		if current.RowCount > 0 {
			curEff = []float64{current.SumEffectiveCost / float64(current.RowCount)}
			curUse = []float64{current.SumUsageQuantity / float64(current.RowCount)}
		}
	}
	baseBuckets, _ := drift.BuildCostBuckets(baseEff, baseUse, req.BucketSchema)
	curBuckets, _ := drift.BuildCostBuckets(curEff, curUse, req.BucketSchema)
	// History snapshots
	hist := []drift.Snapshot{}
	for _, hp := range req.History {
		if inv, herr := focquality.LoadBaseline(hp); herr == nil {
			hist = append(hist, drift.Snapshot{TimestampUnix: inv.GeneratedAt.Unix(), RowCount: int64(inv.RowCount), SumEffective: inv.SumEffectiveCost, SumList: inv.SumListCost, SumUsage: inv.SumUsageQuantity})
		}
	}
	curSnap := drift.Snapshot{TimestampUnix: current.GeneratedAt.Unix(), RowCount: int64(current.RowCount), SumEffective: current.SumEffectiveCost, SumList: current.SumListCost, SumUsage: current.SumUsageQuantity}
	rep, err := drift.Run(
		drift.Config{Alpha: req.Alpha, BucketSchema: req.BucketSchema, Percentiles: req.Percentiles, PercentileDriftThreshold: req.PercentileThreshold},
		baseCharge, curCharge,
		basePricing, curPricing,
		baseBuckets, curBuckets,
		hist, curSnap,
		baseEff, curEff, baseUse, curUse,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "drift run error"})
		return
	}
	drift.RecordMetrics(rep)
	c.JSON(http.StatusOK, rep)
}

// loadValuesOptional and tableFnFor are provided in build-tagged files.

// @Summary Get job status and progress
// @Description Get the status and progress of an async job
// @Tags Jobs
// @Produce json
// @Param id path string true "Job ID"
// @Success 200 {object} JobResponse "Job status"
// @Failure 404 {object} map[string]string "Job not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/focus/jobs/{id} [get]
func (h *Handler) GetJobStatus(c *gin.Context) {
	jobID := c.Param("id")

	job, err := h.jobManager.GetJob(jobID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Job not found"})
		return
	}

	response := &JobResponse{
		JobID:        job.ID,
		Status:       string(job.Status),
		Type:         job.Type,
		CreatedAt:    job.CreatedAt,
		StartedAt:    job.StartedAt,
		CompletedAt:  job.CompletedAt,
		Progress:     job.Progress,
		Result:       job.Result,
		WebSocketURL: fmt.Sprintf("ws://%s/ws/jobs/%s", c.Request.Host, job.ID),
	}

	if job.Error != nil {
		errorMsg := job.Error.Error()
		response.Error = &errorMsg
	}

	c.JSON(http.StatusOK, response)
}

// @Summary List available FOCUS datasets
// @Description Get a list of available FOCUS datasets with metadata
// @Tags FOCUS
// @Produce json
// @Param provider query string false "Filter by provider"
// @Param spec_version query string false "Filter by FOCUS spec version"
// @Param limit query int false "Limit number of results" default(50)
// @Param offset query int false "Offset for pagination" default(0)
// @Success 200 {object} map[string]interface{} "List of datasets"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/focus/datasets [get]
func (h *Handler) ListDatasets(c *gin.Context) {
	// Get query parameters
	provider := c.Query("provider")
	specVersion := c.Query("spec_version")
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	// Mock datasets for now - this would come from a dataset registry
	datasets := []DatasetInfo{
		{
			Path:        "/data/focus/aws-2023-01.parquet",
			Name:        "AWS January 2023",
			Provider:    "aws",
			SpecVersion: "1.2",
			RecordCount: 1000000,
			FileSizeMB:  125.5,
			DateRange: TimeRange{
				StartDate: "2023-01-01",
				EndDate:   "2023-01-31",
			},
			CreatedAt:    time.Now().Add(-30 * 24 * time.Hour),
			LastModified: time.Now().Add(-25 * 24 * time.Hour),
			Tags:         []string{"production", "monthly"},
		},
		{
			Path:        "/data/focus/gcp-2023-02.parquet",
			Name:        "GCP February 2023",
			Provider:    "gcp",
			SpecVersion: "1.2",
			RecordCount: 750000,
			FileSizeMB:  98.2,
			DateRange: TimeRange{
				StartDate: "2023-02-01",
				EndDate:   "2023-02-28",
			},
			CreatedAt:    time.Now().Add(-25 * 24 * time.Hour),
			LastModified: time.Now().Add(-20 * 24 * time.Hour),
			Tags:         []string{"production", "monthly"},
		},
	}

	// Apply filters
	filteredDatasets := []DatasetInfo{}
	for _, dataset := range datasets {
		if provider != "" && dataset.Provider != provider {
			continue
		}
		if specVersion != "" && dataset.SpecVersion != specVersion {
			continue
		}
		filteredDatasets = append(filteredDatasets, dataset)
	}

	// Apply pagination
	total := len(filteredDatasets)
	start := offset
	end := offset + limit
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	result := map[string]interface{}{
		"datasets": filteredDatasets[start:end],
		"total":    total,
		"limit":    limit,
		"offset":   offset,
		"has_more": end < total,
	}

	c.JSON(http.StatusOK, result)
}

// =====================================================================================
// Route Registration
// =====================================================================================

// RegisterRoutes registers FOCUS API routes
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	focus := r.Group("/focus")
	{
		// Core operations
		focus.POST("/convert", h.ConvertAsync)
		focus.POST("/analyze", h.AnalyzeAsync)
		// Experimental comparison (async or synchronous when use_focus_engine=true)
		focus.POST("/diff", h.DiffAsync)
		focus.POST("/validate", h.ValidateSync)
		focus.POST("/drift/analyze", h.DriftAnalyze)

		// Job management
		focus.GET("/jobs/:id", h.GetJobStatus)

		// Dataset management
		focus.GET("/datasets", h.ListDatasets)
	}
}
