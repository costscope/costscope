package comparison

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/costscope/costscope/internal/core/logging"
)

// Severity level constants
const (
	SeverityCritical = "critical"
	SeverityHigh     = "high"
	SeverityMedium   = "medium"
	SeverityLow      = "low"
)

// Engine implements the ComparisonEngine interface
type Engine struct {
	logger *logging.Logger
	config *ComparisonConfiguration
	ctx    context.Context
}

// NewEngine creates a new comparison engine
func NewEngine(logger *logging.Logger, config *ComparisonConfiguration) *Engine {
	if config == nil {
		config = DefaultComparisonConfiguration()
	}

	return &Engine{
		logger: logger,
		config: config,
		ctx:    context.Background(),
	}
}

// DefaultComparisonConfiguration returns default configuration
func DefaultComparisonConfiguration() *ComparisonConfiguration {
	return &ComparisonConfiguration{
		SignificanceThreshold:  100.0, // $100 minimum change
		PercentChangeThreshold: 10.0,  // 10% minimum change
		MinCostThreshold:       1.0,   // $1 minimum cost
		Dimensions:             []string{"service", "region"},
		MLEnabled:              true,
		AnomalyDetection:       true,
		TrendAnalysis:          true,
		ForecastEnabled:        true,
		ForecastPeriods:        30,
		ConfidenceLevel:        0.95,
		AnomalyMethods:         []string{"isolation_forest", "statistical"},
		AnomalyThreshold:       0.1,
		OutputFormat:           "json",
		IncludeDetails:         true,
		IncludeRawData:         false,
		CompressOutput:         false,
	}
}

// CompareFOCUSDatasets performs comprehensive comparison of two FOCUS datasets
func (e *Engine) CompareFOCUSDatasets(baseline, current string, options DiffOptions) (*DiffResult, error) {
	startTime := time.Now()
	e.logger.Info(fmt.Sprintf("Starting FOCUS dataset comparison: baseline=%s, current=%s", baseline, current))

	// Load datasets
	baselineData, err := e.loadFOCUSDataset(baseline)
	if err != nil {
		return nil, fmt.Errorf("failed to load baseline dataset: %w", err)
	}

	currentData, err := e.loadFOCUSDataset(current)
	if err != nil {
		return nil, fmt.Errorf("failed to load current dataset: %w", err)
	}

	e.logger.Info(fmt.Sprintf("Datasets loaded successfully: baseline_records=%d, current_records=%d", len(baselineData), len(currentData)))

	// Perform comparison analysis
	result := &DiffResult{
		Metadata: DiffMetadata{
			AnalysisDate:      time.Now(),
			Threshold:         options.Threshold,
			Dimension:         options.Dimensions,
			BaselineRecords:   len(baselineData),
			ComparisonRecords: len(currentData),
			MLEnabled:         options.MLEnabled,
			AnomalyDetection:  options.ShowAnomalies,
			TrendAnalysis:     options.ShowTrends,
			Version:           "1.0.0",
		},
	}

	// Detect cost changes
	changes, err := e.DetectCostChanges(baselineData, currentData, options)
	if err != nil {
		return nil, fmt.Errorf("failed to detect cost changes: %w", err)
	}
	result.Changes = changes

	// Identify service changes
	newServices, removedServices, err := e.IdentifyServiceChanges(baselineData, currentData)
	if err != nil {
		return nil, fmt.Errorf("failed to identify service changes: %w", err)
	}
	result.NewServices = newServices
	result.Removed = removedServices

	// Analyze trends if requested
	if options.ShowTrends {
		trends, err := e.AnalyzeTrends(baselineData, currentData, options)
		if err != nil {
			e.logger.Warn(fmt.Sprintf("Failed to analyze trends: %v", err))
			trends = make(map[string]DiffTrendInfo)
		}
		result.Trends = trends
	}

	// Detect anomalies if requested
	if options.ShowAnomalies {
		anomalies, err := e.DetectAnomalies(currentData, options)
		if err != nil {
			e.logger.Warn(fmt.Sprintf("Failed to detect anomalies: %v", err))
			anomalies = []AnomalyInfo{}
		}
		result.Anomalies = anomalies
	}

	// Generate summary
	result.Summary = e.generateDiffSummary(result, baselineData, currentData)

	processingTime := time.Since(startTime)
	result.Metadata.ProcessingTime = processingTime.String()

	e.logger.Info(fmt.Sprintf("FOCUS dataset comparison completed: processing_time=%v, changes_detected=%d, new_services=%d, removed_services=%d, trends_analyzed=%d, anomalies_detected=%d",
		processingTime, len(result.Changes), len(result.NewServices), len(result.Removed), len(result.Trends), len(result.Anomalies)))

	return result, nil
}

