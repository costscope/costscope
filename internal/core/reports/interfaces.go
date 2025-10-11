package reports

import (
	"context"

	"local/costscope/internal/core/reports/types"
)

// ReportService defines the interface for report generation and management
type ReportService interface {
	// Generate different types of reports
	GenerateCostAnalysisReport(ctx context.Context, options *types.ReportOptions) (*types.CostAnalysisReport, error)
	GenerateUsageSummaryReport(ctx context.Context, options *types.ReportOptions) (*types.UsageSummaryReport, error)
	GenerateTrendAnalysisReport(ctx context.Context, options *types.ReportOptions) (*types.TrendAnalysisReport, error)
	GenerateAnomalyReport(ctx context.Context, options *types.ReportOptions) (*types.AnomalyReport, error)
	GenerateForecastReport(ctx context.Context, options *types.ReportOptions) (*types.ForecastReport, error)
	GenerateExecutiveSummaryReport(ctx context.Context, options *types.ReportOptions) (*types.ExecutiveSummaryReport, error)

	// Report management
	ListReports(ctx context.Context, filters *types.ReportFilters) ([]*types.ReportInfo, error)
	GetReport(ctx context.Context, reportID string) (*types.ReportInfo, error)
	DeleteReport(ctx context.Context, reportID string) error

	// Export functionality
	ExportReport(ctx context.Context, reportID string, format types.ExportFormat, output string) error
}

// ReportGenerator defines the interface for generating specific report components
type ReportGenerator interface {
	GenerateData(ctx context.Context, options *types.ReportOptions) (interface{}, error)
	ValidateOptions(options *types.ReportOptions) error
	GetSupportedFormats() []types.ExportFormat
}

// ReportExporter defines the interface for exporting reports to different formats
type ReportExporter interface {
	// Export writes the report in the requested format to the destination (file path, object store URL, http endpoint, etc).
	// It returns the number of bytes written (when determinable; 0 otherwise) and a hex-encoded SHA256 checksum
	// of the written payload (empty string when not applicable, e.g. HTTP exporter that streams without local buffering).
	Export(ctx context.Context, report interface{}, format types.ExportFormat, output string) (bytesWritten int64, checksumSHA256 string, err error)
	GetSupportedFormats() []types.ExportFormat
}

// ReportScheduler defines the interface for scheduling automatic report generation
type ReportScheduler interface {
	ScheduleReport(ctx context.Context, schedule *types.ReportSchedule) error
	UpdateSchedule(ctx context.Context, scheduleID string, schedule *types.ReportSchedule) error
	DeleteSchedule(ctx context.Context, scheduleID string) error
	ListSchedules(ctx context.Context) ([]*types.ReportSchedule, error)
	RunScheduledReports(ctx context.Context) error
}

// ReportTemplate defines the interface for report templates
type ReportTemplate interface {
	GetName() string
	GetDescription() string
	GetSupportedReportTypes() []types.ReportType
	ApplyTemplate(report interface{}) (interface{}, error)
	ValidateReport(report interface{}) error
}

// ReportOrchestrator coordinates multiple report generators and manages the overall workflow
type ReportOrchestrator interface {
	// Main orchestration methods
	GenerateReport(ctx context.Context, reportType types.ReportType, options *types.ReportOptions) (interface{}, error)
	GenerateMultipleReports(ctx context.Context, requests []*types.ReportRequest) ([]*types.ReportResult, error)

	// Template management
	RegisterTemplate(template ReportTemplate) error
	GetAvailableTemplates() []ReportTemplate

	// Generator management
	RegisterGenerator(reportType types.ReportType, generator ReportGenerator) error
	GetGenerator(reportType types.ReportType) (ReportGenerator, error)
}
