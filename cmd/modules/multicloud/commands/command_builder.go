package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/spf13/cobra"

	"local/costscope/cmd/modules/multicloud/migration"
	"local/costscope/cmd/modules/multicloud/optimization"
	"local/costscope/cmd/modules/multicloud/reporting"
	"local/costscope/cmd/modules/multicloud/validation"
	"local/costscope/internal/core/logging"
	"local/costscope/internal/core/multicloud"
	"local/costscope/internal/providers"
)

// MulticloudCommands manages all multicloud-related commands
type MulticloudCommands struct {
	service *multicloud.MulticloudService
	logger  *logging.Logger

	// Command flags
	flags *CommandFlags
}

// CommandFlags holds all flag values for multicloud commands
type CommandFlags struct {
	// Global flags
	Providers    []string
	StartDate    string
	EndDate      string
	OutputFormat string
	OutputFile   string
	ConfigFile   string

	// Optimization specific flags
	OptimizationTypes     []string
	RiskTolerance         string
	SavingsThreshold      float64
	AutoApprovalThreshold float64
	MaxRecommendations    int

	// Migration specific flags
	SourceProvider   string
	TargetProvider   string
	ResourceSpecFile string

	// Discovery specific flags
	ResourceTypes   []string
	Regions         []string
	IncludeMetadata bool
	IncludeCosts    bool

	// Report specific flags
	IncludeProviderBreakdown bool
	IncludeOptimizations     bool
	CurrencyNormalization    string
	AggregationLevel         string
}

// NewMulticloudCommands creates a new instance of MulticloudCommands
func NewMulticloudCommands(providerManager *providers.ProviderManager) *MulticloudCommands {
	logger := logging.NewLogger(logging.LevelInfo)
	service := multicloud.NewMulticloudService(providerManager, logger)

	return &MulticloudCommands{
		service: service,
		logger:  logger,
		flags:   &CommandFlags{},
	}
}

// Note: BuildMulticloudCommand is generated (see zz_generated_command_builder.go)
// The enhanced subcommands are attached from root wiring to avoid duplicate builders.

// AttachEnhancedSubcommands attaches manually implemented enhanced commands
// to the provided multicloud root command built by the generated builder.
func (mc *MulticloudCommands) AttachEnhancedSubcommands(root *cobra.Command) {
	if root == nil {
		return
	}
	root.AddCommand(mc.buildAdvancedMigrateCommand())
	root.AddCommand(mc.buildEnhancedOptimizeCommand())
	root.AddCommand(mc.buildEnhancedValidateCommand())
	root.AddCommand(mc.buildEnhancedReportCommand())
	root.AddCommand(mc.buildRecommendationsCommand())
	root.AddCommand(mc.buildInventoryCommand())
	root.AddCommand(mc.buildMigrationPlanCommand())
	root.AddCommand(mc.buildFeasibilityCommand())
}

// buildOptimizeCommand creates the optimize subcommand
// Generated command provides optimize

// buildCompareCommand creates the compare subcommand
// Generated command provides compare

// buildMigrateCommand creates the migrate subcommand
// Generated command provides migrate

// buildDiscoverCommand creates the discover subcommand
// Generated command provides discover

// buildValidateCommand creates the validate subcommand
// Generated command provides validate

// addGlobalFlags adds global flags to the multicloud command
// Global flags are generated via spec and bound to mc.flags

// Command execution methods

