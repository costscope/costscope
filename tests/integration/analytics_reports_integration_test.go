package integration_test

import (
	"context"
	"testing"
	"time"

	analyticsTypes "github.com/costscope/costscope/cmd/modules/analytics/types"
	"github.com/costscope/costscope/internal/core/analytics"
	"github.com/costscope/costscope/internal/core/logging"
	"github.com/costscope/costscope/internal/core/reports"
	reportTypes "github.com/costscope/costscope/internal/core/reports/types"
)

// TestAnalyticsToReportsIntegration validates cross-module basic flow:
// analytics service produces results and reports service generates a report.
func TestAnalyticsToReportsIntegration(t *testing.T) {
	logger := logging.NewLogger("info")

	// Initialize services
	aCfg := &analytics.Config{MLEnabled: true, EnablePredictions: true}
	aSvc := analytics.NewBasicService(aCfg, logger)
	rSvc := reports.NewBasicReportService(logger)

	// Step 1: Run analysis (simulated)
	aOpts := &analyticsTypes.AnalyticsOptions{
		TableName:     "focus_costs",
		Currency:      "USD",
		EnableML:      true,
		Filters:       map[string]interface{}{"region": "us-east-1"},
		GroupByFields: []string{"service", "region"},
		SortOrder:     "desc",
	}

	res, err := aSvc.Analyze(aOpts)
	if err != nil {
		t.Fatalf("analytics Analyze failed: %v", err)
	}
	if res == nil || res.AnalysisResult == nil {
		t.Fatalf("analytics results are nil")
	}

	// Step 2: Generate a report using report service
	rOpts := &reportTypes.ReportOptions{
		ReportType:  reportTypes.ReportTypeCostAnalysis,
		Title:       "Integration Cost Analysis",
		Description: "Report generated from analytics flow",
		DateRange: reportTypes.DateRange{
			StartDate: time.Now().AddDate(0, 0, -7),
			EndDate:   time.Now(),
		},
		Currency:  aOpts.Currency,
		IncludeML: true,
		Metadata: map[string]interface{}{
			"source_table": res.TableName,
			"filters":      res.FiltersCount,
		},
	}

	rep, err := rSvc.GenerateCostAnalysisReport(context.Background(), rOpts)
	if err != nil {
		t.Fatalf("failed to generate report: %v", err)
	}
	if rep == nil || rep.ID == "" {
		t.Fatalf("invalid report generated")
	}
}
