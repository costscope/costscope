package validation

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"local/costscope/internal/core/focus/compliance"
	"local/costscope/internal/core/focus/schemas"
	"local/costscope/internal/core/logging"
)

// File format constants
const (
	FormatParquet = "parquet"
	FormatCSV     = "csv"
	FormatJSON    = "json"
	FormatORC     = "orc"
	FormatAVRO    = "avro"
)

// Engine implements the ValidationEngine interface
type Engine struct {
	validators map[string]Validator
}

// NewEngine creates a new validation engine
func NewEngine() *Engine {
	engine := &Engine{
		validators: make(map[string]Validator),
	}

	// Register default validators
	engine.registerDefaultValidators()

	return engine
}

// registerDefaultValidators registers the built-in validators
func (e *Engine) registerDefaultValidators() {
	logger := logging.GetLogger().WithFields(map[string]interface{}{"component": "validation"})
	// Create compliance manager for compliance validation
	complianceManager := compliance.NewManager()

	// Schema validator (wire schemas manager; falls back to built-ins inside)
	if err := e.RegisterValidator(NewSchemaValidator(schemas.NewManager())); err != nil {
		// Log error but continue - this should not fail in normal circumstances
		logger.WarnWithFields("failed to register schema validator", map[string]interface{}{"err": err})
	}

	// Quality validator
	if err := e.RegisterValidator(NewQualityValidator()); err != nil {
		logger.WarnWithFields("failed to register quality validator", map[string]interface{}{"err": err})
	}

	// Compliance validator with manager
	if err := e.RegisterValidator(NewComplianceValidator(complianceManager)); err != nil {
		logger.WarnWithFields("failed to register compliance validator", map[string]interface{}{"err": err})
	}

	// Performance validator
	if err := e.RegisterValidator(NewPerformanceValidator()); err != nil {
		logger.WarnWithFields("failed to register performance validator", map[string]interface{}{"err": err})
	}

	// Anomaly detector
	if err := e.RegisterValidator(NewAnomalyDetector()); err != nil {
		logger.WarnWithFields("failed to register anomaly detector", map[string]interface{}{"err": err})
	}

	// Strict validator (advisory for now)
	if err := e.RegisterValidator(NewStrictValidator()); err != nil {
		logger.WarnWithFields("failed to register strict validator", map[string]interface{}{"err": err})
	}
}

// Validate validates a single file
func (e *Engine) Validate(filePath string, config ValidationConfig) (*ValidationResult, error) {
	startTime := time.Now()

	// traces removed (E2E debug)

	// Initialize result
	result := &ValidationResult{
		FilePath:       filePath,
		ValidationTime: startTime,
		IsValid:        true,
		OverallScore:   0.0,
		Issues:         []ValidationIssue{},
		Warnings:       []ValidationWarning{},
	}

	// Check if file exists
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("file not found: %w", err)
	}

	result.FileSize = fileInfo.Size()
	result.FileFormat = detectFileFormat(filePath)

	if !config.Quiet {
		fmt.Printf(" Validating: %s\n", filePath)
		fmt.Printf(" Format: %s, Size: %d bytes\n", result.FileFormat, result.FileSize)
	}

	// Run schema validation
	if err := e.runSchemaValidation(filePath, config, result); err != nil {
		return nil, fmt.Errorf("schema validation failed: %w", err)
	}

	// Run quality assessment
	// trace removed
	if config.EnableQuality {
		if err := e.runQualityAssessment(filePath, config, result); err != nil {
			return nil, fmt.Errorf("quality assessment failed: %w", err)
		}
	}

	// Run compliance validation
	// trace removed
	if config.EnableCompliance {
		if err := e.runComplianceValidation(filePath, config, result); err != nil {
			return nil, fmt.Errorf("compliance validation failed: %w", err)
		}
		result.RanCompliance = true
	} else {
		result.RanCompliance = false
		// Ensure score remains neutral (do not penalize)
		result.ComplianceValidation.Valid = true
		result.ComplianceValidation.Score = 0
	}

	// Run performance validation
	// trace removed
	if config.EnablePerformance {
		if err := e.runPerformanceValidation(filePath, config, result); err != nil {
			return nil, fmt.Errorf("performance validation failed: %w", err)
		}
		result.RanPerformance = true
	} else {
		result.RanPerformance = false
		result.PerformanceValidation.Valid = true
		result.PerformanceValidation.Score = 0
	}

	// Run anomaly detection
	// trace removed
	if config.EnableAnomalyDetection {
		if err := e.runAnomalyDetection(filePath, config, result); err != nil {
			return nil, fmt.Errorf("anomaly detection failed: %w", err)
		}
		result.RanAnomalies = true
	} else {
		result.RanAnomalies = false
	}

	// Calculate overall score and finalize result
	e.finalizeResult(result)

	// trace removed

	result.Duration = time.Since(startTime)

	if !config.Quiet {
		e.printValidationSummary(result)
	}

	return result, nil
}

