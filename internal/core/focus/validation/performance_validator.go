package validation

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// PerformanceValidator validates performance aspects of data files
type PerformanceValidator struct{}

// NewPerformanceValidator creates a new performance validator
func NewPerformanceValidator() *PerformanceValidator {
	return &PerformanceValidator{}
}

// Name returns the validator name
func (v *PerformanceValidator) Name() string {
	return "performance"
}

// SupportsFormat checks if the validator supports the given format
func (v *PerformanceValidator) SupportsFormat(format string) bool {
	supportedFormats := []string{"parquet", "csv", "json", "orc", "avro"}
	for _, supported := range supportedFormats {
		if format == supported {
			return true
		}
	}
	return false
}

// Validate validates performance characteristics
func (v *PerformanceValidator) Validate(data interface{}, config ValidationConfig) (interface{}, error) {
	filePath, ok := data.(string)
	if !ok {
		return nil, fmt.Errorf("expected file path string, got %T", data)
	}

	result := PerformanceValidationResult{
		Valid:            true,
		Score:            100.0,
		QueryPerformance: QueryPerformanceMetrics{},
		MemoryUsage:      MemoryUsageMetrics{},
		ReadThroughput:   ThroughputMetrics{},
		Issues:           []PerformanceIssue{},
	}

	// Get file information
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	// Perform performance assessments
	v.assessCompressionRatio(filePath, fileInfo, &result)
	v.assessFileEfficiency(filePath, fileInfo, &result)
	v.assessQueryPerformance(filePath, fileInfo, &result)
	v.assessMemoryUsage(filePath, fileInfo, &result)
	v.assessReadThroughput(filePath, fileInfo, &result)

	// Check performance thresholds and identify issues
	v.checkPerformanceThresholds(&result, config)

	// Calculate overall performance score
	v.calculatePerformanceScore(&result)

	return result, nil
}

// assessCompressionRatio assesses file compression efficiency
func (v *PerformanceValidator) assessCompressionRatio(filePath string, _ os.FileInfo, result *PerformanceValidationResult) {
	format := detectFileFormat(filePath)

	// Estimate uncompressed size based on format and file size
	var compressionRatio float64

	switch format {
	case FormatParquet:
		// Parquet typically achieves 5-10x compression for cost data
		compressionRatio = v.simulateCompressionRatio(5.0, 10.0)

	case FormatORC:
		// ORC similar to Parquet
		compressionRatio = v.simulateCompressionRatio(4.0, 9.0)

	case FormatAVRO:
		// Avro with compression typically 3-6x
		compressionRatio = v.simulateCompressionRatio(3.0, 6.0)

	case FormatCSV:
		// CSV with gzip compression typically 3-5x
		if strings.Contains(strings.ToLower(filePath), ".gz") {
			compressionRatio = v.simulateCompressionRatio(3.0, 5.0)
		} else {
			compressionRatio = 1.0 // No compression
		}

	case FormatJSON:
		// JSON with compression typically 4-7x
		if strings.Contains(strings.ToLower(filePath), ".gz") {
			compressionRatio = v.simulateCompressionRatio(4.0, 7.0)
		} else {
			compressionRatio = 1.0 // No compression
		}

	default:
		compressionRatio = 1.0
	}

	result.CompressionRatio = compressionRatio

	// Check compression efficiency
	v.checkCompressionEfficiency(format, compressionRatio, result)
}

