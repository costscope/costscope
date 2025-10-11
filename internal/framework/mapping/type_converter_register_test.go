package mapping

import (
	"testing"
	"time"
)

func TestTypeConverter_RegisterAndCustomTimeFormat(t *testing.T) {
	c := NewUniversalTypeConverter()
	c.RegisterStringTransform("trim_plus_lower", func(s string) string { return c.stringTransforms["lowercase"](c.stringTransforms["trim"](s)) })
	c.RegisterNumberTransform("neg_to_abs", func(f float64) float64 {
		if f < 0 {
			return -f
		}
		return f
	})
	customFormat := "2006-01-02|15"
	c.AddTimeFormat(customFormat)
	// Ensure custom time format usable
	ts := "2025-09-10|13"
	if _, err := c.ConvertTime(ts, ""); err != nil {
		t.Fatalf("expected custom time parse to succeed: %v", err)
	}
	// Get transforms lists (covers GetAvailableTransforms)
	ss, ns := c.GetAvailableTransforms()
	foundString := false
	for _, n := range ss {
		if n == "trim_plus_lower" {
			foundString = true
			break
		}
	}
	if !foundString {
		t.Fatalf("custom string transform missing")
	}
	foundNumber := false
	for _, n := range ns {
		if n == "neg_to_abs" {
			foundNumber = true
			break
		}
	}
	if !foundNumber {
		t.Fatalf("custom number transform missing")
	}
	// Empty time value error branch
	if _, err := c.ConvertTime("", ""); err == nil {
		t.Fatalf("expected empty time value error")
	}
	// Use registered transforms via ConvertString / ConvertFloat
	v, _ := c.ConvertString("  HELLO  ", FieldMapping{Transform: "trim_plus_lower"})
	if v.(string) != "hello" {
		t.Fatalf("string transform not applied: %v", v)
	}
	vf, _ := c.ConvertFloat(-3.1415, FieldMapping{Transform: "neg_to_abs"})
	if vf.(float64) != 3.1415 {
		t.Fatalf("number transform not applied: %v", vf)
	}
	_ = time.Now() // keep time import used
}
