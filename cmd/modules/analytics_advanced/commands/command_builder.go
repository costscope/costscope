//go:build experimental

package commands

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/costscope/costscope/internal/core/analytics_advanced"
	"github.com/costscope/costscope/internal/core/logging"

	"github.com/spf13/cobra"
)

// AdvancedAnalyticsCommands provides CLI commands for advanced analytics
type AdvancedAnalyticsCommands struct {
	service analytics_advanced.AdvancedAnalyticsService
	logger  *logging.Logger
}

// NewAdvancedAnalyticsCommands creates a new advanced analytics commands instance
func NewAdvancedAnalyticsCommands() *AdvancedAnalyticsCommands {
	return &AdvancedAnalyticsCommands{
		service: analytics_advanced.NewAdvancedAnalyticsService(),
		logger:  logging.NewLogger(logging.LevelInfo),
	}
}

// BuildAdvancedAnalyticsCommand creates the main advanced analytics command
func (aac *AdvancedAnalyticsCommands) BuildAdvancedAnalyticsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "analytics-advanced",
		Short: "Advanced analytics with ML, real-time processing, and predictive insights",
		Long: `Advanced Analytics provides sophisticated cost analysis capabilities including:

• Machine Learning-powered forecasting and anomaly detection
• Real-time streaming analytics and alerts
• Predictive scaling recommendations 
• Advanced optimization algorithms
• Custom model training and deployment
• Type-safe filtering and complex transformations`,
		Example: `  # Run ML-powered cost forecasting
  costscope analytics-advanced forecast --model auto-arima --days 90 --confidence 95

  # Real-time anomaly detection
  costscope analytics-advanced detect --stream --sensitivity high --alerts slack

  # Custom optimization analysis
  costscope analytics-advanced optimize --algorithm genetic --target cost --constraints policy.json

  # Train custom ML model
  costscope analytics-advanced train --model custom-lstm --data historical.csv --features auto`,
	}

	// Add subcommands
	cmd.AddCommand(aac.buildForecastCommand())
	cmd.AddCommand(aac.buildDetectCommand())
	cmd.AddCommand(aac.buildOptimizeCommand())
	cmd.AddCommand(aac.buildTrainCommand())
	cmd.AddCommand(aac.buildStreamCommand())
	cmd.AddCommand(aac.buildCustomCommand())

	return cmd
}

// buildForecastCommand creates the ML-powered forecasting command
func (aac *AdvancedAnalyticsCommands) buildForecastCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "forecast",
		Short: "ML-powered cost forecasting with multiple models",
		Long:  "Generate accurate cost forecasts using advanced machine learning models",
		RunE:  aac.handleForecast,
	}

	cmd.Flags().String("model", "auto-arima", "ML model (auto-arima, lstm, prophet, ensemble)")
	cmd.Flags().Int("days", 30, "Forecast period in days")
	cmd.Flags().Float64("confidence", 95.0, "Confidence interval percentage")
	cmd.Flags().String("features", "auto", "Features to include (auto, cost, usage, time)")
	cmd.Flags().String("seasonality", "auto", "Seasonality detection (auto, daily, weekly, monthly)")
	cmd.Flags().Bool("uncertainty", true, "Include uncertainty quantification")
	cmd.Flags().String("output", "table", "Output format (table, json, csv)")

	return cmd
}

// buildDetectCommand creates the anomaly detection command
func (aac *AdvancedAnalyticsCommands) buildDetectCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "detect",
		Short: "Real-time anomaly detection and alerting",
		Long:  "Detect cost anomalies using statistical and ML-based methods",
		RunE:  aac.handleDetect,
	}

	cmd.Flags().String("method", "isolation-forest", "Detection method (isolation-forest, one-class-svm, statistical)")
	cmd.Flags().String("sensitivity", "medium", "Detection sensitivity (low, medium, high)")
	cmd.Flags().Bool("stream", false, "Enable real-time streaming detection")
	cmd.Flags().String("alerts", "", "Alert channels (slack, email, webhook)")
	cmd.Flags().Float64("threshold", 0.95, "Anomaly score threshold")
	cmd.Flags().String("window", "7d", "Detection window (1h, 1d, 7d, 30d)")

	return cmd
}

