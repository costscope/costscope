package config

import (
	"time"

	"local/costscope/internal/core/config/precedence"
	"local/costscope/internal/core/logging"
	"os"
	"strings"
)

// ResolveIntField is the unified high-level helper for resolving an int configuration field
// with full precedence (explicit > YAML default > ENV > fallback) and audit logging.
// It replaces the previous EnvResolver.ResolveInt method to reduce API duplication.
//
// Parameters:
//   - logger: unified logger (may be nil to disable logging)
//   - field: logical config field name (e.g. "streaming.max_concurrent_jobs") used in audit log
//   - explicit: pointer to explicit value (CLI/API). If non-nil it wins even if 0.
//   - yamlSelect: optional selector extracting a *int from ConsolidatedConfig; only invoked when explicit is nil.
//   - envKey: environment variable name used if explicit and YAML absent.
//   - fallback: default value when no other source provided.
//
// Returns precedence.Result[int] including value and source (explicit|yaml|env|default).
func ResolveIntField(logger *logging.Logger, field string, explicit *int, yamlSelect func(*ConsolidatedConfig) *int, envKey string, fallback int) precedence.Result[int] {
	var yamlPtr *int
	if explicit == nil && yamlSelect != nil {
		if cfg := LoadOptionalYAML(logger); cfg != nil { // reuse optional loader
			yamlPtr = yamlSelect(cfg)
		}
	}
	res := precedence.ResolveInt(explicit, yamlPtr, envKey, fallback)
	precedence.LogResolved(logger, field, res)
	return res
}

// ResolveBoolField high-level helper for booleans (explicit > YAML > ENV > fallback). False is meaningful.
func ResolveBoolField(logger *logging.Logger, field string, explicit *bool, yamlSelect func(*ConsolidatedConfig) *bool, envKey string, fallback bool) precedence.Result[bool] {
	var yamlPtr *bool
	if explicit == nil && yamlSelect != nil {
		if cfg := LoadOptionalYAML(logger); cfg != nil {
			yamlPtr = yamlSelect(cfg)
		}
	}
	// Detect invalid env token (present but not truthy/falsey) for observability. We only warn when
	// explicit and YAML are absent (otherwise env is ignored anyway).
	if explicit == nil && yamlPtr == nil {
		if raw, ok := os.LookupEnv(envKey); ok && raw != "" {
			low := strings.ToLower(strings.TrimSpace(raw))
			valid := map[string]struct{}{"1": {}, "true": {}, "t": {}, "yes": {}, "y": {}, "on": {}, "enable": {}, "enabled": {}, "0": {}, "false": {}, "f": {}, "no": {}, "n": {}, "off": {}, "disable": {}, "disabled": {}}
			if _, ok := valid[low]; !ok {
				if logger != nil {
					logger.WarnWithFields("config_invalid_env_bool", map[string]interface{}{"field": field, "env": envKey, "raw": "[REDACTED]"})
				}
			}
		}
	}
	res := precedence.ResolveBool(explicit, yamlPtr, envKey, fallback)
	precedence.LogResolved(logger, field, res)
	return res
}

// ResolveStringField high-level helper for strings (empty = not provided for explicit/YAML).
func ResolveStringField(logger *logging.Logger, field string, explicit *string, yamlSelect func(*ConsolidatedConfig) *string, envKey, fallback string) precedence.Result[string] {
	var yamlPtr *string
	if explicit == nil && yamlSelect != nil {
		if cfg := LoadOptionalYAML(logger); cfg != nil {
			yamlPtr = yamlSelect(cfg)
		}
	}
	res := precedence.ResolveString(explicit, yamlPtr, envKey, fallback)
	precedence.LogResolved(logger, field, res)
	return res
}

// ResolveDurationField helper for durations; zero duration is meaningful explicit/YAML.
func ResolveDurationField(logger *logging.Logger, field string, explicit *time.Duration, yamlSelect func(*ConsolidatedConfig) *time.Duration, envKey string, fallback time.Duration) precedence.Result[time.Duration] {
	var yamlPtr *time.Duration
	if explicit == nil && yamlSelect != nil {
		if cfg := LoadOptionalYAML(logger); cfg != nil {
			yamlPtr = yamlSelect(cfg)
		}
	}
	res := precedence.ResolveDuration(explicit, yamlPtr, envKey, fallback)
	precedence.LogResolved(logger, field, res)
	return res
}

// ResolveFloatField helper for float64; zero is meaningful explicit/YAML.
func ResolveFloatField(logger *logging.Logger, field string, explicit *float64, yamlSelect func(*ConsolidatedConfig) *float64, envKey string, fallback float64) precedence.Result[float64] {
	var yamlPtr *float64
	if explicit == nil && yamlSelect != nil {
		if cfg := LoadOptionalYAML(logger); cfg != nil {
			yamlPtr = yamlSelect(cfg)
		}
	}
	res := precedence.ResolveFloat(explicit, yamlPtr, envKey, fallback)
	precedence.LogResolved(logger, field, res)
	return res
}
