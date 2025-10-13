//go:build enterprise

package streaming

import (
	"context"
	"fmt"
	"sync"
	"time"

	streamingTypes "github.com/costscope/costscope/cmd/modules/streaming/types"
	"github.com/costscope/costscope/internal/core/logging"
	"github.com/costscope/costscope/internal/providers"
	providerTypes "github.com/costscope/costscope/internal/providers/types"
)

// Job status constants
const (
	JobStatusStarting = "starting"
	JobStatusRunning  = "running"
	JobStatusPaused   = "paused"
	JobStatusStopped  = "stopped"
)

// JobManager interface defines methods for managing streaming jobs
type JobManager interface {
	StartJob(config *streamingTypes.StreamingJobConfig) (*streamingTypes.StreamingJobInfo, error)
	PauseJob(jobID string) error
	ResumeJob(jobID string) error
	StopJob(jobID string) error
	GetJobStatus(jobID string) (*streamingTypes.StreamingJobStatus, error)
	ListJobs() ([]*streamingTypes.StreamingJobInfo, error)
	Shutdown() error
}

// DefaultJobManager implements JobManager interface
type DefaultJobManager struct {
	jobs          map[string]*Job
	mu            sync.RWMutex
	providerMgr   *providers.ProviderManager
	logger        *logging.Logger
	ctx           context.Context
	cancel        context.CancelFunc
	checkpointDir string
}

// Job represents a streaming job with its execution context
type Job struct {
	Config  *streamingTypes.StreamingJobConfig
	Status  *streamingTypes.StreamingJobStatus
	Metrics *streamingTypes.StreamingJobMetrics
	ctx     context.Context
	cancel  context.CancelFunc
	worker  *JobWorker
	mu      sync.RWMutex
	// TenantID optional multi-tenant identifier (skeleton; unused when feature flag disabled)
	TenantID string
}

// JobWorker handles the actual execution of a streaming job
type JobWorker struct {
	job        *Job
	provider   providerTypes.CloudProvider
	logger     *logging.Logger
	checkpoint *Checkpoint
}

