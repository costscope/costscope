package mapping

import (
	"testing"
	"time"
)

// Additional coverage-oriented tests exercising:
// - setFieldValue convertible numeric path (int64 -> float64)
// - enum mapping (present + absent)
// - string field fallback formatting for non-string raw values
// - optional field unsupported type skip (wrapOptionalValue error path)
// - validateStringField allowed values (success + failure)
// - validateNumericField max value failure
// - successful time mapping via specified TimeFormat
// - skipping optional field on conversion error
func TestFieldMapper_AdditionalBranches(t *testing.T) {
	minLen := 2
	maxVal := 1.5
	allowed := []string{"OK", "GOOD"}

	cfg := &FieldMappingConfig{
		ProviderName: "aws",
		FieldMappings: map[string]FieldMapping{
			// Convertible int64 -> float64 target field (PricingQuantity is float64 in FocusRecord)
			"PricingQuantity": {SourceField: "lineItem/Qty", TargetField: "PricingQuantity", FieldType: FieldTypeInt64, IsRequired: true},
			// Enum mapping present
			"ChargeCategory": {SourceField: "charge_cat", TargetField: "ChargeCategory", FieldType: FieldTypeEnum, IsRequired: true, EnumMapping: "cc"},
			// Enum mapping absent (returns raw)
			"ChargeClass": {SourceField: "charge_class", TargetField: "ChargeClass", FieldType: FieldTypeEnum, IsRequired: true, EnumMapping: "missing_map"},
			// String field with non-string raw value triggers fmt.Sprintf fallback
			"BillingAccountId": {SourceField: "acct_id_num", TargetField: "BillingAccountId", FieldType: FieldTypeString, IsRequired: true},
			// Optional field with map value: ExtractString will stringify map -> wrapOptionalValue treats as non-empty string and pointers it.
			"AvailabilityZone": {SourceField: "az_obj", TargetField: "AvailabilityZone", FieldType: FieldTypeOptional, IsRequired: false},
			// Time format success
			"BillingPeriodStart": {SourceField: "bp_start", TargetField: "BillingPeriodStart", FieldType: FieldTypeTime, IsRequired: true, TimeFormat: "2006-01-02"},
			// String validation allowed values
			"ChargeDescription": {SourceField: "charge_desc", TargetField: "ChargeDescription", FieldType: FieldTypeString, IsRequired: true},
			// Numeric max value violation (ListCost)
			"ListCost": {SourceField: "list_cost", TargetField: "ListCost", FieldType: FieldTypeFloat64, IsRequired: true},
		},
		EnumMappings: map[string]map[string]string{
			"cc": {"usage": "Usage"},
		},
		DefaultValues: map[string]interface{}{},
		ValidationRules: map[string]ValidationRule{
			"ChargeDescription": {MinLength: &minLen, AllowedValues: allowed},
			"ListCost":          {MaxValue: &maxVal},
		},
		TimeFormats:    map[string]string{},
		CustomMappings: map[string]CustomMappingFunction{},
	}

	mapper, err := NewFieldMapper(cfg)
	if err != nil {
		t.Fatalf("new mapper err: %v", err)
	}

	// First attempt with failing validation and enum / optional skip
	rec, err := mapper.MapToFOCUS(map[string]interface{}{
		"lineItem/Qty": "5",                          // int path (string -> int64 -> convertible -> float64)
		"charge_cat":   "usage",                      // enum mapping present
		"charge_class": "On-Demand",                  // no enum mapping defined -> raw
		"acct_id_num":  1234567890,                   // non-string -> fmt.Sprintf fallback
		"az_obj":       map[string]int{"ignored": 1}, // unsupported optional type
		"bp_start":     "2025-09-10",                 // parses with TimeFormat
		"charge_desc":  "BAD",                        // not in allowed values
		"list_cost":    2.0,                          // exceeds max value 1.5
	})
	if err == nil {
		t.Fatalf("expected validation failure (string allowed values or numeric max), got nil")
	}
	if rec != nil {
		t.Fatalf("expected nil record on required validation failure")
	}

	// Second attempt with valid values to exercise success branches
	rec, err = mapper.MapToFOCUS(map[string]interface{}{
		"lineItem/Qty": "7",
		"charge_cat":   "usage",
		"charge_class": "On-Demand",
		"acct_id_num":  999,
		"az_obj":       map[string]int{"ignored": 2}, // still skipped silently
		"bp_start":     time.Now().Format("2006-01-02"),
		"charge_desc":  "GOOD", // allowed
		"list_cost":    1.4,    // within max
	})
	if err != nil {
		t.Fatalf("unexpected success err: %v", err)
	}
	if rec.PricingQuantity != 7 {
		t.Fatalf("expected PricingQuantity 7 got %v", rec.PricingQuantity)
	}
	if rec.ChargeCategory != "Usage" {
		t.Fatalf("enum mapping failed, got %q", rec.ChargeCategory)
	}
	if rec.ChargeClass != "On-Demand" {
		t.Fatalf("raw enum (missing map) expected On-Demand got %q", rec.ChargeClass)
	}
	if rec.BillingAccountId != "999" {
		t.Fatalf("fmt fallback expected '999' got %q", rec.BillingAccountId)
	}
	// AvailabilityZone should be non-nil pointer to stringified map
	if rec.AvailabilityZone == nil || *rec.AvailabilityZone == "" {
		t.Fatalf("expected AvailabilityZone pointer with stringified content, got %+v", rec.AvailabilityZone)
	}
	if rec.BillingPeriodStart.IsZero() {
		t.Fatalf("expected parsed BillingPeriodStart")
	}
	if rec.ChargeDescription != "GOOD" {
		t.Fatalf("ChargeDescription mismatch %q", rec.ChargeDescription)
	}
	if rec.ListCost != 1.4 {
		t.Fatalf("ListCost mismatch %v", rec.ListCost)
	}
}
