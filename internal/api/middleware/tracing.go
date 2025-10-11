package middleware

import (
	"strings"
	"time"

	"local/costscope/internal/core/logging"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
)

const (
	HeaderRequestID = "X-Request-ID"
	HeaderTraceID   = "X-Trace-ID"
	HeaderSpanID    = "X-Span-ID"
)

// Tracing sets up request/trace/span IDs and context propagation.
func Tracing(tracerName string) gin.HandlerFunc {
	propagator := propagation.TraceContext{}
	tracer := otel.Tracer(tracerName)

	return func(c *gin.Context) {
		// Extract context from incoming headers
		ctx := propagator.Extract(c.Request.Context(), propagation.HeaderCarrier(c.Request.Header))

		// Ensure request ID
		reqID := c.GetHeader(HeaderRequestID)
		if strings.TrimSpace(reqID) == "" {
			reqID = uuid.NewString()
		}
		c.Header(HeaderRequestID, reqID)

		// Start span
		ctx, span := tracer.Start(ctx, c.FullPath())
		defer span.End()

		// Get trace/span IDs
		spanCtx := span.SpanContext()
		traceID := spanCtx.TraceID().String()
		spanID := spanCtx.SpanID().String()

		// Add headers for downstream
		c.Header(HeaderTraceID, traceID)
		c.Header(HeaderSpanID, spanID)

		// Attach attributes
		span.SetAttributes(
			attribute.String("http.method", c.Request.Method),
			attribute.String("http.route", c.FullPath()),
			attribute.String("http.target", c.Request.URL.Path),
			attribute.String("request.id", reqID),
		)

		// Put into gin context and request context
		c.Set("request_id", reqID)
		c.Set("trace_id", traceID)
		c.Set("span_id", spanID)
		// Use typed keys in context to avoid key collisions
		ctx = logging.ContextWithIDs(ctx, reqID, traceID, spanID)
		c.Request = c.Request.WithContext(ctx)

		start := time.Now()
		c.Next()

		// Record status code and duration
		span.SetAttributes(
			attribute.Int("http.status_code", c.Writer.Status()),
			attribute.Int64("http.duration_ms", time.Since(start).Milliseconds()),
		)
	}
}
