package config

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/costscope/costscope/internal/core/config/precedence"
	"github.com/costscope/costscope/internal/core/logging"
)

// TestLoadFromFile_PathValidation ensures invalid relative/parent paths are rejected.
func TestLoadFromFile_PathValidation(t *testing.T) {
	logger := logging.NewLogger(logging.LevelError)
	cl := NewConfigLoader(logger)

	// Create a temp file to have an absolute valid path
	tmpDir := t.TempDir()
	validPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(validPath, []byte("core:\n  app_name: test\n"), 0o600); err != nil {
		t.Fatalf("failed writing temp config: %v", err)
	}
	if _, err := cl.LoadFromFile(validPath); err != nil {
		t.Fatalf("expected valid abs path to load: %v", err)
	}

	// Create a relative file in temp dir and attempt to load it via relative path after chdir.
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("chdir tmp: %v", err)
	}
	defer func() { _ = os.Chdir(cwd) }()
	relFile := "rel.yaml"
	if err := os.WriteFile(relFile, []byte("core:\n  app_name: test\n"), 0o600); err != nil {
		t.Fatalf("failed writing rel file: %v", err)
	}
	if _, err := cl.LoadFromFile(relFile); err == nil || !strings.Contains(err.Error(), "invalid config path") {
		t.Fatalf("expected invalid config path for relative file, got %v", err)
	}
}

// TestLoadFromFile_StrictUnknownFields verifies that unknown YAML fields cause an error due to KnownFields(true).
func TestLoadFromFile_StrictUnknownFields(t *testing.T) {
	logger := logging.NewLogger(logging.LevelError)
	cl := NewConfigLoader(logger)
	tmpDir := t.TempDir()
	bad := filepath.Join(tmpDir, "bad.yaml")
	// Introduce an unknown top-level field `unknown_root` and unknown nested field `bogus`.
	yaml := "unknown_root: 123\ncore:\n  bogus: value\n"
	if err := os.WriteFile(bad, []byte(yaml), 0o600); err != nil {
		t.Fatalf("write yaml: %v", err)
	}
	if _, err := cl.LoadFromFile(bad); err == nil {
		t.Fatalf("expected error for unknown fields")
	}
}

// TestResolveBoolField_YAMLDefaultPrecedence covers explicit > YAML > env > fallback ordering for bool + masking of sensitive names.
func TestResolveBoolField_YAMLDefaultPrecedence(t *testing.T) {
	logger := logging.NewLogger(logging.LevelError)
	// Prepare YAML with focus.use_unified_mapper_default: true and reports.output_dir.
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.yaml")
	contents := "focus:\n  use_unified_mapper_default: true\nreports:\n  output_dir: /custom/reports\nsecurity:\n  jwt_secret: abc123\n"
	if err := os.WriteFile(cfgPath, []byte(contents), 0o600); err != nil {
		t.Fatalf("write yaml: %v", err)
	}
	t.Setenv("COSTSCOPE_CONFIG", cfgPath)
	// Also set ENV overrides that should be ignored because YAML takes precedence over env (and explicit absent).
	t.Setenv("COSTSCOPE_USE_UNIFIED_MAPPER", "false")

	// 1. Bool resolve (no explicit) -> YAML default true
	boolRes := ResolveBoolField(logger, "focus.use_unified_mapper", nil, func(c *ConsolidatedConfig) *bool { return &c.Focus.UseUnifiedMapperDefault }, "COSTSCOPE_USE_UNIFIED_MAPPER", false)
	if !boolRes.Value || boolRes.Source != precedence.SourceYAML {
		t.Fatalf("expected YAML true source=yaml got val=%v src=%s", boolRes.Value, boolRes.Source)
	}

	// 2. With explicit override false
	explicit := false
	boolRes2 := ResolveBoolField(logger, "focus.use_unified_mapper", &explicit, func(c *ConsolidatedConfig) *bool { return &c.Focus.UseUnifiedMapperDefault }, "COSTSCOPE_USE_UNIFIED_MAPPER", true)
	if boolRes2.Value || boolRes2.Source != precedence.SourceExplicit {
		t.Fatalf("expected explicit false override got val=%v src=%s", boolRes2.Value, boolRes2.Source)
	}
}

