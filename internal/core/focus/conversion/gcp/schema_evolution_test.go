package gcp_test

import (
	"bufio"
	"compress/gzip"
	"encoding/csv"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	gcpp "local/costscope/internal/core/focus/conversion/gcp"
	"local/costscope/internal/core/focus/types"
)

// CSV: tolerate alternate currency field (billing_currency) and labels in resource.labels
func Test_CSV_SchemaEvolution_Tolerant(t *testing.T) {
	headers := []string{"billing_account_id", "billing_account_name", "billing_currency", "project.id", "project.name", "service.description", "service.id", "sku.id", "sku.description", "usage_start_time", "usage_end_time", "usage.amount", "usage.unit", "cost", "resource.labels"}
	record := []string{"BA-X", "NameX", "USD", "p1", "n1", "BigQuery", "bigquery.googleapis.com", "sku-1", "Analysis", "2024-01-02T00:00:00Z", "2024-01-02T01:00:00Z", "2", "GiB", "10.5", "{\"team\":\"alpha\"}"}
	tmp := t.TempDir()
	in := filepath.Join(tmp, "in.csv")
	out := filepath.Join(tmp, "out.ndjson")
	// #nosec G304 - writing to test tempdir
	f, err := os.Create(in)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	w := csv.NewWriter(f)
	_ = w.Write(headers)
	_ = w.Write(record)
	w.Flush()
	_ = f.Close()

	conv := gcpp.NewGCPConverter()
	cfg := &types.ConversionConfig{Provider: "gcp", InputPath: in, OutputPath: out, Streaming: true}
	if err := conv.ValidateInput(t.Context(), cfg); err != nil {
		t.Fatalf("validate: %v", err)
	}
	res, err := conv.ConvertStream(t.Context(), cfg, nil)
	if err != nil {
		t.Fatalf("convert: %v", err)
	}
	if res.OutputRecords != 1 {
		t.Fatalf("want 1, got %d", res.OutputRecords)
	}
	// #nosec G304 - reading from test tempdir
	data, _ := os.ReadFile(out)
	var fr types.FocusRecord
	if err := json.Unmarshal(bytesTrimLine(data), &fr); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if fr.BillingCurrency == "" || fr.Tags["team"] != "alpha" {
		t.Fatalf("missing tolerant fields: currency=%q tags=%v", fr.BillingCurrency, fr.Tags)
	}
}

// JSON Gzip: tolerate pricing.cost, labels at top-level labels and system_labels
func Test_JSON_SchemaEvolution_Gzip_Array(t *testing.T) {
	body := `[
        {"billing_account_id":"BA-Y","billing_account_name":"NameY","billing_currency":"USD","project":{"id":"p2","name":"n2"},"service":{"description":"BigQuery","id":"bigquery.googleapis.com"},"sku":{"id":"sku-2","description":"Storage"},"usage_start_time":"2024-01-03T00:00:00Z","usage_end_time":"2024-01-03T01:00:00Z","usage":{"amount":3,"unit":"GiB"},"pricing":{"cost":7.5},"labels":{"env":"prod"},"system_labels":[{"key":"owner","value":"team-b"}]}
    ]`
	tmp := t.TempDir()
	in := filepath.Join(tmp, "in.json.gz")
	out := filepath.Join(tmp, "out.ndjson")
	// #nosec G304 - writing to test tempdir
	f, _ := os.Create(in)
	gz := gzip.NewWriter(f)
	_, _ = gz.Write([]byte(body))
	_ = gz.Close()
	_ = f.Close()

	conv := gcpp.NewGCPConverter()
	cfg := &types.ConversionConfig{Provider: "gcp", InputPath: in, OutputPath: out, Streaming: true}
	if err := conv.ValidateInput(t.Context(), cfg); err != nil {
		t.Fatalf("validate: %v", err)
	}
	res, err := conv.ConvertStream(t.Context(), cfg, nil)
	if err != nil {
		t.Fatalf("convert: %v", err)
	}
	if res.OutputRecords != 1 {
		t.Fatalf("want 1, got %d", res.OutputRecords)
	}
	// #nosec G304 - opening test temp file
	f2, _ := os.Open(out)
	rd := bufio.NewReader(f2)
	line, _ := rd.ReadString('\n')
	_ = f2.Close()
	line = strings.TrimSpace(line)
	var fr types.FocusRecord
	if err := json.Unmarshal([]byte(line), &fr); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if fr.EffectiveCost != 7.5 {
		t.Fatalf("want cost 7.5, got %f", fr.EffectiveCost)
	}
	if fr.Tags["env"] != "prod" || fr.Tags["owner"] != "team-b" {
		t.Fatalf("labels not merged: %+v", fr.Tags)
	}
}

// helper: trim first line of a byte slice containing ndjson
func bytesTrimLine(b []byte) []byte {
	s := string(b)
	i := strings.IndexByte(s, '\n')
	if i == -1 {
		return []byte(strings.TrimSpace(s))
	}
	return []byte(strings.TrimSpace(s[:i]))
}
