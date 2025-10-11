package production

import "time"

// ProductionSystemMetrics represents comprehensive system metrics for production readiness
type ProductionSystemMetrics struct {
	Timestamp        time.Time          `json:"timestamp"`
	SystemHealth     SystemHealthStatus `json:"system_health"`
	Performance      PerformanceMetrics `json:"performance"`
	Security         SecurityMetrics    `json:"security"`
	Integration      IntegrationMetrics `json:"integration"`
	Analytics        AnalyticsMetrics   `json:"analytics"`
	TotalCommands    int                `json:"total_commands"`
	TotalEndpoints   int                `json:"total_endpoints"`
	TotalFeatures    int                `json:"total_features"`
	CompletionLevel  string             `json:"completion_level"`
	ProductionReady  bool               `json:"production_ready"`
	ReadinessScore   int                `json:"readiness_score"` // 0-100
	Recommendations  []string           `json:"recommendations"`
	CriticalIssues   []string           `json:"critical_issues"`
	ProcessingTimeMs int64              `json:"processing_time_ms"`
}

// SystemHealthStatus represents the overall health of the system
type SystemHealthStatus struct {
	Status          string            `json:"status"` // healthy, degraded, critical
	ComponentHealth map[string]string `json:"component_health"`
	UptimeHours     float64           `json:"uptime_hours"`
	ErrorRate       float64           `json:"error_rate"`
	ResponseTimeMs  float64           `json:"response_time_ms"`
	HealthScore     int               `json:"health_score"` // 0-100
}

// PerformanceMetrics represents system performance metrics
type PerformanceMetrics struct {
	ThroughputOpsPerSec int64   `json:"throughput_ops_per_sec"`
	MemoryUsageMB       float64 `json:"memory_usage_mb"`
	MemoryLimitMB       float64 `json:"memory_limit_mb"`
	CPUUsagePercent     float64 `json:"cpu_usage_percent"`
	DiskUsagePercent    float64 `json:"disk_usage_percent"`
	NetworkLatencyMs    float64 `json:"network_latency_ms"`
	OptimizationScore   int     `json:"optimization_score"` // 0-100
	PerformanceGrade    string  `json:"performance_grade"`  // A, B, C, D, F
}

// SecurityMetrics represents security assessment metrics
type SecurityMetrics struct {
	SecurityScore       int               `json:"security_score"` // 0-100
	VulnerabilitiesOpen int               `json:"vulnerabilities_open"`
	VulnerabilitiesHigh int               `json:"vulnerabilities_high"`
	ComplianceStatus    map[string]string `json:"compliance_status"`
	EncryptionEnabled   bool              `json:"encryption_enabled"`
	AccessViolations    int               `json:"access_violations"`
	AuditScore          int               `json:"audit_score"`    // 0-100
	SecurityGrade       string            `json:"security_grade"` // A, B, C, D, F
}

// IntegrationMetrics represents integration and operational metrics
type IntegrationMetrics struct {
	ConnectedSystems    int               `json:"connected_systems"`
	ActiveWorkflows     int               `json:"active_workflows"`
	AlertChannels       int               `json:"alert_channels"`
	AutomationSavings   float64           `json:"automation_savings"`
	IntegrationHealth   map[string]string `json:"integration_health"`
	DeploymentStatus    string            `json:"deployment_status"`
	IntegrationScore    int               `json:"integration_score"`    // 0-100
	OperationalMaturity string            `json:"operational_maturity"` // basic, intermediate, advanced
}

// AnalyticsMetrics represents analytics and ML capabilities metrics
type AnalyticsMetrics struct {
	MLModelsActive      int     `json:"ml_models_active"`
	PredictionAccuracy  float64 `json:"prediction_accuracy"`
	AnomaliesDetected   int     `json:"anomalies_detected"`
	ForecastReliability float64 `json:"forecast_reliability"`
	InsightsGenerated   int     `json:"insights_generated"`
	DataQualityScore    int     `json:"data_quality_score"` // 0-100
	AnalyticsMaturity   string  `json:"analytics_maturity"` // basic, intermediate, advanced
}

// ProductionOptimizationReport represents comprehensive system optimization report
type ProductionOptimizationReport struct {
	GeneratedAt         time.Time                  `json:"generated_at"`
	SystemOverview      SystemOverview             `json:"system_overview"`
	OptimizationResults OptimizationResults        `json:"optimization_results"`
	Recommendations     []ProductionRecommendation `json:"recommendations"`
	ROIAnalysis         ROIAnalysis                `json:"roi_analysis"`
	FutureRoadmap       []RoadmapItem              `json:"future_roadmap"`
	ExecutiveSummary    string                     `json:"executive_summary"`
	ProcessingTimeMs    int64                      `json:"processing_time_ms"`
}

