package types

import "time"

// AnalyticsCommand represents analytics command configuration
type AnalyticsCommand struct {
	Use   string
	Short string
	Long  string
}

// AnalyticsOptions holds configuration for analytics operations
type AnalyticsOptions struct {
	// Core analysis options
	TableName           string
	Currency            string
	GroupByFields       []string
	SortOrder           string
	Filters             map[string]interface{}
	TransformationRules map[string]string

	// Feature toggles
	EnableML       bool
	EnableCaching  bool
	StrictTypes    bool
	EnableParallel bool

	// ML and forecasting options
	ForecastDays           int
	EnableAnomalyDetection bool
	EnableTrendAnalysis    bool
	EnablePredictions      bool

	// Performance options
	MaxConcurrency int
	CacheTTL       time.Duration
	TimeFormat     string
}

// AnalyticsResults holds the results of analytics operations
type AnalyticsResults struct {
	Timestamp        time.Time              `json:"timestamp"`
	AnalyticsType    string                 `json:"analytics_type"`
	TableName        string                 `json:"table_name"`
	FiltersCount     int                    `json:"filters_count"`
	AnalysisResult   map[string]interface{} `json:"analysis_result"`
	ComparisonResult map[string]interface{} `json:"comparison_result,omitempty"`
	ForecastResult   map[string]interface{} `json:"forecast_result,omitempty"`
	TypeSafetyDemo   map[string]interface{} `json:"type_safety_demo,omitempty"`
}

// AnalyticsStatus represents the status of analytics operation
type AnalyticsStatus string

const (
	AnalyticsStatusPending   AnalyticsStatus = "pending"
	AnalyticsStatusRunning   AnalyticsStatus = "running"
	AnalyticsStatusCompleted AnalyticsStatus = "completed"
	AnalyticsStatusFailed    AnalyticsStatus = "failed"
	AnalyticsStatusCancelled AnalyticsStatus = "cancelled"
)

// String returns string representation of AnalyticsStatus
func (s AnalyticsStatus) String() string {
	return string(s)
}

// NOTE: Only a String() helper is retained. Additional IsValid style helpers are
// intentionally omitted to avoid unused dead code; validation occurs where flags
// or JSON payloads are parsed. Add explicit checks there if new statuses appear.

// AnalyticsJob represents a running analytics job
type AnalyticsJob struct {
	ID        string                 `json:"id"`
	Status    AnalyticsStatus        `json:"status"`
	Options   *AnalyticsOptions      `json:"options"`
	Results   *AnalyticsResults      `json:"results,omitempty"`
	Error     string                 `json:"error,omitempty"`
	StartTime time.Time              `json:"start_time"`
	EndTime   *time.Time             `json:"end_time,omitempty"`
	Progress  float64                `json:"progress"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}
