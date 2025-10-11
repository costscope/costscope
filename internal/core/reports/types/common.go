package types

import (
	"context"
	"time"
)

// ReportType represents the type of report to generate
type ReportType string

const (
	ReportTypeCostAnalysis     ReportType = "cost_analysis"
	ReportTypeUsageSummary     ReportType = "usage_summary"
	ReportTypeTrendAnalysis    ReportType = "trend_analysis"
	ReportTypeAnomaly          ReportType = "anomaly"
	ReportTypeForecast         ReportType = "forecast"
	ReportTypeExecutiveSummary ReportType = "executive_summary"
)

// String returns the string representation of ReportType
func (rt ReportType) String() string {
	return string(rt)
}

// ExportFormat represents the format for exporting reports
type ExportFormat string

const (
	ExportFormatJSON  ExportFormat = "json"
	ExportFormatCSV   ExportFormat = "csv"
	ExportFormatHTML  ExportFormat = "html"
	ExportFormatPDF   ExportFormat = "pdf"
	ExportFormatYAML  ExportFormat = "yaml"
	ExportFormatXML   ExportFormat = "xml"
	ExportFormatExcel ExportFormat = "excel"
	// New formats supported by exporters
	ExportFormatParquet ExportFormat = "parquet"
	ExportFormatHTTP    ExportFormat = "http"
)

// String returns the string representation of ExportFormat
func (ef ExportFormat) String() string {
	return string(ef)
}

// RenderReport abstracts rendering of complex binary formats (PDF/Excel) so the
// exporter layer can remain focused on storage concerns. Implementations may be
// simple stubs initially and swapped for full engines later without touching
// service orchestration code.
type RenderReport interface {
	Render(ctx context.Context, report interface{}) ([]byte, error)
	ContentType() string
}

// ReportStatus represents the status of a report
type ReportStatus string

const (
	ReportStatusPending    ReportStatus = "pending"
	ReportStatusGenerating ReportStatus = "generating"
	ReportStatusCompleted  ReportStatus = "completed"
	ReportStatusFailed     ReportStatus = "failed"
	ReportStatusCancelled  ReportStatus = "cancelled"
)

// String returns the string representation of ReportStatus
func (rs ReportStatus) String() string {
	return string(rs)
}

// Priority represents the priority level
type Priority string

const (
	PriorityLow      Priority = "low"
	PriorityMedium   Priority = "medium"
	PriorityHigh     Priority = "high"
	PriorityCritical Priority = "critical"
)

// String returns the string representation of Priority
func (p Priority) String() string {
	return string(p)
}

// DateRange represents a time range for reports
type DateRange struct {
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
}

// ReportOptions contains configuration for report generation
type ReportOptions struct {
	ReportType  ReportType             `json:"report_type"`
	Title       string                 `json:"title,omitempty"`
	Description string                 `json:"description,omitempty"`
	DateRange   DateRange              `json:"date_range"`
	Providers   []string               `json:"providers,omitempty"`
	Regions     []string               `json:"regions,omitempty"`
	Services    []string               `json:"services,omitempty"`
	Accounts    []string               `json:"accounts,omitempty"`
	Currency    string                 `json:"currency,omitempty"`
	GroupBy     []string               `json:"group_by,omitempty"`
	Filters     map[string]interface{} `json:"filters,omitempty"`
	IncludeML   bool                   `json:"include_ml,omitempty"`
	DetailLevel string                 `json:"detail_level,omitempty"`
	OutputPath  string                 `json:"output_path,omitempty"`
	TemplateID  string                 `json:"template_id,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// ReportFilters contains filters for listing reports
type ReportFilters struct {
	ReportType   *ReportType   `json:"report_type,omitempty"`
	Status       *ReportStatus `json:"status,omitempty"`
	CreatedAfter *time.Time    `json:"created_after,omitempty"`
	CreatedBy    string        `json:"created_by,omitempty"`
	Tags         []string      `json:"tags,omitempty"`
	Limit        int           `json:"limit,omitempty"`
	Offset       int           `json:"offset,omitempty"`
}

// ReportInfo contains metadata about a report
type ReportInfo struct {
	ID             string                 `json:"id"`
	ReportType     ReportType             `json:"report_type"`
	Title          string                 `json:"title"`
	Description    string                 `json:"description"`
	Status         ReportStatus           `json:"status"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
	CompletedAt    *time.Time             `json:"completed_at,omitempty"`
	CreatedBy      string                 `json:"created_by"`
	Size           int64                  `json:"size"`
	FilePath       string                 `json:"file_path,omitempty"`
	ExportFormats  []ExportFormat         `json:"export_formats,omitempty"`
	Tags           []string               `json:"tags,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	ErrorMessage   string                 `json:"error_message,omitempty"`
	ProcessingTime float64                `json:"processing_time_ms,omitempty"`
}

// ReportRequest represents a request to generate a report
type ReportRequest struct {
	ID      string         `json:"id,omitempty"`
	Options *ReportOptions `json:"options"`
	Format  ExportFormat   `json:"format,omitempty"`
	Output  string         `json:"output,omitempty"`
}

// ReportResult represents the result of a report generation
type ReportResult struct {
	RequestID   string      `json:"request_id"`
	ReportInfo  *ReportInfo `json:"report_info"`
	Success     bool        `json:"success"`
	Error       string      `json:"error,omitempty"`
	GeneratedAt time.Time   `json:"generated_at"`
}

// ReportSchedule represents a scheduled report configuration
type ReportSchedule struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Options     *ReportOptions         `json:"options"`
	CronExpr    string                 `json:"cron_expression"`
	Enabled     bool                   `json:"enabled"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	LastRun     *time.Time             `json:"last_run,omitempty"`
	NextRun     *time.Time             `json:"next_run,omitempty"`
	RunCount    int                    `json:"run_count"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// ReportSummary provides a high-level summary of report content
type ReportSummary struct {
	TotalItems  int                    `json:"total_items"`
	TotalCost   float64                `json:"total_cost,omitempty"`
	Currency    string                 `json:"currency,omitempty"`
	DateRange   DateRange              `json:"date_range"`
	TopInsights []string               `json:"top_insights,omitempty"`
	KeyMetrics  map[string]interface{} `json:"key_metrics,omitempty"`
	Alerts      []string               `json:"alerts,omitempty"`
	Confidence  float64                `json:"confidence,omitempty"`
}
