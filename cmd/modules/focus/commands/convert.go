package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"local/costscope/internal/core/config"
	"local/costscope/internal/core/focus/conversion"
	"local/costscope/internal/core/focus/quality"
	"local/costscope/internal/core/focus/types"
	"local/costscope/internal/core/logging"
)

// Config precedence resolution now uses stateless Resolve*Field helpers (no legacy resolver struct).

// =====================================================================================
// FOCUS Convert Command - Core CostScope Functionality
// =====================================================================================
// Main CLI command for converting cloud billing data to FOCUS v1.2 format

var (
	// Provider settings
	convertProvider string

	// Input/Output settings
	convertInput  string
	convertOutput string

	// Processing options
	convertStreaming  bool
	convertValidate   bool
	convertAnalyze    bool
	convertSubmitOnly bool
	convertQuiet      bool
	convertVerbose    bool

	// Performance options
	convertChunkSize   int
	convertWorkers     int
	convertMaxMemory   int
	convertCompression bool
	convertProgress    bool
	// Monitoring control
	convertMonitorTimeout string

	// Experimental flags
	convertUseUnifiedMapper bool

	// Invariants / data quality flags
	convertInvariantsEnabled   bool
	convertInvariantsBaseline  string
	convertInvariantsTolerance float64
	convertInvariantsReport    string
	convertInvariantsFail      bool

	// Parquet-specific options
	parquetCompression   string
	parquetRowGroupBytes int64
	parquetPageSizeBytes int64
	parquetRotateBytes   int64
	parquetRotateEvery   string
	parquetFilePrefix    string

	// Batch processing
	convertInputDir  string
	convertOutputDir string
)

