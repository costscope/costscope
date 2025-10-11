//go:build !sqlite

// Package persistence provides an in-memory stub for SQLite when the 'sqlite'
// build tag is not enabled. This allows producing a slim binary without CGO
// (github.com/mattn/go-sqlite3) while preserving the Repository API.
package persistence

import (
	"context"
	"errors"
	"sort"
	"strings"
	"sync"
	"time"

	streamingTypes "local/costscope/cmd/modules/streaming/types"
	providerTypes "local/costscope/internal/providers/types"
)

// SQLiteRepository is an in-memory implementation of Repository used in slim builds.
// It satisfies tests that exercise basic CRUD semantics without requiring CGO/SQLite.
type SQLiteRepository struct {
	mu        sync.RWMutex
	jobs      map[string]*streamingTypes.StreamingJobInfo
	providers map[string]*providerTypes.ProviderConfig
}

// NewSQLiteRepository returns an in-memory repository instance.
func NewSQLiteRepository(_ *DatabaseConfig) (*SQLiteRepository, error) { // nolint:revive
	return &SQLiteRepository{
		jobs:      make(map[string]*streamingTypes.StreamingJobInfo),
		providers: make(map[string]*providerTypes.ProviderConfig),
	}, nil
}

// SaveJob creates or replaces a job entry.
func (r *SQLiteRepository) SaveJob(_ context.Context, job *streamingTypes.StreamingJobInfo) error {
	if job == nil || job.Config.JobID == "" {
		return errors.New("invalid job")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	// Store a shallow copy to avoid external mutation surprises
	cp := *job
	r.jobs[job.Config.JobID] = &cp
	return nil
}

// GetJob returns a job by ID.
func (r *SQLiteRepository) GetJob(_ context.Context, jobID string) (*streamingTypes.StreamingJobInfo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	job, ok := r.jobs[jobID]
	if !ok {
		return nil, errors.New("job not found: " + jobID)
	}
	cp := *job
	return &cp, nil
}

// ListJobs returns jobs filtered with simple in-memory predicates.
func (r *SQLiteRepository) ListJobs(_ context.Context, filters JobFilters) ([]*streamingTypes.StreamingJobInfo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*streamingTypes.StreamingJobInfo, 0, len(r.jobs))
	for _, j := range r.jobs {
		if len(filters.Status) > 0 {
			match := false
			for _, s := range filters.Status {
				if strings.EqualFold(j.Status.Status, s) {
					match = true
					break
				}
			}
			if !match {
				continue
			}
		}
		if len(filters.Provider) > 0 {
			match := false
			for _, p := range filters.Provider {
				if strings.EqualFold(j.Config.Provider, p) {
					match = true
					break
				}
			}
			if !match {
				continue
			}
		}
		if filters.CreatedAt.From != nil && j.Config.CreatedAt.Before(*filters.CreatedAt.From) {
			continue
		}
		if filters.CreatedAt.To != nil && j.Config.CreatedAt.After(*filters.CreatedAt.To) {
			continue
		}
		if filters.UpdatedAt.From != nil && j.Config.UpdatedAt.Before(*filters.UpdatedAt.From) {
			continue
		}
		if filters.UpdatedAt.To != nil && j.Config.UpdatedAt.After(*filters.UpdatedAt.To) {
			continue
		}
		cp := *j
		out = append(out, &cp)
	}

	// Basic sorting support on created_at/updated_at
	sort.Slice(out, func(i, j int) bool {
		switch strings.ToLower(filters.SortBy) {
		case "updated_at":
			if strings.ToUpper(filters.SortOrder) == "DESC" {
				return out[i].Config.UpdatedAt.After(out[j].Config.UpdatedAt)
			}
			return out[i].Config.UpdatedAt.Before(out[j].Config.UpdatedAt)
		default: // created_at
			if strings.ToUpper(filters.SortOrder) == "DESC" {
				return out[i].Config.CreatedAt.After(out[j].Config.CreatedAt)
			}
			return out[i].Config.CreatedAt.Before(out[j].Config.CreatedAt)
		}
	})

	// Pagination
	start := filters.Offset
	if start < 0 {
		start = 0
	}
	end := len(out)
	if filters.Limit > 0 && start+filters.Limit < end {
		end = start + filters.Limit
	}
	if start > len(out) {
		return []*streamingTypes.StreamingJobInfo{}, nil
	}
	return out[start:end], nil
}

// UpdateJobStatus updates status fields for a job.
func (r *SQLiteRepository) UpdateJobStatus(_ context.Context, jobID string, status *streamingTypes.StreamingJobStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	job, ok := r.jobs[jobID]
	if !ok {
		return errors.New("job not found: " + jobID)
	}
	// Preserve immutable fields; update status and timestamps
	job.Status = *status
	job.Config.UpdatedAt = time.Now()
	return nil
}

// DeleteJob removes a job.
func (r *SQLiteRepository) DeleteJob(_ context.Context, jobID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.jobs[jobID]; !ok {
		return errors.New("job not found: " + jobID)
	}
	delete(r.jobs, jobID)
	return nil
}

// SaveProvider creates or replaces a provider config.
func (r *SQLiteRepository) SaveProvider(_ context.Context, config *providerTypes.ProviderConfig) error {
	if config == nil || config.Name == "" {
		return errors.New("invalid provider config")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	cp := *config
	r.providers[config.Name] = &cp
	return nil
}

// GetProvider returns a provider config by name.
func (r *SQLiteRepository) GetProvider(_ context.Context, name string) (*providerTypes.ProviderConfig, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.providers[name]
	if !ok {
		return nil, errors.New("provider not found: " + name)
	}
	cp := *p
	return &cp, nil
}

// ListProviders lists all providers.
func (r *SQLiteRepository) ListProviders(_ context.Context) ([]*providerTypes.ProviderConfig, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*providerTypes.ProviderConfig, 0, len(r.providers))
	for _, p := range r.providers {
		cp := *p
		out = append(out, &cp)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

// DeleteProvider removes a provider by name.
func (r *SQLiteRepository) DeleteProvider(_ context.Context, name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.providers[name]; !ok {
		return errors.New("provider not found: " + name)
	}
	delete(r.providers, name)
	return nil
}

// Health returns nil if the stub is operational.
func (r *SQLiteRepository) Health(_ context.Context) error { return nil }

// Close releases resources (no-op for in-memory stub).
func (r *SQLiteRepository) Close() error { return nil }
