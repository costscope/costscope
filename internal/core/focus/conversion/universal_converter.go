package conversion

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	awsp "github.com/costscope/costscope/internal/core/focus/conversion/aws"
	azp "github.com/costscope/costscope/internal/core/focus/conversion/azure"
	gcpp "github.com/costscope/costscope/internal/core/focus/conversion/gcp"
	store "github.com/costscope/costscope/internal/core/focus/conversion/store"
	u "github.com/costscope/costscope/internal/core/focus/conversion/universal"
	"github.com/costscope/costscope/internal/core/focus/types"
	"github.com/costscope/costscope/internal/core/logging"
	"github.com/costscope/costscope/internal/core/monitoring/telemetry"
)

// =====================================================================================
// Universal FOCUS Converter - Multi-Cloud Billing Data to FOCUS v1.2
// =====================================================================================
// Central conversion engine supporting AWS CUR, Azure Cost Export, GCP Billing Export

// Backward-compatible alias to the new implementation in subpackage
type UniversalConverter = u.UniversalConverter

// Constructor forwarder
func NewUniversalConverter() *UniversalConverter { return u.NewUniversalConverter() }

// =====================================================================================
// Conversion Manager - Orchestrates Multiple Conversions
// =====================================================================================

// ConversionManager manages and coordinates multiple conversion operations
type ConversionManager struct {
	converter         *UniversalConverter
	activeJobs        map[string]*ConversionJob
	jobHistory        []*types.ConversionResult // retained for backward compatibility (deprecated)
	jobStore          store.JobStore
	logger            *logging.Logger
	mutex             sync.RWMutex
	maxConcurrentJobs int
}

// ConversionJob represents an active conversion job
type ConversionJob struct {
	ID         string                    `json:"id"`
	Config     *types.ConversionConfig   `json:"config"`
	Status     types.ConversionStatus    `json:"status"`
	Progress   *types.ConversionProgress `json:"progress"`
	Result     *types.ConversionResult   `json:"result,omitempty"`
	StartTime  time.Time                 `json:"start_time"`
	EndTime    *time.Time                `json:"end_time,omitempty"`
	CancelFunc context.CancelFunc        `json:"-"`
}

// NewConversionManager creates a new conversion manager
func NewConversionManager(maxConcurrentJobs int, opts ...interface{}) *ConversionManager { // varargs retained for backward compatibility; optional JobStore may be passed
	manager := &ConversionManager{
		converter:         NewUniversalConverter(),
		activeJobs:        make(map[string]*ConversionJob),
		jobHistory:        make([]*types.ConversionResult, 0),
		logger:            logging.NewLogger(logging.LevelInfo),
		maxConcurrentJobs: maxConcurrentJobs,
		jobStore:          nil,
	}
	// accept optional JobStore as first variadic argument
	if len(opts) > 0 {
		if js, ok := opts[0].(store.JobStore); ok {
			manager.jobStore = js
		}
	}
	manager.registerDefaultConverters()
	return manager
}

// NewConfiguredConversionManager creates a ConversionManager using environment
// configuration: if JOB_STORE_PATH is set, a BoltJobStore is created and used.
func NewConfiguredConversionManager(maxConcurrentJobs int) *ConversionManager {
	// read env var JOB_STORE_PATH
	path := os.Getenv("JOB_STORE_PATH")
	if path == "" {
		return NewConversionManager(maxConcurrentJobs)
	}
	if path == "" {
		return NewConversionManager(maxConcurrentJobs)
	}
	// try to create BoltJobStore; if fails, fall back to in-memory but log
	js, err := store.NewBoltJobStore(path)
	if err != nil {
		logging.GetLogger().WarnWithFields("failed to open BoltJobStore, falling back to in-memory", map[string]interface{}{"path": path, "error": err.Error()})
		return NewConversionManager(maxConcurrentJobs)
	}
	cm := NewConversionManager(maxConcurrentJobs, js)
	return cm
}

// GetConverter returns the universal converter
// (no env helper required; os.Getenv is used in the factory)
func (cm *ConversionManager) GetConverter() *UniversalConverter {
	return cm.converter
}

