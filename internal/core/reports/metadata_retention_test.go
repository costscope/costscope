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

// TestInMemoryRetention ensures age and count pruning works for in-memory store.
func TestInMemoryRetention(t *testing.T) {
	log := logging.NewLogger(logging.LevelError)
	store := NewInMemoryMetadataStore(log, 2, 30*time.Minute)
	old := &ReportMetadata{ID: "old", Path: "old.json", Format: "json", SizeBytes: 1, CreatedAt: time.Now().Add(-2 * time.Hour).UTC()}
	_ = store.Save(context.Background(), old)
	for i := 0; i < 3; i++ {
		md := &ReportMetadata{ID: fID(i), Path: fID(i) + ".json", Format: "json", SizeBytes: int64(i), CreatedAt: time.Now().Add(time.Duration(i) * time.Minute).UTC()}
		_ = store.Save(context.Background(), md)
	}
	all, _ := store.List(context.Background())
	if len(all) != 2 {
		// Expect newest two f2,f1
		ids := []string{}
		for _, r := range all {
			ids = append(ids, r.ID)
		}
		t.Fatalf("expected 2 after retention got %d: %v", len(all), ids)
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

// TestFileRetention ensures file store rewrites when thresholds exceeded.
func TestFileRetention(t *testing.T) {
	log := logging.NewLogger(logging.LevelError)
	dir := t.TempDir()
	path := filepath.Join(dir, "meta.jsonl")
	store := NewFileMetadataStore(path, log, 2, 30*time.Minute)
	old := &ReportMetadata{ID: "old", Path: "old.json", Format: "json", SizeBytes: 1, CreatedAt: time.Now().Add(-2 * time.Hour).UTC()}
	_ = store.Save(context.Background(), old)
	for i := 0; i < 3; i++ {
		md := &ReportMetadata{ID: fID(i), Path: fID(i) + ".json", Format: "json", SizeBytes: int64(i), CreatedAt: time.Now().Add(time.Duration(i) * time.Minute).UTC()}
		_ = store.Save(context.Background(), md)
	}
	all, _ := store.List(context.Background())
	if len(all) != 2 {
		t.Fatalf("expected 2 got %d", len(all))
	}
	// Path comes from t.TempDir() + constant filename; safe for test. #nosec G304
	content, err := os.ReadFile(path) // #nosec G304
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	if len(content) == 0 {
		t.Fatalf("expected content after retention rewrite")
	}
}

// local helper (duplicated minimal to avoid test import churn)
func fID(i int) string { return fmt.Sprintf("f%v", i) }
