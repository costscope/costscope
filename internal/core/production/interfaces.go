package production

import (
	"context"
	"time"
)

// ProductionService defines the interface for production readiness assessment
type ProductionService interface {
	// GetSystemStatus returns comprehensive system status and health metrics
	GetSystemStatus(ctx context.Context) (*ProductionSystemMetrics, error)

	// RunOptimization runs comprehensive system optimization analysis
	RunOptimization(ctx context.Context, options *OptimizationOptions) (*ProductionOptimizationReport, error)

	// AssessDeploymentReadiness assesses system readiness for production deployment
	AssessDeploymentReadiness(ctx context.Context, environment string) (*DeploymentReadinessAssessment, error)

	// GenerateExecutiveReport generates executive-level comprehensive report
	GenerateExecutiveReport(ctx context.Context, options *ReportOptions) (*ExecutiveReport, error)

	// ValidateProductionConfiguration validates production configuration
	ValidateProductionConfiguration(ctx context.Context) (*ValidationResult, error)

	// GetHealthChecks returns current health check results
	GetHealthChecks(ctx context.Context) (map[string]CheckResult, error)
}

// MetricsCollector defines the interface for collecting system metrics
type MetricsCollector interface {
	// CollectSystemHealth collects system health metrics
	CollectSystemHealth(ctx context.Context) (*SystemHealthStatus, error)

	// CollectPerformanceMetrics collects performance metrics
	CollectPerformanceMetrics(ctx context.Context) (*PerformanceMetrics, error)

	// CollectSecurityMetrics collects security assessment metrics
	CollectSecurityMetrics(ctx context.Context) (*SecurityMetrics, error)

	// CollectIntegrationMetrics collects integration metrics
	CollectIntegrationMetrics(ctx context.Context) (*IntegrationMetrics, error)

	// CollectAnalyticsMetrics collects analytics metrics
	CollectAnalyticsMetrics(ctx context.Context) (*AnalyticsMetrics, error)
}

// OptimizationEngine defines the interface for system optimization
type OptimizationEngine interface {
	// AnalyzeOptimizations analyzes potential system optimizations
	AnalyzeOptimizations(ctx context.Context, options *OptimizationOptions) (*OptimizationResults, error)

	// GenerateRecommendations generates optimization recommendations
	GenerateRecommendations(ctx context.Context, metrics *ProductionSystemMetrics) ([]ProductionRecommendation, error)

	// CalculateROI calculates return on investment for recommendations
	CalculateROI(ctx context.Context, recommendations []ProductionRecommendation) (*ROIAnalysis, error)

	// PlanRoadmap plans future development roadmap
	PlanRoadmap(ctx context.Context, currentState *ProductionSystemMetrics) ([]RoadmapItem, error)
}

// DeploymentAssessor defines the interface for deployment readiness assessment
type DeploymentAssessor interface {
	// AssessReadiness assesses deployment readiness
	AssessReadiness(ctx context.Context, requirements *DeploymentRequirements) (*DeploymentReadinessAssessment, error)

	// ValidateEnvironment validates deployment environment
	ValidateEnvironment(ctx context.Context, environment string) (*EnvironmentValidation, error)

	// RunHealthChecks runs comprehensive health checks
	RunHealthChecks(ctx context.Context, components []string) (*HealthCheckResults, error)

	// GenerateDeploymentPlan generates deployment execution plan
	GenerateDeploymentPlan(ctx context.Context, strategy string, requirements *DeploymentRequirements) (*DeploymentPlan, error)
}

// Request and configuration types

// OptimizationOptions represents options for optimization analysis
type OptimizationOptions struct {
	Aggressive     bool     `json:"aggressive"`
	DryRun         bool     `json:"dry_run"`
	Categories     []string `json:"categories"` // performance, security, cost, integration
	MinImpact      Impact   `json:"min_impact"`
	MaxEffort      Effort   `json:"max_effort"`
	Budget         float64  `json:"budget"`
	TimelineMonths int      `json:"timeline_months"`
	IncludeRoadmap bool     `json:"include_roadmap"`
}

