//go:build enterprise

package streaming

import (
	"context"
	"fmt"
	"time"

	"local/costscope/cmd/modules/streaming/types"
	"local/costscope/internal/core/logging"
	"local/costscope/internal/core/persistence"
	"local/costscope/internal/providers"
)

// PersistentJobManager extends DefaultJobManager with database persistence
type PersistentJobManager struct {
	*DefaultJobManager
	repository persistence.Repository
	logger     *logging.Logger
}

// NewPersistentJobManager creates a new JobManager with database persistence
func NewPersistentJobManager(repo persistence.Repository, providerMgr *providers.ProviderManager, checkpointDir string) *PersistentJobManager {
	baseManager := NewJobManager(providerMgr, checkpointDir)

	pm := &PersistentJobManager{
		DefaultJobManager: baseManager,
		repository:        repo,
		logger:            logging.NewLogger(logging.LevelInfo),
	}

	// Load existing jobs from database
	if err := pm.loadJobsFromDB(); err != nil {
		pm.logger.Error(fmt.Sprintf("Failed to load jobs from database: %v", err))
	}

	return pm
}

// loadJobsFromDB loads all existing jobs from the database into memory
func (pm *PersistentJobManager) loadJobsFromDB() error {
	ctx := context.Background()

	jobs, err := pm.repository.ListJobs(ctx, persistence.JobFilters{})
	if err != nil {
		return fmt.Errorf("failed to list jobs from database: %w", err)
	}

	pm.mu.Lock()
	defer pm.mu.Unlock()

	for _, jobInfo := range jobs {
		// Convert StreamingJobInfo to Job format expected by DefaultJobManager
		job := &Job{
			Config: &jobInfo.Config,
			Status: &jobInfo.Status,
			Metrics: &types.StreamingJobMetrics{
				JobID: jobInfo.Config.JobID,
			},
		}

		// Add job to in-memory storage
		pm.jobs[jobInfo.Config.JobID] = job

		// For running jobs, log restoration
		if jobInfo.Status.Status == JobStatusRunning {
			pm.logger.Info(fmt.Sprintf("Restoring running job: %s", jobInfo.Config.JobID))
		}
	}

	pm.logger.Info(fmt.Sprintf("Loaded %d jobs from database", len(jobs)))
	return nil
}

// StartJobPersistent creates and starts a new job with database persistence
func (pm *PersistentJobManager) StartJobPersistent(config *types.StreamingJobConfig) (*types.StreamingJobInfo, error) {
	// Set timestamps
	now := time.Now()
	config.CreatedAt = now
	config.UpdatedAt = now

	// Create job info for database
	jobInfo := &types.StreamingJobInfo{
		Config: *config,
		Status: types.StreamingJobStatus{
			JobID:      config.JobID,
			Status:     "created",
			Progress:   0.0,
			StartTime:  now,
			LastUpdate: now,
		},
	}

	// Save to database first
	ctx := context.Background()
	if err := pm.repository.SaveJob(ctx, jobInfo); err != nil {
		return nil, fmt.Errorf("failed to save job to database: %w", err)
	}

	// Start job using base manager
	resultInfo, err := pm.DefaultJobManager.StartJob(config) //nolint:staticcheck
	if err != nil {
		// If start fails, try to clean up database entry
		_ = pm.repository.DeleteJob(ctx, config.JobID)
		return nil, err
	}

	pm.logger.Info(fmt.Sprintf("Created and started persistent job: %s", config.JobID))
	return resultInfo, nil
}

