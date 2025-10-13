package gcp_test

import (
	"testing"

	providergcp "github.com/costscope/costscope/internal/core/focus/conversion/gcp"
	ftypes "github.com/costscope/costscope/internal/core/focus/types"
)

func TestRowMapper_MapRaw(t *testing.T) {
	headers := []string{"project.id", "service.description", "sku.description", "usage.amount", "usage.unit", "cost", "currency", "credits"}
	// Use FullRowMapper to validate key fields instead of the legacy lite wrapper.
	mapper := providergcp.NewFullRowMapper(headers)
	row := []string{"p1", "Compute Engine", "N1 Standard Core", "12.5", "hours", "34.56", "USD", ""}
	fr, _, _, err := mapper.Map(row)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fr.EffectiveCost != 34.56 {
		t.Fatalf("expected cost 34.56 got %v", fr.EffectiveCost)
	}
	if fr.UsageQuantity != 12.5 {
		t.Fatalf("expected usage 12.5 got %v", fr.UsageQuantity)
	}
	if fr.ServiceName == "" {
		t.Fatalf("service name empty")
	}
	if fr.ChargeCategory != "Usage" {
		t.Fatalf("expected Usage got %s", fr.ChargeCategory)
	}
	if fr.ProviderName != ftypes.ProviderNames.GCP {
		t.Fatalf("provider mismatch: want %s got %s", ftypes.ProviderNames.GCP, fr.ProviderName)
	}
	if fr.SourceProvider != "gcp" {
		t.Fatalf("source provider mismatch: want gcp got %s", fr.SourceProvider)
	}
}

func TestRowMapper_CreditDetection(t *testing.T) {
	headers := []string{"project.id", "service.description", "sku.description", "usage.amount", "usage.unit", "cost", "currency", "credits"}
	mapper := providergcp.NewFullRowMapper(headers)
	rowNeg := []string{"p1", "Compute", "SKU", "1", "h", "-1.00", "USD", ""}
	fr, _, _, _ := mapper.Map(rowNeg)
	if fr.ChargeCategory != "Credit" {
		t.Fatalf("expected Credit for negative cost")
	}
	rowCred := []string{"p1", "Compute", "SKU", "1", "h", "0.00", "USD", "[{\"type\":\"PROMO\"}]"}
	fr2, _, _, _ := mapper.Map(rowCred)
	if fr2.ChargeCategory != "Credit" {
		t.Fatalf("expected Credit for credits JSON")
	}
}
