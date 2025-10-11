package multicloud

import (
	"time"
)

// Result types for multicloud operations

// OptimizationResult represents the result of optimization analysis
type OptimizationResult struct {
	RequestID        string                    `json:"request_id"`
	Timestamp        time.Time                 `json:"timestamp"`
	Providers        []string                  `json:"providers"`
	TotalSavings     float64                   `json:"total_savings"`
	Currency         string                    `json:"currency"`
	AnalysisPeriod   DateRange                 `json:"analysis_period"`
	Optimizations    []OptimizationOpportunity `json:"optimizations"`
	Summary          OptimizationSummary       `json:"summary"`
	Recommendations  []Recommendation          `json:"recommendations"`
	ProcessingTimeMs int64                     `json:"processing_time_ms"`
}

// CostComparisonResult represents cost comparison across providers
type CostComparisonResult struct {
	RequestID        string             `json:"request_id"`
	Timestamp        time.Time          `json:"timestamp"`
	ComparisonPeriod DateRange          `json:"comparison_period"`
	Currency         string             `json:"currency"`
	Providers        []ProviderCostData `json:"providers"`
	TotalCosts       map[string]float64 `json:"total_costs"`
	CostBreakdown    CostBreakdown      `json:"cost_breakdown"`
	Trends           []CostTrend        `json:"trends"`
	BestValue        ProviderRanking    `json:"best_value"`
	ProcessingTimeMs int64              `json:"processing_time_ms"`
}

// MigrationEstimate represents migration cost estimation
type MigrationEstimate struct {
	RequestID          string           `json:"request_id"`
	Timestamp          time.Time        `json:"timestamp"`
	SourceProvider     string           `json:"source_provider"`
	TargetProvider     string           `json:"target_provider"`
	EstimatedDuration  time.Duration    `json:"estimated_duration"`
	MigrationCosts     MigrationCosts   `json:"migration_costs"`
	OngoingSavings     float64          `json:"ongoing_savings"`
	BreakevenTimeframe time.Duration    `json:"breakeven_timeframe"`
	RiskAssessment     RiskAssessment   `json:"risk_assessment"`
	Timeline           []MigrationPhase `json:"timeline"`
	ProcessingTimeMs   int64            `json:"processing_time_ms"`
}

// FeasibilityAnalysis represents migration feasibility analysis
type FeasibilityAnalysis struct {
	RequestID          string                `json:"request_id"`
	Timestamp          time.Time             `json:"timestamp"`
	OverallFeasibility FeasibilityLevel      `json:"overall_feasibility"`
	FeasibilityScore   float64               `json:"feasibility_score"`
	ResourceAnalysis   []ResourceFeasibility `json:"resource_analysis"`
	TechnicalBarriers  []TechnicalBarrier    `json:"technical_barriers"`
	CostBenefit        CostBenefitAnalysis   `json:"cost_benefit"`
	Recommendations    []string              `json:"recommendations"`
	ProcessingTimeMs   int64                 `json:"processing_time_ms"`
}

// MigrationPlan represents a detailed migration plan
type MigrationPlan struct {
	RequestID         string           `json:"request_id"`
	Timestamp         time.Time        `json:"timestamp"`
	PlanID            string           `json:"plan_id"`
	SourceProvider    string           `json:"source_provider"`
	TargetProvider    string           `json:"target_provider"`
	EstimatedCost     float64          `json:"estimated_cost"`
	EstimatedDuration time.Duration    `json:"estimated_duration"`
	Phases            []MigrationPhase `json:"phases"`
	Dependencies      []Dependency     `json:"dependencies"`
	RiskMitigation    []RiskMitigation `json:"risk_mitigation"`
	Rollback          RollbackPlan     `json:"rollback"`
	ProcessingTimeMs  int64            `json:"processing_time_ms"`
}

// DiscoveryResult represents resource discovery results
type DiscoveryResult struct {
	RequestID           string               `json:"request_id"`
	Timestamp           time.Time            `json:"timestamp"`
	Providers           []string             `json:"providers"`
	TotalResources      int                  `json:"total_resources"`
	ResourcesByType     map[string]int       `json:"resources_by_type"`
	ResourcesByProvider map[string]int       `json:"resources_by_provider"`
	Resources           []DiscoveredResource `json:"resources"`
	CostSummary         *ResourceCostSummary `json:"cost_summary,omitempty"`
	ProcessingTimeMs    int64                `json:"processing_time_ms"`
}

// ProviderScanResult represents scan results for a specific provider
type ProviderScanResult struct {
	ProviderName     string               `json:"provider_name"`
	Region           string               `json:"region"`
	ScanTimestamp    time.Time            `json:"scan_timestamp"`
	ResourceCount    int                  `json:"resource_count"`
	Resources        []DiscoveredResource `json:"resources"`
	Errors           []string             `json:"errors,omitempty"`
	ProcessingTimeMs int64                `json:"processing_time_ms"`
}

