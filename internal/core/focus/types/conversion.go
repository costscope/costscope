package types

import (
	"context"
	"time"
)

// =====================================================================================
// FOCUS Conversion Types and Interfaces
// =====================================================================================
// Core types and interfaces for cloud billing data conversion to FOCUS format

// ConversionConfig represents configuration for FOCUS conversion
type ConversionConfig struct {
	// Provider settings
	Provider     string `json:"provider" yaml:"provider"`
	ProviderType string `json:"provider_type" yaml:"provider_type"`

	// Input settings
	InputPath   string `json:"input_path" yaml:"input_path"`
	InputFormat string `json:"input_format" yaml:"input_format"`

	// Output settings
	OutputPath   string `json:"output_path" yaml:"output_path"`
	OutputFormat string `json:"output_format" yaml:"output_format"`

	// Output format specific options
	Parquet ParquetOptions `json:"parquet,omitempty" yaml:"parquet,omitempty"`

	// Processing settings
	Streaming   bool `json:"streaming" yaml:"streaming"`
	ChunkSize   int  `json:"chunk_size" yaml:"chunk_size"`
	Workers     int  `json:"workers" yaml:"workers"`
	MaxMemoryMB int  `json:"max_memory_mb" yaml:"max_memory_mb"`
	Compression bool `json:"compression" yaml:"compression"`

	// Validation settings
	ValidateInput  bool `json:"validate_input" yaml:"validate_input"`
	ValidateOutput bool `json:"validate_output" yaml:"validate_output"`
	StrictMode     bool `json:"strict_mode" yaml:"strict_mode"`

	// Metadata
	ConversionId string    `json:"conversion_id" yaml:"conversion_id"`
	CreatedAt    time.Time `json:"created_at" yaml:"created_at"`
	CreatedBy    string    `json:"created_by" yaml:"created_by"`

	// Experiments and feature flags
	UseUnifiedMapper bool `json:"use_unified_mapper" yaml:"use_unified_mapper"`

	// Invariants (post-conversion lightweight quality aggregates)
	InvariantsEnabled   bool    `json:"invariants_enabled" yaml:"invariants_enabled"`
	InvariantsBaseline  string  `json:"invariants_baseline" yaml:"invariants_baseline"`
	InvariantsTolerance float64 `json:"invariants_tolerance" yaml:"invariants_tolerance"`
}

// ParquetOptions controls Parquet writer behavior
type ParquetOptions struct {
	// CompressionCodec sets the Parquet compression codec: snappy (default), gzip, zstd, uncompressed
	CompressionCodec string `json:"compression_codec" yaml:"compression_codec"`
	// RowGroupSizeBytes sets the target row group size in bytes (default 128MB)
	RowGroupSizeBytes int64 `json:"row_group_size_bytes" yaml:"row_group_size_bytes"`
	// PageSizeBytes sets the Parquet page size in bytes (default 8KB)
	PageSizeBytes int64 `json:"page_size_bytes" yaml:"page_size_bytes"`
	// RotateSizeBytes enables rotation when the current file reaches this size in bytes (default 512MB if >0 or unspecified)
	RotateSizeBytes int64 `json:"rotate_size_bytes" yaml:"rotate_size_bytes"`
	// RotateInterval enables time-based rotation using a duration string (e.g. "5m", "1h"); empty disables interval rotation
	RotateInterval string `json:"rotate_interval" yaml:"rotate_interval"`
	// FilePrefix optionally overrides the output filename prefix used for rotated files; if empty, derived from output path
	FilePrefix string `json:"file_prefix" yaml:"file_prefix"`
}

// ConversionProgress represents progress tracking for conversion
type ConversionProgress struct {
	ConversionId           string        `json:"conversion_id"`
	Status                 string        `json:"status"`
	StartTime              time.Time     `json:"start_time"`
	ElapsedTime            time.Duration `json:"elapsed_time"`
	EstimatedTimeRemaining time.Duration `json:"estimated_time_remaining"`

	// Progress metrics
	TotalRecords     int64 `json:"total_records"`
	ProcessedRecords int64 `json:"processed_records"`
	SuccessRecords   int64 `json:"success_records"`
	ErrorRecords     int64 `json:"error_records"`
	SkippedRecords   int64 `json:"skipped_records"`

	// Performance metrics
	RecordsPerSecond float64 `json:"records_per_second"`
	MemoryUsageMB    float64 `json:"memory_usage_mb"`
	CPUUsagePercent  float64 `json:"cpu_usage_percent"`

	// File metrics
	InputFileSizeMB  float64 `json:"input_file_size_mb"`
	OutputFileSizeMB float64 `json:"output_file_size_mb"`
	CompressionRatio float64 `json:"compression_ratio"`

	// Error tracking
	LastError    string   `json:"last_error,omitempty"`
	ErrorDetails []string `json:"error_details,omitempty"`
	Warnings     []string `json:"warnings,omitempty"`
}

