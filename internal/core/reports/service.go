package reports

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"local/costscope/internal/core/logging"
	"local/costscope/internal/core/reports/exporters"
	"local/costscope/internal/core/reports/types"
)

// BasicReportService provides a basic implementation of the ReportService interface.
type BasicReportService struct {
	logger         *logging.Logger
	generators     map[types.ReportType]ReportGenerator
	exporters      map[types.ExportFormat]ReportExporter
	reports        map[string]*types.ReportInfo // In-memory storage
	reportsMu      sync.RWMutex                 // Mutex for thread-safe access to reports
	reportContents map[string]interface{}
	metadataStore  MetadataStore // optional persistence layer
}

// ===== context helpers (non-exported key) for plumbing output dir precedence source through CLI → service =====
type ctxKey string

const outputDirSourceCtxKey ctxKey = "reports.output_dir_source"

// WithOutputDirSource annotates context with the precedence source (explicit|yaml|env|default) of the resolved
// output directory used for an export. The service records this in ReportMetadata when present.
func WithOutputDirSource(ctx context.Context, source string) context.Context {
	if source == "" {
		return ctx
	}
	return context.WithValue(ctx, outputDirSourceCtxKey, source)
}

func outputDirSourceFromContext(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(outputDirSourceCtxKey).(string)
	return v, ok
}

// NewBasicReportService creates a new basic report service
func NewBasicReportService(logger *logging.Logger) *BasicReportService {
	return &BasicReportService{
		logger:         logger,
		generators:     make(map[types.ReportType]ReportGenerator),
		exporters:      make(map[types.ExportFormat]ReportExporter),
		reports:        make(map[string]*types.ReportInfo),
		reportContents: make(map[string]interface{}),
		metadataStore:  nil, // lazy; callers can inject via WithMetadataStore
	}
}

// WithMetadataStore enables metadata persistence (fluent style; returns same instance for chaining).
// Intentional optional wiring to keep public constructor stable.
func (s *BasicReportService) WithMetadataStore(store MetadataStore) *BasicReportService {
	s.metadataStore = store
	return s
}

// GenerateCostAnalysisReport generates a cost analysis report
func (s *BasicReportService) GenerateCostAnalysisReport(ctx context.Context, options *types.ReportOptions) (*types.CostAnalysisReport, error) {
	s.logger.Info("Starting cost analysis report generation")

	reportID := s.generateReportID()
	startTime := time.Now()

	// Create report info
	// Register the report
	s.registerReport(reportID, types.ReportTypeCostAnalysis, options.Title)

	// Simulate report generation
	report := &types.CostAnalysisReport{
		ID:          reportID,
		Title:       options.Title,
		Description: options.Description,
		GeneratedAt: startTime,
		DateRange:   options.DateRange,
		TotalCost:   12345.67,
		Currency:    s.getCurrency(options),
		CostByService: []types.ServiceCostBreakdown{
			{
				ServiceName: "EC2",
				Provider:    "aws",
				Cost:        5000.00,
				Percentage:  40.5,
				Trend:       "increasing",
			},
			{
				ServiceName: "S3",
				Provider:    "aws",
				Cost:        2500.00,
				Percentage:  20.2,
				Trend:       "stable",
			},
		},
		CostByRegion: []types.RegionCostBreakdown{
			{
				Region:     "us-east-1",
				Provider:   "aws",
				Cost:       8000.00,
				Percentage: 64.8,
				Trend:      "increasing",
			},
			{
				Region:     "us-west-2",
				Provider:   "aws",
				Cost:       4345.67,
				Percentage: 35.2,
				Trend:      "stable",
			},
		},
		TopCostDrivers: []types.CostDriver{
			{
				ID:          "driver-1",
				Name:        "Large EC2 Instances",
				Type:        "compute",
				Cost:        3000.00,
				Impact:      "high",
				Trend:       "increasing",
				Description: "Multiple large EC2 instances running 24/7",
			},
		},
		Optimization: []types.OptimizationRecommendation{
			{
				ID:                 "opt-1",
				Type:               "rightsizing",
				Priority:           types.PriorityHigh,
				Title:              "Rightsize EC2 Instances",
				Description:        "Several EC2 instances are underutilized and can be downsized",
				Impact:             "high",
				Effort:             "low",
				PotentialSavings:   1200.00,
				ImplementationTime: "1-2 weeks",
				RiskLevel:          "low",
			},
		},
		Summary: types.ReportSummary{
			TotalItems: 25,
			TotalCost:  12345.67,
			Currency:   s.getCurrency(options),
			DateRange:  options.DateRange,
			TopInsights: []string{
				"Cost increased by 15% compared to previous period",
				"EC2 represents 40% of total costs",
				"Potential savings of $1,200 identified",
			},
			KeyMetrics: map[string]interface{}{
				"cost_per_day":      411.52,
				"top_service":       "EC2",
				"optimization_rate": 9.7,
			},
			Confidence: 0.92,
		},
		Metadata: options.Metadata,
	}

	processingTime := time.Since(startTime).Milliseconds()
	s.logger.Info(fmt.Sprintf("Cost analysis report generated successfully: %s (processing_time: %dms)", reportID, processingTime))

	// Persist full content for export and update status to completed
	s.setReportContent(reportID, report)
	s.completeReport(reportID)

	return report, nil
}

