package commands

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"local/costscope/internal/core/logging"
	"local/costscope/internal/core/reports"
	rtypes "local/costscope/internal/core/reports/types"
)

// TestReportsExportHTTP_E2E spins up a test HTTP server and drives the Cobra command
// `reports export --format http` end-to-end, asserting headers and payload.
func TestReportsExportHTTP_E2E(t *testing.T) {
	t.Setenv("COSTSCOPE_HTTP_BEARER_TOKEN", "e2e-bearer")
	t.Setenv("COSTSCOPE_HTTP_API_KEY", "e2e-apikey")

	// Start a capture server
	type payload struct {
		ID         string            `json:"id"`
		ReportType rtypes.ReportType `json:"report_type"`
		Title      string            `json:"title"`
	}
	got := struct {
		Method string
		Auth   string
		APIKey string
		Body   payload
	}{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got.Method = r.Method
		got.Auth = r.Header.Get("Authorization")
		got.APIKey = r.Header.Get("X-API-Key")
		defer func() { _ = r.Body.Close() }()
		if err := json.NewDecoder(r.Body).Decode(&got.Body); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	// Prepare service and generate a report
	logger := logging.NewLogger("warn")
	svc := reports.NewBasicReportService(logger)
	opts := &rtypes.ReportOptions{
		ReportType: rtypes.ReportTypeCostAnalysis,
		Title:      "E2E Cost Analysis",
		DateRange:  rtypes.DateRange{StartDate: time.Now().AddDate(0, -1, 0), EndDate: time.Now()},
	}
	rep, err := svc.GenerateCostAnalysisReport(context.Background(), opts)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	// Build the reports command and execute export via HTTP
	cmds := NewReportsCommands(svc, logger)
	root := cmds.BuildReportsCommand()
	// Simulate: reports export <id> --format http --output <url>
	root.SetArgs([]string{"export", rep.ID, "--format", "http", "--output", srv.URL})
	if err := root.Execute(); err != nil {
		t.Fatalf("execute export: %v", err)
	}

	// Assertions
	if got.Method != http.MethodPost {
		t.Fatalf("expected POST, got %s", got.Method)
	}
	if got.Auth != "Bearer e2e-bearer" {
		t.Fatalf("authorization header mismatch: %q", got.Auth)
	}
	if got.APIKey != "e2e-apikey" {
		t.Fatalf("api key header mismatch: %q", got.APIKey)
	}
	if got.Body.ID != rep.ID {
		t.Fatalf("payload id mismatch: got %s want %s", got.Body.ID, rep.ID)
	}
	if got.Body.ReportType == "" {
		t.Fatalf("payload report_type should not be empty")
	}
}

func TestReportsExportHTTP_E2E_WithIncludeContent(t *testing.T) {
	t.Setenv("COSTSCOPE_HTTP_BEARER_TOKEN", "e2e-bearer2")
	t.Setenv("COSTSCOPE_HTTP_API_KEY", "e2e-apikey2")

	type payload struct {
		ID         string            `json:"id"`
		ReportType rtypes.ReportType `json:"report_type"`
		Title      string            `json:"title"`
	}
	got := struct {
		Method string
		Auth   string
		APIKey string
		Body   payload
	}{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got.Method = r.Method
		got.Auth = r.Header.Get("Authorization")
		got.APIKey = r.Header.Get("X-API-Key")
		defer func() { _ = r.Body.Close() }()
		_ = json.NewDecoder(r.Body).Decode(&got.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	logger := logging.NewLogger("warn")
	svc := reports.NewBasicReportService(logger)
	rep, err := svc.GenerateCostAnalysisReport(context.Background(), &rtypes.ReportOptions{
		ReportType: rtypes.ReportTypeCostAnalysis,
		Title:      "E2E Include Content",
		DateRange:  rtypes.DateRange{StartDate: time.Now().Add(-24 * time.Hour), EndDate: time.Now()},
	})
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	cmds := NewReportsCommands(svc, logger)
	root := cmds.BuildReportsCommand()
	root.SetArgs([]string{"export", rep.ID, "--format", "http", "--output", srv.URL, "--include-content"})
	if err := root.Execute(); err != nil {
		t.Fatalf("execute export: %v", err)
	}

	if got.Auth != "Bearer e2e-bearer2" || got.APIKey != "e2e-apikey2" {
		t.Fatal("headers mismatch")
	}
	if got.Body.ID != rep.ID {
		t.Fatalf("id mismatch: %s", got.Body.ID)
	}
}