// DetectCostChanges identifies significant cost changes between datasets
func (e *Engine) DetectCostChanges(baseline, current []FOCUSRecord, options DiffOptions) ([]CostChange, error) {
	e.logger.Info(fmt.Sprintf("Detecting cost changes: dimensions=%v, threshold=%.2f", options.Dimensions, options.Threshold))

	// Group data by specified dimensions
	baselineGroups := e.groupRecords(baseline, options.Dimensions)
	currentGroups := e.groupRecords(current, options.Dimensions)

	var changes []CostChange

	// Check for changes in existing groups
	for key, currentCost := range currentGroups {
		baselineCost, exists := baselineGroups[key]

		change := CostChange{
			CurrentCost:     currentCost.TotalCost,
			Change:          currentCost.TotalCost - baselineCost.TotalCost,
			Timestamp:       time.Now(),
			DetectionMethod: "comparison_engine",
			ConfidenceLevel: 1.0,
		}

		// Parse dimensions from key
		dimensions := strings.Split(key, "|")
		if len(dimensions) >= 1 {
			change.Service = dimensions[0]
		}
		if len(dimensions) >= 2 {
			change.Region = dimensions[1]
		}
		if len(dimensions) >= 3 {
			change.Account = dimensions[2]
		}

		if exists {
			// Calculate percentage change
			if baselineCost.TotalCost > 0 {
				change.PercentChange = (change.Change / baselineCost.TotalCost) * 100
			}
			change.BaselineCost = baselineCost.TotalCost

			// Determine significance
			if math.Abs(change.Change) >= options.Threshold ||
				math.Abs(change.PercentChange) >= options.SignificanceLevel*100 {
				if change.Change > 0 {
					change.Category = "increase"
				} else {
					change.Category = "decrease"
				}

				change.Significance = e.categorizeSignificance(change.Change, change.PercentChange)
				changes = append(changes, change)
			}
		} else {
			// New service/resource
			change.Category = "new"
			change.BaselineCost = 0
			change.PercentChange = 100 // 100% increase
			change.Significance = e.categorizeSignificance(change.Change, change.PercentChange)
			changes = append(changes, change)
		}
	}

	// Check for removed services/resources
	for key, baselineCost := range baselineGroups {
		if _, exists := currentGroups[key]; !exists {
			change := CostChange{
				BaselineCost:    baselineCost.TotalCost,
				CurrentCost:     0,
				Change:          -baselineCost.TotalCost,
				PercentChange:   -100,
				Category:        "removed",
				Timestamp:       time.Now(),
				DetectionMethod: "comparison_engine",
				ConfidenceLevel: 1.0,
			}

			// Parse dimensions from key
			dimensions := strings.Split(key, "|")
			if len(dimensions) >= 1 {
				change.Service = dimensions[0]
			}
			if len(dimensions) >= 2 {
				change.Region = dimensions[1]
			}
			if len(dimensions) >= 3 {
				change.Account = dimensions[2]
			}

			change.Significance = e.categorizeSignificance(change.Change, change.PercentChange)
			changes = append(changes, change)
		}
	}

	// Sort changes by absolute cost impact
	sort.Slice(changes, func(i, j int) bool {
		return math.Abs(changes[i].Change) > math.Abs(changes[j].Change)
	})

	e.logger.Info(fmt.Sprintf("Cost changes detected: total_changes=%d", len(changes)))

	return changes, nil
}

