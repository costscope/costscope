package precedence

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/costscope/costscope/internal/core/logging"
)

func TestResolveBool_Precedence(t *testing.T) {
	tru := true
	fal := false
	// explicit wins
	res := ResolveBool(&fal, &tru, "NON_EXIST", true)

	if res.Value != false || res.Source != SourceExplicit {
		t.Fatalf("expected explicit false, got %+v", res)
	}
	// yaml beats env
	if err := os.Setenv("TEST_BOOL_ENV", "true"); err != nil {
		t.Fatalf("set env: %v", err)
	}
	res = ResolveBool(nil, &fal, "TEST_BOOL_ENV", true)
	if res.Value != false || res.Source != SourceYAML {
		t.Fatalf("expected yaml false over env true, got %+v", res)
	}
	// env used when yaml nil
	res = ResolveBool(nil, nil, "TEST_BOOL_ENV", false)
	if res.Value != true || res.Source != SourceEnv {
		t.Fatalf("expected env true, got %+v", res)
	}
	// fallback when nothing set
	if err := os.Unsetenv("TEST_BOOL_ENV"); err != nil {
		t.Fatalf("unset env: %v", err)
	}
	res = ResolveBool(nil, nil, "TEST_BOOL_ENV", true)
	if res.Value != true || res.Source != SourceDefault {
		t.Fatalf("expected fallback true, got %+v", res)
	}
}

func TestResolveInt_Precedence(t *testing.T) {
	five := 5
	three := 3
	if err := os.Setenv("TEST_INT_ENV", "9"); err != nil {
		t.Fatalf("set env: %v", err)
	}
	// explicit
	res := ResolveInt(&three, &five, "TEST_INT_ENV", 1)
	if res.Value != 3 || res.Source != SourceExplicit {
		t.Fatalf("expected explicit 3, got %+v", res)
	}
	// yaml
	res = ResolveInt(nil, &five, "TEST_INT_ENV", 1)
	if res.Value != 5 || res.Source != SourceYAML {
		t.Fatalf("expected yaml 5, got %+v", res)
	}
	// env
	res = ResolveInt(nil, nil, "TEST_INT_ENV", 1)
	if res.Value != 9 || res.Source != SourceEnv {
		t.Fatalf("expected env 9, got %+v", res)
	}
	// fallback
	if err := os.Unsetenv("TEST_INT_ENV"); err != nil {
		t.Fatalf("unset env: %v", err)
	}
	res = ResolveInt(nil, nil, "TEST_INT_ENV", 42)
	if res.Value != 42 || res.Source != SourceDefault {
		t.Fatalf("expected fallback 42, got %+v", res)
	}
}

func TestResolveFloat_Precedence(t *testing.T) {
	twoPointFive := 2.5
	onePointOne := 1.1
	if err := os.Setenv("TEST_FLOAT_ENV", "9.75"); err != nil {
		t.Fatalf("set env: %v", err)
	}
	// explicit
	res := ResolveFloat(&onePointOne, &twoPointFive, "TEST_FLOAT_ENV", 0.5)
	if res.Value != 1.1 || res.Source != SourceExplicit {
		t.Fatalf("expected explicit 1.1 got %+v", res)
	}
	// yaml
	res = ResolveFloat(nil, &twoPointFive, "TEST_FLOAT_ENV", 0.5)
	if res.Value != 2.5 || res.Source != SourceYAML {
		t.Fatalf("expected yaml 2.5 got %+v", res)
	}
	// env
	res = ResolveFloat(nil, nil, "TEST_FLOAT_ENV", 0.5)
	if res.Value != 9.75 || res.Source != SourceEnv {
		t.Fatalf("expected env 9.75 got %+v", res)
	}
	// fallback
	if err := os.Unsetenv("TEST_FLOAT_ENV"); err != nil {
		t.Fatalf("unset env: %v", err)
	}
	res = ResolveFloat(nil, nil, "TEST_FLOAT_ENV", 0.5)
	if res.Value != 0.5 || res.Source != SourceDefault {
		t.Fatalf("expected fallback 0.5 got %+v", res)
	}
}

