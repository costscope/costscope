//go:build experimental

package advanced

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	analysisTypes "local/costscope/internal/core/focus/analysis"
	"local/costscope/internal/core/logging"
)

// NewAnalyzeCommand creates the enhanced analyze command for FOCUS data analysis
func NewAnalyzeCommand(logger *logging.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "analyze <dataset>",
		Short: "Perform comprehensive ML-powered analysis of FOCUS dataset",
		Long: `Perform comprehensive analysis of FOCUS datasets with machine learning capabilities
including anomaly detection, cost optimization recommendations, trend analysis, and forecasting.

This enhanced analyze command provides:
- ML-powered anomaly detection using multiple algorithms
- Cost optimization recommendations with savings potential
- Trend analysis with forecasting capabilities
- Multi-dimensional cost analysis
- Executive reporting and insights

Examples:
  # Basic analysis with all features
  costscope analyze dataset.parquet

  # Analysis with specific dimensions and ML features
  costscope analyze dataset.parquet --dimensions service,region --ml --anomalies

  # Forecasting analysis
  costscope analyze dataset.parquet --forecast --periods 30 --trends

  # Optimization-focused analysis
  costscope analyze dataset.parquet --optimize --recommendations --savings-threshold 1000

  # Executive report generation
  costscope analyze dataset.parquet --executive --format html --output report.html

  # Deep analysis with all ML features
  costscope analyze dataset.parquet --deep --ml --anomalies --trends --forecast --optimize`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAnalyze(cmd, args, logger)
		},
	}

	// Analysis scope flags
	cmd.Flags().StringSlice("dimensions", []string{"service", "region"}, "Dimensions to analyze")
	cmd.Flags().StringSlice("services", []string{}, "Specific services to analyze (empty = all)")
	cmd.Flags().StringSlice("regions", []string{}, "Specific regions to analyze (empty = all)")
	cmd.Flags().StringSlice("accounts", []string{}, "Specific accounts to analyze (empty = all)")
	cmd.Flags().String("time-period", "all", "Time period to analyze (all, last-30d, last-7d, etc.)")
	cmd.Flags().Float64("min-cost", 1.0, "Minimum cost threshold for analysis")

	// ML and advanced analytics flags
	cmd.Flags().Bool("ml", false, "Enable machine learning features")
	cmd.Flags().Bool("deep", false, "Enable deep analysis (all ML features)")
	cmd.Flags().Bool("anomalies", false, "Perform anomaly detection")
	cmd.Flags().StringSlice("anomaly-methods", []string{"statistical", "isolation_forest"}, "Anomaly detection methods")
	cmd.Flags().Float64("anomaly-threshold", 0.1, "Anomaly detection sensitivity threshold")
	cmd.Flags().Bool("trends", false, "Perform trend analysis")
	cmd.Flags().Bool("seasonality", false, "Detect seasonal patterns")
	cmd.Flags().Bool("forecast", false, "Generate cost forecasts")
	cmd.Flags().Int("periods", 30, "Number of forecast periods")
	cmd.Flags().Float64("confidence", 0.95, "Confidence level for predictions")

	// Optimization flags
	cmd.Flags().Bool("optimize", false, "Generate cost optimization recommendations")
	cmd.Flags().Bool("recommendations", false, "Include detailed recommendations")
	cmd.Flags().Float64("savings-threshold", 100.0, "Minimum savings threshold for recommendations")
	cmd.Flags().StringSlice("optimization-types", []string{"rightsizing", "reserved_instances", "spot_instances"}, "Types of optimizations to analyze")

	// Reporting flags
	cmd.Flags().Bool("executive", false, "Generate executive summary")
	cmd.Flags().Bool("detailed", true, "Include detailed analysis results")
	cmd.Flags().Bool("insights", false, "Generate actionable insights")
	cmd.Flags().String("format", "json", "Output format (json, csv, html, pdf)")
	cmd.Flags().String("output", "", "Output file path (default: stdout)")
	cmd.Flags().Bool("charts", false, "Include charts and visualizations")
	cmd.Flags().Bool("compress", false, "Compress output")

	// Performance flags
	cmd.Flags().Int("workers", 4, "Number of parallel workers for analysis")
	cmd.Flags().Bool("cache", true, "Cache intermediate results")
	cmd.Flags().Bool("verbose", false, "Verbose output")

	return cmd
}