// GenerateUsageSummaryReport generates a usage summary report
func (s *BasicReportService) GenerateUsageSummaryReport(ctx context.Context, options *types.ReportOptions) (*types.UsageSummaryReport, error) {
	s.logger.Info("Starting usage summary report generation")

	reportID := s.generateReportID()
	startTime := time.Now()

	// Register the report
	s.registerReport(reportID, types.ReportTypeUsageSummary, options.Title)

	report := &types.UsageSummaryReport{
		ID:          reportID,
		Title:       options.Title,
		Description: options.Description,
		GeneratedAt: startTime,
		DateRange:   options.DateRange,
		ResourceUtilization: []types.ResourceUtilization{
			{
				ResourceID:   "i-1234567890abcdef0",
				ResourceType: "EC2",
				Provider:     "aws",
				Region:       "us-east-1",
				Utilization:  65.5,
				Capacity:     map[string]string{"cpu": "4 vCPU", "memory": "16 GiB"},
				Status:       "running",
			},
		},
		ServiceUsage: []types.ServiceUsage{
			{
				ServiceName: "EC2",
				Provider:    "aws",
				Usage:       map[string]interface{}{"instances": 12, "hours": 8760},
				Cost:        5000.00,
				Trend:       "increasing",
			},
		},
		Summary: types.ReportSummary{
			TotalItems: 50,
			DateRange:  options.DateRange,
			TopInsights: []string{
				"Average resource utilization: 65%",
				"12 EC2 instances currently running",
				"Storage utilization: 78%",
			},
			KeyMetrics: map[string]interface{}{
				"avg_utilization":  65.5,
				"total_resources":  50,
				"efficiency_score": 7.8,
			},
			Confidence: 0.88,
		},
		Metadata: options.Metadata,
	}

	processingTime := time.Since(startTime).Milliseconds()
	s.logger.Info(fmt.Sprintf("Usage summary report generated successfully: %s (processing_time: %dms)", reportID, processingTime))

	// Persist full content for export and update status to completed
	s.setReportContent(reportID, report)
	s.completeReport(reportID)

	return report, nil
}