// ValidateBatch validates multiple files in a directory
func (e *Engine) ValidateBatch(inputDir string, config ValidationConfig) ([]*ValidationResult, error) {
	var results []*ValidationResult

	err := filepath.Walk(inputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Check if file format is supported
		format := detectFileFormat(path)
		if !e.isSupportedFormat(format) {
			return nil
		}

		if !config.Quiet {
			fmt.Printf(" Processing: %s\n", path)
		}

		result, err := e.Validate(path, config)
		if err != nil {
			return fmt.Errorf("validation failed for %s: %w", path, err)
		}

		results = append(results, result)
		return nil
	})

	if err != nil {
		return nil, err
	}

	if !config.Quiet {
		e.printBatchSummary(results)
	}

	return results, nil
}

// runSchemaValidation runs schema validation
func (e *Engine) runSchemaValidation(filePath string, config ValidationConfig, result *ValidationResult) error {
	validator, exists := e.validators["schema"]
	if !exists {
		return fmt.Errorf("schema validator not found")
	}

	if !config.Quiet {
		fmt.Printf(" Running schema validation...\n")
	}

	validationResult, err := validator.Validate(filePath, config)
	if err != nil {
		return err
	}

	schemaResult, ok := validationResult.(SchemaValidationResult)
	if !ok {
		return fmt.Errorf("invalid schema validation result type")
	}

	result.SchemaValidation = schemaResult
	result.OverallScore += schemaResult.Score * 0.3 // 30% weight

	if !schemaResult.Valid {
		result.IsValid = false
		for _, issue := range schemaResult.Issues {
			result.Issues = append(result.Issues, ValidationIssue{
				Type:       "schema",
				Severity:   "high",
				Component:  "schema",
				Message:    issue.Message,
				Location:   issue.Column,
				Suggestion: issue.Suggestion,
			})
		}
	}
	// trace removed

	return nil
}

// runQualityAssessment runs data quality assessment
func (e *Engine) runQualityAssessment(filePath string, config ValidationConfig, result *ValidationResult) error {
	validator, exists := e.validators["quality"]
	if !exists {
		return fmt.Errorf("quality validator not found")
	}

	// trace removed
	if !config.Quiet {
		fmt.Printf(" Running data quality assessment...\n")
	}

	validationResult, err := validator.Validate(filePath, config)
	if err != nil {
		return err
	}

	// trace removed
	qualityResult, ok := validationResult.(QualityAssessmentResult)
	if !ok {
		return fmt.Errorf("invalid quality validation result type")
	}

	// trace removed
	result.QualityAssessment = qualityResult
	result.OverallScore += qualityResult.Score * 0.25 // 25% weight

	if !qualityResult.Valid {
		result.IsValid = false
		for _, issue := range qualityResult.Issues {
			result.Issues = append(result.Issues, ValidationIssue{
				Type:       "quality",
				Severity:   issue.Severity,
				Component:  "quality",
				Message:    issue.Message,
				Location:   issue.Column,
				Suggestion: issue.Suggestion,
			})
		}
	}
	// trace removed

	return nil
}