// registerDefaultConverters registers default provider converters
func (cm *ConversionManager) registerDefaultConverters() {
	// Register AWS converter
	awsConverter := awsp.NewAWSConverter()
	_ = cm.converter.RegisterConverter("aws", awsConverter)

	// Register Azure converter
	azureConverter := azp.NewAzureConverter()
	_ = cm.converter.RegisterConverter("azure", azureConverter)

	// Register GCP converter
	gcpConverter := gcpp.NewGCPConverter()
	_ = cm.converter.RegisterConverter("gcp", gcpConverter)
}

// SubmitJob submits a new conversion job
func (cm *ConversionManager) SubmitJob(config *types.ConversionConfig) (string, error) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	// Check concurrent job limit
	if len(cm.activeJobs) >= cm.maxConcurrentJobs {
		return "", fmt.Errorf("maximum concurrent jobs reached (%d)", cm.maxConcurrentJobs)
	}

	// Create job
	jobId := fmt.Sprintf("job_%d", time.Now().UnixNano())
	ctx, cancelFunc := context.WithCancel(context.Background())

	job := &ConversionJob{
		ID:         jobId,
		Config:     config,
		Status:     types.StatusPending,
		StartTime:  time.Now(),
		CancelFunc: cancelFunc,
		Progress: &types.ConversionProgress{
			ConversionId: jobId,
			Status:       string(types.StatusPending),
			StartTime:    time.Now(),
		},
	}

	cm.activeJobs[jobId] = job
	// Metrics: increment submitted & active
	telemetry.ConversionJobsSubmitted.Inc()
	telemetry.ConversionActiveJobs.Inc()

	// Start job asynchronously
	go cm.executeJob(ctx, job)

	cm.logger.Info(fmt.Sprintf("Submitted conversion job: %s", jobId))
	return jobId, nil
}

// GetJobStatus returns the status of a conversion job
func (cm *ConversionManager) GetJobStatus(jobId string) (*ConversionJob, error) {
	cm.mutex.RLock()
	job, exists := cm.activeJobs[jobId]
	if !exists {
		cm.mutex.RUnlock()
		return nil, fmt.Errorf("job not found: %s", jobId)
	}
	// Create a snapshot copy under the read lock to avoid exposing internal mutable state
	snap := snapshotJob(job)
	cm.mutex.RUnlock()
	return snap, nil
}

// CancelJob cancels a running conversion job
func (cm *ConversionManager) CancelJob(jobId string) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	job, exists := cm.activeJobs[jobId]
	if !exists {
		return fmt.Errorf("job not found: %s", jobId)
	}

	if job.Status == types.StatusCompleted || job.Status == types.StatusFailed {
		return fmt.Errorf("job already finished: %s", string(job.Status))
	}

	job.CancelFunc()
	// Mark status; metrics & active gauge will be handled in executeJob when it observes ctx cancellation.
	job.Status = types.StatusCancelled
	cm.logger.Info(fmt.Sprintf("Cancellation requested for conversion job: %s", jobId))
	return nil
}

// ListActiveJobs returns all active conversion jobs
func (cm *ConversionManager) ListActiveJobs() []*ConversionJob {
	cm.mutex.RLock()
	jobs := make([]*ConversionJob, 0, len(cm.activeJobs))
	for _, job := range cm.activeJobs {
		jobs = append(jobs, snapshotJob(job))
	}
	cm.mutex.RUnlock()
	return jobs
}

// GetJobHistory returns conversion job history
func (cm *ConversionManager) GetJobHistory(limit int) []*types.ConversionResult {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	if limit <= 0 || limit > len(cm.jobHistory) {
		return cm.jobHistory
	}
	return cm.jobHistory[len(cm.jobHistory)-limit:]
}

// RemoveCompletedJob removes a completed job from active jobs
func (cm *ConversionManager) RemoveCompletedJob(jobId string) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	delete(cm.activeJobs, jobId)
}

