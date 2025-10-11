package types

import "testing"

// Minimal smoke test to exercise basic type instantiation for coverage
func TestStructInstantiation(t *testing.T) {
	pi := ProviderInfo{Name: "aws-main", Type: ProviderTypeAWS, Version: "v1", SupportedRegions: []string{"us-east-1"}, Capabilities: []string{"cost"}}
	if pi.Type != ProviderTypeAWS || len(pi.SupportedRegions) != 1 {
		t.Fatalf("unexpected provider info: %+v", pi)
	}
}
