package normalization

import "testing"

func TestNormalizeRegionProviderSpecific(t *testing.T) {
	cases := []struct{ provider, in, want string }{
		{"aws", "US East (N. Virginia)", "us-east-1"},
		{"aws", "us east 1", "us-east-1"},
		{"aws", "us-east1", "us-east-1"},
		{"aws", "US Gov West 1", "us-gov-west-1"},
		{"aws", "cn north 1", "cn-north-1"},
		{"aws", "eu central 1", "eu-central-1"},
		{"azure", "West Europe", "westeurope"},
		{"azure", "WEST EUROPE", "westeurope"},
		{"azure", "US Gov Virginia", "usgovvirginia"},
		{"azure", "China East", "chinaeast"},
		{"gcp", "US Central (Iowa)", "us-central1"},
		{"gcp", "europe west 1", "europe-west1"},
		{"gcp", "us east 5", "us-east5"},
		{"gcp", "asia south 2", "asia-south2"},
	}
	for _, c := range cases {
		got, ok := NormalizeRegion(c.provider, c.in)
		if !ok || got != c.want {
			t.Fatalf("NormalizeRegion(%q,%q)=%q,%v want %q,true", c.provider, c.in, got, ok, c.want)
		}
	}
}

func TestNormalizeRegionHeuristic(t *testing.T) {
	if got, ok := NormalizeRegion("aws", "eu-west-1"); !ok || got != "eu-west-1" {
		t.Fatalf("expected heuristic canonical pass-through eu-west-1, got %q ok=%v", got, ok)
	}
}

func TestNormalizeUnit(t *testing.T) {
	cases := map[string]string{
		"hour":       "Hours",
		"HRS":        "Hours",
		"gib":        "GB",
		"GB-Hours":   "GB-Hours",
		"vcpu hours": "vCPU-Hours",
		"Requests":   "Requests",
	}
	for in, want := range cases {
		got, ok := NormalizeUnit(in)
		if !ok || got != want {
			t.Fatalf("NormalizeUnit(%q)=%q,%v want %q,true", in, got, ok, want)
		}
	}
}
