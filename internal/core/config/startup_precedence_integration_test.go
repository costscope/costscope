package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/costscope/costscope/internal/core/logging"
)

// TestConfigPrecedenceIntegration exercises a mini end-to-end flow:
//  1. Start from a YAML-populated ConsolidatedConfig (simulated via struct fields)
//  2. Apply env overrides (t.Setenv)
//  3. Resolve several fields with explicit / yaml / env / fallback precedence
//  4. Run ValidateAllConfig + EnsureConfigDirectories to confirm the finalized config is valid
//  5. Assert precedence source ordering and value selection semantics (incl. empty explicit string ignored)
func TestConfigPrecedenceIntegration(t *testing.T) {
	const (
		sourceEnv      = "env"
		sourceYAML     = "yaml"
		sourceExplicit = "explicit"
	)
	base := t.TempDir()
	logger := logging.NewLogger(logging.LevelError)

	// Prepare initial YAML file (absolute path required for loader) with reports.output_dir populated.
	yamlReportsDir := filepath.Join(base, "yaml-reports")
	yamlPath := filepath.Join(base, "config.yaml")
	content := "reports:\n  output_dir: " + yamlReportsDir + "\ncore:\n  log_level: warn\n"
	if err := os.WriteFile(yamlPath, []byte(content), 0o600); err != nil {
		t.Fatalf("write yaml: %v", err)
	}
	t.Setenv("COSTSCOPE_CONFIG", yamlPath)
	// ENV override (wins only when YAML missing for this field)
	envReports := filepath.Join(base, "env-reports")
	t.Setenv("COSTSCOPE_REPORTS_DIR", envReports)

	// Case 1: No explicit, YAML wins.
	res1 := ResolveStringField(logger, "reports.output_dir", nil, func(c *ConsolidatedConfig) *string { return &c.Reports.OutputDir }, "COSTSCOPE_REPORTS_DIR", "fallback-dir")
	if res1.Value != yamlReportsDir || res1.Source != sourceYAML {
		t.Fatalf("case1 expected yaml=%s got value=%s source=%s", yamlReportsDir, res1.Value, res1.Source)
	}

	// Rewrite YAML with empty output_dir to force env precedence (simulating YAML absence for this field)
	if err := os.WriteFile(yamlPath, []byte("reports:\n  output_dir: ''\n"), 0o600); err != nil {
		t.Fatalf("rewrite yaml: %v", err)
	}
	res2 := ResolveStringField(logger, "reports.output_dir", nil, func(c *ConsolidatedConfig) *string { return &c.Reports.OutputDir }, "COSTSCOPE_REPORTS_DIR", "fallback-dir")
	if res2.Source != sourceEnv || res2.Value != envReports {
		t.Fatalf("case2 expected env=%s got value=%s source=%s", envReports, res2.Value, res2.Source)
	}

	// Case 3: Explicit beats all.
	explicitDir := filepath.Join(base, "explicit-reports")
	res3 := ResolveStringField(logger, "reports.output_dir", &explicitDir, func(c *ConsolidatedConfig) *string { return &c.Reports.OutputDir }, "COSTSCOPE_REPORTS_DIR", "fallback-dir")
	if res3.Source != sourceExplicit || res3.Value != explicitDir {
		t.Fatalf("case3 expected explicit=%s got value=%s source=%s", explicitDir, res3.Value, res3.Source)
	}

	// Case 4: Explicit empty pointer prevents YAML load (implementation detail) -> env wins (documented behavior).
	empty := ""
	res4 := ResolveStringField(logger, "reports.output_dir", &empty, func(c *ConsolidatedConfig) *string { return &c.Reports.OutputDir }, "COSTSCOPE_REPORTS_DIR", "fallback-dir")
	if res4.Source != sourceEnv || res4.Value != envReports {
		t.Fatalf("case4 expected env due to empty explicit blocking YAML load; got value=%s source=%s", res4.Value, res4.Source)
	}

	// Finalize a config instance and ensure validation + directory preparation still succeed.
	cfg := minimalValidConfig(base)
	cfg.Reports.OutputDir = explicitDir // use resolved explicit
	if err := ValidateAllConfig(cfg); err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}
	cfg.Core.DataDirectory = filepath.Join(base, "coredata")
	cfg.Core.TempDirectory = filepath.Join(base, "tmp")
	if err := EnsureConfigDirectories(cfg); err != nil {
		t.Fatalf("ensure directories: %v", err)
	}
}