// runOptimize executes the optimize command
func (mc *MulticloudCommands) runOptimize(cmd *cobra.Command, args []string) error {
	mc.logger.Info("Running multi-cloud optimization analysis")

	// Parse dates
	startDate, endDate, err := mc.parseDateRange(mc.flags.StartDate, mc.flags.EndDate)
	if err != nil {
		return fmt.Errorf("invalid date range: %w", err)
	}

	// Parse optimization types
	optimizationTypes := mc.parseOptimizationTypes(mc.flags.OptimizationTypes)

	// Parse risk tolerance
	riskTolerance := mc.parseRiskTolerance(mc.flags.RiskTolerance)

	// Create optimization request
	request := &multicloud.OptimizationRequest{
		Providers:             mc.getProviders(),
		StartDate:             startDate,
		EndDate:               endDate,
		OptimizationTypes:     optimizationTypes,
		RiskTolerance:         riskTolerance,
		SavingsThreshold:      mc.flags.SavingsThreshold,
		AutoApprovalThreshold: mc.flags.AutoApprovalThreshold,
	}

	// Execute optimization analysis
	result, err := mc.service.AnalyzeOptimizations(context.Background(), request)
	if err != nil {
		return fmt.Errorf("optimization analysis failed: %w", err)
	}

	// Output results
	return mc.outputResult(result, "optimization_analysis")
}

// runCompare executes the compare command
func (mc *MulticloudCommands) runCompare(cmd *cobra.Command, args []string) error {
	mc.logger.Info("Running multi-cloud cost comparison")

	// Parse dates
	startDate, endDate, err := mc.parseDateRange(mc.flags.StartDate, mc.flags.EndDate)
	if err != nil {
		return fmt.Errorf("invalid date range: %w", err)
	}

	// Parse aggregation level
	aggregationLevel := mc.parseAggregationLevel(mc.flags.AggregationLevel)

	// Create comparison request
	request := &multicloud.CostComparisonRequest{
		Providers:        mc.getProviders(),
		StartDate:        startDate,
		EndDate:          endDate,
		Currency:         mc.flags.CurrencyNormalization,
		AggregationLevel: aggregationLevel,
		NormalizeRegions: true,
	}

	// Execute cost comparison
	result, err := mc.service.CompareCosts(context.Background(), request)
	if err != nil {
		return fmt.Errorf("cost comparison failed: %w", err)
	}

	// Output results
	return mc.outputResult(result, "cost_comparison")
}

// runMigrate executes the migrate command
func (mc *MulticloudCommands) runMigrate(cmd *cobra.Command, args []string) error {
	mc.logger.Info("Running migration analysis")

	// Create migration request
	request := &multicloud.MigrationRequest{
		SourceProvider:      mc.flags.SourceProvider,
		TargetProvider:      mc.flags.TargetProvider,
		Resources:           []multicloud.ResourceSpec{}, // TODO: Load from file
		SourceRegion:        "us-east-1",                 // Default for now
		TargetRegion:        "us-east-1",                 // Default for now
		MigrationTimeframe:  30 * 24 * time.Hour,         // 30 days default
		IncludeDataTransfer: true,
		MigrationStrategy:   multicloud.MigrationStrategyLiftAndShift,
	}

	// Execute migration estimation
	estimate, err := mc.service.EstimateMigrationCosts(context.Background(), request)
	if err != nil {
		return fmt.Errorf("migration estimation failed: %w", err)
	}

	// Also get feasibility analysis
	feasibility, err := mc.service.AnalyzeMigrationFeasibility(context.Background(), request)
	if err != nil {
		return fmt.Errorf("migration feasibility analysis failed: %w", err)
	}

	// Combine results
	migrationResult := map[string]interface{}{
		"estimate":    estimate,
		"feasibility": feasibility,
	}

	// Output results
	return mc.outputResult(migrationResult, "migration_analysis")
}

// runDiscover executes the discover command
func (mc *MulticloudCommands) runDiscover(cmd *cobra.Command, args []string) error {
	mc.logger.Info("Running resource discovery")

	// Create discovery request
	request := &multicloud.DiscoveryRequest{
		Providers:       mc.getProviders(),
		ResourceTypes:   mc.flags.ResourceTypes,
		Regions:         mc.flags.Regions,
		IncludeMetadata: mc.flags.IncludeMetadata,
		IncludeCosts:    mc.flags.IncludeCosts,
	}

	// Execute resource discovery
	result, err := mc.service.DiscoverResources(context.Background(), request)
	if err != nil {
		return fmt.Errorf("resource discovery failed: %w", err)
	}

	// Output results
	return mc.outputResult(result, "resource_discovery")
}