// buildOptimizeCommand creates the advanced optimization command
func (aac *AdvancedAnalyticsCommands) buildOptimizeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "optimize",
		Short: "Advanced cost optimization algorithms",
		Long:  "Apply sophisticated optimization algorithms for cost reduction",
		RunE:  aac.handleOptimize,
	}

	cmd.Flags().String("algorithm", "genetic", "Optimization algorithm (genetic, particle-swarm, simulated-annealing)")
	cmd.Flags().String("target", "cost", "Optimization target (cost, performance, both)")
	cmd.Flags().String("constraints", "", "Constraints file (JSON)")
	cmd.Flags().Int("iterations", 1000, "Maximum iterations")
	cmd.Flags().Float64("tolerance", 0.001, "Convergence tolerance")
	cmd.Flags().Bool("multi-objective", false, "Multi-objective optimization")

	return cmd
}

// buildTrainCommand creates the model training command
func (aac *AdvancedAnalyticsCommands) buildTrainCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "train",
		Short: "Train custom ML models",
		Long:  "Train and deploy custom machine learning models",
		RunE:  aac.handleTrain,
	}

	cmd.Flags().String("model", "lstm", "Model type (lstm, transformer, gradient-boost)")
	cmd.Flags().String("data", "", "Training data file")
	cmd.Flags().String("features", "auto", "Feature selection strategy")
	cmd.Flags().Float64("validation-split", 0.2, "Validation data split ratio")
	cmd.Flags().Int("epochs", 100, "Training epochs")
	cmd.Flags().String("name", "", "Model name for deployment")

	return cmd
}

// buildStreamCommand creates the real-time streaming command
func (aac *AdvancedAnalyticsCommands) buildStreamCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stream",
		Short: "Real-time streaming analytics",
		Long:  "Process real-time cost data streams with low latency",
		RunE:  aac.handleStream,
	}

	cmd.Flags().String("source", "kafka", "Stream source (kafka, kinesis, pubsub)")
	cmd.Flags().String("topic", "cost-events", "Stream topic/channel")
	cmd.Flags().String("window", "5m", "Processing window (1m, 5m, 15m, 1h)")
	cmd.Flags().String("aggregation", "sum", "Aggregation function (sum, avg, max, count)")
	cmd.Flags().Bool("persist", true, "Persist results to database")

	return cmd
}

// buildCustomCommand creates the custom analysis command
func (aac *AdvancedAnalyticsCommands) buildCustomCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "custom",
		Short: "Custom analytics with Python/R integration",
		Long:  "Execute custom analytics scripts with full Python/R integration",
		RunE:  aac.handleCustom,
	}

	cmd.Flags().String("script", "", "Custom script file (.py, .r)")
	cmd.Flags().String("environment", "python", "Execution environment (python, r, julia)")
	cmd.Flags().String("parameters", "", "Script parameters (JSON)")
	cmd.Flags().String("libraries", "", "Required libraries (comma-separated)")
	cmd.Flags().Bool("gpu", false, "Enable GPU acceleration")

	return cmd
}

// Command handlers
func (aac *AdvancedAnalyticsCommands) handleForecast(cmd *cobra.Command, args []string) error {
	model, _ := cmd.Flags().GetString("model")
	days, _ := cmd.Flags().GetInt("days")
	confidence, _ := cmd.Flags().GetFloat64("confidence")
	features, _ := cmd.Flags().GetString("features")
	seasonality, _ := cmd.Flags().GetString("seasonality")
	uncertainty, _ := cmd.Flags().GetBool("uncertainty")
	outputFormat, _ := cmd.Flags().GetString("output")

	aac.logger.Info(fmt.Sprintf("Starting advanced ML forecasting with model: %s, days: %d", model, days))

	request := &analytics_advanced.ForecastRequest{
		Model:       model,
		Days:        days,
		Confidence:  confidence,
		Features:    strings.Split(features, ","),
		Seasonality: seasonality,
		Uncertainty: uncertainty,
	}

	result, err := aac.service.RunMLForecast(request)
	if err != nil {
		return fmt.Errorf("forecast failed: %w", err)
	}

	return aac.outputForecastResult(result, outputFormat)
}

