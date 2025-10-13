package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/costscope/costscope/internal/core/config/precedence"
	"github.com/costscope/costscope/internal/core/logging"
)

// TestResolveBoolField_PrecedenceMatrix validates explicit > YAML > ENV > fallback ordering.
func TestResolveBoolField_PrecedenceMatrix(t *testing.T) {
	logger := logging.NewLogger(logging.LevelError)
	tmp := t.TempDir()
	yamlPath := filepath.Join(tmp, "config.yaml")
	if err := os.WriteFile(yamlPath, []byte("core:\n  metrics_enabled: true\n"), 0o600); err != nil {
		t.Fatalf("write yaml: %v", err)
	}
	trueVal := func(b bool) *bool { return &b }

	cases := []struct {
		name      string
		explicit  *bool
		setYAML   bool
		envVal    string
		expectVal bool
		expectSrc precedence.Source
	}{
		{name: "explicit wins over yaml/env", explicit: trueVal(false), setYAML: true, envVal: "1", expectVal: false, expectSrc: precedence.SourceExplicit},
		{name: "yaml over env", explicit: nil, setYAML: true, envVal: "1", expectVal: true, expectSrc: precedence.SourceYAML},
		{name: "env when no explicit/yaml", explicit: nil, setYAML: false, envVal: "true", expectVal: true, expectSrc: precedence.SourceEnv},
		{name: "fallback when nothing", explicit: nil, setYAML: false, envVal: "", expectVal: false, expectSrc: precedence.SourceDefault},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.envVal != "" {
				t.Setenv("TEST_BOOL_FIELD_MATRIX", tc.envVal)
			} else {
				_ = os.Unsetenv("TEST_BOOL_FIELD_MATRIX")
			}
			if tc.setYAML && tc.explicit == nil {
				t.Setenv("COSTSCOPE_CONFIG", yamlPath)
			} else {
				_ = os.Unsetenv("COSTSCOPE_CONFIG")
			}
			res := ResolveBoolField(logger, "core.metrics_enabled", tc.explicit, func(c *ConsolidatedConfig) *bool { return &c.Core.MetricsEnabled }, "TEST_BOOL_FIELD_MATRIX", false)
			if res.Value != tc.expectVal || res.Source != tc.expectSrc {
				t.Fatalf("want (%v,%s) got (%v,%s)", tc.expectVal, tc.expectSrc, res.Value, res.Source)
			}
		})
	}
}
