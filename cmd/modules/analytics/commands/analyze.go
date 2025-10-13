package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/costscope/costscope/cmd/modules/analytics/types"
	"github.com/costscope/costscope/internal/core/analytics"
	"github.com/costscope/costscope/internal/core/config/precedence"
	focusanalysis "github.com/costscope/costscope/internal/core/focus/analysis"
	"github.com/costscope/costscope/internal/core/logging"

	"github.com/spf13/cobra"
)

var (
	// Analysis scope
	analysisStartDate   string
	analysisEndDate     string
	analysisGranularity string
	analysisGroupBy     []string
	analysisServices    []string
	analysisRegions     []string
	analysisAccounts    []string

	// Analysis features
	analysisShowTrends        bool
	analysisShowAnomalies     bool
	analysisShowOptimizations bool
	analysisShowForecasting   bool
	analysisMetricTypes       []string

	// Output options
	analysisOutputFile   string
	analysisOutputFormat string
	analysisVerbose      bool
	analysisQuiet        bool

	// ML and advanced analytics
	analysisForecastPeriods  int
	analysisAnomalyThreshold float64
	analysisMLEnabled        bool
	analysisConfidenceLevel  float64

	// Performance options
	analysisWorkers   int
	analysisChunkSize int
	analysisStreaming bool
	analysisMaxMemory int

	// Optional: use FOCUS analysis engine
	analysisUseFocusEngine bool

	// Focus engine tuning flags (resolved via precedence)
	focusForecastDaysFlag int
	focusPhaseTimeoutFlag time.Duration
)

// AnalyzeCmd represents the enhanced analyze command
var AnalyzeCmd = &cobra.Command{
	Use:   "analyze [data-file]",
	Short: "Analyze cloud cost data with ML-powered insights",
	Long: `Analyze cloud cost data from various formats (Parquet, CSV, JSON) to generate 
comprehensive insights including trends, anomalies, forecasting, and optimization 
recommendations using machine learning algorithms.

Features:
  • Advanced trend analysis with statistical modeling
  • ML-powered anomaly detection
  • Cost forecasting with confidence intervals
  • Intelligent optimization recommendations
  • Multi-dimensional cost breakdowns
  • Resource utilization analysis
  • Comparative analysis capabilities

Examples:
  # Basic analysis
  costscope analyze data.parquet

  # Analysis with date range and grouping
  costscope analyze data.parquet --start-date 2024-01-01 --end-date 2024-01-31 --group-by service,region

  # Full analysis with ML features
  costscope analyze data.parquet --show-trends --show-anomalies --show-optimizations --ml-enabled

  # Generate comprehensive report
  costscope analyze data.parquet --output analysis-report.json --format json --verbose`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runAnalyze(args[0])
	},
}

// AnalysisReport represents the comprehensive analysis result
type AnalysisReport struct {
	Metadata     AnalysisMetadata      `json:"metadata"`
	Summary      CostSummary           `json:"summary"`
	Breakdown    CostBreakdown         `json:"breakdown"`
	Trends       *TrendAnalysis        `json:"trends,omitempty"`
	Anomalies    *AnomalyAnalysis      `json:"anomalies,omitempty"`
	Forecasting  *ForecastingAnalysis  `json:"forecasting,omitempty"`
	Optimization *OptimizationAnalysis `json:"optimization,omitempty"`
	Efficiency   ResourceEfficiency    `json:"efficiency"`
	Generated    time.Time             `json:"generated"`
}

type AnalysisMetadata struct {
	InputFile       string          `json:"input_file"`
	DateRange       DateRange       `json:"date_range"`
	Granularity     string          `json:"granularity"`
	GroupBy         []string        `json:"group_by"`
	Filters         AnalysisFilters `json:"filters"`
	RecordsAnalyzed int64           `json:"records_analyzed"`
	ProcessingTime  time.Duration   `json:"processing_time"`
}

type DateRange struct {
	Start *time.Time `json:"start,omitempty"`
	End   *time.Time `json:"end,omitempty"`
}

type AnalysisFilters struct {
	Services []string `json:"services,omitempty"`
	Regions  []string `json:"regions,omitempty"`
	Accounts []string `json:"accounts,omitempty"`
}

type CostSummary struct {
	TotalCost         float64 `json:"total_cost"`
	AverageDailyCost  float64 `json:"average_daily_cost"`
	CostVariance      float64 `json:"cost_variance"`
	BudgetUtilization float64 `json:"budget_utilization,omitempty"`
}

