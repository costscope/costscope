package comparison

import (
	"time"

	focusTypes "github.com/costscope/costscope/internal/core/focus/analysis"
)

// Re-export analysis types for comparison module
type (
	DiffResult    = focusTypes.DiffResult
	DiffSummary   = focusTypes.DiffSummary
	CostChange    = focusTypes.CostChange
	ServiceInfo   = focusTypes.ServiceInfo
	DiffTrendInfo = focusTypes.DiffTrendInfo
	AnomalyInfo   = focusTypes.AnomalyInfo
	DiffMetadata  = focusTypes.DiffMetadata
	DiffOptions   = focusTypes.DiffOptions
	Forecast      = focusTypes.Forecast
	DataPoint     = focusTypes.DataPoint
)

// ComparisonEngine provides dataset comparison capabilities
type ComparisonEngine interface {
	// Core comparison operations
	CompareFOCUSDatasets(baseline, current string, options DiffOptions) (*DiffResult, error)
	DetectCostChanges(baseline, current []FOCUSRecord, options DiffOptions) ([]CostChange, error)
	IdentifyServiceChanges(baseline, current []FOCUSRecord) (new, removed []ServiceInfo, err error)

	// Advanced analysis operations
	AnalyzeTrends(baseline, current []FOCUSRecord, options DiffOptions) (map[string]DiffTrendInfo, error)
	DetectAnomalies(data []FOCUSRecord, options DiffOptions) ([]AnomalyInfo, error)
	GenerateForecast(historical []FOCUSRecord, periods int) ([]Forecast, error)

	// Aggregated insights (diff + executive summary + optional forecast)
	GenerateComparisonInsights(baseline, current []FOCUSRecord, opts DiffOptions, forecastPeriods int) (*ComparisonInsights, error)

	// Reporting operations
	GenerateExecutiveSummary(result *DiffResult) (*ExecutiveSummary, error)
	ExportResults(result *DiffResult, format, output string) error
}

// FOCUSRecord represents a FOCUS dataset record for comparison
type FOCUSRecord struct {
	// Core FOCUS v1.2 fields
	BillingAccountID       string    `parquet:"name=billing_account_id"`
	BillingAccountName     string    `parquet:"name=billing_account_name"`
	BillingCurrency        string    `parquet:"name=billing_currency"`
	BillingPeriodEnd       time.Time `parquet:"name=billing_period_end"`
	BillingPeriodStart     time.Time `parquet:"name=billing_period_start"`
	ChargeType             string    `parquet:"name=charge_type"`
	CommitmentDiscountID   string    `parquet:"name=commitment_discount_id"`
	CommitmentDiscountName string    `parquet:"name=commitment_discount_name"`
	CommitmentDiscountType string    `parquet:"name=commitment_discount_type"`

	// Cost and pricing
	BilledCost      float64 `parquet:"name=billed_cost"`
	EffectiveCost   float64 `parquet:"name=effective_cost"`
	ListCost        float64 `parquet:"name=list_cost"`
	ListUnitPrice   float64 `parquet:"name=list_unit_price"`
	PricingCategory string  `parquet:"name=pricing_category"`
	PricingQuantity float64 `parquet:"name=pricing_quantity"`
	PricingUnit     string  `parquet:"name=pricing_unit"`

	// Resource information
	ResourceID      string `parquet:"name=resource_id"`
	ResourceName    string `parquet:"name=resource_name"`
	ResourceType    string `parquet:"name=resource_type"`
	ServiceCategory string `parquet:"name=service_category"`
	ServiceName     string `parquet:"name=service_name"`

	// Location and usage
	Region           string  `parquet:"name=region"`
	AvailabilityZone string  `parquet:"name=availability_zone"`
	UsageQuantity    float64 `parquet:"name=usage_quantity"`
	UsageUnit        string  `parquet:"name=usage_unit"`

	// Provider-specific
	Provider      string `parquet:"name=provider"`
	PublisherName string `parquet:"name=publisher_name"`

	// Sub account information
	SubAccountID   string `parquet:"name=sub_account_id"`
	SubAccountName string `parquet:"name=sub_account_name"`

	// Tags and metadata
	Tags map[string]string `parquet:"name=tags"`

	// Additional computed fields for analysis
	DailyAverageCost     float64 `parquet:"name=daily_average_cost"`
	MonthlyProjectedCost float64 `parquet:"name=monthly_projected_cost"`
	CostCategory         string  `parquet:"name=cost_category"`
	ResourceHash         string  `parquet:"name=resource_hash"`
}

