package gcp_test

import (
	"testing"

	conv "local/costscope/internal/core/focus/conversion"
	providergcp "local/costscope/internal/core/focus/conversion/gcp"
)

func TestLiteParityHash(t *testing.T) {
	headers := []string{"project.id", "service.description", "sku.description", "usage.amount", "usage.unit", "cost", "currency", "credits"}
	mapper := providergcp.NewFullRowMapper(headers)
	rows := [][]string{{"p1", "Compute", "Core", "10", "hours", "5.00", "USD", ""}, {"p1", "Compute", "Core", "2", "hours", "-1.00", "USD", "[{\"type\":\"PROMO\"}]"}}
	lite := make([]conv.FocusRecordLite, 0, len(rows))
	for _, r := range rows {
		fr, _, _, err := mapper.Map(r)
		if err != nil {
			t.Fatalf("map error: %v", err)
		}
		lite = append(lite, conv.FocusRecordLite{
			EffectiveCost:  fr.EffectiveCost,
			UsageQuantity:  fr.UsageQuantity,
			ProviderName:   fr.ProviderName,
			ServiceName:    fr.ServiceName,
			ChargeCategory: fr.ChargeCategory,
		})
	}
	if h := conv.HashFocusLite(lite); h == "" {
		t.Fatalf("empty hash")
	}
}
