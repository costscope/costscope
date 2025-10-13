package analytics

import (
	"path/filepath"
	"testing"

	"github.com/costscope/costscope/cmd/modules/analytics/types"
	"github.com/costscope/costscope/internal/core/logging"
)

func TestNewBasicService(t *testing.T) {
	config := &Config{
		MLEnabled:         true,
		AnomalyDetection:  true,
		TrendAnalysis:     true,
		EnablePredictions: true,
	}
	logger := logging.NewLogger("info")

	service := NewBasicService(config, logger)
	if service == nil {
		t.Fatal("NewBasicService returned nil")
	}

	if service.config != config {
		t.Error("Config not set correctly")
	}

	if service.logger != logger {
		t.Error("Logger not set correctly")
	}
}

func TestAnalyze(t *testing.T) {
	config := &Config{
		MLEnabled:         true,
		AnomalyDetection:  true,
		TrendAnalysis:     true,
		EnablePredictions: true,
	}
	logger := logging.NewLogger("info")
	service := NewBasicService(config, logger)

	opts := &types.AnalyticsOptions{
		TableName:     "test_table",
		Currency:      "USD",
		EnableML:      true,
		Filters:       map[string]interface{}{"region": "us-east-1"},
		GroupByFields: []string{"service", "region"},
		SortOrder:     "desc",
	}

	results, err := service.Analyze(opts)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	if results == nil {
		t.Fatal("Results should not be nil")
	}

	if results.AnalyticsType != "basic_analysis" {
		t.Errorf("Expected analytics_type 'basic_analysis', got %s", results.AnalyticsType)
	}

	if results.TableName != opts.TableName {
		t.Errorf("Expected table_name %s, got %s", opts.TableName, results.TableName)
	}

	if results.FiltersCount != len(opts.Filters) {
		t.Errorf("Expected filters_count %d, got %d", len(opts.Filters), results.FiltersCount)
	}

	// Check ML insights are included when ML is enabled
	if mlInsights, exists := results.AnalysisResult["ml_insights"]; !exists {
		t.Error("ML insights should be included when ML is enabled")
	} else {
		insights := mlInsights.(map[string]interface{})
		if insights["confidence_score"] == nil {
			t.Error("ML insights should include confidence_score")
		}
	}
}

func TestAnalyzeWithoutML(t *testing.T) {
	config := &Config{
		MLEnabled: false, // ML disabled
	}
	logger := logging.NewLogger("info")
	service := NewBasicService(config, logger)

	opts := &types.AnalyticsOptions{
		TableName: "test_table",
		Currency:  "USD",
		EnableML:  true, // User wants ML but it's disabled in config
		Filters:   map[string]interface{}{"region": "us-east-1"},
	}

	results, err := service.Analyze(opts)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	// ML insights should not be included when ML is disabled in config
	if _, exists := results.AnalysisResult["ml_insights"]; exists {
		t.Error("ML insights should not be included when ML is disabled in config")
	}
}

func TestForecast(t *testing.T) {
	config := &Config{
		MLEnabled:         true,
		AnomalyDetection:  true,
		TrendAnalysis:     true,
		EnablePredictions: true,
	}
	logger := logging.NewLogger("info")
	service := NewBasicService(config, logger)

	opts := &types.AnalyticsOptions{
		TableName:              "test_table",
		Currency:               "USD",
		EnableML:               true,
		ForecastDays:           30,
		EnableAnomalyDetection: true,
		Filters:                map[string]interface{}{"region": "us-east-1"},
	}

	results, err := service.Forecast(opts)
	if err != nil {
		t.Fatalf("Forecast failed: %v", err)
	}

	if results == nil {
		t.Fatal("Results should not be nil")
	}

	if results.AnalyticsType != "ml_forecast" {
		t.Errorf("Expected analytics_type 'ml_forecast', got %s", results.AnalyticsType)
	}

	// Check forecast result structure
	if results.ForecastResult == nil {
		t.Fatal("ForecastResult should not be nil")
	}

	requiredFields := []string{"forecast_days", "predicted_cost", "confidence_lower", "confidence_upper", "trend_analysis"}
	for _, field := range requiredFields {
		if _, exists := results.ForecastResult[field]; !exists {
			t.Errorf("ForecastResult should include %s", field)
		}
	}

	// Check anomaly detection is included
	if _, exists := results.ForecastResult["anomaly_detection"]; !exists {
		t.Error("Anomaly detection should be included when enabled")
	}
}

