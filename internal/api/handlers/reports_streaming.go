package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/costscope/costscope/internal/api/jobs"
	"github.com/costscope/costscope/internal/api/response"
	"github.com/costscope/costscope/internal/core/logging"
	"github.com/costscope/costscope/internal/core/monitoring/telemetry"
	"github.com/costscope/costscope/internal/core/reports"
	"github.com/costscope/costscope/internal/core/streaming"
)

// =====================================================================================
// Reports Handler - Report Generation & Management
// =====================================================================================

type ReportsHandler struct {
	logger        *logging.Logger
	jobMgr        *jobs.Manager
	reportService *reports.BasicReportService
}

func NewReportsHandler(logger *logging.Logger, jobMgr *jobs.Manager, svc *reports.BasicReportService) *ReportsHandler {
	return &ReportsHandler{logger: logger, jobMgr: jobMgr, reportService: svc}
}

func (h *ReportsHandler) GenerateReport(c *gin.Context) {
	var request struct {
		TemplateID string                 `json:"template_id" binding:"required"`
		Parameters map[string]interface{} `json:"parameters"`
		Format     string                 `json:"format"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		response.AutoBadRequest(c, "Invalid request format")
		return
	}
	// 202 Accepted retained; using envelope for consistency
	response.AutoOK(c, http.StatusAccepted, gin.H{"job_id": "report-generate", "status": "accepted"})
}

func (h *ReportsHandler) ListReports(c *gin.Context) {
	reports := []gin.H{
		{"id": "rpt-001", "name": "Monthly Cost Report", "status": "completed", "created_at": "2025-01-31T00:00:00Z"},
		{"id": "rpt-002", "name": "Service Usage Report", "status": "running", "created_at": "2025-01-31T01:00:00Z"},
	}
	response.AutoOK200(c, gin.H{"reports": reports, "total": len(reports)})
}

// ListExports returns exported report metadata (if metadata store configured).
// Query params: format, after, before (RFC3339), limit, offset.
func (h *ReportsHandler) ListExports(c *gin.Context) {
	if h.reportService == nil {
		response.AutoOK200(c, gin.H{"exports": []interface{}{}, "count": 0})
		return
	}
	format := c.Query("format")
	afterStr := c.Query("after")
	beforeStr := c.Query("before")
	limitStr := c.Query("limit")
	offsetStr := c.Query("offset")
	var afterPtr *time.Time
	if afterStr != "" {
		if t, err := time.Parse(time.RFC3339, afterStr); err == nil {
			afterPtr = &t
		}
	}
	var beforePtr *time.Time
	if beforeStr != "" {
		if t, err := time.Parse(time.RFC3339, beforeStr); err == nil {
			beforePtr = &t
		}
	}
	limit := 50
	if limitStr != "" {
		if v, err := strconv.Atoi(limitStr); err == nil {
			limit = v
		}
	}
	offset := 0
	if offsetStr != "" {
		if v, err := strconv.Atoi(offsetStr); err == nil {
			offset = v
		}
	}
	list, err := h.reportService.ListReportMetadataOptions(c.Request.Context(), &reports.MetadataListOptions{Format: format, CreatedAfter: afterPtr, CreatedBefore: beforePtr, Limit: limit, Offset: offset})
	if err != nil {
		response.AutoBadRequest(c, "failed to list exports: "+err.Error())
		return
	}
	response.AutoOK200(c, gin.H{"exports": list, "count": len(list)})
}

// VerifyExport compares stored checksum with recalculated checksum.
func (h *ReportsHandler) VerifyExport(c *gin.Context) {
	if h.reportService == nil {
		response.AutoBadRequest(c, "report service not configured")
		return
	}
	id := c.Param("id")
	match, actual, err := h.reportService.VerifyReportIntegrity(c.Request.Context(), id)
	if err != nil {
		response.AutoBadRequest(c, err.Error())
		return
	}
	code := "mismatch"
	if match {
		code = "ok"
	}
	response.AutoOK200(c, gin.H{"id": id, "status": code, "actual_checksum": actual})
}

func (h *ReportsHandler) GetReport(c *gin.Context) {
	id := c.Param("id")
	response.AutoOK200(c, gin.H{"id": id, "name": "Report " + id, "status": "completed"})
}

func (h *ReportsHandler) DeleteReport(c *gin.Context) {
	id := c.Param("id")
	response.AutoOK200(c, gin.H{"id": id, "status": "deleted"})
}

func (h *ReportsHandler) DownloadReport(c *gin.Context) {
	id := c.Param("id")
	response.AutoOK200(c, gin.H{"id": id, "download_url": "/api/v1/reports/" + id + "/download"})
}

func (h *ReportsHandler) ListTemplates(c *gin.Context) {
	templates := []gin.H{
		{"id": "tpl-001", "name": "Cost Analysis Template", "category": "cost"},
		{"id": "tpl-002", "name": "Usage Report Template", "category": "usage"},
	}
	response.AutoOK200(c, gin.H{"templates": templates, "total": len(templates)})
}

func (h *ReportsHandler) ScheduleReport(c *gin.Context) {
	var request struct {
		TemplateID string `json:"template_id" binding:"required"`
		Schedule   string `json:"schedule" binding:"required"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		response.AutoBadRequest(c, "Invalid request format")
		return
	}
	response.AutoCreated201(c, gin.H{"schedule_id": "sch-001", "status": "scheduled"})
}