// BuildConvertCommand creates the convert command
func BuildConvertCommand() *cobra.Command {
	convertCmd := &cobra.Command{
		Use:   "convert",
		Short: "Convert cloud billing data to FOCUS format",
		Long: `Convert cloud billing data from various providers (AWS, Azure, GCP) to the 
FOCUS (FinOps Open Cloud Usage Specification) format.

This command supports multiple input formats and provides efficient processing
for large datasets through streaming and parallel processing.

FOCUS v1.2 Compliance:
- Full schema compliance with FOCUS v1.2 specification
- Streaming processing for files >10GB
- Parallel workers for optimal performance
- Data quality validation and error reporting

Examples:
  # Convert AWS CUR data
  costscope convert --provider aws --input manifest.json --output focus.parquet

  # Convert Azure cost export with streaming
  costscope convert --provider azure --input export.csv --output focus.parquet --streaming

  # Convert GCP BigQuery export with workers
  costscope convert --provider gcp --input bq-export.csv --output focus.parquet --workers 8

  # Batch convert directory
  costscope convert batch --input-dir ./billing-data/ --output-dir ./focus/ --provider aws

Supported Providers:
  - aws: AWS Cost and Usage Reports (CUR)
  - azure: Azure Cost Management exports  
  - gcp: GCP BigQuery billing exports`,
		RunE: runConvert,
	}

	// Provider selection
	convertCmd.Flags().StringVarP(&convertProvider, "provider", "p", "", "Cloud provider (aws, azure, gcp)")
	_ = convertCmd.MarkFlagRequired("provider")

	// Input/Output
	convertCmd.Flags().StringVarP(&convertInput, "input", "i", "", "Input file, directory, or URI")
	convertCmd.Flags().StringVarP(&convertOutput, "output", "o", "", "Output file path")

	// Processing options
	convertCmd.Flags().BoolVar(&convertStreaming, "streaming", false, "Use streaming processing for large files")
	convertCmd.Flags().BoolVar(&convertValidate, "validate", true, "Validate input before conversion")
	convertCmd.Flags().BoolVar(&convertAnalyze, "analyze", false, "Analyze input and display metadata")
	convertCmd.Flags().BoolVar(&convertSubmitOnly, "submit-only", false, "Submit job and return immediately without monitoring (experimental; job runs in-process)")

	// Performance options
	convertCmd.Flags().IntVar(&convertChunkSize, "chunk-size", 10000, "Chunk size for streaming processing")
	convertCmd.Flags().IntVar(&convertWorkers, "workers", 4, "Number of parallel workers")
	convertCmd.Flags().IntVar(&convertMaxMemory, "max-memory", 1024, "Maximum memory usage in MB")
	convertCmd.Flags().BoolVar(&convertCompression, "compression", true, "Enable output compression")

	// Parquet-specific options
	convertCmd.Flags().StringVar(&parquetCompression, "parquet-compression", "snappy", "Parquet compression codec (snappy|zstd|gzip|uncompressed)")
	convertCmd.Flags().Int64Var(&parquetRowGroupBytes, "row-group-size", 0, "Parquet row group size in bytes (default 134217728 for 128MB if 0)")
	convertCmd.Flags().Int64Var(&parquetPageSizeBytes, "page-size", 0, "Parquet page size in bytes (default library value if 0)")
	convertCmd.Flags().Int64Var(&parquetRotateBytes, "rotate-size", 0, "Rotate output Parquet file when it reaches this size in bytes (default 536870912 for 512MB if 0 and rotation enabled)")
	convertCmd.Flags().StringVar(&parquetRotateEvery, "rotate-interval", "", "Time-based rotation interval (e.g. 5m, 1h); empty to disable")
	convertCmd.Flags().StringVar(&parquetFilePrefix, "file-prefix", "", "Prefix for rotated Parquet files (defaults to output basename)")

	// Output options
	convertCmd.Flags().BoolVar(&convertProgress, "progress", true, "Show progress indicators")
	convertCmd.Flags().StringVar(&convertMonitorTimeout, "monitor-timeout", "30s", "Maximum time to monitor conversion before timing out (e.g. 30s, 2m)")
	convertCmd.Flags().BoolVarP(&convertVerbose, "verbose", "v", false, "Verbose output")
	convertCmd.Flags().BoolVarP(&convertQuiet, "quiet", "q", false, "Quiet output (errors only)")

	// Hidden experimental flag: unified mapper
	convertCmd.Flags().BoolVar(&convertUseUnifiedMapper, "use-unified-mapper", false, "EXPERIMENTAL: Use unified field mapper instead of legacy rules (may change output)")
	_ = convertCmd.Flags().MarkHidden("use-unified-mapper")

	// Invariants flags (lightweight quality invariants aggregation)
	convertCmd.Flags().BoolVar(&convertInvariantsEnabled, "invariants", false, "Compute lightweight post-conversion invariants (aggregates & distributions)")
	convertCmd.Flags().StringVar(&convertInvariantsBaseline, "invariants-baseline", "", "Path to invariants baseline JSON for drift comparison (optional)")
	convertCmd.Flags().Float64Var(&convertInvariantsTolerance, "invariants-tolerance", 0.01, "Relative tolerance for aggregate drift (e.g. 0.01 = 1%)")
	convertCmd.Flags().StringVar(&convertInvariantsReport, "invariants-report", "", "Write computed invariants JSON report to this path (optional)")
	convertCmd.Flags().BoolVar(&convertInvariantsFail, "fail-on-invariants", false, "Fail conversion if invariants drift violations detected")

	// Add batch subcommand
	convertCmd.AddCommand(buildBatchConvertCommand())

	return convertCmd
}

// buildBatchConvertCommand creates the batch convert subcommand
func buildBatchConvertCommand() *cobra.Command {
	batchCmd := &cobra.Command{
		Use:   "batch",
		Short: "Batch convert multiple files",
		Long: `Batch convert multiple cloud billing files to FOCUS format.

This command processes all files in an input directory and converts them
to FOCUS format in the output directory.

Examples:
  costscope convert batch --input-dir ./cur-data/ --output-dir ./focus/ --provider aws
  costscope convert batch --input-dir ./azure-exports/ --output-dir ./focus/ --provider azure --streaming`,
		RunE: runBatchConvert,
	}

	// Batch-specific flags
	batchCmd.Flags().StringVar(&convertInputDir, "input-dir", "", "Input directory containing billing files")
	batchCmd.Flags().StringVar(&convertOutputDir, "output-dir", "", "Output directory for FOCUS files")
	_ = batchCmd.MarkFlagRequired("input-dir")
	_ = batchCmd.MarkFlagRequired("output-dir")

	return batchCmd
}