// runRecommendations executes the recommendations command (advanced optimization surface)
func (mc *MulticloudCommands) runRecommendations(cmd *cobra.Command, args []string) error {
	mc.logger.Info("Running recommendations generation")

	// Build request
	req := &multicloud.RecommendationRequest{
		Providers:          mc.getProviders(),
		RiskTolerance:      mc.parseRiskTolerance(mc.flags.RiskTolerance),
		SavingsThreshold:   mc.flags.SavingsThreshold,
		MaxRecommendations: mc.flags.MaxRecommendations,
	}

	res, err := mc.service.GetRecommendations(context.Background(), req)
	if err != nil {
		return fmt.Errorf("recommendations failed: %w", err)
	}
	return mc.outputResult(res, "recommendations")
}

// runInventory executes the unified inventory command
func (mc *MulticloudCommands) runInventory(cmd *cobra.Command, args []string) error {
	mc.logger.Info("Fetching unified inventory")
	inv, err := mc.service.GetUnifiedInventory(context.Background(), nil)
	if err != nil {
		return fmt.Errorf("inventory fetch failed: %w", err)
	}
	return mc.outputResult(inv, "unified_inventory")
}

// runMigrationPlan executes the migration-plan command
func (mc *MulticloudCommands) runMigrationPlan(cmd *cobra.Command, args []string) error {
	mc.logger.Info("Generating migration plan")
	req := &multicloud.MigrationRequest{
		SourceProvider:      mc.flags.SourceProvider,
		TargetProvider:      mc.flags.TargetProvider,
		Resources:           []multicloud.ResourceSpec{},
		SourceRegion:        "us-east-1",
		TargetRegion:        "us-east-1",
		MigrationTimeframe:  30 * 24 * time.Hour,
		MigrationStrategy:   multicloud.MigrationStrategyLiftAndShift,
		IncludeDataTransfer: true,
	}
	plan, err := mc.service.GenerateMigrationPlan(context.Background(), req)
	if err != nil {
		return fmt.Errorf("migration plan failed: %w", err)
	}
	return mc.outputResult(plan, "migration_plan")
}

// runFeasibility executes migration feasibility analysis
func (mc *MulticloudCommands) runFeasibility(cmd *cobra.Command, args []string) error {
	mc.logger.Info("Running migration feasibility analysis")
	req := &multicloud.MigrationRequest{
		SourceProvider:      mc.flags.SourceProvider,
		TargetProvider:      mc.flags.TargetProvider,
		Resources:           []multicloud.ResourceSpec{},
		SourceRegion:        "us-east-1",
		TargetRegion:        "us-east-1",
		MigrationTimeframe:  30 * 24 * time.Hour,
		MigrationStrategy:   multicloud.MigrationStrategyLiftAndShift,
		IncludeDataTransfer: true,
	}
	feas, err := mc.service.AnalyzeMigrationFeasibility(context.Background(), req)
	if err != nil {
		return fmt.Errorf("feasibility failed: %w", err)
	}
	return mc.outputResult(feas, "migration_feasibility")
}

// runValidate executes the validate command
func (mc *MulticloudCommands) runValidate(cmd *cobra.Command, args []string) error {
	mc.logger.Info("Validating multi-cloud providers")

	// Get service status
	status := mc.service.GetServiceStatus()

	// Output status
	return mc.outputResult(status, "validation_status")
}

// Helper methods

// getProviders returns the list of providers to use
func (mc *MulticloudCommands) getProviders() []string {
	if len(mc.flags.Providers) > 0 {
		return mc.flags.Providers
	}

	// Default to all available providers
	// TODO: Get from provider manager
	return []string{"aws", "azure", "gcp"}
}