// GenerateTrendAnalysisReport generates a trend analysis report
func (s *BasicReportService) GenerateTrendAnalysisReport(ctx context.Context, options *types.ReportOptions) (*types.TrendAnalysisReport, error) {
	s.logger.Info("Starting trend analysis report generation")

	reportID := s.generateReportID()
	startTime := time.Now()

	// Register the report
	s.registerReport(reportID, types.ReportTypeTrendAnalysis, options.Title)
	report := &types.TrendAnalysisReport{
		ID:          reportID,
		Title:       options.Title,
		Description: options.Description,
		GeneratedAt: startTime,
		DateRange:   options.DateRange,
		Trends: []types.TrendAnalysis{
			{
				Metric:      "total_cost",
				Direction:   "increasing",
				Strength:    0.75,
				Confidence:  0.92,
				StartDate:   options.DateRange.StartDate,
				EndDate:     options.DateRange.EndDate,
				Description: "Total cost showing strong upward trend",
			},
		},
		Forecasts: []types.ForecastData{
			{
				Metric:         "total_cost",
				Date:           time.Now().AddDate(0, 1, 0),
				Value:          13500.00,
				ConfidenceLow:  12000.00,
				ConfidenceHigh: 15000.00,
				Method:         "linear_regression",
			},
		},
		MLInsights: []types.MLInsight{
			{
				Type:        "trend_prediction",
				Confidence:  0.85,
				Description: "ML model predicts continued cost growth",
				Impact:      "high",
				Data: map[string]interface{}{
					"growth_rate": 0.15,
					"seasonality": "detected",
				},
			},
		},
		Summary: types.ReportSummary{
			TotalItems: 10,
			DateRange:  options.DateRange,
			TopInsights: []string{
				"Strong upward cost trend detected",
				"15% growth rate predicted",
				"Seasonal patterns identified",
			},
			KeyMetrics: map[string]interface{}{
				"trend_strength": 0.75,
				"growth_rate":    0.15,
				"seasonality":    true,
			},
			Confidence: 0.92,
		},
		Metadata: options.Metadata,
	}

	processingTime := time.Since(startTime).Milliseconds()
	s.logger.Info(fmt.Sprintf("Trend analysis report generated successfully: %s (processing_time: %dms)", reportID, processingTime))

	// Persist full content for export and update status to completed
	s.setReportContent(reportID, report)
	s.completeReport(reportID)

	return report, nil
}

// GenerateAnomalyReport generates an anomaly detection report
func (s *BasicReportService) GenerateAnomalyReport(ctx context.Context, options *types.ReportOptions) (*types.AnomalyReport, error) {
	s.logger.Info("Starting anomaly report generation")

	reportID := s.generateReportID()
	startTime := time.Now()

	// Register the report
	s.registerReport(reportID, types.ReportTypeAnomaly, options.Title)
	report := &types.AnomalyReport{
		ID:          reportID,
		Title:       options.Title,
		Description: options.Description,
		GeneratedAt: startTime,
		DateRange:   options.DateRange,
		Anomalies: []types.AnomalyData{
			{
				ID:          "anomaly-1",
				Type:        "cost_spike",
				Metric:      "daily_cost",
				DetectedAt:  time.Now().AddDate(0, 0, -1),
				Severity:    "high",
				Score:       0.89,
				Expected:    400.00,
				Actual:      850.00,
				Deviation:   1.125,
				Description: "Unusual cost spike detected in EC2 service",
				Context: map[string]interface{}{
					"service": "EC2",
					"region":  "us-east-1",
				},
			},
		},
		Alerts: []types.AlertData{
			{
				ID:           "alert-1",
				Type:         "budget_threshold",
				Severity:     "warning",
				Message:      "Monthly budget threshold exceeded",
				CreatedAt:    time.Now(),
				Acknowledged: false,
				Context: map[string]interface{}{
					"threshold": 10000.00,
					"current":   12345.67,
				},
			},
		},
		RiskLevel: "medium",
		Summary: types.ReportSummary{
			TotalItems: 3,
			DateRange:  options.DateRange,
			TopInsights: []string{
				"1 high-severity anomaly detected",
				"Cost spike of 112.5% in EC2",
				"Budget threshold exceeded",
			},
			Alerts: []string{
				"Monthly budget threshold exceeded",
				"Unusual EC2 cost spike detected",
			},
			Confidence: 0.89,
		},
		Metadata: options.Metadata,
	}

	processingTime := time.Since(startTime).Milliseconds()
	s.logger.Info(fmt.Sprintf("Anomaly report generated successfully: %s (processing_time: %dms)", reportID, processingTime))

	// Persist full content for export and update status to completed
	s.setReportContent(reportID, report)
	s.completeReport(reportID)

	return report, nil
}

