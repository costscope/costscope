package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/costscope/costscope/internal/api/jobs"
	"github.com/costscope/costscope/internal/api/response"
	"github.com/costscope/costscope/internal/api/websocket"
	"github.com/costscope/costscope/internal/core/focus/conversion"
	"github.com/costscope/costscope/internal/core/focus/types"
	"github.com/costscope/costscope/internal/core/logging"
)

// =====================================================================================
// FOCUS Handler - Core FOCUS Operations
// =====================================================================================

// FocusHandler provides FOCUS operations endpoints
type FocusHandler struct {
	logger    *logging.Logger
	jobMgr    *jobs.Manager
	wsManager *websocket.Manager
	convMgr   *conversion.ConversionManager // async conversion manager
}

// NewFocusHandler creates a new FOCUS handler
func NewFocusHandler(logger *logging.Logger, jobMgr *jobs.Manager, wsManager *websocket.Manager, convMgr *conversion.ConversionManager) *FocusHandler {
	return &FocusHandler{
		logger:    logger,
		jobMgr:    jobMgr,
		wsManager: wsManager,
		convMgr:   convMgr,
	}
}

// ConvertData converts cost data to FOCUS format
func (h *FocusHandler) ConvertData(c *gin.Context) {
	var req struct {
		Provider            string   `json:"provider" binding:"required"`
		InputPath           string   `json:"input_path" binding:"required"`
		OutputPath          string   `json:"output_path" binding:"required"`
		Streaming           bool     `json:"streaming"`
		ChunkSize           int      `json:"chunk_size"`
		Workers             int      `json:"workers"`
		UseUnifiedMapper    *bool    `json:"use_unified_mapper,omitempty"`
		InvariantsEnabled   *bool    `json:"invariants_enabled,omitempty"`
		InvariantsBaseline  *string  `json:"invariants_baseline,omitempty"`
		InvariantsTolerance *float64 `json:"invariants_tolerance,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.AutoBadRequest(c, "invalid request")
		return
	}
	if h.convMgr == nil {
		response.AutoBadRequestCode(c, "conversion manager unavailable", "bad_request")
		return
	}
	cfg := &types.ConversionConfig{
		Provider:     req.Provider,
		InputPath:    req.InputPath,
		OutputPath:   req.OutputPath,
		Streaming:    req.Streaming,
		ChunkSize:    req.ChunkSize,
		Workers:      req.Workers,
		ConversionId: "api_conv_" + strconv.FormatInt(time.Now().UnixNano(), 10),
		CreatedAt:    time.Now(),
		CreatedBy:    "api",
	}
	if req.UseUnifiedMapper != nil {
		cfg.UseUnifiedMapper = *req.UseUnifiedMapper
	}
	if req.InvariantsEnabled != nil {
		cfg.InvariantsEnabled = *req.InvariantsEnabled
	}
	if req.InvariantsBaseline != nil {
		cfg.InvariantsBaseline = *req.InvariantsBaseline
	}
	if req.InvariantsTolerance != nil {
		cfg.InvariantsTolerance = *req.InvariantsTolerance
	}
	jobID, err := h.convMgr.SubmitJob(cfg)
	if err != nil {
		response.AutoBadRequestCode(c, err.Error(), "bad_request")
		return
	}
	// Return 202 Accepted with a consistent status label "accepted" (was "submitted")
	response.AutoOK(c, http.StatusAccepted, gin.H{"job_id": jobID, "status": "accepted"})
}

// AnalyzeData performs FOCUS data analysis
func (h *FocusHandler) AnalyzeData(c *gin.Context) {
	var request struct {
		InputPath    string   `json:"input_path" binding:"required"`
		AnalysisType string   `json:"analysis_type" binding:"required"`
		Dimensions   []string `json:"dimensions"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		response.AutoBadRequest(c, "Invalid request format")
		return
	}

	jobID := "focus-analyze-" + request.AnalysisType
	h.logger.InfoWithFields("FOCUS analysis requested", map[string]interface{}{
		"job_id":        jobID,
		"input_path":    request.InputPath,
		"analysis_type": request.AnalysisType,
	})

	response.AutoOK(c, http.StatusAccepted, gin.H{
		"job_id":        jobID,
		"status":        "accepted",
		"message":       "FOCUS analysis job created",
		"analysis_type": request.AnalysisType,
	})
}

