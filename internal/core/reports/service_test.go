package reports

import (
	"context"
	"testing"
	"time"

	"github.com/costscope/costscope/internal/core/logging"
	"github.com/costscope/costscope/internal/core/reports/types"
)

func TestBasicReportService_GenerateCostAnalysisReport(t *testing.T) {
	logger := logging.NewLogger("info")
	service := NewBasicReportService(logger)

	options := &types.ReportOptions{
		ReportType:  types.ReportTypeCostAnalysis,
		Title:       "Test Cost Analysis",
		Description: "Test cost analysis report",
		DateRange: types.DateRange{
			StartDate: time.Now().AddDate(0, -1, 0),
			EndDate:   time.Now(),
		},
		Currency:    "USD",
		IncludeML:   true,
		DetailLevel: "standard",
	}

	report, err := service.GenerateCostAnalysisReport(context.Background(), options)
	if err != nil {
		t.Fatalf("Failed to generate cost analysis report: %v", err)
	}

	if report == nil {
		t.Fatal("Report is nil")
	}

	if report.ID == "" {
		t.Error("Report ID is empty")
	}

	if report.Title != options.Title {
		t.Errorf("Expected title %s, got %s", options.Title, report.Title)
	}

	if report.Currency != options.Currency {
		t.Errorf("Expected currency %s, got %s", options.Currency, report.Currency)
	}

	if report.TotalCost <= 0 {
		t.Error("Total cost should be greater than 0")
	}

	if len(report.CostByService) == 0 {
		t.Error("CostByService should not be empty")
	}

	if len(report.CostByRegion) == 0 {
		t.Error("CostByRegion should not be empty")
	}

	if len(report.TopCostDrivers) == 0 {
		t.Error("TopCostDrivers should not be empty")
	}

	if len(report.Optimization) == 0 {
		t.Error("Optimization recommendations should not be empty")
	}

	if report.Summary.Confidence <= 0 || report.Summary.Confidence > 1 {
		t.Error("Confidence should be between 0 and 1")
	}
}

func TestBasicReportService_GenerateUsageSummaryReport(t *testing.T) {
	logger := logging.NewLogger("info")
	service := NewBasicReportService(logger)

	options := &types.ReportOptions{
		ReportType:  types.ReportTypeUsageSummary,
		Title:       "Test Usage Summary",
		Description: "Test usage summary report",
		DateRange: types.DateRange{
			StartDate: time.Now().AddDate(0, -1, 0),
			EndDate:   time.Now(),
		},
		Currency:    "EUR",
		IncludeML:   true,
		DetailLevel: "detailed",
	}

	report, err := service.GenerateUsageSummaryReport(context.Background(), options)
	if err != nil {
		t.Fatalf("Failed to generate usage summary report: %v", err)
	}

	if report == nil {
		t.Fatal("Report is nil")
	}

	if report.ID == "" {
		t.Error("Report ID is empty")
	}

	if report.Title != options.Title {
		t.Errorf("Expected title %s, got %s", options.Title, report.Title)
	}

	if len(report.ResourceUtilization) == 0 {
		t.Error("ResourceUtilization should not be empty")
	}

	if len(report.ServiceUsage) == 0 {
		t.Error("ServiceUsage should not be empty")
	}

	if report.Summary.Confidence <= 0 || report.Summary.Confidence > 1 {
		t.Error("Confidence should be between 0 and 1")
	}
}