// GenerateForecastReport generates a forecast report
func (s *BasicReportService) GenerateForecastReport(ctx context.Context, options *types.ReportOptions) (*types.ForecastReport, error) {
	s.logger.Info("Starting forecast report generation")

	reportID := s.generateReportID()
	startTime := time.Now()

	// Register the report
	s.registerReport(reportID, types.ReportTypeForecast, options.Title)
	report := &types.ForecastReport{
		ID:          reportID,
		Title:       options.Title,
		Description: options.Description,
		GeneratedAt: startTime,
		DateRange:   options.DateRange,
		Forecasts: []types.ForecastData{
			{
				Metric:         "monthly_cost",
				Date:           time.Now().AddDate(0, 1, 0),
				Value:          14500.00,
				ConfidenceLow:  13000.00,
				ConfidenceHigh: 16000.00,
				Method:         "arima",
			},
			{
				Metric:         "monthly_cost",
				Date:           time.Now().AddDate(0, 2, 0),
				Value:          15800.00,
				ConfidenceLow:  14000.00,
				ConfidenceHigh: 17600.00,
				Method:         "arima",
			},
		},
		Scenarios: []types.ScenarioData{
			{
				Name:        "Conservative Growth",
				Description: "Assuming current usage patterns continue",
				Probability: 0.7,
				Impact:      "moderate",
				Results: map[string]interface{}{
					"monthly_cost": 14500.00,
					"growth_rate":  0.08,
				},
			},
			{
				Name:        "Aggressive Growth",
				Description: "Assuming business expansion",
				Probability: 0.3,
				Impact:      "high",
				Results: map[string]interface{}{
					"monthly_cost": 18000.00,
					"growth_rate":  0.25,
				},
			},
		},
		Confidence: 0.87,
		Summary: types.ReportSummary{
			TotalItems: 8,
			DateRange:  options.DateRange,
			TopInsights: []string{
				"Cost expected to grow 8-15% next month",
				"Conservative scenario most likely",
				"Business expansion could increase costs 25%",
			},
			KeyMetrics: map[string]interface{}{
				"forecast_accuracy": 0.87,
				"expected_growth":   0.08,
				"scenarios":         2,
			},
			Confidence: 0.87,
		},
		Metadata: options.Metadata,
	}

	processingTime := time.Since(startTime).Milliseconds()
	s.logger.Info(fmt.Sprintf("Forecast report generated successfully: %s (processing_time: %dms)", reportID, processingTime))

	// Persist full content for export and update status to completed
	s.setReportContent(reportID, report)
	s.completeReport(reportID)

	return report, nil
}

