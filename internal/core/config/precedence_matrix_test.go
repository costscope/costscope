package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"local/costscope/internal/core/config/precedence"
	"local/costscope/internal/core/logging"
)

// TestConfigPrecedence_Matrix covers combined precedence & logging for the GA blocker keys:
//   - jwt_secret_present (derived boolean; value=true when security.jwt_secret resolved non-empty)
//   - multi_tenant.enabled
//   - focus.use_unified_mapper
//   - focus.invariants_enabled_default
//
// Scenarios: only default, env only, yaml only, explicit flag, invalid env (ignored ⇒ fallback), yaml + explicit (explicit wins).
// Assertions: value + source per field, a single config_precedence_resolved log line for each field, and no secret value leakage.
func TestConfigPrecedence_Matrix(t *testing.T) { // refactored to keep cyclomatic complexity low
	type scenario struct {
		name string
		// inputs
		setEnv             bool
		invalidEnv         bool
		writeYAML          bool
		setExplicit        bool
		setYAMLAndExplicit bool // explicit overriding yaml
	}

	scenarios := []scenario{
		{name: "default"},
		{name: "env", setEnv: true},
		{name: "yaml", writeYAML: true},
		{name: "explicit", setExplicit: true},
		{name: "invalid_env", invalidEnv: true},
		{name: "yaml_plus_explicit", writeYAML: true, setExplicit: true, setYAMLAndExplicit: true},
	}

	secretExplicit := "EXPLICIT_SECRET_TEST"
	secretEnv := "ENV_SECRET_TEST"

	for _, sc := range scenarios {
		t.Run(sc.name, func(t *testing.T) {
			// reset env keys each run
			envKeys := []string{"COSTSCOPE_JWT_SECRET", "COSTSCOPE_MULTI_TENANT_ENABLED", "COSTSCOPE_USE_UNIFIED_MAPPER", "COSTSCOPE_INVARIANTS_ENABLED", "COSTSCOPE_CONFIG"}
			for _, k := range envKeys {
				_ = os.Unsetenv(k)
			}

			// optional YAML
			var yamlPath string
			if sc.writeYAML {
				dir := t.TempDir()
				yamlPath = filepath.Join(dir, "config.yaml")
				// YAML defaults set to true / secret value
				contents := "security:\n  jwt_secret: SHARED_YAML_SECRET\nmulti_tenant:\n  enabled: true\nfocus:\n  use_unified_mapper_default: true\n  invariants_enabled_default: true\n"
				if err := os.WriteFile(yamlPath, []byte(contents), 0o600); err != nil {
					t.Fatalf("write yaml: %v", err)
				}
				if err := os.Setenv("COSTSCOPE_CONFIG", yamlPath); err != nil {
					t.Fatalf("set COSTSCOPE_CONFIG: %v", err)
				}
			}

			// env values
			if sc.setEnv {
				_ = os.Setenv("COSTSCOPE_JWT_SECRET", secretEnv)
				_ = os.Setenv("COSTSCOPE_MULTI_TENANT_ENABLED", "true")
				_ = os.Setenv("COSTSCOPE_USE_UNIFIED_MAPPER", "true")
				_ = os.Setenv("COSTSCOPE_INVARIANTS_ENABLED", "true")
			}
			if sc.invalidEnv { // invalid boolean tokens ignored -> fallback
				_ = os.Setenv("COSTSCOPE_MULTI_TENANT_ENABLED", "maybe")
				_ = os.Setenv("COSTSCOPE_USE_UNIFIED_MAPPER", "maybe")
				_ = os.Setenv("COSTSCOPE_INVARIANTS_ENABLED", "maybe")
				// empty secret env yields absence
				_ = os.Setenv("COSTSCOPE_JWT_SECRET", "")
			}

			// capture logs (logger writes to stderr)
			oldStderr := os.Stderr
			r, w, err := os.Pipe()
			if err != nil {
				t.Fatalf("pipe: %v", err)
			}
			os.Stderr = w
			logger := logging.NewLogger(logging.LevelInfo)

			// explicit pointers
			var explicitSecret *string
			var explicitMT, explicitUnified, explicitInv *bool
			if sc.setExplicit {
				explicitSecret = &secretExplicit
				tru := true
				explicitMT, explicitUnified, explicitInv = &tru, &tru, &tru
			}

			// Resolve jwt secret (string) then log presence as a boolean field jwt_secret_present.
			secretRes := ResolveStringField(logger, "security.jwt_secret", explicitSecret, func(cc *ConsolidatedConfig) *string {
				if cc == nil || cc.Security.JWTSecret == "" {
					return nil
				}
				return &cc.Security.JWTSecret
			}, "COSTSCOPE_JWT_SECRET", "")
			// Derive presence result & log
			presenceRes := precedence.Result[bool]{Value: secretRes.Value != "", Source: secretRes.Source}
			precedence.LogResolved(logger, "jwt_secret_present", presenceRes)

			// multi_tenant.enabled
			_ = ResolveBoolField(logger, "multi_tenant.enabled", explicitMT, func(cc *ConsolidatedConfig) *bool { return &cc.MultiTenant.Enabled }, "COSTSCOPE_MULTI_TENANT_ENABLED", false)
			// focus.use_unified_mapper
			_ = ResolveBoolField(logger, "focus.use_unified_mapper", explicitUnified, func(cc *ConsolidatedConfig) *bool { return &cc.Focus.UseUnifiedMapperDefault }, "COSTSCOPE_USE_UNIFIED_MAPPER", false)
			// focus.invariants_enabled_default
			_ = ResolveBoolField(logger, "focus.invariants_enabled_default", explicitInv, func(cc *ConsolidatedConfig) *bool { return &cc.Focus.InvariantsEnabledDefault }, "COSTSCOPE_INVARIANTS_ENABLED", false)

			// close writer, restore stderr, read logs
			_ = w.Close()
			os.Stderr = oldStderr
			var buf bytes.Buffer
			if _, err := buf.ReadFrom(r); err != nil {
				t.Fatalf("read logs: %v", err)
			}
			raw := buf.String()
			if raw == "" {
				t.Fatalf("expected log output")
			}

			// parse lines into map[field]entry
			entries := map[string]map[string]any{}
			for _, line := range strings.Split(strings.TrimSpace(raw), "\n") {
				if strings.TrimSpace(line) == "" {
					continue
				}
				var m map[string]any
				if err := json.Unmarshal([]byte(line), &m); err != nil {
					t.Fatalf("unmarshal log line: %v line=%q", err, line)
				}
				if m["msg"] == "config_precedence_resolved" {
					if fld, ok := m["field"].(string); ok {
						entries[fld] = m
					}
				}
			}

			// expected sources & values per scenario
			want, err := expectedMatrix(sc.name)
			if err != nil {
				t.Fatalf("expectedMatrix: %v", err)
			}

			for field, exp := range want {
				m, ok := entries[field]
				if !ok {
					t.Fatalf("missing log entry for field %s\nall=%v", field, keys(entries))
				}
				gotVal, _ := m["value"].(bool)
				if gotVal != exp.val {
					t.Fatalf("%s value mismatch: want %v got %v (entry=%v)", field, exp.val, gotVal, m)
				}
				if gotSrc, _ := m["source"].(string); gotSrc != exp.source {
					t.Fatalf("%s source mismatch: want %s got %s (entry=%v)", field, exp.source, gotSrc, m)
				}
			}

			if sc.invalidEnv { // expect at least one warning line for invalid bool tokens
				if !strings.Contains(raw, "config_invalid_env_bool") {
					t.Fatalf("expected config_invalid_env_bool warning in logs, got: %s", raw)
				}
			}

			// Ensure secret material not leaked
			if strings.Contains(raw, secretExplicit) || strings.Contains(raw, secretEnv) {
				t.Fatalf("secret value leaked in logs: %q", raw)
			}
		})
	}
}

func keys[K comparable, V any](m map[K]V) []K {
	out := make([]K, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

// expected describes the expected resolved boolean plus its source.
type expected struct {
	val    bool
	source string
}

// expectedMatrix consolidates scenario -> expected map logic to keep the main
// test function's cyclomatic complexity low while preserving readability of
// scenario definitions.
func expectedMatrix(name string) (map[string]expected, error) { // small helper; returns error for unknown scenario
	var src precedence.Source
	var val bool
	switch name { // only four logical buckets
	case "default", "invalid_env":
		src = precedence.SourceDefault
		val = false
	case "env":
		src = precedence.SourceEnv
		val = true
	case "yaml":
		src = precedence.SourceYAML
		val = true
	case "explicit", "yaml_plus_explicit":
		src = precedence.SourceExplicit
		val = true
	default:
		return nil, errors.New("unknown scenario: " + name)
	}
	s := string(src)
	return map[string]expected{
		"jwt_secret_present":               {val, s},
		"multi_tenant.enabled":             {val, s},
		"focus.use_unified_mapper":         {val, s},
		"focus.invariants_enabled_default": {val, s},
	}, nil
}