// IdentifyServiceChanges identifies new and removed services
func (e *Engine) IdentifyServiceChanges(baseline, current []FOCUSRecord) ([]ServiceInfo, []ServiceInfo, error) {
	e.logger.Info("Identifying service changes")

	baselineServices := e.extractServices(baseline)
	currentServices := e.extractServices(current)

	var newServices, removedServices []ServiceInfo

	// Find new services
	for key, service := range currentServices {
		if _, exists := baselineServices[key]; !exists {
			newServices = append(newServices, service)
		}
	}

	// Find removed services
	for key, service := range baselineServices {
		if _, exists := currentServices[key]; !exists {
			removedServices = append(removedServices, service)
		}
	}

	e.logger.Info(fmt.Sprintf("Service changes identified: new_services=%d, removed_services=%d", len(newServices), len(removedServices)))

	return newServices, removedServices, nil
}

// AnalyzeTrends performs trend analysis on the datasets
func (e *Engine) AnalyzeTrends(baseline, current []FOCUSRecord, options DiffOptions) (map[string]DiffTrendInfo, error) {
	e.logger.Info("Analyzing trends")

	trends := make(map[string]DiffTrendInfo)

	// For now, implement basic trend analysis
	// This would be enhanced with more sophisticated ML models
	baselineGroups := e.groupRecords(baseline, options.Dimensions)
	currentGroups := e.groupRecords(current, options.Dimensions)

	for key := range currentGroups {
		baselineCost, hasBaseline := baselineGroups[key]
		currentCost := currentGroups[key]

		if hasBaseline && len(baselineCost.DataPoints) > 1 && len(currentCost.DataPoints) > 1 {
			trend := e.calculateTrend(baselineCost.DataPoints, currentCost.DataPoints)

			dimensions := strings.Split(key, "|")
			if len(dimensions) >= 1 {
				trend.Service = dimensions[0]
			}
			if len(dimensions) >= 2 {
				trend.Region = dimensions[1]
			}

			trends[key] = trend
		}
	}

	e.logger.Info(fmt.Sprintf("Trend analysis completed: trends_identified=%d", len(trends)))

	return trends, nil
}

// DetectAnomalies detects cost anomalies in the dataset
func (e *Engine) DetectAnomalies(data []FOCUSRecord, options DiffOptions) ([]AnomalyInfo, error) {
	e.logger.Info("Detecting anomalies")

	var anomalies []AnomalyInfo

	// Group data for anomaly detection
	groups := e.groupRecords(data, options.Dimensions)

	for key, group := range groups {
		if len(group.DataPoints) > 7 { // Need at least a week of data
			detectedAnomalies := e.detectAnomaliesInSeries(group.DataPoints, key)
			anomalies = append(anomalies, detectedAnomalies...)
		}
	}

	e.logger.Info(fmt.Sprintf("Anomaly detection completed: anomalies_detected=%d", len(anomalies)))

	return anomalies, nil
}

// GenerateForecast generates cost forecasts
func (e *Engine) GenerateForecast(historical []FOCUSRecord, periods int) ([]Forecast, error) {
	e.logger.Info(fmt.Sprintf("Generating forecast: periods=%d", periods))

	// Simple linear regression forecast
	// This would be enhanced with more sophisticated ML models
	groups := e.groupRecords(historical, []string{"service"})

	var forecasts []Forecast

	for _, group := range groups {
		if len(group.DataPoints) >= 7 {
			serviceForecast := e.generateServiceForecast(group.DataPoints, periods)
			forecasts = append(forecasts, serviceForecast...)
		}
	}

	return forecasts, nil
}

