package precedence

// Unified deterministic config precedence resolution utilities.
// Order: Explicit (highest – CLI flag / API request) > YAML config default > Environment variable > Fallback default.
// Callers pass pointers for explicit + YAML values; nil means not provided.
// Environment variables are looked up lazily inside the resolver.
// Logging helpers provide structured audit with masking for secret-like fields.

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/costscope/costscope/internal/core/logging"
)

// Source indicates the provenance of a resolved value.
// exported for diagnostics in tests or logging.
type Source string

const (
	SourceExplicit Source = "explicit"
	SourceYAML     Source = "yaml"
	SourceEnv      Source = "env"
	SourceDefault  Source = "default"
)

// Result wraps a resolved value with its source.
type Result[T any] struct {
	Value  T
	Source Source
}

// ResolveBool applies precedence for booleans.
func ResolveBool(explicit *bool, yaml *bool, envKey string, fallback bool) Result[bool] {
	if explicit != nil { // explicit always wins even if false
		return Result[bool]{Value: *explicit, Source: SourceExplicit}
	}
	if yaml != nil { // yaml wins over env
		return Result[bool]{Value: *yaml, Source: SourceYAML}
	}
	if v, ok := lookupEnvBool(envKey); ok {
		return Result[bool]{Value: v, Source: SourceEnv}
	}
	return Result[bool]{Value: fallback, Source: SourceDefault}
}

// ResolveInt applies precedence for integers (zero is a meaningful explicit value if provided).
func ResolveInt(explicit *int, yaml *int, envKey string, fallback int) Result[int] {
	if explicit != nil {
		return Result[int]{Value: *explicit, Source: SourceExplicit}
	}
	if yaml != nil {
		return Result[int]{Value: *yaml, Source: SourceYAML}
	}
	if v, ok := lookupEnvInt(envKey); ok {
		return Result[int]{Value: v, Source: SourceEnv}
	}
	return Result[int]{Value: fallback, Source: SourceDefault}
}

// ResolveFloat applies precedence for float64 values (zero is meaningful when explicitly provided).
// Environment variable parser uses strconv.ParseFloat with 64-bit precision.
func ResolveFloat(explicit *float64, yaml *float64, envKey string, fallback float64) Result[float64] {
	if explicit != nil { // explicit wins even if 0.0
		return Result[float64]{Value: *explicit, Source: SourceExplicit}
	}
	if yaml != nil { // yaml over env
		return Result[float64]{Value: *yaml, Source: SourceYAML}
	}
	if v, ok := lookupEnvFloat(envKey); ok {
		return Result[float64]{Value: v, Source: SourceEnv}
	}
	return Result[float64]{Value: fallback, Source: SourceDefault}
}

// ResolveDuration applies precedence for durations (zero duration is meaningful when explicitly provided).
// Environment variable is parsed using time.ParseDuration (supports examples like "30s", "5m", "1h30m", or "0").
func ResolveDuration(explicit *time.Duration, yaml *time.Duration, envKey string, fallback time.Duration) Result[time.Duration] {
	if explicit != nil { // explicit always wins even if 0
		return Result[time.Duration]{Value: *explicit, Source: SourceExplicit}
	}
	if yaml != nil { // yaml wins over env
		return Result[time.Duration]{Value: *yaml, Source: SourceYAML}
	}
	if v, ok := lookupEnvDuration(envKey); ok {
		return Result[time.Duration]{Value: v, Source: SourceEnv}
	}
	return Result[time.Duration]{Value: fallback, Source: SourceDefault}
}

// ResolveString applies precedence for strings (empty string does NOT count as provided).
func ResolveString(explicit *string, yaml *string, envKey string, fallback string) Result[string] {
	if explicit != nil && *explicit != "" {
		return Result[string]{Value: *explicit, Source: SourceExplicit}
	}
	if yaml != nil && *yaml != "" {
		return Result[string]{Value: *yaml, Source: SourceYAML}
	}
	if v, ok := os.LookupEnv(envKey); ok && v != "" {
		return Result[string]{Value: v, Source: SourceEnv}
	}
	return Result[string]{Value: fallback, Source: SourceDefault}
}

// LogResolved logs a resolved value with masking for sensitive fields.
// secretKeys are additional substrings (case-insensitive) considered sensitive.
func LogResolved[T any](logger *logging.Logger, field string, res Result[T], secretKeys ...string) {
	if logger == nil {
		return
	}
	masked := shouldMask(field, secretKeys)
	var display any = res.Value
	if masked {
		display = "[REDACTED]"
	}
	logger.InfoWithFields("config_precedence_resolved", map[string]interface{}{
		"field":  field,
		"value":  display,
		"source": string(res.Source),
	})
}

// ------------------------- helpers -------------------------

// internal env parsers (intentionally unexported)
// Previously thin exported wrappers (LookupEnvBool|Int|Duration|Float) existed so other
// packages could reuse the normalization of truthy/falsey sets and numeric parsing.
// After consolidating all config resolution behind Resolve* helpers + LogResolved,
// no external package consumed those wrappers. They were removed pre-1.0 to shrink
// public surface and avoid deadcode noise. If a future external use-case emerges,
// we can re-export a single focused helper.
func lookupEnvBool(key string) (bool, bool) {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return false, false
	}
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "1", "true", "t", "yes", "y", "on", "enable", "enabled":
		return true, true
	case "0", "false", "f", "no", "n", "off", "disable", "disabled":
		return false, true
	default:
		return false, false
	}
}

func lookupEnvInt(key string) (int, bool) {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return 0, false
	}
	parsed, err := strconv.Atoi(strings.TrimSpace(v))
	if err != nil {
		return 0, false
	}
	return parsed, true
}

func lookupEnvFloat(key string) (float64, bool) {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return 0, false
	}
	parsed, err := strconv.ParseFloat(strings.TrimSpace(v), 64)
	if err != nil {
		return 0, false
	}
	return parsed, true
}

func lookupEnvDuration(key string) (time.Duration, bool) {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return 0, false
	}
	d, err := time.ParseDuration(strings.TrimSpace(v))
	if err != nil {
		return 0, false
	}
	return d, true
}

func shouldMask(field string, secretKeys []string) bool {
	lf := strings.ToLower(field)
	// Special-case: presence indicator fields (suffix _present) should never be masked even
	// if they contain substrings like "secret" to allow tests / diagnostics to assert
	// boolean value without leaking underlying secret material. The actual secret field
	// (e.g. security.jwt_secret) remains masked.
	if strings.HasSuffix(lf, "_present") {
		return false
	}
	defaults := []string{"secret", "password", "token", "key"}
	if len(secretKeys) == 0 {
		secretKeys = defaults
	} else {
		secretKeys = append(secretKeys, defaults...)
	}
	for _, s := range secretKeys {
		if strings.Contains(lf, s) {
			return true
		}
	}
	return false
}

// (exported LookupEnv* wrappers removed – see comment above lookupEnvBool)
