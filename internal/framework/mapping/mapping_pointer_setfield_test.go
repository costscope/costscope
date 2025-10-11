package mapping

import (
	"strings"
	"testing"
	"time"
)

// customStructMapping returns a struct{} to force setFieldValue conversion failure to string field.
func customStructMapping(_ interface{}, _ FieldMapping, _ ValueExtractor) (interface{}, error) {
	return struct{}{}, nil
}

func TestFieldMapper_SetFieldValuePointerAndFailures(t *testing.T) {
	t.Run("setFieldValue_failure_custom_mapping", func(t *testing.T) {
		cfg := &FieldMappingConfig{
			ProviderName: "aws",
			FieldMappings: map[string]FieldMapping{
				"BillingCurrency": {SourceField: "cur", TargetField: "BillingCurrency", FieldType: FieldTypeString, IsRequired: true, Transform: "customStruct"},
			},
			EnumMappings:    map[string]map[string]string{},
			DefaultValues:   map[string]interface{}{},
			ValidationRules: map[string]ValidationRule{},
			TimeFormats:     map[string]string{},
			CustomMappings:  map[string]CustomMappingFunction{"customStruct": customStructMapping},
		}
		mapper, err := NewFieldMapper(cfg)
		if err != nil {
			t.Fatalf("mapper err: %v", err)
		}
		rec, err := mapper.MapToFOCUS(map[string]interface{}{"cur": "USD"})
		if err == nil || !strings.Contains(err.Error(), "cannot convert") {
			t.Fatalf("expected conversion error, got %v", err)
		}
		if rec != nil {
			t.Fatalf("expected nil record on setFieldValue failure")
		}
	})

	t.Run("pointer_assignments_and_validations", func(t *testing.T) {
		maxLen := 5
		minVal := 1.0
		maxVal := 10.0
		cfg := &FieldMappingConfig{
			ProviderName: "aws",
			FieldMappings: map[string]FieldMapping{
				// BilledCost is *float64 in FocusRecord; source provided as float64 -> pointer path (assignable to elem) exercised
				"BilledCost": {SourceField: "billed", TargetField: "BilledCost", FieldType: FieldTypeFloat64, IsRequired: false},
				// ConsumedQuantity is *float64; map as int64 to hit convertible-to-pointer branch
				"ConsumedQuantity": {SourceField: "consumed", TargetField: "ConsumedQuantity", FieldType: FieldTypeInt64, IsRequired: false},
				// Region optional pointer string; empty -> nil pointer path
				"Region": {SourceField: "region", TargetField: "Region", FieldType: FieldTypeOptional, IsRequired: false},
				// BillingPeriodEnd time RFC3339 default path (no explicit format)
				"BillingPeriodEnd": {SourceField: "end", TargetField: "BillingPeriodEnd", FieldType: FieldTypeTime, IsRequired: true},
				// ChargeDescription string with MaxLength rule to trigger success and failure
				"ChargeDescription": {SourceField: "desc", TargetField: "ChargeDescription", FieldType: FieldTypeString, IsRequired: true},
				// EffectiveCost numeric to exercise numeric validation pointers (later)
				"EffectiveCost": {SourceField: "eff", TargetField: "EffectiveCost", FieldType: FieldTypeFloat64, IsRequired: true},
			},
			EnumMappings:  map[string]map[string]string{},
			DefaultValues: map[string]interface{}{},
			ValidationRules: map[string]ValidationRule{
				"ChargeDescription": {MaxLength: &maxLen},
				"EffectiveCost":     {MinValue: &minVal, MaxValue: &maxVal},
			},
			TimeFormats:    map[string]string{},
			CustomMappings: map[string]CustomMappingFunction{},
		}
		mapper, err := NewFieldMapper(cfg)
		if err != nil {
			t.Fatalf("mapper err: %v", err)
		}

		// Failure: description exceeds max length, numeric min violation, empty region pointer becomes nil
		rec, err := mapper.MapToFOCUS(map[string]interface{}{
			"billed":   2.34,
			"consumed": "7", // string -> int64 extraction -> convertible to float64 pointer
			"region":   "",  // optional empty -> nil
			"end":      time.Now().UTC().Format(time.RFC3339),
			"desc":     "TOO-LONG", // > maxLen
			"eff":      0.5,        // below minVal
		})
		if err == nil || !strings.Contains(err.Error(), "validation failed") {
			t.Fatalf("expected validation failure, got %v", err)
		}
		if rec != nil {
			t.Fatalf("expected nil record on validation failure")
		}

		// Success case
		rec, err = mapper.MapToFOCUS(map[string]interface{}{
			"billed":   3.21,
			"consumed": "8",
			"region":   "", // remains nil pointer
			"end":      time.Now().UTC().Format(time.RFC3339),
			"desc":     "OK", // within max length
			"eff":      5.5,  // within [1,10]
		})
		if err != nil {
			t.Fatalf("unexpected success err: %v", err)
		}
		if rec.BilledCost == nil || *rec.BilledCost != 3.21 {
			t.Fatalf("BilledCost pointer mismatch %+v", rec.BilledCost)
		}
		if rec.ConsumedQuantity == nil || *rec.ConsumedQuantity != 8.0 {
			t.Fatalf("ConsumedQuantity pointer mismatch %+v", rec.ConsumedQuantity)
		}
		if rec.Region != nil {
			t.Fatalf("Region expected nil pointer, got %v", *rec.Region)
		}
		if rec.ChargeDescription != "OK" {
			t.Fatalf("ChargeDescription mismatch %q", rec.ChargeDescription)
		}
		if rec.EffectiveCost != 5.5 {
			t.Fatalf("EffectiveCost mismatch %v", rec.EffectiveCost)
		}
	})
}
