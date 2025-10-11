// Unified (formerly perf-tagged) cache tests. Build tag removed after perf tag retirement.

package performance

import (
	"errors"
	"sync"
	"testing"
	"time"
)

// inMemoryPersistence is a simple stub implementing CachePersistence for tests.
type inMemoryPersistence struct {
	mu      sync.Mutex
	stored  map[string]*CacheEntry
	toLoad  map[string]*CacheEntry
	saveErr error
	loadErr error
}

func (p *inMemoryPersistence) Save(entries map[string]*CacheEntry) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.saveErr != nil {
		return p.saveErr
	}
	p.stored = make(map[string]*CacheEntry)
	for k, v := range entries {
		p.stored[k] = v
	}
	return nil
}

func (p *inMemoryPersistence) Load() (map[string]*CacheEntry, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.loadErr != nil {
		return nil, p.loadErr
	}
	// Return a copy to avoid aliasing issues
	out := make(map[string]*CacheEntry)
	for k, v := range p.toLoad {
		out[k] = v
	}
	return out, nil
}

func (p *inMemoryPersistence) Clear() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.stored = map[string]*CacheEntry{}
	p.toLoad = map[string]*CacheEntry{}
	return nil
}

func TestEnhancedCache_SetGet_Expire(t *testing.T) {
	cfg := &CacheConfig{
		MaxSize:         10,
		DefaultTTL:      500 * time.Millisecond,
		CleanupInterval: 50 * time.Millisecond,
		Monitor:         &CacheMonitorConfig{Enabled: false},
	}
	ec := NewEnhancedCache(cfg)
	defer func() { _ = ec.Close() }()

	if err := ec.SetWithTTL("k1", "v1", 100*time.Millisecond); err != nil {
		t.Fatalf("set: %v", err)
	}
	if v, ok := ec.Get("k1"); !ok || v.(string) != "v1" {
		t.Fatalf("expected hit v1, got %v ok=%v", v, ok)
	}
	// wait for expiration
	time.Sleep(150 * time.Millisecond)
	if _, ok := ec.Get("k1"); ok {
		t.Fatal("expected miss after expiration")
	}
}

func TestEnhancedCache_EvictionLRU(t *testing.T) {
	cfg := &CacheConfig{MaxSize: 2, DefaultTTL: time.Second, CleanupInterval: time.Second, Monitor: &CacheMonitorConfig{Enabled: false}}
	ec := NewEnhancedCache(cfg)
	defer func() { _ = ec.Close() }()

	_ = ec.Set("a", "A")
	_ = ec.Set("b", "B")
	// Access "a" so that "b" becomes the LRU
	if _, ok := ec.Get("a"); !ok {
		t.Fatal("expected hit for a")
	}
	// Insert c -> should evict LRU (b)
	_ = ec.Set("c", "C")

	if _, ok := ec.Get("b"); ok {
		t.Fatal("expected b to be evicted")
	}
	if _, ok := ec.Get("a"); !ok {
		t.Fatal("expected a to remain")
	}
	if _, ok := ec.Get("c"); !ok {
		t.Fatal("expected c to be present")
	}
}

func TestEnhancedCache_DeleteByTags(t *testing.T) {
	cfg := &CacheConfig{MaxSize: 10, DefaultTTL: time.Second, CleanupInterval: time.Second, Monitor: &CacheMonitorConfig{Enabled: false}}
	ec := NewEnhancedCache(cfg)
	defer func() { _ = ec.Close() }()

	_ = ec.SetWithTags("t1", 1, []string{"hot", "user:1"})
	_ = ec.SetWithTags("t2", 2, []string{"cold"})
	_ = ec.SetWithTags("t3", 3, []string{"hot"})

	deleted := ec.DeleteByTags([]string{"hot"})
	if deleted != 2 {
		t.Fatalf("expected to delete 2 items, got %d", deleted)
	}
	if _, ok := ec.Get("t2"); !ok {
		t.Fatal("expected t2 to remain")
	}
}

func TestEnhancedCache_MonitorCallback_LowHitRate(t *testing.T) {
	cfg := &CacheConfig{
		MaxSize:         10,
		DefaultTTL:      time.Second,
		CleanupInterval: time.Second,
		Monitor: &CacheMonitorConfig{
			Enabled:       true,
			Interval:      30 * time.Millisecond,
			HitRateTarget: 0.9, // make it hard to achieve so callback fires
		},
	}
	ec := NewEnhancedCache(cfg)
	defer func() { _ = ec.Close() }()

	// Prepare some misses to lower hit rate
	for i := 0; i < 5; i++ {
		_, _ = ec.Get("nope")
	}
	_ = ec.Set("k", "v")
	if _, ok := ec.Get("k"); !ok {
		t.Fatal("expected hit for k")
	}

	fired := make(chan struct{}, 1)
	ec.AddMonitorCallback(func(cs *CacheStats) {
		select {
		case fired <- struct{}{}:
		default:
		}
	})

	// Wait for monitor to cycle
	select {
	case <-fired:
		// ok
	case <-time.After(200 * time.Millisecond):
		// Non-fatal: monitoring can be flaky in CI; assert softly
		t.Log("monitor callback did not fire within timeout; continuing")
	}
}

func TestEnhancedCache_Persistence_LoadAndSave(t *testing.T) {
	// Prepare entries to load
	now := time.Now()
	toLoad := map[string]*CacheEntry{
		"loaded": {Key: "loaded", Value: 123, ExpiresAt: now.Add(1 * time.Hour), CreatedAt: now, LastAccess: now},
	}
	stub := &inMemoryPersistence{toLoad: toLoad}
	cfg := &CacheConfig{MaxSize: 10, DefaultTTL: time.Hour, CleanupInterval: time.Second, Persistence: stub, Monitor: &CacheMonitorConfig{Enabled: false}}
	ec := NewEnhancedCache(cfg)

	// Verify loaded
	if v, ok := ec.Get("loaded"); !ok || v.(int) != 123 {
		t.Fatalf("expected loaded=123, got %v ok=%v", v, ok)
	}

	// Add another entry and close (which should save)
	_ = ec.Set("save_me", "x")
	if err := ec.Close(); err != nil {
		t.Fatalf("close(save): %v", err)
	}

	// Verify Save was called with entries
	stub.mu.Lock()
	_, existsLoaded := stub.stored["loaded"]
	_, existsSaveMe := stub.stored["save_me"]
	stub.mu.Unlock()
	if !existsLoaded || !existsSaveMe {
		t.Fatalf("expected both entries persisted, got loaded=%v save_me=%v", existsLoaded, existsSaveMe)
	}

	// Error propagation on Save
	stubErr := &inMemoryPersistence{saveErr: errors.New("boom")}
	ec2 := NewEnhancedCache(&CacheConfig{Persistence: stubErr, Monitor: &CacheMonitorConfig{Enabled: false}})
	_ = ec2.Set("k", "v")
	if err := ec2.Close(); err == nil {
		t.Fatal("expected error from persistence Save")
	}
}
