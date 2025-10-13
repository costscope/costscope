package gcp

import (
	"context"
	"io"
	"path/filepath"
	"time"

	rgcp "github.com/costscope/costscope/internal/core/focus/reader/gcp"
	"github.com/costscope/costscope/internal/core/focus/types"
	"github.com/costscope/costscope/internal/core/monitoring/telemetry"
)

// ProcessCSV streams GCP CSV records in chunks, maps them via the provided mapper,
// writes them using the given DataWriter, and reports progress. It mirrors the
// legacy root implementation semantics (classifier metrics are emitted by the mapper).
//
// Parameters:
// - reader: CSV source (already decompressed if needed)
// - config: conversion configuration
// - dw: data writer for output
// - progress: progress struct to mutate
// - start: conversion start time
// - cb: optional progress callback
// - pathLabel: telemetry label ("legacy" or "unified") for mapper latency metric
// - mapper: function(headers, chunk) -> (mappedRecords, errorsCount)
func ProcessCSV(
	ctx context.Context,
	reader io.Reader,
	config *types.ConversionConfig,
	dw types.DataWriter,
	progress *types.ConversionProgress,
	start time.Time,
	cb types.ProgressCallback,
	pathLabel string,
	mapper func(headers []string, rows [][]string) ([]types.FocusRecord, int),
) (int64, int64, int64, error) {
	cr := rgcp.NewCSVReader(reader)
	headers, err := cr.ReadHeaders()
	if err != nil {
		return 0, 0, 0, err
	}

	var rc, pc, ec int64
	chunkSize := config.ChunkSize
	if chunkSize <= 0 {
		chunkSize = 10000
	}
	baseName := filepath.Base(config.InputPath)

	for {
		chunk, read, rerr := cr.ReadChunk(chunkSize)
		rc += read
		if rerr != nil && len(chunk) == 0 {
			if rerr == io.EOF {
				break
			}
			return rc, pc, ec, rerr
		}
		if len(chunk) == 0 {
			break
		}

		mapStart := time.Now()
		mapped, errs := mapper(headers, chunk)
		telemetry.MapperLatency.WithLabelValues("gcp", pathLabel).Observe(time.Since(mapStart).Seconds())
		ec += int64(errs)

		// Set source filename parity (legacy behavior)
		if len(mapped) > 0 {
			for i := range mapped {
				mapped[i].SourceFileName = baseName
			}
			if err := dw.Write(ctx, mapped); err != nil {
				return rc, pc, ec, err
			}
			pc += int64(len(mapped))
		}

		if cb != nil {
			progress.TotalRecords = rc
			progress.ProcessedRecords = pc
			progress.ErrorRecords = ec
			progress.ElapsedTime = time.Since(start)
			cb(progress)
		}

		if rerr != nil { // after processing current chunk, exit loop
			break
		}
	}
	return rc, pc, ec, nil
}