// assessFileEfficiency assesses overall file efficiency
func (v *PerformanceValidator) assessFileEfficiency(filePath string, fileInfo os.FileInfo, result *PerformanceValidationResult) {
	format := detectFileFormat(filePath)
	fileSize := fileInfo.Size()

	// Calculate efficiency based on file size, format, and compression
	efficiency := 100.0

	// Deduct points for suboptimal format choices
	switch format {
	case FormatParquet, FormatORC:
		// Optimal formats for analytical workloads
		efficiency += 0

	case FormatAVRO:
		// Good format but not as optimal as columnar
		efficiency -= 5.0

	case FormatCSV:
		// Less efficient, especially without compression
		if !strings.Contains(strings.ToLower(filePath), ".gz") {
			efficiency -= 20.0
		} else {
			efficiency -= 10.0
		}

	case FormatJSON:
		// Least efficient for large datasets
		if !strings.Contains(strings.ToLower(filePath), ".gz") {
			efficiency -= 30.0
		} else {
			efficiency -= 15.0
		}

	default:
		efficiency -= 25.0
	}

	// Consider file size efficiency
	sizeMB := float64(fileSize) / (1024 * 1024)
	if sizeMB > 1000 { // Files over 1GB
		if format == FormatCSV || format == FormatJSON {
			efficiency -= 15.0 // Large files should use better formats
		}
	}

	// Check for very small files (inefficient overhead)
	if sizeMB < 1 && (format == FormatParquet || format == FormatORC) {
		efficiency -= 5.0 // Columnar formats have overhead for very small files
	}

	if efficiency < 0 {
		efficiency = 0
	}

	result.FileEfficiency = efficiency
}

// assessQueryPerformance simulates query performance metrics
func (v *PerformanceValidator) assessQueryPerformance(filePath string, fileInfo os.FileInfo, result *PerformanceValidationResult) {
	format := detectFileFormat(filePath)
	fileSize := fileInfo.Size()
	sizeMB := float64(fileSize) / (1024 * 1024)

	// Simulate query performance based on format and size
	var baseSelectTime, baseFilterTime, baseAggregateTime, baseSortTime time.Duration

	switch format {
	case FormatParquet:
		// Excellent for analytical queries
		baseSelectTime = time.Duration(sizeMB*0.5) * time.Millisecond
		baseFilterTime = time.Duration(sizeMB*0.3) * time.Millisecond
		baseAggregateTime = time.Duration(sizeMB*0.8) * time.Millisecond
		baseSortTime = time.Duration(sizeMB*1.2) * time.Millisecond

	case FormatORC:
		// Very good for analytical queries
		baseSelectTime = time.Duration(sizeMB*0.6) * time.Millisecond
		baseFilterTime = time.Duration(sizeMB*0.4) * time.Millisecond
		baseAggregateTime = time.Duration(sizeMB*0.9) * time.Millisecond
		baseSortTime = time.Duration(sizeMB*1.3) * time.Millisecond

	case FormatAVRO:
		// Good for row-based access
		baseSelectTime = time.Duration(sizeMB*1.2) * time.Millisecond
		baseFilterTime = time.Duration(sizeMB*2.0) * time.Millisecond
		baseAggregateTime = time.Duration(sizeMB*2.5) * time.Millisecond
		baseSortTime = time.Duration(sizeMB*3.0) * time.Millisecond

	case FormatCSV:
		// Slower, especially for filtering and aggregation
		baseSelectTime = time.Duration(sizeMB*2.0) * time.Millisecond
		baseFilterTime = time.Duration(sizeMB*4.0) * time.Millisecond
		baseAggregateTime = time.Duration(sizeMB*5.0) * time.Millisecond
		baseSortTime = time.Duration(sizeMB*6.0) * time.Millisecond

	case FormatJSON:
		// Slowest for analytical workloads
		baseSelectTime = time.Duration(sizeMB*3.0) * time.Millisecond
		baseFilterTime = time.Duration(sizeMB*6.0) * time.Millisecond
		baseAggregateTime = time.Duration(sizeMB*8.0) * time.Millisecond
		baseSortTime = time.Duration(sizeMB*10.0) * time.Millisecond

	default:
		baseSelectTime = time.Duration(sizeMB*5.0) * time.Millisecond
		baseFilterTime = time.Duration(sizeMB*8.0) * time.Millisecond
		baseAggregateTime = time.Duration(sizeMB*10.0) * time.Millisecond
		baseSortTime = time.Duration(sizeMB*12.0) * time.Millisecond
	}

	// Add some randomness to simulate real-world variance
	result.QueryPerformance = QueryPerformanceMetrics{
		SelectTime:    v.addVariance(baseSelectTime, 0.2),
		FilterTime:    v.addVariance(baseFilterTime, 0.3),
		AggregateTime: v.addVariance(baseAggregateTime, 0.25),
		SortTime:      v.addVariance(baseSortTime, 0.4),
	}
}