// PauseJobPersistent pauses a job and persists the status change
func (pm *PersistentJobManager) PauseJobPersistent(jobID string) error {
	// Pause using base manager (which now tolerates brief 'starting' phase)
	if err := pm.DefaultJobManager.PauseJob(jobID); err != nil { //nolint:staticcheck
		return err
	}

	// Get current job status with proper locking
	pm.mu.RLock()
	job, exists := pm.jobs[jobID]
	if !exists {
		pm.mu.RUnlock()
		return fmt.Errorf("job %s not found", jobID)
	}

	// Update timestamp with job locking
	now := time.Now()
	job.mu.Lock()
	job.Status.LastUpdate = now
	statusCopy := *job.Status
	job.mu.Unlock()
	job.Config.UpdatedAt = now
	pm.mu.RUnlock()

	// Persist status update
	ctx := context.Background()
	if err := pm.repository.UpdateJobStatus(ctx, jobID, &statusCopy); err != nil {
		pm.logger.Error(fmt.Sprintf("Failed to persist pause status for job %s: %v", jobID, err))
		// Don't return error as the pause operation succeeded
	}

	pm.logger.Info(fmt.Sprintf("Paused persistent job: %s", jobID))
	return nil
}

// ResumeJobPersistent resumes a job and persists the status change
func (pm *PersistentJobManager) ResumeJobPersistent(jobID string) error {
	// Resume using base manager
	if err := pm.DefaultJobManager.ResumeJob(jobID); err != nil { //nolint:staticcheck
		return err
	}

	// Get current job status with proper locking
	pm.mu.RLock()
	job, exists := pm.jobs[jobID]
	if !exists {
		pm.mu.RUnlock()
		return fmt.Errorf("job %s not found", jobID)
	}

	// Update timestamp with job locking
	now := time.Now()
	job.mu.Lock()
	job.Status.LastUpdate = now
	statusCopy := *job.Status
	job.mu.Unlock()
	job.Config.UpdatedAt = now
	pm.mu.RUnlock()

	// Persist status update
	ctx := context.Background()
	if err := pm.repository.UpdateJobStatus(ctx, jobID, &statusCopy); err != nil {
		pm.logger.Error(fmt.Sprintf("Failed to persist resume status for job %s: %v", jobID, err))
	}

	pm.logger.Info(fmt.Sprintf("Resumed persistent job: %s", jobID))
	return nil
}

// StopJobPersistent stops a job and persists the status change
// StopJobPersistent stops a job and persists the status change
func (pm *PersistentJobManager) StopJobPersistent(jobID string) error {
	// Stop using base manager
	if err := pm.DefaultJobManager.StopJob(jobID); err != nil { //nolint:staticcheck
		return err
	}

	// Get current job status with proper locking
	pm.mu.RLock()
	job, exists := pm.jobs[jobID]
	if !exists {
		pm.mu.RUnlock()
		return fmt.Errorf("job %s not found", jobID)
	}

	// Update timestamp with job locking
	now := time.Now()
	job.mu.Lock()
	job.Status.LastUpdate = now
	statusCopy := *job.Status
	job.mu.Unlock()
	job.Config.UpdatedAt = now
	pm.mu.RUnlock()

	// Persist status update
	ctx := context.Background()
	if err := pm.repository.UpdateJobStatus(ctx, jobID, &statusCopy); err != nil {
		pm.logger.Error(fmt.Sprintf("Failed to persist stop status for job %s: %v", jobID, err))
	}

	pm.logger.Info(fmt.Sprintf("Stopped persistent job: %s", jobID))
	return nil
}

// UpdateJobProgressPersistent updates job progress and persists to database
func (pm *PersistentJobManager) UpdateJobProgressPersistent(jobID string, progress float64, processedRows, totalRows int64) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	job, exists := pm.jobs[jobID]
	if !exists {
		return fmt.Errorf("job %s not found", jobID)
	}

	// Update progress with proper locking
	now := time.Now()
	job.mu.Lock()
	job.Status.Progress = progress
	job.Status.ProcessedRows = processedRows
	job.Status.TotalRows = totalRows
	job.Status.LastUpdate = now
	job.mu.Unlock()

	// Update config timestamp
	job.Config.UpdatedAt = now

	// Calculate processing rate and estimated completion
	job.mu.RLock()
	startTime := job.Status.StartTime
	job.mu.RUnlock()

	if startTime.Before(now) {
		elapsed := now.Sub(startTime)
		if processedRows > 0 {
			rate := float64(processedRows) / elapsed.Seconds()
			remaining := totalRows - processedRows
			if rate > 0 && remaining > 0 {
				estimatedSeconds := float64(remaining) / rate
				estimatedEnd := now.Add(time.Duration(estimatedSeconds) * time.Second)
				job.mu.Lock()
				job.Status.EstimatedEnd = estimatedEnd
				job.mu.Unlock()
			}
		}
	}

	// Check if job is completed
	if progress >= 100.0 {
		job.mu.Lock()
		job.Status.Status = "completed"
		job.Status.EstimatedEnd = now
		job.mu.Unlock()
	}

	// Persist status update
	ctx := context.Background()

	// Create a copy of the status to avoid data races during DB update
	job.mu.RLock()
	statusCopy := job.Status
	job.mu.RUnlock()

	if err := pm.repository.UpdateJobStatus(ctx, jobID, statusCopy); err != nil {
		pm.logger.Error(fmt.Sprintf("Failed to persist progress update for job %s: %v", jobID, err))
		// Don't return error as the progress update succeeded in memory
	}

	return nil
}

