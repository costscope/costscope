package api

import "testing"

// TestGateEnabled_Variants verifies gateEnabled behaviour for unset/falsy/truthy env values.
func TestGateEnabled_Variants(t *testing.T) {
	// Unset/empty => enabled
	t.Setenv("MY_TEST_GATE", "")
	if !gateEnabled("MY_TEST_GATE") {
		t.Fatalf("expected gateEnabled to be true when env is empty or unset")
	}

	falsy := []string{"0", "false", "off", "disable", "disabled"}
	for _, v := range falsy {
		t.Setenv("MY_TEST_GATE", v)
		if gateEnabled("MY_TEST_GATE") {
			t.Fatalf("expected gateEnabled=false for value '%s'", v)
		}
	}

	truthy := []string{"1", "true", "on", "yes", "enabled"}
	for _, v := range truthy {
		t.Setenv("MY_TEST_GATE", v)
		if !gateEnabled("MY_TEST_GATE") {
			t.Fatalf("expected gateEnabled=true for value '%s'", v)
		}
	}
}
