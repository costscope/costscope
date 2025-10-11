package mapping

import (
	"strings"
	"testing"
)

// These additional tests focus on previously uncovered / low‑coverage branches:
// 1. NewFieldMapper nil config error.
// 2. Required field missing path in mapSingleField.
// 3. Type conversion mismatches (float/int/bool/time) and setFieldValue mismatch.
// 4. Time parse failure branch.
// 5. Validation failure branches for string length and numeric min.
// 6. Optional pointer wrapping (non‑empty -> pointer, empty -> nil).

func TestFieldMapper_ErrorPathsAndMissingRequired(t *testing.T) {
	if _, err := NewFieldMapper(nil); err == nil { // nil config error
		t.Fatalf("expected error for nil config")
	}

	cfg := &FieldMappingConfig{ // minimal config; only mappings we explicitly add
		ProviderName:    "aws",
		FieldMappings:   map[string]FieldMapping{},
		EnumMappings:    map[string]map[string]string{},
		DefaultValues:   map[string]interface{}{},
		ValidationRules: map[string]ValidationRule{},
		TimeFormats:     map[string]string{},
		CustomMappings:  map[string]CustomMappingFunction{},
	}

	// Required string field missing -> error
	cfg.FieldMappings["BillingAccountId"] = FieldMapping{
		SourceField: "bill/PayerAccountId", TargetField: "BillingAccountId", FieldType: FieldTypeString, IsRequired: true,
	}
	mapper, err := NewFieldMapper(cfg)
	if err != nil {
		t.Fatalf("new mapper err: %v", err)
	}
	if _, err = mapper.MapToFOCUS(map[string]interface{}{}); err == nil || !strings.Contains(err.Error(), "required field bill/PayerAccountId is missing") {
		t.Fatalf("expected missing required field error, got %v", err)
	}

	// Float mapping with non‑numeric value (bool) -> parse error propagated because required
	cfg2 := *cfg
	cfg2.FieldMappings = map[string]FieldMapping{
		"EffectiveCost": {SourceField: "lineItem/UnblendedCost", TargetField: "EffectiveCost", FieldType: FieldTypeFloat64, IsRequired: true},
	}
	mapper2, _ := NewFieldMapper(&cfg2)
	if _, err := mapper2.MapToFOCUS(map[string]interface{}{"lineItem/UnblendedCost": true}); err == nil || !strings.Contains(err.Error(), "failed to extract value") {
		t.Fatalf("expected extract/parse error for float bool, got %v", err)
	}

	// Int mapping with non‑int value (string non‑int) -> parse error
	cfg3 := *cfg
	cfg3.FieldMappings = map[string]FieldMapping{
		"UsageQuantity": {SourceField: "lineItem/UsageAmount", TargetField: "UsageQuantity", FieldType: FieldTypeInt64, IsRequired: true},
	}
	mapper3, _ := NewFieldMapper(&cfg3)
	if _, err := mapper3.MapToFOCUS(map[string]interface{}{"lineItem/UsageAmount": "not-an-int"}); err == nil || !strings.Contains(err.Error(), "failed to extract value") {
		t.Fatalf("expected int parse error, got %v", err)
	}

	// Bool mapping with invalid value -> parse error
	cfg4 := *cfg
	cfg4.FieldMappings = map[string]FieldMapping{
		"ProviderName": {SourceField: "providerEnabled", TargetField: "ProviderName", FieldType: FieldTypeBool, IsRequired: true},
	}
	mapper4, _ := NewFieldMapper(&cfg4)
	if _, err := mapper4.MapToFOCUS(map[string]interface{}{"providerEnabled": "not-bool"}); err == nil || !strings.Contains(err.Error(), "cannot parse bool") {
		t.Fatalf("expected bool parse error, got %v", err)
	}
}

func TestFieldMapper_TimeParseAndValidationFailures(t *testing.T) {
	// Time parse failure (required field)
	cfg := &FieldMappingConfig{
		ProviderName: "aws",
		FieldMappings: map[string]FieldMapping{
			"BillingPeriodStart": {SourceField: "billing_start", TargetField: "BillingPeriodStart", FieldType: FieldTypeTime, IsRequired: true},
		},
		EnumMappings:    map[string]map[string]string{},
		DefaultValues:   map[string]interface{}{},
		ValidationRules: map[string]ValidationRule{},
		TimeFormats:     map[string]string{},
		CustomMappings:  map[string]CustomMappingFunction{},
	}
	mapper, _ := NewFieldMapper(cfg)
	if _, err := mapper.MapToFOCUS(map[string]interface{}{"billing_start": "not-a-time"}); err == nil || !strings.Contains(err.Error(), "failed to extract value") && !strings.Contains(err.Error(), "failed to map required field BillingPeriodStart") {
		t.Fatalf("expected time parse error, got %v", err)
	}

	// Validation failures (string length + numeric min)
	minLen := 5
	minVal := 10.0
	cfg2 := &FieldMappingConfig{
		ProviderName: "aws",
		FieldMappings: map[string]FieldMapping{
			"BillingAccountId": {SourceField: "acct", TargetField: "BillingAccountId", FieldType: FieldTypeString, IsRequired: true},
			"EffectiveCost":    {SourceField: "cost", TargetField: "EffectiveCost", FieldType: FieldTypeFloat64, IsRequired: true},
		},
		EnumMappings:  map[string]map[string]string{},
		DefaultValues: map[string]interface{}{},
		ValidationRules: map[string]ValidationRule{
			"BillingAccountId": {MinLength: &minLen},
			"EffectiveCost":    {MinValue: &minVal},
		},
		TimeFormats:    map[string]string{},
		CustomMappings: map[string]CustomMappingFunction{},
	}
	mapper2, _ := NewFieldMapper(cfg2)
	if _, err := mapper2.MapToFOCUS(map[string]interface{}{"acct": "abc", "cost": 1.0}); err == nil || !strings.Contains(err.Error(), "validation failed") {
		t.Fatalf("expected validation failure, got %v", err)
	}
}

func TestFieldMapper_OptionalWrapping(t *testing.T) {
	cfg := &FieldMappingConfig{
		ProviderName: "aws",
		FieldMappings: map[string]FieldMapping{
			"AvailabilityZone": {SourceField: "az", TargetField: "AvailabilityZone", FieldType: FieldTypeOptional, IsRequired: false},
		},
		EnumMappings:    map[string]map[string]string{},
		DefaultValues:   map[string]interface{}{},
		ValidationRules: map[string]ValidationRule{},
		TimeFormats:     map[string]string{},
		CustomMappings:  map[string]CustomMappingFunction{},
	}
	mapper, _ := NewFieldMapper(cfg)
	rec, err := mapper.MapToFOCUS(map[string]interface{}{"az": "us-east-1a"})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if rec.AvailabilityZone == nil || *rec.AvailabilityZone != "us-east-1a" {
		t.Fatalf("optional value not wrapped correctly: %+v", rec.AvailabilityZone)
	}

	// Empty optional value becomes nil
	rec2, err := mapper.MapToFOCUS(map[string]interface{}{"az": ""})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if rec2.AvailabilityZone != nil {
		t.Fatalf("expected nil optional for empty string, got %v", *rec2.AvailabilityZone)
	}
}