// parseDateRange parses start and end dates
func (mc *MulticloudCommands) parseDateRange(startDate, endDate string) (time.Time, time.Time, error) {
	var start, end time.Time
	var err error

	if startDate == "" {
		start = time.Now().AddDate(0, -1, 0) // Default to 1 month ago
	} else {
		start, err = time.Parse("2006-01-02", startDate)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("invalid start date format: %w", err)
		}
	}

	if endDate == "" {
		end = time.Now() // Default to now
	} else {
		end, err = time.Parse("2006-01-02", endDate)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("invalid end date format: %w", err)
		}
	}

	if start.After(end) {
		return time.Time{}, time.Time{}, fmt.Errorf("start date must be before end date")
	}

	return start, end, nil
}

// parseOptimizationTypes parses optimization types from strings
func (mc *MulticloudCommands) parseOptimizationTypes(types []string) []multicloud.OptimizationType {
	if len(types) == 0 {
		// Default to all types
		return []multicloud.OptimizationType{
			multicloud.OptimizationTypeRightSizing,
			multicloud.OptimizationTypeReservedInstances,
			multicloud.OptimizationTypeSpotInstances,
			multicloud.OptimizationTypeCostArbitrage,
			multicloud.OptimizationTypeRegionSwitching,
		}
	}

	optimizationTypes := make([]multicloud.OptimizationType, 0, len(types))
	for _, t := range types {
		switch strings.ToLower(t) {
		case "right_sizing":
			optimizationTypes = append(optimizationTypes, multicloud.OptimizationTypeRightSizing)
		case "reserved_instances":
			optimizationTypes = append(optimizationTypes, multicloud.OptimizationTypeReservedInstances)
		case "spot_instances":
			optimizationTypes = append(optimizationTypes, multicloud.OptimizationTypeSpotInstances)
		case "cost_arbitrage":
			optimizationTypes = append(optimizationTypes, multicloud.OptimizationTypeCostArbitrage)
		case "region_switching":
			optimizationTypes = append(optimizationTypes, multicloud.OptimizationTypeRegionSwitching)
		}
	}

	return optimizationTypes
}

// parseRiskTolerance parses risk tolerance from string
func (mc *MulticloudCommands) parseRiskTolerance(riskStr string) multicloud.RiskLevel {
	switch strings.ToLower(riskStr) {
	case "low":
		return multicloud.RiskLevelLow
	case "high":
		return multicloud.RiskLevelHigh
	default:
		return multicloud.RiskLevelMedium
	}
}

// parseAggregationLevel parses aggregation level from string
func (mc *MulticloudCommands) parseAggregationLevel(aggStr string) multicloud.AggregationLevel {
	switch strings.ToLower(aggStr) {
	case "hourly":
		return multicloud.AggregationLevelHourly
	case "weekly":
		return multicloud.AggregationLevelWeekly
	case "monthly":
		return multicloud.AggregationLevelMonthly
	default:
		return multicloud.AggregationLevelDaily
	}
}

// outputResult outputs the result in the specified format
func (mc *MulticloudCommands) outputResult(result interface{}, prefix string) error {
	var output []byte
	var err error

	switch mc.flags.OutputFormat {
	case "json":
		output, err = json.MarshalIndent(result, "", "  ")
	case "yaml":
		output, err = yaml.Marshal(result)
	default:
		output, err = json.MarshalIndent(result, "", "  ")
	}

	if err != nil {
		return fmt.Errorf("failed to format output: %w", err)
	}

	// Write to file or stdout
	if mc.flags.OutputFile != "" {
		// Use prefix to make filename more descriptive if not already specified
		filename := mc.flags.OutputFile
		if filename == "output.json" && prefix != "" {
			filename = fmt.Sprintf("%s_%s.json", prefix, time.Now().Format("20060102_150405"))
		}
		err = os.WriteFile(filename, output, 0600)
		if err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		fmt.Printf("Results written to %s\n", filename)
	} else {
		// Add prefix to output for better context
		if prefix != "" {
			fmt.Printf("=== %s ===\n", strings.ToUpper(strings.ReplaceAll(prefix, "_", " ")))
		}
		fmt.Println(string(output))
	}

	return nil
}