// runConvert executes the main convert command
func runConvert(cmd *cobra.Command, args []string) error {
	logger := logging.NewLogger(getLogLevel())
	logger.Info("Starting FOCUS conversion process")

	// Validate flags
	if err := validateConvertFlags(); err != nil {
		return fmt.Errorf("invalid flags: %w", err)
	}

	// Precedence: CLI flag (explicit) > YAML default > ENV > fallback(false)
	var explicitPtr *bool
	if cmd.Flags().Changed("use-unified-mapper") {
		explicitPtr = &convertUseUnifiedMapper
	}
	res := config.ResolveBoolField(logger, "focus.use_unified_mapper", explicitPtr, func(cc *config.ConsolidatedConfig) *bool {
		if cc == nil {
			return nil
		}
		return &cc.Focus.UseUnifiedMapperDefault
	}, "COSTSCOPE_USE_UNIFIED_MAPPER", false)
	useUnified := res.Value

	// Invariants precedence using unified resolver
	var invariantsExplicit *bool
	if cmd.Flags().Changed("invariants") {
		invariantsExplicit = &convertInvariantsEnabled
	}
	invEnabledRes := config.ResolveBoolField(logger, "focus.invariants_enabled", invariantsExplicit, func(cc *config.ConsolidatedConfig) *bool {
		if cc == nil {
			return nil
		}
		return &cc.Focus.InvariantsEnabledDefault
	}, "COSTSCOPE_INVARIANTS_ENABLED", false)
	invariantsEnabled := invEnabledRes.Value

	// Invariants tolerance precedence (float). Only apply if invariants are enabled (still resolve/log source).
	var tolExplicit *float64
	if cmd.Flags().Changed("invariants-tolerance") {
		tolExplicit = &convertInvariantsTolerance
	}
	tolRes := config.ResolveFloatField(logger, "focus.invariants_tolerance", tolExplicit, func(cc *config.ConsolidatedConfig) *float64 {
		if cc == nil {
			return nil
		}
		return &cc.Focus.InvariantsToleranceDefault
	}, "COSTSCOPE_INVARIANTS_TOLERANCE", 0.01)
	invariantsTolerance := tolRes.Value

	// Baseline path precedence (string)
	var baseExplicit *string
	if cmd.Flags().Changed("invariants-baseline") {
		baseExplicit = &convertInvariantsBaseline
	}
	baseRes := config.ResolveStringField(logger, "focus.invariants_baseline", baseExplicit, func(cc *config.ConsolidatedConfig) *string {
		if cc == nil || cc.Focus.InvariantsBaselineDefault == "" {
			return nil
		}
		return &cc.Focus.InvariantsBaselineDefault
	}, "COSTSCOPE_INVARIANTS_BASELINE", "")
	invariantsBaseline := baseRes.Value

	// Create conversion configuration
	config := &types.ConversionConfig{
		Provider:            convertProvider,
		InputPath:           convertInput,
		OutputPath:          convertOutput,
		Streaming:           convertStreaming,
		ChunkSize:           convertChunkSize,
		Workers:             convertWorkers,
		MaxMemoryMB:         convertMaxMemory,
		Compression:         convertCompression,
		ValidateInput:       convertValidate,
		ConversionId:        fmt.Sprintf("conv_%d", time.Now().Unix()),
		CreatedAt:           time.Now(),
		CreatedBy:           "CostScope CLI",
		UseUnifiedMapper:    useUnified,
		InvariantsEnabled:   invariantsEnabled,
		InvariantsBaseline:  invariantsBaseline,
		InvariantsTolerance: invariantsTolerance,
	}

	// Populate Parquet options from flags
	config.Parquet.CompressionCodec = parquetCompression
	config.Parquet.RowGroupSizeBytes = parquetRowGroupBytes
	config.Parquet.PageSizeBytes = parquetPageSizeBytes
	config.Parquet.RotateSizeBytes = parquetRotateBytes
	config.Parquet.RotateInterval = parquetRotateEvery
	config.Parquet.FilePrefix = parquetFilePrefix

	// Create / obtain conversion manager. For submit-only asynchronous mode we must
	// use the shared in-process manager so that subsequent 'focus jobs *' commands
	// can observe and manage the job lifecycle. For synchronous (non submit-only)
	// runs we keep an isolated manager instance to avoid cross-talk with any
	// concurrently monitored jobs.
	var converterManager *conversion.ConversionManager
	if convertSubmitOnly {
		converterManager = getSharedConversionManager()
	} else {
		converterManager = conversion.NewConfiguredConversionManager(1)
	}

	// Register provider-specific converters
	if err := registerConverters(converterManager); err != nil {
		return fmt.Errorf("failed to register converters: %w", err)
	}

	// Perform input analysis if requested
	if convertAnalyze {
		if err := analyzeInput(config, logger); err != nil {
			logger.Error(fmt.Sprintf("Input analysis failed: %v", err))
		}
	}

	// Execute conversion
	jobId, err := converterManager.SubmitJob(config)
	if err != nil {
		return fmt.Errorf("failed to submit conversion job: %w", err)
	}

	logger.Info(fmt.Sprintf("Conversion job submitted: %s", jobId))

	// If submit-only, just print the job id and exit; note that job continues running
	if convertSubmitOnly {
		if _, werr := fmt.Fprintf(cmd.OutOrStdout(), "job_id=%s status=accepted (monitoring disabled)\n", jobId); werr != nil {
			return werr
		}
		return nil
	}

	// Parse monitor timeout
	mt := 30 * time.Second
	if convertMonitorTimeout != "" {
		if dur, err := time.ParseDuration(convertMonitorTimeout); err == nil && dur > 0 {
			mt = dur
		} else if err != nil {
			logger.Warn(fmt.Sprintf("invalid monitor-timeout value %q, using default 30s: %v", convertMonitorTimeout, err))
		}
	}

	// Monitor job progress synchronously with configured timeout
	return monitorConversionJob(converterManager, jobId, logger, mt)
}

