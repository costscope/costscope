package azure_test

import (
	"bytes"
	"encoding/csv"
	azure "local/costscope/internal/core/focus/conversion/azure"
	"os"
	"path/filepath"
	"testing"
	"time"

	dto "github.com/prometheus/client_model/go"

	"local/costscope/internal/core/focus/types"
	"local/costscope/internal/core/monitoring/telemetry"
)

func getAzureUnifiedCount(_ string, _ string, decision string) float64 { // provider=azure, path=unified
	m := &dto.Metric{}
	_ = telemetry.ClassifierDecisions.WithLabelValues("azure", "unified", decision).Write(m)
	if m.GetCounter() == nil {
		return 0
	}
	return m.GetCounter().GetValue()
}

func TestAzure_ClassifierDecisionMetric_Unified_CSV(t *testing.T) {
	az := azure.NewAzureConverter()
	headers := []string{
		"UsageStart", "UsageEnd", "AmortizedCost", "Quantity", "UnitOfMeasure",
		"BillingAccountId", "ServiceName", "SubscriptionId",
	}
	now := time.Now().UTC().Format(time.RFC3339)
	rows := [][]string{
		{now, now, "1.00", "1", "Hours", "A", "Compute", "sub1"},
		{now, now, "-0.50", "0", "Hours", "A", "Compute", "sub1"},
	}
	buf := &bytes.Buffer{}
	cw := csv.NewWriter(buf)
	_ = cw.Write(headers)
	for _, r := range rows {
		_ = cw.Write(r)
	}
	cw.Flush()

	preUsage := getAzureUnifiedCount("azure", "unified", "Usage")
	preCredit := getAzureUnifiedCount("azure", "unified", "Credit")

	tmp := t.TempDir()
	in := filepath.Join(tmp, "in.csv")
	out := filepath.Join(tmp, "out.ndjson")
	if err := os.WriteFile(in, buf.Bytes(), 0o600); err != nil {
		t.Fatalf("write input: %v", err)
	}
	cfg := &types.ConversionConfig{InputPath: in, OutputPath: out, UseUnifiedMapper: true, Streaming: true}
	if _, err := az.ConvertStream(t.Context(), cfg, nil); err != nil {
		t.Fatalf("ConvertStream error: %v", err)
	}

	postUsage := getAzureUnifiedCount("azure", "unified", "Usage")
	postCredit := getAzureUnifiedCount("azure", "unified", "Credit")
	if int(postUsage-preUsage) < 1 {
		t.Fatalf("expected unified CSV Usage increment, delta=%.0f", postUsage-preUsage)
	}
	if int(postCredit-preCredit) < 1 {
		t.Fatalf("expected unified CSV Credit increment, delta=%.0f", postCredit-preCredit)
	}
}

func TestAzure_ClassifierDecisionMetric_Unified_JSON(t *testing.T) {
	az := azure.NewAzureConverter()
	now := time.Now().UTC().Format(time.RFC3339)
	lines := []string{
		`{"UsageStart":"` + now + `","UsageEnd":"` + now + `","AmortizedCost":1.00,"Quantity":1,"UnitOfMeasure":"Hours","BillingAccountId":"A","ServiceName":"Compute","SubscriptionId":"sub1"}`,
		`{"UsageStart":"` + now + `","UsageEnd":"` + now + `","AmortizedCost":-0.50,"Quantity":0,"UnitOfMeasure":"Hours","BillingAccountId":"A","ServiceName":"Compute","SubscriptionId":"sub1"}`,
	}
	var b bytes.Buffer
	for _, l := range lines {
		b.WriteString(l)
		b.WriteByte('\n')
	}

	preUsage := getAzureUnifiedCount("azure", "unified", "Usage")
	preCredit := getAzureUnifiedCount("azure", "unified", "Credit")

	tmp := t.TempDir()
	in := filepath.Join(tmp, "in.json")
	out := filepath.Join(tmp, "out.ndjson")
	if err := os.WriteFile(in, b.Bytes(), 0o600); err != nil {
		t.Fatalf("write input: %v", err)
	}
	cfg := &types.ConversionConfig{InputPath: in, OutputPath: out, UseUnifiedMapper: true, Streaming: true}
	if _, err := az.ConvertStream(t.Context(), cfg, nil); err != nil {
		t.Fatalf("ConvertStream error: %v", err)
	}

	postUsage := getAzureUnifiedCount("azure", "unified", "Usage")
	postCredit := getAzureUnifiedCount("azure", "unified", "Credit")
	if int(postUsage-preUsage) < 1 {
		t.Fatalf("expected unified JSON Usage increment, delta=%.0f", postUsage-preUsage)
	}
	if int(postCredit-preCredit) < 1 {
		t.Fatalf("expected unified JSON Credit increment, delta=%.0f", postCredit-preCredit)
	}
}
