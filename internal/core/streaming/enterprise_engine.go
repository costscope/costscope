//go:build enterprise

package streaming

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/costscope/costscope/internal/core/logging"
	"github.com/costscope/costscope/internal/database/performance"
)

// EnterpriseStreamingEngine handles enterprise-scale streaming operations
type EnterpriseStreamingEngine struct {
	performanceEngine *performance.PerformanceEngine
	logger            *logging.Logger
	config            *EnterpriseStreamingConfig
	activeStreams     map[string]*StreamingOperation
	mu                sync.RWMutex
}

// EnterpriseStreamingConfig holds enterprise streaming configuration
type EnterpriseStreamingConfig struct {
	MaxConcurrentStreams   int     `json:"max_concurrent_streams"`
	MaxFileSize            int64   `json:"max_file_size_gb"`
	ChunkSize              int64   `json:"chunk_size_mb"`
	MemoryLimit            int64   `json:"memory_limit_mb"`
	CompressionEnabled     bool    `json:"compression_enabled"`
	CheckpointInterval     int     `json:"checkpoint_interval_seconds"`
	ProcessingTimeout      int     `json:"processing_timeout_seconds"`
	TargetThroughputGBMin  float64 `json:"target_throughput_gb_per_min"`
	ParallelWorkers        int     `json:"parallel_workers"`
	BufferSize             int     `json:"buffer_size_mb"`
	RetryAttempts          int     `json:"retry_attempts"`
	ProgressReportInterval int     `json:"progress_report_interval_seconds"`
}

