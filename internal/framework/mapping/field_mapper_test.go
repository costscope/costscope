package mapping

import (
	"testing"
)

func TestFieldMapper_BasicMapping(t *testing.T) {
	cfg := &FieldMappingConfig{
		ProviderName: "Amazon Web Services",
		FieldMappings: map[string]FieldMapping{
			"BillingAccountId": {
				SourceField: "bill/PayerAccountId",
				TargetField: "BillingAccountId",
				FieldType:   FieldTypeString,
				IsRequired:  true,
			},
			"BilledCost": {
				SourceField: "lineItem/BlendedCost",
				TargetField: "BilledCost",
				FieldType:   FieldTypeFloat64,
				IsRequired:  false,
			},
		},
		EnumMappings:    map[string]map[string]string{},
		DefaultValues:   map[string]interface{}{},
		ValidationRules: map[string]ValidationRule{},
		TimeFormats:     map[string]string{},
		CustomMappings:  map[string]CustomMappingFunction{},
	}

	mapper, err := NewFieldMapper(cfg)
	if err != nil {
		t.Fatalf("NewFieldMapper error: %v", err)
	}

	// Case 1: numeric source value should be handled (map[string]interface{} -> string -> float)
	rec, err := mapper.MapToFOCUS(map[string]interface{}{
		"bill/PayerAccountId":  "123456789012",
		"lineItem/BlendedCost": 10.50,
	})
	if err != nil {
		t.Fatalf("MapToFOCUS (float input) error: %v", err)
	}
	if rec.BillingAccountId != "123456789012" {
		t.Fatalf("BillingAccountId mismatch: %s", rec.BillingAccountId)
	}
	if rec.BilledCost == nil || *rec.BilledCost != 10.50 {
		if rec.BilledCost == nil {
			t.Fatalf("BilledCost is nil")
		}
		t.Fatalf("BilledCost mismatch: %v", *rec.BilledCost)
	}

	// Case 2: string source value
	rec, err = mapper.MapToFOCUS(map[string]interface{}{
		"bill/PayerAccountId":  "123456789012",
		"lineItem/BlendedCost": "10.50",
	})
	if err != nil {
		t.Fatalf("MapToFOCUS (string input) error: %v", err)
	}
	if rec.BillingAccountId != "123456789012" {
		t.Fatalf("BillingAccountId mismatch: %s", rec.BillingAccountId)
	}
	if rec.BilledCost == nil || *rec.BilledCost != 10.50 {
		if rec.BilledCost == nil {
			t.Fatalf("BilledCost is nil")
		}
		t.Fatalf("BilledCost mismatch: %v", *rec.BilledCost)
	}
}