// GenerateExecutiveSummaryReport generates an executive summary report
func (s *BasicReportService) GenerateExecutiveSummaryReport(ctx context.Context, options *types.ReportOptions) (*types.ExecutiveSummaryReport, error) {
	s.logger.Info("Starting executive summary report generation")

	reportID := s.generateReportID()
	startTime := time.Now()

	// Register the report
	s.registerReport(reportID, types.ReportTypeExecutiveSummary, options.Title)
	report := &types.ExecutiveSummaryReport{
		ID:          reportID,
		Title:       options.Title,
		Description: options.Description,
		GeneratedAt: startTime,
		DateRange:   options.DateRange,
		ExecutiveSummary: types.ExecutiveSummaryData{
			TotalSpend:      12345.67,
			SpendChange:     15.2,
			TopCostDriver:   "EC2 Compute",
			OptimizationOpp: 1200.00,
			RiskLevel:       "medium",
			Recommendations: 5,
			KeyInsights: []string{
				"Cloud spend increased 15.2% this month",
				"EC2 compute drives 40% of total costs",
				"$1,200 in immediate savings opportunities identified",
				"Resource utilization averaging 65%",
			},
		},
		KeyMetrics: []types.KeyMetric{
			{
				Name:        "Total Cloud Spend",
				Value:       12345.67,
				Unit:        "USD",
				Change:      15.2,
				Trend:       "increasing",
				Description: "Total cloud infrastructure costs",
			},
			{
				Name:        "Cost Per Unit",
				Value:       0.85,
				Unit:        "USD/hour",
				Change:      -2.1,
				Trend:       "decreasing",
				Description: "Average cost per compute hour",
			},
		},
		Recommendations: []types.ExecutiveRecommendation{
			{
				ID:               "exec-rec-1",
				Title:            "Implement Reserved Instances",
				Priority:         types.PriorityHigh,
				PotentialSavings: 800.00,
				Timeframe:        "30 days",
				Complexity:       "low",
				Description:      "Purchase reserved instances for predictable workloads",
			},
		},
		Summary: types.ReportSummary{
			TotalItems: 15,
			TotalCost:  12345.67,
			Currency:   s.getCurrency(options),
			DateRange:  options.DateRange,
			TopInsights: []string{
				"15.2% cost increase requires attention",
				"$1,200 optimization opportunity available",
				"Medium risk level - monitoring recommended",
			},
			KeyMetrics: map[string]interface{}{
				"executive_score": 7.2,
				"risk_level":      "medium",
				"action_items":    5,
			},
			Confidence: 0.91,
		},
		Metadata: options.Metadata,
	}

	processingTime := time.Since(startTime).Milliseconds()
	s.logger.Info(fmt.Sprintf("Executive summary report generated successfully: %s (processing_time: %dms)", reportID, processingTime))

	// Persist full content for export and update status to completed
	s.setReportContent(reportID, report)
	s.completeReport(reportID)

	return report, nil
}

// ListReports lists all reports with optional filtering
func (s *BasicReportService) ListReports(ctx context.Context, filters *types.ReportFilters) ([]*types.ReportInfo, error) {
	s.logger.Info(fmt.Sprintf("Listing reports (total: %d)", len(s.reports)))

	var result []*types.ReportInfo
	for _, report := range s.reports {
		if s.matchesFilter(report, filters) {
			result = append(result, report)
		}
	}

	// Apply limit and offset
	if filters != nil {
		if filters.Offset > 0 && filters.Offset < len(result) {
			result = result[filters.Offset:]
		}
		if filters.Limit > 0 && filters.Limit < len(result) {
			result = result[:filters.Limit]
		}
	}

	return result, nil
}

// GetReport retrieves a specific report by ID
func (s *BasicReportService) GetReport(ctx context.Context, reportID string) (*types.ReportInfo, error) {
	s.logger.Info(fmt.Sprintf("Getting report: %s", reportID))

	report, exists := s.reports[reportID]
	if !exists {
		return nil, fmt.Errorf("report not found: %s", reportID)
	}

	return report, nil
}

// DeleteReport removes a report
func (s *BasicReportService) DeleteReport(ctx context.Context, reportID string) error {
	s.logger.Info(fmt.Sprintf("Deleting report: %s", reportID))

	if _, exists := s.reports[reportID]; !exists {
		return fmt.Errorf("report not found: %s", reportID)
	}

	delete(s.reports, reportID)
	s.reportsMu.Lock()
	delete(s.reportContents, reportID)
	s.reportsMu.Unlock()
	return nil
}