// Checkpoint represents a job checkpoint for recovery
type Checkpoint struct {
	JobID          string                 `json:"job_id"`
	ProcessedRows  int64                  `json:"processed_rows"`
	ProcessedBytes int64                  `json:"processed_bytes"`
	LastPosition   string                 `json:"last_position"`
	Timestamp      time.Time              `json:"timestamp"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// NewJobManager creates a new job manager instance
func NewJobManager(providerMgr *providers.ProviderManager, checkpointDir string) *DefaultJobManager {
	ctx, cancel := context.WithCancel(context.Background())

	return &DefaultJobManager{
		jobs:          make(map[string]*Job),
		providerMgr:   providerMgr,
		logger:        logging.NewLogger(logging.LevelInfo),
		ctx:           ctx,
		cancel:        cancel,
		checkpointDir: checkpointDir,
	}
}

// StartJob starts a new streaming job
func (jm *DefaultJobManager) StartJob(config *streamingTypes.StreamingJobConfig) (*streamingTypes.StreamingJobInfo, error) {
	jm.mu.Lock()
	defer jm.mu.Unlock()

	// Validate configuration
	if err := jm.validateJobConfig(config); err != nil {
		return nil, fmt.Errorf("invalid job configuration: %w", err)
	}

	// Check if job already exists
	if _, exists := jm.jobs[config.JobID]; exists {
		return nil, fmt.Errorf("job with ID %s already exists", config.JobID)
	}

	// Get provider
	provider, err := jm.providerMgr.GetProvider(config.Provider)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider %s: %w", config.Provider, err)
	}

	// Validate provider credentials
	ctx := context.Background()
	if err := provider.ValidateCredentials(ctx, map[string]string{}); err != nil {
		return nil, fmt.Errorf("provider credentials validation failed: %w", err)
	}

	// Create job context
	jobCtx, jobCancel := context.WithCancel(jm.ctx)

	// Initialize job status
	status := &streamingTypes.StreamingJobStatus{
		JobID:      config.JobID,
		Status:     JobStatusStarting,
		Progress:   0.0,
		StartTime:  time.Now(),
		LastUpdate: time.Now(),
	}

	// Initialize job metrics
	metrics := &streamingTypes.StreamingJobMetrics{
		JobID:          config.JobID,
		CPUUsage:       0.0,
		MemoryUsage:    0,
		ProcessingRate: 0.0,
		LastUpdate:     time.Now(),
	}

	// Create job worker
	worker := &JobWorker{
		provider: provider,
		logger:   jm.logger,
		checkpoint: &Checkpoint{
			JobID:     config.JobID,
			Timestamp: time.Now(),
			Metadata:  make(map[string]interface{}),
		},
	}

	// Create job
	job := &Job{
		Config:  config,
		Status:  status,
		Metrics: metrics,
		ctx:     jobCtx,
		cancel:  jobCancel,
		worker:  worker,
	}

	// Set worker reference to job
	worker.job = job

	// Store job
	jm.jobs[config.JobID] = job

	// Start job execution in background
	go jm.executeJob(job)

	jm.logger.Info(fmt.Sprintf("Started streaming job: %s", config.JobID))

	// Get status copy after job started (with proper locking)
	job.mu.RLock()
	statusCopy := *job.Status
	job.mu.RUnlock()

	return &streamingTypes.StreamingJobInfo{
		Config: *config,
		Status: statusCopy,
	}, nil
}

// PauseJob pauses a running job
func (jm *DefaultJobManager) PauseJob(jobID string) error {
	// First, get a reference to the job without holding the write lock for the entire duration
	jm.mu.RLock()
	job, exists := jm.jobs[jobID]
	jm.mu.RUnlock()
	if !exists {
		return fmt.Errorf("job %s not found", jobID)
	}

	// If the job is still starting, wait briefly for it to transition to running
	// to avoid racy failures when pausing immediately after StartJob in tests.
	deadline := time.Now().Add(500 * time.Millisecond)
	for {
		job.mu.RLock()
		status := job.Status.Status
		job.mu.RUnlock()
		if status != JobStatusStarting || time.Now().After(deadline) {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	// Now take locks to perform the state change atomically
	jm.mu.Lock()
	defer jm.mu.Unlock()

	// Job might have been removed between the waits; re-check
	job, exists = jm.jobs[jobID]
	if !exists {
		return fmt.Errorf("job %s not found", jobID)
	}

	job.mu.Lock()
	defer job.mu.Unlock()

	if job.Status.Status != JobStatusRunning {
		return fmt.Errorf("job %s is not running (current status: %s)", jobID, job.Status.Status)
	}

	job.Status.Status = JobStatusPaused
	job.Status.LastUpdate = time.Now()

	jm.logger.Info(fmt.Sprintf("Paused streaming job: %s", jobID))
	return nil
}

// ResumeJob resumes a paused job
func (jm *DefaultJobManager) ResumeJob(jobID string) error {
	jm.mu.Lock()
	defer jm.mu.Unlock()

	job, exists := jm.jobs[jobID]
	if !exists {
		return fmt.Errorf("job %s not found", jobID)
	}

	job.mu.Lock()
	defer job.mu.Unlock()

	if job.Status.Status != JobStatusPaused {
		return fmt.Errorf("job %s is not paused (current status: %s)", jobID, job.Status.Status)
	}

	job.Status.Status = JobStatusRunning
	job.Status.LastUpdate = time.Now()

	jm.logger.Info(fmt.Sprintf("Resumed streaming job: %s", jobID))
	return nil
}

// StopJob stops a job
func (jm *DefaultJobManager) StopJob(jobID string) error {
	jm.mu.Lock()
	defer jm.mu.Unlock()

	job, exists := jm.jobs[jobID]
	if !exists {
		return fmt.Errorf("job %s not found", jobID)
	}

	job.mu.Lock()
	defer job.mu.Unlock()

	// Cancel job context
	job.cancel()

	job.Status.Status = JobStatusStopped
	job.Status.LastUpdate = time.Now()

	jm.logger.Info(fmt.Sprintf("Stopped streaming job: %s", jobID))
	return nil
}

// GetJobStatus returns the status of a specific job
func (jm *DefaultJobManager) GetJobStatus(jobID string) (*streamingTypes.StreamingJobStatus, error) {
	jm.mu.RLock()
	defer jm.mu.RUnlock()

	job, exists := jm.jobs[jobID]
	if !exists {
		return nil, fmt.Errorf("job %s not found", jobID)
	}

	job.mu.RLock()
	defer job.mu.RUnlock()

	// Create a copy of the status
	status := *job.Status
	return &status, nil
}

// ListJobs returns all active jobs
func (jm *DefaultJobManager) ListJobs() ([]*streamingTypes.StreamingJobInfo, error) {
	jm.mu.RLock()
	defer jm.mu.RUnlock()

	jobs := make([]*streamingTypes.StreamingJobInfo, 0, len(jm.jobs))
	for _, job := range jm.jobs {
		job.mu.RLock()
		jobInfo := &streamingTypes.StreamingJobInfo{
			Config: *job.Config,
			Status: *job.Status,
		}
		job.mu.RUnlock()
		jobs = append(jobs, jobInfo)
	}

	return jobs, nil
}

// GetJobHistory returns completed/failed jobs (placeholder implementation)

// Shutdown gracefully shuts down the job manager
func (jm *DefaultJobManager) Shutdown() error {
	jm.mu.Lock()
	defer jm.mu.Unlock()

	jm.logger.Info("Shutting down job manager...")

	// Cancel all jobs
	for jobID, job := range jm.jobs {
		job.cancel()
		jm.logger.Info(fmt.Sprintf("Cancelled job: %s", jobID))
	}

	// Cancel manager context
	jm.cancel()

	jm.logger.Info("Job manager shutdown complete")
	return nil
}

// validateJobConfig validates job configuration
func (jm *DefaultJobManager) validateJobConfig(config *streamingTypes.StreamingJobConfig) error {
	if config.JobID == "" {
		return fmt.Errorf("job ID is required")
	}
	if config.Provider == "" {
		return fmt.Errorf("provider is required")
	}
	if config.InputPath == "" {
		return fmt.Errorf("input path is required")
	}
	if config.OutputPath == "" {
		return fmt.Errorf("output path is required")
	}
	if config.Workers <= 0 {
		config.Workers = 4 // Default value
	}
	if config.Memory <= 0 {
		config.Memory = 512 // Default 512MB
	}
	return nil
}

// executeJob runs the actual job processing
func (jm *DefaultJobManager) executeJob(job *Job) {
	job.mu.Lock()
	job.Status.Status = JobStatusRunning
	job.Status.LastUpdate = time.Now()

	// Simulate job processing with checkpoints
	totalRows := int64(1000000) // Simulated total rows
	job.Status.TotalRows = totalRows
	job.mu.Unlock()

	jm.logger.Info(fmt.Sprintf("Executing job: %s", job.Config.JobID))

	for i := int64(0); i < totalRows; i += 10000 {
		// Check for context cancellation
		select {
		case <-job.ctx.Done():
			job.mu.Lock()
			if job.Status.Status == JobStatusRunning {
				job.Status.Status = "cancelled"
			}
			job.Status.LastUpdate = time.Now()
			job.mu.Unlock()
			return
		default:
		}

		// Check if job is paused
		job.mu.RLock()
		isPaused := job.Status.Status == JobStatusPaused
		job.mu.RUnlock()

		if isPaused {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		// Simulate processing
		time.Sleep(10 * time.Millisecond)

		// Update progress
		job.mu.Lock()
		job.Status.ProcessedRows = i + 10000
		if job.Status.ProcessedRows > totalRows {
			job.Status.ProcessedRows = totalRows
		}
		job.Status.Progress = float64(job.Status.ProcessedRows) / float64(totalRows) * 100
		job.Status.LastUpdate = time.Now()

		// Update metrics
		elapsed := time.Since(job.Status.StartTime).Seconds()
		if elapsed > 0 {
			job.Metrics.ProcessingRate = float64(job.Status.ProcessedRows) / elapsed
		}
		job.Metrics.LastUpdate = time.Now()
		job.mu.Unlock()

		// Create checkpoint every 50k rows
		if job.Status.ProcessedRows%50000 == 0 {
			job.worker.checkpoint.ProcessedRows = job.Status.ProcessedRows
			job.worker.checkpoint.Timestamp = time.Now()
		}
	}

	// Job completed
	job.mu.Lock()
	job.Status.Status = "completed"
	job.Status.Progress = 100.0
	job.Status.LastUpdate = time.Now()
	job.Status.EstimatedEnd = time.Now()
	job.mu.Unlock()

	jm.logger.Info(fmt.Sprintf("Job completed: %s", job.Config.JobID))
}
