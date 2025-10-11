package multicloud

import (
	"context"
	"time"
)

// OptimizationEngine defines interface for multi-cloud optimization analysis
type OptimizationEngine interface {
	// AnalyzeOptimizations identifies cost optimization opportunities across clouds
	AnalyzeOptimizations(ctx context.Context, request *OptimizationRequest) (*OptimizationResult, error)

	// CompareCosts compares costs across multiple cloud providers
	CompareCosts(ctx context.Context, request *CostComparisonRequest) (*CostComparisonResult, error)

	// GetRecommendations returns optimization recommendations
	GetRecommendations(ctx context.Context, request *RecommendationRequest) (*RecommendationResult, error)
}

// MigrationEngine defines interface for cross-cloud migration analysis
type MigrationEngine interface {
	// EstimateMigrationCosts estimates costs for migrating resources between clouds
	EstimateMigrationCosts(ctx context.Context, request *MigrationRequest) (*MigrationEstimate, error)

	// AnalyzeMigrationFeasibility checks if migration is feasible
	AnalyzeMigrationFeasibility(ctx context.Context, request *MigrationRequest) (*FeasibilityAnalysis, error)

	// GenerateMigrationPlan creates a detailed migration plan
	GenerateMigrationPlan(ctx context.Context, request *MigrationRequest) (*MigrationPlan, error)
}

// DiscoveryEngine defines interface for unified resource discovery
type DiscoveryEngine interface {
	// DiscoverResources discovers resources across all configured providers
	DiscoverResources(ctx context.Context, request *DiscoveryRequest) (*DiscoveryResult, error)

	// ScanProvider scans a specific provider for resources
	ScanProvider(ctx context.Context, providerName string, filters *ResourceFilters) (*ProviderScanResult, error)

	// GetUnifiedInventory returns unified inventory across all providers
	GetUnifiedInventory(ctx context.Context) (*UnifiedInventory, error)
}

// OptimizationRequest represents a request for optimization analysis
type OptimizationRequest struct {
	Providers             []string           `json:"providers"`
	StartDate             time.Time          `json:"start_date"`
	EndDate               time.Time          `json:"end_date"`
	OptimizationTypes     []OptimizationType `json:"optimization_types"`
	RiskTolerance         RiskLevel          `json:"risk_tolerance"`
	SavingsThreshold      float64            `json:"savings_threshold"`
	AutoApprovalThreshold float64            `json:"auto_approval_threshold"`
	IncludeRegions        []string           `json:"include_regions,omitempty"`
	ExcludeRegions        []string           `json:"exclude_regions,omitempty"`
	ResourceFilters       *ResourceFilters   `json:"resource_filters,omitempty"`
}

// CostComparisonRequest represents a request for cost comparison
type CostComparisonRequest struct {
	Providers        []string         `json:"providers"`
	StartDate        time.Time        `json:"start_date"`
	EndDate          time.Time        `json:"end_date"`
	Currency         string           `json:"currency"`
	AggregationLevel AggregationLevel `json:"aggregation_level"`
	IncludeServices  []string         `json:"include_services,omitempty"`
	ExcludeServices  []string         `json:"exclude_services,omitempty"`
	NormalizeRegions bool             `json:"normalize_regions"`
}

// MigrationRequest represents a migration analysis request
type MigrationRequest struct {
	SourceProvider      string            `json:"source_provider"`
	TargetProvider      string            `json:"target_provider"`
	Resources           []ResourceSpec    `json:"resources"`
	SourceRegion        string            `json:"source_region"`
	TargetRegion        string            `json:"target_region"`
	MigrationTimeframe  time.Duration     `json:"migration_timeframe"`
	IncludeDataTransfer bool              `json:"include_data_transfer"`
	MigrationStrategy   MigrationStrategy `json:"migration_strategy"`
}

// DiscoveryRequest represents a resource discovery request
type DiscoveryRequest struct {
	Providers       []string         `json:"providers"`
	ResourceTypes   []string         `json:"resource_types,omitempty"`
	Regions         []string         `json:"regions,omitempty"`
	IncludeMetadata bool             `json:"include_metadata"`
	IncludeCosts    bool             `json:"include_costs"`
	Filters         *ResourceFilters `json:"filters,omitempty"`
}

// RecommendationRequest represents a request for recommendations
type RecommendationRequest struct {
	Providers          []string  `json:"providers"`
	RiskTolerance      RiskLevel `json:"risk_tolerance"`
	SavingsThreshold   float64   `json:"savings_threshold"`
	MaxRecommendations int       `json:"max_recommendations"`
	IncludeCategories  []string  `json:"include_categories,omitempty"`
	ExcludeCategories  []string  `json:"exclude_categories,omitempty"`
}

// Enums and types
type OptimizationType string

const (
	OptimizationTypeRightSizing         OptimizationType = "right_sizing"
	OptimizationTypeReservedInstances   OptimizationType = "reserved_instances"
	OptimizationTypeSpotInstances       OptimizationType = "spot_instances"
	OptimizationTypeCostArbitrage       OptimizationType = "cost_arbitrage"
	OptimizationTypeRegionSwitching     OptimizationType = "region_switching"
	OptimizationTypeStorageOptimization OptimizationType = "storage_optimization"
	OptimizationTypeNetworkOptimization OptimizationType = "network_optimization"
)

type RiskLevel string

const (
	RiskLevelLow    RiskLevel = "low"
	RiskLevelMedium RiskLevel = "medium"
	RiskLevelHigh   RiskLevel = "high"
)

type AggregationLevel string

const (
	AggregationLevelHourly  AggregationLevel = "hourly"
	AggregationLevelDaily   AggregationLevel = "daily"
	AggregationLevelWeekly  AggregationLevel = "weekly"
	AggregationLevelMonthly AggregationLevel = "monthly"
)

type MigrationStrategy string

const (
	MigrationStrategyLiftAndShift MigrationStrategy = "lift_and_shift"
	MigrationStrategyReplatform   MigrationStrategy = "replatform"
	MigrationStrategyRefactor     MigrationStrategy = "refactor"
	MigrationStrategyHybrid       MigrationStrategy = "hybrid"
)

// ResourceSpec represents a resource specification
type ResourceSpec struct {
	ID            string                 `json:"id"`
	Type          string                 `json:"type"`
	Name          string                 `json:"name"`
	Region        string                 `json:"region"`
	Configuration map[string]interface{} `json:"configuration"`
	Dependencies  []string               `json:"dependencies,omitempty"`
}

// ResourceFilters represents filters for resource discovery/analysis
type ResourceFilters struct {
	Tags          map[string]string `json:"tags,omitempty"`
	MinCost       float64           `json:"min_cost,omitempty"`
	MaxCost       float64           `json:"max_cost,omitempty"`
	CreatedAfter  *time.Time        `json:"created_after,omitempty"`
	CreatedBefore *time.Time        `json:"created_before,omitempty"`
	State         []string          `json:"state,omitempty"`
}