// StreamingOperation represents an active streaming operation
type StreamingOperation struct {
	ID              string                   `json:"id"`
	SourcePath      string                   `json:"source_path"`
	DestinationPath string                   `json:"destination_path"`
	Operation       StreamingOperationType   `json:"operation"`
	Status          StreamingOperationStatus `json:"status"`
	StartTime       time.Time                `json:"start_time"`
	EndTime         *time.Time               `json:"end_time,omitempty"`
	Progress        *StreamingProgress       `json:"progress"`
	Performance     *StreamingPerformance    `json:"performance"`
	Configuration   *StreamingConfiguration  `json:"configuration"`
	ErrorInfo       *StreamingError          `json:"error_info,omitempty"`
	Checkpoints     []StreamingCheckpoint    `json:"checkpoints"`
	WorkerStates    []WorkerState            `json:"worker_states"`
	ctx             context.Context
	cancel          context.CancelFunc
	mu              sync.RWMutex
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

// StreamingProgress tracks the progress of streaming operations
type StreamingProgress struct {
	BytesProcessed    int64     `json:"bytes_processed"`
	TotalBytes        int64     `json:"total_bytes"`
	RecordsProcessed  int64     `json:"records_processed"`
	TotalRecords      int64     `json:"total_records"`
	PercentComplete   float64   `json:"percent_complete"`
	EstimatedTimeLeft string    `json:"estimated_time_left"`
	LastUpdated       time.Time `json:"last_updated"`
}

// StreamingPerformance tracks performance metrics for streaming operations
type StreamingPerformance struct {
	ThroughputGBMin  float64   `json:"throughput_gb_per_min"`
	RecordsPerSecond float64   `json:"records_per_second"`
	MemoryUsageMB    int64     `json:"memory_usage_mb"`
	CPUUsagePercent  float64   `json:"cpu_usage_percent"`
	IOWaitTimeMS     int64     `json:"io_wait_time_ms"`
	CompressionRatio float64   `json:"compression_ratio"`
	ErrorRate        float64   `json:"error_rate"`
	LastMeasured     time.Time `json:"last_measured"`
}

// StreamingConfiguration holds configuration for a specific streaming operation
type StreamingConfiguration struct {
	ChunkSizeMB        int  `json:"chunk_size_mb"`
	ParallelWorkers    int  `json:"parallel_workers"`
	CompressionEnabled bool `json:"compression_enabled"`
	CheckpointEnabled  bool `json:"checkpoint_enabled"`
	ValidationEnabled  bool `json:"validation_enabled"`
	RetryOnError       bool `json:"retry_on_error"`
	MaxRetryAttempts   int  `json:"max_retry_attempts"`
}

// StreamingError represents error information for streaming operations
type StreamingError struct {
	ErrorCode        string    `json:"error_code"`
	ErrorMessage     string    `json:"error_message"`
	ErrorType        string    `json:"error_type"`
	FailedChunk      int64     `json:"failed_chunk"`
	RecoveryStrategy string    `json:"recovery_strategy"`
	RetryCount       int       `json:"retry_count"`
	Timestamp        time.Time `json:"timestamp"`
}

// StreamingCheckpoint represents a checkpoint in streaming operation
type StreamingCheckpoint struct {
	ID             string    `json:"id"`
	Position       int64     `json:"position"`
	RecordsCount   int64     `json:"records_count"`
	BytesProcessed int64     `json:"bytes_processed"`
	Timestamp      time.Time `json:"timestamp"`
	ValidationHash string    `json:"validation_hash"`
}

// WorkerState represents the state of a streaming worker
type WorkerState struct {
	WorkerID        string             `json:"worker_id"`
	Status          WorkerStatus       `json:"status"`
	CurrentChunk    int64              `json:"current_chunk"`
	ProcessedChunks int64              `json:"processed_chunks"`
	ErrorCount      int                `json:"error_count"`
	LastActivity    time.Time          `json:"last_activity"`
	Performance     *WorkerPerformance `json:"performance"`
}

// WorkerStatus defines the status of a streaming worker
type WorkerStatus string

const (
	WorkerStatusIdle       WorkerStatus = "idle"
	WorkerStatusProcessing WorkerStatus = "processing"
	WorkerStatusError      WorkerStatus = "error"
	WorkerStatusCompleted  WorkerStatus = "completed"
)

// WorkerPerformance tracks performance metrics for individual workers
type WorkerPerformance struct {
	ThroughputMBSec float64   `json:"throughput_mb_per_sec"`
	RecordsPerSec   float64   `json:"records_per_sec"`
	MemoryUsageMB   int64     `json:"memory_usage_mb"`
	LastMeasured    time.Time `json:"last_measured"`
}

// NewEnterpriseStreamingEngine creates a new enterprise streaming engine
func NewEnterpriseStreamingEngine(logger *logging.Logger) *EnterpriseStreamingEngine {
	config := &EnterpriseStreamingConfig{
		MaxConcurrentStreams:   10,
		MaxFileSize:            100,  // 100GB
		ChunkSize:              100,  // 100MB chunks
		MemoryLimit:            2048, // 2GB memory limit
		CompressionEnabled:     true,
		CheckpointInterval:     60,   // 1 minute
		ProcessingTimeout:      3600, // 1 hour
		TargetThroughputGBMin:  1.0,  // 1GB/min minimum
		ParallelWorkers:        8,
		BufferSize:             64, // 64MB buffer
		RetryAttempts:          3,
		ProgressReportInterval: 30, // 30 seconds
	}

	performanceEngine := performance.NewPerformanceEngine(performance.DefaultPerformanceConfig())

	return &EnterpriseStreamingEngine{
		performanceEngine: performanceEngine,
		logger:            logger,
		config:            config,
		activeStreams:     make(map[string]*StreamingOperation),
	}
}

// StartStreamingOperation starts a new enterprise streaming operation
func (ese *EnterpriseStreamingEngine) StartStreamingOperation(ctx context.Context, req *StreamingOperationRequest) (*StreamingOperation, error) {
	ese.logger.Info(fmt.Sprintf("Starting enterprise streaming operation: %s -> %s", req.SourcePath, req.DestinationPath))

	// Validate request
	if err := ese.validateStreamingRequest(req); err != nil {
		return nil, fmt.Errorf("invalid streaming request: %w", err)
	}

	// Check concurrent streams limit
	ese.mu.RLock()
	activeCount := len(ese.activeStreams)
	ese.mu.RUnlock()

	if activeCount >= ese.config.MaxConcurrentStreams {
		return nil, fmt.Errorf("maximum concurrent streams (%d) reached", ese.config.MaxConcurrentStreams)
	}

	// Create streaming operation
	operationCtx, cancel := context.WithTimeout(ctx, time.Duration(ese.config.ProcessingTimeout)*time.Second)

	operation := &StreamingOperation{
		ID:              req.OperationID,
		SourcePath:      req.SourcePath,
		DestinationPath: req.DestinationPath,
		Operation:       req.Operation,
		Status:          StreamingStatusPending,
		StartTime:       time.Now(),
		Progress: &StreamingProgress{
			LastUpdated: time.Now(),
		},
		Performance: &StreamingPerformance{
			LastMeasured: time.Now(),
		},
		Configuration: &StreamingConfiguration{
			ChunkSizeMB:        req.Configuration.ChunkSizeMB,
			ParallelWorkers:    req.Configuration.ParallelWorkers,
			CompressionEnabled: req.Configuration.CompressionEnabled,
			CheckpointEnabled:  req.Configuration.CheckpointEnabled,
			ValidationEnabled:  req.Configuration.ValidationEnabled,
			RetryOnError:       req.Configuration.RetryOnError,
			MaxRetryAttempts:   req.Configuration.MaxRetryAttempts,
		},
		ctx:    operationCtx,
		cancel: cancel,
	}

	// Initialize worker states
	operation.WorkerStates = make([]WorkerState, operation.Configuration.ParallelWorkers)
	for i := 0; i < operation.Configuration.ParallelWorkers; i++ {
		operation.WorkerStates[i] = WorkerState{
			WorkerID:     fmt.Sprintf("worker-%d", i),
			Status:       WorkerStatusIdle,
			LastActivity: time.Now(),
			Performance: &WorkerPerformance{
				LastMeasured: time.Now(),
			},
		}
	}

	// Store operation
	ese.mu.Lock()
	ese.activeStreams[operation.ID] = operation
	ese.mu.Unlock()

	// Start operation processing
	go ese.processStreamingOperation(operation)

	return operation, nil
}

// StreamingOperationRequest represents a request to start streaming operation
type StreamingOperationRequest struct {
	OperationID     string                  `json:"operation_id"`
	SourcePath      string                  `json:"source_path"`
	DestinationPath string                  `json:"destination_path"`
	Operation       StreamingOperationType  `json:"operation"`
	Configuration   *StreamingConfiguration `json:"configuration"`
}

// GetStreamingOperation retrieves information about a streaming operation
func (ese *EnterpriseStreamingEngine) GetStreamingOperation(operationID string) (*StreamingOperation, error) {
	ese.mu.RLock()
	defer ese.mu.RUnlock()

	operation, exists := ese.activeStreams[operationID]
	if !exists {
		return nil, fmt.Errorf("streaming operation %s not found", operationID)
	}

	// Return a copy of the operation data without copying the mutex
	operation.mu.RLock()
	defer operation.mu.RUnlock()

	operationCopy := &StreamingOperation{
		ID:              operation.ID,
		SourcePath:      operation.SourcePath,
		DestinationPath: operation.DestinationPath,
		Operation:       operation.Operation,
		Status:          operation.Status,
		StartTime:       operation.StartTime,
		EndTime:         operation.EndTime,
		Progress:        operation.Progress,
		Performance:     operation.Performance,
		Configuration:   operation.Configuration,
		ErrorInfo:       operation.ErrorInfo,
		Checkpoints:     operation.Checkpoints,
		WorkerStates:    operation.WorkerStates,
	}

	return operationCopy, nil
}

// PauseStreamingOperation pauses a streaming operation
func (ese *EnterpriseStreamingEngine) PauseStreamingOperation(operationID string) error {
	ese.mu.RLock()
	operation, exists := ese.activeStreams[operationID]
	ese.mu.RUnlock()

	if !exists {
		return fmt.Errorf("streaming operation %s not found", operationID)
	}

	operation.mu.Lock()
	defer operation.mu.Unlock()

	if operation.Status != StreamingStatusRunning {
		return fmt.Errorf("cannot pause operation in status %s", operation.Status)
	}

	operation.Status = StreamingStatusPaused
	ese.logger.Info(fmt.Sprintf("Streaming operation %s paused", operationID))

	return nil
}

// ResumeStreamingOperation resumes a paused streaming operation
func (ese *EnterpriseStreamingEngine) ResumeStreamingOperation(operationID string) error {
	ese.mu.RLock()
	operation, exists := ese.activeStreams[operationID]
	ese.mu.RUnlock()

	if !exists {
		return fmt.Errorf("streaming operation %s not found", operationID)
	}

	operation.mu.Lock()
	defer operation.mu.Unlock()

	if operation.Status != StreamingStatusPaused {
		return fmt.Errorf("cannot resume operation in status %s", operation.Status)
	}

	operation.Status = StreamingStatusRunning
	ese.logger.Info(fmt.Sprintf("Streaming operation %s resumed", operationID))

	return nil
}

// CancelStreamingOperation cancels a streaming operation
func (ese *EnterpriseStreamingEngine) CancelStreamingOperation(operationID string) error {
	ese.mu.RLock()
	operation, exists := ese.activeStreams[operationID]
	ese.mu.RUnlock()

	if !exists {
		return fmt.Errorf("streaming operation %s not found", operationID)
	}

	operation.cancel()

	operation.mu.Lock()
	operation.Status = StreamingStatusCancelled
	endTime := time.Now()
	operation.EndTime = &endTime
	operation.mu.Unlock()

	ese.logger.Info(fmt.Sprintf("Streaming operation %s cancelled", operationID))

	return nil
}

// ListActiveOperations returns all active streaming operations
func (ese *EnterpriseStreamingEngine) ListActiveOperations() []*StreamingOperation {
	ese.mu.RLock()
	defer ese.mu.RUnlock()

	operations := make([]*StreamingOperation, 0, len(ese.activeStreams))
	for _, operation := range ese.activeStreams {
		operation.mu.RLock()
		operationCopy := &StreamingOperation{
			ID:              operation.ID,
			SourcePath:      operation.SourcePath,
			DestinationPath: operation.DestinationPath,
			Operation:       operation.Operation,
			Status:          operation.Status,
			StartTime:       operation.StartTime,
			EndTime:         operation.EndTime,
			Progress:        operation.Progress,
			Performance:     operation.Performance,
			Configuration:   operation.Configuration,
			ErrorInfo:       operation.ErrorInfo,
			Checkpoints:     operation.Checkpoints,
			WorkerStates:    operation.WorkerStates,
		}
		operation.mu.RUnlock()
		operations = append(operations, operationCopy)
	}

	return operations
}

// processStreamingOperation processes a streaming operation
func (ese *EnterpriseStreamingEngine) processStreamingOperation(operation *StreamingOperation) {
	ese.logger.Info(fmt.Sprintf("Processing streaming operation %s", operation.ID))

	// Update status to running
	operation.mu.Lock()
	operation.Status = StreamingStatusRunning
	operation.mu.Unlock()

	// Get file info
	fileInfo, err := os.Stat(operation.SourcePath)
	if err != nil {
		ese.handleOperationError(operation, fmt.Errorf("failed to get file info: %w", err))
		return
	}

	// Update total bytes
	operation.mu.Lock()
	operation.Progress.TotalBytes = fileInfo.Size()
	operation.mu.Unlock()

	// Start performance monitoring
	go ese.monitorOperationPerformance(operation)

	// Start progress reporting
	go ese.reportOperationProgress(operation)

	// Process based on operation type
	switch operation.Operation {
	case StreamingOperationConvert:
		err = ese.processConversionOperation(operation)
	case StreamingOperationAnalyze:
		err = ese.processAnalysisOperation(operation)
	case StreamingOperationValidate:
		err = ese.processValidationOperation(operation)
	case StreamingOperationDiff:
		err = ese.processDiffOperation(operation)
	case StreamingOperationMerge:
		err = ese.processMergeOperation(operation)
	case StreamingOperationCompress:
		err = ese.processCompressionOperation(operation)
	default:
		err = fmt.Errorf("unsupported operation type: %s", operation.Operation)
	}

	// Handle completion or error
	operation.mu.Lock()
	endTime := time.Now()
	operation.EndTime = &endTime

	if err != nil {
		operation.Status = StreamingStatusFailed
		operation.ErrorInfo = &StreamingError{
			ErrorCode:    "PROCESSING_ERROR",
			ErrorMessage: err.Error(),
			ErrorType:    "processing",
			Timestamp:    endTime,
		}
		ese.logger.Error(fmt.Sprintf("Streaming operation %s failed: %v", operation.ID, err))
	} else {
		operation.Status = StreamingStatusCompleted
		operation.Progress.PercentComplete = 100.0
		ese.logger.Info(fmt.Sprintf("Streaming operation %s completed successfully", operation.ID))
	}
	operation.mu.Unlock()

	// Clean up after delay
	go func() {
		time.Sleep(5 * time.Minute) // Keep completed operations for 5 minutes
		ese.mu.Lock()
		delete(ese.activeStreams, operation.ID)
		ese.mu.Unlock()
	}()
}

// Placeholder processing methods for different operation types

func (ese *EnterpriseStreamingEngine) processConversionOperation(operation *StreamingOperation) error {
	// Simulate FOCUS conversion with chunked processing
	return ese.processWithChunks(operation, func(chunk []byte, chunkIndex int64) error {
		// Simulate conversion processing
		time.Sleep(100 * time.Millisecond)
		return nil
	})
}

func (ese *EnterpriseStreamingEngine) processAnalysisOperation(operation *StreamingOperation) error {
	// Simulate analysis with parallel processing
	return ese.processWithChunks(operation, func(chunk []byte, chunkIndex int64) error {
		// Simulate analysis processing
		time.Sleep(50 * time.Millisecond)
		return nil
	})
}

func (ese *EnterpriseStreamingEngine) processValidationOperation(operation *StreamingOperation) error {
	// Simulate fast schema validation
	return ese.processWithChunks(operation, func(chunk []byte, chunkIndex int64) error {
		// Simulate validation processing
		time.Sleep(25 * time.Millisecond)
		return nil
	})
}

func (ese *EnterpriseStreamingEngine) processDiffOperation(operation *StreamingOperation) error {
	// Simulate memory-efficient dataset comparison
	return ese.processWithChunks(operation, func(chunk []byte, chunkIndex int64) error {
		// Simulate diff processing
		time.Sleep(75 * time.Millisecond)
		return nil
	})
}

func (ese *EnterpriseStreamingEngine) processMergeOperation(operation *StreamingOperation) error {
	// Simulate dataset merging
	return ese.processWithChunks(operation, func(chunk []byte, chunkIndex int64) error {
		// Simulate merge processing
		time.Sleep(80 * time.Millisecond)
		return nil
	})
}

func (ese *EnterpriseStreamingEngine) processCompressionOperation(operation *StreamingOperation) error {
	// Simulate compression
	return ese.processWithChunks(operation, func(chunk []byte, chunkIndex int64) error {
		// Simulate compression processing
		time.Sleep(60 * time.Millisecond)
		return nil
	})
}

// processWithChunks processes operation in chunks using parallel workers
func (ese *EnterpriseStreamingEngine) processWithChunks(operation *StreamingOperation, processor func([]byte, int64) error) error {
	file, err := os.Open(operation.SourcePath)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			ese.logger.Info("Warning: failed to close file: " + closeErr.Error())
		}
	}()

	chunkSize := int64(operation.Configuration.ChunkSizeMB) * 1024 * 1024
	chunkIndex := int64(0)

	for {
		select {
		case <-operation.ctx.Done():
			return operation.ctx.Err()
		default:
		}

		// Check if paused
		operation.mu.RLock()
		status := operation.Status
		operation.mu.RUnlock()

		if status == StreamingStatusPaused {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		// Read chunk
		chunk := make([]byte, chunkSize)
		n, err := file.Read(chunk)
		if err != nil {
			if err == io.EOF {
				break // End of file
			}
			return fmt.Errorf("failed to read chunk: %w", err)
		}

		// Process chunk
		if err := processor(chunk[:n], chunkIndex); err != nil {
			return fmt.Errorf("failed to process chunk %d: %w", chunkIndex, err)
		}

		// Update progress
		operation.mu.Lock()
		operation.Progress.BytesProcessed += int64(n)
		operation.Progress.RecordsProcessed += int64(n / 100) // Approximate records
		if operation.Progress.TotalBytes > 0 {
			operation.Progress.PercentComplete = float64(operation.Progress.BytesProcessed) / float64(operation.Progress.TotalBytes) * 100
		}
		operation.Progress.LastUpdated = time.Now()
		operation.mu.Unlock()

		// Create checkpoint if enabled
		if operation.Configuration.CheckpointEnabled && chunkIndex%10 == 0 {
			ese.createCheckpoint(operation, chunkIndex)
		}

		chunkIndex++
	}

	return nil
}

