package exporters

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/costscope/costscope/internal/core/reports/types"
)

func TestJSONExporter_Export(t *testing.T) {
	exporter := NewJSONExporter()

	// Test supported formats
	formats := exporter.GetSupportedFormats()
	if len(formats) != 1 || formats[0] != types.ExportFormatJSON {
		t.Error("JSON exporter should only support JSON format")
	}

	// Create test report
	report := &types.CostAnalysisReport{
		ID:          "test-report-1",
		Title:       "Test Cost Analysis",
		GeneratedAt: time.Now(),
		TotalCost:   1234.56,
		Currency:    "USD",
		Summary: types.ReportSummary{
			TotalItems: 10,
			TotalCost:  1234.56,
			Currency:   "USD",
		},
	}

	// Create temporary file
	tempDir := t.TempDir()
	outputFile := filepath.Join(tempDir, "test_report.json")

	// Test export
	bytesWritten, checksum, err := exporter.Export(context.Background(), report, types.ExportFormatJSON, outputFile)
	if err != nil {
		t.Fatalf("Failed to export JSON: %v", err)
	}
	if bytesWritten == 0 {
		t.Errorf("expected non-zero bytesWritten")
	}
	if checksum == "" {
		t.Errorf("expected checksum to be set")
	}

	// Verify file exists
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Fatal("Output file was not created")
	}

	// Verify file content
	content, err := os.ReadFile(outputFile) // #nosec G304 - test file in controlled environment
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	if len(content) == 0 {
		t.Error("Output file is empty")
	}

	// Basic JSON validation - should contain key fields
	contentStr := string(content)
	if !contains(contentStr, "test-report-1") {
		t.Error("Output should contain report ID")
	}
	if !contains(contentStr, "Test Cost Analysis") {
		t.Error("Output should contain report title")
	}
	if !contains(contentStr, "1234.56") {
		t.Error("Output should contain total cost")
	}
}

func TestJSONExporter_UnsupportedFormat(t *testing.T) {
	exporter := NewJSONExporter()
	report := &types.CostAnalysisReport{ID: "test"}

	_, _, err := exporter.Export(context.Background(), report, types.ExportFormatCSV, "test.csv")
	if err == nil {
		t.Error("Should return error for unsupported format")
	}
}

func TestCSVExporter_Export(t *testing.T) {
	exporter := NewCSVExporter()

	// Test supported formats
	formats := exporter.GetSupportedFormats()
	if len(formats) != 1 || formats[0] != types.ExportFormatCSV {
		t.Error("CSV exporter should only support CSV format")
	}

	// Create test report
	report := &types.CostAnalysisReport{
		ID:          "test-report-2",
		Title:       "Test Cost Analysis CSV",
		GeneratedAt: time.Now(),
		TotalCost:   2345.67,
		Currency:    "EUR",
		CostByService: []types.ServiceCostBreakdown{
			{ServiceName: "EC2", Cost: 1000.00},
			{ServiceName: "S3", Cost: 500.00},
		},
		CostByRegion: []types.RegionCostBreakdown{
			{Region: "us-east-1", Cost: 1500.00},
		},
		TopCostDrivers: []types.CostDriver{
			{Name: "Compute", Cost: 1000.00},
		},
	}

	// Create temporary file
	tempDir := t.TempDir()
	outputFile := filepath.Join(tempDir, "test_report.csv")

	// Test export
	bytesWritten, checksum, err := exporter.Export(context.Background(), report, types.ExportFormatCSV, outputFile)
	if err != nil {
		t.Fatalf("Failed to export CSV: %v", err)
	}
	if bytesWritten == 0 {
		t.Errorf("expected non-zero bytesWritten")
	}
	if checksum == "" {
		t.Errorf("expected checksum to be set")
	}

	// Verify file exists
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Fatal("Output file was not created")
	}

	// Verify file content
	content, err := os.ReadFile(outputFile) // #nosec G304 - test file in controlled environment
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	if len(content) == 0 {
		t.Error("Output file is empty")
	}

	// Basic CSV validation
	contentStr := string(content)
	if !contains(contentStr, "test-report-2") {
		t.Error("Output should contain report ID")
	}
	if !contains(contentStr, "Test Cost Analysis CSV") {
		t.Error("Output should contain report title")
	}
	if !contains(contentStr, "2345.67") {
		t.Error("Output should contain total cost")
	}
	if !contains(contentStr, "EUR") {
		t.Error("Output should contain currency")
	}
}