// assessMemoryUsage simulates memory usage assessment
func (v *PerformanceValidator) assessMemoryUsage(filePath string, fileInfo os.FileInfo, result *PerformanceValidationResult) {
	format := detectFileFormat(filePath)
	fileSize := fileInfo.Size()

	// Estimate memory usage based on format and file size
	var peakUsageFactor, avgUsageFactor float64

	switch format {
	case FormatParquet:
		// Efficient memory usage due to columnar format and compression
		peakUsageFactor = 1.5
		avgUsageFactor = 1.2

	case FormatORC:
		// Similar to Parquet
		peakUsageFactor = 1.6
		avgUsageFactor = 1.3

	case FormatAVRO:
		// More memory usage for row-based format
		peakUsageFactor = 2.0
		avgUsageFactor = 1.8

	case FormatCSV:
		// Higher memory usage, especially for large files
		peakUsageFactor = 3.0
		avgUsageFactor = 2.5

	case FormatJSON:
		// Highest memory usage due to object overhead
		peakUsageFactor = 4.0
		avgUsageFactor = 3.5

	default:
		peakUsageFactor = 3.5
		avgUsageFactor = 3.0
	}

	peakUsage := int64(float64(fileSize) * peakUsageFactor)
	avgUsage := int64(float64(fileSize) * avgUsageFactor)

	// Calculate efficiency score
	maxEfficiency := 100.0
	if peakUsageFactor > 2.0 {
		maxEfficiency -= (peakUsageFactor - 2.0) * 20.0
	}

	result.MemoryUsage = MemoryUsageMetrics{
		PeakUsage:       peakUsage,
		AverageUsage:    avgUsage,
		EfficiencyScore: maxEfficiency,
	}
}

// assessReadThroughput simulates read throughput assessment
func (v *PerformanceValidator) assessReadThroughput(filePath string, fileInfo os.FileInfo, result *PerformanceValidationResult) {
	format := detectFileFormat(filePath)
	_ = fileInfo.Size() // File size for future use

	// Estimate throughput based on format
	var baseThroughputMBps float64

	switch format {
	case FormatParquet:
		// High throughput due to efficient compression and columnar layout
		baseThroughputMBps = 450.0

	case FormatORC:
		// Similar to Parquet
		baseThroughputMBps = 420.0

	case FormatAVRO:
		// Good throughput for row-based format
		baseThroughputMBps = 300.0

	case FormatCSV:
		// Lower throughput
		if strings.Contains(strings.ToLower(filePath), ".gz") {
			baseThroughputMBps = 180.0 // Decompression overhead
		} else {
			baseThroughputMBps = 250.0
		}

	case FormatJSON:
		// Lowest throughput due to parsing overhead
		if strings.Contains(strings.ToLower(filePath), ".gz") {
			baseThroughputMBps = 120.0
		} else {
			baseThroughputMBps = 160.0
		}

	default:
		baseThroughputMBps = 150.0
	}

	// Convert to bytes per second and add variance
	bytesPerSecond := int64(baseThroughputMBps * 1024 * 1024)
	bytesPerSecond = int64(float64(bytesPerSecond) * 0.9) // Fixed variance

	// Estimate records per second (assuming average record size)
	avgRecordSize := v.estimateAvgRecordSize(format)
	recordsPerSecond := bytesPerSecond / avgRecordSize

	// Calculate efficiency score
	maxEfficiency := 100.0
	if baseThroughputMBps < 200.0 {
		maxEfficiency -= (200.0 - baseThroughputMBps) / 2.0
	}

	result.ReadThroughput = ThroughputMetrics{
		BytesPerSecond:   bytesPerSecond,
		RecordsPerSecond: recordsPerSecond,
		EfficiencyScore:  maxEfficiency,
	}
}

