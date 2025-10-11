package commands

import (
	"testing"

	"local/costscope/internal/core/logging"
	"local/costscope/internal/core/reports"
)

func TestNewReportsCommands(t *testing.T) {
	logger := logging.NewLogger("info")
	reportService := reports.NewBasicReportService(logger)
	commands := NewReportsCommands(reportService, logger)

	if commands == nil {
		t.Fatal("NewReportsCommands returned nil")
	}

	if commands.logger == nil {
		t.Error("Logger should not be nil")
	}

	if commands.reportService == nil {
		t.Error("ReportService should not be nil")
	}
}

func TestBuildReportsCommand(t *testing.T) {
	logger := logging.NewLogger("info")
	reportService := reports.NewBasicReportService(logger)
	commands := NewReportsCommands(reportService, logger)

	cmd := commands.BuildReportsCommand()
	if cmd == nil {
		t.Fatal("BuildReportsCommand returned nil")
	}

	if cmd.Use != "reports" {
		t.Errorf("Expected command use to be 'reports', got '%s'", cmd.Use)
	}
}
