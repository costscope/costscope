package cmd

import (
	"context"

	analysisCommands "local/costscope/cmd/modules/analysis/commands"
	analyticsCommands "local/costscope/cmd/modules/analytics/commands"
	apiCommands "local/costscope/cmd/modules/api"
	complianceCommands "local/costscope/cmd/modules/compliance/commands"
	diagnosticsCommands "local/costscope/cmd/modules/diagnostics/commands"
	driftCommands "local/costscope/cmd/modules/drift/commands"
	focusCommands "local/costscope/cmd/modules/focus/commands"
	validationCommands "local/costscope/cmd/modules/focus/commands/validation"
	integrationCommands "local/costscope/cmd/modules/integration"
	invariantsCommands "local/costscope/cmd/modules/invariants/commands"
	monitoringCommands "local/costscope/cmd/modules/monitoring/commands"
	multicloudCommands "local/costscope/cmd/modules/multicloud/commands"
	productionCommands "local/costscope/cmd/modules/production/commands"
	providerCommands "local/costscope/cmd/modules/providers/commands"
	reportCommands "local/costscope/cmd/modules/reports/commands"
	securityCommands "local/costscope/cmd/modules/security/commands"
	"local/costscope/internal/cli"
	"local/costscope/internal/core/analytics"
	"local/costscope/internal/core/integration"
	"local/costscope/internal/core/logging"
	"local/costscope/internal/core/normalization"
	"local/costscope/internal/core/production"
	"local/costscope/internal/core/reports"
	fw "local/costscope/internal/framework"
	"local/costscope/internal/providers"

	"github.com/spf13/cobra"
)

var (
	rootCmd *cobra.Command
)

// Reachability guards (no runtime effect): ensure certain constructor shims are
// referenced in non-test code so static deadcode analysis does not flag them.
//   - validationCommands.BuildValidateCommand is a thin wrapper delegating to the
//     parent package for import/backward-compatibility.
//   - diagnosticsCommands.NewDiagnosticsCommandsWithService is used by tests and
//     alternative wiring; referencing here keeps it reachable for analyzers.
var (
	_ = validationCommands.BuildValidateCommand
	_ = diagnosticsCommands.NewDiagnosticsCommandsWithService
)

func init() {
	// Initialize framework first
	framework := fw.NewFramework()

	// Start framework
	ctx := context.Background()
	if err := framework.Start(ctx); err != nil {
		logging.GetLogger().WarnWithFields("Failed to start framework", map[string]interface{}{"error": err.Error()})
	}

	// Warm normalization caches (regions/units) early to ensure predictable cache hit metrics
	normalization.PreWarm()

	// Create enhanced CLI
	enhancedCLI := cli.NewEnhancedCLI(framework)
	rootCmd = enhancedCLI.GetRootCommand()

	// Discover and register framework commands
	if err := enhancedCLI.DiscoverAndRegisterCommands(); err != nil {
		logging.GetLogger().WarnWithFields("Failed to discover framework commands", map[string]interface{}{"error": err.Error()})
	}

	// Update root command details
	rootCmd.Use = "costscope"
	rootCmd.Short = "CostScope - Enterprise Cloud Cost Analysis Tool"
	rootCmd.Long = `CostScope is a comprehensive enterprise-grade cloud cost management and analysis tool
that helps you monitor, analyze, and optimize your cloud spending across
multiple cloud providers including AWS, Azure, and Google Cloud Platform.

Features include:
- Advanced analytics and reporting
- Multi-cloud support
- FOCUS standard compliance
- Real-time monitoring
- Extensible plugin architecture
- Production-ready deployment options`

	// Initialize existing components
	setupExistingCommands()
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	return rootCmd.Execute()
}

// GetRootCommand returns the initialized root cobra command.
// NOTE: This is intentionally exported solely for black‑box CLI integration tests
// (tests in external packages import cmd and build command trees). Production code
// uses Execute(). Marked in deadcode allowlist; do not remove without migrating
// tests to an in‑package helper or tagged test constructor.
func GetRootCommand() *cobra.Command { return rootCmd }

