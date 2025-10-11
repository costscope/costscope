package types

import (
	"time"
)

// CostAnalysisReport represents a comprehensive cost analysis report
type CostAnalysisReport struct {
	ID             string                       `json:"id"`
	Title          string                       `json:"title"`
	Description    string                       `json:"description"`
	GeneratedAt    time.Time                    `json:"generated_at"`
	DateRange      DateRange                    `json:"date_range"`
	TotalCost      float64                      `json:"total_cost"`
	Currency       string                       `json:"currency"`
	CostByService  []ServiceCostBreakdown       `json:"cost_by_service"`
	CostByRegion   []RegionCostBreakdown        `json:"cost_by_region"`
	CostByAccount  []AccountCostBreakdown       `json:"cost_by_account"`
	CostTrends     []CostTrendData              `json:"cost_trends"`
	TopCostDrivers []CostDriver                 `json:"top_cost_drivers"`
	Optimization   []OptimizationRecommendation `json:"optimization_recommendations"`
	Summary        ReportSummary                `json:"summary"`
	Metadata       map[string]interface{}       `json:"metadata,omitempty"`
}

// UsageSummaryReport represents a resource usage summary report
type UsageSummaryReport struct {
	ID                  string                 `json:"id"`
	Title               string                 `json:"title"`
	Description         string                 `json:"description"`
	GeneratedAt         time.Time              `json:"generated_at"`
	DateRange           DateRange              `json:"date_range"`
	ResourceUtilization []ResourceUtilization  `json:"resource_utilization"`
	ServiceUsage        []ServiceUsage         `json:"service_usage"`
	UsageTrends         []UsageTrendData       `json:"usage_trends"`
	CapacityAnalysis    []CapacityAnalysis     `json:"capacity_analysis"`
	Summary             ReportSummary          `json:"summary"`
	Metadata            map[string]interface{} `json:"metadata,omitempty"`
}