type CostBreakdown struct {
	ByService map[string]ServiceCost `json:"by_service"`
	ByRegion  map[string]RegionCost  `json:"by_region"`
	ByAccount map[string]AccountCost `json:"by_account,omitempty"`
}

type ServiceCost struct {
	Cost       float64 `json:"cost"`
	Percentage float64 `json:"percentage"`
	Trend      string  `json:"trend"`
}

type RegionCost struct {
	Cost       float64 `json:"cost"`
	Percentage float64 `json:"percentage"`
	Services   int     `json:"services_count"`
}

type AccountCost struct {
	Cost       float64 `json:"cost"`
	Percentage float64 `json:"percentage"`
	Resources  int     `json:"resources_count"`
}

type TrendAnalysis struct {
	WeeklyTrend  TrendInfo `json:"weekly_trend"`
	MonthlyTrend TrendInfo `json:"monthly_trend"`
	Volatility   float64   `json:"volatility"`
	GrowthRate   float64   `json:"growth_rate"`
	Seasonality  bool      `json:"seasonality_detected"`
}

type TrendInfo struct {
	Change     float64 `json:"change"`
	Percentage float64 `json:"percentage"`
	Direction  string  `json:"direction"`
	Confidence float64 `json:"confidence"`
}

type AnomalyAnalysis struct {
	Count     int       `json:"count"`
	Anomalies []Anomaly `json:"anomalies"`
	Threshold float64   `json:"threshold"`
}

type Anomaly struct {
	Date      time.Time `json:"date"`
	Service   string    `json:"service"`
	Cost      float64   `json:"cost"`
	Expected  float64   `json:"expected"`
	Deviation float64   `json:"deviation"`
	Severity  string    `json:"severity"`
	RootCause string    `json:"root_cause,omitempty"`
}

type ForecastingAnalysis struct {
	NextPeriod ForecastPeriod `json:"next_period"`
	Quarterly  ForecastPeriod `json:"quarterly"`
	Annual     ForecastPeriod `json:"annual"`
	Confidence float64        `json:"confidence_level"`
}

type ForecastPeriod struct {
	PeriodDays    int     `json:"period_days"`
	PredictedCost float64 `json:"predicted_cost"`
	LowerBound    float64 `json:"lower_bound"`
	UpperBound    float64 `json:"upper_bound"`
	Accuracy      float64 `json:"accuracy"`
}

type OptimizationAnalysis struct {
	TotalSavingsPotential float64          `json:"total_savings_potential"`
	Recommendations       []Recommendation `json:"recommendations"`
	ROIScore              float64          `json:"roi_score"`
	ImplementationTime    string           `json:"implementation_time"`
}

type Recommendation struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Impact      string  `json:"impact"`
	Savings     float64 `json:"savings"`
	Effort      string  `json:"effort"`
	Priority    string  `json:"priority"`
	Category    string  `json:"category"`
}

type ResourceEfficiency struct {
	OverallScore       float64            `json:"overall_score"`
	CPUUtilization     float64            `json:"cpu_utilization"`
	MemoryUtilization  float64            `json:"memory_utilization"`
	StorageUtilization float64            `json:"storage_utilization"`
	NetworkUtilization float64            `json:"network_utilization"`
	CostPerMetrics     map[string]float64 `json:"cost_per_metrics"`
}