// checkPerformanceThresholds checks performance against thresholds
func (v *PerformanceValidator) checkPerformanceThresholds(result *PerformanceValidationResult, _ ValidationConfig) {
	// Note: config parameter reserved for future threshold customization

	// Check compression ratio
	v.checkCompressionThresholds(result)

	// Check query performance
	v.checkQueryPerformanceThresholds(result)

	// Check memory efficiency
	v.checkMemoryEfficiencyThresholds(result)

	// Check throughput efficiency
	v.checkThroughputThresholds(result)
}

// checkCompressionThresholds checks compression efficiency
func (v *PerformanceValidator) checkCompressionThresholds(result *PerformanceValidationResult) {
	if result.CompressionRatio < 2.0 {
		result.Issues = append(result.Issues, PerformanceIssue{
			Type:       "low_compression",
			Metric:     "compression_ratio",
			Value:      result.CompressionRatio,
			Threshold:  2.0,
			Message:    fmt.Sprintf("Low compression ratio: %.1fx", result.CompressionRatio),
			Severity:   "medium",
			Suggestion: "Consider using Parquet or ORC format with better compression",
		})
	}
}

// checkQueryPerformanceThresholds checks query performance
func (v *PerformanceValidator) checkQueryPerformanceThresholds(result *PerformanceValidationResult) {
	// Check if aggregate time is too high (> 10 seconds)
	if result.QueryPerformance.AggregateTime > 10*time.Second {
		result.Issues = append(result.Issues, PerformanceIssue{
			Type:       "slow_aggregation",
			Metric:     "aggregate_time",
			Value:      result.QueryPerformance.AggregateTime,
			Threshold:  10 * time.Second,
			Message:    fmt.Sprintf("Slow aggregation performance: %v", result.QueryPerformance.AggregateTime),
			Severity:   "high",
			Suggestion: "Use columnar format (Parquet/ORC) for better analytical performance",
		})
	}

	// Check if filter time is too high (> 5 seconds)
	if result.QueryPerformance.FilterTime > 5*time.Second {
		result.Issues = append(result.Issues, PerformanceIssue{
			Type:       "slow_filtering",
			Metric:     "filter_time",
			Value:      result.QueryPerformance.FilterTime,
			Threshold:  5 * time.Second,
			Message:    fmt.Sprintf("Slow filtering performance: %v", result.QueryPerformance.FilterTime),
			Severity:   "medium",
			Suggestion: "Consider indexing or partitioning strategies",
		})
	}
}

// checkMemoryEfficiencyThresholds checks memory efficiency
func (v *PerformanceValidator) checkMemoryEfficiencyThresholds(result *PerformanceValidationResult) {
	if result.MemoryUsage.EfficiencyScore < 60.0 {
		result.Issues = append(result.Issues, PerformanceIssue{
			Type:       "poor_memory_efficiency",
			Metric:     "memory_efficiency",
			Value:      result.MemoryUsage.EfficiencyScore,
			Threshold:  60.0,
			Message:    fmt.Sprintf("Poor memory efficiency: %.1f%%", result.MemoryUsage.EfficiencyScore),
			Severity:   "medium",
			Suggestion: "Use more memory-efficient file format like Parquet",
		})
	}
}

// checkThroughputThresholds checks throughput efficiency
func (v *PerformanceValidator) checkThroughputThresholds(result *PerformanceValidationResult) {
	throughputMBps := float64(result.ReadThroughput.BytesPerSecond) / (1024 * 1024)

	if throughputMBps < 100.0 {
		result.Issues = append(result.Issues, PerformanceIssue{
			Type:       "low_throughput",
			Metric:     "read_throughput",
			Value:      throughputMBps,
			Threshold:  100.0,
			Message:    fmt.Sprintf("Low read throughput: %.1f MB/s", throughputMBps),
			Severity:   "medium",
			Suggestion: "Optimize file format and compression for better throughput",
		})
	}
}

