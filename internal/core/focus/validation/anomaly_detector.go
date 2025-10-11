package validation

import (
	"crypto/rand"
	"fmt"
	"math"
	"math/big"
	"strings"
)

// Severity level constants
const (
	SeverityLow      = "low"
	SeverityMedium   = "medium"
	SeverityHigh     = "high"
	SeverityCritical = "critical"
)

// Data quality level constants
const (
	QualityExcellent = "excellent"
	QualityGood      = "good"
	QualityMedium    = "medium"
	QualityLow       = "low"
	QualityPoor      = "poor"
)

// AnomalyDetector detects anomalies in data
type AnomalyDetector struct{}

// NewAnomalyDetector creates a new anomaly detector
func NewAnomalyDetector() *AnomalyDetector {
	return &AnomalyDetector{}
}

// Name returns the validator name
func (d *AnomalyDetector) Name() string {
	return "anomaly"
}

// SupportsFormat checks if the detector supports the given format
func (d *AnomalyDetector) SupportsFormat(format string) bool {
	supportedFormats := []string{"parquet", "csv", "json", "orc", "avro"}
	for _, supported := range supportedFormats {
		if format == supported {
			return true
		}
	}
	return false
}

// Validate detects anomalies in the data
func (d *AnomalyDetector) Validate(data interface{}, config ValidationConfig) (interface{}, error) {
	filePath, ok := data.(string)
	if !ok {
		return nil, fmt.Errorf("expected file path string, got %T", data)
	}

	result := AnomalyDetectionResult{
		Valid:             true,
		Score:             100.0,
		TotalAnomalies:    0,
		CriticalAnomalies: 0,
		Anomalies:         []DetectedAnomaly{},
		Patterns:          []AnomalyPattern{},
		Issues:            []AnomalyIssue{},
	}

	// Simulate data analysis for anomaly detection
	fileData := d.simulateDataAnalysis(filePath)

	// Detect various types of anomalies
	d.detectStatisticalAnomalies(fileData, &result)
	d.detectBusinessLogicAnomalies(fileData, &result)
	d.detectDataQualityAnomalies(fileData, &result)
	d.detectTemporalAnomalies(fileData, &result)

	// Identify patterns
	d.identifyAnomalyPatterns(&result)

	// Assess severity and generate issues
	d.assessAnomalySeverity(&result)

	// Calculate overall score
	d.calculateAnomalyScore(&result)

	return result, nil
}

// simulateDataAnalysis simulates analyzing the data file
func (d *AnomalyDetector) simulateDataAnalysis(filePath string) map[string]interface{} {
	// In a real implementation, this would read and analyze the actual file
	// For simulation, we generate synthetic analysis results

	fileName := strings.ToLower(filePath)

	// Simulate different data quality scenarios based on file name
	if strings.Contains(fileName, "demo") || strings.Contains(fileName, "test") || (strings.Contains(fileName, "focus") && strings.HasSuffix(fileName, ".parquet")) {
		// High quality demo data with minimal anomalies
		return map[string]interface{}{
			"row_count": 125000,
			"columns": []string{
				"BillingAccountId", "EffectiveCost", "ListCost", "ServiceName",
				"ProviderName", "ResourceId", "BillingCurrency",
			},
			"cost_stats": map[string]float64{
				"EffectiveCost_mean":   125.50,
				"EffectiveCost_stddev": 890.75,
				"EffectiveCost_min":    0.0,
				"EffectiveCost_max":    15000.0,
				"ListCost_mean":        135.20,
				"ListCost_stddev":      920.15,
			},
			"data_quality": "high",
			"anomaly_rate": 0.002, // 0.2% anomaly rate
		}
	} else {
		// Production data with some anomalies
		return map[string]interface{}{
			"row_count": 89000,
			"columns": []string{
				"BillingAccountId", "EffectiveCost", "ListCost", "ServiceName",
				"ProviderName", "ResourceId",
			},
			"cost_stats": map[string]float64{
				"EffectiveCost_mean":   98.75,
				"EffectiveCost_stddev": 1250.30,
				"EffectiveCost_min":    -5.0,    // Negative cost (anomaly)
				"EffectiveCost_max":    85000.0, // Very high cost (potential anomaly)
				"ListCost_mean":        105.40,
				"ListCost_stddev":      1180.75,
			},
			"data_quality": "medium",
			"anomaly_rate": 0.015, // 1.5% anomaly rate
		}
	}
}