func TestCSVExporter_UnsupportedFormat(t *testing.T) {
	exporter := NewCSVExporter()
	report := &types.CostAnalysisReport{ID: "test"}

	_, _, err := exporter.Export(context.Background(), report, types.ExportFormatJSON, "test.json")
	if err == nil {
		t.Error("Should return error for unsupported format")
	}
}

func TestYAMLExporter_Export(t *testing.T) {
	exporter := NewYAMLExporter()

	// Test supported formats
	formats := exporter.GetSupportedFormats()
	if len(formats) != 1 || formats[0] != types.ExportFormatYAML {
		t.Error("YAML exporter should only support YAML format")
	}

	// Create test report
	report := &types.UsageSummaryReport{
		ID:          "test-usage-report",
		Title:       "Test Usage Summary",
		GeneratedAt: time.Now(),
		ResourceUtilization: []types.ResourceUtilization{
			{
				ResourceID:   "i-1234567890abcdef0",
				ResourceType: "EC2",
				Provider:     "aws",
				Utilization:  75.5,
			},
		},
		Summary: types.ReportSummary{
			TotalItems: 5,
		},
	}

	// Create temporary file
	tempDir := t.TempDir()
	outputFile := filepath.Join(tempDir, "test_report.yaml")

	// Test export
	_, _, err := exporter.Export(context.Background(), report, types.ExportFormatYAML, outputFile)
	if err != nil {
		t.Fatalf("Failed to export YAML: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Fatal("Output file was not created")
	}

	// Verify file content
	content, err := os.ReadFile(outputFile) // #nosec G304 - test file in controlled environment
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	if len(content) == 0 {
		t.Error("Output file is empty")
	}

	// Basic YAML validation
	contentStr := string(content)
	if !contains(contentStr, "test-usage-report") {
		t.Error("Output should contain report ID")
	}
	if !contains(contentStr, "Test Usage Summary") {
		t.Error("Output should contain report title")
	}
	if !contains(contentStr, "i-1234567890abcdef0") {
		t.Error("Output should contain resource ID")
	}
}

func TestPDFExporter(t *testing.T) {
	ctx := context.Background()
	exp := NewPDFExporter()
	tmp := t.TempDir() + "/out.pdf"
	report := &types.ReportInfo{ID: "r1", Title: "PDF"}
	if _, _, err := exp.Export(ctx, report, types.ExportFormatPDF, tmp); err != nil {
		t.Fatalf("pdf export failed: %v", err)
	}
	b, err := os.ReadFile(tmp) // #nosec G304 test controlled path
	if err != nil {
		t.Fatalf("read pdf: %v", err)
	}
	if len(b) == 0 {
		t.Fatalf("empty pdf output")
	}
}

func TestExcelExporter(t *testing.T) {
	ctx := context.Background()
	exp := NewExcelExporter()
	tmp := t.TempDir() + "/out.xlsx"
	report := &types.ReportInfo{ID: "r2", Title: "Excel"}
	if _, _, err := exp.Export(ctx, report, types.ExportFormatExcel, tmp); err != nil {
		t.Fatalf("excel export failed: %v", err)
	}
	b, err := os.ReadFile(tmp) // #nosec G304 test controlled path
	if err != nil {
		t.Fatalf("read xlsx: %v", err)
	}
	if len(b) == 0 {
		t.Fatalf("empty xlsx output")
	}
}
func TestYAMLExporter_UnsupportedFormat(t *testing.T) {
	exporter := NewYAMLExporter()
	report := &types.UsageSummaryReport{ID: "test"}

	_, _, err := exporter.Export(context.Background(), report, types.ExportFormatCSV, "test.csv")
	if err == nil {
		t.Error("Should return error for unsupported format")
	}
}

func TestParquetExporter_Export(t *testing.T) {
	exporter := NewParquetExporter()

	// Supported formats
	formats := exporter.GetSupportedFormats()
	if len(formats) != 1 || formats[0] != types.ExportFormatParquet {
		t.Fatal("Parquet exporter should only support parquet format")
	}

	report := &types.ExecutiveSummaryReport{ID: "rep-1", Title: "Exec"}
	tmp := t.TempDir()
	out := filepath.Join(tmp, "rep.parquet")
	if _, _, err := exporter.Export(context.Background(), report, types.ExportFormatParquet, out); err != nil {
		t.Fatalf("parquet export: %v", err)
	}
	fi, err := os.Stat(out)
	if err != nil || fi.Size() == 0 {
		t.Fatalf("expected parquet file, err=%v size=%d", err, fi.Size())
	}
}

func TestHTTPExporter_PostAuth(t *testing.T) {
	// Test server capturing request
	var gotMethod, gotAPIKey, gotAuth string
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		gotMethod = r.Method
		gotAPIKey = r.Header.Get("X-API-Key")
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`ok`))
	}))
	defer srv.Close()

	exp := NewHTTPExporter()
	ctx := WithAPIKey(context.Background(), "k123")
	ctx = WithBearerToken(ctx, "t456")
	if _, _, err := exp.Export(ctx, map[string]any{"id": 1}, types.ExportFormatHTTP, srv.URL); err != nil {
		t.Fatalf("http export: %v", err)
	}
	if gotMethod != http.MethodPost || gotAPIKey != "k123" || gotAuth != "Bearer t456" {
		t.Fatalf("unexpected headers or method: %s %s %s", gotMethod, gotAPIKey, gotAuth)
	}
	if calls == 0 {
		t.Fatal("server was not called")
	}
}

