package azure

import (
	"time"

	"github.com/costscope/costscope/internal/core/focus/types"
)

// UpdateProgress updates and publishes progress via callback (if provided).
// This is provider-scoped to avoid duplication while keeping behavior identical.
func UpdateProgress(cb types.ProgressCallback, progress *types.ConversionProgress, recordCount, processedRecords, errorRecords int64, start time.Time) {
	if cb == nil || progress == nil {
		return
	}
	progress.TotalRecords = recordCount
	progress.ProcessedRecords = processedRecords
	progress.ErrorRecords = errorRecords
	progress.ElapsedTime = time.Since(start)
	cb(progress)
}