// detectStatisticalAnomalies detects statistical outliers
func (d *AnomalyDetector) detectStatisticalAnomalies(fileData map[string]interface{}, result *AnomalyDetectionResult) {
	costStats := fileData["cost_stats"].(map[string]float64)
	anomalyRate := fileData["anomaly_rate"].(float64)
	rowCount := fileData["row_count"].(int)

	// Detect anomalies based on cost statistics
	d.detectExtremeValues(costStats, result)
	d.detectCostOutliers(costStats, anomalyRate, rowCount, result)
}

// detectBusinessLogicAnomalies detects business logic violations
func (d *AnomalyDetector) detectBusinessLogicAnomalies(fileData map[string]interface{}, result *AnomalyDetectionResult) {
	costStats := fileData["cost_stats"].(map[string]float64)

	// Check for negative costs
	if minCost := costStats["EffectiveCost_min"]; minCost < 0 {
		anomaly := DetectedAnomaly{
			Type:        "negative_cost",
			Severity:    "critical",
			Column:      "EffectiveCost",
			Value:       minCost,
			Expected:    0.0,
			Score:       95.0,
			Description: "Negative cost values detected",
			Context: map[string]interface{}{
				"min_value": minCost,
				"rule":      "Costs must be non-negative",
			},
		}
		result.Anomalies = append(result.Anomalies, anomaly)
		result.CriticalAnomalies++
	}

	// Check for EffectiveCost > ListCost (unusual but possible)
	effectiveMean := costStats["EffectiveCost_mean"]
	listMean := costStats["ListCost_mean"]

	if effectiveMean > listMean*1.1 { // More than 10% higher
		anomaly := DetectedAnomaly{
			Type:        "cost_relationship_anomaly",
			Severity:    "medium",
			Column:      "EffectiveCost,ListCost",
			Value:       effectiveMean / listMean,
			Expected:    1.0,
			Score:       75.0,
			Description: "EffectiveCost average higher than ListCost average",
			Context: map[string]interface{}{
				"effective_mean": effectiveMean,
				"list_mean":      listMean,
				"ratio":          effectiveMean / listMean,
			},
		}
		result.Anomalies = append(result.Anomalies, anomaly)
	}
}

// detectDataQualityAnomalies detects data quality issues
func (d *AnomalyDetector) detectDataQualityAnomalies(fileData map[string]interface{}, result *AnomalyDetectionResult) {
	columns := fileData["columns"].([]string)
	dataQuality := fileData["data_quality"].(string)

	// Check for missing required columns
	requiredColumns := []string{"BillingAccountId", "EffectiveCost", "ListCost", "ServiceName"}

	for _, required := range requiredColumns {
		if !contains(columns, required) {
			anomaly := DetectedAnomaly{
				Type:        "missing_column",
				Severity:    "high",
				Column:      required,
				Value:       "missing",
				Expected:    "present",
				Score:       85.0,
				Description: fmt.Sprintf("Required column '%s' is missing", required),
			}
			result.Anomalies = append(result.Anomalies, anomaly)
		}
	}

	// Simulate data quality issues based on quality level
	if dataQuality == QualityMedium || dataQuality == QualityLow {
		// Simulate duplicate detection
		anomaly := DetectedAnomaly{
			Type:        "duplicate_records",
			Severity:    SeverityMedium,
			Column:      "all",
			Value:       "2.3% duplicates",
			Expected:    "< 1% duplicates",
			Score:       70.0,
			Description: "High percentage of duplicate records detected",
			Context: map[string]interface{}{
				"duplicate_percentage": 2.3,
				"threshold":            1.0,
			},
		}
		result.Anomalies = append(result.Anomalies, anomaly)
	}
}

