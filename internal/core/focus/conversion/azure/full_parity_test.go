package azure_test

import (
	"context"
	"fmt"
	conv "local/costscope/internal/core/focus/conversion"
	azure "local/costscope/internal/core/focus/conversion/azure"
	types "local/costscope/internal/core/focus/types"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestAzure_FullParity_LegacyVsLegacy placeholder: ensures hash stable; later second run switches to unified mapper path.
func TestAzure_FullParity_LegacyVsLegacy(t *testing.T) {
	header := "BillingAccountId,CostInBillingCurrency,Cost,AmortizedCost,Quantity,UsageStart,UsageEnd,MeterCategory,ServiceName,MeterName,ChargeType,RetailPrice,UnitOfMeasure,SubscriptionId,SubscriptionName"
	rows := []string{
		"BA-1,1.20,1.20,1.10,2,2024-01-01T00:00:00Z,2024-01-01T01:00:00Z,Compute,VM,Standard_D2s_v3,Usage,0.60,Hours,sub-1,SubOne",
		"BA-1,-0.50,-0.50,-0.50,1,2024-01-01T00:00:00Z,2024-01-01T01:00:00Z,Storage,Blob,Hot,Credit,0.50,GB,sub-1,SubOne",
		"BA-2,3.40,3.50,3.33,5,2024-01-02T00:00:00Z,2024-01-02T01:00:00Z,Database,SQL,GeneralPurpose,Usage,0.70,Hours,sub-2,SubTwo",
	}
	csvData := header + "\n" + rows[0] + "\n" + rows[1] + "\n" + rows[2] + "\n"

	tmp := t.TempDir()
	in := filepath.Join(tmp, "in.csv")
	if err := os.WriteFile(in, []byte(csvData), 0o600); err != nil {
		t.Fatalf("write csv: %v", err)
	}
	out1 := filepath.Join(tmp, "legacy1.ndjson")
	out2 := filepath.Join(tmp, "legacy2.ndjson")

	convr := azure.NewAzureConverter()
	cfg1 := &types.ConversionConfig{Provider: "azure", InputPath: in, OutputPath: out1, Streaming: true, ChunkSize: 1000, Workers: 1, ConversionId: "azure-parity-1"}
	if err := convr.ValidateInput(context.Background(), cfg1); err != nil {
		t.Fatalf("validate1: %v", err)
	}
	if _, err := convr.ConvertStream(context.Background(), cfg1, nil); err != nil {
		t.Fatalf("convert1: %v", err)
	}

	time.Sleep(5 * time.Millisecond)
	useUnified := true
	cfg2 := &types.ConversionConfig{Provider: "azure", InputPath: in, OutputPath: out2, Streaming: true, ChunkSize: 1000, Workers: 1, ConversionId: "azure-parity-2", UseUnifiedMapper: useUnified}
	if err := convr.ValidateInput(context.Background(), cfg2); err != nil {
		t.Fatalf("validate2: %v", err)
	}
	if _, err := convr.ConvertStream(context.Background(), cfg2, nil); err != nil {
		t.Fatalf("convert2: %v", err)
	}

	recs1 := readAllFocusRecordsFromNDJSONLocal(t, out1)
	recs2 := readAllFocusRecordsFromNDJSONLocal(t, out2)
	// Use lite hash parity after removing full hash implementation
	l1 := make([]conv.FocusRecordLite, 0, len(recs1))
	l2 := make([]conv.FocusRecordLite, 0, len(recs2))
	for i := range recs1 {
		l1 = append(l1, conv.FocusRecordLite{EffectiveCost: recs1[i].EffectiveCost, UsageQuantity: recs1[i].UsageQuantity, ProviderName: recs1[i].ProviderName, ServiceName: recs1[i].ServiceName, ChargeCategory: recs1[i].ChargeCategory})
		l2 = append(l2, conv.FocusRecordLite{EffectiveCost: recs2[i].EffectiveCost, UsageQuantity: recs2[i].UsageQuantity, ProviderName: recs2[i].ProviderName, ServiceName: recs2[i].ServiceName, ChargeCategory: recs2[i].ChargeCategory})
	}
	h1 := conv.HashFocusLite(l1)
	h2 := conv.HashFocusLite(l2)
	if h1 != h2 {
		// Build quick diff report for debugging parity divergence
		max := len(recs1)
		if len(recs2) < max {
			max = len(recs2)
		}
		var sb strings.Builder
		sb.WriteString("hash mismatch:\nlegacy=" + h1 + " unified=" + h2 + "\n")
		for i := 0; i < max; i++ {
			c1 := canonicalizeFR(recs1[i])
			c2 := canonicalizeFR(recs2[i])
			if c1 == c2 {
				continue
			}
			// lightweight field subset comparison
			sb.WriteString(fmt.Sprintf("record %d differs\n", i))
			sb.WriteString("legacy:  " + c1 + "\n")
			sb.WriteString("unified: " + c2 + "\n")
			break
		}
		t.Fatal(sb.String())
	}
}
