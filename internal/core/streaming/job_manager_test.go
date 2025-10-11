//go:build enterprise

package streaming

import (
	"context"
	"testing"
	"time"

	streamingTypes "local/costscope/cmd/modules/streaming/types"
	"local/costscope/internal/providers"
	providerTypes "local/costscope/internal/providers/types"
)

// MockProvider implements CloudProvider interface for testing
type MockProvider struct {
	name         string
	providerType providerTypes.ProviderType
}

func (m *MockProvider) ValidateCredentials(ctx context.Context, config map[string]string) error {
	return nil
}

func (m *MockProvider) GetProviderInfo(ctx context.Context) (providerTypes.ProviderInfo, error) {
	return providerTypes.ProviderInfo{
		Name:             m.name,
		Type:             m.providerType,
		Version:          "1.0.0",
		SupportedRegions: []string{"us-east-1", "us-west-2"},
		Capabilities:     []string{"cost", "resources"},
		Metadata:         map[string]string{},
	}, nil
}

func (m *MockProvider) GetCostData(ctx context.Context, params providerTypes.CostDataParams) ([]providerTypes.CostRecord, error) {
	return []providerTypes.CostRecord{}, nil
}

func (m *MockProvider) GetResourceData(ctx context.Context, params providerTypes.ResourceDataParams) ([]providerTypes.ResourceRecord, error) {
	return []providerTypes.ResourceRecord{}, nil
}

func (m *MockProvider) GetName() string {
	return m.name
}

func (m *MockProvider) GetType() providerTypes.ProviderType {
	return m.providerType
}

func (m *MockProvider) GetSupportedRegions() []string {
	return []string{"us-east-1", "us-west-2"}
}