func TestLookupEnvFloat_InvalidValue(t *testing.T) {
	_ = os.Setenv("FLOAT_INVALID", "12.x5")
	if v, ok := lookupEnvFloat("FLOAT_INVALID"); ok {
		t.Fatalf("expected not ok for invalid float, got ok with %v", v)
	}
}

func TestResolveString_Precedence(t *testing.T) {
	foo := "foo"
	bar := "bar"
	if err := os.Setenv("TEST_STR_ENV", "envval"); err != nil {
		t.Fatalf("set env: %v", err)
	}
	// explicit
	res := ResolveString(&foo, &bar, "TEST_STR_ENV", "def")
	if res.Value != "foo" || res.Source != SourceExplicit {
		t.Fatalf("expected explicit foo, got %+v", res)
	}
	// yaml
	res = ResolveString(nil, &bar, "TEST_STR_ENV", "def")
	if res.Value != "bar" || res.Source != SourceYAML {
		t.Fatalf("expected yaml bar, got %+v", res)
	}
	// env
	res = ResolveString(nil, nil, "TEST_STR_ENV", "def")
	if res.Value != "envval" || res.Source != SourceEnv {
		t.Fatalf("expected env envval, got %+v", res)
	}
	// fallback
	if err := os.Unsetenv("TEST_STR_ENV"); err != nil {
		t.Fatalf("unset env: %v", err)
	}
	res = ResolveString(nil, nil, "TEST_STR_ENV", "fallback")
	if res.Value != "fallback" || res.Source != SourceDefault {
		t.Fatalf("expected fallback fallback, got %+v", res)
	}
}

func TestLookupEnvBoolVariants(t *testing.T) {
	cases := map[string]bool{
		"1":        true,
		"true":     true,
		"TRUE":     true,
		"yes":      true,
		"on":       true,
		"enable":   true,
		"0":        false,
		"false":    false,
		"no":       false,
		"off":      false,
		"disabled": false,
	}
	for val, exp := range cases {
		if err := os.Setenv("BOOL_CASE", val); err != nil {
			t.Fatalf("set env: %v", err)
		}
		got, ok := lookupEnvBool("BOOL_CASE")
		if !ok || got != exp {
			t.Fatalf("val %s expected %v got %v ok=%v", val, exp, got, ok)
		}
	}
}

func TestLogResolvedMasking(t *testing.T) {
	logger := logging.NewLogger(logging.LevelError)
	tru := true
	res := ResolveBool(&tru, nil, "NOPE", false)
	LogResolved(logger, "jwt_secret", res) // should be masked
}

func TestResolveString_EmptyExplicitAndYAML(t *testing.T) {
	// empty explicit should be treated as not provided
	empty := ""
	yamlv := "yaml"
	res := ResolveString(&empty, &yamlv, "NOENV", "fb")
	if res.Value != "yaml" || res.Source != SourceYAML {
		t.Fatalf("expected yaml over empty explicit, got %+v", res)
	}
	// empty yaml should be ignored; env wins if present
	_ = os.Setenv("STR_ENV_X", "envv")
	res = ResolveString(nil, &empty, "STR_ENV_X", "fb")
	if res.Value != "envv" || res.Source != SourceEnv {
		t.Fatalf("expected env over empty yaml, got %+v", res)
	}
	_ = os.Unsetenv("STR_ENV_X")
}

func TestLookupEnvBool_InvalidValue(t *testing.T) {
	_ = os.Setenv("BOOL_INVALID", "definitely-not-bool")
	if v, ok := lookupEnvBool("BOOL_INVALID"); ok {
		t.Fatalf("expected not ok for invalid bool, got ok with %v", v)
	}
}

func TestLookupEnvInt_InvalidValue(t *testing.T) {
	_ = os.Setenv("INT_INVALID", "12x")
	if v, ok := lookupEnvInt("INT_INVALID"); ok {
		t.Fatalf("expected not ok for invalid int, got ok with %v", v)
	}
}

func TestShouldMask_CustomKeys(t *testing.T) {
	if !shouldMask("my_api_token", nil) {
		t.Fatalf("expected default mask for token")
	}
	// custom key list
	if !shouldMask("random_field", []string{"random_"}) {
		t.Fatalf("expected mask with custom key substring")
	}
}