// executeJob executes a conversion job
func (cm *ConversionManager) executeJob(ctx context.Context, job *ConversionJob) {
	// Don't remove from active jobs immediately - let monitoring handle it
	defer func() {
		// Job will be removed after monitoring completes
	}()

	// Update job status unless already cancelled before start
	cm.mutex.Lock()
	if job.Status != types.StatusCancelled {
		job.Status = types.StatusRunning
		job.Progress.Status = string(types.StatusRunning)
	}
	cm.mutex.Unlock()

	cm.logger.Info(fmt.Sprintf("Starting conversion job: %s", job.ID))

	// Execute conversion
	result, err := cm.converter.Convert(ctx, job.Config)
	now := time.Now()

	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	job.EndTime = &now
	if err != nil {
		if ctx.Err() != nil { // cancellation path
			job.Status = types.StatusCancelled
			job.Progress.Status = string(types.StatusCancelled)
			telemetry.ConversionJobsCompleted.WithLabelValues("cancelled").Inc()
			telemetry.ConversionActiveJobs.Dec()
			cm.logger.Info(fmt.Sprintf("Conversion job cancelled: %s", job.ID))
			// persist cancelled result placeholder
			if job.Result == nil {
				job.Result = &types.ConversionResult{ConversionId: job.ID, Success: false, StartTime: job.StartTime}
			}
			go cm.persistResultToStore(job.Result)
			return
		}
		cm.logger.Error(fmt.Sprintf("Conversion job failed: %s - %v", job.ID, err))
		job.Status = types.StatusFailed
		job.Progress.Status = string(types.StatusFailed)
		job.Progress.LastError = err.Error()
		telemetry.ConversionJobsCompleted.WithLabelValues("failed").Inc()
		telemetry.ConversionActiveJobs.Dec()
		if job.Result == nil {
			job.Result = &types.ConversionResult{ConversionId: job.ID, Success: false, StartTime: job.StartTime}
		}
		go cm.persistResultToStore(job.Result)
		return
	}

	// Job completed successfully
	job.Status = types.StatusCompleted
	job.Progress.Status = string(types.StatusCompleted)
	job.Result = result
	// Legacy in-memory history (will be removed in future major version)
	cm.jobHistory = append(cm.jobHistory, result)
	// persist to configured JobStore if present
	go cm.persistResultToStore(result)

	cm.logger.Info(fmt.Sprintf("Conversion job completed: %s", job.ID))
	telemetry.ConversionJobsCompleted.WithLabelValues("success").Inc()
	telemetry.ConversionActiveJobs.Dec()
	if job.EndTime != nil {
		telemetry.ConversionJobDuration.Observe(job.EndTime.Sub(job.StartTime).Seconds())
	}
}

// persistResultToStore saves the ConversionResult to the configured JobStore if one exists.
func (cm *ConversionManager) persistResultToStore(res *types.ConversionResult) {
	if cm == nil || cm.jobStore == nil || res == nil {
		return
	}
	// Synchronous persistence with one retry on error.
	err := cm.jobStore.SaveResult(res)
	if err != nil {
		// retry once
		cm.logger.Warn(fmt.Sprintf("job store save failed for %s, retrying: %v", res.ConversionId, err))
		err = cm.jobStore.SaveResult(res)
	}
	if err != nil {
		telemetry.ConversionPersistenceFailure.Inc()
		cm.logger.Error(fmt.Sprintf("failed to persist conversion result %s: %v", res.ConversionId, err))
		return
	}
	telemetry.ConversionPersistenceSuccess.Inc()
}

// snapshotJob performs a shallow copy of the ConversionJob struct plus deep copies of
// pointer fields that are mutated during execution (Progress, EndTime) so that callers
// can read job state without racing with the executing goroutine.
func snapshotJob(j *ConversionJob) *ConversionJob {
	if j == nil {
		return nil
	}
	cp := *j // shallow copy (safe for read-only fields)
	if j.Progress != nil {
		p := *j.Progress
		cp.Progress = &p
	}
	if j.EndTime != nil { // copy value to avoid data race if writer updates pointer target
		t := *j.EndTime
		cp.EndTime = &t
	}
	return &cp
}

// Close gracefully closes any underlying resources used by the ConversionManager (e.g., BoltJobStore).
func (cm *ConversionManager) Close() error {
	if cm == nil || cm.jobStore == nil {
		return nil
	}
	type closer interface{ Close() error }
	if c, ok := cm.jobStore.(closer); ok {
		return c.Close()
	}
	return nil
}
