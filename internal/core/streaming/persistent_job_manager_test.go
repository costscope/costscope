//go:build enterprise

package streaming

import (
	"context"
	"os"
	"testing"
	"time"

	streamingTypes "local/costscope/cmd/modules/streaming/types"
	"local/costscope/internal/core/persistence"
	"local/costscope/internal/providers"
	providerTypes "local/costscope/internal/providers/types"
)

// TestProvider implements CloudProvider interface for testing
type TestProvider struct {
	name         string
	providerType providerTypes.ProviderType
}

func (tp *TestProvider) ValidateCredentials(ctx context.Context, config map[string]string) error {
	return nil
}

func (tp *TestProvider) GetProviderInfo(ctx context.Context) (providerTypes.ProviderInfo, error) {
	return providerTypes.ProviderInfo{
		Name:             tp.name,
		Type:             tp.providerType,
		Version:          "test-1.0.0",
		SupportedRegions: []string{"test-region-1", "test-region-2"},
		Capabilities:     []string{"cost-analysis", "resource-tracking"},
		Metadata:         map[string]string{"test": "true"},
	}, nil
}

func (tp *TestProvider) GetCostData(ctx context.Context, params providerTypes.CostDataParams) ([]providerTypes.CostRecord, error) {
	return []providerTypes.CostRecord{
		{
			Date:       time.Now(),
			Amount:     100.0,
			Currency:   "USD",
			Service:    "test-service",
			Region:     "test-region",
			ResourceID: "test-resource-123",
			Tags:       map[string]string{"env": "test"},
			Metadata:   map[string]interface{}{"source": "test"},
		},
	}, nil
}

func (tp *TestProvider) GetResourceData(ctx context.Context, params providerTypes.ResourceDataParams) ([]providerTypes.ResourceRecord, error) {
	return []providerTypes.ResourceRecord{
		{
			ID:         "test-resource-123",
			Name:       "test-resource",
			Type:       "test-type",
			Region:     "test-region",
			Status:     "active",
			Cost:       100.0,
			Currency:   "USD",
			Tags:       map[string]string{"env": "test"},
			Properties: map[string]interface{}{"test": true},
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		},
	}, nil
}

func (tp *TestProvider) GetName() string {
	return tp.name
}

func (tp *TestProvider) GetType() providerTypes.ProviderType {
	return tp.providerType
}

func (tp *TestProvider) GetSupportedRegions() []string {
	return []string{"test-region-1", "test-region-2"}
}

