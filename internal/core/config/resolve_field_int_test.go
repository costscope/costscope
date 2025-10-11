package config

import (
	"os"
	"path/filepath"
	"testing"

	"local/costscope/internal/core/config/precedence"
	"local/costscope/internal/core/logging"
)

// TestResolveIntField_PrecedenceMatrix validates explicit > YAML > ENV > fallback across combinations.
func TestResolveIntField_PrecedenceMatrix(t *testing.T) {
	logger := logging.NewLogger(logging.LevelError)
	tmp := t.TempDir()
	yamlPath := filepath.Join(tmp, "config.yaml")
	if err := os.WriteFile(yamlPath, []byte("streaming:\n  max_concurrent_jobs: 7\n"), 0o600); err != nil {
		t.Fatalf("write yaml: %v", err)
	}

	cases := []struct {
		name      string
		explicit  *int
		setYAML   bool
		envVal    string
		expectVal int
		expectSrc precedence.Source
	}{
		{name: "explicit wins over yaml/env", explicit: intPtr(3), setYAML: true, envVal: "11", expectVal: 3, expectSrc: precedence.SourceExplicit},
		{name: "yaml over env", explicit: nil, setYAML: true, envVal: "11", expectVal: 7, expectSrc: precedence.SourceYAML},
		{name: "env when no explicit/yaml", explicit: nil, setYAML: false, envVal: "11", expectVal: 11, expectSrc: precedence.SourceEnv},
		{name: "fallback when nothing", explicit: nil, setYAML: false, envVal: "", expectVal: 5, expectSrc: precedence.SourceDefault},
		{name: "explicit zero meaningful", explicit: intPtr(0), setYAML: true, envVal: "11", expectVal: 0, expectSrc: precedence.SourceExplicit},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// env setup
			if tc.envVal != "" {
				t.Setenv("TEST_INT_FIELD_MATRIX", tc.envVal)
			} else {
				_ = os.Unsetenv("TEST_INT_FIELD_MATRIX")
			}
			if tc.setYAML && tc.explicit == nil { // only expose YAML when it can be considered
				t.Setenv("COSTSCOPE_CONFIG", yamlPath)
			} else {
				_ = os.Unsetenv("COSTSCOPE_CONFIG")
			}
			res := ResolveIntField(logger, "streaming.max_concurrent_jobs", tc.explicit, func(c *ConsolidatedConfig) *int { return &c.Streaming.MaxConcurrentJobs }, "TEST_INT_FIELD_MATRIX", 5)
			if res.Value != tc.expectVal || res.Source != tc.expectSrc {
				t.Fatalf("want (%d,%s) got (%d,%s)", tc.expectVal, tc.expectSrc, res.Value, res.Source)
			}
		})
	}
}

func intPtr(v int) *int { return &v }
