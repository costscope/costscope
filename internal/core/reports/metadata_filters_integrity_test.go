package reports

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"local/costscope/internal/core/logging"
)

// TestListReportMetadataOptions exercises format/date/pagination filtering logic.
func TestListReportMetadataOptions(t *testing.T) {
	logger := logging.NewLogger("test")
	store := NewInMemoryMetadataStore(logger, 0, 0)
	svc := NewBasicReportService(logger).WithMetadataStore(store)

	now := time.Now().UTC()
	// Insert three records with staggered timestamps & different formats
	recs := []*ReportMetadata{
		{ID: "a", Path: "p1.json", Format: "json", SizeBytes: 10, ChecksumSHA256: "x", CreatedAt: now.Add(-3 * time.Hour)},
		{ID: "b", Path: "p2.csv", Format: "csv", SizeBytes: 20, ChecksumSHA256: "y", CreatedAt: now.Add(-2 * time.Hour)},
		{ID: "c", Path: "p3.json", Format: "json", SizeBytes: 30, ChecksumSHA256: "z", CreatedAt: now.Add(-1 * time.Hour)},
	}
	for _, r := range recs {
		_ = store.Save(context.Background(), r)
	}

	// Format filter json should return two newest-first (c, a)
	list, err := svc.ListReportMetadataOptions(context.Background(), &MetadataListOptions{Format: "json"})
	if err != nil {
		t.Fatalf("list error: %v", err)
	}
	if len(list) != 2 || list[0].ID != "c" || list[1].ID != "a" {
		t.Fatalf("unexpected format filter result: %#v", list)
	}

	// CreatedAfter filter (after a) should return b,c (newest-first => c,b)
	after := now.Add(-150 * time.Minute) // between a (-180m) and b (-120m)
	list, err = svc.ListReportMetadataOptions(context.Background(), &MetadataListOptions{CreatedAfter: &after})
	if err != nil {
		t.Fatalf("list error: %v", err)
	}
	if len(list) != 2 || list[0].ID != "c" || list[1].ID != "b" {
		t.Fatalf("unexpected CreatedAfter result: %#v", list)
	}

	// Pagination: limit 1 offset 1 over full set (newest order c,b,a) should return b only
	list, err = svc.ListReportMetadataOptions(context.Background(), &MetadataListOptions{Limit: 1, Offset: 1})
	if err != nil {
		t.Fatalf("list error: %v", err)
	}
	if len(list) != 1 || list[0].ID != "b" {
		t.Fatalf("unexpected pagination result: %#v", list)
	}
}

// TestVerifyReportIntegrity covers match and mismatch scenarios.
func TestVerifyReportIntegrity(t *testing.T) {
	logger := logging.NewLogger("test")
	dir := t.TempDir()
	path := filepath.Join(dir, "report.json")
	content := []byte("hello world")
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}
	sum := sha256.Sum256(content)
	checksum := fmt.Sprintf("%x", sum[:])

	store := NewInMemoryMetadataStore(logger, 0, 0)
	svc := NewBasicReportService(logger).WithMetadataStore(store)
	md := &ReportMetadata{ID: "exp1", Path: path, Format: "json", SizeBytes: int64(len(content)), ChecksumSHA256: checksum, CreatedAt: time.Now().UTC()}
	if err := store.Save(context.Background(), md); err != nil {
		t.Fatalf("save: %v", err)
	}

	match, actual, err := svc.VerifyReportIntegrity(context.Background(), "exp1")
	if err != nil {
		t.Fatalf("verify error: %v", err)
	}
	if !match || actual != checksum {
		t.Fatalf("expected match; got match=%v actual=%s", match, actual)
	}

	// Modify file -> mismatch
	if err := os.WriteFile(path, append(content, 'X'), 0o600); err != nil {
		t.Fatalf("modify file: %v", err)
	}
	match, actual, err = svc.VerifyReportIntegrity(context.Background(), "exp1")
	if err != nil {
		t.Fatalf("verify error: %v", err)
	}
	if match || actual == checksum {
		t.Fatalf("expected mismatch; got match=%v actual=%s", match, actual)
	}
}
