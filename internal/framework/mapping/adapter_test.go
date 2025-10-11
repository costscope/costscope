package mapping

import (
	ftypes "local/costscope/internal/core/focus/types"
	"testing"
)

func TestAdapter_FromRules_BasicMappingAndTypes(t *testing.T) {
	rules := &ftypes.MappingRules{
		Provider: "aws",
		Version:  "1.0",
		FieldMaps: map[string]ftypes.FieldMapping{
			"billing_account_id": {SourceField: "bill/BillingAccountId", TargetField: "billing_account_id", Required: true},
			"effective_cost":     {SourceField: "lineItem/UnblendedCost", TargetField: "effective_cost", Required: true},
			"usage_quantity":     {SourceField: "lineItem/UsageAmount", TargetField: "usage_quantity", Required: true},
		},
	}

	cfg := NewAdapter().FromRules(rules)
	if cfg.ProviderName != "aws" {
		t.Fatalf("ProviderName mismatch: %s", cfg.ProviderName)
	}

	// Keys should be FocusRecord field names
	if _, ok := cfg.FieldMappings["BillingAccountId"]; !ok {
		t.Fatalf("expected mapping for BillingAccountId")
	}
	if _, ok := cfg.FieldMappings["EffectiveCost"]; !ok {
		t.Fatalf("expected mapping for EffectiveCost")
	}
	if _, ok := cfg.FieldMappings["UsageQuantity"]; !ok {
		t.Fatalf("expected mapping for UsageQuantity")
	}

	// Type inference
	if cfg.FieldMappings["EffectiveCost"].FieldType != FieldTypeFloat64 {
		t.Fatalf("EffectiveCost type expected float64, got %s", cfg.FieldMappings["EffectiveCost"].FieldType)
	}
	if cfg.FieldMappings["UsageQuantity"].FieldType != FieldTypeFloat64 {
		t.Fatalf("UsageQuantity type expected float64, got %s", cfg.FieldMappings["UsageQuantity"].FieldType)
	}
}
