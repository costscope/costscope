package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"local/costscope/internal/core/logging"
	"local/costscope/internal/core/production"
	"local/costscope/internal/providers"

	"github.com/spf13/cobra"
)

const (
	outputFormatJSON = "json"
)

// ProductionCommands provides production readiness commands
type ProductionCommands struct {
	logger          *logging.Logger
	providerManager *providers.ProviderManager
	productionSvc   production.ProductionService
}

// NewProductionCommands creates a new production commands instance
func NewProductionCommands(logger *logging.Logger, providerManager *providers.ProviderManager) *ProductionCommands {
	productionSvc := production.NewBasicProductionService(providerManager, logger)

	return &ProductionCommands{
		logger:          logger,
		providerManager: providerManager,
		productionSvc:   productionSvc,
	}
}

// AddCommands adds production commands to the parent command
func (pc *ProductionCommands) AddCommands(parentCmd *cobra.Command) {
	productionCmd := &cobra.Command{
		Use:   "production",
		Short: "Production readiness assessment and management",
		Long:  "Comprehensive production readiness assessment, optimization analysis, and deployment management",
	}

	// Add subcommands
	productionCmd.AddCommand(pc.createStatusCommand())
	productionCmd.AddCommand(pc.createOptimizeCommand())
	productionCmd.AddCommand(pc.createHealthCommand())

	parentCmd.AddCommand(productionCmd)
}

// createStatusCommand creates the status command
func (pc *ProductionCommands) createStatusCommand() *cobra.Command {
	var outputFormat string
	var showDetails bool

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Get comprehensive production system status",
		Long:  "Retrieve detailed production system status including health, performance, security, and readiness metrics",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			pc.logger.Info("Getting production system status")

			// Get system status
			status, err := pc.productionSvc.GetSystemStatus(ctx)
			if err != nil {
				return fmt.Errorf("failed to get system status: %w", err)
			}

			return pc.outputStatus(status, outputFormat, showDetails)
		},
	}

	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format (table, json)")
	cmd.Flags().BoolVarP(&showDetails, "details", "d", false, "Show detailed metrics")

	return cmd
}

// createOptimizeCommand creates the optimize command
func (pc *ProductionCommands) createOptimizeCommand() *cobra.Command {
	var outputFormat string

	cmd := &cobra.Command{
		Use:   "optimize",
		Short: "Run system optimization analysis",
		Long:  "Analyze system for optimization opportunities and generate actionable recommendations",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			pc.logger.Info("Running optimization analysis")

			// Prepare optimization options
			options := &production.OptimizationOptions{
				Aggressive:     false,
				DryRun:         true,
				Categories:     []string{"performance", "security", "cost"},
				MinImpact:      production.ImpactMedium,
				MaxEffort:      production.EffortHigh,
				Budget:         100000.0,
				TimelineMonths: 6,
			}

			// Run optimization
			results, err := pc.productionSvc.RunOptimization(ctx, options)
			if err != nil {
				return fmt.Errorf("failed to run optimization: %w", err)
			}

			return pc.outputOptimizationReport(results, outputFormat)
		},
	}

	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format (table, json)")

	return cmd
}

// createHealthCommand creates the health command
func (pc *ProductionCommands) createHealthCommand() *cobra.Command {
	var outputFormat string

	cmd := &cobra.Command{
		Use:   "health",
		Short: "Run comprehensive health checks",
		Long:  "Execute comprehensive health checks across all system components",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			pc.logger.Info("Running health checks")

			healthChecks, err := pc.productionSvc.GetHealthChecks(ctx)
			if err != nil {
				return fmt.Errorf("failed to run health checks: %w", err)
			}

			return pc.outputHealthChecks(healthChecks, outputFormat)
		},
	}

	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format (table, json)")

	return cmd
}

// Output formatting methods

func (pc *ProductionCommands) outputStatus(status *production.ProductionSystemMetrics, format string, details bool) error {
	switch format {
	case outputFormatJSON:
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(status)
	default:
		return pc.printStatusTable(status, details)
	}
}