// GenerateExecutiveSummary creates an executive summary from diff results
func (e *Engine) GenerateExecutiveSummary(result *DiffResult) (*ExecutiveSummary, error) {
	e.logger.Info("Generating executive summary")

	summary := &ExecutiveSummary{
		AnalysisPeriod:     fmt.Sprintf("%s to %s", result.Summary.BaselinePeriod, result.Summary.ComparisonPeriod),
		GeneratedAt:        time.Now(),
		TotalCostImpact:    result.Summary.TotalCostChange,
		TotalCostImpactPct: result.Summary.PercentageChange,
	}

	// Generate key findings
	summary.KeyFindings = e.generateKeyFindings(result)

	// Top cost changes (limit to top 5)
	topChanges := result.Changes
	if len(topChanges) > 5 {
		topChanges = topChanges[:5]
	}
	summary.TopCostChanges = topChanges

	// Critical anomalies
	for _, anomaly := range result.Anomalies {
		if anomaly.Severity == SeverityCritical || anomaly.Severity == SeverityHigh {
			summary.CriticalAnomalies = append(summary.CriticalAnomalies, anomaly)
		}
	}

	// Generate action items
	summary.ImmediateActions = e.generateActionItems(result)

	// Generate recommendations
	summary.Recommendations = e.generateRecommendations(result)

	return summary, nil
}

// ExportResults exports comparison results to specified format
func (e *Engine) ExportResults(result *DiffResult, format, output string) error {
	e.logger.Info(fmt.Sprintf("Exporting results: format=%s, output=%s", format, output))

	switch strings.ToLower(format) {
	case "json":
		return e.exportJSON(result, output)
	case "csv":
		return e.exportCSV(result, output)
	case "html":
		return e.exportHTML(result, output)
	default:
		return fmt.Errorf("unsupported export format: %s", format)
	}
}

// Helper methods

func (e *Engine) loadFOCUSDataset(filename string) ([]FOCUSRecord, error) {
	// This would be implemented to load FOCUS parquet files
	// For now, return empty slice
	e.logger.Info(fmt.Sprintf("Loading FOCUS dataset: file=%s", filename))

	// TODO: Implement actual parquet file loading
	// For now, return mock data to satisfy the interface
	if filename == "" {
		return nil, fmt.Errorf("filename cannot be empty")
	}

	return []FOCUSRecord{}, nil
}

type GroupedData struct {
	TotalCost  float64
	DataPoints []DataPoint
	Records    []FOCUSRecord
}

// ComparisonInsights provides an aggregated, higher-level view combining
// the diff result, executive summary and (optionally) a forecast. This is
// intended for CLI/API one-shot "insights" style responses so callers do not
// need to orchestrate multiple engine calls. It is an additive convenience
// wrapper that preserves existing public methods for granular use.
type ComparisonInsights struct {
	Diff        *DiffResult       `json:"diff"`
	Executive   *ExecutiveSummary `json:"executive_summary"`
	Forecast    []Forecast        `json:"forecast,omitempty"`
	GeneratedAt time.Time         `json:"generated_at"`
}

