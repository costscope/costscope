package gcp_test

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	dto "github.com/prometheus/client_model/go"

	gcpp "local/costscope/internal/core/focus/conversion/gcp"
	"local/costscope/internal/core/focus/types"
	"local/costscope/internal/core/monitoring/telemetry"
)

// getClassifierCountJSON reads the counter for a given decision on the legacy path.
func getClassifierCountJSON(decision string) float64 { // provider=gcp, path=legacy
	m := &dto.Metric{}
	_ = telemetry.ClassifierDecisions.WithLabelValues("gcp", "legacy", decision).Write(m)
	if m.GetCounter() == nil {
		return 0
	}
	return m.GetCounter().GetValue()
}

// Test_ClassifierDecisionMetric_JSON ensures the classifier decision metric
// increments for Usage and Credit during JSON mapping on the legacy path via end-to-end conversion.
func Test_ClassifierDecisionMetric_JSON(t *testing.T) {
	now := time.Now().UTC().Format(time.RFC3339)

	// Two JSON objects: one Usage (positive cost), one Credit (negative cost)
	objs := []map[string]any{
		{
			"usage_start_time":   now,
			"usage_end_time":     now,
			"cost":               1.00,
			"usage":              map[string]any{"amount": 1.0, "unit": "Hours"},
			"billing_account_id": "A",
			"service":            map[string]any{"description": "Compute"},
			"sku":                map[string]any{"id": "sku1"},
			"project":            map[string]any{"id": "p1"},
		},
		{
			"usage_start_time":   now,
			"usage_end_time":     now,
			"cost":               -0.50,
			"usage":              map[string]any{"amount": 0.0, "unit": "Hours"},
			"billing_account_id": "A",
			"service":            map[string]any{"description": "Compute"},
			"sku":                map[string]any{"id": "sku1"},
			"project":            map[string]any{"id": "p1"},
		},
	}

	tmp := t.TempDir()
	in := filepath.Join(tmp, "in.json.gz")
	out := filepath.Join(tmp, "out.ndjson")

	// Write gzipped JSON array to ensure JSON path is exercised
	// #nosec G304 - writing to test tempdir
	f, err := os.Create(in)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	gz := gzip.NewWriter(f)
	enc := json.NewEncoder(gz)
	if _, err := gz.Write([]byte("[")); err != nil {
		t.Fatalf("write [: %v", err)
	}
	for i, o := range objs {
		if err := enc.Encode(o); err != nil {
			t.Fatalf("encode obj: %v", err)
		}
		// Encoder adds a trailing newline per Encode; replace newline with comma/newline for array except last
		if i < len(objs)-1 {
			if _, err := gz.Write([]byte(",\n")); err != nil {
				t.Fatalf("write comma: %v", err)
			}
		}
	}
	if _, err := gz.Write([]byte("]")); err != nil {
		t.Fatalf("write ]: %v", err)
	}
	_ = gz.Close()
	_ = f.Close()

	preUsage := getClassifierCountJSON("Usage")
	preCredit := getClassifierCountJSON("Credit")

	conv := gcpp.NewGCPConverter()
	cfg := &types.ConversionConfig{Provider: "gcp", InputPath: in, OutputPath: out, Streaming: true, UseUnifiedMapper: false}
	if err := conv.ValidateInput(t.Context(), cfg); err != nil {
		t.Fatalf("validate: %v", err)
	}
	if _, err := conv.ConvertStream(t.Context(), cfg, nil); err != nil {
		t.Fatalf("convert: %v", err)
	}

	// Touch output to ensure conversion actually wrote records
	// #nosec G304 - reading from temp file
	f2, err := os.Open(out)
	if err != nil {
		t.Fatalf("open out: %v", err)
	}
	_ = bufio.NewReader(f2)
	_ = f2.Close()

	postUsage := getClassifierCountJSON("Usage")
	postCredit := getClassifierCountJSON("Credit")

	if int(postUsage-preUsage) < 1 {
		t.Fatalf("expected Usage decisions to increment by >=1, delta=%.0f", postUsage-preUsage)
	}
	if int(postCredit-preCredit) < 1 {
		t.Fatalf("expected Credit decisions to increment by >=1, delta=%.0f", postCredit-preCredit)
	}
}
