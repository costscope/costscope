package jobs

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/costscope/costscope/internal/core/logging"
)

// =====================================================================================
// Async Job Management System - Enterprise Job Processing
// =====================================================================================

// JobStatus represents the current status of a job
type JobStatus string

const (
	StatusPending   JobStatus = "pending"
	StatusRunning   JobStatus = "running"
	StatusCompleted JobStatus = "completed"
	StatusFailed    JobStatus = "failed"
	StatusCancelled JobStatus = "cancelled"
)

// JobPriority represents job priority levels
type JobPriority int

const (
	PriorityLow    JobPriority = 1
	PriorityNormal JobPriority = 5
	PriorityHigh   JobPriority = 10
)

// Progress represents job execution progress
type Progress struct {
	Current      int64      `json:"current"`
	Total        int64      `json:"total"`
	Percentage   float64    `json:"percentage"`
	Message      string     `json:"message"`
	Stage        string     `json:"stage"`
	UpdatedAt    time.Time  `json:"updated_at"`
	EstimatedETA *time.Time `json:"estimated_eta,omitempty"`
}

// JobConfig represents job configuration
type JobConfig struct {
	ID           string            `json:"id"`
	Type         string            `json:"type"`
	Priority     JobPriority       `json:"priority"`
	Timeout      time.Duration     `json:"timeout"`
	MaxRetries   int               `json:"max_retries"`
	WebhookURL   string            `json:"webhook_url,omitempty"`
	CallbackData interface{}       `json:"callback_data,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
	// TenantID is an optional multi-tenant identifier (feature-flag gated; skeleton only)
	TenantID string `json:"tenant_id,omitempty"`
}

// Job represents an async job
type Job struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Status      JobStatus              `json:"status"`
	Priority    JobPriority            `json:"priority"`
	CreatedAt   time.Time              `json:"created_at"`
	StartedAt   *time.Time             `json:"started_at,omitempty"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	Progress    *Progress              `json:"progress,omitempty"`
	Result      map[string]interface{} `json:"result,omitempty"`
	Error       error                  `json:"error,omitempty"`
	Config      *JobConfig             `json:"config"`
	Retries     int                    `json:"retries"`
	LastRetry   *time.Time             `json:"last_retry,omitempty"`
	// TenantID is an optional multi-tenant identifier (feature-flag gated; skeleton only)
	TenantID string `json:"tenant_id,omitempty"`
}

// Task represents a job task that can be executed
type Task interface {
	Execute(ctx context.Context, job *Job, progressCallback func(*Progress)) error
	GetType() string
	GetDescription() string
}

// Manager manages async job execution
type Manager struct {
	logger       *logging.Logger
	jobs         map[string]*Job
	jobsMutex    sync.RWMutex
	workers      int
	jobQueue     chan *Job
	stopChan     chan struct{}
	running      bool
	runningMutex sync.RWMutex

	// Multi-tenant quota tracking (in-memory, non-persistent)
	maxJobsPerTenant       int
	maxActiveJobsPerTenant int
	tenantTotals           map[string]int // total submitted
	tenantActive           map[string]int // pending + running

	// Optional real-time broadcaster (e.g., WebSocket manager)
	broadcaster Broadcaster
}

// Broadcaster is a minimal interface for broadcasting job updates without importing concrete WS types.
type Broadcaster interface {
	BroadcastJobProgress(jobID string, progress *Progress)
	BroadcastJobStatus(jobID string, status JobStatus, result map[string]interface{}, error string)
}

// NewManager creates a new job manager
func NewManager(logger *logging.Logger, workers int) *Manager {
	return &Manager{
		logger:       logger,
		jobs:         make(map[string]*Job),
		workers:      workers,
		jobQueue:     make(chan *Job, workers*2), // Buffer for pending jobs
		stopChan:     make(chan struct{}),
		tenantTotals: make(map[string]int),
		tenantActive: make(map[string]int),
	}
}

// SetBroadcaster attaches a real-time broadcaster to the job manager.
func (m *Manager) SetBroadcaster(b Broadcaster) {
	m.jobsMutex.Lock()
	defer m.jobsMutex.Unlock()
	m.broadcaster = b
}

// IsRunning reports whether the job manager is currently running.
// It uses a read lock to safely access the running flag.
func (m *Manager) IsRunning() bool {
	m.runningMutex.RLock()
	defer m.runningMutex.RUnlock()
	return m.running
}

// ConfigureQuotas sets per-tenant quota limits (0 disables respective limit)
func (m *Manager) ConfigureQuotas(maxJobs, maxActive int) {
	m.jobsMutex.Lock()
	m.maxJobsPerTenant = maxJobs
	m.maxActiveJobsPerTenant = maxActive
	m.jobsMutex.Unlock()
	m.logger.Info(fmt.Sprintf("Job quotas configured: max_jobs_per_tenant=%d max_active_jobs_per_tenant=%d", maxJobs, maxActive))
}

