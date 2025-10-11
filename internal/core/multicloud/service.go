package multicloud

//nolint:unused // Advanced operations not yet exposed via CLI/API (kept for planned features).
import (
	"context"
	"fmt"
	"sync"
	"time"

	"local/costscope/internal/core/logging"
	"local/costscope/internal/providers"
)

// MulticloudService provides unified multi-cloud management capabilities
type MulticloudService struct {
	providerManager *providers.ProviderManager
	logger          *logging.Logger

	// Component engines
	optimizationEngine OptimizationEngine
	migrationEngine    MigrationEngine
	discoveryEngine    DiscoveryEngine

	// Configuration snapshot (immutable after construction). Dynamic runtime updates
	// were removed to reduce surface area and align with global precedence engine.
	config *MulticloudConfig

	// Thread safety
	mu sync.RWMutex
}

// MulticloudConfig holds configuration for multicloud operations
type MulticloudConfig struct {
	DefaultCurrency      string        `json:"default_currency"`
	DefaultTimeout       time.Duration `json:"default_timeout"`
	MaxConcurrentScans   int           `json:"max_concurrent_scans"`
	CacheEnabled         bool          `json:"cache_enabled"`
	CacheTTL             time.Duration `json:"cache_ttl"`
	EnableOptimizations  bool          `json:"enable_optimizations"`
	EnableMigrations     bool          `json:"enable_migrations"`
	DefaultRiskTolerance RiskLevel     `json:"default_risk_tolerance"`
}

// NewMulticloudService creates a new multicloud service
func NewMulticloudService(providerManager *providers.ProviderManager, logger *logging.Logger) *MulticloudService {
	config := &MulticloudConfig{
		DefaultCurrency:      "USD",
		DefaultTimeout:       30 * time.Minute,
		MaxConcurrentScans:   5,
		CacheEnabled:         true,
		CacheTTL:             1 * time.Hour,
		EnableOptimizations:  true,
		EnableMigrations:     true,
		DefaultRiskTolerance: RiskLevelMedium,
	}

	service := &MulticloudService{
		providerManager: providerManager,
		logger:          logger,
		config:          config,
	}

	// Initialize engines
	service.optimizationEngine = NewBasicOptimizationEngine(providerManager, logger)
	service.migrationEngine = NewBasicMigrationEngine(providerManager, logger)
	service.discoveryEngine = NewBasicDiscoveryEngine(providerManager, logger)

	return service
}

// Optimization operations

// AnalyzeOptimizations identifies cost optimization opportunities across clouds
func (ms *MulticloudService) AnalyzeOptimizations(ctx context.Context, request *OptimizationRequest) (*OptimizationResult, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	ms.logger.Info("Starting multi-cloud optimization analysis")

	if !ms.config.EnableOptimizations {
		return nil, fmt.Errorf("optimization analysis is disabled")
	}

	// Validate providers
	if err := ms.validateProviders(request.Providers); err != nil {
		ms.logger.Error("Optimization analysis failed - provider validation error")
		return nil, fmt.Errorf("provider validation failed: %w", err)
	}

	// Delegate to optimization engine
	result, err := ms.optimizationEngine.AnalyzeOptimizations(ctx, request)
	if err != nil {
		ms.logger.Error("Optimization analysis failed")
		return nil, fmt.Errorf("optimization analysis failed: %w", err)
	}

	ms.logger.Info("Optimization analysis completed successfully")
	return result, nil
}

// CompareCosts compares costs across multiple cloud providers
func (ms *MulticloudService) CompareCosts(ctx context.Context, request *CostComparisonRequest) (*CostComparisonResult, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	ms.logger.Info("Starting cost comparison")

	// Validate providers
	if err := ms.validateProviders(request.Providers); err != nil {
		return nil, fmt.Errorf("provider validation failed: %w", err)
	}

	// Delegate to optimization engine
	result, err := ms.optimizationEngine.CompareCosts(ctx, request)
	if err != nil {
		ms.logger.Error("Cost comparison failed")
		return nil, fmt.Errorf("cost comparison failed: %w", err)
	}

	ms.logger.Info("Cost comparison completed successfully")
	return result, nil
}

// GetRecommendations returns optimization recommendations
func (ms *MulticloudService) GetRecommendations(ctx context.Context, request *RecommendationRequest) (*RecommendationResult, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	ms.logger.Info("Getting optimization recommendations")

	// Validate providers
	if err := ms.validateProviders(request.Providers); err != nil {
		return nil, fmt.Errorf("provider validation failed: %w", err)
	}

	// Delegate to optimization engine
	result, err := ms.optimizationEngine.GetRecommendations(ctx, request)
	if err != nil {
		ms.logger.Error("Getting recommendations failed")
		return nil, fmt.Errorf("getting recommendations failed: %w", err)
	}

	ms.logger.Info("Optimization analysis completed successfully")
	return result, nil
}

// Migration operations

// EstimateMigrationCosts estimates costs for migrating resources between clouds
func (ms *MulticloudService) EstimateMigrationCosts(ctx context.Context, request *MigrationRequest) (*MigrationEstimate, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	ms.logger.Info("Estimating migration costs")

	if !ms.config.EnableMigrations {
		return nil, fmt.Errorf("migration analysis is disabled")
	}

	// Validate providers
	providers := []string{request.SourceProvider, request.TargetProvider}
	if err := ms.validateProviders(providers); err != nil {
		return nil, fmt.Errorf("provider validation failed: %w", err)
	}

	// Delegate to migration engine
	result, err := ms.migrationEngine.EstimateMigrationCosts(ctx, request)
	if err != nil {
		ms.logger.Error("Migration cost estimation failed")
		return nil, fmt.Errorf("migration cost estimation failed: %w", err)
	}

	ms.logger.Info("Migration cost estimation completed")
	return result, nil
}