// SystemOverview represents high-level system overview
type SystemOverview struct {
	TotalCapabilities   int      `json:"total_capabilities"`
	ActiveFeatures      []string `json:"active_features"`
	SupportedProviders  []string `json:"supported_providers"`
	DeploymentReadiness string   `json:"deployment_readiness"`
	ScalabilityRating   int      `json:"scalability_rating"` // 0-100
	Architecture        string   `json:"architecture"`       // monolithic, microservices, serverless
	TechnologyStack     []string `json:"technology_stack"`
}

// OptimizationResults represents results of optimization analysis
type OptimizationResults struct {
	TotalImprovements    int     `json:"total_improvements"`
	PerformanceGains     float64 `json:"performance_gains"`
	CostSavings          float64 `json:"cost_savings"`
	SecurityEnhancements int     `json:"security_enhancements"`
	EfficiencyGains      float64 `json:"efficiency_gains"`
	OptimizationScore    int     `json:"optimization_score"` // 0-100
}

// ProductionRecommendation represents an actionable recommendation
type ProductionRecommendation struct {
	ID          string   `json:"id"`
	Type        string   `json:"type"` // performance, security, cost, integration
	Priority    Priority `json:"priority"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Impact      Impact   `json:"impact"`
	Effort      Effort   `json:"effort"`
	Timeline    string   `json:"timeline"`
	Cost        float64  `json:"cost"`
	ROI         float64  `json:"roi"`
	Actions     []Action `json:"actions"`
	Resources   []string `json:"resources"`
	Risks       []Risk   `json:"risks"`
}

// ROIAnalysis represents return on investment analysis
type ROIAnalysis struct {
	TotalInvestment   float64            `json:"total_investment"`
	ProjectedSavings  float64            `json:"projected_savings"`
	ROIPercentage     float64            `json:"roi_percentage"`
	PaybackPeriodDays int                `json:"payback_period_days"`
	NPV               float64            `json:"npv"` // Net Present Value
	IRR               float64            `json:"irr"` // Internal Rate of Return
	SavingsBreakdown  map[string]float64 `json:"savings_breakdown"`
	CostBenefitRatio  float64            `json:"cost_benefit_ratio"`
}

// RoadmapItem represents a future development item
type RoadmapItem struct {
	ID           string   `json:"id"`
	Title        string   `json:"title"`
	Description  string   `json:"description"`
	Quarter      string   `json:"quarter"`
	Priority     Priority `json:"priority"`
	Category     string   `json:"category"`
	Dependencies []string `json:"dependencies"`
	Resources    []string `json:"resources"`
	Value        int      `json:"value"`  // Business value score 0-100
	Effort       int      `json:"effort"` // Effort score 0-100
}

// DeploymentReadinessAssessment represents production deployment assessment
type DeploymentReadinessAssessment struct {
	OverallReadiness      string                 `json:"overall_readiness"` // ready, needs_work, not_ready
	ReadinessScore        int                    `json:"readiness_score"`   // 0-100
	ReadinessStatus       string                 `json:"readiness_status"`
	CriticalIssues        []Issue                `json:"critical_issues"`
	Warnings              []Issue                `json:"warnings"`
	ReadinessChecklist    []ChecklistItem        `json:"readiness_checklist"`
	EnvironmentChecks     map[string]CheckResult `json:"environment_checks"`
	HealthChecks          map[string]bool        `json:"health_checks"`
	PerformanceTests      map[string]bool        `json:"performance_tests"`
	SecurityValidation    map[string]bool        `json:"security_validation"`
	EnvironmentValidation map[string]bool        `json:"environment_validation"`
	BlockingIssues        []Issue                `json:"blocking_issues"`
	AssessmentTimestamp   time.Time              `json:"assessment_timestamp"`
	DeploymentPlan        DeploymentPlan         `json:"deployment_plan"`
	RollbackPlan          RollbackPlan           `json:"rollback_plan"`
	ProcessingTimeMs      int64                  `json:"processing_time_ms"`
}

// Supporting types

// Priority represents priority levels
type Priority string

const (
	PriorityLow      Priority = "low"
	PriorityMedium   Priority = "medium"
	PriorityHigh     Priority = "high"
	PriorityCritical Priority = "critical"
)

// Impact represents impact levels
type Impact string

const (
	ImpactLow      Impact = "low"
	ImpactMedium   Impact = "medium"
	ImpactHigh     Impact = "high"
	ImpactCritical Impact = "critical"
)

// Effort represents effort levels
type Effort string

const (
	EffortLow    Effort = "low"
	EffortMedium Effort = "medium"
	EffortHigh   Effort = "high"
)

// Action represents an actionable item
type Action struct {
	ID          string   `json:"id"`
	Description string   `json:"description"`
	Command     string   `json:"command,omitempty"`
	Resources   []string `json:"resources"`
	Order       int      `json:"order"`
}

// Risk represents a potential risk
type Risk struct {
	ID          string  `json:"id"`
	Description string  `json:"description"`
	Probability float64 `json:"probability"` // 0.0 - 1.0
	Impact      Impact  `json:"impact"`
	Mitigation  string  `json:"mitigation"`
}

// Issue represents a system issue
type Issue struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	Severity    string `json:"severity"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Component   string `json:"component"`
	Resolution  string `json:"resolution"`
}

