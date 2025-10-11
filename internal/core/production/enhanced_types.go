package production

import (
	"time"
)

// Enhanced Production System Types

// EnhancedProductionSystemMetrics extends base metrics with integration and automation data
type EnhancedProductionSystemMetrics struct {
	BaseMetrics                ProductionSystemMetrics  `json:"base_metrics"`
	IntegrationSystemMetrics   IntegrationSystemMetrics `json:"integration_metrics"`
	AutomationMetrics          AutomationMetrics        `json:"automation_metrics"`
	OperationalMetrics         OperationalMetrics       `json:"operational_metrics"`
	EnhancedReadinessScore     int                      `json:"enhanced_readiness_score"`
	IntegrationRecommendations []string                 `json:"integration_recommendations"`
	CollectionTimestamp        time.Time                `json:"collection_timestamp"`
	ProcessingTimeMs           int64                    `json:"processing_time_ms"`
}

// IntegrationSystemMetrics represents comprehensive integration metrics
type IntegrationSystemMetrics struct {
	TotalIntegrations  int       `json:"total_integrations"`
	ConnectedSystems   int       `json:"connected_systems"`
	HealthySystems     int       `json:"healthy_systems"`
	ActiveAlerts       int       `json:"active_alerts"`
	ActiveWorkflows    int       `json:"active_workflows"`
	IntegrationHealth  float64   `json:"integration_health"`   // Percentage
	AlertEffectiveness float64   `json:"alert_effectiveness"`  // Percentage
	WorkflowSuccess    float64   `json:"workflow_success"`     // Percentage
	DataFlowThroughput float64   `json:"data_flow_throughput"` // MB/s
	ConnectionUptime   float64   `json:"connection_uptime"`    // Percentage
	ErrorRate          float64   `json:"error_rate"`           // Percentage
	LastHealthCheck    time.Time `json:"last_health_check"`
}

// AutomationMetrics represents automation coverage and effectiveness
type AutomationMetrics struct {
	AutomatedProcesses      int     `json:"automated_processes"`
	ManualProcesses         int     `json:"manual_processes"`
	AutomationCoverage      float64 `json:"automation_coverage"`       // Percentage
	SavingsPerMonth         float64 `json:"savings_per_month"`         // USD
	ProcessingTimeReduction float64 `json:"processing_time_reduction"` // Percentage
	ErrorReduction          float64 `json:"error_reduction"`           // Percentage
	ComplianceAutomation    float64 `json:"compliance_automation"`     // Percentage
	SecurityAutomation      float64 `json:"security_automation"`       // Percentage
	MonitoringAutomation    float64 `json:"monitoring_automation"`     // Percentage
	BackupAutomation        float64 `json:"backup_automation"`         // Percentage
	RecoveryAutomation      float64 `json:"recovery_automation"`       // Percentage
	ScalingAutomation       float64 `json:"scaling_automation"`        // Percentage
	CostOptimization        float64 `json:"cost_optimization"`         // Percentage
	AlertAutomation         float64 `json:"alert_automation"`          // Percentage
	ReportingAutomation     float64 `json:"reporting_automation"`      // Percentage
}

// OperationalMetrics represents operational excellence metrics
type OperationalMetrics struct {
	MTBF                  float64 `json:"mtbf"`                    // Mean Time Between Failures (hours)
	MTTR                  float64 `json:"mttr"`                    // Mean Time To Recovery (minutes)
	ChangeSuccessRate     float64 `json:"change_success_rate"`     // Percentage
	RollbackRate          float64 `json:"rollback_rate"`           // Percentage
	IncidentCount         int     `json:"incident_count"`          // Total incidents
	P1IncidentCount       int     `json:"p1_incident_count"`       // Critical incidents
	P2IncidentCount       int     `json:"p2_incident_count"`       // High priority incidents
	P3IncidentCount       int     `json:"p3_incident_count"`       // Medium priority incidents
	SLACompliance         float64 `json:"sla_compliance"`          // Percentage
	CapacityUtilization   float64 `json:"capacity_utilization"`    // Percentage
	CostPerTransaction    float64 `json:"cost_per_transaction"`    // USD
	CustomerSatisfaction  float64 `json:"customer_satisfaction"`   // 1-5 scale
	TeamProductivity      float64 `json:"team_productivity"`       // Percentage
	KnowledgeBaseAccuracy float64 `json:"knowledge_base_accuracy"` // Percentage
	DocumentationCoverage float64 `json:"documentation_coverage"`  // Percentage
	TrainingCompletion    float64 `json:"training_completion"`     // Percentage
}

// Enhanced Optimization Types

