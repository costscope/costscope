package azure_test

import (
	"bytes"
	"encoding/csv"
	"os"
	"path/filepath"
	"testing"
	"time"

	dto "github.com/prometheus/client_model/go"

	azure "local/costscope/internal/core/focus/conversion/azure"
	"local/costscope/internal/core/focus/types"
	"local/costscope/internal/core/monitoring/telemetry"
)

// getAzureClassifierCount reads the current counter value for the given label set.
func getAzureClassifierCount(_ string, _ string, decision string) float64 { // provider=azure, path=legacy
	m := &dto.Metric{}
	_ = telemetry.ClassifierDecisions.WithLabelValues("azure", "legacy", decision).Write(m)
	if m.GetCounter() == nil {
		return 0
	}
	return m.GetCounter().GetValue()
}

// TestAzure_ClassifierDecisionMetric_CSV ensures costscope_classifier_decisions_total
// increments for Usage and Credit during CSV conversion on the legacy path.
func TestAzure_ClassifierDecisionMetric_CSV(t *testing.T) {
	az := azure.NewAzureConverter()

	headers := []string{
		"UsageStart", "UsageEnd", "AmortizedCost", "Quantity", "UnitOfMeasure",
		"BillingAccountId", "ServiceName", "SubscriptionId",
	}
	now := time.Now().UTC().Format(time.RFC3339)
	rows := [][]string{
		// Positive cost -> Usage
		{now, now, "1.00", "1", "Hours", "A", "Compute", "sub1"},
		// Negative cost -> Credit
		{now, now, "-0.50", "0", "Hours", "A", "Compute", "sub1"},
	}

	preUsage := getAzureClassifierCount("azure", "legacy", "Usage")
	preCredit := getAzureClassifierCount("azure", "legacy", "Credit")

	// processAzureCSV increments metrics; feed it a small CSV reader.
	buf := &bytes.Buffer{}
	cw := csv.NewWriter(buf)
	_ = cw.Write(headers)
	for _, r := range rows {
		_ = cw.Write(r)
	}
	cw.Flush()

	tmp := t.TempDir()
	in := filepath.Join(tmp, "in.csv")
	out := filepath.Join(tmp, "out.ndjson")
	if err := os.WriteFile(in, buf.Bytes(), 0o600); err != nil {
		t.Fatalf("write input: %v", err)
	}
	cfg := &types.ConversionConfig{InputPath: in, OutputPath: out, Streaming: true}
	if _, err := az.ConvertStream(t.Context(), cfg, nil); err != nil {
		t.Fatalf("convert: %v", err)
	}

	postUsage := getAzureClassifierCount("azure", "legacy", "Usage")
	postCredit := getAzureClassifierCount("azure", "legacy", "Credit")
	if int(postUsage-preUsage) < 1 {
		t.Fatalf("expected legacy CSV Usage increment, delta=%.0f", postUsage-preUsage)
	}
	if int(postCredit-preCredit) < 1 {
		t.Fatalf("expected legacy CSV Credit increment, delta=%.0f", postCredit-preCredit)
	}
}

// TestAzure_ClassifierDecisionMetric_JSON ensures increments for Usage and Credit
// on the legacy JSON path (NDJSON input).
func TestAzure_ClassifierDecisionMetric_JSON(t *testing.T) {
	az := azure.NewAzureConverter()
	now := time.Now().UTC().Format(time.RFC3339)

	// Two NDJSON objects: one Usage (positive cost), one Credit (negative cost)
	lines := []string{
		`{"UsageStart":"` + now + `","UsageEnd":"` + now + `","AmortizedCost":1.00,"Quantity":1,"UnitOfMeasure":"Hours","BillingAccountId":"A","ServiceName":"Compute","SubscriptionId":"sub1"}`,
		`{"UsageStart":"` + now + `","UsageEnd":"` + now + `","AmortizedCost":-0.50,"Quantity":0,"UnitOfMeasure":"Hours","BillingAccountId":"A","ServiceName":"Compute","SubscriptionId":"sub1"}`,
	}
	var b bytes.Buffer
	for _, l := range lines {
		b.WriteString(l)
		b.WriteByte('\n')
	}

	preUsage := getAzureClassifierCount("azure", "legacy", "Usage")
	preCredit := getAzureClassifierCount("azure", "legacy", "Credit")

	tmp := t.TempDir()
	in := filepath.Join(tmp, "in.json")
	out := filepath.Join(tmp, "out.ndjson")
	if err := os.WriteFile(in, b.Bytes(), 0o600); err != nil {
		t.Fatalf("write input: %v", err)
	}
	cfg := &types.ConversionConfig{InputPath: in, OutputPath: out, Streaming: true}
	if _, err := az.ConvertStream(t.Context(), cfg, nil); err != nil {
		t.Fatalf("convert: %v", err)
	}

	postUsage := getAzureClassifierCount("azure", "legacy", "Usage")
	postCredit := getAzureClassifierCount("azure", "legacy", "Credit")
	if int(postUsage-preUsage) < 1 {
		t.Fatalf("expected legacy JSON Usage increment, delta=%.0f", postUsage-preUsage)
	}
	if int(postCredit-preCredit) < 1 {
		t.Fatalf("expected legacy JSON Credit increment, delta=%.0f", postCredit-preCredit)
	}
}
