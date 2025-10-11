package config

import (
	"os"
	"path/filepath"
	"testing"

	"local/costscope/internal/core/logging"
)

func TestResolveBoolField_Precedence_Log(t *testing.T) {
	logger := logging.NewLogger(logging.LevelInfo)

	// Prepare a YAML file with focus.use_unified_mapper_default: true
	dir := t.TempDir()
	yamlPath := filepath.Join(dir, "config.yaml")
	// #nosec G306 temporary test file
	if err := os.WriteFile(yamlPath, []byte("focus:\n  use_unified_mapper_default: true\n"), 0o600); err != nil {
		t.Fatalf("write yaml: %v", err)
	}
	// Point resolver to YAML
	t.Setenv("COSTSCOPE_CONFIG", yamlPath)
	t.Setenv("COSTSCOPE_USE_UNIFIED_MAPPER", "") // ensure env not used

	// No explicit -> YAML should win
	res := ResolveBoolField(logger, "focus.use_unified_mapper", nil, func(c *ConsolidatedConfig) *bool { return &c.Focus.UseUnifiedMapperDefault }, "COSTSCOPE_USE_UNIFIED_MAPPER", false)
	if res.Value != true || string(res.Source) != "yaml" {
		t.Fatalf("expected yaml true, got %+v", res)
	}

	// Explicit should override YAML
	explicit := false
	res = ResolveBoolField(logger, "focus.use_unified_mapper", &explicit, func(c *ConsolidatedConfig) *bool { return &c.Focus.UseUnifiedMapperDefault }, "COSTSCOPE_USE_UNIFIED_MAPPER", true)
	if res.Value != false || string(res.Source) != "explicit" {
		t.Fatalf("expected explicit false, got %+v", res)
	}

	// ENV should be used when no yaml + no explicit
	if err := os.Unsetenv("COSTSCOPE_CONFIG"); err != nil {
		t.Fatalf("unset COSTSCOPE_CONFIG: %v", err)
	}
	t.Setenv("COSTSCOPE_USE_UNIFIED_MAPPER", "yes")
	res = ResolveBoolField(logger, "focus.use_unified_mapper", nil, nil, "COSTSCOPE_USE_UNIFIED_MAPPER", false)
	if res.Value != true || string(res.Source) != "env" {
		t.Fatalf("expected env true, got %+v", res)
	}
}