// EnhancedOptimizationOptions extends base optimization options
type EnhancedOptimizationOptions struct {
	*OptimizationOptions           `json:"base_options"`
	IncludeIntegrationOptimization bool     `json:"include_integration_optimization"`
	IncludeAutomationOptimization  bool     `json:"include_automation_optimization"`
	IncludeWorkflowOptimization    bool     `json:"include_workflow_optimization"`
	IncludeOperationalOptimization bool     `json:"include_operational_optimization"`
	IntegrationFocus               []string `json:"integration_focus"`     // e.g., ["latency", "throughput", "reliability"]
	AutomationPriorities           []string `json:"automation_priorities"` // e.g., ["cost", "time", "quality"]
}

// IntegratedOptimizationReport combines all optimization results
type IntegratedOptimizationReport struct {
	BaseOptimizationReport       ProductionOptimizationReport    `json:"base_optimization"`
	IntegrationOptimization      *IntegrationOptimizationResults `json:"integration_optimization,omitempty"`
	AutomationOptimization       *AutomationOptimizationResults  `json:"automation_optimization,omitempty"`
	WorkflowOptimization         *WorkflowOptimizationResults    `json:"workflow_optimization,omitempty"`
	IntegratedROI                *IntegratedROIAnalysis          `json:"integrated_roi"`
	StrategicRecommendations     []StrategicRecommendation       `json:"strategic_recommendations"`
	OptimizationTimestamp        time.Time                       `json:"optimization_timestamp"`
	ProcessingTimeMs             int64                           `json:"processing_time_ms"`
	OverallEfficiencyImprovement float64                         `json:"overall_efficiency_improvement"` // Percentage
}

// IntegrationOptimizationResults represents integration-specific optimizations
type IntegrationOptimizationResults struct {
	OptimizedConnections int      `json:"optimized_connections"`
	ReducedLatency       float64  `json:"reduced_latency"`     // Percentage
	ImprovedThroughput   float64  `json:"improved_throughput"` // Percentage
	CostSavings          float64  `json:"cost_savings"`        // USD
	Recommendations      []string `json:"recommendations"`
}

// AutomationOptimizationResults represents automation-specific optimizations
type AutomationOptimizationResults struct {
	AutomatableProcesses int      `json:"automatable_processes"`
	PotentialSavings     float64  `json:"potential_savings"` // USD
	EfficiencyGain       float64  `json:"efficiency_gain"`   // Percentage
	ROI                  float64  `json:"roi"`               // Percentage
	Recommendations      []string `json:"recommendations"`
}

// WorkflowOptimizationResults represents workflow-specific optimizations
type WorkflowOptimizationResults struct {
	OptimizedWorkflows int      `json:"optimized_workflows"`
	TimeReduction      float64  `json:"time_reduction"`      // Percentage
	CostReduction      float64  `json:"cost_reduction"`      // Percentage
	QualityImprovement float64  `json:"quality_improvement"` // Percentage
	Recommendations    []string `json:"recommendations"`
}

// IntegratedROIAnalysis provides comprehensive ROI analysis
type IntegratedROIAnalysis struct {
	TotalInvestment float64 `json:"total_investment"` // USD
	TotalSavings    float64 `json:"total_savings"`    // USD
	ROIPercentage   float64 `json:"roi_percentage"`   // Percentage
	PaybackMonths   int     `json:"payback_months"`   // Months
	NPV             float64 `json:"npv"`              // Net Present Value (USD)
	IRR             float64 `json:"irr"`              // Internal Rate of Return (Percentage)
}

// StrategicRecommendation represents a strategic recommendation
type StrategicRecommendation struct {
	Priority    string `json:"priority"` // high, medium, low
	Category    string `json:"category"` // integration, automation, security, etc.
	Title       string `json:"title"`
	Description string `json:"description"`
	Impact      string `json:"impact"`   // high, medium, low
	Effort      string `json:"effort"`   // high, medium, low
	Timeline    string `json:"timeline"` // e.g., "2-3 months"
}

// Enhanced Deployment Types

// IntegratedDeploymentOptions extends deployment options with integration considerations
type IntegratedDeploymentOptions struct {
	*DeploymentRequirements      `json:"base_requirements"`
	RequireIntegrationValidation bool     `json:"require_integration_validation"`
	RequireAutomationValidation  bool     `json:"require_automation_validation"`
	RequireOperationalValidation bool     `json:"require_operational_validation"`
	IntegrationTestingRequired   bool     `json:"integration_testing_required"`
	AutomationDeploymentRequired bool     `json:"automation_deployment_required"`
	OperationalReadinessRequired bool     `json:"operational_readiness_required"`
	CriticalIntegrations         []string `json:"critical_integrations"`
	AutomationDependencies       []string `json:"automation_dependencies"`
	OperationalPrerequisites     []string `json:"operational_prerequisites"`
}

