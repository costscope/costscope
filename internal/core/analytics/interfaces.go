package analytics

import "github.com/costscope/costscope/cmd/modules/analytics/types"

// Service defines the interface for analytics operations
type Service interface {
	// Analyze performs cost analysis with the given options
	Analyze(opts *types.AnalyticsOptions) (*types.AnalyticsResults, error)

	// Forecast generates cost forecasts using ML algorithms
	Forecast(opts *types.AnalyticsOptions) (*types.AnalyticsResults, error)

	// Compare compares costs across time periods or providers
	Compare(opts *types.AnalyticsOptions, period string, providers []string) (*types.AnalyticsResults, error)

	// Export exports analytics results to specified format and location
	Export(opts *types.AnalyticsOptions, format, output string) (string, error)
}

// Config holds configuration for analytics service
type Config struct {
	// ML and forecasting configuration
	MLEnabled           bool
	AnomalyDetection    bool
	TrendAnalysis       bool
	EnablePredictions   bool
	EnableOptimizations bool

	// Performance configuration
	EnableCaching   bool
	DefaultCacheTTL string
	MaxConcurrency  int

	// Data processing configuration
	DefaultCurrency    string
	DefaultTimeFormat  string
	StrictTypeChecking bool
}