// runComplianceValidation runs compliance validation
func (e *Engine) runComplianceValidation(filePath string, config ValidationConfig, result *ValidationResult) error {
	validator, exists := e.validators["compliance"]
	if !exists {
		return fmt.Errorf("compliance validator not found")
	}

	if !config.Quiet {
		fmt.Printf("️ Running compliance validation...\n")
	}

	validationResult, err := validator.Validate(filePath, config)
	if err != nil {
		return err
	}

	complianceResult, ok := validationResult.(ComplianceValidationResult)
	if !ok {
		return fmt.Errorf("invalid compliance validation result type")
	}

	result.ComplianceValidation = complianceResult
	result.OverallScore += complianceResult.Score * 0.25 // 25% weight

	if !complianceResult.Valid {
		result.IsValid = false
		for _, issue := range complianceResult.Issues {
			result.Issues = append(result.Issues, ValidationIssue{
				Type:       "compliance",
				Severity:   issue.Severity,
				Component:  "compliance",
				Message:    issue.Message,
				Suggestion: issue.Suggestion,
			})
		}
	}

	return nil
}

// runPerformanceValidation runs performance validation
func (e *Engine) runPerformanceValidation(filePath string, config ValidationConfig, result *ValidationResult) error {
	validator, exists := e.validators["performance"]
	if !exists {
		return fmt.Errorf("performance validator not found")
	}

	if !config.Quiet {
		fmt.Printf(" Running performance validation...\n")
	}

	validationResult, err := validator.Validate(filePath, config)
	if err != nil {
		return err
	}

	performanceResult, ok := validationResult.(PerformanceValidationResult)
	if !ok {
		return fmt.Errorf("invalid performance validation result type")
	}

	result.PerformanceValidation = performanceResult
	result.OverallScore += performanceResult.Score * 0.2 // 20% weight

	if !performanceResult.Valid {
		for _, issue := range performanceResult.Issues {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Type:       "performance",
				Component:  "performance",
				Message:    issue.Message,
				Suggestion: issue.Suggestion,
			})
		}
	}

	return nil
}

// runAnomalyDetection runs anomaly detection
func (e *Engine) runAnomalyDetection(filePath string, config ValidationConfig, result *ValidationResult) error {
	validator, exists := e.validators["anomaly"]
	if !exists {
		return fmt.Errorf("anomaly detector not found")
	}

	if !config.Quiet {
		fmt.Printf(" Running anomaly detection...\n")
	}

	validationResult, err := validator.Validate(filePath, config)
	if err != nil {
		return err
	}

	anomalyResult, ok := validationResult.(AnomalyDetectionResult)
	if !ok {
		return fmt.Errorf("invalid anomaly detection result type")
	}

	result.AnomalyDetection = &anomalyResult

	// Add anomaly issues as warnings unless critical
	for _, issue := range anomalyResult.Issues {
		if issue.Severity == SeverityCritical {
			result.Issues = append(result.Issues, ValidationIssue{
				Type:       "anomaly",
				Severity:   issue.Severity,
				Component:  "anomaly",
				Message:    issue.Message,
				Suggestion: issue.Suggestion,
			})
			result.IsValid = false
		} else {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Type:       "anomaly",
				Component:  "anomaly",
				Message:    issue.Message,
				Suggestion: issue.Suggestion,
			})
		}
	}

	return nil
}