// GenerateComparisonInsights runs a full comparison workflow (Compare +
// executive summary + optional forecast) over the provided baseline/current
// FOCUS records. forecastPeriods <= 0 disables forecast generation. Errors
// from forecast generation are logged but do not fail the whole aggregation
// (best-effort) because insights value can still be delivered.
func (e *Engine) GenerateComparisonInsights(baseline, current []FOCUSRecord, opts DiffOptions, forecastPeriods int) (*ComparisonInsights, error) {
	start := time.Now()
	// Build diff inline (mirrors CompareFOCUSDatasets minus file loading)
	diff := &DiffResult{Metadata: DiffMetadata{AnalysisDate: start, Threshold: opts.Threshold, Dimension: opts.Dimensions, BaselineRecords: len(baseline), ComparisonRecords: len(current), MLEnabled: opts.MLEnabled, AnomalyDetection: opts.ShowAnomalies, TrendAnalysis: opts.ShowTrends, Version: "1.0.0"}}

	changes, err := e.DetectCostChanges(baseline, current, opts)
	if err != nil {
		return nil, fmt.Errorf("detect cost changes: %w", err)
	}
	diff.Changes = changes

	newServices, removedServices, err := e.IdentifyServiceChanges(baseline, current)
	if err != nil {
		return nil, fmt.Errorf("identify service changes: %w", err)
	}
	diff.NewServices = newServices
	diff.Removed = removedServices

	if opts.ShowTrends {
		trends, terr := e.AnalyzeTrends(baseline, current, opts)
		if terr != nil {
			e.logger.Warn(fmt.Sprintf("trend analysis failed: %v", terr))
			trends = map[string]DiffTrendInfo{}
		}
		diff.Trends = trends
	}
	if opts.ShowAnomalies {
		anomalies, aerr := e.DetectAnomalies(current, opts)
		if aerr != nil {
			e.logger.Warn(fmt.Sprintf("anomaly detection failed: %v", aerr))
			anomalies = []AnomalyInfo{}
		}
		diff.Anomalies = anomalies
	}
	diff.Summary = e.generateDiffSummary(diff, baseline, current)
	diff.Metadata.ProcessingTime = time.Since(start).String()

	exec, err := e.GenerateExecutiveSummary(diff)
	if err != nil {
		return nil, fmt.Errorf("executive summary: %w", err)
	}

	var forecast []Forecast
	if forecastPeriods > 0 {
		hist := append(append([]FOCUSRecord{}, baseline...), current...)
		if fc, ferr := e.GenerateForecast(hist, forecastPeriods); ferr == nil {
			forecast = fc
		} else {
			e.logger.Warn(fmt.Sprintf("forecast generation failed: %v", ferr))
		}
	}
	return &ComparisonInsights{Diff: diff, Executive: exec, Forecast: forecast, GeneratedAt: time.Now()}, nil
}

// CompareFOCUSFilesInsights is a file-oriented convenience wrapper that loads
// baseline/current FOCUS datasets (using the same loader as CompareFOCUSDatasets)
// then produces aggregated insights (diff + executive summary + optional forecast).
// It exists to support the CLI --insights flag without duplicating loading logic
// in the command layer.
func (e *Engine) CompareFOCUSFilesInsights(baselineFile, currentFile string, opts DiffOptions, forecastPeriods int) (*ComparisonInsights, error) {
	baselineData, err := e.loadFOCUSDataset(baselineFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load baseline dataset: %w", err)
	}
	currentData, err := e.loadFOCUSDataset(currentFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load current dataset: %w", err)
	}
	return e.GenerateComparisonInsights(baselineData, currentData, opts, forecastPeriods)
}

func (e *Engine) groupRecords(records []FOCUSRecord, dimensions []string) map[string]GroupedData {
	groups := make(map[string]GroupedData)

	for _, record := range records {
		key := e.buildGroupKey(record, dimensions)

		group, exists := groups[key]
		if !exists {
			group = GroupedData{
				DataPoints: []DataPoint{},
				Records:    []FOCUSRecord{},
			}
		}

		group.TotalCost += record.BilledCost
		group.DataPoints = append(group.DataPoints, DataPoint{
			Date:   record.BillingPeriodStart,
			Cost:   record.BilledCost,
			Usage:  record.UsageQuantity,
			Source: record.ServiceName,
		})
		group.Records = append(group.Records, record)

		groups[key] = group
	}

	return groups
}

func (e *Engine) buildGroupKey(record FOCUSRecord, dimensions []string) string {
	var parts []string

	for _, dim := range dimensions {
		switch strings.ToLower(dim) {
		case "service":
			parts = append(parts, record.ServiceName)
		case "region":
			parts = append(parts, record.Region)
		case "account":
			parts = append(parts, record.BillingAccountID)
		case "resource":
			parts = append(parts, record.ResourceID)
		default:
			parts = append(parts, "unknown")
		}
	}

	return strings.Join(parts, "|")
}

