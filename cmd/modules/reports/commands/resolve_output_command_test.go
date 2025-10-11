package commands

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"

	"local/costscope/internal/core/logging"
	"local/costscope/internal/core/monitoring/telemetry"
	"local/costscope/internal/core/reports"
)

// execCommand runs the root command with args capturing stdout
func execCommand(t *testing.T, root *cobra.Command, args ...string) (string, error) {
	t.Helper()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)
	_, err := root.ExecuteC()
	return strings.TrimSpace(buf.String()), err
}

func newReportsRoot(t *testing.T) *cobra.Command { // nolint:unparam // test helper signature uniform across packages
	// Ensure metrics registered once; ignoring duplicate registration panics by letting
	// first call succeed (subsequent tests in package reuse process). Safe because
	// prometheus client returns panic on duplicate registration; we guard with recover.
	func() {
		defer func() { _ = recover() }()
		telemetry.Register()
	}()
	logger := logging.NewLogger(logging.LevelError)
	svc := reports.NewBasicReportService(logger)
	cmds := NewReportsCommands(svc, logger)
	return cmds.BuildReportsCommand()
}

func TestReportsResolveOutputExplicitFile(t *testing.T) {
	root := newReportsRoot(t)
	tmp := t.TempDir()
	outPath := filepath.Join(tmp, "custom_name")
	got, err := execCommand(t, root, "resolve-output", "--output", outPath, "--format", "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(got, outPath) || !strings.HasSuffix(got, ".json") {
		t.Fatalf("expected explicit path with extension, got %s", got)
	}
}

func TestReportsResolveOutputExplicitBaseDir(t *testing.T) {
	root := newReportsRoot(t)
	base := t.TempDir()
	got, err := execCommand(t, root, "resolve-output", "--base-dir", base, "--format", "csv")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(got, base) || !strings.HasSuffix(got, ".csv") {
		t.Fatalf("expected path in base dir with .csv suffix, got %s", got)
	}
}

func TestReportsResolveOutputEnvDir(t *testing.T) {
	root := newReportsRoot(t)
	envDir := t.TempDir()
	if err := os.Setenv("COSTSCOPE_REPORTS_DIR", envDir); err != nil {
		t.Fatalf("set env: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Unsetenv("COSTSCOPE_REPORTS_DIR"); err != nil {
			t.Fatalf("unset env: %v", err)
		}
	})
	got, err := execCommand(t, root, "resolve-output", "--format", "yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(got, envDir) || !strings.HasSuffix(got, ".yaml") {
		t.Fatalf("expected env dir path with .yaml suffix, got %s", got)
	}
}

func TestReportsResolveOutputYAMLConfig(t *testing.T) {
	root := newReportsRoot(t)
	cfgDir := t.TempDir()
	reportsDir := filepath.Join(cfgDir, "reports-out")
	cfgPath := filepath.Join(cfgDir, "config.yaml")
	content := []byte("reports:\n  output_dir: " + reportsDir + "\n")
	if err := os.WriteFile(cfgPath, content, 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	if err := os.Setenv("COSTSCOPE_CONFIG", cfgPath); err != nil {
		t.Fatalf("set env: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Unsetenv("COSTSCOPE_CONFIG"); err != nil {
			t.Fatalf("unset env: %v", err)
		}
	})
	got, err := execCommand(t, root, "resolve-output", "--format", "parquet")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(got, reportsDir) || !strings.HasSuffix(got, ".parquet") {
		t.Fatalf("expected yaml dir precedence with .parquet suffix, got %s", got)
	}
}

func TestReportsResolveOutputDefault(t *testing.T) {
	root := newReportsRoot(t)
	got, err := execCommand(t, root, "resolve-output")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(got, "costscope-data/reports") {
		t.Fatalf("expected default directory in path, got %s", got)
	}
}

func TestReportGenerateDryRunOutputResolution(t *testing.T) {
	root := newReportsRoot(t)
	base := t.TempDir()
	got, err := execCommand(t, root, "generate", "cost-analysis", "--dry-run-output-resolution", "--format", "csv", "--base-dir", base)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(got, base) || !strings.HasSuffix(got, ".csv") {
		t.Fatalf("expected resolved csv path in base dir, got %s", got)
	}
	if strings.Contains(got, "Report ID:") {
		t.Fatalf("dry run should not generate report content, got %s", got)
	}
}

func TestReportsResolveOutputTimestamped(t *testing.T) {
	root := newReportsRoot(t)
	got, err := execCommand(t, root, "resolve-output")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	now := time.Now().UTC().Format("20060102-")
	if !strings.Contains(got, now) {
		t.Fatalf("expected timestamp prefix %s in %s", now, got)
	}
}

func TestReportsResolveCollisionMetric(t *testing.T) {
	root := newReportsRoot(t)
	dir := t.TempDir()
	file := filepath.Join(dir, "existing.json")
	if err := os.WriteFile(file, []byte("abc"), 0o600); err != nil {
		t.Fatalf("prep: %v", err)
	}
	_, err := execCommand(t, root, "resolve-output", "--output", file, "--format", "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rw := httptest.NewRecorder()
	promhttp.Handler().ServeHTTP(rw, req)
	body := rw.Body.String()
	if !strings.Contains(body, "costscope_reports_resolve_output_collisions_total") {
		if len(body) > 200 {
			body = body[:200]
		}
		t.Fatalf("expected collisions metric, metrics: %s", body)
	}
}
