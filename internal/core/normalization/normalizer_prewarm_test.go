package normalization

import "testing"

func TestPreWarmAndCacheStats(t *testing.T) {
	PreWarm()
	h1, m1, _, sz1 := RegionCacheStats()
	if sz1 == 0 {
		t.Fatalf("expected region cache prewarmed")
	}
	// trigger a known hit
	if v, ok := NormalizeRegion("aws", "us east 1"); !ok || v != "us-east-1" {
		t.Fatalf("normalize mismatch: %v %v", v, ok)
	}
	h2, m2, _, _ := RegionCacheStats()
	if h2 < h1 {
		t.Fatalf("expected hits to increase")
	}
	if m2 < m1 {
		t.Fatalf("misses should be monotonic non-decreasing")
	}
}
