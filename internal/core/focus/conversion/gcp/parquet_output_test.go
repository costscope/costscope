//go:build !race
// +build !race

package gcp_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	gcpp "local/costscope/internal/core/focus/conversion/gcp"
	"local/costscope/internal/core/focus/types"
)

// Ensures GCP converter can write Parquet when requested
func Test_CSVStreaming_ParquetOutput(t *testing.T) {
	csv := strings.Join([]string{
		"billing_account_id,billing_account_name,currency,project.id,project.name,service.description,service.id,sku.id,sku.description,usage_start_time,usage_end_time,usage.amount,usage.unit,cost,labels",
		"BA-123,Main,USD,proj-1,My Project,BigQuery,bigquery.googleapis.com,12345,Analysis,2024-01-01T00:00:00Z,2024-01-01T01:00:00Z,10,GiB,5.0,{\"env\":\"dev\"}",
	}, "\n")

	tmp := t.TempDir()
	in := filepath.Join(tmp, "in.csv")
	out := filepath.Join(tmp, "out.parquet")
	// #nosec G304 - writing file within test temp directory
	if err := os.WriteFile(in, []byte(csv), 0o600); err != nil {
		t.Fatalf("write input: %v", err)
	}

	conv := gcpp.NewGCPConverter()
	cfg := &types.ConversionConfig{
		Provider:     "gcp",
		InputPath:    in,
		OutputPath:   out,
		OutputFormat: "parquet",
		Streaming:    true,
		ChunkSize:    1000,
		Workers:      1,
		ConversionId: "test-gcp-parquet",
	}
	// Disable rotation for deterministic single-file output path
	cfg.Parquet.RotateSizeBytes = -1

	if err := conv.ValidateInput(context.Background(), cfg); err != nil {
		t.Fatalf("validate: %v", err)
	}

	res, err := conv.ConvertStream(context.Background(), cfg, nil)
	if err != nil {
		t.Fatalf("convert: %v", err)
	}
	if !res.Success {
		t.Fatalf("expected success: %+v", res)
	}
	if res.OutputRecords != 1 {
		t.Fatalf("expected 1 output record, got %d", res.OutputRecords)
	}

	// #nosec G304 - stat file produced within test temp directory
	fi, err := os.Stat(out)
	if err != nil {
		t.Fatalf("expected parquet output: %v", err)
	}
	if fi.Size() == 0 {
		t.Fatalf("parquet output should not be empty")
	}
}