// AnalyzeMigrationFeasibility analyzes feasibility of migration
func (ms *MulticloudService) AnalyzeMigrationFeasibility(ctx context.Context, request *MigrationRequest) (*MigrationFeasibility, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	ms.logger.Info("Analyzing migration feasibility")

	// Validate providers
	providers := []string{request.SourceProvider, request.TargetProvider}
	if err := ms.validateProviders(providers); err != nil {
		return nil, fmt.Errorf("provider validation failed: %w", err)
	}

	// Delegate to migration engine
	feasibilityResult, err := ms.migrationEngine.AnalyzeMigrationFeasibility(ctx, request)
	if err != nil {
		ms.logger.Error("Migration feasibility analysis failed")
		return nil, fmt.Errorf("migration feasibility analysis failed: %w", err)
	}

	// Convert to expected return type
	result := &MigrationFeasibility{
		OverallFeasibility: FeasibilityMedium, // Default
		FeasibilityScore:   feasibilityResult.FeasibilityScore,
		Blockers:           []MigrationBlocker{},
		Warnings:           []MigrationWarning{},
		Recommendations:    feasibilityResult.Recommendations,
		EstimatedEffort:    EffortMedium,    // Default
		RiskAssessment:     RiskLevelMedium, // Default
		ProcessingTimeMs:   feasibilityResult.ProcessingTimeMs,
	}

	ms.logger.Info("Migration feasibility analysis completed")
	return result, nil
}

// GenerateMigrationPlan generates a detailed migration plan
func (ms *MulticloudService) GenerateMigrationPlan(ctx context.Context, request *MigrationRequest) (*MigrationPlan, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	ms.logger.Info("Generating migration plan")

	// Validate providers
	providers := []string{request.SourceProvider, request.TargetProvider}
	if err := ms.validateProviders(providers); err != nil {
		return nil, fmt.Errorf("provider validation failed: %w", err)
	}

	// Delegate to migration engine
	result, err := ms.migrationEngine.GenerateMigrationPlan(ctx, request)
	if err != nil {
		ms.logger.Error("Migration plan generation failed")
		return nil, fmt.Errorf("migration plan generation failed: %w", err)
	}

	ms.logger.Info("Migration plan generation completed")
	return result, nil
}

// Discovery operations

// DiscoverResources discovers resources across multiple cloud providers
func (ms *MulticloudService) DiscoverResources(ctx context.Context, request *DiscoveryRequest) (*DiscoveryResult, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	ms.logger.Info("Starting resource discovery")

	// Validate providers
	if err := ms.validateProviders(request.Providers); err != nil {
		return nil, fmt.Errorf("provider validation failed: %w", err)
	}

	// Delegate to discovery engine
	result, err := ms.discoveryEngine.DiscoverResources(ctx, request)
	if err != nil {
		ms.logger.Error("Resource discovery failed")
		return nil, fmt.Errorf("resource discovery failed: %w", err)
	}

	ms.logger.Info("Resource discovery completed")
	return result, nil
}

// GetUnifiedInventory gets unified inventory across providers
func (ms *MulticloudService) GetUnifiedInventory(ctx context.Context, filter *InventoryFilter) (*UnifiedInventory, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	// Delegate to discovery engine
	result, err := ms.discoveryEngine.GetUnifiedInventory(ctx)
	if err != nil {
		ms.logger.Error("Getting unified inventory failed")
		return nil, fmt.Errorf("getting unified inventory failed: %w", err)
	}

	ms.logger.Info("Unified inventory retrieved")
	return result, nil
}

// NOTE: Former dynamic configuration methods (UpdateConfig / GetConfig) and
// the ScanProvider wrapper were removed as dead code. If hot-reload of
// multicloud settings becomes a real requirement, introduce a centralized
// runtime config manager instead of per-service setters.

// GetServiceStatus returns the current service status
func (ms *MulticloudService) GetServiceStatus() *ServiceStatus {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	// Count available providers
	availableProviders := 0
	if ms.providerManager != nil {
		// This would normally call a method to count configured providers
		// For now, we'll use a placeholder count of 0
		availableProviders = 0
	}

	return &ServiceStatus{
		ServiceName:          "multicloud",
		Status:               "running",
		AvailableProviders:   availableProviders,
		DefaultCurrency:      ms.config.DefaultCurrency,
		OptimizationsEnabled: ms.config.EnableOptimizations,
		MigrationsEnabled:    ms.config.EnableMigrations,
		CacheEnabled:         ms.config.CacheEnabled,
		DefaultRiskTolerance: string(ms.config.DefaultRiskTolerance),
		Providers:            make(map[string]*ProviderStatus),
	}
}

// Helper methods

// validateProviders validates that all requested providers are available
func (ms *MulticloudService) validateProviders(providers []string) error {
	if len(providers) == 0 {
		return fmt.Errorf("no providers specified")
	}

	// For now, we'll just check that providers are known types
	validProviders := map[string]bool{
		"aws":   true,
		"azure": true,
		"gcp":   true,
	}

	for _, provider := range providers {
		if !validProviders[provider] {
			return fmt.Errorf("provider %s is not configured", provider)
		}
	}

	return nil
}