func runAnalyze(cmd *cobra.Command, args []string, logger *logging.Logger) error {
	datasetFile := args[0]

	// Get flags
	dimensions, _ := cmd.Flags().GetStringSlice("dimensions")
	services, _ := cmd.Flags().GetStringSlice("services")
	regions, _ := cmd.Flags().GetStringSlice("regions")
	accounts, _ := cmd.Flags().GetStringSlice("accounts")
	timePeriod, _ := cmd.Flags().GetString("time-period")
	minCost, _ := cmd.Flags().GetFloat64("min-cost")

	// ML flags
	mlEnabled, _ := cmd.Flags().GetBool("ml")
	deepAnalysis, _ := cmd.Flags().GetBool("deep")
	anomalies, _ := cmd.Flags().GetBool("anomalies")
	anomalyMethods, _ := cmd.Flags().GetStringSlice("anomaly-methods")
	anomalyThreshold, _ := cmd.Flags().GetFloat64("anomaly-threshold")
	trends, _ := cmd.Flags().GetBool("trends")
	seasonality, _ := cmd.Flags().GetBool("seasonality")
	forecast, _ := cmd.Flags().GetBool("forecast")
	periods, _ := cmd.Flags().GetInt("periods")
	confidence, _ := cmd.Flags().GetFloat64("confidence")

	// Optimization flags
	optimize, _ := cmd.Flags().GetBool("optimize")
	savingsThreshold, _ := cmd.Flags().GetFloat64("savings-threshold")
	optimizationTypes, _ := cmd.Flags().GetStringSlice("optimization-types")

	// Reporting flags
	executive, _ := cmd.Flags().GetBool("executive")
	detailed, _ := cmd.Flags().GetBool("detailed")
	insights, _ := cmd.Flags().GetBool("insights")
	format, _ := cmd.Flags().GetString("format")
	output, _ := cmd.Flags().GetString("output")
	charts, _ := cmd.Flags().GetBool("charts")
	compress, _ := cmd.Flags().GetBool("compress")

	// Performance flags
	workers, _ := cmd.Flags().GetInt("workers")
	cache, _ := cmd.Flags().GetBool("cache")
	verbose, _ := cmd.Flags().GetBool("verbose")

	// Validate dataset file
	if !isValidFOCUSFile(datasetFile) {
		return fmt.Errorf("dataset file must be a valid FOCUS dataset (.parquet)")
	}

	// Enable all ML features if deep analysis is requested
	if deepAnalysis {
		mlEnabled = true
		anomalies = true
		trends = true
		seasonality = true
		forecast = true
		optimize = true
		insights = true
	}

	// Create analysis configuration
	config := &analysisTypes.AnalysisConfiguration{
		Dimensions:           dimensions,
		Services:             services,
		Regions:              regions,
		Accounts:             accounts,
		TimePeriod:           timePeriod,
		MinCostThreshold:     minCost,
		MLEnabled:            mlEnabled,
		AnomalyDetection:     anomalies,
		AnomalyMethods:       anomalyMethods,
		AnomalyThreshold:     anomalyThreshold,
		TrendAnalysis:        trends,
		SeasonalityDetection: seasonality,
		ForecastEnabled:      forecast,
		ForecastPeriods:      periods,
		ConfidenceLevel:      confidence,
		OptimizationEnabled:  optimize,
		SavingsThreshold:     savingsThreshold,
		OptimizationTypes:    optimizationTypes,
		ExecutiveSummary:     executive,
		DetailedResults:      detailed,
		GenerateInsights:     insights,
		OutputFormat:         format,
		IncludeCharts:        charts,
		CompressOutput:       compress,
		Workers:              workers,
		CacheEnabled:         cache,
		Verbose:              verbose,
	}

	// Create analysis options
	options := analysisTypes.AnalysisOptions{
		MLEnabled:            mlEnabled,
		AnomalyDetection:     anomalies,
		TrendAnalysis:        trends,
		OptimizationAnalysis: optimize,
		ForecastDays:         periods,
		ConfidenceLevel:      confidence,
		OutputFormat:         format,
		Verbose:              verbose,
	}

	logger.Info(fmt.Sprintf("Starting enhanced FOCUS dataset analysis: dataset=%s, dimensions=%v, ml_enabled=%v, deep_analysis=%v",
		datasetFile, dimensions, mlEnabled, deepAnalysis))

	// Create analysis engine (this would be implemented)
	engine := analysisTypes.NewEngine(logger, config)

	// Perform comprehensive analysis
	result, err := engine.AnalyzeFOCUSDataset(datasetFile, options)
	if err != nil {
		return fmt.Errorf("analysis failed: %w", err)
	}

	// Output results
	if output != "" {
		err = engine.ExportResults(result, format, output)
		if err != nil {
			return fmt.Errorf("failed to export results: %w", err)
		}
		logger.Info(fmt.Sprintf("Analysis results exported to: %s", output))
	} else {
		// Print to stdout
		err = printAnalysisResults(result, format, logger)
		if err != nil {
			return fmt.Errorf("failed to print results: %w", err)
		}
	}

	// Print analysis summary
	printAnalysisSummary(result, logger)

	return nil
}