// finalizeResult calculates final scores and generates summary
func (e *Engine) finalizeResult(result *ValidationResult) {
	// Calculate overall score (weighted average)
	totalWeight := 0.3 // schema
	if result.QualityAssessment.Score > 0 {
		totalWeight += 0.25 // quality
	}
	if result.ComplianceValidation.Score > 0 {
		totalWeight += 0.25 // compliance
	}
	if result.PerformanceValidation.Score > 0 {
		totalWeight += 0.2 // performance
	}

	if totalWeight > 0 {
		result.OverallScore = result.OverallScore / totalWeight
	}

	// Generate grade
	grade := "F"
	recommendation := "Significant issues found. Immediate action required."
	nextSteps := []string{
		"Review and fix critical issues",
		"Improve data quality",
		"Ensure FOCUS compliance",
	}

	if result.OverallScore >= 95 {
		grade = "A"
		recommendation = "Excellent! Dataset meets high quality standards."
		nextSteps = []string{"Continue monitoring", "Regular validation checks"}
	} else if result.OverallScore >= 85 {
		grade = "B"
		recommendation = "Good quality with minor improvements needed."
		nextSteps = []string{"Address minor issues", "Optimize performance"}
	} else if result.OverallScore >= 75 {
		grade = "C"
		recommendation = "Acceptable quality but improvements recommended."
		nextSteps = []string{"Fix identified issues", "Improve data quality"}
	} else if result.OverallScore >= 65 {
		grade = "D"
		recommendation = "Below standards. Significant improvements needed."
		nextSteps = []string{"Review all validation issues", "Implement quality controls"}
	}

	// Count issues by severity
	criticalIssues := 0
	highIssues := 0
	mediumIssues := 0
	lowIssues := 0

	for _, issue := range result.Issues {
		switch issue.Severity {
		case SeverityCritical:
			criticalIssues++
		case SeverityHigh:
			highIssues++
		case SeverityMedium:
			mediumIssues++
		case SeverityLow:
			lowIssues++
		}
	}

	result.Summary = ValidationSummary{
		TotalIssues:    len(result.Issues),
		CriticalIssues: criticalIssues,
		HighIssues:     highIssues,
		MediumIssues:   mediumIssues,
		LowIssues:      lowIssues,
		TotalWarnings:  len(result.Warnings),
		OverallGrade:   grade,
		Recommendation: recommendation,
		NextSteps:      nextSteps,
	}

	// Compute top issue types (simple count by message prefix)
	issueCounts := map[string]int{}
	for _, is := range result.Issues {
		key := is.Type
		issueCounts[key]++
	}
	// Attach as warnings for quick visibility
	for k, c := range issueCounts {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Type:      "top_issue",
			Component: "summary",
			Message:   fmt.Sprintf("%s: %d occurrences", k, c),
		})
	}
}

// printValidationSummary prints validation summary
func (e *Engine) printValidationSummary(result *ValidationResult) {
	separator := strings.Repeat("=", 60)
	fmt.Printf("\n%s\n", separator)
	fmt.Printf(" VALIDATION SUMMARY\n")
	fmt.Printf("%s\n", separator)

	status := " PASSED"
	if !result.IsValid {
		status = " FAILED"
	}

	fmt.Printf("Status: %s\n", status)
	fmt.Printf("Overall Score: %.1f%%\n", result.OverallScore)
	fmt.Printf("Grade: %s\n", result.Summary.OverallGrade)
	fmt.Printf("Duration: %v\n", result.Duration)

	if len(result.Issues) > 0 {
		fmt.Printf("\n Issues Found: %d\n", len(result.Issues))
		fmt.Printf("  Critical: %d, High: %d, Medium: %d, Low: %d\n",
			result.Summary.CriticalIssues,
			result.Summary.HighIssues,
			result.Summary.MediumIssues,
			result.Summary.LowIssues)
	}

	if len(result.Warnings) > 0 {
		fmt.Printf("️  Warnings: %d\n", len(result.Warnings))
	}

	fmt.Printf("\n Recommendation: %s\n", result.Summary.Recommendation)
}