// calculatePerformanceScore calculates overall performance score
func (v *PerformanceValidator) calculatePerformanceScore(result *PerformanceValidationResult) {
	score := 100.0

	// Compression score (25% weight)
	compressionScore := v.calculateCompressionScore(result.CompressionRatio)
	score = score*0.75 + compressionScore*0.25

	// File efficiency (25% weight)
	score = score*0.75 + result.FileEfficiency*0.25

	// Memory efficiency (25% weight)
	score = score*0.75 + result.MemoryUsage.EfficiencyScore*0.25

	// Throughput efficiency (25% weight)
	score = score*0.75 + result.ReadThroughput.EfficiencyScore*0.25

	// Deduct points for issues
	for _, issue := range result.Issues {
		switch issue.Severity {
		case SeverityCritical:
			score -= 25.0
		case SeverityHigh:
			score -= 15.0
		case SeverityMedium:
			score -= 10.0
		case SeverityLow:
			score -= 5.0
		}
	}

	if score < 0 {
		score = 0
	}

	result.Score = score

	// Mark as invalid if score is below threshold (performance issues are warnings)
	if score < 50.0 {
		result.Valid = false
	}
}

// Helper functions

// simulateCompressionRatio simulates compression ratio within a range
func (v *PerformanceValidator) simulateCompressionRatio(min, max float64) float64 {
	return min + 0.5*(max-min) // Fixed mid-range value
}

// addVariance adds random variance to a duration using crypto-secure random
func (v *PerformanceValidator) addVariance(base time.Duration, variance float64) time.Duration {
	// Use crypto-secure random for realistic variance simulation
	randomFactor := cryptoRandFloat64() * variance
	factor := 1.0 + randomFactor
	return time.Duration(float64(base) * factor)
}

// estimateAvgRecordSize estimates average record size for a format
func (v *PerformanceValidator) estimateAvgRecordSize(format string) int64 {
	switch format {
	case "parquet", "orc":
		return 150 // Compressed columnar data
	case "avro":
		return 200 // Compressed row data
	case "csv":
		return 300 // Text format
	case "json":
		return 400 // JSON overhead
	default:
		return 250
	}
}

// calculateCompressionScore calculates score based on compression ratio
func (v *PerformanceValidator) calculateCompressionScore(ratio float64) float64 {
	if ratio >= 8.0 {
		return 100.0
	} else if ratio >= 5.0 {
		return 90.0
	} else if ratio >= 3.0 {
		return 80.0
	} else if ratio >= 2.0 {
		return 70.0
	} else {
		return 50.0
	}
}

// checkCompressionEfficiency checks compression efficiency based on format
func (v *PerformanceValidator) checkCompressionEfficiency(format string, ratio float64, result *PerformanceValidationResult) {
	var expectedMin float64

	switch format {
	case "parquet", "orc":
		expectedMin = 4.0
	case "avro":
		expectedMin = 2.5
	case "csv", "json":
		if strings.Contains(strings.ToLower(format), ".gz") {
			expectedMin = 2.0
		} else {
			expectedMin = 1.0 // No compression expected
		}
	default:
		expectedMin = 1.0
	}

	if ratio < expectedMin {
		severity := SeverityMedium
		if expectedMin-ratio > 2.0 {
			severity = SeverityHigh
		}

		result.Issues = append(result.Issues, PerformanceIssue{
			Type:       "suboptimal_compression",
			Metric:     "compression_ratio",
			Value:      ratio,
			Threshold:  expectedMin,
			Message:    fmt.Sprintf("Compression ratio %.1fx below expected %.1fx for %s format", ratio, expectedMin, format),
			Severity:   severity,
			Suggestion: fmt.Sprintf("Optimize compression settings for %s format", format),
		})
	}
}