func (aac *AdvancedAnalyticsCommands) handleDetect(cmd *cobra.Command, args []string) error {
	method, _ := cmd.Flags().GetString("method")
	sensitivity, _ := cmd.Flags().GetString("sensitivity")
	stream, _ := cmd.Flags().GetBool("stream")
	alerts, _ := cmd.Flags().GetString("alerts")
	threshold, _ := cmd.Flags().GetFloat64("threshold")
	window, _ := cmd.Flags().GetString("window")

	aac.logger.Info(fmt.Sprintf("Starting anomaly detection with method: %s, sensitivity: %s", method, sensitivity))

	request := &analytics_advanced.AnomalyDetectionRequest{
		Method:      method,
		Sensitivity: sensitivity,
		Stream:      stream,
		Alerts:      strings.Split(alerts, ","),
		Threshold:   threshold,
		Window:      window,
	}

	result, err := aac.service.DetectAnomalies(request)
	if err != nil {
		return fmt.Errorf("anomaly detection failed: %w", err)
	}

	return aac.outputAnomalyResult(result)
}

func (aac *AdvancedAnalyticsCommands) handleOptimize(cmd *cobra.Command, args []string) error {
	algorithm, _ := cmd.Flags().GetString("algorithm")
	target, _ := cmd.Flags().GetString("target")
	constraints, _ := cmd.Flags().GetString("constraints")
	iterations, _ := cmd.Flags().GetInt("iterations")
	tolerance, _ := cmd.Flags().GetFloat64("tolerance")
	multiObjective, _ := cmd.Flags().GetBool("multi-objective")

	aac.logger.Info(fmt.Sprintf("Starting advanced optimization with algorithm: %s, target: %s", algorithm, target))

	request := &analytics_advanced.OptimizationRequest{
		Algorithm:      algorithm,
		Target:         target,
		Constraints:    constraints,
		Iterations:     iterations,
		Tolerance:      tolerance,
		MultiObjective: multiObjective,
	}

	result, err := aac.service.RunAdvancedOptimization(request)
	if err != nil {
		return fmt.Errorf("optimization failed: %w", err)
	}

	return aac.outputOptimizationResult(result)
}

func (aac *AdvancedAnalyticsCommands) handleTrain(cmd *cobra.Command, args []string) error {
	modelType, _ := cmd.Flags().GetString("model")
	data, _ := cmd.Flags().GetString("data")
	features, _ := cmd.Flags().GetString("features")
	validationSplit, _ := cmd.Flags().GetFloat64("validation-split")
	epochs, _ := cmd.Flags().GetInt("epochs")
	name, _ := cmd.Flags().GetString("name")

	aac.logger.Info(fmt.Sprintf("Starting model training with model: %s, data: %s", modelType, data))

	request := &analytics_advanced.ModelTrainingRequest{
		ModelType:       modelType,
		DataFile:        data,
		Features:        features,
		ValidationSplit: validationSplit,
		Epochs:          epochs,
		Name:            name,
	}

	result, err := aac.service.TrainCustomModel(request)
	if err != nil {
		return fmt.Errorf("model training failed: %w", err)
	}

	return aac.outputTrainingResult(result)
}

func (aac *AdvancedAnalyticsCommands) handleStream(cmd *cobra.Command, args []string) error {
	source, _ := cmd.Flags().GetString("source")
	topic, _ := cmd.Flags().GetString("topic")
	window, _ := cmd.Flags().GetString("window")
	aggregation, _ := cmd.Flags().GetString("aggregation")
	persist, _ := cmd.Flags().GetBool("persist")

	aac.logger.Info(fmt.Sprintf("Starting streaming analytics with source: %s, topic: %s", source, topic))

	request := &analytics_advanced.StreamingRequest{
		Source:      source,
		Topic:       topic,
		Window:      window,
		Aggregation: aggregation,
		Persist:     persist,
	}

	result, err := aac.service.StartStreamProcessing(request)
	if err != nil {
		return fmt.Errorf("streaming failed: %w", err)
	}

	return aac.outputStreamingResult(result)
}

