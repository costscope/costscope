package writers

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/costscope/costscope/internal/core/focus/types"
)

// NDJSONWriter implements types.DataWriter to write one JSON record per line
type NDJSONWriter struct {
	file     *os.File
	enc      *json.Encoder
	metadata *types.DataDestinationMetadata
}

func (w *NDJSONWriter) Open(_ context.Context, path string, _ *types.FocusSchema) error {
	// #nosec G304 - path comes from validated config
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	w.file = f
	w.enc = json.NewEncoder(f)
	w.metadata = &types.DataDestinationMetadata{
		FilePath: path,
		Created:  time.Now(),
		Format:   "ndjson",
		Schema:   "FOCUS_1.2",
	}
	return nil
}

func (w *NDJSONWriter) Write(_ context.Context, records []types.FocusRecord) error {
	for i := range records {
		if err := w.enc.Encode(records[i]); err != nil {
			return fmt.Errorf("failed to encode record: %w", err)
		}
		w.metadata.RecordCount++
	}
	return nil
}

func (w *NDJSONWriter) WriteChunk(_ context.Context, data []byte) error {
	if w.file == nil {
		return fmt.Errorf("writer not opened")
	}
	if _, err := w.file.Write(data); err != nil {
		return fmt.Errorf("failed to write chunk: %w", err)
	}
	return nil
}

func (w *NDJSONWriter) Flush(_ context.Context) error {
	if w.file != nil {
		return w.file.Sync()
	}
	return nil
}

func (w *NDJSONWriter) Close() error {
	if w.file != nil {
		if stat, err := w.file.Stat(); err == nil {
			w.metadata.FileSize = stat.Size()
		}
		return w.file.Close()
	}
	return nil
}

func (w *NDJSONWriter) GetMetadata() *types.DataDestinationMetadata {
	return w.metadata
}
