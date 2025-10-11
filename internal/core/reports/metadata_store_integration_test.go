package reports

import (
	"context"
	"local/costscope/internal/core/logging"
	"local/costscope/internal/core/reports/types"
	"testing"
)

// Test that injecting a metadata store results in a Save call (implicitly via presence of entry).
func TestBasicReportService_Export_PersistsMetadataWhenStoreProvided(t *testing.T) {
	logger := logging.NewLogger(logging.LevelError)
	svc := NewBasicReportService(logger).WithMetadataStore(NewInMemoryMetadataStore(logger, 0, 0))

	// Generate a minimal executive summary report (reusing existing service method)
	rep, err := svc.GenerateExecutiveSummaryReport(context.Background(), &types.ReportOptions{Title: "Exec", Description: "", Currency: "USD"})
	if err != nil {
		t.Fatalf("generate report: %v", err)
	}
	if rep == nil {
		t.Fatalf("expected report")
	}

	// Export (JSON) – output path is a temp file path (not actually written by stub exporter maybe) but acceptable.
	if err := svc.ExportReport(context.Background(), rep.ID, types.ExportFormatJSON, "./tmp-export.json"); err != nil {
		t.Fatalf("export: %v", err)
	}

	store := svc.metadataStore.(*InMemoryMetadataStore)
	if _, ok := store.items[rep.ID]; !ok {
		t.Fatalf("expected metadata stored for report %s", rep.ID)
	}
}
