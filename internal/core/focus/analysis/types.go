package analysis

import (
	"time"
)

// DiffResult represents the comprehensive comparison result between two FOCUS datasets
type DiffResult struct {
	Summary     DiffSummary              `json:"summary"`
	Changes     []CostChange             `json:"changes"`
	NewServices []ServiceInfo            `json:"new_services"`
	Removed     []ServiceInfo            `json:"removed_services"`
	Trends      map[string]DiffTrendInfo `json:"trends"`
	Anomalies   []AnomalyInfo            `json:"anomalies"`
	Metadata    DiffMetadata             `json:"metadata"`
}

// DiffSummary provides high-level comparison metrics
type DiffSummary struct {
	TotalCostChange      float64 `json:"total_cost_change"`
	PercentageChange     float64 `json:"percentage_change"`
	SignificantChanges   int     `json:"significant_changes"`
	NewServicesCount     int     `json:"new_services_count"`
	RemovedServicesCount int     `json:"removed_services_count"`
	AnomaliesDetected    int     `json:"anomalies_detected"`
	BaselinePeriod       string  `json:"baseline_period"`
	ComparisonPeriod     string  `json:"comparison_period"`
}

// CostChange represents a detected cost change between datasets
type CostChange struct {
	Service         string    `json:"service"`
	Region          string    `json:"region"`
	Account         string    `json:"account"`
	ResourceID      string    `json:"resource_id"`
	BaselineCost    float64   `json:"baseline_cost"`
	CurrentCost     float64   `json:"current_cost"`
	Change          float64   `json:"change"`
	PercentChange   float64   `json:"percent_change"`
	Significance    string    `json:"significance"` // "low", "medium", "high", "critical"
	Category        string    `json:"category"`     // "increase", "decrease", "new", "removed"
	ConfidenceLevel float64   `json:"confidence_level"`
	DetectionMethod string    `json:"detection_method"`
	Timestamp       time.Time `json:"timestamp"`
}

// ServiceInfo represents service-level information
type ServiceInfo struct {
	Service       string    `json:"service"`
	Region        string    `json:"region"`
	Account       string    `json:"account"`
	Cost          float64   `json:"cost"`
	UsageQuantity float64   `json:"usage_quantity"`
	UsageUnit     string    `json:"usage_unit"`
	FirstSeen     time.Time `json:"first_seen"`
	LastSeen      time.Time `json:"last_seen"`
	ResourceCount int       `json:"resource_count"`
}

// DiffTrendInfo provides trend analysis for services
type DiffTrendInfo struct {
	Service     string     `json:"service"`
	Region      string     `json:"region"`
	Trend       string     `json:"trend"`       // "increasing", "decreasing", "stable", "volatile"
	Velocity    float64    `json:"velocity"`    // Rate of change
	Prediction  float64    `json:"prediction"`  // Predicted next period cost
	Confidence  float64    `json:"confidence"`  // Prediction confidence 0-1
	DataPoints  []float64  `json:"data_points"` // Historical data points
	Seasonality []float64  `json:"seasonality"` // Detected seasonal patterns
	Forecasts   []Forecast `json:"forecasts"`   // Multi-period forecasts
}

// Forecast represents a future cost prediction
type Forecast struct {
	Period          string    `json:"period"`
	PredictedCost   float64   `json:"predicted_cost"`
	ConfidenceUpper float64   `json:"confidence_upper"`
	ConfidenceLower float64   `json:"confidence_lower"`
	PredictionDate  time.Time `json:"prediction_date"`
}

// AnomalyInfo represents detected cost anomalies
type AnomalyInfo struct {
	Service         string    `json:"service"`
	Region          string    `json:"region"`
	Account         string    `json:"account"`
	ResourceID      string    `json:"resource_id"`
	AnomalyType     string    `json:"anomaly_type"` // "spike", "drop", "trend_break", "outlier"
	Severity        string    `json:"severity"`     // "low", "medium", "high", "critical"
	DetectedAt      time.Time `json:"detected_at"`
	AnomalyScore    float64   `json:"anomaly_score"` // 0-1 anomaly score
	ExpectedCost    float64   `json:"expected_cost"`
	ActualCost      float64   `json:"actual_cost"`
	Deviation       float64   `json:"deviation"`
	ConfidenceLevel float64   `json:"confidence_level"`
	DetectionMethod string    `json:"detection_method"` // "isolation_forest", "statistical", "ml_model"
	Description     string    `json:"description"`
	Recommendations []string  `json:"recommendations"`
}

