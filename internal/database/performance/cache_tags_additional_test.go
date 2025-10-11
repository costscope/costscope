package performance

import (
	"testing"
	"time"
)

// Additional edge-case tests for tag deletion semantics.
func TestEnhancedCache_DeleteByTags_EdgeCases(t *testing.T) {
	cfg := &CacheConfig{MaxSize: 10, DefaultTTL: defaultTestTTL, CleanupInterval: defaultTestCleanup, Monitor: &CacheMonitorConfig{Enabled: false}}
	ec := NewEnhancedCache(cfg)
	defer func() { _ = ec.Close() }()

	// Seed entries
	_ = ec.SetWithTags("k_hot", "v1", []string{"hot"})
	_ = ec.SetWithTags("k_cold", "v2", []string{"cold"})
	_ = ec.SetWithTags("k_multi", "v3", []string{"hot", "region:us"})

	// (a) Empty tag slice should delete nothing (explicit no-op)
	if del := ec.DeleteByTags([]string{}); del != 0 {
		t.Fatalf("expected 0 deletions for empty tag slice, got %d", del)
	}

	// (b) Non-matching tag
	if del := ec.DeleteByTags([]string{"nope"}); del != 0 {
		// should not delete anything
		t.Fatalf("expected 0 deletions for non-matching tag, got %d", del)
	}
	// All keys should still exist
	for _, k := range []string{"k_hot", "k_cold", "k_multi"} {
		if _, ok := ec.Get(k); !ok {
			// they may have aged; none should be expired yet (default TTL 1h). If this fails it's a logic issue.
			t.Fatalf("expected key %s to remain after non-matching delete", k)
		}
	}

	// (c) Partial intersection: delete by one of multiple tags
	if del := ec.DeleteByTags([]string{"hot"}); del != 2 { // k_hot & k_multi
		t.Fatalf("expected to delete 2 hot-tagged entries, got %d", del)
	}
	if _, ok := ec.Get("k_cold"); !ok {
		// unaffected entry should remain
		t.Fatal("expected k_cold to remain after hot deletion")
	}
}

// Provide defaults used in tests to avoid magic numbers
const (
	defaultTestTTL     = time.Hour
	defaultTestCleanup = time.Minute
)