// runBatchConvert executes batch conversion
func runBatchConvert(cmd *cobra.Command, args []string) error {
	logger := logging.NewLogger(getLogLevel())
	logger.Info("Starting FOCUS batch conversion")

	// Validate directories
	if err := validateBatchDirectories(); err != nil {
		return err
	}

	// Find input files
	inputFiles := findInputFiles(convertInputDir, convertProvider)
	if len(inputFiles) == 0 {
		return fmt.Errorf("no input files found in %s", convertInputDir)
	}

	logger.Info(fmt.Sprintf("Found %d files for batch conversion", len(inputFiles)))

	// Precedence for batch identical
	var explicitPtr *bool
	if cmd.Flags().Changed("use-unified-mapper") {
		explicitPtr = &convertUseUnifiedMapper
	}
	res := config.ResolveBoolField(logger, "focus.use_unified_mapper", explicitPtr, func(cc *config.ConsolidatedConfig) *bool {
		if cc == nil {
			return nil
		}
		return &cc.Focus.UseUnifiedMapperDefault
	}, "COSTSCOPE_USE_UNIFIED_MAPPER", false)
	useUnified := res.Value

	// Create conversion configs for each file
	configs := make([]*types.ConversionConfig, 0, len(inputFiles))
	for _, inputFile := range inputFiles {
		outputFile := generateOutputPath(inputFile, convertInputDir, convertOutputDir)

		config := &types.ConversionConfig{
			Provider:         convertProvider,
			InputPath:        inputFile,
			OutputPath:       outputFile,
			Streaming:        convertStreaming,
			ChunkSize:        convertChunkSize,
			Workers:          1, // Single worker per file in batch mode
			MaxMemoryMB:      convertMaxMemory,
			Compression:      convertCompression,
			ValidateInput:    convertValidate,
			ConversionId:     fmt.Sprintf("batch_%d_%s", time.Now().Unix(), filepath.Base(inputFile)),
			CreatedAt:        time.Now(),
			CreatedBy:        "CostScope CLI Batch",
			UseUnifiedMapper: useUnified,
		}
		configs = append(configs, config)
	}

	// Create converter manager with multiple workers for batch
	converterManager := conversion.NewConfiguredConversionManager(convertWorkers)

	// Register converters
	if err := registerConverters(converterManager); err != nil {
		return fmt.Errorf("failed to register converters: %w", err)
	}

	// Execute batch conversion
	return executeBatchConversion(converterManager, configs, logger)
}

