package store

import (
	"sync"
	"time"

	focustypes "local/costscope/internal/core/focus/types"
)

// JobStore is a small persistence abstraction used to store conversion results
// for local inspection or future durable backends. This in-memory implementation
// is intentionally minimal and suitable for tests and single-process usage.
type JobStore interface {
	SaveResult(res *focustypes.ConversionResult) error
	ListResults(limit int) []*focustypes.ConversionResult
	FinalizeResultTiming(conversionID string, end time.Time, duration time.Duration) error
}

// InMemoryJobStore is a threadsafe, in-memory store for ConversionResult
type InMemoryJobStore struct {
	mu    sync.RWMutex
	items []*focustypes.ConversionResult
}

// NewInMemoryJobStore creates a new empty in-memory store
func NewInMemoryJobStore() *InMemoryJobStore {
	return &InMemoryJobStore{items: make([]*focustypes.ConversionResult, 0)}
}

// SaveResult appends or replaces a conversion result
func (s *InMemoryJobStore) SaveResult(res *focustypes.ConversionResult) error {
	if res == nil {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	// If a result with same ConversionId exists, replace it
	for i, e := range s.items {
		if e != nil && e.ConversionId == res.ConversionId {
			s.items[i] = res
			return nil
		}
	}
	// Otherwise append
	s.items = append(s.items, res)
	return nil
}

// ListResults returns up to `limit` results in insertion order (most recent last).
// If limit <= 0, all results are returned.
func (s *InMemoryJobStore) ListResults(limit int) []*focustypes.ConversionResult {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if limit <= 0 || limit >= len(s.items) {
		// return a copy to avoid callers mutating internal slice
		out := make([]*focustypes.ConversionResult, len(s.items))
		copy(out, s.items)
		return out
	}
	out := make([]*focustypes.ConversionResult, limit)
	copy(out, s.items[len(s.items)-limit:])
	return out
}

// FinalizeResultTiming updates timing fields for an existing result.
func (s *InMemoryJobStore) FinalizeResultTiming(conversionID string, end time.Time, duration time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, e := range s.items {
		if e != nil && e.ConversionId == conversionID {
			e.EndTime = end
			e.Duration = duration
			return nil
		}
	}
	return nil
}