// Helper methods

func (ese *EnterpriseStreamingEngine) validateStreamingRequest(req *StreamingOperationRequest) error {
	if req.OperationID == "" {
		return fmt.Errorf("operation ID is required")
	}
	if req.SourcePath == "" {
		return fmt.Errorf("source path is required")
	}
	if req.DestinationPath == "" {
		return fmt.Errorf("destination path is required")
	}

	// Check if source file exists
	if _, err := os.Stat(req.SourcePath); os.IsNotExist(err) {
		return fmt.Errorf("source file does not exist: %s", req.SourcePath)
	}

	return nil
}

func (ese *EnterpriseStreamingEngine) handleOperationError(operation *StreamingOperation, err error) {
	operation.mu.Lock()
	operation.Status = StreamingStatusFailed
	endTime := time.Now()
	operation.EndTime = &endTime
	operation.ErrorInfo = &StreamingError{
		ErrorCode:    "OPERATION_ERROR",
		ErrorMessage: err.Error(),
		ErrorType:    "system",
		Timestamp:    endTime,
	}
	operation.mu.Unlock()

	ese.logger.Error(fmt.Sprintf("Streaming operation %s failed: %v", operation.ID, err))
}

func (ese *EnterpriseStreamingEngine) monitorOperationPerformance(operation *StreamingOperation) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-operation.ctx.Done():
			return
		case <-ticker.C:
			// Update performance metrics
			operation.mu.Lock()
			if operation.Progress.BytesProcessed > 0 && !operation.StartTime.IsZero() {
				elapsed := time.Since(operation.StartTime).Minutes()
				if elapsed > 0 {
					gbProcessed := float64(operation.Progress.BytesProcessed) / (1024 * 1024 * 1024)
					operation.Performance.ThroughputGBMin = gbProcessed / elapsed
				}
			}
			operation.Performance.LastMeasured = time.Now()
			operation.mu.Unlock()
		}
	}
}

