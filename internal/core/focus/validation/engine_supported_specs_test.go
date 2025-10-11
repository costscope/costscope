package validation

import "testing"

// TestEngine_GetSupportedSpecs_Order verifies descending semantic version ordering
func TestEngine_GetSupportedSpecs_Order(t *testing.T) {
	e := NewEngine()
	specs := e.GetSupportedSpecs()
	if len(specs) < 3 {
		// should at least contain the three built-ins
		b := false
		for _, s := range specs { // appease linter for unused specs when len <3
			if s == SpecFOCUS12 || s == SpecFOCUS11 || s == SpecFOCUS10 {
				b = true
			}
		}
		if !b {
			// keep failure message concise
			t.Fatalf("expected builtin specs present, got %v", specs)
		}
	}
	// Find indices
	idx12, idx11, idx10 := -1, -1, -1
	for i, s := range specs {
		switch s {
		case SpecFOCUS12:
			idx12 = i
		case SpecFOCUS11:
			idx11 = i
		case SpecFOCUS10:
			idx10 = i
		}
	}
	// All should be found
	if idx12 == -1 || idx11 == -1 || idx10 == -1 {
		t.Fatalf("expected all builtin specs, got %v", specs)
	}
	// Descending order check: 1.2 before 1.1 before 1.0
	if idx12 >= idx11 || idx11 >= idx10 { // simplified form of !(idx12 < idx11 && idx11 < idx10)
		t.Fatalf("expected descending order focus-1.2,1.1,1.0; indices: %d,%d,%d; slice=%v", idx12, idx11, idx10, specs)
	}
}