// Enhanced command builders for advanced features

// buildAdvancedMigrateCommand creates the advanced migrate command
func (mc *MulticloudCommands) buildAdvancedMigrateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate-advanced",
		Short: "Advanced cross-cloud migration analysis with comprehensive assessment",
		Long: `Advanced migration analysis with detailed cost estimates, risk assessment,
timeline planning, and comprehensive feasibility studies across cloud providers.

This command provides enterprise-grade migration analysis including:
- Comprehensive migration cost analysis
- Risk assessment and mitigation strategies
- Resource compatibility mapping
- Timeline and execution planning
- Performance impact assessment`,
		Example: `  # Advanced migration analysis from AWS to Azure
  costscope multicloud migrate-advanced --source aws --target azure --risk-tolerance medium

  # Comprehensive analysis with detailed reporting
  costscope multicloud migrate-advanced --source aws --target gcp --include-all --output migration-report.json`,
		RunE: mc.runAdvancedMigrate,
	}

	// Advanced migration flags
	cmd.Flags().StringVar(&mc.flags.SourceProvider, "source", "", "Source cloud provider (required)")
	cmd.Flags().StringVar(&mc.flags.TargetProvider, "target", "", "Target cloud provider (required)")
	cmd.Flags().StringVar(&mc.flags.RiskTolerance, "risk-tolerance", "medium", "Risk tolerance (low,medium,high)")
	cmd.Flags().BoolVar(&mc.flags.IncludeMetadata, "include-all", false, "Include all comprehensive analysis")
	cmd.Flags().StringVar(&mc.flags.ResourceSpecFile, "resources", "", "Resource specification file")

	// Mark required flags
	if err := cmd.MarkFlagRequired("source"); err != nil {
		mc.logger.Error(fmt.Sprintf("Failed to mark source flag as required: %v", err))
	}
	if err := cmd.MarkFlagRequired("target"); err != nil {
		mc.logger.Error(fmt.Sprintf("Failed to mark target flag as required: %v", err))
	}

	return cmd
}

// buildEnhancedOptimizeCommand creates the enhanced optimize command
func (mc *MulticloudCommands) buildEnhancedOptimizeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "optimize-enhanced",
		Short: "Enhanced optimization analysis with ML predictions and advanced algorithms",
		Long: `Enhanced cross-cloud optimization analysis using machine learning predictions,
advanced algorithms, and comprehensive cost optimization strategies.

Features:
- ML-powered cost predictions
- Advanced optimization algorithms
- Simulation and scenario modeling
- Automated recommendation generation
- Risk-adjusted optimization strategies`,
		Example: `  # Enhanced optimization with ML predictions
  costscope multicloud optimize-enhanced --providers aws,azure --include-ml --simulation-mode

  # Advanced optimization with custom thresholds
  costscope multicloud optimize-enhanced --providers aws,gcp --savings-threshold 1000 --auto-approval 500`,
		RunE: mc.runEnhancedOptimize,
	}

	// Enhanced optimization flags
	cmd.Flags().StringSliceVar(&mc.flags.Providers, "providers", []string{}, "Cloud providers (aws,azure,gcp)")
	cmd.Flags().StringSliceVar(&mc.flags.OptimizationTypes, "types", []string{}, "Optimization types")
	cmd.Flags().StringVar(&mc.flags.RiskTolerance, "risk-tolerance", "medium", "Risk tolerance level")
	cmd.Flags().Float64Var(&mc.flags.SavingsThreshold, "savings-threshold", 100.0, "Minimum savings threshold")
	cmd.Flags().Float64Var(&mc.flags.AutoApprovalThreshold, "auto-approval", 1000.0, "Auto-approval threshold")
	cmd.Flags().BoolVar(&mc.flags.IncludeMetadata, "include-ml", false, "Include ML predictions")
	cmd.Flags().BoolVar(&mc.flags.IncludeCosts, "simulation-mode", false, "Enable simulation mode")

	return cmd
}

