package types

import (
	"testing"
	"time"
)

func TestStreamingJobConfig(t *testing.T) {
	now := time.Now()
	config := &StreamingJobConfig{
		JobID:      "test-job-123",
		Provider:   "aws",
		InputPath:  "/path/to/input.csv",
		OutputPath: "/path/to/output.parquet",
		Workers:    4,
		Memory:     1024,
		CreatedAt:  now,
	}

	if config.JobID != "test-job-123" {
		t.Errorf("Expected JobID to be 'test-job-123', got '%s'", config.JobID)
	}

	if config.Provider != "aws" {
		t.Errorf("Expected Provider to be 'aws', got '%s'", config.Provider)
	}

	if config.InputPath != "/path/to/input.csv" {
		t.Errorf("Expected InputPath to be '/path/to/input.csv', got '%s'", config.InputPath)
	}

	if config.OutputPath != "/path/to/output.parquet" {
		t.Errorf("Expected OutputPath to be '/path/to/output.parquet', got '%s'", config.OutputPath)
	}

	if config.Workers != 4 {
		t.Errorf("Expected Workers to be 4, got %d", config.Workers)
	}

	if config.Memory != 1024 {
		t.Errorf("Expected Memory to be 1024, got %d", config.Memory)
	}

	if config.CreatedAt != now {
		t.Errorf("Expected CreatedAt to be %v, got %v", now, config.CreatedAt)
	}
}

func TestStreamingJobStatus(t *testing.T) {
	status := &StreamingJobStatus{
		JobID:         "test-job-123",
		Status:        "running",
		Progress:      50.5,
		ProcessedRows: 1000,
		TotalRows:     2000,
	}

	if status.JobID != "test-job-123" {
		t.Errorf("Expected JobID to be 'test-job-123', got '%s'", status.JobID)
	}

	if status.Status != "running" {
		t.Errorf("Expected Status to be 'running', got '%s'", status.Status)
	}

	if status.Progress != 50.5 {
		t.Errorf("Expected Progress to be 50.5, got %f", status.Progress)
	}

	if status.ProcessedRows != 1000 {
		t.Errorf("Expected ProcessedRows to be 1000, got %d", status.ProcessedRows)
	}

	if status.TotalRows != 2000 {
		t.Errorf("Expected TotalRows to be 2000, got %d", status.TotalRows)
	}
}

func TestStreamingFlags(t *testing.T) {
	flags := &StreamingFlags{
		Provider: "azure",
		Input:    "input.csv",
		Output:   "output.parquet",
		Workers:  8,
		JobID:    "job-456",
		Verbose:  true,
		JSON:     false,
	}

	if flags.Provider != "azure" {
		t.Errorf("Expected Provider to be 'azure', got '%s'", flags.Provider)
	}

	if flags.Input != "input.csv" {
		t.Errorf("Expected Input to be 'input.csv', got '%s'", flags.Input)
	}

	if flags.Output != "output.parquet" {
		t.Errorf("Expected Output to be 'output.parquet', got '%s'", flags.Output)
	}

	if flags.Workers != 8 {
		t.Errorf("Expected Workers to be 8, got %d", flags.Workers)
	}

	if flags.JobID != "job-456" {
		t.Errorf("Expected JobID to be 'job-456', got '%s'", flags.JobID)
	}

	if !flags.Verbose {
		t.Error("Expected Verbose to be true")
	}

	if flags.JSON {
		t.Error("Expected JSON to be false")
	}
}
