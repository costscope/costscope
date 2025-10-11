package reports

import "context"

// includeContentKey is used to gate whether ExportReport should prefer full content
// over metadata when exporting. Exposed via helper to avoid leaking key type.
type includeContentKey struct{}

// WithIncludeContent returns a context that signals the exporter to include full
// report content when available instead of just metadata.
func WithIncludeContent(ctx context.Context, include bool) context.Context {
	return context.WithValue(ctx, includeContentKey{}, include)
}

func includeContentFromContext(ctx context.Context) bool {
	if ctx == nil {
		return false
	}
	v, _ := ctx.Value(includeContentKey{}).(bool)
	return v
}
