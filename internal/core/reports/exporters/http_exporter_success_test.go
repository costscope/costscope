package exporters

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/costscope/costscope/internal/core/reports/types"
)

// TestHTTPExporter_SimpleSuccess covers 200 branch without retries.
func TestHTTPExporter_SimpleSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			t.Fatalf("missing content-type")
		}
		w.WriteHeader(200)
	}))
	defer srv.Close()
	exp := NewHTTPExporter()
	if _, _, err := exp.Export(context.Background(), map[string]any{"ok": true}, types.ExportFormatHTTP, srv.URL); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
}