func TestBasicReportService_GenerateTrendAnalysisReport(t *testing.T) {
	logger := logging.NewLogger("info")
	service := NewBasicReportService(logger)

	options := &types.ReportOptions{
		ReportType:  types.ReportTypeTrendAnalysis,
		Title:       "Test Trend Analysis",
		Description: "Test trend analysis report",
		DateRange: types.DateRange{
			StartDate: time.Now().AddDate(0, -3, 0),
			EndDate:   time.Now(),
		},
		Currency:    "GBP",
		IncludeML:   true,
		DetailLevel: "standard",
	}

	report, err := service.GenerateTrendAnalysisReport(context.Background(), options)
	if err != nil {
		t.Fatalf("Failed to generate trend analysis report: %v", err)
	}

	if report == nil {
		t.Fatal("Report is nil")
	}

	if report.ID == "" {
		t.Error("Report ID is empty")
	}

	if len(report.Trends) == 0 {
		t.Error("Trends should not be empty")
	}

	if len(report.Forecasts) == 0 {
		t.Error("Forecasts should not be empty")
	}

	if len(report.MLInsights) == 0 {
		t.Error("MLInsights should not be empty")
	}

	if report.Summary.Confidence <= 0 || report.Summary.Confidence > 1 {
		t.Error("Confidence should be between 0 and 1")
	}
}

func TestBasicReportService_GenerateAnomalyReport(t *testing.T) {
	logger := logging.NewLogger("info")
	service := NewBasicReportService(logger)

	options := &types.ReportOptions{
		ReportType:  types.ReportTypeAnomaly,
		Title:       "Test Anomaly Detection",
		Description: "Test anomaly detection report",
		DateRange: types.DateRange{
			StartDate: time.Now().AddDate(0, 0, -7),
			EndDate:   time.Now(),
		},
		Currency:    "USD",
		IncludeML:   true,
		DetailLevel: "detailed",
	}

	report, err := service.GenerateAnomalyReport(context.Background(), options)
	if err != nil {
		t.Fatalf("Failed to generate anomaly report: %v", err)
	}

	if report == nil {
		t.Fatal("Report is nil")
	}

	if report.ID == "" {
		t.Error("Report ID is empty")
	}

	if len(report.Anomalies) == 0 {
		t.Error("Anomalies should not be empty")
	}

	if len(report.Alerts) == 0 {
		t.Error("Alerts should not be empty")
	}

	if report.RiskLevel == "" {
		t.Error("RiskLevel should not be empty")
	}

	if report.Summary.Confidence <= 0 || report.Summary.Confidence > 1 {
		t.Error("Confidence should be between 0 and 1")
	}
}

func TestBasicReportService_GenerateForecastReport(t *testing.T) {
	logger := logging.NewLogger("info")
	service := NewBasicReportService(logger)

	options := &types.ReportOptions{
		ReportType:  types.ReportTypeForecast,
		Title:       "Test Forecast",
		Description: "Test forecast report",
		DateRange: types.DateRange{
			StartDate: time.Now().AddDate(0, -2, 0),
			EndDate:   time.Now(),
		},
		Currency:    "USD",
		IncludeML:   true,
		DetailLevel: "standard",
	}

	report, err := service.GenerateForecastReport(context.Background(), options)
	if err != nil {
		t.Fatalf("Failed to generate forecast report: %v", err)
	}

	if report == nil {
		t.Fatal("Report is nil")
	}

	if report.ID == "" {
		t.Error("Report ID is empty")
	}

	if len(report.Forecasts) == 0 {
		t.Error("Forecasts should not be empty")
	}

	if len(report.Scenarios) == 0 {
		t.Error("Scenarios should not be empty")
	}

	if report.Confidence <= 0 || report.Confidence > 1 {
		t.Error("Confidence should be between 0 and 1")
	}

	if report.Summary.Confidence <= 0 || report.Summary.Confidence > 1 {
		t.Error("Summary confidence should be between 0 and 1")
	}
}

