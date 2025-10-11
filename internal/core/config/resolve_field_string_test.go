package config

import (
	"os"
	"path/filepath"
	"testing"

	"local/costscope/internal/core/config/precedence"
	"local/costscope/internal/core/logging"
)

// TestResolveStringField_PrecedenceMatrix validates explicit > YAML > ENV > fallback ordering with empty-string semantics.
func TestResolveStringField_PrecedenceMatrix(t *testing.T) {
	logger := logging.NewLogger(logging.LevelError)
	tmp := t.TempDir()
	yamlPath := filepath.Join(tmp, "config.yaml")
	if err := os.WriteFile(yamlPath, []byte("core:\n  log_level: info\n"), 0o600); err != nil {
		t.Fatalf("write yaml: %v", err)
	}

	strPtr := func(s string) *string { return &s }

	cases := []struct {
		name      string
		explicit  *string
		setYAML   bool
		envVal    string
		expectVal string
		expectSrc precedence.Source
	}{
		{name: "explicit wins over yaml/env", explicit: strPtr("debug"), setYAML: true, envVal: "warn", expectVal: "debug", expectSrc: precedence.SourceExplicit},
		{name: "yaml over env", explicit: nil, setYAML: true, envVal: "warn", expectVal: "info", expectSrc: precedence.SourceYAML},
		{name: "env when no explicit/yaml", explicit: nil, setYAML: false, envVal: "warn", expectVal: "warn", expectSrc: precedence.SourceEnv},
		{name: "fallback when nothing", explicit: nil, setYAML: false, envVal: "", expectVal: "error", expectSrc: precedence.SourceDefault},
		{name: "explicit empty ignored -> env (yaml not loaded when explicit ptr present)", explicit: strPtr(""), setYAML: true, envVal: "warn", expectVal: "warn", expectSrc: precedence.SourceEnv},
		{name: "explicit empty no yaml -> env", explicit: strPtr(""), setYAML: false, envVal: "warn", expectVal: "warn", expectSrc: precedence.SourceEnv},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.envVal != "" {
				t.Setenv("TEST_STRING_FIELD_MATRIX", tc.envVal)
			} else {
				_ = os.Unsetenv("TEST_STRING_FIELD_MATRIX")
			}
			if tc.setYAML && (tc.explicit == nil || *tc.explicit == "") {
				t.Setenv("COSTSCOPE_CONFIG", yamlPath)
			} else {
				_ = os.Unsetenv("COSTSCOPE_CONFIG")
			}
			res := ResolveStringField(logger, "core.log_level", tc.explicit, func(c *ConsolidatedConfig) *string { return &c.Core.LogLevel }, "TEST_STRING_FIELD_MATRIX", "error")
			if res.Value != tc.expectVal || res.Source != tc.expectSrc {
				t.Fatalf("want (%s,%s) got (%s,%s)", tc.expectVal, tc.expectSrc, res.Value, res.Source)
			}
		})
	}
}