// buildEnhancedValidateCommand creates the enhanced validate command
func (mc *MulticloudCommands) buildEnhancedValidateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate-enhanced",
		Short: "Enhanced multi-cloud validation with security, compliance, and auto-fix",
		Long: `Enhanced multi-cloud provider validation including connectivity, permissions,
security assessments, compliance checking, and automated issue resolution.

Comprehensive validation features:
- Provider connectivity and authentication
- Permission and access validation
- Security posture assessment
- Compliance framework checking
- Performance and cost validation
- Automated issue remediation`,
		Example: `  # Enhanced validation with all checks
  costscope multicloud validate-enhanced --all-checks --detailed

  # Security-focused validation with auto-fix
  costscope multicloud validate-enhanced --security --compliance --fix-issues`,
		RunE: mc.runEnhancedValidate,
	}

	// Enhanced validation flags
	cmd.Flags().StringSliceVar(&mc.flags.Providers, "providers", []string{}, "Cloud providers to validate")
	cmd.Flags().StringVar(&mc.flags.ConfigFile, "config", "", "Configuration file")
	cmd.Flags().BoolVar(&mc.flags.IncludeMetadata, "all-checks", false, "Include all validation checks")
	cmd.Flags().BoolVar(&mc.flags.IncludeCosts, "detailed", false, "Detailed validation output")
	cmd.Flags().BoolVar(&mc.flags.IncludeProviderBreakdown, "security", false, "Include security validation")
	cmd.Flags().BoolVar(&mc.flags.IncludeOptimizations, "compliance", false, "Include compliance validation")
	cmd.Flags().StringVar(&mc.flags.ResourceSpecFile, "fix-issues", "", "Auto-fix identified issues")

	return cmd
}

// buildEnhancedReportCommand creates the enhanced report command
func (mc *MulticloudCommands) buildEnhancedReportCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "report-enhanced",
		Short: "Enhanced multi-cloud reporting with executive summaries and forecasting",
		Long: `Generate comprehensive multi-cloud reports with executive summaries,
cost forecasting, trend analysis, and custom dashboard generation.

Advanced reporting features:
- Executive summary generation
- Cost forecasting and trend analysis
- Compliance reporting
- Custom dashboard creation
- Automated report scheduling
- Multi-format output support`,
		Example: `  # Generate comprehensive executive report
  costscope multicloud report-enhanced --type executive --providers aws,azure --forecasting

  # Detailed compliance report with charts
  costscope multicloud report-enhanced --type compliance --providers aws,gcp --charts --output report.json`,
		RunE: mc.runEnhancedReport,
	}

	// Enhanced reporting flags
	cmd.Flags().StringSliceVar(&mc.flags.Providers, "providers", []string{}, "Cloud providers")
	cmd.Flags().StringVar(&mc.flags.StartDate, "start-date", "", "Start date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&mc.flags.EndDate, "end-date", "", "End date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&mc.flags.ResourceSpecFile, "type", "comprehensive", "Report type (comprehensive,executive,compliance)")
	cmd.Flags().StringVar(&mc.flags.CurrencyNormalization, "currency", "USD", "Currency for normalization")
	cmd.Flags().StringVar(&mc.flags.AggregationLevel, "aggregation", "daily", "Aggregation level")
	cmd.Flags().BoolVar(&mc.flags.IncludeProviderBreakdown, "provider-breakdown", true, "Include provider breakdown")
	cmd.Flags().BoolVar(&mc.flags.IncludeOptimizations, "optimizations", false, "Include optimization recommendations")
	cmd.Flags().BoolVar(&mc.flags.IncludeMetadata, "forecasting", false, "Include cost forecasting")
	cmd.Flags().BoolVar(&mc.flags.IncludeCosts, "charts", false, "Include charts and visualizations")

	return cmd
}