// =====================================================================================
// Helper Functions
// =====================================================================================

// validateConvertFlags validates convert command flags
func validateConvertFlags() error {
	if convertProvider == "" {
		return fmt.Errorf("provider is required")
	}

	if !isProviderSupported(convertProvider) {
		return fmt.Errorf("unsupported provider: %s (supported: aws, azure, gcp)", convertProvider)
	}

	if convertInput == "" {
		return fmt.Errorf("input is required")
	}

	if convertOutput == "" {
		return fmt.Errorf("output is required")
	}

	// Validate input file exists
	if _, err := os.Stat(convertInput); os.IsNotExist(err) {
		return fmt.Errorf("input file does not exist: %s", convertInput)
	}

	// Validate output directory
	outputDir := filepath.Dir(convertOutput)
	if err := os.MkdirAll(outputDir, 0750); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Parquet flags validation (only when parquet output requested by extension or later selection)
	if strings.EqualFold(filepath.Ext(convertOutput), ".parquet") || strings.EqualFold(strings.TrimSpace(filepath.Ext(convertOutput)), ".pq") {
		// Compression codec validation
		switch strings.ToLower(parquetCompression) {
		case "snappy", "zstd", "gzip", "uncompressed", "none":
			// ok; normalize "none" -> "uncompressed"
			if parquetCompression == "none" {
				parquetCompression = "uncompressed"
			}
		default:
			return fmt.Errorf("invalid parquet-compression: %s (supported: snappy|zstd|gzip|uncompressed)", parquetCompression)
		}

		// Sizes must be >= 0
		if parquetRowGroupBytes < 0 || parquetPageSizeBytes < 0 || parquetRotateBytes < 0 {
			return fmt.Errorf("row-group-size, page-size, and rotate-size must be >= 0")
		}

		// rotate-interval, if set, must parse
		if strings.TrimSpace(parquetRotateEvery) != "" {
			if _, err := time.ParseDuration(parquetRotateEvery); err != nil {
				return fmt.Errorf("invalid rotate-interval: %v", err)
			}
		}
	}

	return nil
}