func TestBasicReportService_GenerateExecutiveSummaryReport(t *testing.T) {
	logger := logging.NewLogger("info")
	service := NewBasicReportService(logger)

	options := &types.ReportOptions{
		ReportType:  types.ReportTypeExecutiveSummary,
		Title:       "Test Executive Summary",
		Description: "Test executive summary report",
		DateRange: types.DateRange{
			StartDate: time.Now().AddDate(0, -1, 0),
			EndDate:   time.Now(),
		},
		Currency:    "USD",
		IncludeML:   false,
		DetailLevel: "basic",
	}

	report, err := service.GenerateExecutiveSummaryReport(context.Background(), options)
	if err != nil {
		t.Fatalf("Failed to generate executive summary report: %v", err)
	}

	if report == nil {
		t.Fatal("Report is nil")
	}

	if report.ID == "" {
		t.Error("Report ID is empty")
	}

	if report.ExecutiveSummary.TotalSpend <= 0 {
		t.Error("TotalSpend should be greater than 0")
	}

	if report.ExecutiveSummary.TopCostDriver == "" {
		t.Error("TopCostDriver should not be empty")
	}

	if report.ExecutiveSummary.RiskLevel == "" {
		t.Error("RiskLevel should not be empty")
	}

	if len(report.ExecutiveSummary.KeyInsights) == 0 {
		t.Error("KeyInsights should not be empty")
	}

	if len(report.KeyMetrics) == 0 {
		t.Error("KeyMetrics should not be empty")
	}

	if len(report.Recommendations) == 0 {
		t.Error("Recommendations should not be empty")
	}
}

func TestBasicReportService_ListReports(t *testing.T) {
	logger := logging.NewLogger("info")
	service := NewBasicReportService(logger)

	// Generate a few reports first
	options := &types.ReportOptions{
		ReportType: types.ReportTypeCostAnalysis,
		Title:      "Test Report 1",
		DateRange: types.DateRange{
			StartDate: time.Now().AddDate(0, -1, 0),
			EndDate:   time.Now(),
		},
	}

	_, err := service.GenerateCostAnalysisReport(context.Background(), options)
	if err != nil {
		t.Fatalf("Failed to generate test report: %v", err)
	}

	options.Title = "Test Report 2"
	options.ReportType = types.ReportTypeUsageSummary
	_, err = service.GenerateUsageSummaryReport(context.Background(), options)
	if err != nil {
		t.Fatalf("Failed to generate test report: %v", err)
	}

	// Test listing all reports
	reports, err := service.ListReports(context.Background(), nil)
	if err != nil {
		t.Fatalf("Failed to list reports: %v", err)
	}

	// We should have exactly 2 reports from this test instance
	if len(reports) < 2 {
		t.Errorf("Expected at least 2 reports, got %d", len(reports))
	}

	// Test filtering by report type
	costAnalysisType := types.ReportTypeCostAnalysis
	filters := &types.ReportFilters{
		ReportType: &costAnalysisType,
	}

	filteredReports, err := service.ListReports(context.Background(), filters)
	if err != nil {
		t.Fatalf("Failed to list filtered reports: %v", err)
	}

	if len(filteredReports) != 1 {
		t.Errorf("Expected 1 filtered report, got %d", len(filteredReports))
	}

	if filteredReports[0].ReportType != types.ReportTypeCostAnalysis {
		t.Errorf("Expected cost analysis report, got %s", filteredReports[0].ReportType)
	}
}

func TestBasicReportService_GetReport(t *testing.T) {
	logger := logging.NewLogger("info")
	service := NewBasicReportService(logger)

	// Generate a test report
	options := &types.ReportOptions{
		ReportType: types.ReportTypeCostAnalysis,
		Title:      "Test Report",
		DateRange: types.DateRange{
			StartDate: time.Now().AddDate(0, -1, 0),
			EndDate:   time.Now(),
		},
	}

	report, err := service.GenerateCostAnalysisReport(context.Background(), options)
	if err != nil {
		t.Fatalf("Failed to generate test report: %v", err)
	}

	// Test getting the report
	retrievedReport, err := service.GetReport(context.Background(), report.ID)
	if err != nil {
		t.Fatalf("Failed to get report: %v", err)
	}

	if retrievedReport.ID != report.ID {
		t.Errorf("Expected report ID %s, got %s", report.ID, retrievedReport.ID)
	}

	if retrievedReport.Title != report.Title {
		t.Errorf("Expected title %s, got %s", report.Title, retrievedReport.Title)
	}

	// Test getting non-existent report
	_, err = service.GetReport(context.Background(), "non-existent-id")
	if err == nil {
		t.Error("Expected error for non-existent report")
	}
}