func (ese *EnterpriseStreamingEngine) reportOperationProgress(operation *StreamingOperation) {
	ticker := time.NewTicker(time.Duration(ese.config.ProgressReportInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-operation.ctx.Done():
			return
		case <-ticker.C:
			operation.mu.RLock()
			progress := operation.Progress.PercentComplete
			throughput := operation.Performance.ThroughputGBMin
			operation.mu.RUnlock()

			ese.logger.Info(fmt.Sprintf("Operation %s progress: %.1f%% (%.2f GB/min)",
				operation.ID, progress, throughput))
		}
	}
}

func (ese *EnterpriseStreamingEngine) createCheckpoint(operation *StreamingOperation, chunkIndex int64) {
	checkpoint := StreamingCheckpoint{
		ID:             fmt.Sprintf("%s-checkpoint-%d", operation.ID, chunkIndex),
		Position:       chunkIndex,
		RecordsCount:   operation.Progress.RecordsProcessed,
		BytesProcessed: operation.Progress.BytesProcessed,
		Timestamp:      time.Now(),
		ValidationHash: fmt.Sprintf("hash-%d", chunkIndex), // Placeholder
	}

	operation.mu.Lock()
	operation.Checkpoints = append(operation.Checkpoints, checkpoint)
	operation.mu.Unlock()

	ese.logger.Debug(fmt.Sprintf("Created checkpoint %s for operation %s", checkpoint.ID, operation.ID))
}

// Stop gracefully shuts down the streaming engine
func (ese *EnterpriseStreamingEngine) Stop() error {
	ese.logger.Info("Stopping enterprise streaming engine")

	// Cancel all active operations
	ese.mu.RLock()
	operations := make([]*StreamingOperation, 0, len(ese.activeStreams))
	for _, op := range ese.activeStreams {
		operations = append(operations, op)
	}
	ese.mu.RUnlock()

	for _, operation := range operations {
		operation.cancel()
	}

	// Stop performance engine
	return ese.performanceEngine.Stop()
}

// Compile-time guarantee that enterprise implementation satisfies the reduced
// public StreamingEngine interface used by HTTP handlers. Note: methods like
// PauseStreamingOperation and Stop() are intentionally NOT part of the
// interface surface today because no public endpoints invoke them. If in the
// future pause / stop (per-operation or global engine shutdown wiring) is
// exposed via API, extend StreamingEngine additively and implement matching
// no-op + metrics logic in the non-enterprise stub to preserve SemVer
// compatibility.
var _ StreamingEngine = (*EnterpriseStreamingEngine)(nil)
