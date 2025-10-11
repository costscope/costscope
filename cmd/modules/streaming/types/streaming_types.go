package types

import (
	"time"
)

// StreamingJobConfig represents configuration for a streaming job
type StreamingJobConfig struct {
	JobID      string            `json:"job_id"`
	Provider   string            `json:"provider"`
	InputPath  string            `json:"input_path"`
	OutputPath string            `json:"output_path"`
	Workers    int               `json:"workers"`
	Memory     int               `json:"memory_mb"`
	Parameters map[string]string `json:"parameters"`
	CreatedAt  time.Time         `json:"created_at"`
	UpdatedAt  time.Time         `json:"updated_at"`
}

// StreamingJobStatus represents the status of a streaming job
type StreamingJobStatus struct {
	JobID          string    `json:"job_id"`
	Status         string    `json:"status"` // running, paused, stopped, completed, failed
	Progress       float64   `json:"progress_percent"`
	ProcessedRows  int64     `json:"processed_rows"`
	TotalRows      int64     `json:"total_rows"`
	ProcessedBytes int64     `json:"processed_bytes"`
	TotalBytes     int64     `json:"total_bytes"`
	StartTime      time.Time `json:"start_time"`
	LastUpdate     time.Time `json:"last_update"`
	EstimatedEnd   time.Time `json:"estimated_end"`
	ErrorMessage   string    `json:"error_message,omitempty"`
}

// StreamingJobInfo combines configuration and status information
type StreamingJobInfo struct {
	Config StreamingJobConfig `json:"config"`
	Status StreamingJobStatus `json:"status"`
}

// StreamingJobMetrics represents performance metrics for a job
type StreamingJobMetrics struct {
	JobID            string    `json:"job_id"`
	CPUUsage         float64   `json:"cpu_usage_percent"`
	MemoryUsage      int64     `json:"memory_usage_bytes"`
	DiskIORead       int64     `json:"disk_io_read_bytes"`
	DiskIOWrite      int64     `json:"disk_io_write_bytes"`
	NetworkBytesIn   int64     `json:"network_bytes_in"`
	NetworkBytesOut  int64     `json:"network_bytes_out"`
	ProcessingRate   float64   `json:"processing_rate_rows_per_sec"`
	ErrorCount       int       `json:"error_count"`
	WarningCount     int       `json:"warning_count"`
	LastMetricUpdate time.Time `json:"last_metric_update"`
	LastUpdate       time.Time `json:"last_update"`
}

// StreamingJobList represents a list of streaming jobs
type StreamingJobList struct {
	Jobs  []StreamingJobInfo `json:"jobs"`
	Total int                `json:"total"`
	Page  int                `json:"page"`
	Size  int                `json:"size"`
}

// StreamingJobOperation represents an operation on a streaming job
type StreamingJobOperation struct {
	Operation  string                 `json:"operation"` // start, pause, resume, stop
	JobID      string                 `json:"job_id"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
	Timestamp  time.Time              `json:"timestamp"`
	Success    bool                   `json:"success"`
	Message    string                 `json:"message"`
}

// NOTE: Previous enum-style types (JobStatusValue/ProviderType/OperationType) and IsValid helpers
// were removed as they were unused across the codebase. Validation should be performed
// at flag / request parsing boundaries if reintroduced in the future.

// StreamingFlags represents command line flags for streaming operations
type StreamingFlags struct {
	Provider string `json:"provider"`
	Input    string `json:"input"`
	Output   string `json:"output"`
	Workers  int    `json:"workers"`
	JobID    string `json:"job_id"`
	Verbose  bool   `json:"verbose"`
	JSON     bool   `json:"json"`
	Limit    int    `json:"limit"`
}
