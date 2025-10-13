//go:build never

package conversion

import (
	"context"
	"fmt"
	"sync"

	u "github.com/costscope/costscope/internal/core/focus/conversion/universal"
	"github.com/costscope/costscope/internal/core/focus/types"
	"github.com/costscope/costscope/internal/core/logging"
)

// Re-export to preserve backward-compatible imports
type UniversalConverter = u.UniversalConverter

func NewUniversalConverter() *UniversalConverter { return u.NewUniversalConverter() }

// ConversionManager manages and coordinates multiple conversion operations
type ConversionManager struct {
	converter         *UniversalConverter
	activeJobs        map[string]*ConversionJob
	jobHistory        []*types.ConversionResult
	logger            *logging.Logger
	mutex             sync.RWMutex
	maxConcurrentJobs int
}

// ConversionJob represents an active conversion job
type ConversionJob struct {
	ID       string                    `json:"id"`
	Config   *types.ConversionConfig   `json:"config"`
	Status   types.ConversionStatus    `json:"status"`
	Progress *types.ConversionProgress `json:"progress"`
	Result   *types.ConversionResult   `json:"result,omitempty"`
}

// NewConversionManager creates a new conversion manager
func NewConversionManager(maxConcurrentJobs int) *ConversionManager {
	manager := &ConversionManager{
		converter:         NewUniversalConverter(),
		activeJobs:        make(map[string]*ConversionJob),
		jobHistory:        make([]*types.ConversionResult, 0),
		logger:            logging.NewLogger(logging.LevelInfo),
		maxConcurrentJobs: maxConcurrentJobs,
	}
	return manager
}

// Minimal forwarders for compatibility (methods used externally remain available via the embedded converter)

func (cm *ConversionManager) GetConverter() *UniversalConverter { return cm.converter }

// SubmitJob is preserved for API compatibility; implementation delegated to universal package would be planned in future split
func (cm *ConversionManager) SubmitJob(config *types.ConversionConfig) (string, error) {
	// Keep behavior minimal here; the full job orchestration lived in the old file.
	if config == nil {
		return "", fmt.Errorf("config cannot be nil")
	}
	// Generate a simple job id and run synchronously to preserve surface; callers expecting async should use API/job manager.
	jobId := fmt.Sprintf("job_%s", config.ConversionId)
	// Synchronous conversion
	_, err := cm.converter.Convert(context.Background(), config)
	if err != nil {
		return "", err
	}
	return jobId, nil
}