// IntegratedDeploymentReadinessAssessment combines all deployment readiness aspects
type IntegratedDeploymentReadinessAssessment struct {
	BaseAssessment       DeploymentReadinessAssessment  `json:"base_assessment"`
	IntegrationReadiness IntegrationDeploymentReadiness `json:"integration_readiness"`
	AutomationReadiness  AutomationDeploymentReadiness  `json:"automation_readiness"`
	OperationalReadiness OperationalDeploymentReadiness `json:"operational_readiness"`
	IntegratedScore      int                            `json:"integrated_score"`
	EnhancedPlan         *EnhancedDeploymentPlan        `json:"enhanced_plan"`
	AssessmentTimestamp  time.Time                      `json:"assessment_timestamp"`
	ProcessingTimeMs     int64                          `json:"processing_time_ms"`
}

// IntegrationDeploymentReadiness represents integration-specific deployment readiness
type IntegrationDeploymentReadiness struct {
	ConnectionsReady   bool     `json:"connections_ready"`
	AlertsConfigured   bool     `json:"alerts_configured"`
	WorkflowsValidated bool     `json:"workflows_validated"`
	MonitoringEnabled  bool     `json:"monitoring_enabled"`
	BackupsConfigured  bool     `json:"backups_configured"`
	SecurityValidated  bool     `json:"security_validated"`
	ReadinessScore     int      `json:"readiness_score"`
	BlockingIssues     []string `json:"blocking_issues"`
	Recommendations    []string `json:"recommendations"`
}

// AutomationDeploymentReadiness represents automation-specific deployment readiness
type AutomationDeploymentReadiness struct {
	AutomationReady     bool     `json:"automation_ready"`
	ScriptsValidated    bool     `json:"scripts_validated"`
	SchedulesConfigured bool     `json:"schedules_configured"`
	ErrorHandling       bool     `json:"error_handling"`
	RollbackReady       bool     `json:"rollback_ready"`
	ReadinessScore      int      `json:"readiness_score"`
	BlockingIssues      []string `json:"blocking_issues"`
	Recommendations     []string `json:"recommendations"`
}

// OperationalDeploymentReadiness represents operational-specific deployment readiness
type OperationalDeploymentReadiness struct {
	OperationalReady   bool     `json:"operational_ready"`
	DocumentationReady bool     `json:"documentation_ready"`
	TeamTrained        bool     `json:"team_trained"`
	ProcessesDefined   bool     `json:"processes_defined"`
	SupportReady       bool     `json:"support_ready"`
	ReadinessScore     int      `json:"readiness_score"`
	BlockingIssues     []string `json:"blocking_issues"`
	Recommendations    []string `json:"recommendations"`
}

// EnhancedDeploymentPlan extends deployment plan with integration considerations
type EnhancedDeploymentPlan struct {
	BaseSteps         []DeploymentStep            `json:"base_steps"`
	IntegrationSteps  []IntegrationDeploymentStep `json:"integration_steps"`
	AutomationSteps   []AutomationDeploymentStep  `json:"automation_steps"`
	RollbackStrategy  string                      `json:"rollback_strategy"`
	EstimatedDuration time.Duration               `json:"estimated_duration"`
	RiskLevel         string                      `json:"risk_level"`
	ApprovalRequired  bool                        `json:"approval_required"`
}

// IntegrationDeploymentStep represents a deployment step specific to integrations
type IntegrationDeploymentStep struct {
	Name     string        `json:"name"`
	Duration time.Duration `json:"duration"`
	Critical bool          `json:"critical"`
}

// AutomationDeploymentStep represents a deployment step specific to automation
type AutomationDeploymentStep struct {
	Name     string        `json:"name"`
	Duration time.Duration `json:"duration"`
	Critical bool          `json:"critical"`
}

// Enhanced Report Types

// EnhancedReportOptions extends report options with integration insights
type EnhancedReportOptions struct {
	*ReportOptions                `json:"base_options"`
	IncludeIntegrationAnalysis    bool   `json:"include_integration_analysis"`
	IncludeAutomationAnalysis     bool   `json:"include_automation_analysis"`
	IncludeOperationalAnalysis    bool   `json:"include_operational_analysis"`
	IncludeStrategicRoadmap       bool   `json:"include_strategic_roadmap"`
	IncludeBusinessImpactAnalysis bool   `json:"include_business_impact_analysis"`
	IntegrationDetailLevel        string `json:"integration_detail_level"` // summary, detailed, comprehensive
	AutomationDetailLevel         string `json:"automation_detail_level"`  // summary, detailed, comprehensive
}

