package exporters

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"local/costscope/internal/core/reports/types"
)

// HTTPExporter sends the report as JSON to an HTTP endpoint via POST or PUT with retries and auth headers.
type HTTPExporter struct {
	// Optional custom client; if nil, a default with timeout is used
	Client *http.Client
}

func NewHTTPExporter() *HTTPExporter { return &HTTPExporter{} }

func (e *HTTPExporter) Export(ctx context.Context, report interface{}, format types.ExportFormat, output string) (int64, string, error) {
	if format != types.ExportFormatHTTP {
		return 0, "", fmt.Errorf("unsupported format for HTTP exporter: %s", format)
	}
	if output == "" {
		return 0, "", errors.New("http exporter requires output as URL")
	}
	// Marshal as JSON body
	body, err := json.Marshal(report)
	if err != nil {
		return 0, "", fmt.Errorf("marshal report: %w", err)
	}
	// Use POST for report ingestion/export by default
	req, err := http.NewRequest(http.MethodPost, output, bytes.NewReader(body))
	if err != nil {
		return 0, "", err
	}
	req.Header.Set("Content-Type", "application/json")
	// Auth headers from context or env fallbacks (for CLI/API process-level config)
	if v, ok := ctx.Value(apiKeyKey{}).(string); ok && v != "" {
		req.Header.Set("X-API-Key", v)
	} else if ev := os.Getenv("COSTSCOPE_HTTP_API_KEY"); ev != "" {
		req.Header.Set("X-API-Key", ev)
	}
	if v, ok := ctx.Value(bearerTokenKey{}).(string); ok && v != "" {
		req.Header.Set("Authorization", "Bearer "+v)
	} else if ev := os.Getenv("COSTSCOPE_HTTP_BEARER_TOKEN"); ev != "" {
		req.Header.Set("Authorization", "Bearer "+ev)
	}

	client := e.Client
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}

	// Retries with backoff
	maxRetries := 3
	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		resp, err := httpDo(ctx, req.Clone(ctx), client)
		if err == nil {
			_ = resp.Body.Close()
			return int64(len(body)), "", nil // checksum intentionally empty (not materialized post-send)
		}
		lastErr = err
		// backoff
		time.Sleep(time.Duration(attempt+1) * 500 * time.Millisecond)
	}
	return 0, "", fmt.Errorf("http export failed after %d retries: %w", maxRetries, lastErr)
}

func (e *HTTPExporter) GetSupportedFormats() []types.ExportFormat {
	return []types.ExportFormat{types.ExportFormatHTTP}
}

// Context keys (auth only)
type apiKeyKey struct{}
type bearerTokenKey struct{}

// Helpers to attach auth options to context
func WithAPIKey(ctx context.Context, key string) context.Context {
	return context.WithValue(ctx, apiKeyKey{}, key)
}
func WithBearerToken(ctx context.Context, tok string) context.Context {
	return context.WithValue(ctx, bearerTokenKey{}, tok)
}