// setupExistingCommands configures all existing commands
func setupExistingCommands() {
	// Initialize logger
	logger := logging.NewLogger("info")

	// Initialize provider manager
	providerManager := &providers.ProviderManager{}

	// Initialize analyze-enhanced command
	initAnalyzeEnhanced()
	rootCmd.AddCommand(analyzeEnhancedCmd)

	// Initialize config command
	initConfigCommands()
	rootCmd.AddCommand(configCmd)

	// Initialize version command
	initVersionCommand()
	rootCmd.AddCommand(versionCmd)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// Add streaming command
	streamingManager := NewStreamingManager()
	rootCmd.AddCommand(streamingManager.BuildStreamingCommand())

	// Add providers command
	providerCmds := providerCommands.NewProviderCommands()
	rootCmd.AddCommand(providerCmds.BuildProvidersCommand())

	// Legacy config_advanced module removed (unused)

	// Add analytics command
	analyticsConfig := &analytics.Config{
		MLEnabled:           true,
		AnomalyDetection:    true,
		TrendAnalysis:       true,
		EnablePredictions:   true,
		EnableOptimizations: true,
		EnableCaching:       true,
		DefaultCacheTTL:     "1h",
		MaxConcurrency:      10,
		DefaultCurrency:     "USD",
		DefaultTimeFormat:   "2006-01-02",
		StrictTypeChecking:  false,
	}
	analyticsService := analytics.NewBasicService(analyticsConfig, logger)
	analyticsCmds := analyticsCommands.NewAnalyticsCommands(logger, analyticsService)
	analyticsMainCmd := analyticsCmds.BuildAnalyticsCommand()

	// Add enhanced diff command to analytics module
	enhancedDiffCmd := analyticsCommands.BuildEnhancedDiffCommand()
	analyticsMainCmd.AddCommand(enhancedDiffCmd)
	rootCmd.AddCommand(analyticsMainCmd)

	// Add advanced analytics command
	// Experimental analytics commands are registered only when built with -tags experimental
	registerExperimental(rootCmd)

	// Register developer tiny CLI utilities (only when built with -tags devcli)
	registerDevCLI(rootCmd)

	// Add Advanced Analysis module
	analysisCmds := analysisCommands.NewAnalysisCommands(logger)
	rootCmd.AddCommand(analysisCmds.BuildAnalysisCommand())

	// Add diagnostics command (quick status)
	diagCmds := diagnosticsCommands.NewDiagnosticsCommands(logger, providerManager)
	rootCmd.AddCommand(diagCmds.BuildDiagnosticsCommand())

	// No-op reference to alternate DI constructor to keep it reachable for
	// deadcode analyzers without changing runtime behavior.
	_ = diagnosticsCommands.NewDiagnosticsCommandsWithService(
		logger,
		providerManager,
		production.NewBasicProductionService(providerManager, logger),
	)

	// Add reports command with enhanced capabilities
	reportService := reports.NewBasicReportService(logger)
	reportsCmds := reportCommands.NewReportsCommands(reportService, logger)
	reportsMainCmd := reportsCmds.BuildReportsCommand()

	// Add enhanced report command to reports module
	enhancedReportCmd := reportCommands.NewEnhancedReportCommand(logger)
	reportsMainCmd.AddCommand(enhancedReportCmd)
	rootCmd.AddCommand(reportsMainCmd)

	// Add multicloud command
	multicloudCmds := multicloudCommands.NewMulticloudCommands(providerManager)
	mcRoot := multicloudCmds.BuildMulticloudCommand()
	multicloudCmds.AttachEnhancedSubcommands(mcRoot)
	rootCmd.AddCommand(mcRoot)

	// Add production command
	productionCmds := productionCommands.NewProductionCommands(logger, providerManager)
	productionCmds.AddCommands(rootCmd)

	// Add production readiness commands
	productionReadinessCmds := productionCommands.NewProductionReadinessCommands(logger, providerManager)
	rootCmd.AddCommand(productionReadinessCmds.BuildProductionReadinessCommands())

	// Add enhanced production command
	enhancedProductionCmd := productionCommands.BuildEnhancedProductionCommands()
	rootCmd.AddCommand(enhancedProductionCmd)

	// Add integration command
	integrationCmd := integrationCommands.CreateIntegrationCommands()
	rootCmd.AddCommand(integrationCmd)

	// Add monitoring command
	integrationService := integration.NewService()
	productionService := production.NewBasicProductionService(providerManager, logger)
	monitoringCmds := monitoringCommands.NewMonitoringCommands(
		productionService,
		integrationService,
		logger,
	)
	rootCmd.AddCommand(monitoringCmds.BuildCommands())

	// Add security command
	securityCmds := securityCommands.NewSecurityCommands(logger)
	rootCmd.AddCommand(securityCmds.BuildSecurityCommand())

	// Add compliance command
	compCmds := complianceCommands.NewComplianceCommands()
	rootCmd.AddCommand(compCmds.BuildComplianceCommand())

	// Add FOCUS convert command - CORE FUNCTIONALITY
	convertCmd := focusCommands.BuildConvertCommand()
	// Add enhanced convert subcommand
	enhancedConvertCmd := focusCommands.BuildEnhancedConvertCommand()
	convertCmd.AddCommand(enhancedConvertCmd)
	rootCmd.AddCommand(convertCmd)

	// Add FOCUS validate command - ENTERPRISE DATA GOVERNANCE
	validateCmd := validationCommands.BuildValidateCommand()
	rootCmd.AddCommand(validateCmd)

	// Grouped top-level focus command (convert + validate only)
	focusRoot := focusCommands.BuildFocusCommand()
	rootCmd.AddCommand(focusRoot)

	// Add API server command - ENTERPRISE REST API
	apiCmd := apiCommands.BuildAPICommand()
	// Add enhanced API server subcommand
	enhancedAPICmd := apiCommands.BuildEnhancedAPICommand()
	apiCmd.AddCommand(enhancedAPICmd)
	rootCmd.AddCommand(apiCmd)

	// Add Enterprise API server command - COMPLETE ENTERPRISE API
	enterpriseCmd := apiCommands.BuildEnterpriseAPICommand()
	rootCmd.AddCommand(enterpriseCmd)

	// Add optimize command - ENTERPRISE PERFORMANCE OPTIMIZATION
	optimizeCmd := BuildOptimizeCommand()
	rootCmd.AddCommand(optimizeCmd)

	// Add invariants command (utilities for data quality drift)
	invCmd := invariantsCommands.BuildInvariantsCommand()
	rootCmd.AddCommand(invCmd)

	// Add advanced drift command (semantic distribution & trend checks)
	driftCmd := driftCommands.BuildDriftCommand()
	rootCmd.AddCommand(driftCmd)
}
