package reports

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/costscope/costscope/internal/core/logging"
)

// ReportMetadata holds persisted export level metadata (MVP – minimal fields)
type ReportMetadata struct {
	ID              string    `json:"id"`
	Path            string    `json:"path"`
	Format          string    `json:"format"`
	SizeBytes       int64     `json:"size_bytes"`
	ChecksumSHA256  string    `json:"checksum_sha256"`
	CreatedAt       time.Time `json:"created_at"`
	OutputDirSource string    `json:"output_dir_source"`
}

// MetadataListOptions filters listing results (all fields optional).
type MetadataListOptions struct {
	Format        string
	CreatedAfter  *time.Time
	CreatedBefore *time.Time
	Offset        int
	Limit         int // 0 or negative => no explicit limit (return all after offset)
}

// MetadataStore abstracts persistence for report metadata.
type MetadataStore interface {
	Save(ctx context.Context, md *ReportMetadata) error
	Get(ctx context.Context, id string) (*ReportMetadata, error)
	// List returns all metadata records (unfiltered); retained for backward compatibility with simple stores.
	List(ctx context.Context) ([]*ReportMetadata, error)
	// ListOptions returns filtered/paginated metadata. Default implementation may fall back to List and filter in-memory.
	ListOptions(ctx context.Context, opts *MetadataListOptions) ([]*ReportMetadata, error)
}

// InMemoryMetadataStore simple in-memory implementation (thread-safe via channel ops minimal needs).
type InMemoryMetadataStore struct {
	items   map[string]*ReportMetadata
	log     *logging.Logger
	maxRecs int
	maxAge  time.Duration
}

func NewInMemoryMetadataStore(log *logging.Logger, maxRecords int, maxAge time.Duration) *InMemoryMetadataStore {
	return &InMemoryMetadataStore{items: make(map[string]*ReportMetadata), log: log, maxRecs: maxRecords, maxAge: maxAge}
}

func (s *InMemoryMetadataStore) Save(_ context.Context, md *ReportMetadata) error {
	s.items[md.ID] = md
	s.prune()
	return nil
}

func (s *InMemoryMetadataStore) Get(_ context.Context, id string) (*ReportMetadata, error) {
	md, ok := s.items[id]
	if !ok {
		return nil, fmt.Errorf("metadata not found: %s", id)
	}
	return md, nil
}

func (s *InMemoryMetadataStore) List(_ context.Context) ([]*ReportMetadata, error) {
	out := make([]*ReportMetadata, 0, len(s.items))
	for _, v := range s.items {
		out = append(out, v)
	}
	return out, nil
}

func (s *InMemoryMetadataStore) ListOptions(ctx context.Context, opts *MetadataListOptions) ([]*ReportMetadata, error) {
	all, err := s.List(ctx)
	if err != nil {
		return nil, err
	}
	return filterAndPageMetadata(all, opts), nil
}

// FileMetadataStore provides a simple durable metadata store backed by JSON lines file.
// Each Save appends (best-effort). List rewinds and decodes all lines. Not optimized for very large counts.
type FileMetadataStore struct {
	path    string
	log     *logging.Logger
	maxRecs int
	maxAge  time.Duration
}

func NewFileMetadataStore(path string, log *logging.Logger, opts ...interface{}) *FileMetadataStore {
	// Backward-compatible variadic: allow passing maxRecs (int) and maxAge (time.Duration) in order.
	fs := &FileMetadataStore{path: path, log: log}
	if len(opts) > 0 {
		if v, ok := opts[0].(int); ok {
			fs.maxRecs = v
		}
	}
	if len(opts) > 1 {
		if v, ok := opts[1].(time.Duration); ok {
			fs.maxAge = v
		}
	}
	return fs
}

func (s *FileMetadataStore) Save(_ context.Context, md *ReportMetadata) error {
	f, err := openFileAppend(s.path)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	line, err := jsonMarshalNoErr(md)
	if err != nil {
		return err
	}
	if _, err := f.Write(append(line, '\n')); err != nil {
		return err
	}
	// Opportunistic pruning (best-effort). Only rewrite file when thresholds exceeded to limit IO.
	if s.maxRecs > 0 || s.maxAge > 0 {
		s.pruneIfNeeded()
	}
	return nil
}

func (s *FileMetadataStore) Get(ctx context.Context, id string) (*ReportMetadata, error) {
	all, err := s.List(ctx)
	if err != nil {
		return nil, err
	}
	for _, md := range all {
		if md.ID == id {
			return md, nil
		}
	}
	return nil, fmt.Errorf("metadata not found: %s", id)
}

func (s *FileMetadataStore) List(_ context.Context) ([]*ReportMetadata, error) {
	data, err := os.ReadFile(s.path) // #nosec G304 path controlled by config
	if err != nil {
		if os.IsNotExist(err) {
			return []*ReportMetadata{}, nil
		}
		return nil, err
	}
	lines := bytes.Split(data, []byte("\n"))
	out := make([]*ReportMetadata, 0, len(lines))
	for _, ln := range lines {
		if len(ln) == 0 {
			continue
		}
		var md ReportMetadata
		if err := json.Unmarshal(ln, &md); err == nil {
			out = append(out, &md)
		}
	}
	return out, nil
}