// ComparisonConfiguration holds configuration for comparison operations
type ComparisonConfiguration struct {
	// Comparison settings
	SignificanceThreshold  float64  `json:"significance_threshold"`
	PercentChangeThreshold float64  `json:"percent_change_threshold"`
	MinCostThreshold       float64  `json:"min_cost_threshold"`
	Dimensions             []string `json:"dimensions"`

	// ML and analysis settings
	MLEnabled        bool    `json:"ml_enabled"`
	AnomalyDetection bool    `json:"anomaly_detection"`
	TrendAnalysis    bool    `json:"trend_analysis"`
	ForecastEnabled  bool    `json:"forecast_enabled"`
	ForecastPeriods  int     `json:"forecast_periods"`
	ConfidenceLevel  float64 `json:"confidence_level"`

	// Anomaly detection settings
	AnomalyMethods        []string               `json:"anomaly_methods"`
	AnomalyThreshold      float64                `json:"anomaly_threshold"`
	IsolationForestParams map[string]interface{} `json:"isolation_forest_params"`
	StatisticalParams     map[string]interface{} `json:"statistical_params"`

	// Filtering settings
	ExcludeServices []string `json:"exclude_services"`
	IncludeServices []string `json:"include_services"`
	ExcludeRegions  []string `json:"exclude_regions"`
	IncludeRegions  []string `json:"include_regions"`
	ExcludeAccounts []string `json:"exclude_accounts"`
	IncludeAccounts []string `json:"include_accounts"`

	// Output settings
	OutputFormat   string `json:"output_format"`
	IncludeDetails bool   `json:"include_details"`
	IncludeRawData bool   `json:"include_raw_data"`
	CompressOutput bool   `json:"compress_output"`
}

// ExecutiveSummary provides high-level insights for executives
type ExecutiveSummary struct {
	// Overview
	AnalysisPeriod     string    `json:"analysis_period"`
	GeneratedAt        time.Time `json:"generated_at"`
	TotalCostImpact    float64   `json:"total_cost_impact"`
	TotalCostImpactPct float64   `json:"total_cost_impact_pct"`

	// Key metrics
	KeyFindings       []KeyFinding  `json:"key_findings"`
	TopCostChanges    []CostChange  `json:"top_cost_changes"`
	CriticalAnomalies []AnomalyInfo `json:"critical_anomalies"`
	ImmediateActions  []ActionItem  `json:"immediate_actions"`

	// Strategic insights
	CostDrivers     []CostDriver              `json:"cost_drivers"`
	OptimizationOpp []OptimizationOpportunity `json:"optimization_opportunities"`
	RiskAreas       []RiskArea                `json:"risk_areas"`
	Recommendations []Recommendation          `json:"recommendations"`

	// Forecast and trends
	ForecastSummary ForecastSummary `json:"forecast_summary"`
	TrendInsights   []TrendInsight  `json:"trend_insights"`
}

// KeyFinding represents a key finding from the analysis
type KeyFinding struct {
	Category    string  `json:"category"` // "cost_increase", "cost_decrease", "new_service", "anomaly"
	Impact      string  `json:"impact"`   // "high", "medium", "low"
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Value       float64 `json:"value"`
	Unit        string  `json:"unit"`
	Confidence  float64 `json:"confidence"`
}