// CompareData compares FOCUS datasets
func (h *FocusHandler) CompareData(c *gin.Context) {
	var request struct {
		Dataset1 string `json:"dataset1" binding:"required"`
		Dataset2 string `json:"dataset2" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		response.AutoBadRequest(c, "Invalid request format")
		return
	}

	jobID := "focus-compare"
	response.AutoOK(c, http.StatusAccepted, gin.H{
		"job_id":  jobID,
		"status":  "accepted",
		"message": "FOCUS comparison job created",
	})
}

// ValidateData validates FOCUS data compliance
func (h *FocusHandler) ValidateData(c *gin.Context) {
	var request struct {
		InputPath   string `json:"input_path" binding:"required"`
		SpecVersion string `json:"spec_version"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		response.AutoBadRequest(c, "Invalid request format")
		return
	}

	jobID := "focus-validate"
	response.AutoOK(c, http.StatusAccepted, gin.H{
		"job_id":  jobID,
		"status":  "accepted",
		"message": "FOCUS validation job created",
	})
}

// GetJob returns job status and details
func (h *FocusHandler) GetJob(c *gin.Context) {
	id := c.Param("id")
	if h.convMgr == nil {
		response.AutoNotFound404(c, "conversion manager unavailable")
		return
	}
	job, err := h.convMgr.GetJobStatus(id)
	if err != nil {
		response.AutoNotFound404(c, "job not found")
		return
	}
	payload := gin.H{
		"id":         job.ID,
		"status":     job.Status,
		"started_at": job.StartTime.UTC().Format(time.RFC3339),
	}
	if job.EndTime != nil {
		payload["completed_at"] = job.EndTime.UTC().Format(time.RFC3339)
	}
	if job.Progress != nil {
		payload["progress"] = job.Progress
	}
	if job.Result != nil {
		payload["result"] = job.Result
	}
	response.AutoOK200(c, payload)
}

// ListJobs returns all jobs for the user
func (h *FocusHandler) ListJobs(c *gin.Context) {
	if h.convMgr == nil {
		response.AutoOK200(c, gin.H{"jobs": []any{}, "total": 0})
		return
	}
	list := h.convMgr.ListActiveJobs()
	out := make([]gin.H, 0, len(list))
	for _, j := range list {
		item := gin.H{"id": j.ID, "status": j.Status, "started_at": j.StartTime.UTC().Format(time.RFC3339)}
		if j.Progress != nil {
			item["progress"] = j.Progress
		}
		out = append(out, item)
	}
	response.AutoOK200(c, gin.H{"jobs": out, "total": len(out)})
}

// CancelJob cancels a running job
func (h *FocusHandler) CancelJob(c *gin.Context) {
	id := c.Param("id")
	if h.convMgr == nil {
		response.AutoNotFound404(c, "conversion manager unavailable")
		return
	}
	if err := h.convMgr.CancelJob(id); err != nil {
		response.AutoNotFound404(c, "job not found or not cancellable")
		return
	}
	response.AutoOK200(c, gin.H{"id": id, "status": "cancelled"})
}

// ListJobHistory returns recently completed jobs (bounded by ?limit=)
func (h *FocusHandler) ListJobHistory(c *gin.Context) {
	if h.convMgr == nil {
		response.AutoOK200(c, gin.H{"history": []any{}, "total": 0})
		return
	}
	limStr := c.Query("limit")
	limit := 0
	if limStr != "" {
		if v, err := strconv.Atoi(limStr); err == nil {
			limit = v
		}
	}
	hist := h.convMgr.GetJobHistory(limit)
	out := make([]gin.H, 0, len(hist))
	for _, r := range hist {
		out = append(out, gin.H{
			"id":          r.ConversionId,
			"input_file":  r.InputFile,
			"output_file": r.OutputFile,
			"success":     r.Success,
			"duration_ms": r.Duration.Milliseconds(),
			"records_in":  r.InputRecords,
			"records_out": r.OutputRecords,
		})
	}
	response.AutoOK200(c, gin.H{"history": out, "total": len(out)})
}
