package gcp_test

import (
	"testing"

	providergcp "github.com/costscope/costscope/internal/core/focus/conversion/gcp"
	"github.com/costscope/costscope/internal/core/focus/types"
)

func TestFullRowMapper_NegativeCostClassification(t *testing.T) {
	headers := []string{"usage_start_time", "usage_end_time", "cost", "usage.amount", "service.description", "sku.description", "usage.unit", "credits"}
	m := providergcp.NewFullRowMapper(headers)
	row := []string{"2024-01-01T00:00:00Z", "2024-01-01T01:00:00Z", "-0.50", "1", "Compute Engine", "vCPU-Hour", "hours", ""}
	fr, _, _, err := m.Map(row)
	if err != nil {
		t.Fatalf("map error: %v", err)
	}
	if fr.ChargeCategory != types.ChargeCategories.Credit {
		t.Fatalf("expected Credit for negative cost, got %q", fr.ChargeCategory)
	}
}

func TestFullRowMapper_CreditsForceCredit(t *testing.T) {
	headers := []string{"usage_start_time", "usage_end_time", "cost", "usage.amount", "service.description", "sku.description", "usage.unit", "credits"}
	m := providergcp.NewFullRowMapper(headers)
	row := []string{"2024-01-01T00:00:00Z", "2024-01-01T01:00:00Z", "0.00", "1", "Compute Engine", "vCPU-Hour", "hours", "[{\"id\":\"c1\",\"type\":\"COMMITMENT\",\"name\":\"CUD\"}]"}
	fr, _, _, err := m.Map(row)
	if err != nil {
		t.Fatalf("map error: %v", err)
	}
	if fr.ChargeCategory != types.ChargeCategories.Credit {
		t.Fatalf("expected Credit when credits present, got %q", fr.ChargeCategory)
	}
	if fr.CommitmentDiscountId == nil || *fr.CommitmentDiscountId == "" {
		t.Fatalf("expected commitment discount id populated")
	}
}

func TestFullRowMapper_SpotDetectionFromCredits(t *testing.T) {
	headers := []string{"usage_start_time", "usage_end_time", "cost", "usage.amount", "service.description", "sku.description", "usage.unit", "credits"}
	m := providergcp.NewFullRowMapper(headers)
	row := []string{"2024-01-01T00:00:00Z", "2024-01-01T01:00:00Z", "0.10", "1", "Compute Engine", "vCPU-Hour", "hours", "[{\"type\":\"SPOT\",\"name\":\"Spot VM discount\"}]"}
	fr, _, _, err := m.Map(row)
	if err != nil {
		t.Fatalf("map error: %v", err)
	}
	if fr.PricingCategory != types.PricingCategories.Spot {
		t.Fatalf("expected Spot pricing category, got %q", fr.PricingCategory)
	}
}

func TestFullRowMapper_TagsAggregated(t *testing.T) {
	headers := []string{"usage_start_time", "usage_end_time", "cost", "usage.amount", "service.description", "sku.description", "usage.unit", "labels", "system_labels", "resource.labels"}
	m := providergcp.NewFullRowMapper(headers)
	labels := "[{\"key\":\"Env\",\"value\":\"Prod\"}]"
	sys := "{\"team\":\"platform\"}"
	res := "{\"kubernetes_namespace\":\"default\"}"
	row := []string{"2024-01-01T00:00:00Z", "2024-01-01T01:00:00Z", "0.10", "1", "Compute Engine", "vCPU-Hour", "hours", labels, sys, res}
	fr, _, _, err := m.Map(row)
	if err != nil {
		t.Fatalf("map error: %v", err)
	}
	if fr.Tags["env"] != "Prod" || fr.Tags["team"] != "platform" || fr.Tags["kubernetes_namespace"] != "default" {
		t.Fatalf("expected aggregated labels in tags, got %+v", fr.Tags)
	}
}
