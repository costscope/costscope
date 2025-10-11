package conversion

import (
	"bufio"
	"encoding/json"
	"local/costscope/internal/core/focus/types"
	"os"
	"testing"
)

// readAllFocusRecordsFromNDJSON loads every line as FocusRecord.
func readAllFocusRecordsFromNDJSON(t *testing.T, path string) []*types.FocusRecord {
	f, err := os.Open(path) // #nosec G304 - test helper reads fixture path controlled by tests
	if err != nil {
		t.Fatalf("open ndjson: %v", err)
	}
	t.Cleanup(func() { _ = f.Close() })
	var out []*types.FocusRecord
	s := bufio.NewScanner(f)
	for s.Scan() {
		line := s.Bytes()
		if len(line) == 0 {
			continue
		}
		var fr types.FocusRecord
		if err := json.Unmarshal(line, &fr); err != nil {
			t.Fatalf("unmarshal line: %v", err)
		}
		out = append(out, &fr)
	}
	if err := s.Err(); err != nil {
		t.Fatalf("scan: %v", err)
	}
	return out
}

// eqString asserts that a[key] and b[key] are equal strings.
func eqString(t *testing.T, a, b map[string]interface{}, key string) {
	t.Helper()
	av, aok := a[key].(string)
	bv, bok := b[key].(string)
	if !aok || !bok || av != bv {
		t.Fatalf("string field %s mismatch: %q vs %q", key, a[key], b[key])
	}
}

// eqFloat asserts that a[key] and b[key] are numerically equal within a tiny epsilon.
func eqFloat(t *testing.T, a, b map[string]interface{}, key string) {
	t.Helper()
	af, aok := toFloat(a[key])
	bf, bok := toFloat(b[key])
	if !aok || !bok {
		t.Fatalf("float field %s type mismatch: %T vs %T", key, a[key], b[key])
	}
	if (af-bf) > 1e-9 || (bf-af) > 1e-9 {
		t.Fatalf("float field %s mismatch: %v vs %v", key, af, bf)
	}
}

func toFloat(v interface{}) (float64, bool) {
	switch x := v.(type) {
	case float64:
		return x, true
	case float32:
		return float64(x), true
	case int:
		return float64(x), true
	case int64:
		return float64(x), true
	case json.Number:
		f, err := x.Float64()
		if err != nil {
			return 0, false
		}
		return f, true
	default:
		return 0, false
	}
}