// ChecklistItem represents a deployment checklist item
type ChecklistItem struct {
	ID          string `json:"id"`
	Category    string `json:"category"`
	Description string `json:"description"`
	Status      string `json:"status"` // passed, failed, pending, na
	Required    bool   `json:"required"`
	Notes       string `json:"notes,omitempty"`
}

// CheckResult represents the result of an environment check
type CheckResult struct {
	Status      string    `json:"status"` // passed, failed, warning
	Message     string    `json:"message"`
	Details     string    `json:"details,omitempty"`
	CheckedAt   time.Time `json:"checked_at"`
	Remediation string    `json:"remediation,omitempty"`
}

// DeploymentPlan represents a deployment plan
type DeploymentPlan struct {
	Environment       string           `json:"environment"`
	Strategy          string           `json:"strategy"` // blue_green, rolling, canary
	Steps             []DeploymentStep `json:"steps"`
	RollbackSteps     []DeploymentStep `json:"rollback_steps"`
	Prerequisites     []string         `json:"prerequisites"`
	HealthChecks      []HealthCheck    `json:"health_checks"`
	EstimatedTimeMs   int64            `json:"estimated_time_ms"`
	EstimatedDuration time.Duration    `json:"estimated_duration"`
	RiskLevel         string           `json:"risk_level"`
	ApprovalRequired  bool             `json:"approval_required"`
	PlanCreatedAt     time.Time        `json:"plan_created_at"`
}

// DeploymentStep represents a step in deployment
type DeploymentStep struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Command     string        `json:"command"`
	Order       int           `json:"order"`
	Timeout     int           `json:"timeout"` // seconds
	Duration    time.Duration `json:"duration"`
	Type        string        `json:"type"`
	Required    bool          `json:"required"`
	Rollback    string        `json:"rollback,omitempty"`
}

// HealthCheck represents a health check
type HealthCheck struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"` // http, tcp, command
	Target   string `json:"target"`
	Timeout  int    `json:"timeout"` // seconds
	Retries  int    `json:"retries"`
	Required bool   `json:"required"`
}

// RollbackPlan represents a rollback plan
type RollbackPlan struct {
	Strategy        string           `json:"strategy"`
	Steps           []DeploymentStep `json:"steps"`
	TriggerPoints   []string         `json:"trigger_points"`
	EstimatedTimeMs int64            `json:"estimated_time_ms"`
	DataRecovery    bool             `json:"data_recovery"`
}

// DeploymentRequirements represents deployment requirements
type DeploymentRequirements struct {
	Environment          string             `json:"environment"`
	MinHealthScore       int                `json:"min_health_score"`
	RequiredChecks       []string           `json:"required_checks"`
	PerformanceTargets   map[string]float64 `json:"performance_targets"`
	SecurityRequirements []string           `json:"security_requirements"`
	ComplianceStandards  []string           `json:"compliance_standards"`
	ResourceLimits       map[string]string  `json:"resource_limits"`
	MonitoringEnabled    bool               `json:"monitoring_enabled"`
	BackupRequired       bool               `json:"backup_required"`
}

// EnvironmentValidation represents environment validation results
type EnvironmentValidation struct {
	Environment         string    `json:"environment"`
	ResourcesAvailable  bool      `json:"resources_available"`
	ConfigurationValid  bool      `json:"configuration_valid"`
	ConnectivityOK      bool      `json:"connectivity_ok"`
	DependenciesReady   bool      `json:"dependencies_ready"`
	ValidationScore     int       `json:"validation_score"`
	Issues              []Issue   `json:"issues"`
	ValidationTimestamp time.Time `json:"validation_timestamp"`
}

// HealthCheckResults represents health check results
type HealthCheckResults struct {
	OverallHealthScore int             `json:"overall_health_score"`
	ComponentResults   map[string]bool `json:"component_results"`
	FailedChecks       []string        `json:"failed_checks"`
	CheckTimestamp     time.Time       `json:"check_timestamp"`
}