// DiffMetadata contains analysis metadata
type DiffMetadata struct {
	AnalysisDate      time.Time `json:"analysis_date"`
	Threshold         float64   `json:"threshold"`
	Dimension         []string  `json:"dimension"` // ["service", "region", "account"]
	ProcessingTime    string    `json:"processing_time"`
	BaselineRecords   int       `json:"baseline_records"`
	ComparisonRecords int       `json:"comparison_records"`
	MLEnabled         bool      `json:"ml_enabled"`
	AnomalyDetection  bool      `json:"anomaly_detection"`
	TrendAnalysis     bool      `json:"trend_analysis"`
	Version           string    `json:"version"`
}

// DiffOptions represents configuration for diff operations
type DiffOptions struct {
	Threshold         float64  `json:"threshold"`
	Dimensions        []string `json:"dimensions"`
	ShowAnomalies     bool     `json:"show_anomalies"`
	ShowTrends        bool     `json:"show_trends"`
	MLEnabled         bool     `json:"ml_enabled"`
	SignificanceLevel float64  `json:"significance_level"`
	ConfidenceLevel   float64  `json:"confidence_level"`
	ForecastPeriods   int      `json:"forecast_periods"`
	AnomalyMethods    []string `json:"anomaly_methods"`
	ExcludeServices   []string `json:"exclude_services"`
	IncludeServices   []string `json:"include_services"`
	ExcludeRegions    []string `json:"exclude_regions"`
	IncludeRegions    []string `json:"include_regions"`
	OutputFormat      string   `json:"output_format"`
	OutputFile        string   `json:"output_file"`
	Verbose           bool     `json:"verbose"`
}

// AnalysisResult represents comprehensive analysis results
type AnalysisResult struct {
	Summary          AnalysisSummary   `json:"summary"`
	CostTrends       []TrendAnalysis   `json:"cost_trends"`
	Anomalies        []AnomalyInfo     `json:"anomalies"`
	Optimizations    []OptimizationRec `json:"optimizations"`
	Forecasts        []Forecast        `json:"forecasts"`
	ServiceBreakdown []ServiceSummary  `json:"service_breakdown"`
	Metadata         AnalysisMetadata  `json:"metadata"`
	Extended         interface{}       `json:"extended,omitempty"`
}

// AnalysisSummary provides high-level analysis metrics
type AnalysisSummary struct {
	TotalCost          float64   `json:"total_cost"`
	PeriodStart        time.Time `json:"period_start"`
	PeriodEnd          time.Time `json:"period_end"`
	DaysAnalyzed       int       `json:"days_analyzed"`
	ServicesCount      int       `json:"services_count"`
	RegionsCount       int       `json:"regions_count"`
	AccountsCount      int       `json:"accounts_count"`
	ResourcesCount     int       `json:"resources_count"`
	TrendsDetected     int       `json:"trends_detected"`
	AnomaliesDetected  int       `json:"anomalies_detected"`
	OptimizationsFound int       `json:"optimizations_found"`
	PotentialSavings   float64   `json:"potential_savings"`
	CostGrowthRate     float64   `json:"cost_growth_rate"`
	TopCostServices    []string  `json:"top_cost_services"`
	TopGrowthServices  []string  `json:"top_growth_services"`
}

// TrendAnalysis represents trend analysis for a service/resource
type TrendAnalysis struct {
	Service        string      `json:"service"`
	Region         string      `json:"region"`
	Account        string      `json:"account"`
	TrendType      string      `json:"trend_type"` // "linear", "exponential", "seasonal", "irregular"
	Direction      string      `json:"direction"`  // "increasing", "decreasing", "stable"
	Strength       float64     `json:"strength"`   // 0-1 trend strength
	Seasonality    bool        `json:"seasonality"`
	SeasonalPeriod int         `json:"seasonal_period"` // Days in seasonal cycle
	GrowthRate     float64     `json:"growth_rate"`     // Daily growth rate
	RSquared       float64     `json:"r_squared"`       // Trend fit quality
	DataPoints     []DataPoint `json:"data_points"`
	Forecasts      []Forecast  `json:"forecasts"`
}

