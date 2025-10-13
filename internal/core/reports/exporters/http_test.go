package exporters

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/costscope/costscope/internal/core/reports/types"
)

// simpleReport is a tiny struct to send
type simpleReport struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

func TestHTTPExporter_BasicPOST(t *testing.T) {
	var gotMethod string
	var gotAuth string
	var gotBody simpleReport
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotAuth = r.Header.Get("Authorization")
		defer func() { _ = r.Body.Close() }()
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	exp := NewHTTPExporter()
	report := &simpleReport{ID: "r1", Title: "hello"}
	ctx := WithBearerToken(context.Background(), "tkn")
	if _, _, err := exp.Export(ctx, report, types.ExportFormatHTTP, server.URL); err != nil {
		t.Fatalf("export: %v", err)
	}
	if gotMethod != http.MethodPost {
		t.Fatalf("method %s", gotMethod)
	}
	if gotAuth != "Bearer tkn" {
		t.Fatalf("auth header %q", gotAuth)
	}
	if gotBody.ID != "r1" || gotBody.Title != "hello" {
		t.Fatalf("body %+v", gotBody)
	}
}

func TestHTTPExporter_WithAPIKey(t *testing.T) {
	var gotMethod string
	var gotAPIKey string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotAPIKey = r.Header.Get("X-API-Key")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	exp := NewHTTPExporter()
	report := &simpleReport{ID: "r2"}
	ctx := WithAPIKey(context.Background(), "apikey")
	if _, _, err := exp.Export(ctx, report, types.ExportFormatHTTP, server.URL+"/ingest"); err != nil {
		t.Fatalf("export: %v", err)
	}
	if gotMethod != http.MethodPost {
		t.Fatalf("method %s", gotMethod)
	}
	if gotAPIKey != "apikey" {
		t.Fatalf("api key header %q", gotAPIKey)
	}
}

func TestHTTPExporter_EnvFallback(t *testing.T) {
	t.Setenv("COSTSCOPE_HTTP_BEARER_TOKEN", "envtok")
	t.Setenv("COSTSCOPE_HTTP_API_KEY", "envkey")

	var gotAuth string
	var gotAPIKey string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotAPIKey = r.Header.Get("X-API-Key")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	exp := NewHTTPExporter()
	report := &simpleReport{ID: "r3"}
	// No With* on context; should read from env
	if _, _, err := exp.Export(context.Background(), report, types.ExportFormatHTTP, server.URL); err != nil {
		t.Fatalf("export: %v", err)
	}
	if gotAuth != "Bearer envtok" {
		t.Fatalf("auth header %q", gotAuth)
	}
	if gotAPIKey != "envkey" {
		t.Fatalf("api key header %q", gotAPIKey)
	}
}
