package cache

import (
	"container/list"
	"sync"
	"sync/atomic"
)

// lruEntry holds a single key/value pair in the list.
type lruEntry struct {
	key string
	val string
}

// LRU is a bounded thread-safe string->string cache optimized for read-heavy workloads.
// Fast path uses an atomic snapshot map; writes rebuild the snapshot under a lock.
type LRU struct {
	mu       sync.Mutex
	capacity int
	ll       *list.List
	items    map[string]*list.Element
	snap     atomic.Value // map[string]string
	hits     atomic.Uint64
	misses   atomic.Uint64
	evicts   atomic.Uint64
}

// NewLRU creates a new LRU. capacity <= 0 means unbounded (not recommended for hot paths).
func NewLRU(capacity int) *LRU {
	l := &LRU{capacity: capacity, ll: list.New(), items: make(map[string]*list.Element)}
	l.snap.Store(make(map[string]string))
	return l
}

// Get returns (value,true) if present; best-effort promotion on hit.
func (l *LRU) Get(key string) (string, bool) {
	if key == "" {
		return "", false
	}
	if m, _ := l.snap.Load().(map[string]string); m != nil {
		if v, ok := m[key]; ok {
			l.hits.Add(1)
			if el, exists := l.items[key]; exists { // promotion (best-effort)
				l.mu.Lock()
				l.ll.MoveToFront(el)
				l.mu.Unlock()
			}
			return v, true
		}
	}
	l.misses.Add(1)
	return "", false
}

// Add inserts or updates key with value, evicting LRU if capacity exceeded.
func (l *LRU) Add(key, val string) {
	if key == "" {
		return
	}
	l.mu.Lock()
	if el, ok := l.items[key]; ok {
		el.Value.(*lruEntry).val = val
		l.ll.MoveToFront(el)
		l.rebuildSnapshotLocked()
		l.mu.Unlock()
		return
	}
	el := l.ll.PushFront(&lruEntry{key: key, val: val})
	l.items[key] = el
	if l.capacity > 0 && l.ll.Len() > l.capacity {
		if tail := l.ll.Back(); tail != nil {
			ent := tail.Value.(*lruEntry)
			delete(l.items, ent.key)
			l.ll.Remove(tail)
			l.evicts.Add(1)
		}
	}
	l.rebuildSnapshotLocked()
	l.mu.Unlock()
}

func (l *LRU) rebuildSnapshotLocked() {
	snap := make(map[string]string, len(l.items))
	for k, el := range l.items {
		snap[k] = el.Value.(*lruEntry).val
	}
	l.snap.Store(snap)
}

// Stats exposes internal counters for tests/metrics exposition wrappers.
func (l *LRU) Stats() (hits, misses, evicts uint64, size int) {
	hits = l.hits.Load()
	misses = l.misses.Load()
	evicts = l.evicts.Load()
	size = len(l.items)
	return
}