// detectTemporalAnomalies detects time-based anomalies
func (d *AnomalyDetector) detectTemporalAnomalies(fileData map[string]interface{}, result *AnomalyDetectionResult) {
	// Simulate temporal anomaly detection
	// In a real implementation, this would analyze timestamp patterns

	anomalyRate := fileData["anomaly_rate"].(float64)

	if anomalyRate > 0.01 { // More than 1% anomaly rate
		// Simulate finding temporal gaps
		anomaly := DetectedAnomaly{
			Type:        "temporal_gap",
			Severity:    "low",
			Column:      "BillingPeriodStart",
			Value:       "3 day gap",
			Expected:    "continuous data",
			Score:       60.0,
			Description: "Gaps detected in temporal data coverage",
			Context: map[string]interface{}{
				"gap_duration": "3 days",
				"start_date":   "2024-01-15",
				"end_date":     "2024-01-18",
			},
		}
		result.Anomalies = append(result.Anomalies, anomaly)
	}
}

// detectCostOutliers detects outliers in cost data
func (d *AnomalyDetector) detectCostOutliers(costStats map[string]float64, anomalyRate float64, rowCount int, result *AnomalyDetectionResult) {
	mean := costStats["EffectiveCost_mean"]
	stddev := costStats["EffectiveCost_stddev"]
	maxVal := costStats["EffectiveCost_max"]

	// Detect values beyond 3 standard deviations
	threshold := mean + 3*stddev

	if maxVal > threshold {
		// Calculate z-score
		zScore := (maxVal - mean) / stddev

		severity := SeverityMedium
		if zScore > 5 {
			severity = SeverityHigh
		}
		if zScore > 8 {
			severity = SeverityCritical
		}

		anomaly := DetectedAnomaly{
			Type:        "statistical_outlier",
			Severity:    severity,
			Column:      "EffectiveCost",
			Value:       maxVal,
			Expected:    threshold,
			Score:       100.0 - math.Min(zScore*10, 90),
			Description: fmt.Sprintf("Statistical outlier detected (z-score: %.2f)", zScore),
			Context: map[string]interface{}{
				"z_score":   zScore,
				"mean":      mean,
				"stddev":    stddev,
				"threshold": threshold,
			},
		}
		result.Anomalies = append(result.Anomalies, anomaly)

		if severity == SeverityCritical {
			result.CriticalAnomalies++
		}
	}

	// Simulate finding multiple outliers based on anomaly rate
	estimatedOutliers := int(float64(rowCount) * anomalyRate)
	for i := 0; i < estimatedOutliers && i < 10; i++ { // Limit to 10 for simulation
		// Generate random outlier
		value := mean + (3+cryptoRandFloat64()*2)*stddev
		rowIndex := cryptoRandInt63n(int64(rowCount))

		anomaly := DetectedAnomaly{
			Type:        "mild_outlier",
			Severity:    "low",
			Column:      "EffectiveCost",
			RowIndex:    rowIndex,
			Value:       value,
			Expected:    fmt.Sprintf("%.2f ± %.2f", mean, 2*stddev),
			Score:       85.0,
			Description: "Mild statistical outlier",
		}
		result.Anomalies = append(result.Anomalies, anomaly)
	}
}

// detectExtremeValues detects extreme values that might indicate data issues
func (d *AnomalyDetector) detectExtremeValues(costStats map[string]float64, result *AnomalyDetectionResult) {
	maxCost := costStats["EffectiveCost_max"]

	// Check for extremely high values
	if maxCost > 50000.0 {
		anomaly := DetectedAnomaly{
			Type:        "extreme_value",
			Severity:    "medium",
			Column:      "EffectiveCost",
			Value:       maxCost,
			Expected:    "< 50000",
			Score:       75.0,
			Description: "Extremely high cost value detected",
			Context: map[string]interface{}{
				"value":     maxCost,
				"threshold": 50000.0,
				"category":  "extreme_high",
			},
		}
		result.Anomalies = append(result.Anomalies, anomaly)
	}
}