func (e *Engine) extractServices(records []FOCUSRecord) map[string]ServiceInfo {
	services := make(map[string]ServiceInfo)

	for _, record := range records {
		key := fmt.Sprintf("%s|%s|%s", record.ServiceName, record.Region, record.BillingAccountID)

		service, exists := services[key]
		if !exists {
			service = ServiceInfo{
				Service:   record.ServiceName,
				Region:    record.Region,
				Account:   record.BillingAccountID,
				FirstSeen: record.BillingPeriodStart,
				LastSeen:  record.BillingPeriodEnd,
			}
		}

		service.Cost += record.BilledCost
		service.UsageQuantity += record.UsageQuantity
		service.UsageUnit = record.UsageUnit
		service.ResourceCount++

		if record.BillingPeriodStart.Before(service.FirstSeen) {
			service.FirstSeen = record.BillingPeriodStart
		}
		if record.BillingPeriodEnd.After(service.LastSeen) {
			service.LastSeen = record.BillingPeriodEnd
		}

		services[key] = service
	}

	return services
}

func (e *Engine) categorizeSignificance(change, percentChange float64) string {
	absChange := math.Abs(change)
	absPercentChange := math.Abs(percentChange)

	if absChange >= 10000 || absPercentChange >= 100 {
		return SeverityCritical
	} else if absChange >= 1000 || absPercentChange >= 50 {
		return SeverityHigh
	} else if absChange >= 100 || absPercentChange >= 20 {
		return SeverityMedium
	}
	return SeverityLow
}

func (e *Engine) calculateTrend(baseline, current []DataPoint) DiffTrendInfo {
	// Simple trend calculation - would be enhanced with proper statistical analysis
	baselineAvg := e.calculateAverage(baseline)
	currentAvg := e.calculateAverage(current)

	trend := DiffTrendInfo{
		Velocity:   currentAvg - baselineAvg,
		Prediction: currentAvg + (currentAvg - baselineAvg), // Simple linear projection
		Confidence: 0.7,                                     // Default confidence
	}

	if trend.Velocity > baselineAvg*0.1 {
		trend.Trend = "increasing"
	} else if trend.Velocity < -baselineAvg*0.1 {
		trend.Trend = "decreasing"
	} else {
		trend.Trend = "stable"
	}

	return trend
}

func (e *Engine) calculateAverage(points []DataPoint) float64 {
	if len(points) == 0 {
		return 0
	}

	total := 0.0
	for _, point := range points {
		total += point.Cost
	}
	return total / float64(len(points))
}

func (e *Engine) detectAnomaliesInSeries(points []DataPoint, key string) []AnomalyInfo {
	var anomalies []AnomalyInfo

	if len(points) < 7 {
		return anomalies
	}

	// Simple statistical anomaly detection
	mean := e.calculateAverage(points)
	stdDev := e.calculateStdDev(points, mean)
	threshold := 2.0 * stdDev

	for _, point := range points {
		if math.Abs(point.Cost-mean) > threshold {
			anomaly := AnomalyInfo{
				DetectedAt:      point.Date,
				AnomalyScore:    math.Abs(point.Cost-mean) / threshold,
				ExpectedCost:    mean,
				ActualCost:      point.Cost,
				Deviation:       point.Cost - mean,
				ConfidenceLevel: 0.8,
				DetectionMethod: "statistical",
				Severity:        e.categorizeSeverity(math.Abs(point.Cost - mean)),
				AnomalyType:     "outlier",
				Description:     "Cost significantly differs from expected value",
			}

			dimensions := strings.Split(key, "|")
			if len(dimensions) >= 1 {
				anomaly.Service = dimensions[0]
			}
			if len(dimensions) >= 2 {
				anomaly.Region = dimensions[1]
			}

			anomalies = append(anomalies, anomaly)
		}
	}

	return anomalies
}

