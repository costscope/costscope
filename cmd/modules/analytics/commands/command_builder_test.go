package commands

import (
	"bytes"
	"testing"

	"github.com/costscope/costscope/internal/core/analytics"
	"github.com/costscope/costscope/internal/core/logging"
)

func TestNewAnalyticsCommands(t *testing.T) {
	logger := logging.NewLogger("info")
	config := &analytics.Config{
		MLEnabled:         true,
		AnomalyDetection:  true,
		TrendAnalysis:     true,
		EnablePredictions: true,
	}
	service := analytics.NewBasicService(config, logger)

	commands := NewAnalyticsCommands(logger, service)
	if commands == nil {
		t.Fatal("NewAnalyticsCommands returned nil")
	}

	if commands.logger != logger {
		t.Error("Logger not set correctly")
	}

	if commands.service != service {
		t.Error("Service not set correctly")
	}
}

func TestBuildAnalyticsCommand(t *testing.T) {
	logger := logging.NewLogger("info")
	config := &analytics.Config{
		MLEnabled:         true,
		AnomalyDetection:  true,
		TrendAnalysis:     true,
		EnablePredictions: true,
	}
	service := analytics.NewBasicService(config, logger)
	commands := NewAnalyticsCommands(logger, service)

	cmd := commands.BuildAnalyticsCommand()
	if cmd == nil {
		t.Fatal("BuildAnalyticsCommand returned nil")
	}

	if cmd.Use != "analytics" {
		t.Errorf("Expected Use to be 'analytics', got %s", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Short description should not be empty")
	}

	if cmd.Long == "" {
		t.Error("Long description should not be empty")
	}

	// Snapshot check: help text must stay stable
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--help"})
	_ = cmd.Execute()
	got := buf.String()
	if got == "" {
		t.Fatal("expected help output, got empty")
	}
	if !contains(got, "Advanced cost analytics") {
		t.Errorf("help output missing expected phrase, got: %q", got)
	}
}

func TestAnalyzeCommand(t *testing.T) {
	logger := logging.NewLogger("info")
	config := &analytics.Config{
		MLEnabled:         true,
		AnomalyDetection:  true,
		TrendAnalysis:     true,
		EnablePredictions: true,
	}
	service := analytics.NewBasicService(config, logger)
	commands := NewAnalyticsCommands(logger, service)

	cmd := commands.BuildAnalyticsCommand()
	cmd, _, _ = cmd.Find([]string{"analyze"})
	if cmd == nil {
		t.Fatal("buildAnalyzeCommand returned nil")
	}

	if cmd.Use != "analyze" {
		t.Errorf("Expected Use to be 'analyze', got %s", cmd.Use)
	}

	// Check that required flags are present (generated)
	requiredFlags := []string{"table", "filters", "currency", "group-by", "sort-order"}
	for _, flagName := range requiredFlags {
		flag := cmd.Flags().Lookup(flagName)
		if flag == nil {
			t.Errorf("Required flag %s not found", flagName)
		}
	}
}

func TestForecastCommand(t *testing.T) {
	logger := logging.NewLogger("info")
	config := &analytics.Config{
		MLEnabled:         true,
		AnomalyDetection:  true,
		TrendAnalysis:     true,
		EnablePredictions: true,
	}
	service := analytics.NewBasicService(config, logger)
	commands := NewAnalyticsCommands(logger, service)

	root := commands.BuildAnalyticsCommand()
	cmd, _, _ := root.Find([]string{"forecast"})
	if cmd == nil {
		t.Fatal("buildForecastCommand returned nil")
	}

	if cmd.Use != "forecast" {
		t.Errorf("Expected Use to be 'forecast', got %s", cmd.Use)
	}

	// Check forecast-specific flags
	forecastFlags := []string{"forecast-days", "include-trends", "detect-anomalies"}
	for _, flagName := range forecastFlags {
		flag := cmd.Flags().Lookup(flagName)
		if flag == nil {
			t.Errorf("Forecast flag %s not found", flagName)
		}
	}
}

func TestCompareCommand(t *testing.T) {
	logger := logging.NewLogger("info")
	config := &analytics.Config{
		MLEnabled:         true,
		AnomalyDetection:  true,
		TrendAnalysis:     true,
		EnablePredictions: true,
	}
	service := analytics.NewBasicService(config, logger)
	commands := NewAnalyticsCommands(logger, service)

	root := commands.BuildAnalyticsCommand()
	cmd, _, _ := root.Find([]string{"compare"})
	if cmd == nil {
		t.Fatal("buildCompareCommand returned nil")
	}

	if cmd.Use != "compare" {
		t.Errorf("Expected Use to be 'compare', got %s", cmd.Use)
	}

	// Check compare-specific flags
	compareFlags := []string{"compare-period", "compare-providers"}
	for _, flagName := range compareFlags {
		flag := cmd.Flags().Lookup(flagName)
		if flag == nil {
			t.Errorf("Compare flag %s not found", flagName)
		}
	}
}

func TestExportCommand(t *testing.T) {
	logger := logging.NewLogger("info")
	config := &analytics.Config{
		MLEnabled:         true,
		AnomalyDetection:  true,
		TrendAnalysis:     true,
		EnablePredictions: true,
	}
	service := analytics.NewBasicService(config, logger)
	commands := NewAnalyticsCommands(logger, service)

	root := commands.BuildAnalyticsCommand()
	cmd, _, _ := root.Find([]string{"export"})
	if cmd == nil {
		t.Fatal("buildExportCommand returned nil")
	}

	if cmd.Use != "export" {
		t.Errorf("Expected Use to be 'export', got %s", cmd.Use)
	}

	// Check export-specific flags
	exportFlags := []string{"format", "output"}
	for _, flagName := range exportFlags {
		flag := cmd.Flags().Lookup(flagName)
		if flag == nil {
			t.Errorf("Export flag %s not found", flagName)
		}
	}
}

// shared tiny helper
func contains(s, sub string) bool { return bytes.Contains([]byte(s), []byte(sub)) }
