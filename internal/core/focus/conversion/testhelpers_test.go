package conversion

import (
	"context"

	"github.com/costscope/costscope/internal/core/focus/types"
)

// fakeWriter is a minimal in-memory DataWriter for tests.
type fakeWriter struct{ recs []types.FocusRecord }

func (w *fakeWriter) Open(ctx context.Context, path string, schema *types.FocusSchema) error {
	return nil
}
func (w *fakeWriter) Write(ctx context.Context, records []types.FocusRecord) error {
	w.recs = append(w.recs, records...)
	return nil
}
func (w *fakeWriter) WriteChunk(ctx context.Context, data []byte) error { return nil }
func (w *fakeWriter) Flush(ctx context.Context) error                   { return nil }
func (w *fakeWriter) Close() error                                      { return nil }
func (w *fakeWriter) GetMetadata() *types.DataDestinationMetadata {
	return &types.DataDestinationMetadata{}
}