// UnifiedInventory represents unified inventory across all providers
type UnifiedInventory struct {
	Timestamp      time.Time                   `json:"timestamp"`
	TotalResources int                         `json:"total_resources"`
	TotalCost      float64                     `json:"total_cost"`
	Currency       string                      `json:"currency"`
	Providers      []ProviderInventory         `json:"providers"`
	ResourceTypes  map[string]ResourceTypeInfo `json:"resource_types"`
	CostBreakdown  CostBreakdown               `json:"cost_breakdown"`
	LastUpdated    time.Time                   `json:"last_updated"`
}

// RecommendationResult represents optimization recommendations
type RecommendationResult struct {
	RequestID             string           `json:"request_id"`
	Timestamp             time.Time        `json:"timestamp"`
	TotalRecommendations  int              `json:"total_recommendations"`
	TotalPotentialSavings float64          `json:"total_potential_savings"`
	Currency              string           `json:"currency"`
	Recommendations       []Recommendation `json:"recommendations"`
	ProcessingTimeMs      int64            `json:"processing_time_ms"`
}

// Supporting types

// DateRange represents a date range
type DateRange struct {
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
}

// OptimizationOpportunity represents a specific optimization opportunity
type OptimizationOpportunity struct {
	ID                string             `json:"id"`
	Type              OptimizationType   `json:"type"`
	Provider          string             `json:"provider"`
	Resource          ResourceSpec       `json:"resource"`
	CurrentCost       float64            `json:"current_cost"`
	OptimizedCost     float64            `json:"optimized_cost"`
	PotentialSavings  float64            `json:"potential_savings"`
	SavingsPercentage float64            `json:"savings_percentage"`
	Risk              RiskLevel          `json:"risk"`
	Confidence        float64            `json:"confidence"`
	Description       string             `json:"description"`
	Implementation    ImplementationInfo `json:"implementation"`
}

// OptimizationSummary provides summary statistics
type OptimizationSummary struct {
	TotalOpportunities    int                          `json:"total_opportunities"`
	OpportunitiesByType   map[OptimizationType]int     `json:"opportunities_by_type"`
	OpportunitiesByRisk   map[RiskLevel]int            `json:"opportunities_by_risk"`
	SavingsByProvider     map[string]float64           `json:"savings_by_provider"`
	SavingsByType         map[OptimizationType]float64 `json:"savings_by_type"`
	HighConfidenceSavings float64                      `json:"high_confidence_savings"`
	LowRiskSavings        float64                      `json:"low_risk_savings"`
}

// Recommendation represents an actionable recommendation
type Recommendation struct {
	ID               string         `json:"id"`
	Type             string         `json:"type"`
	Priority         Priority       `json:"priority"`
	Title            string         `json:"title"`
	Description      string         `json:"description"`
	PotentialSavings float64        `json:"potential_savings"`
	Risk             RiskLevel      `json:"risk"`
	Effort           EffortLevel    `json:"effort"`
	Timeline         time.Duration  `json:"timeline"`
	Actions          []Action       `json:"actions"`
	Resources        []string       `json:"resources"`
	Dependencies     []string       `json:"dependencies,omitempty"`
	Validation       ValidationInfo `json:"validation"`
}

// ProviderCostData represents cost data for a provider
type ProviderCostData struct {
	ProviderName  string             `json:"provider_name"`
	TotalCost     float64            `json:"total_cost"`
	CostByService map[string]float64 `json:"cost_by_service"`
	CostByRegion  map[string]float64 `json:"cost_by_region"`
	Trends        []DataPoint        `json:"trends"`
	Metadata      ProviderMetadata   `json:"metadata"`
}

// CostBreakdown represents detailed cost breakdown
type CostBreakdown struct {
	ByProvider map[string]float64 `json:"by_provider"`
	ByService  map[string]float64 `json:"by_service"`
	ByRegion   map[string]float64 `json:"by_region"`
	ByPeriod   []PeriodCost       `json:"by_period"`
}

// CostTrend represents cost trend data
type CostTrend struct {
	Provider      string      `json:"provider"`
	Service       string      `json:"service,omitempty"`
	Trend         TrendType   `json:"trend"`
	ChangePercent float64     `json:"change_percent"`
	Period        DateRange   `json:"period"`
	DataPoints    []DataPoint `json:"data_points"`
}

// ProviderRanking represents provider ranking by value
type ProviderRanking struct {
	BestOverall     string            `json:"best_overall"`
	BestByService   map[string]string `json:"best_by_service"`
	Rankings        []ProviderRank    `json:"rankings"`
	RankingCriteria RankingCriteria   `json:"ranking_criteria"`
}

