package config

import (
	"os"
	"testing"

	"local/costscope/internal/core/config/precedence"
	"local/costscope/internal/core/logging"
)

// TestResolveFloatField_Precedence verifies explicit > yaml > env > default ordering and source tagging.
func TestResolveFloatField_Precedence(t *testing.T) {
	logger := logging.NewLogger(logging.LevelError) // quiet

	// prepare a temporary YAML config file with focus.invariants_tolerance_default
	yamlContent := []byte("focus:\n  invariants_tolerance_default: 2.25\n")
	tmp, err := os.CreateTemp(t.TempDir(), "cfg-*.yaml")
	if err != nil {
		t.Fatalf("temp file: %v", err)
	}
	if _, err := tmp.Write(yamlContent); err != nil {
		t.Fatalf("write yaml: %v", err)
	}
	_ = tmp.Close()
	if err := os.Setenv("COSTSCOPE_CONFIG", tmp.Name()); err != nil {
		t.Fatalf("set env: %v", err)
	}
	defer func() {
		if err := os.Unsetenv("COSTSCOPE_CONFIG"); err != nil {
			t.Logf("unset COSTSCOPE_CONFIG: %v", err)
		}
	}()

	// env value (should lose to YAML when YAML present and explicit nil)
	if err := os.Setenv("TEST_TOL_ENV", "3.5"); err != nil {
		t.Fatalf("set env: %v", err)
	}
	defer func() {
		if err := os.Unsetenv("TEST_TOL_ENV"); err != nil {
			t.Logf("unset TEST_TOL_ENV: %v", err)
		}
	}()

	// case 1: explicit wins
	exp := 1.5
	res := ResolveFloatField(logger, "focus.invariants_tolerance", &exp, func(c *ConsolidatedConfig) *float64 {
		return &c.Focus.InvariantsToleranceDefault
	}, "TEST_TOL_ENV", 0.75)
	if res.Value != 1.5 || res.Source != precedence.SourceExplicit {
		t.Fatalf("expected explicit 1.5 got %+v", res)
	}

	// case 2: yaml when explicit nil
	res = ResolveFloatField(logger, "focus.invariants_tolerance", nil, func(c *ConsolidatedConfig) *float64 {
		return &c.Focus.InvariantsToleranceDefault
	}, "TEST_TOL_ENV", 0.75)
	if res.Value != 2.25 || res.Source != precedence.SourceYAML {
		t.Fatalf("expected yaml 2.25 got %+v", res)
	}

	// case 3: env when yaml absent (supply selector returning nil)
	res = ResolveFloatField(logger, "focus.invariants_tolerance", nil, func(c *ConsolidatedConfig) *float64 { return nil }, "TEST_TOL_ENV", 0.75)
	if res.Value != 3.5 || res.Source != precedence.SourceEnv {
		t.Fatalf("expected env 3.5 got %+v", res)
	}

	// case 4: default when nothing provided (unset env, nil selector)
	if err := os.Unsetenv("TEST_TOL_ENV"); err != nil {
		t.Fatalf("unset env: %v", err)
	}
	res = ResolveFloatField(logger, "focus.invariants_tolerance", nil, func(c *ConsolidatedConfig) *float64 { return nil }, "TEST_TOL_ENV", 0.75)
	if res.Value != 0.75 || res.Source != precedence.SourceDefault {
		t.Fatalf("expected default 0.75 got %+v", res)
	}
}