// EnhancedExecutiveReport combines all reporting aspects
type EnhancedExecutiveReport struct {
	BaseReport          ExecutiveReport           `json:"base_report"`
	IntegrationAnalysis *IntegrationAnalysis      `json:"integration_analysis,omitempty"`
	AutomationAnalysis  *AutomationAnalysis       `json:"automation_analysis,omitempty"`
	OperationalAnalysis *OperationalAnalysis      `json:"operational_analysis,omitempty"`
	StrategicRoadmap    *StrategicRoadmap         `json:"strategic_roadmap,omitempty"`
	BusinessImpact      *BusinessImpactAssessment `json:"business_impact,omitempty"`
	ReportTimestamp     time.Time                 `json:"report_timestamp"`
	ProcessingTimeMs    int64                     `json:"processing_time_ms"`
}

// IntegrationAnalysis provides comprehensive integration analysis
type IntegrationAnalysis struct {
	TotalIntegrations    int                `json:"total_integrations"`
	HealthyIntegrations  int                `json:"healthy_integrations"`
	CriticalIntegrations int                `json:"critical_integrations"`
	PerformanceScore     float64            `json:"performance_score"`
	ReliabilityScore     float64            `json:"reliability_score"`
	CostEfficiencyScore  float64            `json:"cost_efficiency_score"`
	SecurityScore        float64            `json:"security_score"`
	Recommendations      []string           `json:"recommendations"`
	TrendAnalysis        map[string]float64 `json:"trend_analysis"`
}

// AutomationAnalysis provides comprehensive automation analysis
type AutomationAnalysis struct {
	AutomationCoverage  float64  `json:"automation_coverage"` // Percentage
	EfficiencyGains     float64  `json:"efficiency_gains"`    // Percentage
	CostSavings         float64  `json:"cost_savings"`        // USD per month
	QualityImprovement  float64  `json:"quality_improvement"` // Percentage
	TimeReduction       float64  `json:"time_reduction"`      // Percentage
	ErrorReduction      float64  `json:"error_reduction"`     // Percentage
	ROI                 float64  `json:"roi"`                 // Percentage
	Recommendations     []string `json:"recommendations"`
	FutureOpportunities []string `json:"future_opportunities"`
}

// OperationalAnalysis provides comprehensive operational analysis
type OperationalAnalysis struct {
	OperationalExcellence float64  `json:"operational_excellence"` // Score 0-100
	ProcessMaturity       float64  `json:"process_maturity"`       // Score 0-100
	TeamEffectiveness     float64  `json:"team_effectiveness"`     // Score 0-100
	KnowledgeManagement   float64  `json:"knowledge_management"`   // Score 0-100
	ContinuousImprovement float64  `json:"continuous_improvement"` // Score 0-100
	CustomerImpact        float64  `json:"customer_impact"`        // 1-5 scale
	BusinessValue         string   `json:"business_value"`         // high, medium, low
	Recommendations       []string `json:"recommendations"`
	MaturityRoadmap       []string `json:"maturity_roadmap"`
}

// StrategicRoadmap provides strategic planning guidance
type StrategicRoadmap struct {
	CurrentState    string         `json:"current_state"` // e.g., "advanced"
	TargetState     string         `json:"target_state"`  // e.g., "world_class"
	Phases          []RoadmapPhase `json:"phases"`
	TotalInvestment float64        `json:"total_investment"` // USD
	ExpectedROI     float64        `json:"expected_roi"`     // Percentage
	BusinessImpact  string         `json:"business_impact"`  // transformational, significant, moderate
}

// RoadmapPhase represents a phase in the strategic roadmap
type RoadmapPhase struct {
	Name         string   `json:"name"`
	Duration     string   `json:"duration"` // e.g., "3 months"
	Objectives   []string `json:"objectives"`
	Deliverables []string `json:"deliverables"`
	Investment   float64  `json:"investment"` // USD
}

// BusinessImpactAssessment provides business impact analysis
type BusinessImpactAssessment struct {
	RevenueImpact        string   `json:"revenue_impact"`
	CostImpact           string   `json:"cost_impact"`
	RiskImpact           string   `json:"risk_impact"`
	QualityImpact        string   `json:"quality_impact"`
	TimeToMarketImpact   string   `json:"time_to_market_impact"`
	CustomerImpact       string   `json:"customer_impact"`
	CompetitiveAdvantage string   `json:"competitive_advantage"`
	StrategicAlignment   string   `json:"strategic_alignment"`
	Recommendations      []string `json:"recommendations"`
	OverallAssessment    string   `json:"overall_assessment"`
}