func TestForecastMLDisabled(t *testing.T) {
	config := &Config{
		MLEnabled: false, // ML disabled
	}
	logger := logging.NewLogger("info")
	service := NewBasicService(config, logger)

	opts := &types.AnalyticsOptions{
		TableName:    "test_table",
		EnableML:     true, // User wants ML but it's disabled
		ForecastDays: 30,
	}

	_, err := service.Forecast(opts)
	if err == nil {
		t.Error("Forecast should fail when ML is disabled")
	}

	expectedError := "ML forecasting is disabled"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestCompare(t *testing.T) {
	config := &Config{
		MLEnabled: true,
	}
	logger := logging.NewLogger("info")
	service := NewBasicService(config, logger)

	opts := &types.AnalyticsOptions{
		TableName: "test_table",
		Currency:  "USD",
		Filters:   map[string]interface{}{"region": "us-east-1"},
	}

	period := "previous-month"
	providers := []string{"aws", "azure", "gcp"}

	results, err := service.Compare(opts, period, providers)
	if err != nil {
		t.Fatalf("Compare failed: %v", err)
	}

	if results == nil {
		t.Fatal("Results should not be nil")
	}

	if results.AnalyticsType != "cost_comparison" {
		t.Errorf("Expected analytics_type 'cost_comparison', got %s", results.AnalyticsType)
	}

	// Check comparison result structure
	if results.ComparisonResult == nil {
		t.Fatal("ComparisonResult should not be nil")
	}

	requiredFields := []string{"comparison_period", "current_cost", "previous_cost", "change_percent", "change_direction"}
	for _, field := range requiredFields {
		if _, exists := results.ComparisonResult[field]; !exists {
			t.Errorf("ComparisonResult should include %s", field)
		}
	}

	// Check provider comparison is included
	if providerCosts, exists := results.ComparisonResult["provider_costs"]; !exists {
		t.Error("Provider costs should be included when providers are specified")
	} else {
		costs := providerCosts.(map[string]float64)
		for _, provider := range providers {
			if _, exists := costs[provider]; !exists {
				t.Errorf("Provider %s should be included in costs", provider)
			}
		}
	}
}

func TestExportJSON(t *testing.T) {
	config := &Config{
		MLEnabled: true,
	}
	logger := logging.NewLogger("info")
	service := NewBasicService(config, logger)

	opts := &types.AnalyticsOptions{
		TableName: "test_table",
		Currency:  "USD",
		Filters:   map[string]interface{}{"region": "us-east-1"},
	}

	// Test JSON export with temporary file
	outputPath, err := service.Export(opts, "json", "")
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	if outputPath == "" {
		t.Error("Output path should not be empty")
	}

	// Check file extension
	if filepath.Ext(outputPath) != ".json" {
		t.Errorf("Expected .json extension, got %s", filepath.Ext(outputPath))
	}

	// Clean up - in real test would check file contents
	// os.Remove(outputPath)
}

func TestExportCSV(t *testing.T) {
	config := &Config{
		MLEnabled: true,
	}
	logger := logging.NewLogger("info")
	service := NewBasicService(config, logger)

	opts := &types.AnalyticsOptions{
		TableName: "test_table",
		Currency:  "USD",
		Filters:   map[string]interface{}{"region": "us-east-1"},
	}

	// Test CSV export
	outputPath, err := service.Export(opts, "csv", "")
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	if filepath.Ext(outputPath) != ".csv" {
		t.Errorf("Expected .csv extension, got %s", filepath.Ext(outputPath))
	}
}

func TestExportUnsupportedFormat(t *testing.T) {
	config := &Config{
		MLEnabled: true,
	}
	logger := logging.NewLogger("info")
	service := NewBasicService(config, logger)

	opts := &types.AnalyticsOptions{
		TableName: "test_table",
		Currency:  "USD",
	}

	// Test unsupported format
	_, err := service.Export(opts, "unsupported", "")
	if err == nil {
		t.Error("Export should fail for unsupported format")
	}

	expectedError := "unsupported export format: unsupported"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}