func (aac *AdvancedAnalyticsCommands) handleCustom(cmd *cobra.Command, args []string) error {
	script, _ := cmd.Flags().GetString("script")
	environment, _ := cmd.Flags().GetString("environment")
	parameters, _ := cmd.Flags().GetString("parameters")
	libraries, _ := cmd.Flags().GetString("libraries")
	gpu, _ := cmd.Flags().GetBool("gpu")

	aac.logger.Info(fmt.Sprintf("Running custom analytics with script: %s, environment: %s", script, environment))

	var params map[string]interface{}
	if parameters != "" {
		if err := json.Unmarshal([]byte(parameters), &params); err != nil {
			return fmt.Errorf("invalid parameters JSON: %w", err)
		}
	}

	request := &analytics_advanced.CustomAnalyticsRequest{
		Script:      script,
		Environment: environment,
		Parameters:  params,
		Libraries:   strings.Split(libraries, ","),
		GPU:         gpu,
	}

	result, err := aac.service.RunCustomAnalytics(request)
	if err != nil {
		return fmt.Errorf("custom analytics failed: %w", err)
	}

	return aac.outputCustomResult(result)
}

// Output formatting methods
func (aac *AdvancedAnalyticsCommands) outputForecastResult(result *analytics_advanced.ForecastResult, format string) error {
	switch format {
	case "json":
		data, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
	case "table":
		aac.printForecastTable(result)
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}
	return nil
}

func (aac *AdvancedAnalyticsCommands) outputAnomalyResult(result *analytics_advanced.AnomalyDetectionResult) error {
	fmt.Printf("═══════════════════════════════════════════════════\n")
	fmt.Printf("              ANOMALY DETECTION RESULTS\n")
	fmt.Printf("═══════════════════════════════════════════════════\n")
	fmt.Printf("Total Anomalies Found: %d\n", len(result.Anomalies))
	fmt.Printf("Detection Method:      %s\n", result.Method)
	fmt.Printf("Sensitivity Level:     %s\n", result.Sensitivity)
	fmt.Printf("Average Score:         %.3f\n", result.AverageScore)
	fmt.Printf("Processing Time:       %dms\n", result.ProcessingTimeMs)

	if len(result.Anomalies) > 0 {
		fmt.Printf("\n--- TOP ANOMALIES ---\n")
		for i, anomaly := range result.Anomalies {
			if i >= 5 { // Show top 5
				break
			}
			fmt.Printf("• [%.3f] %s: %s\n", anomaly.Score, anomaly.Timestamp, anomaly.Description)
		}
	}
	return nil
}

func (aac *AdvancedAnalyticsCommands) outputOptimizationResult(result *analytics_advanced.OptimizationResult) error {
	fmt.Printf("═══════════════════════════════════════════════════\n")
	fmt.Printf("            ADVANCED OPTIMIZATION RESULTS\n")
	fmt.Printf("═══════════════════════════════════════════════════\n")
	fmt.Printf("Algorithm:            %s\n", result.Algorithm)
	fmt.Printf("Optimization Target:  %s\n", result.Target)
	fmt.Printf("Iterations:           %d\n", result.Iterations)
	fmt.Printf("Convergence:          %.6f\n", result.Convergence)
	fmt.Printf("Improvement:          %.2f%%\n", result.ImprovementPercent)
	fmt.Printf("Estimated Savings:    $%.2f\n", result.EstimatedSavings)

	if len(result.Recommendations) > 0 {
		fmt.Printf("\n--- OPTIMIZATION RECOMMENDATIONS ---\n")
		for _, rec := range result.Recommendations {
			fmt.Printf("• %s: %s (Impact: %.1f%%)\n", rec.Category, rec.Description, rec.Impact)
		}
	}
	return nil
}

