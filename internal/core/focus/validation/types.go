package validation

import (
	"time"
)

// ValidationLevel represents the level of validation
type ValidationLevel string

const (
	ValidationLevelBasic      ValidationLevel = "basic"
	ValidationLevelStandard   ValidationLevel = "standard"
	ValidationLevelStrict     ValidationLevel = "strict"
	ValidationLevelEnterprise ValidationLevel = "enterprise"
)

// ValidationSpec represents FOCUS specification version
type ValidationSpec string

const (
	SpecFOCUS12 ValidationSpec = "focus-1.2"
	SpecFOCUS11 ValidationSpec = "focus-1.1"
	SpecFOCUS10 ValidationSpec = "focus-1.0"
)

// ValidationConfig holds configuration for validation
type ValidationConfig struct {
	Level                  ValidationLevel        `json:"level"`
	Spec                   ValidationSpec         `json:"spec"`
	Format                 string                 `json:"format"`
	Strict                 bool                   `json:"strict"`
	EnableCompliance       bool                   `json:"enable_compliance"`
	EnableQuality          bool                   `json:"enable_quality"`
	EnablePerformance      bool                   `json:"enable_performance"`
	EnableAnomalyDetection bool                   `json:"enable_anomaly_detection"`
	AutoFix                bool                   `json:"auto_fix"`
	OutputPath             string                 `json:"output_path"`
	Quiet                  bool                   `json:"quiet"`
	Verbose                bool                   `json:"verbose"`
	SchemaOverrides        map[string]interface{} `json:"schema_overrides,omitempty"`
	ComplianceRules        []ComplianceRule       `json:"compliance_rules,omitempty"`
}

// ValidationResult represents the result of validation
type ValidationResult struct {
	FilePath       string        `json:"file_path"`
	FileFormat     string        `json:"file_format"`
	FileSize       int64         `json:"file_size"`
	ValidationTime time.Time     `json:"validation_time"`
	Duration       time.Duration `json:"duration"`
	IsValid        bool          `json:"is_valid"`
	OverallScore   float64       `json:"overall_score"`

	// Core validation results
	SchemaValidation      SchemaValidationResult      `json:"schema_validation"`
	QualityAssessment     QualityAssessmentResult     `json:"quality_assessment"`
	ComplianceValidation  ComplianceValidationResult  `json:"compliance_validation"`
	PerformanceValidation PerformanceValidationResult `json:"performance_validation"`

	// Execution flags to indicate which domains actually ran (for UI/reporting)
	RanCompliance  bool `json:"ran_compliance"`
	RanPerformance bool `json:"ran_performance"`
	RanAnomalies   bool `json:"ran_anomalies"`

	// Enterprise features
	AnomalyDetection       *AnomalyDetectionResult `json:"anomaly_detection,omitempty"`
	RemediationSuggestions []RemediationSuggestion `json:"remediation_suggestions,omitempty"`

	// Summary
	Issues   []ValidationIssue   `json:"issues"`
	Warnings []ValidationWarning `json:"warnings"`
	Summary  ValidationSummary   `json:"summary"`
}

// SchemaValidationResult holds schema validation results
type SchemaValidationResult struct {
	Valid                bool                            `json:"valid"`
	Score                float64                         `json:"score"`
	RequiredColumns      map[string]bool                 `json:"required_columns"`
	OptionalColumns      map[string]bool                 `json:"optional_columns"`
	UnknownColumns       []string                        `json:"unknown_columns"`
	TypeValidation       map[string]TypeValidation       `json:"type_validation"`
	ConstraintValidation map[string]ConstraintValidation `json:"constraint_validation"`
	Issues               []SchemaIssue                   `json:"issues"`
}

// QualityAssessmentResult holds data quality assessment results
type QualityAssessmentResult struct {
	Valid          bool                        `json:"valid"`
	Score          float64                     `json:"score"`
	RowCount       int64                       `json:"row_count"`
	ColumnCount    int                         `json:"column_count"`
	MissingValues  map[string]int64            `json:"missing_values"`
	DataTypes      map[string]string           `json:"data_types"`
	UniqueValues   map[string]int64            `json:"unique_values"`
	DuplicateRows  int64                       `json:"duplicate_rows"`
	NullPercentage map[string]float64          `json:"null_percentage"`
	Statistics     map[string]ColumnStatistics `json:"statistics"`
	// Optional sample values for additional validations (e.g., enum checks)
	ValueSamples map[string][]string `json:"value_samples,omitempty"`
	Issues       []QualityIssue      `json:"issues"`
}

// ComplianceValidationResult holds compliance validation results
type ComplianceValidationResult struct {
	Valid                bool                       `json:"valid"`
	Score                float64                    `json:"score"`
	FOCUSCompliance      FOCUSComplianceResult      `json:"focus_compliance"`
	RegulatoryCompliance RegulatoryComplianceResult `json:"regulatory_compliance"`
	IndustryStandards    IndustryStandardsResult    `json:"industry_standards"`
	AuditTrail           []AuditEvent               `json:"audit_trail"`
	Issues               []ComplianceIssue          `json:"issues"`
}