// TrendAnalysisReport represents a trend analysis report
type TrendAnalysisReport struct {
	ID          string                 `json:"id"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	GeneratedAt time.Time              `json:"generated_at"`
	DateRange   DateRange              `json:"date_range"`
	Trends      []TrendAnalysis        `json:"trends"`
	Forecasts   []ForecastData         `json:"forecasts"`
	Seasonality []SeasonalPattern      `json:"seasonality"`
	MLInsights  []MLInsight            `json:"ml_insights,omitempty"`
	Summary     ReportSummary          `json:"summary"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// AnomalyReport represents an anomaly detection report
type AnomalyReport struct {
	ID          string                 `json:"id"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	GeneratedAt time.Time              `json:"generated_at"`
	DateRange   DateRange              `json:"date_range"`
	Anomalies   []AnomalyData          `json:"anomalies"`
	Alerts      []AlertData            `json:"alerts"`
	RiskLevel   string                 `json:"risk_level"`
	Summary     ReportSummary          `json:"summary"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// ForecastReport represents a cost forecasting report
type ForecastReport struct {
	ID          string                 `json:"id"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	GeneratedAt time.Time              `json:"generated_at"`
	DateRange   DateRange              `json:"date_range"`
	Forecasts   []ForecastData         `json:"forecasts"`
	Scenarios   []ScenarioData         `json:"scenarios"`
	Confidence  float64                `json:"confidence"`
	Summary     ReportSummary          `json:"summary"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// ExecutiveSummaryReport represents an executive summary report
type ExecutiveSummaryReport struct {
	ID               string                    `json:"id"`
	Title            string                    `json:"title"`
	Description      string                    `json:"description"`
	GeneratedAt      time.Time                 `json:"generated_at"`
	DateRange        DateRange                 `json:"date_range"`
	ExecutiveSummary ExecutiveSummaryData      `json:"executive_summary"`
	KeyMetrics       []KeyMetric               `json:"key_metrics"`
	CostHighlights   []CostHighlight           `json:"cost_highlights"`
	Recommendations  []ExecutiveRecommendation `json:"recommendations"`
	BudgetVariance   []BudgetVarianceData      `json:"budget_variance"`
	Summary          ReportSummary             `json:"summary"`
	Metadata         map[string]interface{}    `json:"metadata,omitempty"`
}

// Supporting data structures

// ServiceCostBreakdown represents cost breakdown by service
type ServiceCostBreakdown struct {
	ServiceName string  `json:"service_name"`
	Provider    string  `json:"provider"`
	Cost        float64 `json:"cost"`
	Percentage  float64 `json:"percentage"`
	Trend       string  `json:"trend"`
}

// RegionCostBreakdown represents cost breakdown by region
type RegionCostBreakdown struct {
	Region     string  `json:"region"`
	Provider   string  `json:"provider"`
	Cost       float64 `json:"cost"`
	Percentage float64 `json:"percentage"`
	Trend      string  `json:"trend"`
}

// AccountCostBreakdown represents cost breakdown by account
type AccountCostBreakdown struct {
	AccountID   string  `json:"account_id"`
	AccountName string  `json:"account_name"`
	Provider    string  `json:"provider"`
	Cost        float64 `json:"cost"`
	Percentage  float64 `json:"percentage"`
	Trend       string  `json:"trend"`
}

// CostTrendData represents cost trend information
type CostTrendData struct {
	Date     time.Time `json:"date"`
	Cost     float64   `json:"cost"`
	Service  string    `json:"service,omitempty"`
	Region   string    `json:"region,omitempty"`
	Provider string    `json:"provider,omitempty"`
}

// CostDriver represents a significant cost driver
type CostDriver struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Type        string  `json:"type"`
	Cost        float64 `json:"cost"`
	Impact      string  `json:"impact"`
	Trend       string  `json:"trend"`
	Description string  `json:"description"`
}

// OptimizationRecommendation represents a cost optimization recommendation
type OptimizationRecommendation struct {
	ID                 string   `json:"id"`
	Type               string   `json:"type"`
	Priority           Priority `json:"priority"`
	Title              string   `json:"title"`
	Description        string   `json:"description"`
	Impact             string   `json:"impact"`
	Effort             string   `json:"effort"`
	PotentialSavings   float64  `json:"potential_savings"`
	ImplementationTime string   `json:"implementation_time"`
	RiskLevel          string   `json:"risk_level"`
	Resources          []string `json:"resources,omitempty"`
}

// ResourceUtilization represents resource utilization data
type ResourceUtilization struct {
	ResourceID   string            `json:"resource_id"`
	ResourceType string            `json:"resource_type"`
	Provider     string            `json:"provider"`
	Region       string            `json:"region"`
	Utilization  float64           `json:"utilization_percent"`
	Capacity     map[string]string `json:"capacity"`
	Status       string            `json:"status"`
}

// ServiceUsage represents service usage statistics
type ServiceUsage struct {
	ServiceName string                 `json:"service_name"`
	Provider    string                 `json:"provider"`
	Usage       map[string]interface{} `json:"usage"`
	Cost        float64                `json:"cost"`
	Trend       string                 `json:"trend"`
}

// UsageTrendData represents usage trend information
type UsageTrendData struct {
	Date     time.Time              `json:"date"`
	Service  string                 `json:"service"`
	Provider string                 `json:"provider"`
	Usage    map[string]interface{} `json:"usage"`
}

// CapacityAnalysis represents capacity analysis data
type CapacityAnalysis struct {
	ResourceType      string  `json:"resource_type"`
	Provider          string  `json:"provider"`
	CurrentCapacity   float64 `json:"current_capacity"`
	UsedCapacity      float64 `json:"used_capacity"`
	AvailableCapacity float64 `json:"available_capacity"`
	UtilizationRate   float64 `json:"utilization_rate"`
	Recommendation    string  `json:"recommendation"`
}

// TrendAnalysis represents trend analysis data
type TrendAnalysis struct {
	Metric      string    `json:"metric"`
	Direction   string    `json:"direction"`
	Strength    float64   `json:"strength"`
	Confidence  float64   `json:"confidence"`
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
	Description string    `json:"description"`
}

// ForecastData represents forecast information
type ForecastData struct {
	Metric         string    `json:"metric"`
	Date           time.Time `json:"date"`
	Value          float64   `json:"value"`
	ConfidenceLow  float64   `json:"confidence_low"`
	ConfidenceHigh float64   `json:"confidence_high"`
	Method         string    `json:"method"`
}

// SeasonalPattern represents seasonal pattern data
type SeasonalPattern struct {
	Pattern     string             `json:"pattern"`
	Strength    float64            `json:"strength"`
	Periods     []SeasonalPeriod   `json:"periods"`
	Factors     map[string]float64 `json:"factors"`
	Description string             `json:"description"`
}

// SeasonalPeriod represents a seasonal period
type SeasonalPeriod struct {
	Name      string  `json:"name"`
	StartDate string  `json:"start_date"`
	EndDate   string  `json:"end_date"`
	Factor    float64 `json:"factor"`
}

// MLInsight represents machine learning insights
type MLInsight struct {
	Type        string                 `json:"type"`
	Confidence  float64                `json:"confidence"`
	Description string                 `json:"description"`
	Impact      string                 `json:"impact"`
	Data        map[string]interface{} `json:"data"`
}

// AnomalyData represents anomaly detection data
type AnomalyData struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Metric      string                 `json:"metric"`
	DetectedAt  time.Time              `json:"detected_at"`
	Severity    string                 `json:"severity"`
	Score       float64                `json:"score"`
	Expected    float64                `json:"expected"`
	Actual      float64                `json:"actual"`
	Deviation   float64                `json:"deviation"`
	Description string                 `json:"description"`
	Context     map[string]interface{} `json:"context"`
}

// AlertData represents alert information
type AlertData struct {
	ID           string                 `json:"id"`
	Type         string                 `json:"type"`
	Severity     string                 `json:"severity"`
	Message      string                 `json:"message"`
	CreatedAt    time.Time              `json:"created_at"`
	Acknowledged bool                   `json:"acknowledged"`
	Context      map[string]interface{} `json:"context"`
}

// ScenarioData represents scenario analysis data
type ScenarioData struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Probability float64                `json:"probability"`
	Impact      string                 `json:"impact"`
	Results     map[string]interface{} `json:"results"`
}

