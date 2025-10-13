package persistence

import (
	"context"
	"os"
	"testing"
	"time"

	streamingTypes "github.com/costscope/costscope/cmd/modules/streaming/types"
	providerTypes "github.com/costscope/costscope/internal/providers/types"
)

func TestSQLiteRepository(t *testing.T) {
	// Create temporary database
	tempFile := "test_costscope.db"
	defer func() {
		if err := os.Remove(tempFile); err != nil {
			t.Logf("Failed to remove temp file: %v", err)
		}
	}()

	config := &DatabaseConfig{
		Type:     DatabaseTypeSQLite,
		FilePath: tempFile,
	}

	repo, err := NewSQLiteRepository(config)
	if err != nil {
		t.Fatalf("Failed to create SQLite repository: %v", err)
	}
	defer func() {
		if err := repo.Close(); err != nil {
			t.Logf("Failed to close repository: %v", err)
		}
	}()

	ctx := context.Background()

	// Test health check
	if err := repo.Health(ctx); err != nil {
		t.Errorf("Health check failed: %v", err)
	}

	// Test job operations
	testJobOperations(t, repo, ctx)

	// Test provider operations
	testProviderOperations(t, repo, ctx)
}

func testJobOperations(t *testing.T, repo Repository, ctx context.Context) {
	// Create test job
	job := &streamingTypes.StreamingJobInfo{
		Config: streamingTypes.StreamingJobConfig{
			JobID:      "test-job-001",
			Provider:   "aws",
			InputPath:  "/test/input.csv",
			OutputPath: "/test/output.parquet",
			Workers:    4,
			Memory:     512,
			Parameters: map[string]string{
				"format":      "parquet",
				"compression": "snappy",
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Status: streamingTypes.StreamingJobStatus{
			JobID:          "test-job-001",
			Status:         "running",
			Progress:       25.5,
			ProcessedRows:  10000,
			TotalRows:      40000,
			ProcessedBytes: 1024000,
			TotalBytes:     4096000,
			StartTime:      time.Now().Add(-10 * time.Minute),
			LastUpdate:     time.Now(),
			EstimatedEnd:   time.Now().Add(30 * time.Minute),
		},
	}

	// Test SaveJob
	if err := repo.SaveJob(ctx, job); err != nil {
		t.Errorf("SaveJob failed: %v", err)
	}

	// Test GetJob
	retrievedJob, err := repo.GetJob(ctx, job.Config.JobID)
	if err != nil {
		t.Errorf("GetJob failed: %v", err)
	}

	if retrievedJob.Config.JobID != job.Config.JobID {
		t.Errorf("Expected job ID %s, got %s", job.Config.JobID, retrievedJob.Config.JobID)
	}

	if retrievedJob.Config.Provider != job.Config.Provider {
		t.Errorf("Expected provider %s, got %s", job.Config.Provider, retrievedJob.Config.Provider)
	}

	if retrievedJob.Status.Status != job.Status.Status {
		t.Errorf("Expected status %s, got %s", job.Status.Status, retrievedJob.Status.Status)
	}

	if retrievedJob.Status.Progress != job.Status.Progress {
		t.Errorf("Expected progress %f, got %f", job.Status.Progress, retrievedJob.Status.Progress)
	}

	// Test parameters
	if len(retrievedJob.Config.Parameters) != len(job.Config.Parameters) {
		t.Errorf("Expected %d parameters, got %d", len(job.Config.Parameters), len(retrievedJob.Config.Parameters))
	}

	for key, value := range job.Config.Parameters {
		if retrievedJob.Config.Parameters[key] != value {
			t.Errorf("Expected parameter %s = %s, got %s", key, value, retrievedJob.Config.Parameters[key])
		}
	}

	// Test UpdateJobStatus
	newStatus := &streamingTypes.StreamingJobStatus{
		JobID:          job.Config.JobID,
		Status:         "completed",
		Progress:       100.0,
		ProcessedRows:  40000,
		TotalRows:      40000,
		ProcessedBytes: 4096000,
		TotalBytes:     4096000,
		LastUpdate:     time.Now(),
		EstimatedEnd:   time.Now(),
	}

	if err := repo.UpdateJobStatus(ctx, job.Config.JobID, newStatus); err != nil {
		t.Errorf("UpdateJobStatus failed: %v", err)
	}

	// Verify update
	updatedJob, err := repo.GetJob(ctx, job.Config.JobID)
	if err != nil {
		t.Errorf("GetJob after update failed: %v", err)
	}

	if updatedJob.Status.Status != "completed" {
		t.Errorf("Expected status 'completed', got %s", updatedJob.Status.Status)
	}

	if updatedJob.Status.Progress != 100.0 {
		t.Errorf("Expected progress 100.0, got %f", updatedJob.Status.Progress)
	}

	// Test ListJobs with no filters
	jobs, err := repo.ListJobs(ctx, JobFilters{})
	if err != nil {
		t.Errorf("ListJobs failed: %v", err)
	}

	if len(jobs) != 1 {
		t.Errorf("Expected 1 job, got %d", len(jobs))
	}

	// Test ListJobs with status filter
	jobs, err = repo.ListJobs(ctx, JobFilters{
		Status: []string{"completed"},
	})
	if err != nil {
		t.Errorf("ListJobs with filter failed: %v", err)
	}

	if len(jobs) != 1 {
		t.Errorf("Expected 1 job with status filter, got %d", len(jobs))
	}

	// Test ListJobs with provider filter
	jobs, err = repo.ListJobs(ctx, JobFilters{
		Provider: []string{"aws"},
	})
	if err != nil {
		t.Errorf("ListJobs with provider filter failed: %v", err)
	}

	if len(jobs) != 1 {
		t.Errorf("Expected 1 job with provider filter, got %d", len(jobs))
	}

	// Test ListJobs with limit
	jobs, err = repo.ListJobs(ctx, JobFilters{
		Limit: 10,
	})
	if err != nil {
		t.Errorf("ListJobs with limit failed: %v", err)
	}

	if len(jobs) > 10 {
		t.Errorf("Expected at most 10 jobs, got %d", len(jobs))
	}

	// Test DeleteJob
	if err := repo.DeleteJob(ctx, job.Config.JobID); err != nil {
		t.Errorf("DeleteJob failed: %v", err)
	}

	// Verify deletion
	_, err = repo.GetJob(ctx, job.Config.JobID)
	if err == nil {
		t.Error("Expected error when getting deleted job")
	}

	// Test GetJob with non-existent job
	_, err = repo.GetJob(ctx, "non-existent-job")
	if err == nil {
		t.Error("Expected error when getting non-existent job")
	}
}

func testProviderOperations(t *testing.T, repo Repository, ctx context.Context) {
	// Create test provider config
	config := &providerTypes.ProviderConfig{
		Name: "test-aws-provider",
		Type: providerTypes.ProviderTypeAWS,
		Credentials: map[string]string{
			"access_key_id":     "test-access-key",
			"secret_access_key": "test-secret-key",
			"region":            "us-east-1",
		},
		Settings: map[string]interface{}{
			"timeout": 30,
			"retries": 3,
		},
		Regions:   []string{"us-east-1", "us-west-2"},
		IsDefault: true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Test SaveProvider
	if err := repo.SaveProvider(ctx, config); err != nil {
		t.Errorf("SaveProvider failed: %v", err)
	}

	// Test GetProvider
	retrievedConfig, err := repo.GetProvider(ctx, config.Name)
	if err != nil {
		t.Errorf("GetProvider failed: %v", err)
	}

	if retrievedConfig.Name != config.Name {
		t.Errorf("Expected name %s, got %s", config.Name, retrievedConfig.Name)
	}

	if retrievedConfig.Type != config.Type {
		t.Errorf("Expected type %s, got %s", config.Type, retrievedConfig.Type)
	}

	if retrievedConfig.IsDefault != config.IsDefault {
		t.Errorf("Expected is_default %t, got %t", config.IsDefault, retrievedConfig.IsDefault)
	}

	// Test credentials
	if len(retrievedConfig.Credentials) != len(config.Credentials) {
		t.Errorf("Expected %d credentials, got %d", len(config.Credentials), len(retrievedConfig.Credentials))
	}

	for key, value := range config.Credentials {
		if retrievedConfig.Credentials[key] != value {
			t.Errorf("Expected credential %s = %s, got %s", key, value, retrievedConfig.Credentials[key])
		}
	}

	// Test settings
	if len(retrievedConfig.Settings) != len(config.Settings) {
		t.Errorf("Expected %d settings, got %d", len(config.Settings), len(retrievedConfig.Settings))
	}

	// Test regions
	if len(retrievedConfig.Regions) != len(config.Regions) {
		t.Errorf("Expected %d regions, got %d", len(config.Regions), len(retrievedConfig.Regions))
	}

	for i, region := range config.Regions {
		if retrievedConfig.Regions[i] != region {
			t.Errorf("Expected region %s, got %s", region, retrievedConfig.Regions[i])
		}
	}

	// Test ListProviders
	providers, err := repo.ListProviders(ctx)
	if err != nil {
		t.Errorf("ListProviders failed: %v", err)
	}

	if len(providers) != 1 {
		t.Errorf("Expected 1 provider, got %d", len(providers))
	}

	// Test DeleteProvider
	if err := repo.DeleteProvider(ctx, config.Name); err != nil {
		t.Errorf("DeleteProvider failed: %v", err)
	}

	// Verify deletion
	_, err = repo.GetProvider(ctx, config.Name)
	if err == nil {
		t.Error("Expected error when getting deleted provider")
	}

	// Test GetProvider with non-existent provider
	_, err = repo.GetProvider(ctx, "non-existent-provider")
	if err == nil {
		t.Error("Expected error when getting non-existent provider")
	}
}

func TestDatabaseConfig(t *testing.T) {
	// Test default config
	config := DefaultDatabaseConfig()
	// Validate default config still points to sqlite and connection string equals file path
	if config.Type != DatabaseTypeSQLite {
		t.Fatalf("expected default type sqlite, got %s", config.Type)
	}
	if config.FilePath != "./costscope.db" {
		t.Fatalf("expected default file path ./costscope.db, got %s", config.FilePath)
	}
	if got := config.GetConnectionString(); got != config.FilePath {
		t.Fatalf("expected connection string %s, got %s", config.FilePath, got)
	}
	if !DatabaseTypeSQLite.IsValid() {
		t.Fatalf("expected sqlite type to be valid")
	}
}

// TestDatabaseType ensures IsValid only returns true for sqlite (postgres stub removed)
func TestDatabaseType(t *testing.T) {
	if !DatabaseTypeSQLite.IsValid() {
		t.Errorf("sqlite should be valid")
	}
	if DatabaseType("bogus").IsValid() {
		t.Errorf("bogus type should be invalid")
	}
}