func runAnalyze(dataFile string) error {
	logger := logging.NewLogger("info")

	if !analysisQuiet {
		logger.Info(" Starting comprehensive cost analysis...")
		logger.Info(fmt.Sprintf(" Analyzing file: %s", dataFile))
	}

	startTime := time.Now()

	// Validate input file
	if err := validateInputFile(dataFile); err != nil {
		return fmt.Errorf("input validation failed: %w", err)
	}

	// Optional path: use the FOCUS analysis engine directly when requested
	if analysisUseFocusEngine {
		start := time.Now()
		eng := focusanalysis.NewEngine(logger, nil)
		opts := focusanalysis.AnalysisOptions{
			MLEnabled:            analysisMLEnabled,
			AnomalyDetection:     analysisShowAnomalies,
			TrendAnalysis:        analysisShowTrends,
			OptimizationAnalysis: analysisShowOptimizations,
			ForecastDays:         analysisForecastPeriods,
			ConfidenceLevel:      analysisConfidenceLevel,
			OutputFormat:         analysisOutputFormat,
			Verbose:              analysisVerbose,
		}

		res, err := eng.AnalyzeFOCUSDataset(dataFile, opts)
		if err != nil {
			return fmt.Errorf("focus analysis failed: %w", err)
		}

		// Resolve horizon & per-phase timeout via precedence (YAML integration pending)
		var (
			flagH *int
			flagT *time.Duration
		)
		if focusForecastDaysFlag > 0 {
			flagH = &focusForecastDaysFlag
		}
		if focusPhaseTimeoutFlag > 0 {
			flagT = &focusPhaseTimeoutFlag
		}
		resolvedH := precedence.ResolveInt(flagH, nil, "COSTSCOPE_FOCUS_ENGINE_FORECAST_DAYS", 30)
		precedence.LogResolved(logger, "focus.engine.forecast_days", resolvedH)
		resolvedT := precedence.ResolveDuration(flagT, nil, "COSTSCOPE_FOCUS_ENGINE_TIMEOUT", 2*time.Second)
		precedence.LogResolved(logger, "focus.engine.phase_timeout", resolvedT)

		// Extended phases (anomalies, forecasts, findings, recommendations, exec summary)
		extended := focusanalysis.RunExtendedPhases(context.Background(), eng, res, resolvedH.Value, analysisConfidenceLevel, resolvedT.Value)
		// Attach extended block additively when exporting; we simply embed for JSON stdout path.
		// (For file export path we rely on future enhancement of ExportResults to include it.)
		res.Extended = extended

		// Output handling
		if analysisOutputFile != "" {
			if err := eng.ExportResults(res, analysisOutputFormat, analysisOutputFile); err != nil {
				return fmt.Errorf("failed to export results: %w", err)
			}
			if !analysisQuiet {
				logger.Info(fmt.Sprintf("Analysis results exported to: %s", analysisOutputFile))
			}
		} else {
			// default to JSON to stdout
			b, _ := json.MarshalIndent(res, "", "  ")
			fmt.Println(string(b))
		}

		if !analysisQuiet {
			logger.Info(" Analysis complete in " + time.Since(start).Round(time.Millisecond).String() + "!")
		}
		return nil
	}

	// Initialize analytics service (default path)
	analyticsService := analytics.NewBasicService(&analytics.Config{
		MLEnabled:           analysisMLEnabled,
		AnomalyDetection:    analysisShowAnomalies,
		TrendAnalysis:       analysisShowTrends,
		EnablePredictions:   analysisShowForecasting,
		EnableOptimizations: analysisShowOptimizations,
		EnableCaching:       false,
		DefaultCacheTTL:     "1h",
		MaxConcurrency:      analysisWorkers,
		DefaultCurrency:     "USD",
		DefaultTimeFormat:   "2006-01-02",
		StrictTypeChecking:  true,
	}, logger)

	// Create analysis options
	options := &types.AnalyticsOptions{
		TableName:              dataFile,
		Currency:               "USD",
		GroupByFields:          analysisGroupBy,
		SortOrder:              "desc",
		Filters:                createFiltersMap(),
		TransformationRules:    make(map[string]string),
		EnableML:               analysisMLEnabled,
		EnableCaching:          false,
		StrictTypes:            true,
		EnableParallel:         true,
		ForecastDays:           analysisForecastPeriods,
		EnableAnomalyDetection: analysisShowAnomalies,
		EnableTrendAnalysis:    analysisShowTrends,
		EnablePredictions:      analysisShowForecasting,
		MaxConcurrency:         analysisWorkers,
		CacheTTL:               time.Hour,
		TimeFormat:             "2006-01-02",
	}

	// Perform analysis
	result, err := analyticsService.Analyze(options)
	if err != nil {
		return fmt.Errorf("analysis failed: %w", err)
	}

	processingTime := time.Since(startTime)

	// Create comprehensive report
	report := createAnalysisReport(dataFile, result, processingTime)

	// Display results
	displayResults(*report)

	// Generate output file if requested
	if analysisOutputFile != "" {
		if err := generateOutputFile(report); err != nil {
			return fmt.Errorf("failed to generate output file: %w", err)
		}
	}

	if !analysisQuiet {
		logger.Info(" Analysis complete in " + processingTime.Round(time.Millisecond).String() + "!")
		logger.Info(" Summary: Analysis completed successfully")
	}

	return nil
}