func TestBasicReportService_DeleteReport(t *testing.T) {
	logger := logging.NewLogger("info")
	service := NewBasicReportService(logger)

	// Generate a test report
	options := &types.ReportOptions{
		ReportType: types.ReportTypeCostAnalysis,
		Title:      "Test Report to Delete",
		DateRange: types.DateRange{
			StartDate: time.Now().AddDate(0, -1, 0),
			EndDate:   time.Now(),
		},
	}

	report, err := service.GenerateCostAnalysisReport(context.Background(), options)
	if err != nil {
		t.Fatalf("Failed to generate test report: %v", err)
	}

	// Verify report exists
	_, err = service.GetReport(context.Background(), report.ID)
	if err != nil {
		t.Fatalf("Report should exist before deletion: %v", err)
	}

	// Delete the report
	err = service.DeleteReport(context.Background(), report.ID)
	if err != nil {
		t.Fatalf("Failed to delete report: %v", err)
	}

	// Verify report is deleted
	_, err = service.GetReport(context.Background(), report.ID)
	if err == nil {
		t.Error("Report should not exist after deletion")
	}

	// Test deleting non-existent report
	err = service.DeleteReport(context.Background(), "non-existent-id")
	if err == nil {
		t.Error("Expected error for deleting non-existent report")
	}
}

func TestReportTypes_String(t *testing.T) {
	tests := []struct {
		reportType types.ReportType
		expected   string
	}{
		{types.ReportTypeCostAnalysis, "cost_analysis"},
		{types.ReportTypeUsageSummary, "usage_summary"},
		{types.ReportTypeTrendAnalysis, "trend_analysis"},
		{types.ReportTypeAnomaly, "anomaly"},
		{types.ReportTypeForecast, "forecast"},
		{types.ReportTypeExecutiveSummary, "executive_summary"},
	}

	for _, test := range tests {
		if test.reportType.String() != test.expected {
			t.Errorf("Expected %s, got %s", test.expected, test.reportType.String())
		}
	}
}

func TestExportFormats_String(t *testing.T) {
	tests := []struct {
		format   types.ExportFormat
		expected string
	}{
		{types.ExportFormatJSON, "json"},
		{types.ExportFormatCSV, "csv"},
		{types.ExportFormatHTML, "html"},
		{types.ExportFormatPDF, "pdf"},
		{types.ExportFormatYAML, "yaml"},
		{types.ExportFormatXML, "xml"},
	}

	for _, test := range tests {
		if test.format.String() != test.expected {
			t.Errorf("Expected %s, got %s", test.expected, test.format.String())
		}
	}
}

func TestReportStatus_String(t *testing.T) {
	tests := []struct {
		status   types.ReportStatus
		expected string
	}{
		{types.ReportStatusPending, "pending"},
		{types.ReportStatusGenerating, "generating"},
		{types.ReportStatusCompleted, "completed"},
		{types.ReportStatusFailed, "failed"},
		{types.ReportStatusCancelled, "cancelled"},
	}

	for _, test := range tests {
		if test.status.String() != test.expected {
			t.Errorf("Expected %s, got %s", test.expected, test.status.String())
		}
	}
}

func TestPriority_String(t *testing.T) {
	tests := []struct {
		priority types.Priority
		expected string
	}{
		{types.PriorityLow, "low"},
		{types.PriorityMedium, "medium"},
		{types.PriorityHigh, "high"},
		{types.PriorityCritical, "critical"},
	}

	for _, test := range tests {
		if test.priority.String() != test.expected {
			t.Errorf("Expected %s, got %s", test.expected, test.priority.String())
		}
	}
}