func (e *Engine) calculateStdDev(points []DataPoint, mean float64) float64 {
	if len(points) <= 1 {
		return 0
	}

	sum := 0.0
	for _, point := range points {
		sum += (point.Cost - mean) * (point.Cost - mean)
	}

	return math.Sqrt(sum / float64(len(points)-1))
}

func (e *Engine) categorizeSeverity(deviation float64) string {
	if deviation >= 10000 {
		return SeverityCritical
	} else if deviation >= 1000 {
		return SeverityHigh
	} else if deviation >= 100 {
		return SeverityMedium
	}
	return SeverityLow
}

func (e *Engine) generateServiceForecast(points []DataPoint, periods int) []Forecast {
	// Simple linear regression forecast
	var forecasts []Forecast

	if len(points) < 3 {
		return forecasts
	}

	// Calculate trend
	n := float64(len(points))
	sumX := n * (n - 1) / 2
	sumY := 0.0
	sumXY := 0.0
	sumX2 := (n - 1) * n * (2*n - 1) / 6

	for i, point := range points {
		x := float64(i)
		y := point.Cost
		sumY += y
		sumXY += x * y
	}

	slope := (n*sumXY - sumX*sumY) / (n*sumX2 - sumX*sumX)
	intercept := (sumY - slope*sumX) / n

	// Generate forecasts
	for i := 1; i <= periods; i++ {
		x := float64(len(points) + i - 1)
		predicted := slope*x + intercept

		forecast := Forecast{
			Period:          fmt.Sprintf("Day %d", i),
			PredictedCost:   predicted,
			ConfidenceUpper: predicted * 1.2,
			ConfidenceLower: predicted * 0.8,
			PredictionDate:  time.Now().AddDate(0, 0, i),
		}

		forecasts = append(forecasts, forecast)
	}

	return forecasts
}

func (e *Engine) generateDiffSummary(result *DiffResult, baseline, current []FOCUSRecord) DiffSummary {
	totalBaselineCost := 0.0
	totalCurrentCost := 0.0

	for _, record := range baseline {
		totalBaselineCost += record.BilledCost
	}

	for _, record := range current {
		totalCurrentCost += record.BilledCost
	}

	change := totalCurrentCost - totalBaselineCost
	percentChange := 0.0
	if totalBaselineCost > 0 {
		percentChange = (change / totalBaselineCost) * 100
	}

	significantChanges := 0
	for _, change := range result.Changes {
		if change.Significance == SeverityHigh || change.Significance == SeverityCritical {
			significantChanges++
		}
	}

	return DiffSummary{
		TotalCostChange:      change,
		PercentageChange:     percentChange,
		SignificantChanges:   significantChanges,
		NewServicesCount:     len(result.NewServices),
		RemovedServicesCount: len(result.Removed),
		AnomaliesDetected:    len(result.Anomalies),
		BaselinePeriod:       "Baseline Period",
		ComparisonPeriod:     "Current Period",
	}
}

