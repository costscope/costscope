//go:build ignore
// +build ignore

package gcp_test

import (
	"context"

	"github.com/costscope/costscope/internal/core/focus/types"
)

// testWriter is a minimal in-memory DataWriter for provider tests.
type testWriter struct{ recs []types.FocusRecord }

func (w *testWriter) Open(ctx context.Context, path string, schema *types.FocusSchema) error {
	return nil
}
func (w *testWriter) Write(ctx context.Context, records []types.FocusRecord) error {
	w.recs = append(w.recs, records...)
	return nil
}
func (w *testWriter) WriteChunk(ctx context.Context, data []byte) error { return nil }
func (w *testWriter) Flush(ctx context.Context) error                   { return nil }
func (w *testWriter) Close() error                                      { return nil }
func (w *testWriter) GetMetadata() *types.DataDestinationMetadata {
	return &types.DataDestinationMetadata{}
}
