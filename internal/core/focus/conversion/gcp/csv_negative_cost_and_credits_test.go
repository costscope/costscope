package gcp_test

import (
	"bufio"
	"context"
	"encoding/csv"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	gcpp "local/costscope/internal/core/focus/conversion/gcp"
	"local/costscope/internal/core/focus/types"
)

func TestCSV_NegativeCostAndCredits(t *testing.T) {
	headers := []string{"billing_account_id", "billing_account_name", "currency", "project.id", "project.name", "service.description", "service.id", "sku.id", "sku.description", "usage_start_time", "usage_end_time", "usage.amount", "usage.unit", "cost", "credits"}
	rows := [][]string{{"BA-321", "Ops", "USD", "proj-csv", "CSV Project", "Compute Engine", "compute.googleapis.com", "sku-csv", "vCPU", "2024-02-01T00:00:00Z", "2024-02-01T01:00:00Z", "1", "hour", "-0.5", "[{\"type\":\"CommittedUseDiscount\",\"name\":\"CUD-Compute\"}]"}}

	tmp := t.TempDir()
	in := filepath.Join(tmp, "in.csv")
	out := filepath.Join(tmp, "out.ndjson")
	f, err := os.Create(in) // #nosec G304 - path is controlled by test TempDir
	if err != nil {
		t.Fatalf("create input: %v", err)
	}
	w := csv.NewWriter(f)
	_ = w.Write(headers)
	for _, r := range rows {
		_ = w.Write(r)
	}
	w.Flush()
	_ = f.Close()

	convr := gcpp.NewGCPConverter()
	cfg := &types.ConversionConfig{Provider: "gcp", InputPath: in, OutputPath: out, Streaming: true}

	if err := convr.ValidateInput(context.Background(), cfg); err != nil {
		t.Fatalf("validate: %v", err)
	}
	if _, err := convr.ConvertStream(context.Background(), cfg, nil); err != nil {
		t.Fatalf("convert: %v", err)
	}

	f, err = os.Open(out) // #nosec G304 - path is controlled by test TempDir
	if err != nil {
		t.Fatalf("open out: %v", err)
	}
	defer func() { _ = f.Close() }()
	rd := bufio.NewReader(f)
	line, _ := rd.ReadString('\n')
	line = strings.TrimSpace(line)
	if line == "" {
		t.Fatalf("empty output")
	}

	var fr types.FocusRecord
	if err := json.Unmarshal([]byte(line), &fr); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if fr.EffectiveCost >= 0 {
		t.Fatalf("expected negative cost, got %f", fr.EffectiveCost)
	}
	if fr.ChargeCategory != types.ChargeCategories.Credit {
		t.Fatalf("expected ChargeCategory=Credit, got %s", fr.ChargeCategory)
	}
}