// MigrationCosts represents detailed migration costs
type MigrationCosts struct {
	DataTransfer     float64            `json:"data_transfer"`
	Downtime         float64            `json:"downtime"`
	Personnel        float64            `json:"personnel"`
	Tooling          float64            `json:"tooling"`
	Testing          float64            `json:"testing"`
	Contingency      float64            `json:"contingency"`
	Total            float64            `json:"total"`
	BreakdownByPhase map[string]float64 `json:"breakdown_by_phase"`
}

// RiskAssessment represents risk assessment for migration
type RiskAssessment struct {
	OverallRisk     RiskLevel        `json:"overall_risk"`
	RiskFactors     []RiskFactor     `json:"risk_factors"`
	MitigationPlans []RiskMitigation `json:"mitigation_plans"`
	RiskScore       float64          `json:"risk_score"`
}

// MigrationPhase represents a phase in migration timeline
type MigrationPhase struct {
	ID                string        `json:"id"`
	Name              string        `json:"name"`
	Description       string        `json:"description"`
	EstimatedDuration time.Duration `json:"estimated_duration"`
	Dependencies      []string      `json:"dependencies"`
	Resources         []string      `json:"resources"`
	Tasks             []Task        `json:"tasks"`
	Milestones        []Milestone   `json:"milestones"`
	RiskLevel         RiskLevel     `json:"risk_level"`
	Cost              float64       `json:"cost"`
}

// Additional supporting types

type FeasibilityLevel string

const (
	FeasibilityHigh   FeasibilityLevel = "high"
	FeasibilityMedium FeasibilityLevel = "medium"
	FeasibilityLow    FeasibilityLevel = "low"
)

type Priority string

const (
	PriorityHigh   Priority = "high"
	PriorityMedium Priority = "medium"
	PriorityLow    Priority = "low"
)

type EffortLevel string

const (
	EffortLow    EffortLevel = "low"
	EffortMedium EffortLevel = "medium"
	EffortHigh   EffortLevel = "high"
)

type TrendType string

const (
	TrendIncreasing TrendType = "increasing"
	TrendDecreasing TrendType = "decreasing"
	TrendStable     TrendType = "stable"
	TrendVolatile   TrendType = "volatile"
)

// Simple types for completion
type ImplementationInfo struct {
	Steps           []string      `json:"steps"`
	Duration        time.Duration `json:"duration"`
	Complexity      string        `json:"complexity"`
	AutomationLevel string        `json:"automation_level"`
}

type Action struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	Command     string `json:"command,omitempty"`
	Manual      bool   `json:"manual"`
}

type ValidationInfo struct {
	Checks       []string `json:"checks"`
	Verification string   `json:"verification"`
}

type ProviderMetadata struct {
	Region      string    `json:"region"`
	LastUpdated time.Time `json:"last_updated"`
	DataQuality float64   `json:"data_quality"`
}

type DataPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}

type PeriodCost struct {
	Period DateRange `json:"period"`
	Cost   float64   `json:"cost"`
}

type ProviderRank struct {
	Provider string  `json:"provider"`
	Rank     int     `json:"rank"`
	Score    float64 `json:"score"`
	Reason   string  `json:"reason"`
}

type RankingCriteria struct {
	CostWeight        float64 `json:"cost_weight"`
	PerformanceWeight float64 `json:"performance_weight"`
	ReliabilityWeight float64 `json:"reliability_weight"`
}

type ResourceFeasibility struct {
	ResourceID   string           `json:"resource_id"`
	Feasibility  FeasibilityLevel `json:"feasibility"`
	Barriers     []string         `json:"barriers"`
	Requirements []string         `json:"requirements"`
}

type TechnicalBarrier struct {
	Type        string    `json:"type"`
	Description string    `json:"description"`
	Severity    RiskLevel `json:"severity"`
	Solutions   []string  `json:"solutions"`
}

type CostBenefitAnalysis struct {
	TotalCosts    float64       `json:"total_costs"`
	TotalBenefits float64       `json:"total_benefits"`
	NetBenefit    float64       `json:"net_benefit"`
	ROI           float64       `json:"roi"`
	PaybackPeriod time.Duration `json:"payback_period"`
}

type Dependency struct {
	ID          string   `json:"id"`
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Criticality Priority `json:"criticality"`
}

type RiskMitigation struct {
	RiskType    string        `json:"risk_type"`
	Description string        `json:"description"`
	Actions     []string      `json:"actions"`
	Timeline    time.Duration `json:"timeline"`
}

type RollbackPlan struct {
	Triggers   []string      `json:"triggers"`
	Steps      []string      `json:"steps"`
	Duration   time.Duration `json:"duration"`
	DataBackup bool          `json:"data_backup"`
}

