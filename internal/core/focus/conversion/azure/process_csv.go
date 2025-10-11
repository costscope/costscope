package azure

import (
	"context"
	"io"
	"time"

	"local/costscope/internal/core/focus/types"
)

// ProcessCSV streams CSV records in chunks, maps them via the provided mapper, writes using dw,
// and updates progress via the callback. It mirrors the legacy root implementation's behavior
// but lives in the provider package to maintain separation without changing public APIs.
//
// Parameters:
// - reader: the CSV source (already opened and decompressed if needed)
// - config: conversion configuration (chunk size, etc.)
// - dw: data writer for output
// - progress: progress struct to mutate
// - start: conversion start time
// - cb: optional progress callback
// - pathLabel: "legacy" or "unified" for telemetry labels (determined by caller)
// - mapper: function(headers, chunk) -> (mappedRecords, errorsCount)
//
// Returns: (recordCount, processedRecords, errorRecords, error)
func ProcessCSV(
	ctx context.Context,
	reader io.Reader,
	config *types.ConversionConfig,
	dw types.DataWriter,
	progress *types.ConversionProgress,
	start time.Time,
	cb types.ProgressCallback,
	pathLabel string,
	mapper func([]string, [][]string) ([]types.FocusRecord, int),
) (int64, int64, int64, error) {
	src, headers, err := NewCSVRowSourceFromReader(reader)
	if err != nil {
		return 0, 0, 0, err
	}
	defer func() { _ = src.Close() }()

	// Copy headers for safety
	headersCopy := make([]string, len(headers))
	copy(headersCopy, headers)

	var recordCount, processedRecords, errorRecords int64
	chunkSize := config.ChunkSize
	if chunkSize <= 0 {
		// Fallback to a sane default mirroring legacy behavior when unset
		chunkSize = 10000
	}

	for {
		select {
		case <-ctx.Done():
			return 0, 0, 0, ctx.Err()
		default:
		}

		chunk := make([][]string, 0, chunkSize)
		for i := 0; i < chunkSize; i++ {
			rec, err := src.Next(ctx)
			if err == io.EOF {
				break
			}
			if err != nil {
				// Mirror legacy behavior: count read error and continue
				// (logger lives in root; avoid cycles here)
				errorRecords++
				continue
			}
			chunk = append(chunk, rec)
			recordCount++
		}

		if len(chunk) == 0 {
			break
		}

		// Map, emit metrics, and write this chunk using the provider-scoped helper
		mapped, errs, werr := MapAndWriteCSVChunk(ctx, dw, headersCopy, chunk, config.InputPath, pathLabel, mapper)
		if werr != nil {
			return 0, 0, 0, werr
		}
		processedRecords += int64(mapped)
		errorRecords += int64(errs)

		// Progress updates via provider-scoped helper
		UpdateProgress(cb, progress, recordCount, processedRecords, errorRecords, start)
	}

	return recordCount, processedRecords, errorRecords, nil
}