// =====================================================================================
// Streaming Handler - Real-time Streaming Operations
// =====================================================================================

type StreamingHandler struct {
	logger *logging.Logger
	jobMgr *jobs.Manager
	engine streaming.StreamingEngine
}

func NewStreamingHandler(logger *logging.Logger, jobMgr *jobs.Manager, engine streaming.StreamingEngine) *StreamingHandler {
	return &StreamingHandler{logger: logger, jobMgr: jobMgr, engine: engine}
}

func (h *StreamingHandler) CreateStreamingJob(c *gin.Context) {
	// Enterprise engine: available only with 'enterprise' tag; slim builds return a helpful error via stub methods
	var body struct {
		Source      string `json:"source" binding:"required"`
		Destination string `json:"destination" binding:"required"`
		Operation   string `json:"operation"` // convert|analyze|validate|diff|merge|compress
		Config      struct {
			ChunkSizeMB        int  `json:"chunk_size_mb"`
			ParallelWorkers    int  `json:"parallel_workers"`
			CompressionEnabled bool `json:"compression_enabled"`
			CheckpointEnabled  bool `json:"checkpoint_enabled"`
			ValidationEnabled  bool `json:"validation_enabled"`
			RetryOnError       bool `json:"retry_on_error"`
			MaxRetryAttempts   int  `json:"max_retry_attempts"`
		} `json:"config"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.AutoBadRequest(c, "invalid request: "+err.Error())
		return
	}

	opType := streaming.StreamingOperationConvert
	switch body.Operation {
	case "analyze":
		opType = streaming.StreamingOperationAnalyze
	case "validate":
		opType = streaming.StreamingOperationValidate
	case "diff":
		opType = streaming.StreamingOperationDiff
	case "merge":
		opType = streaming.StreamingOperationMerge
	case "compress":
		opType = streaming.StreamingOperationCompress
	default:
		// default convert
	}

	req := &streaming.StreamingOperationRequest{
		OperationID:     generateStreamingID(),
		SourcePath:      body.Source,
		DestinationPath: body.Destination,
		Operation:       opType,
		Configuration: &streaming.StreamingConfiguration{
			ChunkSizeMB:        body.Config.ChunkSizeMB,
			ParallelWorkers:    body.Config.ParallelWorkers,
			CompressionEnabled: body.Config.CompressionEnabled,
			CheckpointEnabled:  body.Config.CheckpointEnabled,
			ValidationEnabled:  body.Config.ValidationEnabled,
			RetryOnError:       body.Config.RetryOnError,
			MaxRetryAttempts:   body.Config.MaxRetryAttempts,
		},
	}

	op, err := h.engine.StartStreamingOperation(context.Background(), req)
	if err != nil {
		// In slim builds, this returns an enterprise-only error; surface as 501 Not Implemented
		// Keep 501 semantics but wrap envelope
		response.AutoFail(c, http.StatusNotImplemented, err.Error(), "not_implemented")
		return
	}
	telemetry.StreamingEvents.WithLabelValues(op.ID, "created").Inc()
	response.AutoCreated201(c, gin.H{"id": op.ID, "status": op.Status, "operation": op.Operation})
}

func (h *StreamingHandler) ListStreamingJobs(c *gin.Context) {
	ops := h.engine.ListActiveOperations()
	list := make([]gin.H, 0, len(ops))
	for _, op := range ops {
		list = append(list, gin.H{
			"id":          op.ID,
			"source":      op.SourcePath,
			"destination": op.DestinationPath,
			"status":      op.Status,
			"operation":   op.Operation,
			"progress":    op.Progress.PercentComplete,
		})
	}
	response.AutoOK200(c, gin.H{"jobs": list, "total": len(list)})
}

func (h *StreamingHandler) GetStreamingJob(c *gin.Context) {
	id := c.Param("id")
	op, err := h.engine.GetStreamingOperation(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	response.AutoOK200(c, gin.H{
		"id":           op.ID,
		"status":       op.Status,
		"operation":    op.Operation,
		"progress":     op.Progress.PercentComplete,
		"bytes":        op.Progress.BytesProcessed,
		"total_bytes":  op.Progress.TotalBytes,
		"started_at":   op.StartTime,
		"completed_at": op.EndTime,
	})
}

func (h *StreamingHandler) StartJob(c *gin.Context) {
	id := c.Param("id")
	if err := h.engine.ResumeStreamingOperation(id); err != nil {
		response.AutoBadRequest(c, err.Error())
		return
	}
	h.streamEventAndRespond(c, id, "start", "running")
}

func (h *StreamingHandler) StopJob(c *gin.Context) {
	id := c.Param("id")
	if err := h.engine.CancelStreamingOperation(id); err != nil {
		response.AutoBadRequest(c, err.Error())
		return
	}
	h.streamEventAndRespond(c, id, "stop", "cancelled")
}

func (h *StreamingHandler) DeleteJob(c *gin.Context) {
	id := c.Param("id")
	// Cancel if still running; engine cleans up completed ops after a delay
	_ = h.engine.CancelStreamingOperation(id)
	h.streamEventAndRespond(c, id, "deleted", "deleted")
}

func (h *StreamingHandler) ListSources(c *gin.Context) {
	sources := []gin.H{
		{"id": "aws-cur", "name": "AWS Cost and Usage Reports", "type": "batch"},
		{"id": "azure-billing", "name": "Azure Billing API", "type": "streaming"},
		{"id": "gcp-export", "name": "GCP Billing Export", "type": "batch"},
	}
	response.AutoOK200(c, gin.H{"sources": sources, "total": len(sources)})
}

// streamEventAndRespond is a tiny helper to record a streaming event and emit a uniform JSON response.
// It centralizes a pattern repeated across Start/Stop/Delete while preserving behavior.
func (h *StreamingHandler) streamEventAndRespond(c *gin.Context, id, action, status string) {
	telemetry.StreamingEvents.WithLabelValues(id, action).Inc()
	response.AutoOK200(c, gin.H{"id": id, "status": status})
}

// generateStreamingID returns a simple time-based identifier.
func generateStreamingID() string {
	return time.Now().UTC().Format("20060102-150405.000000000")
}
