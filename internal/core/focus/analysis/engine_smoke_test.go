package analysis_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	analysis "github.com/costscope/costscope/internal/core/focus/analysis"
	"github.com/costscope/costscope/internal/core/logging"
)

func newTestLogger() *logging.Logger { return logging.NewLogger(logging.LevelInfo) }

// Smoke test: end-to-end AnalyzeFOCUSDataset runs and returns a result
func TestAnalysisEngine_AnalyzeFOCUSDataset_Smoke(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "sample.parquet")
	if err := os.WriteFile(f, []byte(""), 0o600); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	eng := analysis.NewEngine(newTestLogger(), nil)
	opts := analysis.AnalysisOptions{MLEnabled: true, TrendAnalysis: true, ForecastDays: 7}

	res, err := eng.AnalyzeFOCUSDataset(f, opts)
	if err != nil {
		t.Fatalf("AnalyzeFOCUSDataset error: %v", err)
	}
	if res == nil {
		t.Fatalf("expected non-nil result")
	}
	if res.Summary.ServicesCount == 0 {
		t.Errorf("expected ServicesCount > 0")
	}
	if res.Metadata.Version == "" {
		t.Errorf("expected non-empty version")
	}
}

// Smoke test: individual analysis helpers operate on small synthetic inputs
func TestAnalysisEngine_Methods_Smoke(t *testing.T) {
	eng := analysis.NewEngine(newTestLogger(), nil)

	// Create an 8-point series to exercise anomaly detection thresholding
	now := time.Now()
	pts := []analysis.DataPoint{
		{Date: now.AddDate(0, 0, -8), Cost: 100, Usage: 10, Source: "EC2"},
		{Date: now.AddDate(0, 0, -7), Cost: 110, Usage: 11, Source: "EC2"},
		{Date: now.AddDate(0, 0, -6), Cost: 95, Usage: 9, Source: "EC2"},
		{Date: now.AddDate(0, 0, -5), Cost: 102, Usage: 10, Source: "EC2"},
		{Date: now.AddDate(0, 0, -4), Cost: 98, Usage: 10, Source: "EC2"},
		{Date: now.AddDate(0, 0, -3), Cost: 101, Usage: 10, Source: "EC2"},
		{Date: now.AddDate(0, 0, -2), Cost: 99, Usage: 10, Source: "EC2"},
		{Date: now.AddDate(0, 0, -1), Cost: 150, Usage: 12, Source: "EC2"}, // potential anomaly
	}

	if _, err := eng.DetectAnomalies(pts, []string{"statistical"}); err != nil {
		t.Fatalf("DetectAnomalies error: %v", err)
	}

	if _, err := eng.AnalyzeTrends(pts, true); err != nil {
		t.Fatalf("AnalyzeTrends error: %v", err)
	}

	if _, err := eng.GenerateForecasts(pts, 3, 0.95); err != nil {
		t.Fatalf("GenerateForecasts error: %v", err)
	}

	svcs := []analysis.ServiceSummary{{Service: "EC2", Region: "us-east-1", Account: "123", TotalCost: 1000, ResourceCount: 1}}
	if _, err := eng.FindOptimizations(svcs, []string{"rightsizing"}); err != nil {
		t.Fatalf("FindOptimizations error: %v", err)
	}

	// Export is a no-op placeholder; ensure it returns nil
	res := &analysis.AnalysisResult{Summary: analysis.AnalysisSummary{ServicesCount: 1}}
	if err := eng.ExportResults(res, "json", ""); err != nil {
		t.Fatalf("ExportResults error: %v", err)
	}
}