// printBatchSummary prints batch validation summary
func (e *Engine) printBatchSummary(results []*ValidationResult) {
	if len(results) == 0 {
		return
	}

	separator := strings.Repeat("=", 60)
	fmt.Printf("\n%s\n", separator)
	fmt.Printf(" BATCH VALIDATION SUMMARY\n")
	fmt.Printf("%s\n", separator)

	totalFiles := len(results)
	passedFiles := 0
	totalScore := 0.0
	totalIssues := 0
	totalWarnings := 0

	for _, result := range results {
		if result.IsValid {
			passedFiles++
		}
		totalScore += result.OverallScore
		totalIssues += len(result.Issues)
		totalWarnings += len(result.Warnings)
	}

	avgScore := totalScore / float64(totalFiles)

	fmt.Printf("Files Processed: %d\n", totalFiles)
	fmt.Printf("Files Passed: %d (%.1f%%)\n", passedFiles, float64(passedFiles)/float64(totalFiles)*100)
	fmt.Printf("Average Score: %.1f%%\n", avgScore)
	fmt.Printf("Total Issues: %d\n", totalIssues)
	fmt.Printf("Total Warnings: %d\n", totalWarnings)
}

// RegisterValidator registers a new validator
func (e *Engine) RegisterValidator(validator Validator) error {
	if validator == nil {
		return fmt.Errorf("validator cannot be nil")
	}

	name := validator.Name()
	if name == "" {
		return fmt.Errorf("validator name cannot be empty")
	}

	e.validators[name] = validator
	return nil
}

// GetSupportedFormats returns list of supported file formats
func (e *Engine) GetSupportedFormats() []string {
	return []string{"parquet", "csv", "json", "orc", "avro"}
}

// GetSupportedSpecs returns list of supported FOCUS specifications
func (e *Engine) GetSupportedSpecs() []ValidationSpec {
	// Prefer dynamic discovery from the schemas.Manager to stay in sync with available schemas.
	// Falls back to the static list if discovery fails.
	m := schemas.NewManager()
	names := m.GetAvailableSchemas()
	if len(names) == 0 {
		return []ValidationSpec{SpecFOCUS12, SpecFOCUS11, SpecFOCUS10}
	}

	// Ensure stable, descending order by semantic version after the "focus-" prefix (e.g., focus-1.2 > focus-1.1 > focus-1.0).
	type ver struct{ major, minor int }
	parse := func(name string) (ver, bool) {
		if !strings.HasPrefix(name, "focus-") {
			return ver{}, false
		}
		rest := strings.TrimPrefix(name, "focus-")
		parts := strings.SplitN(rest, ".", 3)
		if len(parts) < 2 {
			return ver{}, false
		}
		var v ver
		// best-effort parse with error checks
		maj, err1 := strconv.Atoi(parts[0])
		if err1 != nil {
			return ver{}, false
		}
		min, err2 := strconv.Atoi(parts[1])
		if err2 != nil {
			return ver{}, false
		}
		v.major = maj
		v.minor = min
		return v, true
	}

	// Copy and sort
	sorted := append([]string(nil), names...)
	sort.Slice(sorted, func(i, j int) bool {
		vi, okI := parse(sorted[i])
		vj, okJ := parse(sorted[j])
		if okI && okJ {
			if vi.major != vj.major {
				return vi.major > vj.major
			}
			return vi.minor > vj.minor
		}
		// Non-conforming names go last, keep original order among them
		if okI != okJ {
			return okI // parsed first
		}
		return sorted[i] < sorted[j]
	})

	// Map to ValidationSpec
	specs := make([]ValidationSpec, 0, len(sorted))
	for _, n := range sorted {
		specs = append(specs, ValidationSpec(n))
	}
	return specs
}

// isSupportedFormat checks if file format is supported
func (e *Engine) isSupportedFormat(format string) bool {
	for _, supported := range e.GetSupportedFormats() {
		if format == supported {
			return true
		}
	}
	return false
}

// detectFileFormat detects file format from extension
func detectFileFormat(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".parquet":
		return FormatParquet
	case ".csv":
		return FormatCSV
	case ".json":
		return FormatJSON
	case ".orc":
		return FormatORC
	case ".avro":
		return FormatAVRO
	default:
		return "unknown"
	}
}