func TestCSVExporter_DifferentReportTypes(t *testing.T) {
	exporter := NewCSVExporter()
	tempDir := t.TempDir()

	// Test different report types
	testCases := []struct {
		name   string
		report interface{}
		file   string
	}{
		{
			name: "UsageSummaryReport",
			report: &types.UsageSummaryReport{
				ID:    "usage-test",
				Title: "Usage Test",
			},
			file: "usage.csv",
		},
		{
			name: "TrendAnalysisReport",
			report: &types.TrendAnalysisReport{
				ID:    "trend-test",
				Title: "Trend Test",
			},
			file: "trend.csv",
		},
		{
			name: "AnomalyReport",
			report: &types.AnomalyReport{
				ID:        "anomaly-test",
				Title:     "Anomaly Test",
				RiskLevel: "medium",
			},
			file: "anomaly.csv",
		},
		{
			name: "ForecastReport",
			report: &types.ForecastReport{
				ID:         "forecast-test",
				Title:      "Forecast Test",
				Confidence: 0.85,
			},
			file: "forecast.csv",
		},
		{
			name: "ExecutiveSummaryReport",
			report: &types.ExecutiveSummaryReport{
				ID:    "exec-test",
				Title: "Executive Test",
				ExecutiveSummary: types.ExecutiveSummaryData{
					TotalSpend:    1000.00,
					SpendChange:   15.5,
					TopCostDriver: "EC2",
					RiskLevel:     "low",
				},
			},
			file: "exec.csv",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			outputFile := filepath.Join(tempDir, tc.file)

			_, _, err := exporter.Export(context.Background(), tc.report, types.ExportFormatCSV, outputFile)
			if err != nil {
				t.Fatalf("Failed to export %s: %v", tc.name, err)
			}

			// Verify file exists and has content
			if _, err := os.Stat(outputFile); os.IsNotExist(err) {
				t.Fatalf("Output file was not created for %s", tc.name)
			}

			content, err := os.ReadFile(outputFile) // #nosec G304 - test file in controlled environment
			if err != nil {
				t.Fatalf("Failed to read output file for %s: %v", tc.name, err)
			}

			if len(content) == 0 {
				t.Errorf("Output file is empty for %s", tc.name)
			}
		})
	}
}

func TestCSVExporter_GenericStruct(t *testing.T) {
	exporter := NewCSVExporter()

	// Test with a generic struct (not a known report type)
	type GenericReport struct {
		ID     string  `json:"id"`
		Name   string  `json:"name"`
		Value  float64 `json:"value"`
		Active bool    `json:"active"`
	}

	report := &GenericReport{
		ID:     "generic-1",
		Name:   "Generic Test",
		Value:  123.45,
		Active: true,
	}

	tempDir := t.TempDir()
	outputFile := filepath.Join(tempDir, "generic.csv")

	_, _, err := exporter.Export(context.Background(), report, types.ExportFormatCSV, outputFile)
	if err != nil {
		t.Fatalf("Failed to export generic report: %v", err)
	}

	// Verify file content
	content, err := os.ReadFile(outputFile) // #nosec G304 - test file in controlled environment
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	contentStr := string(content)
	if !contains(contentStr, "ID") || !contains(contentStr, "Name") {
		t.Error("Output should contain field headers")
	}
	if !contains(contentStr, "generic-1") || !contains(contentStr, "Generic Test") {
		t.Error("Output should contain field values")
	}
}