// DataPoint represents a single cost data point
type DataPoint struct {
	Date   time.Time `json:"date"`
	Cost   float64   `json:"cost"`
	Usage  float64   `json:"usage"`
	Source string    `json:"source"`
}

// OptimizationRec represents a cost optimization recommendation
type OptimizationRec struct {
	Type           string   `json:"type"` // "rightsizing", "reserved_instances", "spot_instances", "scheduling"
	Service        string   `json:"service"`
	Region         string   `json:"region"`
	Account        string   `json:"account"`
	ResourceID     string   `json:"resource_id"`
	CurrentCost    float64  `json:"current_cost"`
	OptimizedCost  float64  `json:"optimized_cost"`
	Savings        float64  `json:"savings"`
	SavingsPercent float64  `json:"savings_percent"`
	Confidence     float64  `json:"confidence"`
	Priority       string   `json:"priority"` // "low", "medium", "high", "critical"
	Effort         string   `json:"effort"`   // "low", "medium", "high"
	Risk           string   `json:"risk"`     // "low", "medium", "high"
	Description    string   `json:"description"`
	Implementation []string `json:"implementation"`
	Validation     []string `json:"validation"`
	Timeline       string   `json:"timeline"` // "immediate", "short_term", "long_term"
	Dependencies   []string `json:"dependencies"`
}

// ServiceSummary provides service-level summary
type ServiceSummary struct {
	Service            string  `json:"service"`
	Region             string  `json:"region"`
	Account            string  `json:"account"`
	TotalCost          float64 `json:"total_cost"`
	CostPercentage     float64 `json:"cost_percentage"`
	ResourceCount      int     `json:"resource_count"`
	UsageQuantity      float64 `json:"usage_quantity"`
	UsageUnit          string  `json:"usage_unit"`
	AvgDailyCost       float64 `json:"avg_daily_cost"`
	TrendDirection     string  `json:"trend_direction"`
	AnomaliesCount     int     `json:"anomalies_count"`
	OptimizationsCount int     `json:"optimizations_count"`
	PotentialSavings   float64 `json:"potential_savings"`
}

// AnalysisMetadata contains analysis metadata
type AnalysisMetadata struct {
	AnalysisDate     time.Time              `json:"analysis_date"`
	ProcessingTime   string                 `json:"processing_time"`
	RecordsProcessed int                    `json:"records_processed"`
	MLModelsUsed     []string               `json:"ml_models_used"`
	AnalysisTypes    []string               `json:"analysis_types"`
	Version          string                 `json:"version"`
	Configuration    map[string]interface{} `json:"configuration"`
}

// AnalysisOptions represents configuration for analysis operations
type AnalysisOptions struct {
	Type                 string    `json:"type"` // "comprehensive", "anomaly-detection", "optimization", "forecasting"
	StartDate            time.Time `json:"start_date"`
	EndDate              time.Time `json:"end_date"`
	Granularity          string    `json:"granularity"` // "daily", "weekly", "monthly"
	GroupBy              []string  `json:"group_by"`    // ["service", "region", "account"]
	Services             []string  `json:"services"`
	Regions              []string  `json:"regions"`
	Accounts             []string  `json:"accounts"`
	MLEnabled            bool      `json:"ml_enabled"`
	AnomalyDetection     bool      `json:"anomaly_detection"`
	TrendAnalysis        bool      `json:"trend_analysis"`
	OptimizationAnalysis bool      `json:"optimization_analysis"`
	ForecastDays         int       `json:"forecast_days"`
	ConfidenceLevel      float64   `json:"confidence_level"`
	AnomalyThreshold     float64   `json:"anomaly_threshold"`
	OutputFormat         string    `json:"output_format"`
	OutputFile           string    `json:"output_file"`
	Verbose              bool      `json:"verbose"`
}