// Start starts the job manager
func (m *Manager) Start() error {
	m.runningMutex.Lock()
	defer m.runningMutex.Unlock()

	if m.running {
		return errors.New("job manager is already running")
	}

	m.running = true

	// Start worker goroutines
	for i := 0; i < m.workers; i++ {
		go m.worker(i)
	}

	m.logger.Info(fmt.Sprintf("Job manager started with %d workers", m.workers))
	return nil
}

// Stop stops the job manager
func (m *Manager) Stop() error {
	m.runningMutex.Lock()
	defer m.runningMutex.Unlock()

	if !m.running {
		return errors.New("job manager is not running")
	}

	close(m.stopChan)
	m.running = false

	m.logger.Info("Job manager stopped")
	return nil
}

// SubmitJob submits a job for execution
func (m *Manager) SubmitJob(config *JobConfig, task Task) (*Job, error) {
	if config.ID == "" {
		return nil, errors.New("job ID is required")
	}

	// Quota enforcement (best-effort, in-memory)
	tenantID := config.TenantID
	if tenantID != "" { // only enforce when tenant present (multi-tenant mode)
		m.jobsMutex.Lock()
		if m.maxJobsPerTenant > 0 && m.tenantTotals[tenantID] >= m.maxJobsPerTenant {
			m.jobsMutex.Unlock()
			return nil, fmt.Errorf("tenant job quota exceeded (max %d total)", m.maxJobsPerTenant)
		}
		if m.maxActiveJobsPerTenant > 0 && m.tenantActive[tenantID] >= m.maxActiveJobsPerTenant {
			m.jobsMutex.Unlock()
			return nil, fmt.Errorf("tenant active job quota exceeded (max %d active)", m.maxActiveJobsPerTenant)
		}
		// Pre-increment totals (active includes pending)
		m.tenantTotals[tenantID]++
		m.tenantActive[tenantID]++
		m.jobsMutex.Unlock()
	}

	// Create job
	job := &Job{
		ID:        config.ID,
		Type:      config.Type,
		Status:    StatusPending,
		Priority:  config.Priority,
		CreatedAt: time.Now(),
		Config:    config,
		Result:    make(map[string]interface{}),
	}

	// Store job
	m.jobsMutex.Lock()
	m.jobs[job.ID] = job
	m.jobsMutex.Unlock()

	// Queue job for execution
	go func() {
		select {
		case m.jobQueue <- job:
			// Job queued successfully
		default:
			// Queue is full, mark job as failed
			m.jobsMutex.Lock()
			job.Status = StatusFailed
			job.Error = errors.New("job queue is full")
			job.CompletedAt = &[]time.Time{time.Now()}[0]
			m.jobsMutex.Unlock()
		}
	}()

	m.logger.Info(fmt.Sprintf("Job submitted: %s (%s)", job.ID, job.Type))
	return job, nil
}

// GetJob retrieves a job by ID
func (m *Manager) GetJob(jobID string) (*Job, error) {
	m.jobsMutex.RLock()
	job, exists := m.jobs[jobID]
	if !exists {
		m.jobsMutex.RUnlock()
		return nil, errors.New("job not found")
	}
	// Clone under lock to avoid exposing internal mutable state to callers
	clone := cloneJob(job)
	m.jobsMutex.RUnlock()
	return clone, nil
}

// ListJobs returns a list of jobs with optional filtering
func (m *Manager) ListJobs(status JobStatus, jobType string, limit int) ([]*Job, error) {
	m.jobsMutex.RLock()
	defer m.jobsMutex.RUnlock()

	var result []*Job
	count := 0

	for _, job := range m.jobs {
		if limit > 0 && count >= limit {
			break
		}

		// Apply filters
		if status != "" && job.Status != status {
			continue
		}
		if jobType != "" && job.Type != jobType {
			continue
		}

		result = append(result, job)
		count++
	}

	return result, nil
}

// CancelJob cancels a running job
func (m *Manager) CancelJob(jobID string) error {
	m.jobsMutex.Lock()
	defer m.jobsMutex.Unlock()

	job, exists := m.jobs[jobID]
	if !exists {
		return errors.New("job not found")
	}

	if job.Status == StatusCompleted || job.Status == StatusFailed || job.Status == StatusCancelled {
		return errors.New("job cannot be cancelled")
	}

	job.Status = StatusCancelled
	job.CompletedAt = &[]time.Time{time.Now()}[0]

	m.logger.Info(fmt.Sprintf("Job cancelled: %s", jobID))
	return nil
}