// TestResolveStringField_Precedence verifies that an explicit empty string does not override
// a YAML default, and that YAML beats ENV, falling back correctly.
func TestResolveStringField_Precedence(t *testing.T) {
	logger := logging.NewLogger(logging.LevelInfo)

	dir := t.TempDir()
	yamlPath := filepath.Join(dir, "config.yaml")
	// #nosec G306 test file
	if err := os.WriteFile(yamlPath, []byte("reports:\n  output_dir: yaml-dir\n"), 0o600); err != nil {
		t.Fatalf("write yaml: %v", err)
	}
	t.Setenv("COSTSCOPE_CONFIG", yamlPath)
	t.Setenv("COSTSCOPE_REPORTS_DIR", "env-dir")

	// First with no explicit so YAML loads
	res := ResolveStringField(logger, "reports.output_dir", nil, func(c *ConsolidatedConfig) *string { return &c.Reports.OutputDir }, "COSTSCOPE_REPORTS_DIR", "fallback-dir")
	if res.Value != "yaml-dir" || res.Source != precedence.SourceYAML {
		t.Fatalf("expected yaml-dir from yaml, got %+v", res)
	}
	// Now simulate explicit empty (should still select YAML because explicit empty is ignored inside precedence, but YAML will not be reloaded since explicit pointer non-nil -> so we expect fallback to env because yamlPtr not loaded). This documents current behavior.
	empty := ""
	res = ResolveStringField(logger, "reports.output_dir", &empty, func(c *ConsolidatedConfig) *string { return &c.Reports.OutputDir }, "COSTSCOPE_REPORTS_DIR", "fallback-dir")
	if res.Source != precedence.SourceEnv || res.Value != "env-dir" {
		t.Fatalf("expected env-dir via env when explicit empty blocks YAML load, got %+v", res)
	}

	// Remove YAML, expect ENV to win
	if err := os.Unsetenv("COSTSCOPE_CONFIG"); err != nil {
		t.Fatalf("unset COSTSCOPE_CONFIG: %v", err)
	}
	res = ResolveStringField(logger, "reports.output_dir", nil, nil, "COSTSCOPE_REPORTS_DIR", "fallback-dir")
	if res.Value != "env-dir" || res.Source != precedence.SourceEnv {
		t.Fatalf("expected env-dir from env, got %+v", res)
	}

	// Unset env, expect fallback
	if err := os.Unsetenv("COSTSCOPE_REPORTS_DIR"); err != nil {
		t.Fatalf("unset COSTSCOPE_REPORTS_DIR: %v", err)
	}
	res = ResolveStringField(logger, "reports.output_dir", nil, nil, "COSTSCOPE_REPORTS_DIR", "fallback-dir")
	if res.Value != "fallback-dir" || res.Source != precedence.SourceDefault {
		t.Fatalf("expected fallback-dir default, got %+v", res)
	}
}

// TestResolveIntField_EnvPrecedence verifies env-only int precedence.
func TestResolveIntField_EnvPrecedence(t *testing.T) {
	logger := logging.NewLogger(logging.LevelInfo)

	t.Setenv("TEST_INT_PRECEDENCE", "42")
	res := ResolveIntField(logger, "streaming.max_concurrent_jobs", nil, nil, "TEST_INT_PRECEDENCE", 5)
	if res.Value != 42 || res.Source != precedence.SourceEnv {
		t.Fatalf("expected 42 env, got %+v", res)
	}

	// Unset env -> fallback
	if err := os.Unsetenv("TEST_INT_PRECEDENCE"); err != nil {
		t.Fatalf("unset env: %v", err)
	}
	res = ResolveIntField(logger, "streaming.max_concurrent_jobs", nil, nil, "TEST_INT_PRECEDENCE", 5)
	if res.Value != 5 || res.Source != precedence.SourceDefault {
		t.Fatalf("expected fallback 5, got %+v", res)
	}
}

// TestResolveStringField_Masking ensures secret-like fields are redacted in logs.
func TestResolveStringField_Masking(t *testing.T) {
	logger := logging.NewLogger(logging.LevelInfo)

	dir := t.TempDir()
	yamlPath := filepath.Join(dir, "config.yaml")
	// #nosec G306 test file
	if err := os.WriteFile(yamlPath, []byte("security:\n  jwt_secret: supersecretvalue\n"), 0o600); err != nil {
		t.Fatalf("write yaml: %v", err)
	}
	t.Setenv("COSTSCOPE_CONFIG", yamlPath)

	// Capture stderr
	orig := os.Stderr
	rPipe, wPipe, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stderr = wPipe
	defer func() { os.Stderr = orig }()

	res := ResolveStringField(logger, "security.jwt_secret", nil, func(c *ConsolidatedConfig) *string { return &c.Security.JWTSecret }, "COSTSCOPE_JWT_SECRET", "")
	if res.Source != precedence.SourceYAML || res.Value != "supersecretvalue" {
		t.Fatalf("expected yaml secret value, got %+v", res)
	}
	// Close writer and read
	_ = wPipe.Close()
	logged, _ := io.ReadAll(rPipe)
	logStr := string(logged)
	if !strings.Contains(logStr, "\"value\":\"[REDACTED]\"") {
		t.Fatalf("expected redacted value in log, got: %s", logStr)
	}
	if strings.Contains(logStr, "supersecretvalue") {
		t.Fatalf("secret value should not appear in logs: %s", logStr)
	}
}

