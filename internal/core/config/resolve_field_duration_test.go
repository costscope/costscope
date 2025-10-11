package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"local/costscope/internal/core/config/precedence"
	"local/costscope/internal/core/logging"
)

// TestResolveDurationField_PrecedenceMatrix validates explicit > YAML > ENV > fallback ordering (zero meaningful).
func TestResolveDurationField_PrecedenceMatrix(t *testing.T) {
	logger := logging.NewLogger(logging.LevelError)
	tmp := t.TempDir()
	yamlPath := filepath.Join(tmp, "config.yaml")
	if err := os.WriteFile(yamlPath, []byte("core:\n  timeout: 5s\n"), 0o600); err != nil {
		t.Fatalf("write yaml: %v", err)
	}

	dPtr := func(d time.Duration) *time.Duration { return &d }

	cases := []struct {
		name      string
		explicit  *time.Duration
		setYAML   bool
		envVal    string
		expectVal time.Duration
		expectSrc precedence.Source
	}{
		{name: "explicit wins over yaml/env", explicit: dPtr(1500 * time.Millisecond), setYAML: true, envVal: "10s", expectVal: 1500 * time.Millisecond, expectSrc: precedence.SourceExplicit},
		{name: "yaml over env", explicit: nil, setYAML: true, envVal: "10s", expectVal: 5 * time.Second, expectSrc: precedence.SourceYAML},
		{name: "env when no explicit/yaml", explicit: nil, setYAML: false, envVal: "10s", expectVal: 10 * time.Second, expectSrc: precedence.SourceEnv},
		{name: "fallback when nothing", explicit: nil, setYAML: false, envVal: "", expectVal: 2 * time.Second, expectSrc: precedence.SourceDefault},
		{name: "explicit zero meaningful", explicit: dPtr(0), setYAML: true, envVal: "10s", expectVal: 0, expectSrc: precedence.SourceExplicit},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.envVal != "" {
				t.Setenv("TEST_DURATION_FIELD_MATRIX", tc.envVal)
			} else {
				_ = os.Unsetenv("TEST_DURATION_FIELD_MATRIX")
			}
			if tc.setYAML && tc.explicit == nil {
				t.Setenv("COSTSCOPE_CONFIG", yamlPath)
			} else {
				_ = os.Unsetenv("COSTSCOPE_CONFIG")
			}
			res := ResolveDurationField(logger, "core.timeout", tc.explicit, func(c *ConsolidatedConfig) *time.Duration { return &c.Core.Timeout }, "TEST_DURATION_FIELD_MATRIX", 2*time.Second)
			if res.Value != tc.expectVal || res.Source != tc.expectSrc {
				t.Fatalf("want (%v,%s) got (%v,%s)", tc.expectVal, tc.expectSrc, res.Value, res.Source)
			}
		})
	}
}
