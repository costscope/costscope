package writers

import (
	"context"
	"path/filepath"
	"strings"

	"local/costscope/internal/core/focus/types"
)

// parquetOptionsKey is a private context key for passing ParquetOptions
type parquetOptionsKey struct{}

// NewWriter returns a DataWriter based on output format or file extension, and a normalized result output format string.
// It opens the writer before returning; caller should defer Close and call Flush when done.
func NewWriter(ctx context.Context, outputPath string, outputFormat string, schema *types.FocusSchema) (types.DataWriter, string, error) {
	outExt := strings.ToLower(filepath.Ext(outputPath))
	useParquet := strings.EqualFold(outputFormat, "parquet") || outExt == ".parquet"
	if useParquet {
		// Extract parquet options from context if present
		var opts *types.ParquetOptions
		if v := ctx.Value(parquetOptionsKey{}); v != nil {
			if po, ok := v.(*types.ParquetOptions); ok {
				opts = po
			}
		}
		pw := &ParquetWriter{Options: opts}
		if err := pw.Open(ctx, outputPath, schema); err != nil {
			return nil, "", err
		}
		return pw, "FOCUS_PARQUET", nil
	}
	nw := &NDJSONWriter{}
	if err := nw.Open(ctx, outputPath, schema); err != nil {
		return nil, "", err
	}
	return nw, "FOCUS_NDJSON", nil
}

// WithParquetOptions attaches ParquetOptions to context for consumption by NewWriter/ParquetWriter.
func WithParquetOptions(ctx context.Context, opts *types.ParquetOptions) context.Context {
	if opts == nil {
		return ctx
	}
	return context.WithValue(ctx, parquetOptionsKey{}, opts)
}