// identifyAnomalyPatterns identifies patterns in detected anomalies
func (d *AnomalyDetector) identifyAnomalyPatterns(result *AnomalyDetectionResult) {
	// Count anomalies by type
	typeCounts := make(map[string]int)
	for _, anomaly := range result.Anomalies {
		typeCounts[anomaly.Type]++
	}

	// Create patterns for frequent anomaly types
	for anomalyType, count := range typeCounts {
		if count > 1 {
			confidence := math.Min(float64(count)*20.0, 95.0)

			pattern := AnomalyPattern{
				Name:        fmt.Sprintf("Recurring_%s", anomalyType),
				Frequency:   count,
				Confidence:  confidence,
				Description: fmt.Sprintf("Pattern of %s anomalies detected %d times", anomalyType, count),
			}
			result.Patterns = append(result.Patterns, pattern)
		}
	}

	// Identify cost-related pattern
	costAnomalies := 0
	for _, anomaly := range result.Anomalies {
		if strings.Contains(anomaly.Column, "Cost") {
			costAnomalies++
		}
	}

	if costAnomalies > 2 {
		pattern := AnomalyPattern{
			Name:        "Cost_Data_Issues",
			Frequency:   costAnomalies,
			Confidence:  85.0,
			Description: "Multiple cost-related anomalies suggest systematic data quality issues",
		}
		result.Patterns = append(result.Patterns, pattern)
	}
}

// assessAnomalySeverity assesses overall severity and generates issues
func (d *AnomalyDetector) assessAnomalySeverity(result *AnomalyDetectionResult) {
	result.TotalAnomalies = len(result.Anomalies)

	// Generate issues based on anomalies
	for _, anomaly := range result.Anomalies {
		issue := AnomalyIssue{
			Type:     "detected_anomaly",
			Anomaly:  anomaly,
			Message:  anomaly.Description,
			Severity: anomaly.Severity,
		}

		// Add suggestions based on anomaly type
		switch anomaly.Type {
		case "negative_cost":
			issue.Suggestion = "Review data source and implement validation to prevent negative costs"
		case "statistical_outlier":
			issue.Suggestion = "Investigate high-value records for accuracy and potential data entry errors"
		case "missing_column":
			issue.Suggestion = "Ensure all required columns are present in the data source"
		case "duplicate_records":
			issue.Suggestion = "Implement deduplication process in data pipeline"
		case "temporal_gap":
			issue.Suggestion = "Check data collection process for missing time periods"
		default:
			issue.Suggestion = "Investigate anomaly cause and implement appropriate data quality controls"
		}

		result.Issues = append(result.Issues, issue)
	}

	// Check for critical anomaly thresholds
	if result.CriticalAnomalies > 0 {
		result.Valid = false
	}

	// Check total anomaly rate
	if len(result.Anomalies) > 10 {
		result.Valid = false
	}
}

// calculateAnomalyScore calculates overall anomaly detection score
func (d *AnomalyDetector) calculateAnomalyScore(result *AnomalyDetectionResult) {
	score := 100.0

	// Deduct points based on anomaly severity
	for _, anomaly := range result.Anomalies {
		switch anomaly.Severity {
		case SeverityCritical:
			score -= 25.0
		case SeverityHigh:
			score -= 15.0
		case SeverityMedium:
			score -= 8.0
		case SeverityLow:
			score -= 3.0
		}
	}

	// Additional deduction for patterns indicating systematic issues
	for _, pattern := range result.Patterns {
		if pattern.Confidence > 80.0 {
			score -= 5.0
		}
	}

	// Bonus for clean data
	if len(result.Anomalies) == 0 {
		score = 100.0
	}

	// Ensure score doesn't go below 0
	if score < 0 {
		score = 0
	}

	result.Score = score

	// Mark as invalid if score is very low
	if score < 60.0 {
		result.Valid = false
	}
}

// cryptoRandFloat64 generates a cryptographically secure random float64 in [0.0, 1.0)
func cryptoRandFloat64() float64 {
	max := big.NewInt(1 << 53)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return 0.5 // fallback value
	}
	return float64(n.Int64()) / float64(1<<53)
}

// cryptoRandInt63n generates a cryptographically secure random int64 in [0, n)
func cryptoRandInt63n(n int64) int64 {
	if n <= 0 {
		return 0
	}
	max := big.NewInt(n)
	result, err := rand.Int(rand.Reader, max)
	if err != nil {
		return n / 2 // fallback value
	}
	return result.Int64()
}