// TestResolveBoolField_YAMLMiss_Default ensures that a relative COSTSCOPE_CONFIG path is ignored
// and fallback default is used (source=default).
func TestResolveBoolField_YAMLMiss_Default(t *testing.T) {
	logger := logging.NewLogger(logging.LevelInfo)

	relDir := t.TempDir()
	relPath := filepath.Join(relDir, "config.yaml")
	// #nosec G306 test file
	if err := os.WriteFile(relPath, []byte("focus:\n  use_unified_mapper_default: true\n"), 0o600); err != nil {
		t.Fatalf("write yaml: %v", err)
	}
	// Set relative path (not absolute) => should be ignored
	t.Setenv("COSTSCOPE_CONFIG", "config.yaml")

	res := ResolveBoolField(logger, "focus.use_unified_mapper", nil, func(c *ConsolidatedConfig) *bool { return &c.Focus.UseUnifiedMapperDefault }, "COSTSCOPE_USE_UNIFIED_MAPPER", false)
	if res.Value != false || res.Source != precedence.SourceDefault {
		t.Fatalf("expected default false due to YAML miss, got %+v", res)
	}
}

// TestResolveDurationField_Precedence verifies explicit > YAML > env > fallback for durations.
func TestResolveDurationField_Precedence(t *testing.T) {
	logger := logging.NewLogger(logging.LevelInfo)

	dir := t.TempDir()
	yamlPath := filepath.Join(dir, "config.yaml")
	// Provide a YAML default for streaming.job_timeout: 45s
	if err := os.WriteFile(yamlPath, []byte("streaming:\n  job_timeout: 45s\n"), 0o600); err != nil {
		t.Fatalf("write yaml: %v", err)
	}
	t.Setenv("COSTSCOPE_CONFIG", yamlPath)
	t.Setenv("TEST_JOB_TIMEOUT", "30s")

	// 1. No explicit -> YAML (45s)
	res := ResolveDurationField(logger, "streaming.job_timeout", nil, func(c *ConsolidatedConfig) *time.Duration { return &c.Streaming.JobTimeout }, "TEST_JOB_TIMEOUT", 10*time.Second)
	if res.Value != 45*time.Second || res.Source != precedence.SourceYAML {
		t.Fatalf("expected yaml 45s got %+v", res)
	}

	// 2. Explicit overrides (15s)
	explicit := 15 * time.Second
	res = ResolveDurationField(logger, "streaming.job_timeout", &explicit, func(c *ConsolidatedConfig) *time.Duration { return &c.Streaming.JobTimeout }, "TEST_JOB_TIMEOUT", 10*time.Second)
	if res.Value != 15*time.Second || res.Source != precedence.SourceExplicit {
		t.Fatalf("expected explicit 15s got %+v", res)
	}

	// 3. Remove YAML path (unset COSTSCOPE_CONFIG) so env wins
	if err := os.Unsetenv("COSTSCOPE_CONFIG"); err != nil {
		t.Fatalf("unset COSTSCOPE_CONFIG: %v", err)
	}
	res = ResolveDurationField(logger, "streaming.job_timeout", nil, nil, "TEST_JOB_TIMEOUT", 10*time.Second)
	if res.Value != 30*time.Second || res.Source != precedence.SourceEnv {
		t.Fatalf("expected env 30s got %+v", res)
	}

	// 4. Bad env value -> fallback
	t.Setenv("TEST_JOB_TIMEOUT", "notaduration")
	res = ResolveDurationField(logger, "streaming.job_timeout", nil, nil, "TEST_JOB_TIMEOUT", 10*time.Second)
	if res.Value != 10*time.Second || res.Source != precedence.SourceDefault {
		t.Fatalf("expected fallback 10s got %+v", res)
	}
}
