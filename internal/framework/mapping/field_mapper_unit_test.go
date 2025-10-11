package mapping

import (
	"testing"
)

type srcWithZone struct {
	SrcZone string
}

// Test that optional string fields are wrapped into pointer fields on the FocusRecord
func Test_FieldMapper_MapToFOCUS_optionalString_setsPointer(t *testing.T) {
	cfg := &FieldMappingConfig{
		ProviderName: "test",
		FieldMappings: map[string]FieldMapping{
			"AvailabilityZone": {SourceField: "SrcZone", TargetField: "AvailabilityZone", FieldType: FieldTypeOptional, IsRequired: false},
		},
	}

	fm, err := NewFieldMapper(cfg)
	if err != nil {
		t.Fatalf("failed to create FieldMapper: %v", err)
	}

	src := srcWithZone{SrcZone: "eu-west-1a"}
	rec, err := fm.MapToFOCUS(src)
	if err != nil {
		t.Fatalf("MapToFOCUS failed: %v", err)
	}

	if rec.AvailabilityZone == nil {
		t.Fatalf("expected AvailabilityZone to be non-nil pointer")
	}
	if *rec.AvailabilityZone != "eu-west-1a" {
		t.Fatalf("unexpected AvailabilityZone value: %v", *rec.AvailabilityZone)
	}
}

// Test that missing optional fields result in nil pointer on the FocusRecord
func Test_FieldMapper_MapToFOCUS_missingOptional_returnsNilPointer(t *testing.T) {
	cfg := &FieldMappingConfig{
		ProviderName: "test",
		FieldMappings: map[string]FieldMapping{
			"AvailabilityZone": {SourceField: "SrcZone", TargetField: "AvailabilityZone", FieldType: FieldTypeOptional, IsRequired: false},
		},
	}

	fm, err := NewFieldMapper(cfg)
	if err != nil {
		t.Fatalf("failed to create FieldMapper: %v", err)
	}

	// source missing the field (zero value empty string will be treated as missing by extractor)
	src := struct{ Other string }{Other: "x"}
	rec, err := fm.MapToFOCUS(src)
	if err != nil {
		t.Fatalf("MapToFOCUS failed: %v", err)
	}

	if rec.AvailabilityZone != nil {
		t.Fatalf("expected AvailabilityZone to be nil when source missing, got: %v", rec.AvailabilityZone)
	}
}
