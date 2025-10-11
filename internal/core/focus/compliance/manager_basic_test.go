package compliance

import "testing"

// Basic tests to exercise Manager initialization & framework/rule retrieval
func TestManager_DefaultFrameworksLoaded(t *testing.T) {
	m := NewManager()
	if len(m.frameworks) == 0 {
		t.Fatalf("expected default frameworks loaded")
	}
	if _, ok := m.frameworks["focus"]; !ok {
		t.Fatalf("expected focus framework present")
	}
	if len(m.rules) == 0 {
		t.Fatalf("expected default rules loaded")
	}
}
