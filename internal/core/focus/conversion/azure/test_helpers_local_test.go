package azure_test

import (
	"bufio"
	"encoding/json"
	"os"
	"testing"

	"local/costscope/internal/core/focus/types"
)

// readAllFocusRecordsFromNDJSONLocal loads all NDJSON FocusRecord lines from the given path.
func readAllFocusRecordsFromNDJSONLocal(t *testing.T, path string) []*types.FocusRecord {
	t.Helper()
	// #nosec G304 - path is produced in test temp dir
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open ndjson: %v", err)
	}
	defer func() { _ = f.Close() }()
	var out []*types.FocusRecord
	s := bufio.NewScanner(f)
	for s.Scan() {
		line := s.Bytes()
		if len(line) == 0 {
			continue
		}
		var fr types.FocusRecord
		if err := json.Unmarshal(line, &fr); err != nil {
			t.Fatalf("unmarshal line: %v", err)
		}
		out = append(out, &fr)
	}
	if err := s.Err(); err != nil {
		t.Fatalf("scan: %v", err)
	}
	return out
}

// canonicalizeFR builds a stable JSON string of selected important fields for debug diffs in tests.
func canonicalizeFR(fr *types.FocusRecord) string {
	m := map[string]interface{}{
		"billing_account_id":  fr.BillingAccountId,
		"billing_currency":    fr.BillingCurrency,
		"charge_category":     fr.ChargeCategory,
		"effective_cost":      fr.EffectiveCost,
		"usage_quantity":      fr.UsageQuantity,
		"usage_unit":          fr.UsageUnit,
		"service_name":        fr.ServiceName,
		"provider_name":       fr.ProviderName,
		"charge_period_start": fr.ChargePeriodStart.UnixMilli(),
		"charge_period_end":   fr.ChargePeriodEnd.UnixMilli(),
	}
	b, _ := json.Marshal(m)
	return string(b)
}