// ExportReport exports a report to the specified format
func (s *BasicReportService) ExportReport(ctx context.Context, reportID string, format types.ExportFormat, output string) error {
	s.logger.Info(fmt.Sprintf("Exporting report: %s (format: %s, output: %s)", reportID, format, output))

	// Locate exporter
	exporter, ok := s.exporters[format]
	if !ok {
		// Lazily register defaults
		s.exporters[types.ExportFormatJSON] = exporters.NewJSONExporter()
		s.exporters[types.ExportFormatCSV] = exporters.NewCSVExporter()
		s.exporters[types.ExportFormatYAML] = exporters.NewYAMLExporter()
		s.exporters[types.ExportFormatParquet] = exporters.NewParquetExporter()
		s.exporters[types.ExportFormatHTTP] = exporters.NewHTTPExporter()
		// Extended exporters (PDF / Excel) – lightweight stubs until full engines integrated
		if _, exists := s.exporters[types.ExportFormatPDF]; !exists {
			s.exporters[types.ExportFormatPDF] = exporters.NewPDFExporter()
		}
		if _, exists := s.exporters[types.ExportFormatExcel]; !exists {
			s.exporters[types.ExportFormatExcel] = exporters.NewExcelExporter()
		}
		exporter = s.exporters[format]
		if exporter == nil {
			return fmt.Errorf("no exporter for format: %s", format)
		}
	}

	// Choose payload: full content (when available and requested) or metadata fallback
	var payload interface{}
	if includeContentFromContext(ctx) {
		if full := s.getReportContent(reportID); full != nil {
			payload = full
		}
	}
	if payload == nil {
		// Fallback to metadata (prefer existing)
		if r, ok := s.reports[reportID]; ok && r != nil {
			payload = r
		} else {
			payload = &types.ReportInfo{ID: reportID, ReportType: types.ReportTypeExecutiveSummary, Title: "Export Stub"}
		}
	}

	// Perform export (capture size & checksum when supported)
	bytesWritten, checksum, err := exporter.Export(ctx, payload, format, output)
	if err != nil {
		return err
	}

	// Opportunistically persist metadata (best-effort; ignore errors for now to avoid changing behavior)
	if s.metadataStore != nil {
		if r, ok := s.reports[reportID]; ok && r != nil {
			// Determine precedence source (default explicit if none set – maintains previous behavior)
			source := "explicit"
			if v, ok := outputDirSourceFromContext(ctx); ok && v != "" {
				source = v
			}
			md := &ReportMetadata{
				ID:              r.ID,
				Path:            output,
				Format:          string(format),
				SizeBytes:       bytesWritten,
				ChecksumSHA256:  checksum,
				CreatedAt:       time.Now().UTC(),
				OutputDirSource: source,
			}
			_ = s.metadataStore.Save(ctx, md)
		}
	}

	return nil
}

// Helper methods

func (s *BasicReportService) generateReportID() string {
	return fmt.Sprintf("report_%d", time.Now().UnixNano())
}

func (s *BasicReportService) getCurrency(options *types.ReportOptions) string {
	if options.Currency != "" {
		return options.Currency
	}
	return "USD"
}

func (s *BasicReportService) matchesFilter(report *types.ReportInfo, filters *types.ReportFilters) bool {
	if filters == nil {
		return true
	}

	if filters.ReportType != nil && *filters.ReportType != report.ReportType {
		return false
	}

	if filters.Status != nil && *filters.Status != report.Status {
		return false
	}

	if filters.CreatedAfter != nil && report.CreatedAt.Before(*filters.CreatedAfter) {
		return false
	}

	if filters.CreatedBy != "" && filters.CreatedBy != report.CreatedBy {
		return false
	}

	return true
}

// ListReportMetadata returns stored export metadata when a metadataStore is configured.
// Returns empty slice if store is nil to keep behavior predictable.
func (s *BasicReportService) ListReportMetadata(ctx context.Context) ([]*ReportMetadata, error) {
	if s.metadataStore == nil {
		return []*ReportMetadata{}, nil
	}
	return s.metadataStore.List(ctx)
}