type DiscoveredResource struct {
	ID           string                 `json:"id"`
	Type         string                 `json:"type"`
	Name         string                 `json:"name"`
	Provider     string                 `json:"provider"`
	Region       string                 `json:"region"`
	State        string                 `json:"state"`
	Cost         *ResourceCost          `json:"cost,omitempty"`
	Tags         map[string]string      `json:"tags,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
	LastModified time.Time              `json:"last_modified"`
}

type ResourceCostSummary struct {
	TotalCost      float64            `json:"total_cost"`
	Currency       string             `json:"currency"`
	CostByProvider map[string]float64 `json:"cost_by_provider"`
	CostByType     map[string]float64 `json:"cost_by_type"`
	CostByRegion   map[string]float64 `json:"cost_by_region"`
}

type ProviderInventory struct {
	ProviderName    string         `json:"provider_name"`
	ResourceCount   int            `json:"resource_count"`
	TotalCost       float64        `json:"total_cost"`
	ResourcesByType map[string]int `json:"resources_by_type"`
	LastScan        time.Time      `json:"last_scan"`
}

type ResourceTypeInfo struct {
	Count             int            `json:"count"`
	TotalCost         float64        `json:"total_cost"`
	AverageCost       float64        `json:"average_cost"`
	ProviderBreakdown map[string]int `json:"provider_breakdown"`
}

type ResourceCost struct {
	HourlyCost  float64 `json:"hourly_cost"`
	DailyCost   float64 `json:"daily_cost"`
	MonthlyCost float64 `json:"monthly_cost"`
	Currency    string  `json:"currency"`
}

type RiskFactor struct {
	Type        string    `json:"type"`
	Description string    `json:"description"`
	Impact      RiskLevel `json:"impact"`
	Probability float64   `json:"probability"`
	Mitigation  string    `json:"mitigation"`
}

type Task struct {
	ID           string        `json:"id"`
	Name         string        `json:"name"`
	Description  string        `json:"description"`
	Duration     time.Duration `json:"duration"`
	Dependencies []string      `json:"dependencies"`
	Assignee     string        `json:"assignee,omitempty"`
	Status       string        `json:"status"`
}

type Milestone struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	DueDate     time.Time `json:"due_date"`
	Criteria    []string  `json:"criteria"`
}

// Additional service types

// ServiceStatus represents the status of the multicloud service
type ServiceStatus struct {
	ServiceName          string                     `json:"service_name"`
	Status               string                     `json:"status"`
	AvailableProviders   int                        `json:"available_providers"`
	DefaultCurrency      string                     `json:"default_currency"`
	OptimizationsEnabled bool                       `json:"optimizations_enabled"`
	MigrationsEnabled    bool                       `json:"migrations_enabled"`
	CacheEnabled         bool                       `json:"cache_enabled"`
	DefaultRiskTolerance string                     `json:"default_risk_tolerance"`
	Providers            map[string]*ProviderStatus `json:"providers"`
}

// ProviderStatus represents the status of a cloud provider
type ProviderStatus struct {
	Name          string    `json:"name"`
	Status        string    `json:"status"`
	LastChecked   time.Time `json:"last_checked"`
	ResourceCount int       `json:"resource_count"`
	ErrorMessage  string    `json:"error_message,omitempty"`
}

// MigrationFeasibility represents the feasibility analysis of a migration
type MigrationFeasibility struct {
	OverallFeasibility FeasibilityLevel   `json:"overall_feasibility"`
	FeasibilityScore   float64            `json:"feasibility_score"`
	Blockers           []MigrationBlocker `json:"blockers"`
	Warnings           []MigrationWarning `json:"warnings"`
	Recommendations    []string           `json:"recommendations"`
	EstimatedEffort    EffortLevel        `json:"estimated_effort"`
	RiskAssessment     RiskLevel          `json:"risk_assessment"`
	ProcessingTimeMs   int64              `json:"processing_time_ms"`
}

// MigrationBlocker represents a migration blocking issue
type MigrationBlocker struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	Severity    string `json:"severity"`
	Resolution  string `json:"resolution,omitempty"`
}

// MigrationWarning represents a migration warning
type MigrationWarning struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	Impact      string `json:"impact"`
}

// InventoryFilter represents filters for inventory queries
type InventoryFilter struct {
	Providers     []string          `json:"providers,omitempty"`
	ResourceTypes []string          `json:"resource_types,omitempty"`
	Regions       []string          `json:"regions,omitempty"`
	Tags          map[string]string `json:"tags,omitempty"`
	MinCost       *float64          `json:"min_cost,omitempty"`
	MaxCost       *float64          `json:"max_cost,omitempty"`
	CreatedAfter  *time.Time        `json:"created_after,omitempty"`
	CreatedBefore *time.Time        `json:"created_before,omitempty"`
}
