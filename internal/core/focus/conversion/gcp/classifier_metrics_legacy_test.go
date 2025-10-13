package gcp_test

import (
	"bytes"
	"context"
	"encoding/csv"
	"os"
	"path/filepath"
	"testing"
	"time"

	gcpp "github.com/costscope/costscope/internal/core/focus/conversion/gcp"
	"github.com/costscope/costscope/internal/core/focus/types"
	"github.com/costscope/costscope/internal/core/monitoring/telemetry"

	dto "github.com/prometheus/client_model/go"
)

func getLegacyCount(decision string) float64 {
	m := &dto.Metric{}
	_ = telemetry.ClassifierDecisions.WithLabelValues("gcp", "legacy", decision).Write(m)
	if m.GetCounter() == nil {
		return 0
	}
	return m.GetCounter().GetValue()
}

func TestClassifierDecisionMetric_CSV_Legacy(t *testing.T) {
	headers := []string{"usage_start_time", "usage_end_time", "cost", "usage.amount", "usage.unit", "billing_account_id", "service.description", "sku.id", "project.id"}
	now := time.Now().UTC().Format(time.RFC3339)
	rows := [][]string{{now, now, "1.00", "1", "Hours", "A", "Compute", "sku1", "p1"}, {now, now, "-0.50", "0", "Hours", "A", "Compute", "sku1", "p1"}}
	buf := &bytes.Buffer{}
	w := csv.NewWriter(buf)
	_ = w.Write(headers)
	for _, r := range rows {
		_ = w.Write(r)
	}
	w.Flush()

	tmp := t.TempDir()
	in := filepath.Join(tmp, "in.csv")
	out := filepath.Join(tmp, "out.ndjson")
	if err := os.WriteFile(in, buf.Bytes(), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}

	preU, preC := getLegacyCount("Usage"), getLegacyCount("Credit")

	convr := gcpp.NewGCPConverter()
	cfg := &types.ConversionConfig{Provider: "gcp", InputPath: in, OutputPath: out, Streaming: true}
	if err := convr.ValidateInput(context.Background(), cfg); err != nil {
		t.Fatalf("validate: %v", err)
	}
	if _, err := convr.ConvertStream(context.Background(), cfg, nil); err != nil {
		t.Fatalf("convert: %v", err)
	}

	postU, postC := getLegacyCount("Usage"), getLegacyCount("Credit")
	if int(postU-preU) < 1 {
		t.Fatalf("expected legacy CSV Usage increment, delta=%.0f", postU-preU)
	}
	if int(postC-preC) < 1 {
		t.Fatalf("expected legacy CSV Credit increment, delta=%.0f", postC-preC)
	}
}
