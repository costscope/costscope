package conversion

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"local/costscope/internal/core/focus/types"
)

// TestLintUtilization references shared helpers to avoid unused warnings from the linter.
func TestLintUtilization(t *testing.T) {
	// 1) Exercise readAllFocusRecordsFromNDJSON
	tmp := t.TempDir()
	p := filepath.Join(tmp, "one.ndjson")
	if err := os.WriteFile(p, []byte(`{"effective_cost":1,"billing_account_id":"A"}
`), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	recs := readAllFocusRecordsFromNDJSON(t, p)
	if len(recs) != 1 {
		t.Fatalf("expected 1 record, got %d", len(recs))
	}

	// 2) Exercise eqString/eqFloat
	a := map[string]interface{}{"s": "x", "f": 1.0}
	b := map[string]interface{}{"s": "x", "f": 1.0}
	eqString(t, a, b, "s")
	eqFloat(t, a, b, "f")

	// 3) Exercise toFloat on json.Number
	if f, ok := toFloat(json.Number("2.5")); !ok || f != 2.5 {
		t.Fatalf("toFloat failed: %v %v", f, ok)
	}

	// 4) Exercise fakeWriter methods
	var w fakeWriter
	if err := w.Open(context.Background(), "", &types.FocusSchema{}); err != nil {
		t.Fatalf("open: %v", err)
	}
	if err := w.Write(context.Background(), []types.FocusRecord{}); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := w.WriteChunk(context.Background(), nil); err != nil {
		t.Fatalf("writechunk: %v", err)
	}
	if err := w.Flush(context.Background()); err != nil {
		t.Fatalf("flush: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
	_ = w.GetMetadata()
}