// ReportOptions represents options for report generation
type ReportOptions struct {
	Format          string   `json:"format"`           // json, pdf, html, markdown
	IncludeSections []string `json:"include_sections"` // overview, metrics, recommendations, roadmap
	DetailLevel     string   `json:"detail_level"`     // summary, detailed, comprehensive
	Audience        string   `json:"audience"`         // executive, technical, operational
	OutputPath      string   `json:"output_path"`
	IncludeCharts   bool     `json:"include_charts"`
	IncludeAppendix bool     `json:"include_appendix"`
}

// ValidationResult represents validation results
type ValidationResult struct {
	Valid            bool                   `json:"valid"`
	Score            int                    `json:"score"` // 0-100
	Issues           []Issue                `json:"issues"`
	Warnings         []Issue                `json:"warnings"`
	Recommendations  []string               `json:"recommendations"`
	CheckResults     map[string]CheckResult `json:"check_results"`
	ProcessingTimeMs int64                  `json:"processing_time_ms"`
}

// ExecutiveReport represents executive-level report
type ExecutiveReport struct {
	GeneratedAt      time.Time        `json:"generated_at"`
	Executive        ExecutiveSummary `json:"executive_summary"`
	SystemOverview   SystemOverview   `json:"system_overview"`
	KeyMetrics       KeyMetrics       `json:"key_metrics"`
	Achievements     []Achievement    `json:"achievements"`
	Challenges       []Challenge      `json:"challenges"`
	Recommendations  []StrategicRec   `json:"recommendations"`
	ROIAnalysis      ROIAnalysis      `json:"roi_analysis"`
	FutureOutlook    FutureOutlook    `json:"future_outlook"`
	Appendices       []Appendix       `json:"appendices"`
	ProcessingTimeMs int64            `json:"processing_time_ms"`
}

// ExecutiveSummary represents executive summary
type ExecutiveSummary struct {
	OverallHealth     string   `json:"overall_health"`
	BusinessValue     string   `json:"business_value"`
	KeyWins           []string `json:"key_wins"`
	PrimaryChallenges []string `json:"primary_challenges"`
	StrategicFocus    []string `json:"strategic_focus"`
	ROIHighlights     string   `json:"roi_highlights"`
	NextQuarter       string   `json:"next_quarter"`
}

// KeyMetrics represents key business metrics
type KeyMetrics struct {
	ProductionReadiness float64            `json:"production_readiness"` // 0-100
	SystemHealth        float64            `json:"system_health"`        // 0-100
	SecurityPosture     float64            `json:"security_posture"`     // 0-100
	OperationalMaturity float64            `json:"operational_maturity"` // 0-100
	CostEfficiency      float64            `json:"cost_efficiency"`      // 0-100
	TrendAnalysis       map[string]float64 `json:"trend_analysis"`       // month-over-month changes
}

// Achievement represents a significant achievement
type Achievement struct {
	Title        string   `json:"title"`
	Description  string   `json:"description"`
	Impact       string   `json:"impact"`
	Metrics      []string `json:"metrics"`
	Timeline     string   `json:"timeline"`
	Contributors []string `json:"contributors"`
}

// Challenge represents a significant challenge
type Challenge struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Impact      Impact   `json:"impact"`
	Priority    Priority `json:"priority"`
	Actions     []string `json:"actions"`
	Timeline    string   `json:"timeline"`
	Owner       string   `json:"owner"`
}

// StrategicRec represents strategic recommendation
type StrategicRec struct {
	Title        string   `json:"title"`
	Description  string   `json:"description"`
	BusinessCase string   `json:"business_case"`
	Priority     Priority `json:"priority"`
	Investment   float64  `json:"investment"`
	ExpectedROI  float64  `json:"expected_roi"`
	Timeline     string   `json:"timeline"`
	Success      []string `json:"success_criteria"`
}

// FutureOutlook represents future outlook and strategy
type FutureOutlook struct {
	Vision              string             `json:"vision"`
	StrategicGoals      []string           `json:"strategic_goals"`
	TechnologyTrends    []string           `json:"technology_trends"`
	MarketOpportunities []string           `json:"market_opportunities"`
	RiskFactors         []string           `json:"risk_factors"`
	Investments         map[string]float64 `json:"planned_investments"`
	Timeline            string             `json:"timeline"`
}

// Appendix represents report appendix
type Appendix struct {
	Title   string      `json:"title"`
	Type    string      `json:"type"` // metrics, technical, financial
	Content interface{} `json:"content"`
}