// createTestJobManager creates a job manager with mock provider for testing
func createTestJobManager(t *testing.T) *DefaultJobManager {
	// Create provider manager
	providerMgr := providers.NewProviderManager()

	// Register mock provider
	mockProvider := &MockProvider{
		name:         "test-aws",
		providerType: providerTypes.ProviderTypeAWS,
	}

	config := &providerTypes.ProviderConfig{
		Name:        "test-aws",
		Type:        providerTypes.ProviderTypeAWS,
		Credentials: map[string]string{},
		Settings:    map[string]interface{}{},
		Regions:     []string{"us-east-1"},
		IsDefault:   false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := providerMgr.RegisterProvider("test-aws", mockProvider, config)
	if err != nil {
		t.Fatalf("Failed to register mock provider: %v", err)
	}

	return NewJobManager(providerMgr, "/tmp/checkpoints")
}

// createTestJobConfig creates a test job configuration
func createTestJobConfig() *streamingTypes.StreamingJobConfig {
	return &streamingTypes.StreamingJobConfig{
		JobID:      "test-job-123",
		Provider:   "test-aws",
		InputPath:  "/test/input.csv",
		OutputPath: "/test/output.parquet",
		Workers:    4,
		Memory:     512,
		Parameters: map[string]string{
			"format": "parquet",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func TestNewJobManager(t *testing.T) {
	providerMgr := providers.NewProviderManager()
	jm := NewJobManager(providerMgr, "/tmp/checkpoints")

	if jm == nil {
		t.Fatal("NewJobManager returned nil")
	}

	if jm.jobs == nil {
		t.Error("Jobs map not initialized")
	}

	if jm.providerMgr != providerMgr {
		t.Error("Provider manager not set correctly")
	}

	if jm.checkpointDir != "/tmp/checkpoints" {
		t.Error("Checkpoint directory not set correctly")
	}
}

func TestStartJob(t *testing.T) {
	jm := createTestJobManager(t)
	defer func() {
		if err := jm.Shutdown(); err != nil {
			t.Logf("Failed to shutdown job manager: %v", err)
		}
	}()

	config := createTestJobConfig()

	jobInfo, err := jm.StartJob(config)
	if err != nil {
		t.Fatalf("StartJob failed: %v", err)
	}

	if jobInfo == nil {
		t.Fatal("StartJob returned nil job info")
	}

	if jobInfo.Config.JobID != config.JobID {
		t.Errorf("Expected job ID %s, got %s", config.JobID, jobInfo.Config.JobID)
	}

	// Allow for immediate transition to running due to fast execution
	if jobInfo.Status.Status != "starting" && jobInfo.Status.Status != JobStatusRunning {
		fresh, _ := jm.GetJobStatus(config.JobID)
		if fresh == nil || (fresh.Status != "starting" && fresh.Status != JobStatusRunning) {
			t.Errorf("Expected status 'starting' or 'running', got %s", jobInfo.Status.Status)
		}
	}

	// Wait briefly then assert running
	time.Sleep(100 * time.Millisecond)
	status, err := jm.GetJobStatus(config.JobID)
	if err != nil {
		t.Errorf("GetJobStatus failed: %v", err)
	} else if status.Status != JobStatusRunning {
		t.Errorf("Expected job to be running, got status: %s", status.Status)
	}
}

func TestStartJobDuplicate(t *testing.T) {
	jm := createTestJobManager(t)
	defer func() { _ = jm.Shutdown() }()

	config := createTestJobConfig()

	// Start first job
	_, err := jm.StartJob(config)
	if err != nil {
		t.Fatalf("First StartJob failed: %v", err)
	}

	// Try to start same job again
	_, err = jm.StartJob(config)
	if err == nil {
		t.Error("Expected error when starting duplicate job")
	}
}

func TestStartJobInvalidConfig(t *testing.T) {
	jm := createTestJobManager(t)
	defer func() { _ = jm.Shutdown() }()

	// Test with empty job ID
	config := createTestJobConfig()
	config.JobID = ""

	_, err := jm.StartJob(config)
	if err == nil {
		t.Error("Expected error with empty job ID")
	}

	// Test with empty provider
	config = createTestJobConfig()
	config.Provider = ""

	_, err = jm.StartJob(config)
	if err == nil {
		t.Error("Expected error with empty provider")
	}
}

func TestPauseJob(t *testing.T) {
	jm := createTestJobManager(t)
	defer func() {
		_ = jm.Shutdown()
	}()

	config := createTestJobConfig()

	// Start job
	_, err := jm.StartJob(config)
	if err != nil {
		t.Fatalf("StartJob failed: %v", err)
	}

	// Wait for job to start running
	time.Sleep(100 * time.Millisecond)

	// Pause job
	err = jm.PauseJob(config.JobID)
	if err != nil {
		t.Errorf("PauseJob failed: %v", err)
	}

	// Check job is paused
	status, err := jm.GetJobStatus(config.JobID)
	if err != nil {
		t.Errorf("GetJobStatus failed: %v", err)
	} else if status.Status != JobStatusPaused {
		t.Errorf("Expected job to be paused, got status: %s", status.Status)
	}
}

func TestResumeJob(t *testing.T) {
	jm := createTestJobManager(t)
	defer func() { _ = jm.Shutdown() }()

	config := createTestJobConfig()

	// Start and pause job
	_, err := jm.StartJob(config)
	if err != nil {
		t.Fatalf("StartJob failed: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	err = jm.PauseJob(config.JobID)
	if err != nil {
		t.Fatalf("PauseJob failed: %v", err)
	}

	// Resume job
	err = jm.ResumeJob(config.JobID)
	if err != nil {
		t.Errorf("ResumeJob failed: %v", err)
	}

	// Check job is running
	status, err := jm.GetJobStatus(config.JobID)
	if err != nil {
		t.Errorf("GetJobStatus failed: %v", err)
	} else if status.Status != JobStatusRunning {
		t.Errorf("Expected job to be running, got status: %s", status.Status)
	}
}

func TestStopJob(t *testing.T) {
	jm := createTestJobManager(t)
	defer func() { _ = jm.Shutdown() }()

	config := createTestJobConfig()

	// Start job
	_, err := jm.StartJob(config)
	if err != nil {
		t.Fatalf("StartJob failed: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	// Stop job
	err = jm.StopJob(config.JobID)
	if err != nil {
		t.Errorf("StopJob failed: %v", err)
	}

	// Check job is stopped
	status, err := jm.GetJobStatus(config.JobID)
	if err != nil {
		t.Errorf("GetJobStatus failed: %v", err)
	} else if status.Status != JobStatusStopped {
		t.Errorf("Expected job to be stopped, got status: %s", status.Status)
	}
}

func TestGetJobStatusNotFound(t *testing.T) {
	jm := createTestJobManager(t)
	defer func() { _ = jm.Shutdown() }()

	_, err := jm.GetJobStatus("nonexistent-job")
	if err == nil {
		t.Error("Expected error for nonexistent job")
	}
}

func TestListJobs(t *testing.T) {
	jm := createTestJobManager(t)
	defer func() { _ = jm.Shutdown() }()

	// Initially no jobs
	jobs, err := jm.ListJobs()
	if err != nil {
		t.Errorf("ListJobs failed: %v", err)
	}
	if len(jobs) != 0 {
		t.Errorf("Expected 0 jobs, got %d", len(jobs))
	}

	// Start a job
	config := createTestJobConfig()
	_, err = jm.StartJob(config)
	if err != nil {
		t.Fatalf("StartJob failed: %v", err)
	}

	// List jobs
	jobs, err = jm.ListJobs()
	if err != nil {
		t.Errorf("ListJobs failed: %v", err)
	}
	if len(jobs) != 1 {
		t.Errorf("Expected 1 job, got %d", len(jobs))
	}

	if jobs[0].Config.JobID != config.JobID {
		t.Errorf("Expected job ID %s, got %s", config.JobID, jobs[0].Config.JobID)
	}
}

func TestJobProgress(t *testing.T) {
	jm := createTestJobManager(t)
	defer func() { _ = jm.Shutdown() }()

	config := createTestJobConfig()

	// Start job
	_, err := jm.StartJob(config)
	if err != nil {
		t.Fatalf("StartJob failed: %v", err)
	}

	// Wait for some progress
	time.Sleep(200 * time.Millisecond)

	// Check progress
	status, err := jm.GetJobStatus(config.JobID)
	if err != nil {
		t.Errorf("GetJobStatus failed: %v", err)
	} else {
		if status.Progress <= 0 {
			t.Errorf("Expected progress > 0, got %f", status.Progress)
		}
		if status.ProcessedRows <= 0 {
			t.Errorf("Expected processed rows > 0, got %d", status.ProcessedRows)
		}
	}
}

func TestShutdown(t *testing.T) {
	jm := createTestJobManager(t)

	config := createTestJobConfig()

	// Start job
	_, err := jm.StartJob(config)
	if err != nil {
		t.Fatalf("StartJob failed: %v", err)
	}

	// Shutdown
	err = jm.Shutdown()
	if err != nil {
		t.Errorf("Shutdown failed: %v", err)
	}

	// Check job is cancelled after shutdown
	time.Sleep(100 * time.Millisecond)
	status, err := jm.GetJobStatus(config.JobID)
	if err != nil {
		t.Errorf("GetJobStatus failed: %v", err)
	} else if status.Status != "cancelled" && status.Status != "stopped" {
		t.Errorf("Expected job to be cancelled or stopped, got status: %s", status.Status)
	}
}
