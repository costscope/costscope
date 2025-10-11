package config

import (
	"os"
	"path/filepath"
	"testing"

	"local/costscope/internal/core/config/precedence"
	"local/costscope/internal/core/logging"
)

// TestResolveFloatField_PrecedenceMatrix validates explicit > YAML > ENV > fallback ordering for floats (zero meaningful).
func TestResolveFloatField_PrecedenceMatrix(t *testing.T) {
	logger := logging.NewLogger(logging.LevelError)
	tmp := t.TempDir()
	yamlPath := filepath.Join(tmp, "config.yaml")
	if err := os.WriteFile(yamlPath, []byte("focus:\n  invariants_tolerance_default: 0.02\n"), 0o600); err != nil {
		t.Fatalf("write yaml: %v", err)
	}

	fPtr := func(f float64) *float64 { return &f }

	cases := []struct {
		name      string
		explicit  *float64
		setYAML   bool
		envVal    string
		expectVal float64
		expectSrc precedence.Source
	}{
		{name: "explicit wins over yaml/env", explicit: fPtr(0.15), setYAML: true, envVal: "0.33", expectVal: 0.15, expectSrc: precedence.SourceExplicit},
		{name: "yaml over env", explicit: nil, setYAML: true, envVal: "0.33", expectVal: 0.02, expectSrc: precedence.SourceYAML},
		{name: "env when no explicit/yaml", explicit: nil, setYAML: false, envVal: "0.33", expectVal: 0.33, expectSrc: precedence.SourceEnv},
		{name: "fallback when nothing", explicit: nil, setYAML: false, envVal: "", expectVal: 0.01, expectSrc: precedence.SourceDefault},
		{name: "explicit zero meaningful", explicit: fPtr(0.0), setYAML: true, envVal: "0.33", expectVal: 0.0, expectSrc: precedence.SourceExplicit},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.envVal != "" {
				t.Setenv("TEST_FLOAT_FIELD_MATRIX", tc.envVal)
			} else {
				_ = os.Unsetenv("TEST_FLOAT_FIELD_MATRIX")
			}
			if tc.setYAML && tc.explicit == nil {
				t.Setenv("COSTSCOPE_CONFIG", yamlPath)
			} else {
				_ = os.Unsetenv("COSTSCOPE_CONFIG")
			}
			res := ResolveFloatField(logger, "focus.invariants_tolerance_default", tc.explicit, func(c *ConsolidatedConfig) *float64 { return &c.Focus.InvariantsToleranceDefault }, "TEST_FLOAT_FIELD_MATRIX", 0.01)
			if res.Value != tc.expectVal || res.Source != tc.expectSrc {
				t.Fatalf("want (%v,%s) got (%v,%s)", tc.expectVal, tc.expectSrc, res.Value, res.Source)
			}
		})
	}
}
