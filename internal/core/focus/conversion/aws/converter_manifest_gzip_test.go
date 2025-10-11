package aws

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"local/costscope/internal/core/focus/types"
)

func TestAWSManifest_GzipInput(t *testing.T) {
	// Prepare a tiny CUR CSV.gz
	header := strings.Join([]string{
		"bill/BillingAccountId",
		"bill/BillingAccountName",
		"bill/BillingCurrency",
		"lineItem/UnblendedCost",
		"lineItem/UsageAmount",
		"lineItem/UsageStartDate",
		"lineItem/UsageEndDate",
		"lineItem/LineItemDescription",
		"lineItem/Operation",
		"lineItem/UsageType",
		"product/ProductName",
		"product/ProductFamily",
		"lineItem/ResourceId",
		"pricing/PriceId",
		"lineItem/UsageAccountId",
	}, ",")

	row := strings.Join([]string{
		"123456789012",
		"Master",
		"USD",
		"2.50",
		"5",
		"2024-02-01 00:00:00",
		"2024-02-01 01:00:00",
		"S3 storage",
		"PutObject",
		"USW2-TimedStorage-ByteHrs",
		"AmazonS3",
		"Storage",
		"bucket-abc",
		"price-2",
		"222222222222",
	}, ",")

	csv := strings.Join([]string{header, row}, "\n")

	tmp := t.TempDir()
	dataDir := filepath.Join(tmp, "data")
	if err := os.MkdirAll(dataDir, 0o750); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	// Write report file at reportKeys location
	reportRel := filepath.Join("20240201-20240228", "000000.csv.gz")
	reportPath := filepath.Join(dataDir, reportRel)
	if err := os.MkdirAll(filepath.Dir(reportPath), 0o750); err != nil {
		t.Fatalf("mkdir report dir: %v", err)
	}
	// #nosec G304 - writing test file under temp dir
	rf, err := os.Create(reportPath)
	if err != nil {
		t.Fatalf("create report: %v", err)
	}
	rgz := gzip.NewWriter(rf)
	if _, err := rgz.Write([]byte(csv)); err != nil {
		t.Fatalf("write csv.gz: %v", err)
	}
	_ = rgz.Close()
	_ = rf.Close()

	// Build manifest.json and gzip it
	manifest := types.CURManifest{
		AssemblyId:  "asm-1",
		Account:     "123456789012",
		Columns:     nil,
		Charset:     "UTF-8",
		Compression: "GZIP",
		ContentType: "text/csv",
		ReportId:    "r-1",
		ReportName:  "cur-test",
		ReportKeys:  []string{filepath.Join("data", reportRel)},
	}
	manifest.BillingPeriod.Start = "20240201"
	manifest.BillingPeriod.End = "20240228"

	manifestPath := filepath.Join(tmp, "manifest.json.gz")
	// #nosec G304 - writing test file under temp dir
	mf, err := os.Create(manifestPath)
	if err != nil {
		t.Fatalf("create manifest.gz: %v", err)
	}
	mgz := gzip.NewWriter(mf)
	enc := json.NewEncoder(mgz)
	if err := enc.Encode(manifest); err != nil {
		t.Fatalf("encode manifest: %v", err)
	}
	_ = mgz.Close()
	_ = mf.Close()

	out := filepath.Join(tmp, "out.ndjson")

	conv := NewAWSConverter()
	cfg := &types.ConversionConfig{
		Provider:     "aws",
		InputPath:    manifestPath,
		OutputPath:   out,
		Streaming:    true,
		ChunkSize:    1000,
		Workers:      1,
		ConversionId: "test-aws-manifest-gzip",
	}

	if err := conv.ValidateInput(context.Background(), cfg); err != nil {
		t.Fatalf("validate: %v", err)
	}

	res, err := conv.Convert(context.Background(), cfg)
	if err != nil {
		t.Fatalf("convert: %v", err)
	}
	if !res.Success || res.OutputRecords != 1 {
		t.Fatalf("unexpected result: %+v", res)
	}

	// With manifest processing we write per-file outputs suffixed by the report file stem
	perFile := filepath.Join(tmp, "out_000000.ndjson")
	// #nosec G304 - reading file created in temp dir
	if b, err := os.ReadFile(perFile); err != nil || len(b) == 0 {
		t.Fatalf("expected non-empty output: %v, size=%d", err, len(b))
	}
}