func (pc *ProductionCommands) outputOptimizationReport(report *production.ProductionOptimizationReport, format string) error {
	switch format {
	case outputFormatJSON:
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(report)
	default:
		return pc.printOptimizationTable(report)
	}
}

func (pc *ProductionCommands) outputHealthChecks(checks map[string]production.CheckResult, format string) error {
	switch format {
	case outputFormatJSON:
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(checks)
	default:
		return pc.printHealthTable(checks)
	}
}

// Table formatting methods

func (pc *ProductionCommands) printStatusTable(status *production.ProductionSystemMetrics, details bool) error {
	fmt.Println("═══════════════════════════════════════════════════")
	fmt.Println("              PRODUCTION SYSTEM STATUS")
	fmt.Println("═══════════════════════════════════════════════════")

	fmt.Printf("Overall Readiness Score: %d/100\n", status.ReadinessScore)
	fmt.Printf("Health Score:           %d/100\n", status.SystemHealth.HealthScore)
	fmt.Printf("Performance Score:      %d/100\n", status.Performance.OptimizationScore)
	fmt.Printf("Security Score:         %d/100\n", status.Security.SecurityScore)
	fmt.Printf("Integration Score:      %d/100\n", status.Integration.IntegrationScore)
	fmt.Printf("Analytics Models:       %d\n", status.Analytics.MLModelsActive)

	fmt.Println("\n--- CRITICAL ISSUES ---")
	if len(status.CriticalIssues) == 0 {
		fmt.Println("No critical issues found")
	} else {
		for _, issue := range status.CriticalIssues {
			fmt.Printf("• %s\n", issue)
		}
	}

	if details {
		fmt.Println("\n--- DETAILED METRICS ---")
		fmt.Printf("Total Commands:    %d\n", status.TotalCommands)
		fmt.Printf("Total Endpoints:   %d\n", status.TotalEndpoints)
		fmt.Printf("Total Features:    %d\n", status.TotalFeatures)
		fmt.Printf("Completion Level:  %s\n", status.CompletionLevel)
		fmt.Printf("Production Ready:  %t\n", status.ProductionReady)
	}

	fmt.Printf("\nProcessing Time: %dms\n", status.ProcessingTimeMs)

	return nil
}

func (pc *ProductionCommands) printOptimizationTable(report *production.ProductionOptimizationReport) error {
	fmt.Println("═══════════════════════════════════════════════════")
	fmt.Println("              OPTIMIZATION ANALYSIS")
	fmt.Println("═══════════════════════════════════════════════════")

	if report.OptimizationResults.TotalImprovements > 0 {
		fmt.Printf("Total Improvements:     %d\n", report.OptimizationResults.TotalImprovements)
		fmt.Printf("Performance Gains:      %.1f%%\n", report.OptimizationResults.PerformanceGains)
		fmt.Printf("Cost Savings:          $%.0f\n", report.OptimizationResults.CostSavings)
		fmt.Printf("Security Enhancements:  %d\n", report.OptimizationResults.SecurityEnhancements)
		fmt.Printf("Efficiency Gains:       %.1f%%\n", report.OptimizationResults.EfficiencyGains)
		fmt.Printf("Optimization Score:     %d/100\n", report.OptimizationResults.OptimizationScore)
	}

	if len(report.Recommendations) > 0 {
		fmt.Println("\n--- RECOMMENDATIONS ---")
		for _, rec := range report.Recommendations {
			fmt.Printf("• [%s] %s: %s\n", string(rec.Priority), rec.Type, rec.Title)
		}
	}

	return nil
}

func (pc *ProductionCommands) printHealthTable(checks map[string]production.CheckResult) error {
	fmt.Println("═══════════════════════════════════════════════════")
	fmt.Println("              HEALTH CHECK RESULTS")
	fmt.Println("═══════════════════════════════════════════════════")

	for component, result := range checks {
		status := " HEALTHY"
		if result.Status != statusPassed {
			status = " UNHEALTHY"
		}
		fmt.Printf("%-20s %s - %s\n", component, status, result.Message)
	}

	return nil
}
