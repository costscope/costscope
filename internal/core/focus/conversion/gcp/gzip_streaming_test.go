package gcp_test

import (
	"compress/gzip"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	gcpp "github.com/costscope/costscope/internal/core/focus/conversion/gcp"
	"github.com/costscope/costscope/internal/core/focus/types"
)

// Ensures gzipped CSV and JSON inputs are supported end-to-end for GCP
func TestCSVStreaming_GzipInput(t *testing.T) {
	csv := strings.Join([]string{
		"billing_account_id,billing_account_name,currency,project.id,project.name,service.description,service.id,sku.id,sku.description,usage_start_time,usage_end_time,usage.amount,usage.unit,cost,labels",
		"BA-123,Main,USD,proj-1,My Project,BigQuery,bigquery.googleapis.com,12345,Analysis,2024-01-01T00:00:00Z,2024-01-01T01:00:00Z,10,GiB,5.0,{\"env\":\"dev\"}",
	}, "\n")

	tmp := t.TempDir()
	in := filepath.Join(tmp, "in.csv.gz")
	out := filepath.Join(tmp, "out.ndjson")

	f, err := os.Create(in) // #nosec G304 - path is controlled by test TempDir
	if err != nil {
		t.Fatalf("create input: %v", err)
	}
	gz := gzip.NewWriter(f)
	if _, err := gz.Write([]byte(csv)); err != nil {
		t.Fatalf("write gzip: %v", err)
	}
	_ = gz.Close()
	_ = f.Close()

	convr := gcpp.NewGCPConverter()
	cfg := &types.ConversionConfig{Provider: "gcp", InputPath: in, OutputPath: out, Streaming: true, ChunkSize: 1000, Workers: 1, ConversionId: "test-gcp-csv-gzip"}

	if err := convr.ValidateInput(context.Background(), cfg); err != nil {
		t.Fatalf("validate: %v", err)
	}
	res, err := convr.ConvertStream(context.Background(), cfg, nil)
	if err != nil {
		t.Fatalf("convert: %v", err)
	}
	if !res.Success || res.OutputRecords != 1 {
		t.Fatalf("unexpected result: %+v", res)
	}

	// nolint:gosec // path is controlled by test TempDir
	data, err := os.ReadFile(out)
	if err != nil || len(data) == 0 {
		t.Fatalf("expected output file: %v size=%d", err, len(data))
	}
}

func TestJSONStreaming_Array_Gzip(t *testing.T) {
	body := `[
        {
            "billing_account_id": "BA-123",
            "billing_account_name": "Main",
            "currency": "USD",
            "project": {"id": "proj-1", "name": "My Project"},
            "service": {"description": "BigQuery", "id": "bigquery.googleapis.com"},
            "sku": {"id": "12345", "description": "Analysis"},
            "usage_start_time": "2024-01-01T00:00:00Z",
            "usage_end_time": "2024-01-01T01:00:00Z",
            "usage": {"amount": 10, "unit": "GiB"},
            "cost": 5.0,
            "labels": {"env": "dev"}
        }
    ]`

	tmp := t.TempDir()
	in := filepath.Join(tmp, "in.json.gz")
	out := filepath.Join(tmp, "out.ndjson")

	f, err := os.Create(in) // #nosec G304 - path is controlled by test TempDir
	if err != nil {
		t.Fatalf("create input: %v", err)
	}
	gz := gzip.NewWriter(f)
	if _, err := gz.Write([]byte(body)); err != nil {
		t.Fatalf("write gzip: %v", err)
	}
	_ = gz.Close()
	_ = f.Close()

	convr := gcpp.NewGCPConverter()
	cfg := &types.ConversionConfig{Provider: "gcp", InputPath: in, OutputPath: out, Streaming: true, ChunkSize: 1000, Workers: 1, ConversionId: "test-gcp-json-gzip"}

	if err := convr.ValidateInput(context.Background(), cfg); err != nil {
		t.Fatalf("validate: %v", err)
	}
	res, err := convr.ConvertStream(context.Background(), cfg, nil)
	if err != nil {
		t.Fatalf("convert: %v", err)
	}
	if !res.Success || res.OutputRecords != 1 {
		t.Fatalf("unexpected result: %+v", res)
	}

	// nolint:gosec // path is controlled by test TempDir
	data2, err := os.ReadFile(out)
	if err != nil || len(data2) == 0 {
		t.Fatalf("expected output file: %v size=%d", err, len(data2))
	}
}