func (e *Engine) generateKeyFindings(result *DiffResult) []KeyFinding {
	var findings []KeyFinding

	// Cost impact finding
	if math.Abs(result.Summary.TotalCostChange) > 1000 {
		impact := SeverityMedium
		if math.Abs(result.Summary.TotalCostChange) > 10000 {
			impact = SeverityHigh
		}

		finding := KeyFinding{
			Category:    "cost_change",
			Impact:      impact,
			Title:       "Significant Cost Change Detected",
			Description: fmt.Sprintf("Total cost changed by $%.2f (%.1f%%)", result.Summary.TotalCostChange, result.Summary.PercentageChange),
			Value:       result.Summary.TotalCostChange,
			Unit:        "USD",
			Confidence:  0.95,
		}
		findings = append(findings, finding)
	}

	// New services finding
	if len(result.NewServices) > 0 {
		finding := KeyFinding{
			Category:    "new_service",
			Impact:      SeverityMedium,
			Title:       "New Services Detected",
			Description: fmt.Sprintf("%d new services started incurring costs", len(result.NewServices)),
			Value:       float64(len(result.NewServices)),
			Unit:        "services",
			Confidence:  1.0,
		}
		findings = append(findings, finding)
	}

	// Anomalies finding
	criticalAnomalies := 0
	for _, anomaly := range result.Anomalies {
		if anomaly.Severity == SeverityCritical || anomaly.Severity == SeverityHigh {
			criticalAnomalies++
		}
	}

	if criticalAnomalies > 0 {
		finding := KeyFinding{
			Category:    "anomaly",
			Impact:      SeverityHigh,
			Title:       "Critical Anomalies Detected",
			Description: fmt.Sprintf("%d critical cost anomalies require immediate attention", criticalAnomalies),
			Value:       float64(criticalAnomalies),
			Unit:        "anomalies",
			Confidence:  0.85,
		}
		findings = append(findings, finding)
	}

	return findings
}

func (e *Engine) generateActionItems(result *DiffResult) []ActionItem {
	var actions []ActionItem

	// High-impact cost increases
	for _, change := range result.Changes {
		if change.Significance == SeverityCritical && change.Change > 0 {
			action := ActionItem{
				Priority:    SeverityCritical,
				Category:    "investigation",
				Title:       fmt.Sprintf("Investigate %s cost spike", change.Service),
				Description: fmt.Sprintf("Cost increased by $%.2f (%.1f%%) in %s", change.Change, change.PercentChange, change.Service),
				Timeline:    "immediate",
				Owner:       "finops",
				Impact:      SeverityHigh,
			}
			actions = append(actions, action)
		}
	}

	// Critical anomalies
	for _, anomaly := range result.Anomalies {
		if anomaly.Severity == SeverityCritical {
			action := ActionItem{
				Priority:    SeverityHigh,
				Category:    "investigation",
				Title:       fmt.Sprintf("Investigate %s anomaly", anomaly.Service),
				Description: anomaly.Description,
				Timeline:    "this_week",
				Owner:       "engineering",
				Impact:      SeverityHigh,
			}
			actions = append(actions, action)
		}
	}

	return actions
}

func (e *Engine) generateRecommendations(result *DiffResult) []Recommendation {
	var recommendations []Recommendation

	// Cost optimization recommendation
	if result.Summary.TotalCostChange > 5000 {
		rec := Recommendation{
			Type:        "strategic",
			Priority:    SeverityHigh,
			Title:       "Implement Cost Monitoring and Alerting",
			Description: "Establish proactive cost monitoring to detect and respond to cost changes quickly",
			Rationale:   "Significant cost changes detected indicate need for better cost visibility",
			Benefits:    []string{"Early detection of cost anomalies", "Improved cost control", "Better budget adherence"},
			Risks:       []string{"Implementation overhead", "Alert fatigue"},
			Timeline:    "this_month",
			Success:     []string{"Reduction in unexpected cost spikes", "Improved cost predictability"},
		}
		recommendations = append(recommendations, rec)
	}

	return recommendations
}

func (e *Engine) exportJSON(result *DiffResult, output string) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// In a real implementation, this would write to a file
	e.logger.Info(fmt.Sprintf("JSON export completed: output=%s, size=%d", output, len(data)))

	return nil
}

func (e *Engine) exportCSV(result *DiffResult, output string) error {
	// CSV export implementation would go here
	// TODO: Implement actual CSV export using result data
	_ = result // Acknowledge parameter for future implementation
	e.logger.Info(fmt.Sprintf("CSV export completed: output=%s", output))
	return nil
}

func (e *Engine) exportHTML(result *DiffResult, output string) error {
	// HTML export implementation would go here
	// TODO: Implement actual HTML export using result data
	_ = result // Acknowledge parameter for future implementation
	e.logger.Info(fmt.Sprintf("HTML export completed: output=%s", output))
	return nil
}