// ListJobsPersistent returns jobs with optional filtering via database query
func (pm *PersistentJobManager) ListJobsPersistent(filters persistence.JobFilters) ([]*types.StreamingJobInfo, error) {
	ctx := context.Background()

	jobs, err := pm.repository.ListJobs(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to list jobs from database: %w", err)
	}

	return jobs, nil
}

// GetJobFromDB retrieves a job from the database
func (pm *PersistentJobManager) GetJobFromDB(jobID string) (*types.StreamingJobInfo, error) {
	ctx := context.Background()

	job, err := pm.repository.GetJob(ctx, jobID)
	if err != nil {
		return nil, fmt.Errorf("failed to get job from database: %w", err)
	}

	return job, nil
}

// DeleteJobPersistent removes a job from both memory and database
func (pm *PersistentJobManager) DeleteJobPersistent(jobID string) error {
	// Stop job first if running
	if err := pm.DefaultJobManager.StopJob(jobID); err != nil { //nolint:staticcheck
		// Log error but continue with deletion
		pm.logger.Error(fmt.Sprintf("Failed to stop job %s before deletion: %v", jobID, err))
	}

	// Delete from database
	ctx := context.Background()
	if err := pm.repository.DeleteJob(ctx, jobID); err != nil {
		return fmt.Errorf("failed to delete job from database: %w", err)
	}

	// Remove from memory
	pm.mu.Lock()
	delete(pm.jobs, jobID)
	pm.mu.Unlock()

	pm.logger.Info(fmt.Sprintf("Deleted persistent job: %s", jobID))
	return nil
}

// SyncJobToDB synchronizes current in-memory job state to database
func (pm *PersistentJobManager) SyncJobToDB(jobID string) error {
	pm.mu.RLock()
	job, exists := pm.jobs[jobID]
	if !exists {
		pm.mu.RUnlock()
		return fmt.Errorf("job %s not found in memory", jobID)
	}

	// Create job info from current state with proper locking
	job.mu.RLock()
	jobInfo := &types.StreamingJobInfo{
		Config: *job.Config,
		Status: *job.Status,
	}
	job.mu.RUnlock()
	pm.mu.RUnlock()

	// Save to database
	ctx := context.Background()
	if err := pm.repository.SaveJob(ctx, jobInfo); err != nil {
		return fmt.Errorf("failed to sync job to database: %w", err)
	}

	return nil
}

// GetRepository returns the persistence repository
func (pm *PersistentJobManager) GetRepository() persistence.Repository {
	return pm.repository
}

// ShutdownPersistent gracefully shuts down the persistent job manager
func (pm *PersistentJobManager) ShutdownPersistent() error {
	pm.logger.Info("Shutting down persistent job manager")

	// Shutdown base manager first
	if err := pm.DefaultJobManager.Shutdown(); err != nil { //nolint:staticcheck
		pm.logger.Error(fmt.Sprintf("Failed to shutdown base job manager: %v", err))
	}

	// Close repository connection
	if pm.repository != nil {
		if err := pm.repository.Close(); err != nil {
			pm.logger.Error(fmt.Sprintf("Failed to close repository: %v", err))
			return err
		}
	}

	return nil
}