// ConversionResult represents the result of a conversion operation
type ConversionResult struct {
	Success      bool          `json:"success"`
	ConversionId string        `json:"conversion_id"`
	StartTime    time.Time     `json:"start_time"`
	EndTime      time.Time     `json:"end_time"`
	Duration     time.Duration `json:"duration"`

	// Input metrics
	InputFile       string  `json:"input_file"`
	InputFormat     string  `json:"input_format"`
	InputRecords    int64   `json:"input_records"`
	InputFileSizeMB float64 `json:"input_file_size_mb"`

	// Output metrics
	OutputFile       string  `json:"output_file"`
	OutputFormat     string  `json:"output_format"`
	OutputRecords    int64   `json:"output_records"`
	OutputFileSizeMB float64 `json:"output_file_size_mb"`
	CompressionRatio float64 `json:"compression_ratio"`

	// Processing metrics
	SuccessRecords   int64   `json:"success_records"`
	ErrorRecords     int64   `json:"error_records"`
	SkippedRecords   int64   `json:"skipped_records"`
	RecordsPerSecond float64 `json:"records_per_second"`

	// Quality metrics
	DataQualityScore float64 `json:"data_quality_score"`
	ComplianceScore  float64 `json:"compliance_score"`

	// Error tracking
	Errors   []ConversionError `json:"errors,omitempty"`
	Warnings []string          `json:"warnings,omitempty"`

	// Metadata
	FocusVersion     string `json:"focus_version"`
	SchemaVersion    string `json:"schema_version"`
	ConverterVersion string `json:"converter_version"`

	// Optional invariants (present when enabled). We keep it as raw map to avoid
	// import cycle on quality package here; populated by caller.
	Invariants any `json:"invariants,omitempty"`
}

// ConversionError represents an error during conversion
type ConversionError struct {
	RecordNumber int64     `json:"record_number"`
	FieldName    string    `json:"field_name"`
	ErrorType    string    `json:"error_type"`
	ErrorMessage string    `json:"error_message"`
	RawValue     string    `json:"raw_value"`
	Timestamp    time.Time `json:"timestamp"`
}

// ConversionStatus represents possible conversion statuses
type ConversionStatus string

const (
	StatusPending   ConversionStatus = "pending"
	StatusRunning   ConversionStatus = "running"
	StatusCompleted ConversionStatus = "completed"
	StatusFailed    ConversionStatus = "failed"
	StatusCancelled ConversionStatus = "cancelled"
	StatusPaused    ConversionStatus = "paused"
)

// ProgressCallback is a function type for progress updates
type ProgressCallback func(progress *ConversionProgress)

// ErrorCallback is a function type for error notifications
type ErrorCallback func(error *ConversionError)

// =====================================================================================
// Core Conversion Interfaces
// =====================================================================================

// Converter defines the interface for converting cloud billing data to FOCUS format
type Converter interface {
	// Convert performs the main conversion operation
	Convert(ctx context.Context, config *ConversionConfig) (*ConversionResult, error)
}

// StreamingConverter defines the interface for streaming conversion
type StreamingConverter interface {
	Converter

	// ConvertStream performs streaming conversion
	ConvertStream(ctx context.Context, config *ConversionConfig, progressCallback ProgressCallback) (*ConversionResult, error)

	// ProcessChunk processes a single chunk of data
	ProcessChunk(ctx context.Context, chunk []byte, chunkNumber int) ([]FocusRecord, error)
}

// BatchConverter defines the interface for batch conversion
type BatchConverter interface { // deprecated; retained only for backward reference
}

// DataReader defines the interface for reading cloud billing data
type DataReader interface {
	// Open opens the data source for reading
	Open(ctx context.Context, path string) error

	// Read reads the next batch of records
	Read(ctx context.Context, batchSize int) ([]interface{}, error)

	// ReadChunk reads a chunk of raw data
	ReadChunk(ctx context.Context, chunkSize int) ([]byte, error)

	// Close closes the data source
	Close() error

	// GetMetadata returns metadata about the data source
	GetMetadata() *DataSourceMetadata
}

// DataWriter defines the interface for writing FOCUS data
type DataWriter interface {
	// Open opens the data destination for writing
	Open(ctx context.Context, path string, schema *FocusSchema) error

	// Write writes a batch of FOCUS records
	Write(ctx context.Context, records []FocusRecord) error

	// WriteChunk writes a chunk of data
	WriteChunk(ctx context.Context, data []byte) error

	// Flush flushes any buffered data
	Flush(ctx context.Context) error

	// Close closes the data destination
	Close() error

	// GetMetadata returns metadata about the written data
	GetMetadata() *DataDestinationMetadata
}

// RecordMapper defines the interface for mapping provider records to FOCUS
type RecordMapper interface {
	// MapRecord maps a single provider record to FOCUS format
	MapRecord(ctx context.Context, record interface{}) (*FocusRecord, error)

	// MapBatch maps a batch of provider records
	MapBatch(ctx context.Context, records []interface{}) ([]FocusRecord, error)

	// ValidateMapping validates the mapping rules
	ValidateMapping() error

	// GetMappingRules returns the mapping rules
	GetMappingRules() *MappingRules
}