// worker is the worker goroutine that processes jobs
func (m *Manager) worker(workerID int) {
	m.logger.Info(fmt.Sprintf("Worker %d started", workerID))

	for {
		select {
		case <-m.stopChan:
			m.logger.Info(fmt.Sprintf("Worker %d stopped", workerID))
			return

		case job := <-m.jobQueue:
			if job == nil {
				continue
			}

			m.executeJob(job, workerID)
		}
	}
}

// executeJob executes a single job
func (m *Manager) executeJob(job *Job, workerID int) {
	defer func() {
		if r := recover(); r != nil {
			m.jobsMutex.Lock()
			job.Status = StatusFailed
			job.Error = fmt.Errorf("job panicked: %v", r)
			job.CompletedAt = &[]time.Time{time.Now()}[0]
			m.jobsMutex.Unlock()

			m.logger.Error(fmt.Sprintf("Job %s panicked: %v", job.ID, r))
		}
	}()

	// Update job status to running
	m.jobsMutex.Lock()
	job.Status = StatusRunning
	job.StartedAt = &[]time.Time{time.Now()}[0]
	m.jobsMutex.Unlock()

	m.logger.Info(fmt.Sprintf("Worker %d executing job: %s (%s)", workerID, job.ID, job.Type))

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), job.Config.Timeout)
	defer cancel()

	// Progress callback
	progressCallback := func(progress *Progress) {
		m.jobsMutex.Lock()
		job.Progress = progress
		m.jobsMutex.Unlock()

		// Broadcast progress if configured
		if m.broadcaster != nil {
			m.broadcaster.BroadcastJobProgress(job.ID, progress)
		}
	}

	// Execute the task (mock implementation for now)
	var err error

	// Simulate task execution
	task := &MockTask{Type: job.Type}
	err = task.Execute(ctx, job, progressCallback)

	// Update job status
	m.jobsMutex.Lock()
	if err != nil {
		job.Status = StatusFailed
		job.Error = err
	} else {
		job.Status = StatusCompleted
		job.Result["completed_at"] = time.Now()
		job.Result["worker_id"] = workerID
	}
	job.CompletedAt = &[]time.Time{time.Now()}[0]
	// Decrement active counter if tenant set
	if job.TenantID != "" && m.tenantActive[job.TenantID] > 0 {
		m.tenantActive[job.TenantID]--
	}
	// Snapshot for broadcasting outside lock
	status := job.Status
	result := job.Result
	var errMsg string
	if job.Error != nil {
		errMsg = job.Error.Error()
	}
	m.jobsMutex.Unlock()

	// Broadcast terminal status
	if m.broadcaster != nil {
		m.broadcaster.BroadcastJobStatus(job.ID, status, result, errMsg)
	}

	if err != nil {
		m.logger.Error(fmt.Sprintf("Job %s failed: %s", job.ID, err.Error()))
	} else {
		m.logger.Info(fmt.Sprintf("Job %s completed successfully", job.ID))
	}
}

// =====================================================================================
// Mock Task Implementation (temporary)
// =====================================================================================

// MockTask is a mock task implementation for testing
type MockTask struct {
	Type string
}

func (t *MockTask) Execute(ctx context.Context, job *Job, progressCallback func(*Progress)) error {
	// Simulate work with progress updates
	stages := []string{"initializing", "processing", "finalizing"}

	for i, stage := range stages {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Update progress
		progress := &Progress{
			Current:    int64(i + 1),
			Total:      int64(len(stages)),
			Percentage: float64(i+1) / float64(len(stages)) * 100,
			Message:    fmt.Sprintf("Executing stage: %s", stage),
			Stage:      stage,
			UpdatedAt:  time.Now(),
		}

		if i < len(stages)-1 {
			eta := time.Now().Add(time.Duration(len(stages)-i-1) * time.Second)
			progress.EstimatedETA = &eta
		}

		progressCallback(progress)

		// Simulate work
		time.Sleep(time.Second)
	}

	return nil
}

// cloneJob creates a deep-ish copy of Job safe for concurrent reads by callers.
// It copies value fields, maps, and referenced Progress struct to avoid data races
// when the manager mutates the original.
func cloneJob(src *Job) *Job {
	if src == nil {
		return nil
	}
	dst := *src // copy value fields
	// Copy Result map
	if src.Result != nil {
		dst.Result = make(map[string]interface{}, len(src.Result))
		for k, v := range src.Result {
			dst.Result[k] = v
		}
	}
	// Copy Progress struct
	if src.Progress != nil {
		p := *src.Progress
		dst.Progress = &p
	}
	// Note: Config pointer is shared for read-only access; callers MUST NOT mutate.
	// If future use requires, we can add a shallow copy of Config.
	return &dst
}
