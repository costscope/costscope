package normalization

import (
	"testing"
	"time"
)

// TestNormalizeUnitVariants covers additional unit variants & cache hit path
func TestNormalizeUnitVariants(t *testing.T) {
	PreWarm()
	// Known variants
	cases := []struct{ in, want string }{
		{"gib", "GB"},
		{"GB-Hours", "GB-Hours"}, // canonical should return as-is
		{"vcpu-hours", "vCPU-Hours"},
		{"requests", "Requests"},
	}
	for _, c := range cases {
		if got, ok := NormalizeUnit(c.in); !ok || got != c.want {
			t.Fatalf("NormalizeUnit(%s) = %s,%v want %s,true", c.in, got, ok, c.want)
		}
	}
	// Cache hit: call again and ensure hits increase (non-zero)
	NormalizeUnit("gib")
	if h, _, _, _ := UnitCacheStats(); h == 0 {
		t.Fatalf("expected cache hit for second gib normalization")
	}
}

// TestStartCacheMetricsRefresher exercises background goroutine (no assertions on metrics values)
func TestStartCacheMetricsRefresher(t *testing.T) {
	stop := make(chan struct{})
	StartCacheMetricsRefresher(10*time.Millisecond, stop)
	time.Sleep(25 * time.Millisecond)
	close(stop)
	// interval<=0 no-op
	StartCacheMetricsRefresher(0, make(chan struct{}))
}
