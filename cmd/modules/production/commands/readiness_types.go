package commands

import (
	"time"

	"github.com/costscope/costscope/internal/core/production"
)

// Production assessment result types

// ProductionAssessmentResult represents comprehensive production assessment results
type ProductionAssessmentResult struct {
	Environment         string                                    `json:"environment"`
	Timestamp           time.Time                                 `json:"timestamp"`
	AssessmentDuration  time.Duration                             `json:"assessment_duration"`
	ReadinessAssessment *production.DeploymentReadinessAssessment `json:"readiness_assessment"`
	SystemStatus        *production.ProductionSystemMetrics       `json:"system_status"`
	Recommendations     []production.ProductionRecommendation     `json:"recommendations"`
	Summary             *AssessmentSummary                        `json:"summary"`
}

// AssessmentSummary provides high-level assessment summary
type AssessmentSummary struct {
	OverallScore   int      `json:"overall_score"`
	ReadinessLevel string   `json:"readiness_level"`
	CriticalIssues []string `json:"critical_issues"`
	ActionRequired bool     `json:"action_required"`
}

// ProductionCheckResult represents health and compliance check results
type ProductionCheckResult struct {
	Timestamp         time.Time                         `json:"timestamp"`
	HealthChecks      map[string]production.CheckResult `json:"health_checks"`
	ConfigValidation  *production.ValidationResult      `json:"config_validation"`
	ComponentsChecked []string                          `json:"components_checked"`
	OverallStatus     string                            `json:"overall_status"`
	IssuesFound       int                               `json:"issues_found"`
	FixesApplied      []string                          `json:"fixes_applied"`
}

// DeploymentPlanInfo represents deployment planning information
type DeploymentPlanInfo struct {
	Strategy            string                                    `json:"strategy"`
	Environment         string                                    `json:"environment"`
	ReadinessAssessment *production.DeploymentReadinessAssessment `json:"readiness_assessment"`
	DeploymentPlan      *DeploymentPlan                           `json:"deployment_plan"`
	RiskAssessment      *RiskAssessment                           `json:"risk_assessment"`
	RollbackPlan        *RollbackPlan                             `json:"rollback_plan"`
	EstimatedDuration   time.Duration                             `json:"estimated_duration"`
	Prerequisites       []string                                  `json:"prerequisites"`
}

// DeploymentPlan represents detailed deployment execution plan
type DeploymentPlan struct {
	Strategy   string           `json:"strategy"`
	Steps      []DeploymentStep `json:"steps"`
	Rollout    *RolloutStrategy `json:"rollout"`
	Validation *ValidationSteps `json:"validation"`
	Monitoring *MonitoringPlan  `json:"monitoring"`
}

// DeploymentStep represents individual deployment step
type DeploymentStep struct {
	Order        int           `json:"order"`
	Name         string        `json:"name"`
	Description  string        `json:"description"`
	Duration     time.Duration `json:"duration"`
	Dependencies []string      `json:"dependencies"`
	Commands     []string      `json:"commands"`
	Validation   []string      `json:"validation"`
}

// RolloutStrategy represents deployment rollout configuration
type RolloutStrategy struct {
	Type           string        `json:"type"`
	BatchSize      int           `json:"batch_size"`
	BatchDelay     time.Duration `json:"batch_delay"`
	MaxSurge       string        `json:"max_surge"`
	MaxUnavailable string        `json:"max_unavailable"`
}

// ValidationSteps represents deployment validation steps
type ValidationSteps struct {
	PreDeployment  []string `json:"pre_deployment"`
	PostDeployment []string `json:"post_deployment"`
	HealthChecks   []string `json:"health_checks"`
	Smoke          []string `json:"smoke_tests"`
}

// MonitoringPlan represents deployment monitoring configuration
type MonitoringPlan struct {
	Metrics    []string `json:"metrics"`
	Alerts     []string `json:"alerts"`
	Dashboards []string `json:"dashboards"`
	SLIs       []string `json:"slis"`
}

// RiskAssessment represents deployment risk analysis
type RiskAssessment struct {
	OverallRisk     string       `json:"overall_risk"`
	RiskFactors     []RiskFactor `json:"risk_factors"`
	Mitigations     []string     `json:"mitigations"`
	ContingencyPlan string       `json:"contingency_plan"`
}

// RiskFactor represents individual risk factor
type RiskFactor struct {
	Name        string `json:"name"`
	Level       string `json:"level"`
	Impact      string `json:"impact"`
	Probability string `json:"probability"`
	Mitigation  string `json:"mitigation"`
}

// RollbackPlan represents deployment rollback strategy
type RollbackPlan struct {
	Strategy          string        `json:"strategy"`
	TriggerConditions []string      `json:"trigger_conditions"`
	Steps             []string      `json:"steps"`
	Duration          time.Duration `json:"duration"`
	DataRecovery      string        `json:"data_recovery"`
}

// MetricsCollectionResult represents metrics collection results
type MetricsCollectionResult struct {
	Type               string                         `json:"type"`
	Timestamp          time.Time                      `json:"timestamp"`
	SystemHealth       *production.SystemHealthStatus `json:"system_health,omitempty"`
	Performance        *production.PerformanceMetrics `json:"performance,omitempty"`
	Security           *production.SecurityMetrics    `json:"security,omitempty"`
	Integration        *production.IntegrationMetrics `json:"integration,omitempty"`
	Analytics          *production.AnalyticsMetrics   `json:"analytics,omitempty"`
	CollectionDuration time.Duration                  `json:"collection_duration"`
}

// Production readiness levels
const (
	ReadinessLevelNotReady  = "not_ready"
	ReadinessLevelPartial   = "partial"
	ReadinessLevelReady     = "ready"
	ReadinessLevelOptimized = "optimized"
)

// Health status levels
const (
	HealthStatusHealthy  = "healthy"
	HealthStatusWarning  = "warning"
	HealthStatusCritical = "critical"
	HealthStatusUnknown  = "unknown"
)

// Risk levels
const (
	RiskLevelLow      = "low"
	RiskLevelMedium   = "medium"
	RiskLevelHigh     = "high"
	RiskLevelCritical = "critical"
)