// buildRecommendationsCommand wires the recommendations surface of MulticloudService
func (mc *MulticloudCommands) buildRecommendationsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "recommendations",
		Short: "Generate multi-cloud cost optimization recommendations",
		Long:  "Generate multi-cloud cost optimization recommendations across providers (currently mock/demo data).",
		RunE:  mc.runRecommendations,
	}
	cmd.Flags().StringSliceVar(&mc.flags.Providers, "providers", []string{}, "Cloud providers (aws,azure,gcp)")
	cmd.Flags().StringVar(&mc.flags.RiskTolerance, "risk-tolerance", "medium", "Risk tolerance level")
	cmd.Flags().Float64Var(&mc.flags.SavingsThreshold, "savings-threshold", 0, "Minimum savings filter")
	cmd.Flags().IntVar(&mc.flags.MaxRecommendations, "max", 0, "Maximum recommendations to return (0 = default)")
	return cmd
}

// buildInventoryCommand wires unified inventory retrieval
func (mc *MulticloudCommands) buildInventoryCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "inventory",
		Short: "Get unified multi-cloud resource inventory",
		Long:  "Retrieve a unified inventory summary across registered providers (currently mock/demo data).",
		RunE:  mc.runInventory,
	}
	return cmd
}

// buildMigrationPlanCommand wires migration plan generation
func (mc *MulticloudCommands) buildMigrationPlanCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migration-plan",
		Short: "Generate a detailed cross-cloud migration plan",
		Long:  "Generate a detailed cross-cloud migration plan (currently mock/demo data).",
		RunE:  mc.runMigrationPlan,
	}
	cmd.Flags().StringVar(&mc.flags.SourceProvider, "source", "", "Source cloud provider (required)")
	cmd.Flags().StringVar(&mc.flags.TargetProvider, "target", "", "Target cloud provider (required)")
	if err := cmd.MarkFlagRequired("source"); err != nil {
		mc.logger.Error(err.Error())
	}
	if err := cmd.MarkFlagRequired("target"); err != nil {
		mc.logger.Error(err.Error())
	}
	return cmd
}

// buildFeasibilityCommand wires migration feasibility analysis
func (mc *MulticloudCommands) buildFeasibilityCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "feasibility",
		Short: "Analyze migration feasibility between providers",
		Long:  "Analyze migration feasibility between providers (currently mock/demo data).",
		RunE:  mc.runFeasibility,
	}
	cmd.Flags().StringVar(&mc.flags.SourceProvider, "source", "", "Source cloud provider (required)")
	cmd.Flags().StringVar(&mc.flags.TargetProvider, "target", "", "Target cloud provider (required)")
	if err := cmd.MarkFlagRequired("source"); err != nil {
		mc.logger.Error(err.Error())
	}
	if err := cmd.MarkFlagRequired("target"); err != nil {
		mc.logger.Error(err.Error())
	}
	return cmd
}

// Enhanced command execution methods

// runAdvancedMigrate executes the advanced migrate command
func (mc *MulticloudCommands) runAdvancedMigrate(cmd *cobra.Command, args []string) error {
	mc.logger.Info("Running advanced migration analysis")

	// Create migration runner
	migrationRunner := migration.NewMigrationRunner()

	// Create migration flags
	migrationFlags := &migration.CommandFlags{
		SourceProvider:           mc.flags.SourceProvider,
		TargetProvider:           mc.flags.TargetProvider,
		ResourceSpecFile:         mc.flags.ResourceSpecFile,
		RiskTolerance:            mc.flags.RiskTolerance,
		IncludeDataTransfer:      true,
		IncludePerformanceImpact: mc.flags.IncludeMetadata,
		ComprehensiveAnalysis:    mc.flags.IncludeMetadata,
		OutputFormat:             mc.flags.OutputFormat,
		OutputFile:               mc.flags.OutputFile,
		DetailLevel:              "detailed",
	}

	// Execute advanced migration analysis
	return migrationRunner.RunAdvancedMigrationAnalysis(migrationFlags)
}