func printAnalysisResults(result *analysisTypes.AnalysisResult, format string, logger *logging.Logger) error {
	logger.Info(fmt.Sprintf("Printing analysis results in %s format", format))

	switch strings.ToLower(format) {
	case "json":
		return printJSONAnalysisResults(result)
	case "csv":
		return printCSVAnalysisResults(result)
	case "table":
		return printTableAnalysisResults(result)
	default:
		logger.Warn("Unknown format specified, defaulting to JSON")
		return printJSONAnalysisResults(result) // Default to JSON
	}
}

func printJSONAnalysisResults(result *analysisTypes.AnalysisResult) error {
	fmt.Printf("{\n")
	fmt.Printf("  \"summary\": {\n")
	fmt.Printf("    \"total_cost\": %.2f,\n", result.Summary.TotalCost)
	fmt.Printf("    \"cost_growth_rate\": %.2f,\n", result.Summary.CostGrowthRate)
	fmt.Printf("    \"services_count\": %d,\n", result.Summary.ServicesCount)
	fmt.Printf("    \"regions_count\": %d,\n", result.Summary.RegionsCount)
	fmt.Printf("    \"anomalies_detected\": %d,\n", result.Summary.AnomaliesDetected)
	fmt.Printf("    \"optimizations_found\": %d,\n", result.Summary.OptimizationsFound)
	fmt.Printf("    \"potential_savings\": %.2f\n", result.Summary.PotentialSavings)
	fmt.Printf("  },\n")
	fmt.Printf("  \"trends_count\": %d,\n", len(result.CostTrends))
	fmt.Printf("  \"anomalies_count\": %d,\n", len(result.Anomalies))
	fmt.Printf("  \"optimizations_count\": %d,\n", len(result.Optimizations))
	fmt.Printf("  \"forecasts_count\": %d\n", len(result.Forecasts))
	fmt.Printf("}\n")
	return nil
}

func printCSVAnalysisResults(result *analysisTypes.AnalysisResult) error {
	// Print CSV header for service breakdown
	fmt.Printf("service,region,total_cost,daily_average,growth_rate,anomalies,optimization_potential\n")

	// Print service breakdown
	for _, service := range result.ServiceBreakdown {
		fmt.Printf("%s,%s,%.2f,%.2f,%.2f,%d,%.2f\n",
			service.Service, service.Region, service.TotalCost,
			service.AvgDailyCost, service.CostPercentage,
			service.AnomaliesCount, service.PotentialSavings)
	}

	return nil
}

func printTableAnalysisResults(result *analysisTypes.AnalysisResult) error {
	fmt.Printf("\n╔════════════════════════════════════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║                                 FOCUS ANALYSIS RESULTS                                    ║\n")
	fmt.Printf("╠════════════════════════════════════════════════════════════════════════════════════════╣\n")

	// Summary
	fmt.Printf("║ SUMMARY                                                                                    ║\n")
	fmt.Printf("║   Total Cost: $%18.2f   Average Daily: $%12.2f                          ║\n",
		result.Summary.TotalCost, result.Summary.TotalCost/30) // Estimate daily average
	fmt.Printf("║   Growth Rate: %6.2f%%              Services: %3d      Regions: %3d          ║\n",
		result.Summary.CostGrowthRate, result.Summary.ServicesCount, result.Summary.RegionsCount)
	fmt.Printf("║   Anomalies: %3d                    Optimizations: %3d                           ║\n",
		result.Summary.AnomaliesDetected, result.Summary.OptimizationsFound)
	fmt.Printf("║   Potential Savings: $%12.2f                                               ║\n",
		result.Summary.PotentialSavings)
	fmt.Printf("╠════════════════════════════════════════════════════════════════════════════════════════╣\n")

	// Top services
	if len(result.ServiceBreakdown) > 0 {
		fmt.Printf("║ TOP SERVICES                                                                               ║\n")
		fmt.Printf("║ Service                   Cost ($)      Daily ($)    Growth (%%)   Savings ($)          ║\n")
		fmt.Printf("║────────────────────────────────────────────────────────────────────────────────────────║\n")

		maxDisplayed := 10
		if len(result.ServiceBreakdown) < maxDisplayed {
			maxDisplayed = len(result.ServiceBreakdown)
		}

		for i := 0; i < maxDisplayed; i++ {
			service := result.ServiceBreakdown[i]
			fmt.Printf("║ %-24s %10.2f %10.2f %10.2f %10.2f        ║\n",
				truncateString(service.Service, 24),
				service.TotalCost,
				service.AvgDailyCost,
				service.CostPercentage,
				service.PotentialSavings)
		}
	}

	fmt.Printf("╚════════════════════════════════════════════════════════════════════════════════════════╝\n")

	return nil
}