// PerformanceValidationResult holds performance validation results
type PerformanceValidationResult struct {
	Valid            bool                    `json:"valid"`
	Score            float64                 `json:"score"`
	CompressionRatio float64                 `json:"compression_ratio"`
	FileEfficiency   float64                 `json:"file_efficiency"`
	QueryPerformance QueryPerformanceMetrics `json:"query_performance"`
	MemoryUsage      MemoryUsageMetrics      `json:"memory_usage"`
	ReadThroughput   ThroughputMetrics       `json:"read_throughput"`
	Issues           []PerformanceIssue      `json:"issues"`
}

// AnomalyDetectionResult holds anomaly detection results
type AnomalyDetectionResult struct {
	Valid             bool              `json:"valid"`
	Score             float64           `json:"score"`
	TotalAnomalies    int               `json:"total_anomalies"`
	CriticalAnomalies int               `json:"critical_anomalies"`
	Anomalies         []DetectedAnomaly `json:"anomalies"`
	Patterns          []AnomalyPattern  `json:"patterns"`
	Issues            []AnomalyIssue    `json:"issues"`
}

// Supporting types
type TypeValidation struct {
	Expected string `json:"expected"`
	Actual   string `json:"actual"`
	Valid    bool   `json:"valid"`
	Message  string `json:"message,omitempty"`
}

type ConstraintValidation struct {
	ConstraintType string      `json:"constraint_type"`
	Expected       interface{} `json:"expected"`
	Actual         interface{} `json:"actual"`
	Valid          bool        `json:"valid"`
	Message        string      `json:"message,omitempty"`
}

type ColumnStatistics struct {
	Min         interface{}            `json:"min,omitempty"`
	Max         interface{}            `json:"max,omitempty"`
	Mean        interface{}            `json:"mean,omitempty"`
	Median      interface{}            `json:"median,omitempty"`
	StdDev      interface{}            `json:"std_dev,omitempty"`
	Percentiles map[string]interface{} `json:"percentiles,omitempty"`
}

type FOCUSComplianceResult struct {
	SpecVersion    string          `json:"spec_version"`
	RequiredFields map[string]bool `json:"required_fields"`
	OptionalFields map[string]bool `json:"optional_fields"`
	CustomFields   []string        `json:"custom_fields"`
	MetadataValid  bool            `json:"metadata_valid"`
	DimensionValid bool            `json:"dimension_valid"`
	MetricValid    bool            `json:"metric_valid"`
	Issues         []FOCUSIssue    `json:"issues"`
}

type RegulatoryComplianceResult struct {
	GDPR   ComplianceStatus  `json:"gdpr"`
	SOX    ComplianceStatus  `json:"sox"`
	HIPAA  ComplianceStatus  `json:"hipaa"`
	PCI    ComplianceStatus  `json:"pci"`
	Issues []RegulatoryIssue `json:"issues"`
}

type IndustryStandardsResult struct {
	ISO27001 ComplianceStatus `json:"iso_27001"`
	SOC2     ComplianceStatus `json:"soc2"`
	NIST     ComplianceStatus `json:"nist"`
	Issues   []StandardIssue  `json:"issues"`
}

type ComplianceStatus struct {
	Applicable bool    `json:"applicable"`
	Compliant  bool    `json:"compliant"`
	Score      float64 `json:"score"`
	Details    string  `json:"details,omitempty"`
}

type QueryPerformanceMetrics struct {
	SelectTime    time.Duration `json:"select_time"`
	FilterTime    time.Duration `json:"filter_time"`
	AggregateTime time.Duration `json:"aggregate_time"`
	SortTime      time.Duration `json:"sort_time"`
}

type MemoryUsageMetrics struct {
	PeakUsage       int64   `json:"peak_usage"`
	AverageUsage    int64   `json:"average_usage"`
	EfficiencyScore float64 `json:"efficiency_score"`
}

type ThroughputMetrics struct {
	BytesPerSecond   int64   `json:"bytes_per_second"`
	RecordsPerSecond int64   `json:"records_per_second"`
	EfficiencyScore  float64 `json:"efficiency_score"`
}

type DetectedAnomaly struct {
	Type        string                 `json:"type"`
	Severity    string                 `json:"severity"`
	Column      string                 `json:"column"`
	RowIndex    int64                  `json:"row_index,omitempty"`
	Value       interface{}            `json:"value"`
	Expected    interface{}            `json:"expected,omitempty"`
	Score       float64                `json:"score"`
	Description string                 `json:"description"`
	Context     map[string]interface{} `json:"context,omitempty"`
}

