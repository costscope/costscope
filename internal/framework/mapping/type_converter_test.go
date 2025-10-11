package mapping

import (
	"testing"
	"time"
)

func TestUniversalTypeConverter_StringTransforms(t *testing.T) {
	utc := NewUniversalTypeConverter()
	out, err := utc.ConvertString("  HeLLo  ", FieldMapping{Transform: "lowercase"})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if out.(string) != "hello" {
		t.Fatalf("expected lowercase 'hello', got %q", out)
	}
	out, _ = utc.ConvertString("a   b\t c\n", FieldMapping{Transform: "normalize_whitespace"})
	if out.(string) != "a b c" {
		t.Fatalf("normalize_whitespace failed: %q", out)
	}
}

func TestUniversalTypeConverter_NumberTransforms(t *testing.T) {
	utc := NewUniversalTypeConverter()
	out, err := utc.ConvertFloat(-1.236, FieldMapping{Transform: "round_to_cents"})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if out.(float64) != -1.24 { // bankers not needed; simple round
		t.Fatalf("round_to_cents failed: %v", out)
	}
	out, _ = utc.ConvertFloat(-5.0, FieldMapping{Transform: "absolute"})
	if out.(float64) != 5.0 {
		t.Fatalf("absolute failed: %v", out)
	}
}

func TestUniversalTypeConverter_TimeFormats(t *testing.T) {
	utc := NewUniversalTypeConverter()
	// RFC3339
	ts, err := utc.ConvertTime("2024-01-02T03:04:05Z", "")
	if err != nil || ts.UTC().Format(time.RFC3339) != "2024-01-02T03:04:05Z" {
		t.Fatalf("parse RFC3339 failed: %v %v", ts, err)
	}
	// Custom listed format
	ts, err = utc.ConvertTime("01/02/2024 03:04:05", "")
	if err != nil || ts.UTC().Format("01/02/2006 15:04:05") != "01/02/2024 03:04:05" {
		t.Fatalf("parse custom failed: ts=%v err=%v", ts, err)
	}
	// Specific format argument has priority
	ts, err = utc.ConvertTime("02-01-2024 03:04:05", "02-01-2006 15:04:05")
	if err != nil || ts.Year() != 2024 || ts.Month() != time.January || ts.Day() != 2 {
		t.Fatalf("specific format failed: %v %v", ts, err)
	}
}

func TestUniversalTypeConverter_EnumCaseInsensitive(t *testing.T) {
	utc := NewUniversalTypeConverter()
	mapped, err := utc.ConvertEnum("Discount", map[string]string{"discount": "Discount", "CREDIT": "Credit"})
	if err != nil {
		t.Fatalf("enum err: %v", err)
	}
	if mapped != "Discount" {
		t.Fatalf("expected Discount, got %q", mapped)
	}
	mapped, _ = utc.ConvertEnum("CREDIT", map[string]string{"discount": "Discount", "CREDIT": "Credit"})
	if mapped != "Credit" {
		t.Fatalf("expected Credit, got %q", mapped)
	}
}
