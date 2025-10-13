package common

import (
	"fmt"
	"os"
	"time"

	"github.com/costscope/costscope/internal/core/focus/types"
)

// EstimateFromFile returns a heuristic conversion estimate based on input file size and config.
func EstimateFromFile(inputPath string, cfg *types.ConversionConfig) (*types.ConversionEstimate, error) {
	fi, err := os.Stat(inputPath)
	if err != nil {
		return nil, fmt.Errorf("stat failed: %w", err)
	}
	sizeMB := float64(fi.Size()) / (1024 * 1024)
	estRecords := int64(sizeMB * 1200)
	estDuration := time.Duration(sizeMB/12) * time.Second
	estMem := 128
	if cfg.Streaming {
		estMem = cfg.ChunkSize / 128
		if estMem < 64 {
			estMem = 64
		}
	}
	return &types.ConversionEstimate{
		EstimatedDuration:     estDuration,
		EstimatedMemoryMB:     estMem,
		EstimatedOutputSizeMB: sizeMB * 0.75,
		EstimatedRecords:      estRecords,
		RecommendedChunkSize:  10000,
		RecommendedWorkers:    4,
	}, nil
}

// CommonSupportedFormats is a shared set of formats used by Azure and GCP paths.
var CommonSupportedFormats = &types.SupportedFormats{
	InputFormats:  []string{"csv", "json"},
	OutputFormats: []string{"parquet", "json"},
}
