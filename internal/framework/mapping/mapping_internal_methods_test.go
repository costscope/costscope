package mapping

import (
	ftypes "local/costscope/internal/core/focus/types"
	"testing"
	"time"
)

// Helper to build a mapper quickly
func newTestMapper() *FieldMapper {
	m, _ := NewFieldMapper(&FieldMappingConfig{ProviderName: "aws", FieldMappings: map[string]FieldMapping{}})
	return m
}

func TestWrapOptionalValue_AllBranches(t *testing.T) {
	fm := newTestMapper()
	cases := []struct {
		name    string
		in      interface{}
		wantNil bool
		wantErr bool
	}{
		{"nil", nil, true, false},
		{"empty string", "", true, false},
		{"non-empty string", "us-east-1a", false, false},
		{"float64", float64(1.23), false, false},
		{"int64", int64(5), false, false},
		{"bool true", true, false, false},
		{"bool false", false, false, false},
		{"time zero", time.Time{}, true, false},
		{"time nonzero", time.Now(), false, false},
		{"unsupported struct", struct{}{}, true, true},
	}
	for _, tc := range cases {
		v, err := fm.wrapOptionalValue(tc.in)
		if tc.wantErr && err == nil {
			t.Fatalf("%s expected error", tc.name)
		}
		if !tc.wantErr && err != nil {
			t.Fatalf("%s unexpected err: %v", tc.name, err)
		}
		if tc.wantNil && v != nil {
			t.Fatalf("%s expected nil got %v", tc.name, v)
		}
		if !tc.wantNil && v == nil {
			t.Fatalf("%s expected non-nil", tc.name)
		}
	}
}

func TestValidateFieldValue_PointerVariants(t *testing.T) {
	fm := newTestMapper()
	// Pointer string allowed values
	s := "GOOD"
	rule := ValidationRule{AllowedValues: []string{"OK", "GOOD"}}
	if err := fm.validateFieldValue(&s, rule); err != nil {
		t.Fatalf("pointer string allowed failed: %v", err)
	}
	s2 := "BAD"
	if err := fm.validateFieldValue(&s2, rule); err == nil {
		t.Fatalf("expected disallowed value error")
	}

	// Pointer float numeric min/max
	f := 5.0
	min := 1.0
	max := 10.0
	nrule := ValidationRule{MinValue: &min, MaxValue: &max}
	if err := fm.validateFieldValue(&f, nrule); err != nil {
		t.Fatalf("pointer float validation failed: %v", err)
	}
	f2 := 0.5
	if err := fm.validateFieldValue(&f2, nrule); err == nil {
		t.Fatalf("expected min violation")
	}
}

func TestConvertEnum_NilMapReturnsRaw(t *testing.T) {
	// Config without enum mapping entry
	cfg := &FieldMappingConfig{ProviderName: "aws", FieldMappings: map[string]FieldMapping{
		"ChargeClass": {SourceField: "cls", TargetField: "ChargeClass", FieldType: FieldTypeEnum, IsRequired: true, EnumMapping: "nope"},
	}, EnumMappings: map[string]map[string]string{}}
	mapper, _ := NewFieldMapper(cfg)
	rec, err := mapper.MapToFOCUS(map[string]interface{}{"cls": "On-Demand"})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if rec.ChargeClass != "On-Demand" {
		t.Fatalf("expected raw enum returned, got %q", rec.ChargeClass)
	}
}

func TestTimeConversion_AlternateFormat(t *testing.T) {
	cfg := &FieldMappingConfig{ProviderName: "aws", FieldMappings: map[string]FieldMapping{
		"BillingPeriodEnd": {SourceField: "end", TargetField: "BillingPeriodEnd", FieldType: FieldTypeTime, IsRequired: true}, // no explicit format
	}}
	mapper, _ := NewFieldMapper(cfg)
	input := time.Now().UTC().Format("2006/01/02 15:04:05") // one of utc.timeFormats list
	rec, err := mapper.MapToFOCUS(map[string]interface{}{"end": input})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if rec.BillingPeriodEnd.IsZero() {
		t.Fatalf("expected parsed time")
	}
}

func TestSetFieldValue_ConvertibleAndPointerBranches(t *testing.T) {
	fm := newTestMapper()
	realRec := &ftypes.FocusRecord{}
	// 1. int64 -> float64 field EffectiveCost (convertible)
	if err := fm.setFieldValue(realRec, "EffectiveCost", int64(7)); err != nil {
		t.Fatalf("convert int64->float64 err: %v", err)
	}
	if realRec.EffectiveCost != 7 {
		t.Fatalf("expected 7 got %v", realRec.EffectiveCost)
	}
	// 2. int64 -> *float64 (pointer element convertible) BilledCost
	if err := fm.setFieldValue(realRec, "BilledCost", int64(3)); err != nil {
		t.Fatalf("convert int64->*float64 err: %v", err)
	}
	if realRec.BilledCost == nil || *realRec.BilledCost != 3 {
		t.Fatalf("BilledCost mismatch %+v", realRec.BilledCost)
	}
}
