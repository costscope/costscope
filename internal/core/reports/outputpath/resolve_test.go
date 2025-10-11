package outputpath

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"local/costscope/internal/core/logging"
	rtypes "local/costscope/internal/core/reports/types"
)

func TestResolveOutputPath_Table(t *testing.T) {
	tmp := t.TempDir()
	cases := []struct {
		name       string
		base       string
		explicit   string
		format     rtypes.ExportFormat
		env        string
		assertFunc func(t *testing.T, out string, base string, explicit string)
	}{
		{
			name:   "auto local base provided",
			base:   filepath.Join(tmp, "reports"),
			format: rtypes.ExportFormatJSON,
			assertFunc: func(t *testing.T, out, base, _ string) {
				if !strings.HasPrefix(out, base) || !strings.HasSuffix(out, ".json") {
					t.Fatalf("unexpected path: %s", out)
				}
				if _, err := os.Stat(base); err != nil {
					t.Fatalf("expected dir created: %v", err)
				}
			},
		},
		{
			name:   "env override base",
			env:    filepath.Join(tmp, "env-reports"),
			format: rtypes.ExportFormatParquet,
			assertFunc: func(t *testing.T, out, base, _ string) {
				if !strings.HasPrefix(out, base) || !strings.HasSuffix(out, ".parquet") {
					t.Fatalf("bad env path: %s", out)
				}
			},
		},
		{
			name:   "object store prefix",
			base:   "s3://bucket/path",
			format: rtypes.ExportFormatParquet,
			assertFunc: func(t *testing.T, out, base, _ string) {
				if !strings.HasPrefix(out, base+"/") {
					t.Fatalf("s3 prefix wrong: %s", out)
				}
			},
		},
		{
			name:     "explicit local adds extension",
			explicit: filepath.Join(tmp, "myreport"),
			format:   rtypes.ExportFormatJSON,
			assertFunc: func(t *testing.T, out, _, explicit string) {
				if out != explicit+".json" {
					t.Fatalf("expected %s.json got %s", explicit, out)
				}
			},
		},
		{
			name:   "filename pattern",
			base:   filepath.Join(tmp, "pattern"),
			format: rtypes.ExportFormatJSON,
			assertFunc: func(t *testing.T, out, _, _ string) {
				re := regexp.MustCompile(`\d{8}-\d{6}-report\.json$`)
				if !re.MatchString(filepath.Base(out)) {
					t.Fatalf("pattern mismatch: %s", out)
				}
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if c.env != "" {
				if err := os.Setenv("COSTSCOPE_REPORTS_DIR", c.env); err != nil {
					t.Fatalf("failed to set env: %v", err)
				}
				defer func() {
					if err := os.Unsetenv("COSTSCOPE_REPORTS_DIR"); err != nil {
						t.Fatalf("failed to unset env: %v", err)
					}
				}()
			}
			out, err := ResolveOutputPath(c.base, c.explicit, c.format)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			base := c.base
			if base == "" && c.env != "" {
				base = c.env
			}
			c.assertFunc(t, out, base, c.explicit)
		})
	}
}

// Test that a single structured log line `config_precedence_resolved` is emitted
// when resolving reports.output_dir and that default policy points under costscope-data/
func TestResolveOutputPath_LogsConfigPrecedenceAndDefault(t *testing.T) {
	// Capture stderr (logger output)
	old := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stderr = w
	defer func() { os.Stderr = old }()

	// Ensure no env override
	if err := os.Unsetenv("COSTSCOPE_REPORTS_DIR"); err != nil {
		t.Fatalf("unset env: %v", err)
	}

	// Call with no explicit base; should use default costscope-data/reports
	out, err := ResolveOutputPath("", "", rtypes.ExportFormatJSON)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if !strings.Contains(out, "costscope-data/reports/") || !strings.HasSuffix(out, ".json") {
		t.Fatalf("expected default under costscope-data/reports, got %s", out)
	}

	// Close writer and inspect a single log line
	if err := w.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf("read: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) < 1 {
		t.Fatalf("expected at least 1 log line, got %d", len(lines))
	}
	// find the config_precedence_resolved line
	var found map[string]any
	for _, ln := range lines {
		var m map[string]any
		if err := json.Unmarshal([]byte(ln), &m); err == nil {
			if m["msg"] == "config_precedence_resolved" {
				found = m
				break
			}
		}
	}
	if found == nil {
		t.Fatalf("did not find config_precedence_resolved log line in: %v", lines)
	}
	if got := found["field"]; got != "reports.output_dir" {
		t.Fatalf("field mismatch: %v", got)
	}
	if got := found["level"]; got != string(logging.LevelInfo) {
		t.Fatalf("level mismatch: %v", got)
	}
}

// YAML precedence should override ENV for reports.output_dir and log source="yaml".
func TestResolveOutputPath_YAMLPrecedence_OverridesEnv_AndLogsSource(t *testing.T) {
	tmp := t.TempDir()
	yamlPath := filepath.Join(tmp, "config.yaml")
	yamlReportsDir := filepath.Join(tmp, "yaml-reports")

	// Write minimal YAML config with reports.output_dir
	content := []byte("reports:\n  output_dir: " + strings.ReplaceAll(yamlReportsDir, "\\", "/") + "\n")
	if err := os.WriteFile(yamlPath, content, 0o600); err != nil {
		t.Fatalf("write yaml: %v", err)
	}

	// Set ENV that should be overridden by YAML
	envDir := filepath.Join(tmp, "env-reports")
	if err := os.Setenv("COSTSCOPE_REPORTS_DIR", envDir); err != nil {
		t.Fatalf("set env: %v", err)
	}
	defer func() {
		if err := os.Unsetenv("COSTSCOPE_REPORTS_DIR"); err != nil {
			t.Fatalf("unset env: %v", err)
		}
	}()

	// Point loader to our YAML file
	if err := os.Setenv("COSTSCOPE_CONFIG", yamlPath); err != nil {
		t.Fatalf("set COSTSCOPE_CONFIG: %v", err)
	}
	defer func() {
		if err := os.Unsetenv("COSTSCOPE_CONFIG"); err != nil {
			t.Fatalf("unset COSTSCOPE_CONFIG: %v", err)
		}
	}()

	// Capture stderr for structured logs
	old := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stderr = w
	defer func() { os.Stderr = old }()

	// Resolve with no explicit inputs; YAML should win over ENV
	out, err := ResolveOutputPath("", "", rtypes.ExportFormatJSON)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if !strings.HasPrefix(out, yamlReportsDir) || !strings.HasSuffix(out, ".json") {
		t.Fatalf("expected YAML reports dir prefix %s, got %s", yamlReportsDir, out)
	}

	// Close writer and inspect logs for source: yaml
	if err := w.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf("read: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	var found map[string]any
	for _, ln := range lines {
		var m map[string]any
		if err := json.Unmarshal([]byte(ln), &m); err == nil {
			if m["msg"] == "config_precedence_resolved" && m["field"] == "reports.output_dir" {
				found = m
				break
			}
		}
	}
	if found == nil {
		t.Fatalf("did not find config_precedence_resolved log line for reports.output_dir")
	}
	if got := found["source"]; got != "yaml" {
		t.Fatalf("expected source 'yaml', got %v (line: %v)", got, found)
	}
	// Value should match our YAML path (normalized)
	if got := found["value"]; got != strings.ReplaceAll(yamlReportsDir, "\\", "/") {
		t.Fatalf("expected value %s, got %v", yamlReportsDir, got)
	}
}
