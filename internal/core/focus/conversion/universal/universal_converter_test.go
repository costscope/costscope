package universal

import "testing"

// Minimal smoke test invoking NewUniversalConverter (ensures constructor stable)
func TestNewUniversalConverter_Smoke(t *testing.T) {
	c := NewUniversalConverter()
	if c == nil {
		t.Fatalf("expected converter instance")
	}
}