// ListReportMetadataOptions returns filtered/paginated metadata when supported.
func (s *BasicReportService) ListReportMetadataOptions(ctx context.Context, opts *MetadataListOptions) ([]*ReportMetadata, error) {
	if s.metadataStore == nil {
		return []*ReportMetadata{}, nil
	}
	// Capability check: if store implements ListOptions use it; otherwise fallback to List
	type listOpt interface {
		ListOptions(context.Context, *MetadataListOptions) ([]*ReportMetadata, error)
	}
	if lo, ok := s.metadataStore.(listOpt); ok {
		return lo.ListOptions(ctx, opts)
	}
	all, err := s.metadataStore.List(ctx)
	if err != nil {
		return nil, err
	}
	return filterAndPageMetadata(all, opts), nil
}

// VerifyReportIntegrity re-hashes the exported file (if accessible locally) and compares with stored checksum.
// Returns (match, actualChecksum, error). If no checksum stored or file missing, match=false with explanatory error.
func (s *BasicReportService) VerifyReportIntegrity(ctx context.Context, id string) (bool, string, error) {
	if s.metadataStore == nil {
		return false, "", fmt.Errorf("metadata store not configured")
	}
	md, err := s.metadataStore.Get(ctx, id)
	if err != nil {
		return false, "", err
	}
	if md.ChecksumSHA256 == "" {
		return false, "", fmt.Errorf("no checksum stored for report %s", id)
	}
	// For HTTP exports path might be a URL; only verify local filesystem paths.
	if strings.HasPrefix(md.Path, "http://") || strings.HasPrefix(md.Path, "https://") {
		return false, "", fmt.Errorf("cannot verify remote HTTP export for report %s", id)
	}
	data, err := os.ReadFile(md.Path)
	if err != nil {
		return false, "", err
	}
	sum := sha256.Sum256(data)
	actual := fmt.Sprintf("%x", sum[:])
	return actual == md.ChecksumSHA256, actual, nil
}

// Note: registry-style mutators (RegisterGenerator/RegisterExporter) were removed
// as unused. Default exporters are lazily registered in ExportReport().

// NOTE: Historical wrapper reports.ResolveOutputPath removed (unused). Callers must import
// internal/core/reports/outputpath and use outputpath.ResolveOutputPath directly. This reduces
// surface area and avoids duplicated documentation. (SemVer: internal package, safe removal.)

// registerReport registers a report in the service's reports map
func (s *BasicReportService) registerReport(reportID string, reportType types.ReportType, title string) {
	s.reportsMu.Lock()
	defer s.reportsMu.Unlock()

	startTime := time.Now()
	reportInfo := &types.ReportInfo{
		ID:         reportID,
		ReportType: reportType,
		Title:      title,
		Status:     types.ReportStatusGenerating,
		CreatedAt:  startTime,
		UpdatedAt:  startTime,
		CreatedBy:  "system",
	}

	s.reports[reportID] = reportInfo
}

// completeReport marks a report as completed
func (s *BasicReportService) completeReport(reportID string) {
	s.reportsMu.Lock()
	defer s.reportsMu.Unlock()

	if reportInfo, exists := s.reports[reportID]; exists {
		reportInfo.Status = types.ReportStatusCompleted
		reportInfo.UpdatedAt = time.Now()
		s.reports[reportID] = reportInfo
	}
}

// setReportContent stores the full generated report for later export
func (s *BasicReportService) setReportContent(reportID string, content interface{}) {
	s.reportsMu.Lock()
	defer s.reportsMu.Unlock()
	s.reportContents[reportID] = content
}

// getReportContent returns a previously stored full report by ID
func (s *BasicReportService) getReportContent(reportID string) interface{} {
	s.reportsMu.RLock()
	defer s.reportsMu.RUnlock()
	return s.reportContents[reportID]
}