func (s *FileMetadataStore) ListOptions(ctx context.Context, opts *MetadataListOptions) ([]*ReportMetadata, error) {
	all, err := s.List(ctx)
	if err != nil {
		return nil, err
	}
	return filterAndPageMetadata(all, opts), nil
}

// helpers (local minimal to avoid new imports at call sites)
func openFileAppend(path string) (*os.File, error) {
	// Directories should be at most 0750 per security lint (G301)
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return nil, err
	}
	// Path originates from validated configuration; #nosec G304 to acknowledge controlled input
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600) // #nosec G304
	if err != nil {
		return nil, err
	}
	return f, nil
}
func jsonMarshalNoErr(v interface{}) ([]byte, error) { return json.Marshal(v) }

// NOTE: checksum helper removed (unused in MVP); reintroduce when persisted exports stored.

// filterAndPageMetadata applies filtering and pagination on an in-memory slice.
func filterAndPageMetadata(in []*ReportMetadata, opts *MetadataListOptions) []*ReportMetadata {
	if opts == nil {
		// ensure stable ordering (newest first) even for unfiltered results
		sort.Slice(in, func(i, j int) bool { return in[i].CreatedAt.After(in[j].CreatedAt) })
		return in
	}
	filtered := make([]*ReportMetadata, 0, len(in))
	for _, md := range in {
		if opts.Format != "" && opts.Format != md.Format {
			continue
		}
		if opts.CreatedAfter != nil && md.CreatedAt.Before(*opts.CreatedAfter) {
			continue
		}
		if opts.CreatedBefore != nil && md.CreatedAt.After(*opts.CreatedBefore) {
			continue
		}
		filtered = append(filtered, md)
	}
	// stable ordering newest first
	sort.Slice(filtered, func(i, j int) bool { return filtered[i].CreatedAt.After(filtered[j].CreatedAt) })
	start := opts.Offset
	if start < 0 {
		start = 0
	}
	if start > len(filtered) {
		return []*ReportMetadata{}
	}
	end := len(filtered)
	if opts.Limit > 0 && start+opts.Limit < end {
		end = start + opts.Limit
	}
	return filtered[start:end]
}

// prune applies retention rules (in-memory only). Age pruning then count pruning (newest kept).
func (s *InMemoryMetadataStore) prune() {
	if s.maxRecs <= 0 && s.maxAge <= 0 {
		return
	}
	now := time.Now()
	if s.maxAge > 0 {
		cutoff := now.Add(-s.maxAge)
		for id, md := range s.items {
			if md.CreatedAt.Before(cutoff) {
				delete(s.items, id)
			}
		}
	}
	if s.maxRecs > 0 && len(s.items) > s.maxRecs {
		// Collect and sort by CreatedAt desc keep newest maxRecs
		slice := make([]*ReportMetadata, 0, len(s.items))
		for _, md := range s.items {
			slice = append(slice, md)
		}
		sort.Slice(slice, func(i, j int) bool { return slice[i].CreatedAt.After(slice[j].CreatedAt) })
		for i := s.maxRecs; i < len(slice); i++ {
			delete(s.items, slice[i].ID)
		}
	}
}

// pruneIfNeeded loads all records, applies age + count pruning, and rewrites file when changes occur.
func (s *FileMetadataStore) pruneIfNeeded() {
	if s.maxRecs <= 0 && s.maxAge <= 0 {
		return
	}
	records, err := s.List(context.Background())
	if err != nil {
		return
	}
	changed := false
	if s.maxAge > 0 {
		cutoff := time.Now().Add(-s.maxAge)
		tmp := records[:0]
		for _, r := range records {
			if r.CreatedAt.Before(cutoff) {
				changed = true
				continue
			}
			tmp = append(tmp, r)
		}
		records = tmp
	}
	if s.maxRecs > 0 && len(records) > s.maxRecs {
		sort.Slice(records, func(i, j int) bool { return records[i].CreatedAt.After(records[j].CreatedAt) })
		records = records[:s.maxRecs]
		changed = true
	}
	if !changed {
		return
	}
	// Rewrite file with remaining records (newest-first ordering already ensured above when count pruning; apply ordering regardless for determinism).
	sort.Slice(records, func(i, j int) bool { return records[i].CreatedAt.After(records[j].CreatedAt) })
	buf := bytes.Buffer{}
	for _, r := range records {
		b, err := json.Marshal(r)
		if err != nil {
			continue
		}
		buf.Write(b)
		buf.WriteByte('\n')
	}
	// Atomic rewrite: write to temp then rename.
	tmpPath := s.path + ".tmp"
	if err := os.WriteFile(tmpPath, buf.Bytes(), 0o600); err == nil {
		_ = os.Rename(tmpPath, s.path)
	} else {
		_ = os.Remove(tmpPath)
	}
}
