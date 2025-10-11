package telemetry

import (
	"context"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// InitTracing initializes a basic tracer provider with resource info.
// If OTEL_EXPORTER_OTLP_ENDPOINT is set, an external exporter may be configured
// by the runtime (not handled here to keep dependencies minimal).
func InitTracing(ctx context.Context) (shutdown func(context.Context) error, err error) {
	svc := os.Getenv("OTEL_SERVICE_NAME")
	if svc == "" {
		svc = "costscope"
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(svc),
		)),
	)
	otel.SetTracerProvider(tp)
	return tp.Shutdown, nil
}