func TestJSONExporter_Golden(t *testing.T) {
	exp := NewJSONExporter()
	tmp := t.TempDir()
	out := filepath.Join(tmp, "golden.json")
	// Build deterministic report matching fixture
	t0, _ := time.Parse(time.RFC3339, "2000-01-01T00:00:00Z")
	t1, _ := time.Parse(time.RFC3339, "2000-01-31T00:00:00Z")
	r := &types.CostAnalysisReport{
		ID:          "golden-1",
		Title:       "Golden Cost Analysis",
		Description: "Fixture",
		GeneratedAt: t0,
		DateRange:   types.DateRange{StartDate: t0, EndDate: t1},
		TotalCost:   100.5,
		Currency:    "USD",
		// Ensure empty arrays, not nulls
		CostByService:  []types.ServiceCostBreakdown{},
		CostByRegion:   []types.RegionCostBreakdown{},
		CostByAccount:  []types.AccountCostBreakdown{},
		CostTrends:     []types.CostTrendData{},
		TopCostDrivers: []types.CostDriver{},
		Optimization:   []types.OptimizationRecommendation{},
		Summary:        types.ReportSummary{TotalItems: 0, TotalCost: 100.5, Currency: "USD", DateRange: types.DateRange{StartDate: t0, EndDate: t1}},
	}
	if _, _, err := exp.Export(context.Background(), r, types.ExportFormatJSON, out); err != nil {
		t.Fatalf("export: %v", err)
	}
	got, err := os.ReadFile(out) // #nosec G304 - test file in controlled environment
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	// Prefer embedded golden JSON to avoid fs path issues in containerized act runs.
	var want []byte
	if len(goldenFixture) > 0 {
		want = goldenFixture
	} else {
		_, file, _, ok := runtime.Caller(0)
		if !ok {
			t.Fatal("runtime caller unavailable")
		}
		// Prefer repo-level fixture if present; fall back to package-local testdata copy.
		baseDir := filepath.Dir(file)
		candRepo := filepath.Join(baseDir, "..", "..", "..", "..", "tests", "fixtures", "reports", "golden.json")
		candPkg := filepath.Join(baseDir, "testdata", "golden.json")
		candCWD := filepath.Join("tests", "fixtures", "reports", "golden.json")
		fixture := ""
		for _, p := range []string{candRepo, candPkg, candCWD} {
			if _, err := os.Stat(p); err == nil {
				fixture = p
				break
			}
		}
		if fixture == "" {
			t.Fatalf("golden fixture not found; tried: %s, %s, %s", candRepo, candPkg, candCWD)
		}
		var err error
		want, err = os.ReadFile(fixture) // #nosec G304 - controlled test path
		if err != nil {
			t.Fatalf("read golden: %v", err)
		}
	}
	// Canonicalize both JSON documents before comparing
	var gotV any
	var wantV any
	if err := json.Unmarshal(got, &gotV); err != nil {
		t.Fatalf("unmarshal got: %v", err)
	}
	if err := json.Unmarshal(want, &wantV); err != nil {
		t.Fatalf("unmarshal want: %v", err)
	}
	gotNorm, _ := json.MarshalIndent(gotV, "", "  ")
	wantNorm, _ := json.MarshalIndent(wantV, "", "  ")
	if !bytes.Equal(bytes.TrimSpace(gotNorm), bytes.TrimSpace(wantNorm)) {
		t.Fatalf("golden mismatch\nGOT:\n%s\nWANT:\n%s", string(gotNorm), string(wantNorm))
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && containsAt(s, substr, 0)))
}

func containsAt(s, substr string, start int) bool {
	if start < 0 || start > len(s)-len(substr) {
		return false
	}
	for i := 0; i < len(substr); i++ {
		if s[start+i] != substr[i] {
			if start+1 <= len(s)-len(substr) {
				return containsAt(s, substr, start+1)
			}
			return false
		}
	}
	return true
}