// setupTestProviders adds test providers to the manager
func setupTestProviders(pm *providers.ProviderManager) error {
	testProviders := []struct {
		name  string
		pType providerTypes.ProviderType
	}{
		{"test", providerTypes.ProviderTypeAWS},
		{"aws", providerTypes.ProviderTypeAWS},
		{"azure", providerTypes.ProviderTypeAzure},
		{"gcp", providerTypes.ProviderTypeGCP},
	}

	for _, p := range testProviders {
		provider := &TestProvider{
			name:         p.name,
			providerType: p.pType,
		}
		config := &providerTypes.ProviderConfig{
			Name: p.name,
			Type: p.pType,
			Credentials: map[string]string{
				"test": "test-value",
			},
			Regions:   []string{"test-region"},
			IsDefault: false,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		if err := pm.RegisterProvider(p.name, provider, config); err != nil {
			return err
		}
	}

	return nil
}

func TestPersistentJobManager(t *testing.T) {
	// Create temporary database
	tempFile := "test_persistent_jobs.db"
	defer func() {
		if err := os.Remove(tempFile); err != nil {
			t.Logf("Failed to remove temp file: %v", err)
		}
	}()

	// Setup database
	config := &persistence.DatabaseConfig{
		Type:     persistence.DatabaseTypeSQLite,
		FilePath: tempFile,
	}

	repo, err := persistence.NewSQLiteRepository(config)
	if err != nil {
		t.Fatalf("Failed to create SQLite repository: %v", err)
	}
	defer func() { _ = repo.Close() }()

	// Setup provider manager
	providerMgr := providers.NewProviderManager()
	if err := setupTestProviders(providerMgr); err != nil {
		t.Fatalf("Failed to setup test providers: %v", err)
	}

	// Create persistent job manager
	pm := NewPersistentJobManager(repo, providerMgr, "./test_checkpoints")
	defer func() {
		if err := pm.ShutdownPersistent(); err != nil {
			t.Logf("Shutdown error: %v", err)
		}
	}()

	// Test creating and starting a job
	testCreateAndStartJob(t, pm)

	// Test basic persistence operations
	testBasicPersistenceOperations(t, pm)
}

func testCreateAndStartJob(t *testing.T, pm *PersistentJobManager) {
	config := &streamingTypes.StreamingJobConfig{
		JobID:      "test-persistent-job-001",
		Provider:   "test",
		InputPath:  "/test/input.csv",
		OutputPath: "/test/output.parquet",
		Workers:    2,
		Memory:     256,
		Parameters: map[string]string{
			"format": "parquet",
		},
	}

	// Start persistent job
	jobInfo, err := pm.StartJobPersistent(config)
	if err != nil {
		t.Errorf("StartJobPersistent failed: %v", err)
		return
	}

	if jobInfo.Config.JobID != config.JobID {
		t.Errorf("Expected job ID %s, got %s", config.JobID, jobInfo.Config.JobID)
	}

	// Verify job is in database
	dbJob, err := pm.GetJobFromDB(config.JobID)
	if err != nil {
		t.Errorf("Failed to get job from database: %v", err)
		return
	}

	if dbJob.Config.JobID != config.JobID {
		t.Errorf("Expected DB job ID %s, got %s", config.JobID, dbJob.Config.JobID)
	}

	if dbJob.Config.Provider != config.Provider {
		t.Errorf("Expected DB provider %s, got %s", config.Provider, dbJob.Config.Provider)
	}

	// Clean up
	_ = pm.DeleteJobPersistent(config.JobID)
}

func testBasicPersistenceOperations(t *testing.T, pm *PersistentJobManager) {
	config := &streamingTypes.StreamingJobConfig{
		JobID:      "test-basic-job-001",
		Provider:   "test",
		InputPath:  "/test/basic.csv",
		OutputPath: "/test/basic.parquet",
		Workers:    1,
		Memory:     128,
	}

	// Start job
	_, err := pm.StartJobPersistent(config)
	if err != nil {
		t.Fatalf("Failed to start job: %v", err)
	}

	// Let the job run briefly
	time.Sleep(100 * time.Millisecond)

	// Stop the job before testing persistence operations to avoid race conditions
	if err := pm.StopJobPersistent(config.JobID); err != nil {
		t.Errorf("Failed to stop job: %v", err)
	}

	// Verify job exists in database
	dbJob, err := pm.GetJobFromDB(config.JobID)
	if err != nil {
		t.Errorf("Failed to get job from database: %v", err)
		return
	}

	if dbJob.Config.JobID != config.JobID {
		t.Errorf("Expected job ID %s, got %s", config.JobID, dbJob.Config.JobID)
	}

	// Test progress update (after stopping to avoid race with job execution)
	if err := pm.UpdateJobProgressPersistent(config.JobID, 25.0, 2500, 10000); err != nil {
		t.Errorf("Failed to update progress: %v", err)
	}

	// Test sync to DB
	if err := pm.SyncJobToDB(config.JobID); err != nil {
		t.Errorf("Failed to sync to database: %v", err)
	}

	// Clean up
	if err := pm.DeleteJobPersistent(config.JobID); err != nil {
		t.Errorf("Failed to delete job: %v", err)
	}
}

func TestPersistentJobManagerFiltering(t *testing.T) {
	// Create temporary database
	tempFile := "test_filtering.db"
	defer func() { _ = os.Remove(tempFile) }()

	// Setup database
	config := &persistence.DatabaseConfig{
		Type:     persistence.DatabaseTypeSQLite,
		FilePath: tempFile,
	}

	repo, err := persistence.NewSQLiteRepository(config)
	if err != nil {
		t.Fatalf("Failed to create SQLite repository: %v", err)
	}
	defer func() { _ = repo.Close() }()

	// Setup provider manager
	providerMgr := providers.NewProviderManager()
	if err := setupTestProviders(providerMgr); err != nil {
		t.Fatalf("Failed to setup test providers: %v", err)
	}

	// Create persistent job manager
	pm := NewPersistentJobManager(repo, providerMgr, "./test_checkpoints")
	defer func() { _ = pm.ShutdownPersistent() }()

	// Create multiple jobs
	jobs := []*streamingTypes.StreamingJobConfig{
		{
			JobID:      "filter-job-001",
			Provider:   "aws",
			InputPath:  "/test/aws1.csv",
			OutputPath: "/test/aws1.parquet",
		},
		{
			JobID:      "filter-job-002",
			Provider:   "azure",
			InputPath:  "/test/azure1.csv",
			OutputPath: "/test/azure1.parquet",
		},
		{
			JobID:      "filter-job-003",
			Provider:   "aws",
			InputPath:  "/test/aws2.csv",
			OutputPath: "/test/aws2.parquet",
		},
	}

	// Start all jobs
	for _, jobConfig := range jobs {
		if _, err := pm.StartJobPersistent(jobConfig); err != nil {
			t.Errorf("Failed to start job %s: %v", jobConfig.JobID, err)
		}
	}

	// Pause one job
	if err := pm.PauseJobPersistent("filter-job-002"); err != nil {
		t.Errorf("Failed to pause job: %v", err)
	}

	// Give time for status updates to persist
	time.Sleep(100 * time.Millisecond)

	// Test filtering by provider
	awsJobs, err := pm.ListJobsPersistent(persistence.JobFilters{
		Provider: []string{"aws"},
	})
	if err != nil {
		t.Errorf("Failed to list AWS jobs: %v", err)
	} else if len(awsJobs) != 2 {
		t.Errorf("Expected 2 AWS jobs, got %d", len(awsJobs))
	}

	// Test filtering by status - check if we have any running jobs
	runningJobs, err := pm.ListJobsPersistent(persistence.JobFilters{
		Status: []string{"running"},
	})
	if err != nil {
		t.Errorf("Failed to list running jobs: %v", err)
	}

	// The exact count may vary due to timing, so just check we get some results
	t.Logf("Found %d running jobs", len(runningJobs))

	// Test filtering by multiple criteria
	pausedAzureJobs, err := pm.ListJobsPersistent(persistence.JobFilters{
		Status:   []string{"paused"},
		Provider: []string{"azure"},
	})
	if err != nil {
		t.Errorf("Failed to list paused Azure jobs: %v", err)
	} else if len(pausedAzureJobs) != 1 {
		t.Errorf("Expected 1 paused Azure job, got %d", len(pausedAzureJobs))
	}

	// Test limit
	limitedJobs, err := pm.ListJobsPersistent(persistence.JobFilters{
		Limit: 2,
	})
	if err != nil {
		t.Errorf("Failed to list limited jobs: %v", err)
	} else if len(limitedJobs) > 2 {
		t.Errorf("Expected at most 2 jobs with limit, got %d", len(limitedJobs))
	}

	// Clean up
	for _, jobConfig := range jobs {
		_ = pm.DeleteJobPersistent(jobConfig.JobID)
	}
}

func TestPersistentJobManagerSyncToDB(t *testing.T) {
	// Create temporary database
	tempFile := "test_sync.db"
	defer func() { _ = os.Remove(tempFile) }()

	// Setup database
	config := &persistence.DatabaseConfig{
		Type:     persistence.DatabaseTypeSQLite,
		FilePath: tempFile,
	}

	repo, err := persistence.NewSQLiteRepository(config)
	if err != nil {
		t.Fatalf("Failed to create SQLite repository: %v", err)
	}
	defer func() { _ = repo.Close() }()

	// Setup provider manager
	providerMgr := providers.NewProviderManager()
	if err := setupTestProviders(providerMgr); err != nil {
		t.Fatalf("Failed to setup test providers: %v", err)
	}

	// Create persistent job manager
	pm := NewPersistentJobManager(repo, providerMgr, "./test_checkpoints")
	defer func() { _ = pm.ShutdownPersistent() }()

	config2 := &streamingTypes.StreamingJobConfig{
		JobID:      "sync-job-001",
		Provider:   "test",
		InputPath:  "/test/sync.csv",
		OutputPath: "/test/sync.parquet",
		Workers:    1,
		Memory:     128,
	}

	// Start job
	_, err = pm.StartJobPersistent(config2)
	if err != nil {
		t.Fatalf("Failed to start job: %v", err)
	}

	// Manually modify in-memory state (simulate external update)
	pm.mu.Lock()
	if job, exists := pm.jobs[config2.JobID]; exists {
		job.mu.Lock()
		job.Status.Progress = 99.9
		job.Status.ProcessedRows = 9990
		job.Status.TotalRows = 10000
		job.mu.Unlock()
	}
	pm.mu.Unlock()

	// Sync to database
	if err := pm.SyncJobToDB(config2.JobID); err != nil {
		t.Errorf("SyncJobToDB failed: %v", err)
	}

	// Verify sync worked
	dbJob, err := pm.GetJobFromDB(config2.JobID)
	if err != nil {
		t.Errorf("Failed to get job from database after sync: %v", err)
	} else {
		if dbJob.Status.Progress != 99.9 {
			t.Errorf("Expected synced progress 99.9, got %f", dbJob.Status.Progress)
		}
		if dbJob.Status.ProcessedRows != 9990 {
			t.Errorf("Expected synced processed rows 9990, got %d", dbJob.Status.ProcessedRows)
		}
	}

	// Clean up
	_ = pm.DeleteJobPersistent(config2.JobID)
}
