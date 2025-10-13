package exporters

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/costscope/costscope/internal/core/reports/types"
)

// TestHTTPExporter_RetryFailure exercises the retry loop and final error path
func TestHTTPExporter_RetryFailure(t *testing.T) {
	var calls int32
	// Always return 500 so httpDo returns status error
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(500)
	}))
	defer srv.Close()

	exp := NewHTTPExporter()
	// Inject fast timeout client so we don't block long on retries
	exp.Client = &http.Client{Timeout: 2 * time.Second}
	_, _, err := exp.Export(context.Background(), map[string]any{"id": 1}, types.ExportFormatHTTP, srv.URL)
	if err == nil {
		t.Fatalf("expected error")
	}
	// Should attempt maxRetries + 1 (4) requests
	if c := atomic.LoadInt32(&calls); c != 4 { // (attempt 0..3)
		t.Fatalf("expected 4 attempts, got %d", c)
	}
}

// TestHTTPExporter_RetryThenSuccess ensures success after transient failures
func TestHTTPExporter_RetryThenSuccess(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&calls, 1)
		if n < 2 { // first attempt fails
			w.WriteHeader(502)
			return
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte("ok"))
	}))
	defer srv.Close()

	exp := NewHTTPExporter()
	exp.Client = &http.Client{Timeout: 2 * time.Second}
	if _, _, err := exp.Export(context.Background(), map[string]any{"id": 2}, types.ExportFormatHTTP, srv.URL); err != nil {
		t.Fatalf("unexpected error after retries: %v", err)
	}
	if c := atomic.LoadInt32(&calls); c != 2 { // one failure then success
		t.Fatalf("expected 2 attempts, got %d", c)
	}
}

// TestHTTPExporter_UnsupportedFormat guards unsupported format branch
func TestHTTPExporter_UnsupportedFormat(t *testing.T) {
	exp := NewHTTPExporter()
	if _, _, err := exp.Export(context.Background(), map[string]any{}, types.ExportFormatJSON, "http://example"); err == nil {
		t.Fatalf("expected unsupported format error")
	}
}

// TestHTTPExporter_EmptyOutput verifies validation of empty URL
func TestHTTPExporter_EmptyOutput(t *testing.T) {
	exp := NewHTTPExporter()
	if _, _, err := exp.Export(context.Background(), map[string]any{}, types.ExportFormatHTTP, ""); err == nil {
		t.Fatalf("expected empty output error")
	}
}

// Compile time assertion helper to avoid unused warnings for local errors
var _ = errors.New
