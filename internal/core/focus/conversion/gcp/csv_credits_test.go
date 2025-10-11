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

func TestCSV_CreditsAndCreditCategory(t *testing.T) {
	headers := []string{"billing_account_id", "billing_account_name", "currency", "project.id", "project.name", "service.description", "service.id", "sku.id", "sku.description", "usage_start_time", "usage_end_time", "usage.amount", "usage.unit", "cost", "credits"}
	record := []string{"BA-321", "Ops", "USD", "proj-csv", "CSV Project", "Compute Engine", "compute.googleapis.com", "sku-csv", "vCPU", "2024-02-01T00:00:00Z", "2024-02-01T01:00:00Z", "1", "hour", "-0.5", "[{\"id\":\"cud-123\",\"type\":\"CommittedUseDiscount\",\"name\":\"CUD-Compute\"}]"}

	tmp := t.TempDir()
	in := filepath.Join(tmp, "in.csv")
	out := filepath.Join(tmp, "out.ndjson")
	f, err := os.Create(in) // #nosec G304 - path is controlled by test TempDir
	if err != nil {
		t.Fatalf("create input: %v", err)
	}
	w := csv.NewWriter(f)
	_ = w.Write(headers)
	_ = w.Write(record)
	w.Flush()
	_ = f.Close()

	convr := gcpp.NewGCPConverter()
	cfg := &types.ConversionConfig{Provider: "gcp", InputPath: in, OutputPath: out, Streaming: true, ValidateOutput: false, ConversionId: "test-gcp-csv-credits"}

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
	if fr.ChargeCategory != types.ChargeCategories.Credit {
		t.Fatalf("expected Credit, got %s", fr.ChargeCategory)
	}
	if fr.CommitmentDiscountId == nil || *fr.CommitmentDiscountId != "cud-123" {
		t.Fatalf("expected CommitmentDiscountId=cud-123, got %v", fr.CommitmentDiscountId)
	}
}
