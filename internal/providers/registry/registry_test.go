package registry

import "testing"

type testProvider struct{ id string }

func (p *testProvider) Name() string { return p.id }

func TestRegistry_RegisterAndGet(t *testing.T) {
	Register("example", func(opts ...interface{}) (Provider, error) { return &testProvider{id: "example"}, nil })
	p, err := Get("example")
	if err != nil {
		t.Fatalf("expected provider, got err: %v", err)
	}
	if p.Name() != "example" {
		t.Fatalf("unexpected name: %s", p.Name())
	}
}

func TestRegistry_GetUnknown(t *testing.T) {
	_, err := Get("does-not-exist")
	if err == nil {
		t.Fatalf("expected error for unknown provider")
	}
}

// TestRegistry_List ensures List returns registered provider names so the symbol
// is considered reachable (deadcode tool previously flagged it). The List API is
// used (or will be) by higher-level CLI / API endpoints that enumerate providers.
func TestRegistry_List(t *testing.T) {
	// Use a unique name to avoid collision with other tests.
	name := "example-list"
	Register(name, func(opts ...interface{}) (Provider, error) { return &testProvider{id: name}, nil })
	got := List()
	found := false
	for _, n := range got {
		if n == name {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected provider %s in List() result: %v", name, got)
	}
}
