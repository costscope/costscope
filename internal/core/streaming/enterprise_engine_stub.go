//go:build !enterprise

package streaming

import (
	"context"
	"time"

	"local/costscope/internal/core/enterprise"
	"local/costscope/internal/core/logging"
)

// Intentional enterprise stub for EnterpriseStreamingEngine.
// Without -tags enterprise this placeholder preserves constructor & method
// signatures. All operations emit disabled metrics and return
// enterprise.DisabledError("streaming_engine"). Real logic lives in the
// enterprise build.
type EnterpriseStreamingEngine struct{}

type EnterpriseStreamingConfig struct {
	MaxConcurrentStreams   int
	MaxFileSize            int64
	ChunkSize              int64
	MemoryLimit            int64
	CompressionEnabled     bool
	CheckpointInterval     int
	ProcessingTimeout      int
	TargetThroughputGBMin  float64
	ParallelWorkers        int
	BufferSize             int
	RetryAttempts          int
	ProgressReportInterval int
}

// StreamingOperationType defines the type of streaming operation
type StreamingOperationType string

const (
	StreamingOperationConvert  StreamingOperationType = "convert"
	StreamingOperationAnalyze  StreamingOperationType = "analyze"
	StreamingOperationValidate StreamingOperationType = "validate"
	StreamingOperationDiff     StreamingOperationType = "diff"
	StreamingOperationMerge    StreamingOperationType = "merge"
	StreamingOperationCompress StreamingOperationType = "compress"
)

// StreamingOperationStatus defines the status of streaming operation
type StreamingOperationStatus string

const (
	StreamingStatusPending   StreamingOperationStatus = "pending"
	StreamingStatusRunning   StreamingOperationStatus = "running"
	StreamingStatusPaused    StreamingOperationStatus = "paused"
	StreamingStatusCompleted StreamingOperationStatus = "completed"
	StreamingStatusFailed    StreamingOperationStatus = "failed"
	StreamingStatusCancelled StreamingOperationStatus = "cancelled"
)

type StreamingProgress struct {
	BytesProcessed    int64
	TotalBytes        int64
	RecordsProcessed  int64
	TotalRecords      int64
	PercentComplete   float64
	EstimatedTimeLeft string
	LastUpdated       time.Time
}

type StreamingPerformance struct {
	ThroughputGBMin  float64
	RecordsPerSecond float64
	MemoryUsageMB    int64
	CPUUsagePercent  float64
	IOWaitTimeMS     int64
	CompressionRatio float64
	ErrorRate        float64
	LastMeasured     time.Time
}

type StreamingConfiguration struct {
	ChunkSizeMB        int
	ParallelWorkers    int
	CompressionEnabled bool
	CheckpointEnabled  bool
	ValidationEnabled  bool
	RetryOnError       bool
	MaxRetryAttempts   int
}

type StreamingError struct {
	ErrorCode        string
	ErrorMessage     string
	ErrorType        string
	FailedChunk      int64
	RecoveryStrategy string
	RetryCount       int
	Timestamp        time.Time
}

type StreamingCheckpoint struct {
	ID             string
	Position       int64
	RecordsCount   int64
	BytesProcessed int64
	Timestamp      time.Time
	ValidationHash string
}

type WorkerStatus string

const (
	WorkerStatusIdle       WorkerStatus = "idle"
	WorkerStatusProcessing WorkerStatus = "processing"
	WorkerStatusError      WorkerStatus = "error"
	WorkerStatusCompleted  WorkerStatus = "completed"
)

type WorkerPerformance struct {
	ThroughputMBSec float64
	RecordsPerSec   float64
	MemoryUsageMB   int64
	LastMeasured    time.Time
}

type WorkerState struct {
	WorkerID        string
	Status          WorkerStatus
	CurrentChunk    int64
	ProcessedChunks int64
	ErrorCount      int
	LastActivity    time.Time
	Performance     *WorkerPerformance
}

type StreamingOperation struct {
	ID              string
	SourcePath      string
	DestinationPath string
	Operation       StreamingOperationType
	Status          StreamingOperationStatus
	StartTime       time.Time
	EndTime         *time.Time
	Progress        *StreamingProgress
	Performance     *StreamingPerformance
	Configuration   *StreamingConfiguration
	ErrorInfo       *StreamingError
	Checkpoints     []StreamingCheckpoint
	WorkerStates    []WorkerState
}

type StreamingOperationRequest struct {
	OperationID     string
	SourcePath      string
	DestinationPath string
	Operation       StreamingOperationType
	Configuration   *StreamingConfiguration
}

const streamingFeature = "streaming_engine"

// disabledStreamingAction records a disabled invocation for a specific action to give
// finer observability in community builds without expanding metric families.
func disabledStreamingAction(action string) error {
	enterprise.ObserveInvocation(streamingFeature+"."+action, false)
	enterprise.ObserveError(streamingFeature+"."+action, "disabled")
	return enterprise.DisabledError(streamingFeature)
}

// Intentional stub (enterprise gating): constructor returns disabled streaming engine.
func NewEnterpriseStreamingEngine(_ *logging.Logger) *EnterpriseStreamingEngine {
	return &EnterpriseStreamingEngine{}
}

// Intentional stub (enterprise gating): always returns enterprise disabled error.
func (e *EnterpriseStreamingEngine) StartStreamingOperation(context.Context, *StreamingOperationRequest) (*StreamingOperation, error) {
	return nil, disabledStreamingAction("start")
}

// Intentional stub (enterprise gating): always returns enterprise disabled error.
func (e *EnterpriseStreamingEngine) GetStreamingOperation(string) (*StreamingOperation, error) {
	return nil, disabledStreamingAction("get")
}

// Resume, Cancel, List are the only control/query operations currently exposed via handlers.
// Intentional stub (enterprise gating): always returns enterprise disabled error.
func (e *EnterpriseStreamingEngine) ResumeStreamingOperation(string) error {
	return disabledStreamingAction("resume")
}

// Intentional stub (enterprise gating): always returns enterprise disabled error.
func (e *EnterpriseStreamingEngine) CancelStreamingOperation(string) error {
	return disabledStreamingAction("cancel")
}

// Intentional stub (enterprise gating): returns empty slice; emits disabled metric.
func (e *EnterpriseStreamingEngine) ListActiveOperations() []*StreamingOperation {
	// Listing returns empty slice to avoid nil surprises in JSON marshalling paths.
	// Intentionally ignore returned disabled error; interface does not surface errors.
	_ = disabledStreamingAction("list") //nolint:errcheck
	return []*StreamingOperation{}
}

// Compile‑time assertion: stub implements minimal interface.
var _ StreamingEngine = (*EnterpriseStreamingEngine)(nil)
