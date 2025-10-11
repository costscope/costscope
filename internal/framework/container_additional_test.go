package framework

import (
	"testing"
)

// Test that RegisterSingletonFactory lazily initializes exactly once
func TestContainerSingletonFactoryLazyInit(t *testing.T) {
	c := NewContainer()
	calls := 0
	c.RegisterSingletonFactory("svc", func(c *Container) interface{} {
		calls++
		return &struct{ N int }{N: 42}
	})

	// Before Get, factory should not be called
	if calls != 0 {
		t.Fatalf("factory called eagerly: %d", calls)
	}

	// First Get should call factory
	v1, err := c.Get("svc")
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected 1 factory call, got %d", calls)
	}

	// Second Get returns same instance without calling factory again
	v2, err := c.Get("svc")
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if calls != 1 {
		t.Fatalf("factory called more than once: %d", calls)
	}
	if v1 != v2 {
		t.Fatalf("expected same singleton instance")
	}
}

// Test Inject error surfaces when service missing or type mismatch
type Dep interface{ isDep() }
type depImpl struct{}

func (depImpl) isDep() {}

type Target struct {
	D Dep `inject:"dep"`
}

func TestContainerInjectErrors(t *testing.T) {
	c := NewContainer()

	// Missing service
	var tgt Target
	if err := c.Inject(&tgt); err == nil {
		t.Fatalf("expected error for missing service")
	}

	// Register wrong type and expect mismatch error
	c.RegisterSingleton("dep", &struct{}{})
	if err := c.Inject(&tgt); err == nil {
		t.Fatalf("expected type mismatch error")
	}

	// Register correct type and expect success
	c.Unregister("dep")
	c.RegisterSingleton("dep", depImpl{})
	if err := c.Inject(&tgt); err != nil {
		t.Fatalf("inject failed: %v", err)
	}
	if tgt.D == nil {
		t.Fatalf("dependency not injected")
	}
}