// =====================================================================================
// Supporting Types
// =====================================================================================

// ConversionEstimate provides estimates for conversion operations
type ConversionEstimate struct {
	EstimatedDuration     time.Duration `json:"estimated_duration"`
	EstimatedMemoryMB     int           `json:"estimated_memory_mb"`
	EstimatedOutputSizeMB float64       `json:"estimated_output_size_mb"`
	EstimatedRecords      int64         `json:"estimated_records"`
	RecommendedChunkSize  int           `json:"recommended_chunk_size"`
	RecommendedWorkers    int           `json:"recommended_workers"`
}

// SupportedFormats defines supported input/output formats
type SupportedFormats struct {
	InputFormats  []string `json:"input_formats"`
	OutputFormats []string `json:"output_formats"`
}

// DataSourceMetadata contains metadata about the input data source
type DataSourceMetadata struct {
	FilePath     string    `json:"file_path"`
	FileSize     int64     `json:"file_size"`
	RecordCount  int64     `json:"record_count"`
	Format       string    `json:"format"`
	Encoding     string    `json:"encoding"`
	Compression  string    `json:"compression"`
	LastModified time.Time `json:"last_modified"`
	Schema       string    `json:"schema"`
}

// DataDestinationMetadata contains metadata about the output destination
type DataDestinationMetadata struct {
	FilePath         string    `json:"file_path"`
	FileSize         int64     `json:"file_size"`
	RecordCount      int64     `json:"record_count"`
	Format           string    `json:"format"`
	Compression      string    `json:"compression"`
	CompressionRatio float64   `json:"compression_ratio"`
	Created          time.Time `json:"created"`
	Schema           string    `json:"schema"`
}

// MappingRules defines rules for mapping provider data to FOCUS
type MappingRules struct {
	Provider    string                    `json:"provider"`
	Version     string                    `json:"version"`
	FieldMaps   map[string]FieldMapping   `json:"field_maps"`
	Transforms  map[string]TransformRule  `json:"transforms"`
	Validations map[string]ValidationRule `json:"validations"`
}

// FieldMapping defines how to map a provider field to FOCUS
type FieldMapping struct {
	SourceField     string   `json:"source_field"`
	TargetField     string   `json:"target_field"`
	Required        bool     `json:"required"`
	DefaultValue    string   `json:"default_value"`
	Transformations []string `json:"transformations"`
}

// TransformRule defines a transformation rule
type TransformRule struct {
	Type       string                 `json:"type"`
	Parameters map[string]interface{} `json:"parameters"`
}

// ValidationRule defines a validation rule
type ValidationRule struct {
	Type       string                 `json:"type"`
	Parameters map[string]interface{} `json:"parameters"`
}

// =====================================================================================
// Provider-Specific Interfaces
// =====================================================================================

// AWSConverter defines AWS-specific conversion interface
type AWSConverter interface {
	StreamingConverter

	// ConvertCUR converts AWS Cost and Usage Reports
	ConvertCUR(ctx context.Context, config *ConversionConfig) (*ConversionResult, error)

	// ProcessManifest processes AWS CUR manifest files
	ProcessManifest(ctx context.Context, manifestPath string) (*CURManifest, error)
}

// AzureConverter defines Azure-specific conversion interface
type AzureConverter interface {
	StreamingConverter

	// ConvertCostExport converts Azure Cost Management exports
	ConvertCostExport(ctx context.Context, config *ConversionConfig) (*ConversionResult, error)

	// ProcessBillingAccount processes Azure billing account data
	ProcessBillingAccount(ctx context.Context, billingAccount string) error
}

// GCPConverter defines GCP-specific conversion interface
type GCPConverter interface {
	StreamingConverter

	// ConvertBillingExport converts GCP BigQuery billing exports
	ConvertBillingExport(ctx context.Context, config *ConversionConfig) (*ConversionResult, error)

	// ProcessBigQueryTable processes GCP BigQuery billing table
	ProcessBigQueryTable(ctx context.Context, projectId, datasetId, tableId string) error
}

// =====================================================================================
// AWS-Specific Types
// =====================================================================================

// CURManifest represents AWS CUR manifest structure
type CURManifest struct {
	AssemblyId    string      `json:"assemblyId"`
	Account       string      `json:"account"`
	Columns       []CURColumn `json:"columns"`
	Charset       string      `json:"charset"`
	Compression   string      `json:"compression"`
	ContentType   string      `json:"contentType"`
	ReportId      string      `json:"reportId"`
	ReportName    string      `json:"reportName"`
	BillingPeriod struct {
		Start string `json:"start"`
		End   string `json:"end"`
	} `json:"billingPeriod"`
	ReportKeys []string `json:"reportKeys"`
}

// CURColumn represents a column in AWS CUR
type CURColumn struct {
	Category string `json:"category"`
	Name     string `json:"name"`
	Type     string `json:"type"`
}