type AnomalyPattern struct {
	Name        string  `json:"name"`
	Frequency   int     `json:"frequency"`
	Confidence  float64 `json:"confidence"`
	Description string  `json:"description"`
}

// Issue types
type ValidationIssue struct {
	Type       string `json:"type"`
	Severity   string `json:"severity"`
	Component  string `json:"component"`
	Message    string `json:"message"`
	Location   string `json:"location,omitempty"`
	Suggestion string `json:"suggestion,omitempty"`
}

type ValidationWarning struct {
	Type       string `json:"type"`
	Component  string `json:"component"`
	Message    string `json:"message"`
	Location   string `json:"location,omitempty"`
	Suggestion string `json:"suggestion,omitempty"`
}

type SchemaIssue struct {
	Type       string `json:"type"`
	Column     string `json:"column"`
	Message    string `json:"message"`
	Suggestion string `json:"suggestion,omitempty"`
}

type QualityIssue struct {
	Type       string `json:"type"`
	Column     string `json:"column,omitempty"`
	Message    string `json:"message"`
	Severity   string `json:"severity"`
	Count      int64  `json:"count,omitempty"`
	Suggestion string `json:"suggestion,omitempty"`
}

type ComplianceIssue struct {
	Type       string `json:"type"`
	Standard   string `json:"standard"`
	Rule       string `json:"rule"`
	Message    string `json:"message"`
	Severity   string `json:"severity"`
	Suggestion string `json:"suggestion,omitempty"`
}

type FOCUSIssue struct {
	Type       string `json:"type"`
	Field      string `json:"field"`
	Message    string `json:"message"`
	Suggestion string `json:"suggestion,omitempty"`
}

type RegulatoryIssue struct {
	Regulation  string `json:"regulation"`
	Requirement string `json:"requirement"`
	Message     string `json:"message"`
	Severity    string `json:"severity"`
	Suggestion  string `json:"suggestion,omitempty"`
}

type StandardIssue struct {
	Standard   string `json:"standard"`
	Control    string `json:"control"`
	Message    string `json:"message"`
	Severity   string `json:"severity"`
	Suggestion string `json:"suggestion,omitempty"`
}

type PerformanceIssue struct {
	Type       string      `json:"type"`
	Metric     string      `json:"metric"`
	Value      interface{} `json:"value"`
	Threshold  interface{} `json:"threshold"`
	Message    string      `json:"message"`
	Severity   string      `json:"severity"`
	Suggestion string      `json:"suggestion,omitempty"`
}

type AnomalyIssue struct {
	Type       string          `json:"type"`
	Anomaly    DetectedAnomaly `json:"anomaly"`
	Message    string          `json:"message"`
	Severity   string          `json:"severity"`
	Suggestion string          `json:"suggestion,omitempty"`
}

type RemediationSuggestion struct {
	Type        string              `json:"type"`
	Priority    string              `json:"priority"`
	Title       string              `json:"title"`
	Description string              `json:"description"`
	Actions     []RemediationAction `json:"actions"`
	Impact      string              `json:"impact"`
	Effort      string              `json:"effort"`
	Automated   bool                `json:"automated"`
}

type RemediationAction struct {
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Command     string                 `json:"command,omitempty"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
}

type ValidationSummary struct {
	TotalIssues    int      `json:"total_issues"`
	CriticalIssues int      `json:"critical_issues"`
	HighIssues     int      `json:"high_issues"`
	MediumIssues   int      `json:"medium_issues"`
	LowIssues      int      `json:"low_issues"`
	TotalWarnings  int      `json:"total_warnings"`
	OverallGrade   string   `json:"overall_grade"`
	Recommendation string   `json:"recommendation"`
	NextSteps      []string `json:"next_steps"`
}

type ComplianceRule struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Standard    string                 `json:"standard"`
	Description string                 `json:"description"`
	Severity    string                 `json:"severity"`
	Parameters  map[string]interface{} `json:"parameters"`
	Enabled     bool                   `json:"enabled"`
}

type AuditEvent struct {
	Timestamp time.Time              `json:"timestamp"`
	Type      string                 `json:"type"`
	Component string                 `json:"component"`
	Action    string                 `json:"action"`
	User      string                 `json:"user,omitempty"`
	Details   map[string]interface{} `json:"details"`
	Result    string                 `json:"result"`
}

// Validator interface defines the contract for validation components
type Validator interface {
	Name() string
	Validate(data interface{}, config ValidationConfig) (interface{}, error)
	SupportsFormat(format string) bool
}

// ValidationEngine interface defines the main validation engine
type ValidationEngine interface {
	Validate(filePath string, config ValidationConfig) (*ValidationResult, error)
	ValidateBatch(inputDir string, config ValidationConfig) ([]*ValidationResult, error)
	RegisterValidator(validator Validator) error
	GetSupportedFormats() []string
	GetSupportedSpecs() []ValidationSpec
}