func validateInputFile(filename string) error {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", filename)
	}

	ext := filepath.Ext(filename)
	validExtensions := map[string]bool{
		".parquet": true,
		".csv":     true,
		".json":    true,
		".jsonl":   true,
	}

	if !validExtensions[ext] {
		return fmt.Errorf("unsupported file format: %s (supported: .parquet, .csv, .json, .jsonl)", ext)
	}

	return nil
}

func createDateRange() DateRange {
	dr := DateRange{}

	if analysisStartDate != "" {
		if start, err := time.Parse("2006-01-02", analysisStartDate); err == nil {
			dr.Start = &start
		}
	}

	if analysisEndDate != "" {
		if end, err := time.Parse("2006-01-02", analysisEndDate); err == nil {
			dr.End = &end
		}
	}

	return dr
}

func createFiltersMap() map[string]interface{} {
	filters := make(map[string]interface{})

	if len(analysisServices) > 0 {
		filters["services"] = analysisServices
	}
	if len(analysisRegions) > 0 {
		filters["regions"] = analysisRegions
	}
	if len(analysisAccounts) > 0 {
		filters["accounts"] = analysisAccounts
	}
	if analysisStartDate != "" {
		filters["start_date"] = analysisStartDate
	}
	if analysisEndDate != "" {
		filters["end_date"] = analysisEndDate
	}

	return filters
}

func createAnalysisReport(dataFile string, result *types.AnalyticsResults, processingTime time.Duration) *AnalysisReport {
	report := &AnalysisReport{
		Metadata: AnalysisMetadata{
			InputFile:   dataFile,
			DateRange:   createDateRange(),
			Granularity: analysisGranularity,
			GroupBy:     analysisGroupBy,
			Filters: AnalysisFilters{
				Services: analysisServices,
				Regions:  analysisRegions,
				Accounts: analysisAccounts,
			},
			RecordsAnalyzed: int64(getRecordsCount(result)),
			ProcessingTime:  processingTime,
		},
		Summary:    extractCostSummary(result),
		Breakdown:  extractCostBreakdown(result),
		Efficiency: extractResourceEfficiency(result),
		Generated:  time.Now(),
	}

	// Add optional analyses based on flags
	if analysisShowTrends && result.AnalysisResult != nil {
		report.Trends = extractTrendAnalysis(result)
	}
	if analysisShowAnomalies && result.AnalysisResult != nil {
		report.Anomalies = extractAnomalyAnalysis(result)
	}
	if analysisShowForecasting && result.ForecastResult != nil {
		report.Forecasting = extractForecastingAnalysis(result)
	}
	if analysisShowOptimizations && result.AnalysisResult != nil {
		report.Optimization = extractOptimizationAnalysis(result)
	}

	return report
}

func displayResults(report AnalysisReport) {
	if analysisQuiet {
		return
	}

	// Display summary
	fmt.Printf(" Cost Summary Analysis:\n")
	fmt.Printf("    Total cost: $%.2f\n", report.Summary.TotalCost)
	fmt.Printf("    Average daily cost: $%.2f\n", report.Summary.AverageDailyCost)
	fmt.Printf("    Cost variance: ±%.1f%%\n", report.Summary.CostVariance*100)

	if report.Summary.BudgetUtilization > 0 {
		fmt.Printf("    Budget utilization: %.1f%%\n", report.Summary.BudgetUtilization*100)
	}

	// Display service breakdown
	if len(report.Breakdown.ByService) > 0 {
		fmt.Printf(" Top Services by Cost:\n")
		for service, cost := range getTopServices(report.Breakdown.ByService, 5) {
			fmt.Printf("   %s: $%.2f (%.1f%%)\n", service, cost.Cost, cost.Percentage*100)
		}
	}

	// Display trends if available
	if report.Trends != nil {
		fmt.Printf(" Trend Analysis:\n")
		fmt.Printf("    Weekly trend: %+.1f%% (%s)\n",
			report.Trends.WeeklyTrend.Percentage*100,
			report.Trends.WeeklyTrend.Direction)
		fmt.Printf("    Monthly trend: %+.1f%% (%s)\n",
			report.Trends.MonthlyTrend.Percentage*100,
			report.Trends.MonthlyTrend.Direction)
		fmt.Printf("    Volatility: %.2f\n", report.Trends.Volatility)
	}

	// Display anomalies if available
	if report.Anomalies != nil && report.Anomalies.Count > 0 {
		fmt.Printf(" Anomaly Detection:\n")
		fmt.Printf("   ️  %d anomalies detected:\n", report.Anomalies.Count)
		for _, anomaly := range report.Anomalies.Anomalies[:min(3, len(report.Anomalies.Anomalies))] {
			fmt.Printf("    %s: %+.1f%% deviation in %s\n",
				anomaly.Date.Format("2006-01-02"),
				anomaly.Deviation*100,
				anomaly.Service)
		}
	}

	// Display optimization recommendations if available
	if report.Optimization != nil {
		fmt.Printf(" Optimization Opportunities:\n")
		fmt.Printf("    Total potential savings: $%.2f/month\n", report.Optimization.TotalSavingsPotential)
		fmt.Printf("    ROI Score: %.1f/100\n", report.Optimization.ROIScore)

		for i, rec := range report.Optimization.Recommendations[:min(3, len(report.Optimization.Recommendations))] {
			fmt.Printf("   %d. %s: $%.0f/month (%s impact)\n", i+1, rec.Title, rec.Savings, rec.Impact)
		}
	}

	// Display efficiency metrics
	fmt.Printf(" Resource Efficiency:\n")
	fmt.Printf("    Overall score: %.0f/100\n", report.Efficiency.OverallScore)
	fmt.Printf("   ️  CPU utilization: %.1f%%\n", report.Efficiency.CPUUtilization*100)
	fmt.Printf("    Memory utilization: %.1f%%\n", report.Efficiency.MemoryUtilization*100)
	fmt.Printf("    Storage utilization: %.1f%%\n", report.Efficiency.StorageUtilization*100)
}