// ExecutiveSummaryData represents executive summary content
type ExecutiveSummaryData struct {
	TotalSpend      float64  `json:"total_spend"`
	SpendChange     float64  `json:"spend_change_percent"`
	TopCostDriver   string   `json:"top_cost_driver"`
	OptimizationOpp float64  `json:"optimization_opportunity"`
	RiskLevel       string   `json:"risk_level"`
	Recommendations int      `json:"total_recommendations"`
	KeyInsights     []string `json:"key_insights"`
}

// KeyMetric represents a key metric
type KeyMetric struct {
	Name        string      `json:"name"`
	Value       interface{} `json:"value"`
	Unit        string      `json:"unit,omitempty"`
	Change      float64     `json:"change_percent,omitempty"`
	Trend       string      `json:"trend,omitempty"`
	Description string      `json:"description,omitempty"`
}

// CostHighlight represents a cost highlight
type CostHighlight struct {
	Title       string  `json:"title"`
	Amount      float64 `json:"amount"`
	Change      float64 `json:"change_percent"`
	Impact      string  `json:"impact"`
	Description string  `json:"description"`
}

// ExecutiveRecommendation represents an executive-level recommendation
type ExecutiveRecommendation struct {
	ID               string   `json:"id"`
	Title            string   `json:"title"`
	Priority         Priority `json:"priority"`
	PotentialSavings float64  `json:"potential_savings"`
	Timeframe        string   `json:"timeframe"`
	Complexity       string   `json:"complexity"`
	Description      string   `json:"description"`
}

// BudgetVarianceData represents budget variance information
type BudgetVarianceData struct {
	Category        string  `json:"category"`
	Budget          float64 `json:"budget"`
	Actual          float64 `json:"actual"`
	Variance        float64 `json:"variance"`
	VariancePercent float64 `json:"variance_percent"`
	Status          string  `json:"status"`
}
