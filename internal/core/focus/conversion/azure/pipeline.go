package azure

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"

	"local/costscope/internal/core/focus/types"
	"local/costscope/internal/core/monitoring/telemetry"
)

// MapAndWriteCSVChunk maps a CSV chunk to FOCUS via the provided mapper, emits mapping metrics,
// sets source filename, and writes via the DataWriter. Returns processed records, mapping errors,
// and a terminal error if the write fails.
//
// pathLabel should be "legacy" or "unified" (resolved in the root package). Provider label is "azure".
func MapAndWriteCSVChunk(
	ctx context.Context,
	dw types.DataWriter,
	headers []string,
	chunk [][]string,
	inputPath string,
	pathLabel string,
	mapper func([]string, [][]string) ([]types.FocusRecord, int),
) (int, int, error) {
	// Mapping span & latency metrics
	mapStart := time.Now()
	mCtx, mapSpan := otel.Tracer("costscope.converter").Start(ctx, "mapping")
	mapSpan.SetAttributes(attribute.Int("chunk.size", len(chunk)))
	focusRecords, errs := mapper(headers, chunk)
	mapSpan.SetAttributes(
		attribute.Int("mapped.records", len(focusRecords)),
		attribute.Int("mapping.errors", errs),
	)
	mapSpan.End()

	telemetry.MapperLatency.WithLabelValues("azure", pathLabel).Observe(time.Since(mapStart).Seconds())

	// Per-record post mapping fields + classifier metrics
	for i := range focusRecords {
		focusRecords[i].SourceFileName = filepath.Base(inputPath)
		telemetry.ClassifierDecisions.WithLabelValues("azure", pathLabel, focusRecords[i].ChargeCategory).Inc()
	}
	if len(focusRecords) > 0 {
		telemetry.MapperRowsTotal.WithLabelValues("azure", pathLabel).Add(float64(len(focusRecords)))
	}

	if len(focusRecords) > 0 {
		if err := dw.Write(mCtx, focusRecords); err != nil {
			return 0, 0, fmt.Errorf("write failed: %w", err)
		}
	}
	return len(focusRecords), errs, nil
}
