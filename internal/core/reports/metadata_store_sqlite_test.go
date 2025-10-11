//go:build sqlite

package reports

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"local/costscope/internal/core/logging"
)

func TestSQLiteMetadataStoreBasicListingAndPagination(t *testing.T) {
	log := logging.NewLogger(logging.LevelDebug)
	tmp := t.TempDir()
	path := filepath.Join(tmp, "meta.db")
	store, err := NewSQLiteMetadataStore(path, log, 0, 0)
	if err != nil {
		t.Fatalf("init sqlite store: %v", err)
	}
	ctx := context.Background()
	// Insert 3 records with staggered CreatedAt
	for i := 0; i < 3; i++ {
		md := &ReportMetadata{ID: fIDSQLite(i), Path: filepath.Join(tmp, fIDSQLite(i)+".json"), Format: "json", SizeBytes: int64(10 + i), CreatedAt: time.Now().Add(time.Duration(i) * time.Minute).UTC()}
		if err := store.Save(ctx, md); err != nil {
			t.Fatalf("save: %v", err)
		}
	}
	listAll, err := store.List(ctx)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(listAll) != 3 {
		t.Fatalf("expected 3, got %d", len(listAll))
	}
	// Pagination limit=1 offset=1
	opts := &MetadataListOptions{Limit: 1, Offset: 1}
	paged, err := store.ListOptions(ctx, opts)
	if err != nil {
		t.Fatalf("list options: %v", err)
	}
	if len(paged) != 1 {
		t.Fatalf("expected 1 paged, got %d", len(paged))
	}
}

func TestSQLiteMetadataStoreFormatAndDateFilters(t *testing.T) {
	log := logging.NewLogger(logging.LevelDebug)
	path := filepath.Join(t.TempDir(), "meta.db")
	store, err := NewSQLiteMetadataStore(path, log, 0, 0)
	if err != nil {
		t.Fatalf("init: %v", err)
	}
	ctx := context.Background()
	now := time.Now().UTC()
	md1 := &ReportMetadata{ID: "a", Path: "a.json", Format: "json", SizeBytes: 1, CreatedAt: now.Add(-2 * time.Hour)}
	md2 := &ReportMetadata{ID: "b", Path: "b.csv", Format: "csv", SizeBytes: 2, CreatedAt: now.Add(-1 * time.Hour)}
	_ = store.Save(ctx, md1)
	_ = store.Save(ctx, md2)
	after := now.Add(-90 * time.Minute)
	list, err := store.ListOptions(ctx, &MetadataListOptions{Format: "csv", CreatedAfter: &after})
	if err != nil {
		t.Fatalf("list opts: %v", err)
	}
	if len(list) != 1 || list[0].ID != "b" {
		t.Fatalf("expected record b, got %+v", list)
	}
}

func TestSQLiteMetadataStoreRetention(t *testing.T) {
	log := logging.NewLogger(logging.LevelDebug)
	path := filepath.Join(t.TempDir(), "meta.db")
	// retention: max 2 records, max age 30m
	store, err := NewSQLiteMetadataStore(path, log, 2, 30*time.Minute)
	if err != nil {
		t.Fatalf("init: %v", err)
	}
	ctx := context.Background()

	old := &ReportMetadata{ID: "old", Path: "old.json", Format: "json", SizeBytes: 1, CreatedAt: time.Now().Add(-2 * time.Hour).UTC()}
	if err := store.Save(ctx, old); err != nil {
		t.Fatalf("save old: %v", err)
	}
	for i := 0; i < 3; i++ {
		md := &ReportMetadata{ID: fIDSQLite(i), Path: fIDSQLite(i) + ".json", Format: "json", SizeBytes: int64(i), CreatedAt: time.Now().Add(time.Duration(i) * time.Minute).UTC()}
		if err := store.Save(ctx, md); err != nil {
			t.Fatalf("save: %v", err)
		}
	}
	// After retention prune: age-based removes 'old'; count-based leaves 2 newest of remaining 3 (f2,f1)
	all, err := store.List(ctx)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(all) != 2 {
		t.Fatalf("expected 2 after retention, got %d", len(all))
	}
	ids := map[string]struct{}{}
	for _, r := range all {
		ids[r.ID] = struct{}{}
	}
	if _, ok := ids["f2"]; !ok {
		t.Fatalf("expected f2 present")
	}
	if _, ok := ids["f1"]; !ok {
		t.Fatalf("expected f1 present")
	}
	if _, ok := ids["old"]; ok {
		t.Fatalf("old should be pruned")
	}
}

// Helper
// fIDSQLite is a local helper distinct from other test helpers to avoid redeclaration under combined build tags.
func fIDSQLite(i int) string { return fmt.Sprintf("f%d", i) }

func TestMigrateFileMetadataToSQLite(t *testing.T) {
	log := logging.NewLogger(logging.LevelDebug)
	tmp := t.TempDir()
	filePath := filepath.Join(tmp, "meta.jsonl")
	// seed file store
	fs := NewFileMetadataStore(filePath, log)
	_ = fs.Save(context.Background(), &ReportMetadata{ID: "x", Path: "x.json", Format: "json", SizeBytes: 1, CreatedAt: time.Now().UTC()})
	sqlitePath := filepath.Join(tmp, "meta.db")
	store, err := NewSQLiteMetadataStore(sqlitePath, log, 0, 0)
	if err != nil {
		t.Fatalf("init sqlite: %v", err)
	}
	if err := MigrateFileMetadataToSQLite(context.Background(), filePath, store); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	res, err := store.Get(context.Background(), "x")
	if err != nil {
		t.Fatalf("get migrated: %v", err)
	}
	if res.ID != "x" {
		t.Fatalf("expected migrated id x")
	}
	if _, err := os.Stat(filePath); err != nil {
		t.Fatalf("expected file to remain: %v", err)
	}
}