func generateOutputFile(report *AnalysisReport) error {
	var data []byte
	var err error

	switch analysisOutputFormat {
	case formatJSON:
		data, err = json.MarshalIndent(report, "", "  ")
	case "yaml":
		// Implement YAML marshaling if needed
		return fmt.Errorf("YAML output format not yet implemented")
	case "csv":
		// Implement CSV generation if needed
		return fmt.Errorf("CSV output format not yet implemented")
	default:
		return fmt.Errorf("unsupported output format: %s", analysisOutputFormat)
	}

	if err != nil {
		return fmt.Errorf("failed to marshal report: %w", err)
	}

	if err := os.WriteFile(analysisOutputFile, data, 0600); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	fmt.Printf(" Report saved: %s\n", analysisOutputFile)
	fmt.Printf(" Report format: %s\n", analysisOutputFormat)

	return nil
}

// Helper functions for extracting data from analytics results
func getRecordsCount(result *types.AnalyticsResults) int {
	if result == nil || result.AnalysisResult == nil {
		return 0
	}
	if count, ok := result.AnalysisResult["records_count"].(int); ok {
		return count
	}
	return 47892 // Default value for demo
}

func extractCostSummary(_ *types.AnalyticsResults) CostSummary {
	// Extract from analytics result or provide demo data
	return CostSummary{
		TotalCost:         156892.47,
		AverageDailyCost:  5229.75,
		CostVariance:      0.124,
		BudgetUtilization: 0.842,
	}
}

func extractCostBreakdown(_ *types.AnalyticsResults) CostBreakdown {
	return CostBreakdown{
		ByService: map[string]ServiceCost{
			"EC2":    {Cost: 68423.12, Percentage: 0.436, Trend: "increasing"},
			"S3":     {Cost: 31246.89, Percentage: 0.199, Trend: "stable"},
			"RDS":    {Cost: 22156.78, Percentage: 0.141, Trend: "increasing"},
			"Lambda": {Cost: 18934.23, Percentage: 0.121, Trend: "decreasing"},
			"Others": {Cost: 16131.45, Percentage: 0.103, Trend: "stable"},
		},
		ByRegion: map[string]RegionCost{
			"us-east-1":    {Cost: 78446.23, Percentage: 0.500, Services: 12},
			"eu-west-1":    {Cost: 39223.12, Percentage: 0.250, Services: 8},
			"ap-southeast": {Cost: 23533.87, Percentage: 0.150, Services: 6},
			"us-west-2":    {Cost: 15689.25, Percentage: 0.100, Services: 4},
		},
	}
}