func printAnalysisSummary(result *analysisTypes.AnalysisResult, logger *logging.Logger) {
	logger.Info("Printing analysis summary")

	fmt.Printf("\n Enhanced FOCUS Dataset Analysis Summary\n")
	fmt.Printf("═══════════════════════════════════════════\n")

	// Cost overview
	fmt.Printf(" Total Cost: $%.2f (Daily Avg: $%.2f)\n",
		result.Summary.TotalCost, result.Summary.TotalCost/30) // Estimate daily average

	// Growth rate
	growthIcon := ""
	if result.Summary.CostGrowthRate < 0 {
		growthIcon = ""
	}
	fmt.Printf("%s Cost Growth Rate: %.2f%%\n", growthIcon, result.Summary.CostGrowthRate)

	// Coverage
	fmt.Printf(" Coverage: %d services across %d regions\n",
		result.Summary.ServicesCount, result.Summary.RegionsCount)

	// ML insights
	if result.Summary.AnomaliesDetected > 0 {
		fmt.Printf(" Anomalies Detected: %d\n", result.Summary.AnomaliesDetected)
	}

	if result.Summary.OptimizationsFound > 0 {
		fmt.Printf(" Optimization Opportunities: %d\n", result.Summary.OptimizationsFound)
		fmt.Printf(" Potential Savings: $%.2f\n", result.Summary.PotentialSavings)
	}

	// Analysis details
	if len(result.CostTrends) > 0 {
		fmt.Printf(" Trends Analyzed: %d\n", len(result.CostTrends))
	}

	if len(result.Forecasts) > 0 {
		fmt.Printf(" Forecasts Generated: %d periods\n", len(result.Forecasts))
	}

	// Processing info
	fmt.Printf("⏱️  Processing Time: %s\n", result.Metadata.ProcessingTime)
	fmt.Printf(" Analysis Date: %s\n", result.Metadata.AnalysisDate.Format("2006-01-02 15:04:05"))

	// Top cost services
	if len(result.Summary.TopCostServices) > 0 {
		fmt.Printf("\n Top Cost Services: %s\n", strings.Join(result.Summary.TopCostServices, ", "))
	}

	// Top growth services
	if len(result.Summary.TopGrowthServices) > 0 {
		fmt.Printf(" Fastest Growing: %s\n", strings.Join(result.Summary.TopGrowthServices, ", "))
	}

	fmt.Printf("\n")
}

func isValidFOCUSFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return ext == ".parquet"
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// GetAnalyzeUsageExamples returns usage examples for the analyze command
func GetAnalyzeUsageExamples() []string {
	return []string{
		"# Basic comprehensive analysis",
		"costscope analyze dataset.parquet",
		"",
		"# Deep ML-powered analysis with all features",
		"costscope analyze dataset.parquet --deep",
		"",
		"# Focus on specific services and regions",
		"costscope analyze dataset.parquet \\",
		"  --services ec2,s3,rds \\",
		"  --regions us-east-1,us-west-2",
		"",
		"# Anomaly detection with custom sensitivity",
		"costscope analyze dataset.parquet \\",
		"  --anomalies --anomaly-threshold 0.05 \\",
		"  --anomaly-methods statistical,isolation_forest",
		"",
		"# Cost optimization analysis",
		"costscope analyze dataset.parquet \\",
		"  --optimize --recommendations \\",
		"  --savings-threshold 500 \\",
		"  --optimization-types rightsizing,reserved_instances",
		"",
		"# Forecasting analysis",
		"costscope analyze dataset.parquet \\",
		"  --forecast --periods 60 \\",
		"  --trends --seasonality \\",
		"  --confidence 0.95",
		"",
		"# Executive report generation",
		"costscope analyze dataset.parquet \\",
		"  --executive --insights \\",
		"  --format html --output executive-report.html \\",
		"  --charts",
		"",
		"# Performance tuned analysis",
		"costscope analyze dataset.parquet \\",
		"  --workers 8 --cache \\",
		"  --ml --verbose",
	}
}