// ActionItem represents an immediate action item
type ActionItem struct {
	Priority    string `json:"priority"` // "critical", "high", "medium", "low"
	Category    string `json:"category"` // "investigation", "optimization", "governance"
	Title       string `json:"title"`
	Description string `json:"description"`
	Timeline    string `json:"timeline"` // "immediate", "this_week", "this_month"
	Owner       string `json:"owner"`    // "finops", "engineering", "management"
	Impact      string `json:"impact"`
}

// CostDriver represents a major cost driver
type CostDriver struct {
	Service          string  `json:"service"`
	Region           string  `json:"region"`
	CostContribution float64 `json:"cost_contribution"`
	ContributionPct  float64 `json:"contribution_pct"`
	GrowthRate       float64 `json:"growth_rate"`
	Trend            string  `json:"trend"`
	Explanation      string  `json:"explanation"`
}

// OptimizationOpportunity represents a cost optimization opportunity
type OptimizationOpportunity struct {
	Type             string  `json:"type"`
	Service          string  `json:"service"`
	PotentialSavings float64 `json:"potential_savings"`
	SavingsPercent   float64 `json:"savings_percent"`
	Effort           string  `json:"effort"`
	Risk             string  `json:"risk"`
	Description      string  `json:"description"`
	BusinessCase     string  `json:"business_case"`
}

// RiskArea represents a cost risk area
type RiskArea struct {
	Category    string   `json:"category"` // "spending_acceleration", "budget_overrun", "anomaly_cluster"
	Severity    string   `json:"severity"` // "critical", "high", "medium", "low"
	Service     string   `json:"service"`
	Region      string   `json:"region"`
	RiskScore   float64  `json:"risk_score"`
	Impact      float64  `json:"impact"`
	Probability float64  `json:"probability"`
	Description string   `json:"description"`
	Mitigation  []string `json:"mitigation"`
}

// Recommendation represents a strategic recommendation
type Recommendation struct {
	Type        string   `json:"type"` // "strategic", "tactical", "operational"
	Priority    string   `json:"priority"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Rationale   string   `json:"rationale"`
	Benefits    []string `json:"benefits"`
	Risks       []string `json:"risks"`
	Timeline    string   `json:"timeline"`
	Success     []string `json:"success_metrics"`
}

// ForecastSummary provides forecast summary
type ForecastSummary struct {
	NextPeriodCost    float64 `json:"next_period_cost"`
	GrowthRate        float64 `json:"growth_rate"`
	ConfidenceLevel   float64 `json:"confidence_level"`
	UpperBound        float64 `json:"upper_bound"`
	LowerBound        float64 `json:"lower_bound"`
	SeasonalityFactor float64 `json:"seasonality_factor"`
	TrendStrength     float64 `json:"trend_strength"`
}

// TrendInsight provides trend analysis insights
type TrendInsight struct {
	Service    string  `json:"service"`
	Region     string  `json:"region"`
	TrendType  string  `json:"trend_type"`
	Strength   float64 `json:"strength"`
	Impact     string  `json:"impact"`
	Insight    string  `json:"insight"`
	Prediction string  `json:"prediction"`
}

// ComparisonMetrics holds metrics for comparison operations
type ComparisonMetrics struct {
	ProcessingStartTime    time.Time     `json:"processing_start_time"`
	ProcessingEndTime      time.Time     `json:"processing_end_time"`
	ProcessingDuration     time.Duration `json:"processing_duration"`
	BaselineRecordsCount   int           `json:"baseline_records_count"`
	CurrentRecordsCount    int           `json:"current_records_count"`
	MatchedRecordsCount    int           `json:"matched_records_count"`
	NewRecordsCount        int           `json:"new_records_count"`
	RemovedRecordsCount    int           `json:"removed_records_count"`
	ChangesDetectedCount   int           `json:"changes_detected_count"`
	AnomaliesDetectedCount int           `json:"anomalies_detected_count"`
	MLModelsExecuted       []string      `json:"ml_models_executed"`
	MemoryUsageMB          float64       `json:"memory_usage_mb"`
	CPUUsagePercent        float64       `json:"cpu_usage_percent"`
}
