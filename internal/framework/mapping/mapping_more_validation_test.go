package mapping

import "testing"

func TestFieldMapper_AdditionalValidationPointerAndStringSuccess(t *testing.T) {
	maxLen := 10
	minVal := 1.0
	maxVal := 2.0
	cfg := &FieldMappingConfig{
		ProviderName: "aws",
		FieldMappings: map[string]FieldMapping{
			"BilledCost":        {SourceField: "billed", TargetField: "BilledCost", FieldType: FieldTypeFloat64, IsRequired: false},
			"ChargeDescription": {SourceField: "desc", TargetField: "ChargeDescription", FieldType: FieldTypeString, IsRequired: true},
		},
		ValidationRules: map[string]ValidationRule{
			"BilledCost":        {MinValue: &minVal, MaxValue: &maxVal},
			"ChargeDescription": {MaxLength: &maxLen},
		},
	}
	mapper, _ := NewFieldMapper(cfg)
	// failure path: BilledCost > max
	if _, err := mapper.MapToFOCUS(map[string]interface{}{"billed": 3.5, "desc": "OK"}); err == nil {
		t.Fatalf("expected max violation")
	}
	// success path: within range & string length ok
	rec, err := mapper.MapToFOCUS(map[string]interface{}{"billed": 1.5, "desc": "OK"})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if rec.BilledCost == nil || *rec.BilledCost != 1.5 {
		t.Fatalf("billed cost mismatch")
	}
	if rec.ChargeDescription != "OK" {
		t.Fatalf("charge desc mismatch")
	}
}