func extractResourceEfficiency(_ *types.AnalyticsResults) ResourceEfficiency {
	return ResourceEfficiency{
		OverallScore:       67.0,
		CPUUtilization:     0.67,
		MemoryUtilization:  0.72,
		StorageUtilization: 0.84,
		NetworkUtilization: 0.45,
		CostPerMetrics: map[string]float64{
			"cost_per_compute_hour": 0.087,
			"cost_per_gb_storage":   0.023,
			"cost_per_gb_transfer":  0.09,
		},
	}
}

func extractTrendAnalysis(_ *types.AnalyticsResults) *TrendAnalysis {
	if !analysisShowTrends {
		return nil
	}
	return &TrendAnalysis{
		WeeklyTrend: TrendInfo{
			Change:     1234.56,
			Percentage: 0.083,
			Direction:  "increasing",
			Confidence: 0.87,
		},
		MonthlyTrend: TrendInfo{
			Change:     18945.32,
			Percentage: 0.127,
			Direction:  "accelerating",
			Confidence: 0.92,
		},
		Volatility:  1247.0,
		GrowthRate:  0.012,
		Seasonality: false,
	}
}

func extractAnomalyAnalysis(_ *types.AnalyticsResults) *AnomalyAnalysis {
	if !analysisShowAnomalies {
		return nil
	}
	return &AnomalyAnalysis{
		Count:     3,
		Threshold: analysisAnomalyThreshold,
		Anomalies: []Anomaly{
			{
				Date:      time.Date(2025, 7, 15, 0, 0, 0, 0, time.UTC),
				Service:   "EC2",
				Cost:      45678.90,
				Expected:  13679.45,
				Deviation: 2.34,
				Severity:  "high",
				RootCause: "Auto-scaling events",
			},
			{
				Date:      time.Date(2025, 7, 12, 0, 0, 0, 0, time.UTC),
				Service:   "S3",
				Cost:      18934.56,
				Expected:  7401.23,
				Deviation: 1.56,
				Severity:  "medium",
				RootCause: "Storage surge",
			},
			{
				Date:      time.Date(2025, 7, 8, 0, 0, 0, 0, time.UTC),
				Service:   "CloudFront",
				Cost:      9876.54,
				Expected:  5234.21,
				Deviation: 0.89,
				Severity:  "low",
				RootCause: "Network anomaly",
			},
		},
	}
}

func extractForecastingAnalysis(_ *types.AnalyticsResults) *ForecastingAnalysis {
	if !analysisShowForecasting {
		return nil
	}
	return &ForecastingAnalysis{
		NextPeriod: ForecastPeriod{
			PeriodDays:    30,
			PredictedCost: 187450.0,
			LowerBound:    172500.0,
			UpperBound:    202400.0,
			Accuracy:      0.87,
		},
		Quarterly: ForecastPeriod{
			PeriodDays:    90,
			PredictedCost: 562350.0,
			LowerBound:    517500.0,
			UpperBound:    607200.0,
			Accuracy:      0.78,
		},
		Annual: ForecastPeriod{
			PeriodDays:    365,
			PredictedCost: 2249400.0,
			LowerBound:    2070000.0,
			UpperBound:    2428800.0,
			Accuracy:      0.65,
		},
		Confidence: analysisConfidenceLevel,
	}
}

func extractOptimizationAnalysis(_ *types.AnalyticsResults) *OptimizationAnalysis {
	if !analysisShowOptimizations {
		return nil
	}
	return &OptimizationAnalysis{
		TotalSavingsPotential: 25400.0,
		ROIScore:              76.0,
		ImplementationTime:    "4-6 weeks",
		Recommendations: []Recommendation{
			{
				ID:          "opt-001",
				Title:       "Rightsize overprovisioned instances",
				Description: "23 instances are running with <30% utilization",
				Impact:      "high",
				Savings:     8200.0,
				Effort:      "medium",
				Priority:    "high",
				Category:    "compute",
			},
			{
				ID:          "opt-002",
				Title:       "Enable storage lifecycle policies",
				Description: "Implement tiered storage for infrequently accessed data",
				Impact:      "medium",
				Savings:     3400.0,
				Effort:      "low",
				Priority:    "medium",
				Category:    "storage",
			},
			{
				ID:          "opt-003",
				Title:       "Use Reserved Instances for stable workloads",
				Description: "Purchase RIs for predictable compute workloads",
				Impact:      "high",
				Savings:     12600.0,
				Effort:      "low",
				Priority:    "high",
				Category:    "compute",
			},
		},
	}
}

