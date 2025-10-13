package aws

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/costscope/costscope/internal/core/focus/types"
)

func TestAWSManifest_MultiReportKeys_OutputPerFile(t *testing.T) {
	mkCSV := func(cost, qty, rid string) string {
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
			cost,
			qty,
			"2024-03-01 00:00:00",
			"2024-03-01 01:00:00",
			"EC2 usage",
			"RunInstances",
			"USW2-BoxUsage:t3.nano",
			"AmazonEC2",
			"Compute",
			rid,
			"price-x",
			"333333333333",
		}, ",")
		return strings.Join([]string{header, row}, "\n")
	}

	tmp := t.TempDir()
	dataDir := filepath.Join(tmp, "data")
	if err := os.MkdirAll(dataDir, 0o750); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	// Two gzipped CSVs
	rel1 := filepath.Join("20240301-20240331", "000001.csv.gz")
	rel2 := filepath.Join("20240301-20240331", "000002.csv.gz")
	for i, rel := range []string{rel1, rel2} {
		p := filepath.Join(dataDir, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o750); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		// #nosec G304 - writing test file under temp dir
		f, err := os.Create(p)
		if err != nil {
			t.Fatalf("create: %v", err)
		}
		gz := gzip.NewWriter(f)
		// Different content so we can validate both are processed
		body := mkCSV("1.00", "1", "i-test-"+string('A'+rune(i)))
		if _, err := gz.Write([]byte(body)); err != nil {
			t.Fatalf("write: %v", err)
		}
		_ = gz.Close()
		_ = f.Close()
	}

	manifest := types.CURManifest{
		AssemblyId:  "asm-2",
		Account:     "123456789012",
		Charset:     "UTF-8",
		Compression: "GZIP",
		ContentType: "text/csv",
		ReportId:    "r-2",
		ReportName:  "cur-test-multi",
		ReportKeys:  []string{filepath.Join("data", rel1), filepath.Join("data", rel2)},
	}
	manifest.BillingPeriod.Start = "20240301"
	manifest.BillingPeriod.End = "20240331"

	mpath := filepath.Join(tmp, "manifest.json.gz")
	// #nosec G304 - writing test file under temp dir
	mf, err := os.Create(mpath)
	if err != nil {
		t.Fatalf("create manifest: %v", err)
	}
	mgz := gzip.NewWriter(mf)
	if err := json.NewEncoder(mgz).Encode(manifest); err != nil {
		t.Fatalf("encode manifest: %v", err)
	}
	_ = mgz.Close()
	_ = mf.Close()

	out := filepath.Join(tmp, "out.ndjson")
	conv := NewAWSConverter()
	cfg := &types.ConversionConfig{
		Provider:     "aws",
		InputPath:    mpath,
		OutputPath:   out,
		Streaming:    true,
		ChunkSize:    1000,
		Workers:      1,
		ConversionId: "test-aws-manifest-multi",
	}

	if err := conv.ValidateInput(context.Background(), cfg); err != nil {
		t.Fatalf("validate: %v", err)
	}
	res, err := conv.Convert(context.Background(), cfg)
	if err != nil {
		t.Fatalf("convert: %v", err)
	}
	if !res.Success || res.OutputRecords != 2 {
		t.Fatalf("unexpected result: %+v", res)
	}

	// Check both per-file outputs exist
	out1 := filepath.Join(tmp, "out_000001.ndjson")
	out2 := filepath.Join(tmp, "out_000002.ndjson")
	// #nosec G304 - reading files we created in temp dir in test
	if b, err := os.ReadFile(out1); err != nil || len(b) == 0 {
		t.Fatalf("missing or empty %s: %v", out1, err)
	}
	// #nosec G304 - reading files we created in temp dir in test
	if b, err := os.ReadFile(out2); err != nil || len(b) == 0 {
		t.Fatalf("missing or empty %s: %v", out2, err)
	}
}