// runEnhancedOptimize executes the enhanced optimize command
func (mc *MulticloudCommands) runEnhancedOptimize(cmd *cobra.Command, args []string) error {
	mc.logger.Info("Running enhanced optimization analysis")

	// Create optimization runner
	optimizationRunner := optimization.NewOptimizationRunner()

	// Create optimization flags
	optimizationFlags := &optimization.CommandFlags{
		Providers:             mc.flags.Providers,
		StartDate:             mc.flags.StartDate,
		EndDate:               mc.flags.EndDate,
		OptimizationTypes:     mc.flags.OptimizationTypes,
		RiskTolerance:         mc.flags.RiskTolerance,
		SavingsThreshold:      mc.flags.SavingsThreshold,
		AutoApprovalThreshold: mc.flags.AutoApprovalThreshold,
		IncludeMLPredictions:  mc.flags.IncludeMetadata,
		SimulationMode:        mc.flags.IncludeCosts,
		EnableAdvancedRules:   true,
		OutputFormat:          mc.flags.OutputFormat,
		OutputFile:            mc.flags.OutputFile,
		AggregationLevel:      "detailed",
	}

	// Execute enhanced optimization analysis
	return optimizationRunner.RunEnhancedOptimizationAnalysis(optimizationFlags)
}

// runEnhancedValidate executes the enhanced validate command
func (mc *MulticloudCommands) runEnhancedValidate(cmd *cobra.Command, args []string) error {
	mc.logger.Info("Running enhanced provider validation")

	// Create validation runner
	validationRunner := validation.NewValidationRunner()

	// Create validation flags
	validationFlags := &validation.CommandFlags{
		ConfigFile:            mc.flags.ConfigFile,
		ValidationLevel:       "comprehensive",
		IncludeConnectivity:   true,
		IncludePermissions:    true,
		IncludeCompliance:     mc.flags.IncludeOptimizations,
		IncludeSecurity:       mc.flags.IncludeProviderBreakdown,
		IncludePerformance:    mc.flags.IncludeMetadata,
		IncludeCostValidation: true,
		OutputFormat:          mc.flags.OutputFormat,
		OutputFile:            mc.flags.OutputFile,
		Detailed:              mc.flags.IncludeCosts,
		FixIssues:             mc.flags.ResourceSpecFile != "",
		DryRun:                false,
	}

	// Execute enhanced validation
	return validationRunner.RunEnhancedProviderValidation(validationFlags)
}

// runEnhancedReport executes the enhanced report command
func (mc *MulticloudCommands) runEnhancedReport(cmd *cobra.Command, args []string) error {
	mc.logger.Info("Running enhanced report generation")

	// Create reporting runner
	reportingRunner := reporting.NewReportingRunner()

	// Create reporting flags
	reportingFlags := &reporting.CommandFlags{
		Providers:                mc.flags.Providers,
		StartDate:                mc.flags.StartDate,
		EndDate:                  mc.flags.EndDate,
		ReportType:               mc.flags.ResourceSpecFile,
		IncludeProviderBreakdown: mc.flags.IncludeProviderBreakdown,
		IncludeOptimizations:     mc.flags.IncludeOptimizations,
		CurrencyNormalization:    mc.flags.CurrencyNormalization,
		AggregationLevel:         mc.flags.AggregationLevel,
		IncludeForecasting:       mc.flags.IncludeMetadata,
		IncludeCharts:            mc.flags.IncludeCosts,
		ExecutiveSummary:         true,
		DetailLevel:              "comprehensive",
		OutputFormat:             mc.flags.OutputFormat,
		OutputFile:               mc.flags.OutputFile,
	}

	// Execute enhanced reporting
	return reportingRunner.RunEnhancedReportGeneration(reportingFlags)
}