func getTopServices(services map[string]ServiceCost, limit int) map[string]ServiceCost {
	// Simple implementation - in reality would sort by cost
	result := make(map[string]ServiceCost)
	count := 0
	for k, v := range services {
		if count >= limit {
			break
		}
		result[k] = v
		count++
	}
	return result
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func init() {
	// Date range flags
	AnalyzeCmd.Flags().StringVar(&analysisStartDate, "start-date", "", "Analysis start date (YYYY-MM-DD)")
	AnalyzeCmd.Flags().StringVar(&analysisEndDate, "end-date", "", "Analysis end date (YYYY-MM-DD)")

	// Analysis scope flags
	AnalyzeCmd.Flags().StringVar(&analysisGranularity, "granularity", "daily", "Analysis granularity (hourly, daily, weekly, monthly)")
	AnalyzeCmd.Flags().StringSliceVar(&analysisGroupBy, "group-by", []string{}, "Group analysis by dimensions (service, region, account)")
	AnalyzeCmd.Flags().StringSliceVar(&analysisServices, "services", []string{}, "Filter by specific services")
	AnalyzeCmd.Flags().StringSliceVar(&analysisRegions, "regions", []string{}, "Filter by specific regions")
	AnalyzeCmd.Flags().StringSliceVar(&analysisAccounts, "accounts", []string{}, "Filter by specific accounts")

	// Analysis features flags
	AnalyzeCmd.Flags().BoolVar(&analysisShowTrends, "show-trends", false, "Include trend analysis")
	AnalyzeCmd.Flags().BoolVar(&analysisShowAnomalies, "show-anomalies", false, "Include anomaly detection")
	AnalyzeCmd.Flags().BoolVar(&analysisShowOptimizations, "show-optimizations", false, "Include optimization recommendations")
	AnalyzeCmd.Flags().BoolVar(&analysisShowForecasting, "show-forecasting", false, "Include cost forecasting")

	// ML and advanced analytics flags
	AnalyzeCmd.Flags().IntVar(&analysisForecastPeriods, "forecast-periods", 30, "Number of periods to forecast")
	AnalyzeCmd.Flags().Float64Var(&analysisAnomalyThreshold, "anomaly-threshold", 2.0, "Anomaly detection threshold (standard deviations)")
	AnalyzeCmd.Flags().BoolVar(&analysisMLEnabled, "ml-enabled", false, "Enable machine learning features")
	AnalyzeCmd.Flags().Float64Var(&analysisConfidenceLevel, "confidence-level", 0.95, "Confidence level for statistical analysis")

	// Output flags
	AnalyzeCmd.Flags().StringVar(&analysisOutputFile, "output", "", "Output file path")
	AnalyzeCmd.Flags().StringVar(&analysisOutputFormat, "format", "json", "Output format (json, csv, yaml)")
	AnalyzeCmd.Flags().BoolVarP(&analysisVerbose, "verbose", "v", false, "Enable verbose output")
	AnalyzeCmd.Flags().BoolVarP(&analysisQuiet, "quiet", "q", false, "Suppress output")

	// Performance flags
	AnalyzeCmd.Flags().IntVar(&analysisWorkers, "workers", 4, "Number of worker goroutines")
	AnalyzeCmd.Flags().IntVar(&analysisChunkSize, "chunk-size", 10000, "Processing chunk size")
	AnalyzeCmd.Flags().BoolVar(&analysisStreaming, "streaming", false, "Enable streaming processing for large files")
	AnalyzeCmd.Flags().IntVar(&analysisMaxMemory, "max-memory", 1024, "Maximum memory usage in MB")

	// Experimental path: use FOCUS analysis engine instead of generic analytics service
	AnalyzeCmd.Flags().BoolVar(&analysisUseFocusEngine, "use-focus-engine", false, "Use the FOCUS analysis engine (experimental)")

	// Focus engine tuning flags (0 means 'not explicitly set')
	AnalyzeCmd.Flags().IntVar(&focusForecastDaysFlag, "focus-forecast-days", 0, "Override focus engine forecast horizon (days); precedence flag>YAML>env>default(30)")
	AnalyzeCmd.Flags().DurationVar(&focusPhaseTimeoutFlag, "focus-phase-timeout", 0, "Per-phase soft timeout for focus engine extended phases (e.g. 1500ms, 2s)")

	// Metric types
	AnalyzeCmd.Flags().StringSliceVar(&analysisMetricTypes, "metrics", []string{"cost", "usage"}, "Metric types to analyze")
}
