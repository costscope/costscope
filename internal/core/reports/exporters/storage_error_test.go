package exporters

import (
	"bytes"
	"context"
	"testing"
)

// TestS3Store_Put_InvalidURL covers early validation error path (missing bucket host).
func TestS3Store_Put_InvalidURL(t *testing.T) {
	t.Parallel()
	s := &s3Store{} // nil client fine because branch returns before use
	if err := s.Put(context.Background(), "s3://", bytes.NewReader([]byte("x"))); err == nil {
		t.Fatalf("expected invalid s3 url error")
	}
}

// TestGCSStore_Put_InvalidURL covers early validation error path (missing bucket host).
func TestGCSStore_Put_InvalidURL(t *testing.T) {
	t.Parallel()
	g := &gcsStore{} // nil client fine because branch returns before use
	if err := g.Put(context.Background(), "gs://", bytes.NewReader([]byte("x"))); err == nil {
		t.Fatalf("expected invalid gs url error")
	}
}