func (aac *AdvancedAnalyticsCommands) outputTrainingResult(result *analytics_advanced.ModelTrainingResult) error {
	fmt.Printf("═══════════════════════════════════════════════════\n")
	fmt.Printf("              MODEL TRAINING RESULTS\n")
	fmt.Printf("═══════════════════════════════════════════════════\n")
	fmt.Printf("Model Type:           %s\n", result.ModelType)
	fmt.Printf("Training Accuracy:    %.3f\n", result.TrainingAccuracy)
	fmt.Printf("Validation Accuracy:  %.3f\n", result.ValidationAccuracy)
	fmt.Printf("Training Time:        %s\n", result.TrainingTime)
	fmt.Printf("Model Size:           %s\n", result.ModelSize)
	fmt.Printf("Deployment Status:    %s\n", result.DeploymentStatus)

	if result.ModelID != "" {
		fmt.Printf("Model ID:             %s\n", result.ModelID)
	}
	return nil
}

func (aac *AdvancedAnalyticsCommands) outputStreamingResult(result *analytics_advanced.StreamingResult) error {
	fmt.Printf("═══════════════════════════════════════════════════\n")
	fmt.Printf("             STREAMING ANALYTICS STATUS\n")
	fmt.Printf("═══════════════════════════════════════════════════\n")
	fmt.Printf("Stream Source:        %s\n", result.Source)
	fmt.Printf("Processing Status:    %s\n", result.Status)
	fmt.Printf("Events Processed:     %d\n", result.EventsProcessed)
	fmt.Printf("Processing Rate:      %.1f events/sec\n", result.ProcessingRate)
	fmt.Printf("Average Latency:      %dms\n", result.LatencyMs)
	return nil
}

func (aac *AdvancedAnalyticsCommands) outputCustomResult(result *analytics_advanced.CustomAnalyticsResult) error {
	fmt.Printf("═══════════════════════════════════════════════════\n")
	fmt.Printf("             CUSTOM ANALYTICS RESULTS\n")
	fmt.Printf("═══════════════════════════════════════════════════\n")
	fmt.Printf("Script:               %s\n", result.Script)
	fmt.Printf("Environment:          %s\n", result.Environment)
	fmt.Printf("Execution Status:     %s\n", result.Status)
	fmt.Printf("Execution Time:       %s\n", result.ExecutionTime)

	if result.Output != "" {
		fmt.Printf("\n--- SCRIPT OUTPUT ---\n")
		fmt.Printf("%s\n", result.Output)
	}

	if result.Error != "" {
		fmt.Printf("\n--- ERRORS ---\n")
		fmt.Printf("%s\n", result.Error)
	}
	return nil
}

func (aac *AdvancedAnalyticsCommands) printForecastTable(result *analytics_advanced.ForecastResult) {
	fmt.Printf("═══════════════════════════════════════════════════\n")
	fmt.Printf("                ML FORECAST RESULTS\n")
	fmt.Printf("═══════════════════════════════════════════════════\n")
	fmt.Printf("Model:                %s\n", result.Model)
	fmt.Printf("Forecast Period:      %d days\n", result.Days)
	fmt.Printf("Confidence Level:     %.1f%%\n", result.Confidence)
	fmt.Printf("Model Accuracy:       %.3f\n", result.Accuracy)
	fmt.Printf("Seasonality:          %s\n", result.Seasonality)
	fmt.Printf("Trend:                %s\n", result.Trend)

	if len(result.Predictions) > 0 {
		fmt.Printf("\n--- FORECAST DATA (Next 7 Days) ---\n")
		fmt.Printf("Date          | Predicted | Lower CI  | Upper CI  \n")
		fmt.Printf("------------- | --------- | --------- | ---------\n")
		for i, pred := range result.Predictions {
			if i >= 7 { // Show first 7 days
				break
			}
			fmt.Printf("%-13s | $%8.2f | $%8.2f | $%8.2f\n",
				pred.Date, pred.Value, pred.LowerCI, pred.UpperCI)
		}
		if len(result.Predictions) > 7 {
			fmt.Printf("... and %d more days\n", len(result.Predictions)-7)
		}
	}
}
