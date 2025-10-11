package gcp_test

import (
	"bytes"
	"context"
	"encoding/csv"
	"os"
	"path/filepath"
	"testing"
	"time"

	gcpp "local/costscope/internal/core/focus/conversion/gcp"
	"local/costscope/internal/core/focus/types"
	"local/costscope/internal/core/monitoring/telemetry"

	dto "github.com/prometheus/client_model/go"
)

func getUnifiedCount(decision string) float64 {
	m := &dto.Metric{}
	_ = telemetry.ClassifierDecisions.WithLabelValues("gcp", "unified", decision).Write(m)
	if m.GetCounter() == nil {
		return 0
	}
	return m.GetCounter().GetValue()
}

func TestClassifierDecisionMetric_Unified_CSV(t *testing.T) {
	headers := []string{"usage_start_time", "usage_end_time", "cost", "usage.amount", "usage.unit", "billing_account_id", "service.description", "sku.id", "project.id"}
	now := time.Now().UTC().Format(time.RFC3339)
	rows := [][]string{
		{now, now, "1.00", "1", "Hours", "A", "Compute", "sku1", "p1"},
		{now, now, "-0.50", "0", "Hours", "A", "Compute", "sku1", "p1"},
	}
	buf := &bytes.Buffer{}
	w := csv.NewWriter(buf)
	_ = w.Write(headers)
	for _, r := range rows {
		_ = w.Write(r)
	}
	w.Flush()

	tmp := t.TempDir()
	in := filepath.Join(tmp, "in.csv")
	if err := os.WriteFile(in, buf.Bytes(), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	out := filepath.Join(tmp, "out.ndjson")

	preU, preC := getUnifiedCount("Usage"), getUnifiedCount("Credit")

	convr := gcpp.NewGCPConverter()
	cfg := &types.ConversionConfig{Provider: "gcp", InputPath: in, OutputPath: out, Streaming: true, UseUnifiedMapper: true}
	if err := convr.ValidateInput(context.Background(), cfg); err != nil {
		t.Fatalf("validate: %v", err)
	}
	if _, err := convr.ConvertStream(context.Background(), cfg, nil); err != nil {
		t.Fatalf("convert: %v", err)
	}

	postU, postC := getUnifiedCount("Usage"), getUnifiedCount("Credit")
	if int(postU-preU) < 1 {
		t.Fatalf("expected unified CSV Usage increment, delta=%.0f", postU-preU)
	}
	if int(postC-preC) < 1 {
		t.Fatalf("expected unified CSV Credit increment, delta=%.0f", postC-preC)
	}
}

func TestClassifierDecisionMetric_Unified_JSON(t *testing.T) {
	now := time.Now().UTC().Format(time.RFC3339)
	lines := []string{
		`{"usage_start_time":"` + now + `","usage_end_time":"` + now + `","cost":1.00,"usage":{"amount":1,"unit":"Hours"},"billing_account_id":"A","service":{"description":"Compute"},"sku":{"id":"sku1"},"project":{"id":"p1"}}`,
		`{"usage_start_time":"` + now + `","usage_end_time":"` + now + `","cost":-0.50,"usage":{"amount":0,"unit":"Hours"},"billing_account_id":"A","service":{"description":"Compute"},"sku":{"id":"sku1"},"project":{"id":"p1"}}`,
	}
	tmp := t.TempDir()
	in := filepath.Join(tmp, "in.json")
	out := filepath.Join(tmp, "out.ndjson")
	var buf bytes.Buffer
	for _, l := range lines {
		buf.WriteString(l)
		buf.WriteByte('\n')
	}
	if err := os.WriteFile(in, buf.Bytes(), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}

	preU, preC := getUnifiedCount("Usage"), getUnifiedCount("Credit")

	convr := gcpp.NewGCPConverter()
	cfg := &types.ConversionConfig{Provider: "gcp", InputPath: in, OutputPath: out, Streaming: true, UseUnifiedMapper: true}
	if err := convr.ValidateInput(context.Background(), cfg); err != nil {
		t.Fatalf("validate: %v", err)
	}
	if _, err := convr.ConvertStream(context.Background(), cfg, nil); err != nil {
		t.Fatalf("convert: %v", err)
	}

	postU, postC := getUnifiedCount("Usage"), getUnifiedCount("Credit")
	if int(postU-preU) < 1 {
		t.Fatalf("expected unified JSON Usage increment, delta=%.0f", postU-preU)
	}
	if int(postC-preC) < 1 {
		t.Fatalf("expected unified JSON Credit increment, delta=%.0f", postC-preC)
	}
}