// TestLogResolved_ConfigPrecedenceResolvedLine asserts that a single structured log line
// with msg "config_precedence_resolved" is emitted and contains the expected fields.
func TestLogResolved_ConfigPrecedenceResolvedLine(t *testing.T) {
	logger := logging.NewLogger(logging.LevelInfo)

	// Capture os.Stderr where the logger writes JSON lines
	oldStderr := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stderr = w
	defer func() { os.Stderr = oldStderr }()

	// Emit a resolved config line
	field := "focus.use_unified_mapper"
	res := Result[bool]{Value: true, Source: SourceEnv}
	LogResolved(logger, field, res)

	// Close writer and read what was logged
	if err := w.Close(); err != nil { // ensure writer is closed before reading
		t.Fatalf("close writer: %v", err)
	}
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("read stderr: %v", err)
	}

	raw := strings.TrimSpace(buf.String())
	if raw == "" {
		t.Fatal("expected a log line, got empty output")
	}
	lines := strings.Split(raw, "\n")
	if len(lines) != 1 {
		t.Fatalf("expected exactly 1 log line, got %d: %q", len(lines), raw)
	}

	var entry map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &entry); err != nil {
		t.Fatalf("unmarshal log json: %v; line: %q", err, lines[0])
	}

	// Assert core fields
	if got := entry["msg"]; got != "config_precedence_resolved" {
		t.Fatalf("msg mismatch: want config_precedence_resolved, got %v", got)
	}
	if got := entry["level"]; got != string(logging.LevelInfo) {
		t.Fatalf("level mismatch: want %s, got %v", logging.LevelInfo, got)
	}
	if got := entry["field"]; got != field {
		t.Fatalf("field mismatch: want %s, got %v", field, got)
	}
	if got := entry["value"]; got != true {
		t.Fatalf("value mismatch: want true, got %v", got)
	}
	if got := entry["source"]; got != string(SourceEnv) {
		t.Fatalf("source mismatch: want %s, got %v", SourceEnv, got)
	}

	// Bonus: service base field should be present for structured logs
	if _, ok := entry["service"]; !ok {
		t.Fatalf("expected 'service' base field to be present in structured log: %v", entry)
	}
}

func TestResolveDuration_Precedence(t *testing.T) {
	explicit := 30 * time.Second
	yamlv := 10 * time.Second
	if err := os.Setenv("TEST_DURATION_ENV", "5s"); err != nil {
		t.Fatalf("set env: %v", err)
	}
	defer func() { _ = os.Unsetenv("TEST_DURATION_ENV") }()

	res := ResolveDuration(&explicit, &yamlv, "TEST_DURATION_ENV", 1*time.Second)
	if res.Value != explicit || res.Source != SourceExplicit {
		t.Fatalf("expected explicit 30s got %+v", res)
	}
	res = ResolveDuration(nil, &yamlv, "TEST_DURATION_ENV", 1*time.Second)
	if res.Value != yamlv || res.Source != SourceYAML {
		t.Fatalf("expected yaml 10s got %+v", res)
	}
	res = ResolveDuration(nil, nil, "TEST_DURATION_ENV", 1*time.Second)
	if res.Value != 5*time.Second || res.Source != SourceEnv {
		t.Fatalf("expected env 5s got %+v", res)
	}
	_ = os.Unsetenv("TEST_DURATION_ENV")
	res = ResolveDuration(nil, nil, "TEST_DURATION_ENV", 42*time.Second)
	if res.Value != 42*time.Second || res.Source != SourceDefault {
		t.Fatalf("expected fallback 42s got %+v", res)
	}
}

func TestResolveDuration_EnvParseErrorFallsBack(t *testing.T) {
	if err := os.Setenv("TEST_BAD_DURATION", "notaduration"); err != nil {
		t.Fatalf("set env: %v", err)
	}
	defer func() { _ = os.Unsetenv("TEST_BAD_DURATION") }()
	res := ResolveDuration(nil, nil, "TEST_BAD_DURATION", 7*time.Second)
	if res.Value != 7*time.Second || res.Source != SourceDefault {
		t.Fatalf("expected fallback 7s got %+v", res)
	}
}
