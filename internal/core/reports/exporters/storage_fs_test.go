package exporters

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestFSStore_Put_FileSchemeAndPlain exercises fsStore normal and file:// paths.
func TestFSStore_Put_FileSchemeAndPlain(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	destPlain := filepath.Join(tmp, "nested", "report.txt")
	payload := []byte("hello")

	// Plain path via NewObjectStore fallback
	store, kind, err := NewObjectStore(context.Background(), destPlain)
	if err != nil || kind != "fs" {
		t.Fatalf("expected fs kind plain path, got kind=%s err=%v", kind, err)
	}
	if err := store.Put(context.Background(), destPlain, bytes.NewReader(payload)); err != nil {
		t.Fatalf("put plain: %v", err)
	}
	// This read is safe: the path is created within t.TempDir() and not influenced by external input.
	// nolint:gosec // G304: file read is constrained to test-controlled temporary directory
	if b, err := os.ReadFile(destPlain); err != nil || string(b) != "hello" {
		t.Fatalf("unexpected file content: %v %q", err, string(b))
	}

	// file:// scheme path to cover URL parsing branch
	destScheme := "file://" + filepath.Join(tmp, "scheme", "report2.txt")
	if err := store.Put(context.Background(), destScheme, bytes.NewReader([]byte("world"))); err != nil {
		t.Fatalf("put scheme: %v", err)
	}
	// Resolve actual path (strip scheme) for existence check
	actual := filepath.Join(tmp, "scheme", "report2.txt")
	if _, err := os.Stat(actual); err != nil {
		t.Fatalf("expected file at %s: %v", actual, err)
	}
}

// TestFSStore_Put_InvalidParent triggers MkdirAll error by using a parent path that is a file.
func TestFSStore_Put_InvalidParent(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	parentFile := filepath.Join(tmp, "parentAsFile")
	if err := os.WriteFile(parentFile, []byte("x"), 0o600); err != nil {
		t.Fatalf("prep parent file: %v", err)
	}
	dest := filepath.Join(parentFile, "child.txt")
	store := &fsStore{}
	if err := store.Put(context.Background(), dest, bytes.NewReader([]byte("x"))); err == nil {
		t.Fatalf("expected error due to parent being file")
	}
}

// TestNewObjectStore_UnsupportedScheme ensures unknown schemes fall back to fs.
func TestNewObjectStore_UnsupportedScheme(t *testing.T) {
	t.Parallel()
	store, kind, err := NewObjectStore(context.Background(), "ftp://example.com/file.txt")
	if err != nil || kind != "fs" || store == nil {
		t.Fatalf("expected fs fallback, kind=%s err=%v store=%v", kind, err, store)
	}
}

// TestNewObjectStore_S3Constructor covers newS3Store creation (may succeed without credentials).
func TestNewObjectStore_S3Constructor(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	_, kind, err := NewObjectStore(ctx, "s3://bucket/key.txt")
	if err != nil { // allow aws config issues in unusual envs but still covered
		t.Logf("s3 constructor error (acceptable for coverage): %v", err)
	} else if kind != "s3" {
		t.Fatalf("expected kind s3 got %s", kind)
	}
}

// TestNewObjectStore_GCSConstructor covers newGCSStore path; error acceptable when ADC not present.
func TestNewObjectStore_GCSConstructor(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	_, kind, err := NewObjectStore(ctx, "gs://bucket/key.txt")
	if err != nil {
		// In minimal CI environments without ADC credentials this will error; still consider covered
		t.Logf("gcs constructor error (acceptable): %v", err)
		return
	}
	if kind != "gcs" {
		t.Fatalf("expected kind gcs got %s", kind)
	}
}
