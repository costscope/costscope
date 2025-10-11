package cache

import "testing"

func TestLRUEvictionAndStats(t *testing.T) {
	l := NewLRU(3)
	l.Add("a", "1")
	l.Add("b", "2")
	l.Add("c", "3")
	if _, ok := l.Get("a"); !ok { // access a so it becomes MRU
		t.Fatalf("expected hit for a")
	}
	if _, ok := l.Get("missing"); ok { // should be miss
		t.Fatalf("unexpected hit for missing key")
	}
	l.Add("d", "4")              // should evict LRU among b/c (b if a promoted)
	if _, ok := l.Get("b"); ok { // b expected evicted
		t.Fatalf("expected b to be evicted")
	}
	if v, ok := l.Get("a"); !ok || v != "1" { // still present
		t.Fatalf("expected a present after eviction, got %q present=%v", v, ok)
	}
	hits, misses, evicts, sz := l.Stats()
	if hits == 0 || misses == 0 || evicts == 0 {
		t.Fatalf("expected non-zero stats got hits=%d misses=%d evicts=%d", hits, misses, evicts)
	}
	if sz != 3 { // capacity maintained
		t.Fatalf("expected size 3 got %d", sz)
	}
}

func TestLRUPromotion(t *testing.T) {
	l := NewLRU(2)
	l.Add("x", "1")
	l.Add("y", "2")
	// Promote x by Get so y becomes LRU, then add z and ensure y evicted not x.
	if _, ok := l.Get("x"); !ok {
		t.Fatalf("expected hit for x")
	}
	l.Add("z", "3")
	if _, ok := l.Get("y"); ok {
		t.Fatalf("expected y eviction after adding z")
	}
	if v, ok := l.Get("x"); !ok || v != "1" {
		t.Fatalf("expected x retained, got %q present=%v", v, ok)
	}
}
