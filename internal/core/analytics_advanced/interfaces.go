//go:build experimental
// +build experimental

package analytics_advanced

// AdvancedAnalyticsService defines the interface for advanced analytics operations
type AdvancedAnalyticsService interface {
	RunMLForecast(request *ForecastRequest) (*ForecastResult, error)
	DetectAnomalies(request *AnomalyDetectionRequest) (*AnomalyDetectionResult, error)
	RunAdvancedOptimization(request *OptimizationRequest) (*OptimizationResult, error)
	TrainCustomModel(request *ModelTrainingRequest) (*ModelTrainingResult, error)
	StartStreamProcessing(request *StreamingRequest) (*StreamingResult, error)
	RunCustomAnalytics(request *CustomAnalyticsRequest) (*CustomAnalyticsResult, error)
}

// Request types
type ForecastRequest struct {
	Model       string   `json:"model"`
	Days        int      `json:"days"`
	Confidence  float64  `json:"confidence"`
	Features    []string `json:"features"`
	Seasonality string   `json:"seasonality"`
	Uncertainty bool     `json:"uncertainty"`
}

type AnomalyDetectionRequest struct {
	Method      string   `json:"method"`
	Sensitivity string   `json:"sensitivity"`
	Stream      bool     `json:"stream"`
	Alerts      []string `json:"alerts"`
	Threshold   float64  `json:"threshold"`
	Window      string   `json:"window"`
}

type OptimizationRequest struct {
	Algorithm      string  `json:"algorithm"`
	Target         string  `json:"target"`
	Constraints    string  `json:"constraints"`
	Iterations     int     `json:"iterations"`
	Tolerance      float64 `json:"tolerance"`
	MultiObjective bool    `json:"multi_objective"`
}

type ModelTrainingRequest struct {
	ModelType       string  `json:"model_type"`
	DataFile        string  `json:"data_file"`
	Features        string  `json:"features"`
	ValidationSplit float64 `json:"validation_split"`
	Epochs          int     `json:"epochs"`
	Name            string  `json:"name"`
}

type StreamingRequest struct {
	Source      string `json:"source"`
	Topic       string `json:"topic"`
	Window      string `json:"window"`
	Aggregation string `json:"aggregation"`
	Persist     bool   `json:"persist"`
}

type CustomAnalyticsRequest struct {
	Script      string                 `json:"script"`
	Environment string                 `json:"environment"`
	Parameters  map[string]interface{} `json:"parameters"`
	Libraries   []string               `json:"libraries"`
	GPU         bool                   `json:"gpu"`
}

// Result types
type ForecastResult struct {
	Model       string               `json:"model"`
	Days        int                  `json:"days"`
	Confidence  float64              `json:"confidence"`
	Accuracy    float64              `json:"accuracy"`
	Seasonality string               `json:"seasonality"`
	Trend       string               `json:"trend"`
	Predictions []ForecastPrediction `json:"predictions"`
}

type ForecastPrediction struct {
	Date    string  `json:"date"`
	Value   float64 `json:"value"`
	LowerCI float64 `json:"lower_ci"`
	UpperCI float64 `json:"upper_ci"`
}

type AnomalyDetectionResult struct {
	Method           string    `json:"method"`
	Sensitivity      string    `json:"sensitivity"`
	AverageScore     float64   `json:"average_score"`
	ProcessingTimeMs int       `json:"processing_time_ms"`
	Anomalies        []Anomaly `json:"anomalies"`
}

type Anomaly struct {
	Timestamp   string  `json:"timestamp"`
	Score       float64 `json:"score"`
	Description string  `json:"description"`
}

type OptimizationResult struct {
	Algorithm          string                       `json:"algorithm"`
	Target             string                       `json:"target"`
	Iterations         int                          `json:"iterations"`
	Convergence        float64                      `json:"convergence"`
	ImprovementPercent float64                      `json:"improvement_percent"`
	EstimatedSavings   float64                      `json:"estimated_savings"`
	Recommendations    []OptimizationRecommendation `json:"recommendations"`
}

type OptimizationRecommendation struct {
	Category    string  `json:"category"`
	Description string  `json:"description"`
	Impact      float64 `json:"impact"`
}

type ModelTrainingResult struct {
	ModelType          string  `json:"model_type"`
	TrainingAccuracy   float64 `json:"training_accuracy"`
	ValidationAccuracy float64 `json:"validation_accuracy"`
	TrainingTime       string  `json:"training_time"`
	ModelSize          string  `json:"model_size"`
	DeploymentStatus   string  `json:"deployment_status"`
	ModelID            string  `json:"model_id"`
}

type StreamingResult struct {
	Source          string  `json:"source"`
	Status          string  `json:"status"`
	EventsProcessed int     `json:"events_processed"`
	ProcessingRate  float64 `json:"processing_rate"`
	LatencyMs       int     `json:"latency_ms"`
}

type CustomAnalyticsResult struct {
	Script        string `json:"script"`
	Environment   string `json:"environment"`
	Status        string `json:"status"`
	ExecutionTime string `json:"execution_time"`
	Output        string `json:"output"`
	Error         string `json:"error"`
}
