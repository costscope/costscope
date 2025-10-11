package exporters

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// genRandSuffix returns a short random hex suffix
func genRandSuffix(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// TestObjectStore_S3Smoke uploads a tiny blob to s3 when bucket env is present.
func TestObjectStore_S3Smoke(t *testing.T) {
	bucket := os.Getenv("COSTSCOPE_TEST_S3_BUCKET")
	if bucket == "" {
		t.Skip("COSTSCOPE_TEST_S3_BUCKET not set; skipping S3 smoke test")
	}
	key := filepath.Join("costscope-test", "reports", "smoke-"+genRandSuffix(4)+".json")
	dest := "s3://" + bucket + "/" + key
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	store, _, err := NewObjectStore(ctx, dest)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	payload := []byte(`{"ok":true}`)
	if err := store.Put(ctx, dest, bytes.NewReader(payload)); err != nil {
		t.Fatalf("s3 put: %v", err)
	}
}

// TestObjectStore_GCSSmoke uploads a tiny blob to gcs when bucket env is present.
func TestObjectStore_GCSSmoke(t *testing.T) {
	bucket := os.Getenv("COSTSCOPE_TEST_GCS_BUCKET")
	if bucket == "" {
		t.Skip("COSTSCOPE_TEST_GCS_BUCKET not set; skipping GCS smoke test")
	}
	key := filepath.Join("costscope-test", "reports", "smoke-"+genRandSuffix(4)+".json")
	dest := "gs://" + bucket + "/" + key
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	store, _, err := NewObjectStore(ctx, dest)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	payload := []byte(`{"ok":true}`)
	if err := store.Put(ctx, dest, bytes.NewReader(payload)); err != nil {
		t.Fatalf("gcs put: %v", err)
	}
}
