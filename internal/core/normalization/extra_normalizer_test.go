package normalization

import (
	"testing"
)

// New targeted tests to increase coverage of unit and region normalization edge cases.
func TestNormalizeRegion_EmptyAndUnknown(t *testing.T) {
	if out, ok := NormalizeRegion("aws", ""); ok || out != "" {
		t.Fatalf("expected empty input to return empty, false; got %q, %v", out, ok)
	}
	if out, ok := NormalizeRegion("aws", "definitely-unknown-region"); ok || out != "definitely-unknown-region" {
		t.Fatalf("expected unknown passthrough; got %q %v", out, ok)
	}
	// provider unspecified should search all dicts
	if out, ok := NormalizeRegion("", "us east 5"); !ok || out != "us-east5" {
		t.Fatalf("expected cross-provider lookup us-east5, got %q %v", out, ok)
	}
}

func TestNormalizeUnit_WhitespaceAndCase(t *testing.T) {
	cases := map[string]string{
		"  HRs  ":    "Hours",
		"GiB ":       "GB",
		"vcpu-hours": "vCPU-Hours",
	}
	for in, want := range cases {
		got, ok := NormalizeUnit(in)
		if !ok || got != want {
			t.Fatalf("NormalizeUnit(%q)=%q,%v want %q,true", in, got, ok, want)
		}
	}
}