// validateBatchDirectories validates batch conversion directories
func validateBatchDirectories() error {
	// Check input directory exists
	if _, err := os.Stat(convertInputDir); os.IsNotExist(err) {
		return fmt.Errorf("input directory does not exist: %s", convertInputDir)
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(convertOutputDir, 0750); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	return nil
}

// findInputFiles finds input files for batch processing
func findInputFiles(inputDir, provider string) []string {
	var patterns []string

	switch provider {
	case "aws":
		patterns = []string{"*.csv", "*.json", "*manifest.json"}
	case "azure":
		patterns = []string{"*.csv", "*.json"}
	case "gcp":
		patterns = []string{"*.csv", "*.json"}
	default:
		patterns = []string{"*.csv", "*.json"}
	}

	var files []string
	for _, pattern := range patterns {
		matches, err := filepath.Glob(filepath.Join(inputDir, pattern))
		if err != nil {
			continue
		}
		files = append(files, matches...)
	}

	return files
}

// generateOutputPath generates output path for batch conversion
func generateOutputPath(inputFile, inputDir, outputDir string) string {
	relPath, _ := filepath.Rel(inputDir, inputFile)
	baseName := strings.TrimSuffix(relPath, filepath.Ext(relPath))
	return filepath.Join(outputDir, baseName+".parquet")
}

// isProviderSupported checks if provider is supported
func isProviderSupported(provider string) bool {
	supportedProviders := []string{"aws", "azure", "gcp"}
	for _, p := range supportedProviders {
		if strings.EqualFold(provider, p) {
			return true
		}
	}
	return false
}

// getLogLevel returns log level based on flags
func getLogLevel() logging.LogLevel {
	if convertQuiet {
		return logging.LevelError
	}
	if convertVerbose {
		return logging.LevelDebug
	}
	return logging.LevelInfo
}

// registerConverters registers provider-specific converters
func registerConverters(_ *conversion.ConversionManager) error {
	// Converters are automatically registered in NewConversionManager
	// Additional converters can be registered here if needed
	return nil
}

// analyzeInput analyzes input file and displays metadata
func analyzeInput(config *types.ConversionConfig, logger *logging.Logger) error {
	logger.Info("Analyzing input file...")

	fileInfo, err := os.Stat(config.InputPath)
	if err != nil {
		return err
	}

	// Display file information
	logger.Info("Input Analysis Results:")
	logger.Info(fmt.Sprintf("  File: %s", config.InputPath))
	logger.Info(fmt.Sprintf("  Size: %.2f MB", float64(fileInfo.Size())/(1024*1024)))
	logger.Info(fmt.Sprintf("  Provider: %s", config.Provider))
	logger.Info(fmt.Sprintf("  Format: %s", strings.ToUpper(filepath.Ext(config.InputPath))))
	logger.Info(fmt.Sprintf("  Modified: %s", fileInfo.ModTime().Format("2006-01-02 15:04:05")))

	// TODO: Add provider-specific analysis
	// - Record count estimation
	// - Schema validation
	// - Data quality checks

	return nil
}

// monitorConversionJob monitors conversion job progress
func monitorConversionJob(manager *conversion.ConversionManager, jobId string, logger *logging.Logger, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()
	// immediate first poll without waiting for ticker
	for first := true; ; first = false {
		if !first {
			select {
			case <-ticker.C:
			case <-time.After(time.Until(deadline)):
				return fmt.Errorf("conversion monitoring timed out")
			}
		}
		job, err := manager.GetJobStatus(jobId)
		if err != nil {
			return fmt.Errorf("failed to get job status: %w", err)
		}
		if convertProgress && job.Progress != nil {
			showProgress(job.Progress, logger)
		}
		done, perr := processConversionStatus(manager, job, jobId, logger)
		if done {
			return perr
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("conversion monitoring timed out")
		}
	}
}

// processConversionStatus handles terminal states and invariants.
func processConversionStatus(manager *conversion.ConversionManager, job *conversion.ConversionJob, jobId string, logger *logging.Logger) (bool, error) {
	switch job.Status {
	case types.StatusCompleted:
		logger.Info("Conversion completed successfully!")
		if err := handleConversionInvariants(job.Result, logger); err != nil {
			manager.RemoveCompletedJob(jobId)
			return true, err
		}
		displayConversionResults(job.Result, logger)
		manager.RemoveCompletedJob(jobId)
		return true, nil
	case types.StatusFailed:
		logger.Error("Conversion failed!")
		if job.Progress != nil && job.Progress.LastError != "" {
			logger.Error(fmt.Sprintf("Error: %s", job.Progress.LastError))
		}
		manager.RemoveCompletedJob(jobId)
		return true, fmt.Errorf("conversion job failed")
	case types.StatusCancelled:
		logger.Info("Conversion was cancelled")
		return true, fmt.Errorf("conversion job cancelled")
	default:
		return false, nil
	}
}

// handleConversionInvariants isolates invariants logic from monitor loop.
func handleConversionInvariants(result *types.ConversionResult, logger *logging.Logger) error { //nolint:cyclop
	if result == nil || result.Invariants == nil || !convertInvariantsEnabled {
		return nil
	}
	var cur quality.InvariantMetrics
	switch invTyped := result.Invariants.(type) {
	case map[string]any:
		b, _ := json.Marshal(invTyped)
		_ = json.Unmarshal(b, &cur)
	case quality.InvariantMetrics:
		cur = invTyped
	case *quality.InvariantMetrics:
		cur = *invTyped
	default:
		b, _ := json.Marshal(invTyped)
		_ = json.Unmarshal(b, &cur)
	}
	if convertInvariantsBaseline != "" {
		if baseline, err := quality.LoadBaseline(convertInvariantsBaseline); err != nil {
			logger.Warn(fmt.Sprintf("Failed to load invariants baseline: %v", err))
		} else {
			quality.CompareInvariants(&cur, baseline, convertInvariantsTolerance)
			if len(cur.Violations) > 0 {
				logger.Warn(fmt.Sprintf("Invariants drift violations: %d", len(cur.Violations)))
				for _, v := range cur.Violations {
					logger.Warn("  " + v)
				}
				// Defer failure until after we optionally write the report below
				// by recording a flag and returning at end of function.
				// Failure handled after report write below; no action needed here.
			}
		}
	}
	bCur, _ := json.Marshal(cur)
	var curMap map[string]any
	_ = json.Unmarshal(bCur, &curMap)
	result.Invariants = curMap
	if convertInvariantsReport != "" {
		b, _ := json.MarshalIndent(curMap, "", "  ")
		if err := os.WriteFile(convertInvariantsReport, b, 0o600); err != nil { //nolint:gosec
			logger.Warn(fmt.Sprintf("Failed to write invariants report: %v", err))
		} else {
			logger.Info(fmt.Sprintf("Invariants report written: %s", convertInvariantsReport))
		}
	}
	// Emit a compact summary log (avoids dumping full JSON while giving key signals)
	logger.Info(fmt.Sprintf(
		"Invariants summary: row_count=%d eff_cost=%.4f list_cost=%.4f usage_qty=%.4f violations=%d",
		cur.RowCount, cur.SumEffectiveCost, cur.SumListCost, cur.SumUsageQuantity, len(cur.Violations)))
	// If fail-on-invariants is set and we had violations, return error now (after report write)
	if convertInvariantsFail && len(cur.Violations) > 0 {
		return fmt.Errorf("invariants drift violations detected")
	}
	return nil
}

// executeBatchConversion executes batch conversion with progress tracking
func executeBatchConversion(manager *conversion.ConversionManager, configs []*types.ConversionConfig, logger *logging.Logger) error {
	logger.Info(fmt.Sprintf("Starting batch conversion of %d files", len(configs)))

	// Submit all jobs
	jobIds := make([]string, 0, len(configs))
	for _, config := range configs {
		jobId, err := manager.SubmitJob(config)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to submit job for %s: %v", config.InputPath, err))
			continue
		}
		jobIds = append(jobIds, jobId)
	}

	logger.Info(fmt.Sprintf("Submitted %d conversion jobs", len(jobIds)))

	// Monitor all jobs
	completedJobs := 0
	failedJobs := 0
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		activeJobs := manager.ListActiveJobs()
		if len(activeJobs) == 0 {
			break // All jobs completed
		}

		// Count statuses
		completed := 0
		failed := 0
		for _, job := range activeJobs {
			switch job.Status {
			case types.StatusCompleted:
				completed++
			case types.StatusFailed:
				failed++
			}
		}

		if completed+failed == len(jobIds) {
			completedJobs = completed
			failedJobs = failed
			break // All jobs finished
		}

		// Show batch progress
		logger.Info(fmt.Sprintf("Batch progress: %d completed, %d failed, %d active",
			completed, failed, len(activeJobs)-(completed+failed)))

		if completedJobs+failedJobs == len(jobIds) {
			break
		}
	}

	// Display batch results
	logger.Info(fmt.Sprintf("Batch conversion completed: %d successful, %d failed, %d total",
		completedJobs, failedJobs, len(configs)))

	if failedJobs > 0 {
		return fmt.Errorf("batch conversion completed with %d failures", failedJobs)
	}

	return nil
}

// showProgress displays lightweight progress information (no invariants handling here)
func showProgress(progress *types.ConversionProgress, logger *logging.Logger) {
	if progress == nil {
		return
	}
	logger.Info(fmt.Sprintf(
		"Progress: status=%s processed=%d success=%d errors=%d skipped=%d rps=%.1f elapsed=%s",
		progress.Status,
		progress.ProcessedRecords,
		progress.SuccessRecords,
		progress.ErrorRecords,
		progress.SkippedRecords,
		progress.RecordsPerSecond,
		time.Since(progress.StartTime).Truncate(time.Second),
	))
}

// displayConversionResults logs key metrics for a completed conversion.
// (Previously displayed more verbose information; trimmed for clarity.)
func displayConversionResults(result *types.ConversionResult, logger *logging.Logger) {
	if result == nil {
		return
	}
	logger.Info(fmt.Sprintf(
		"Result: success=%v input_records=%d output_records=%d effective_duration=%s input=%s output=%s",
		result.Success,
		result.InputRecords,
		result.OutputRecords,
		result.Duration,
		result.InputFile,
		result.OutputFile,
	))
}
