//go:build experimental
// +build experimental

package analytics_advanced

import (
	"strings"
	"testing"
)

// These tests intentionally validate only structural & invariant properties (sizes, value ranges)
// because the implementation produces randomized mock data intended for experimental usage.
// This guards against accidental interface/shape regressions while the module remains behind
// the `experimental` build tag for CLI exposure.

func TestRunMLForecast(t *testing.T) {
	svc := NewAdvancedAnalyticsService()
	days := 14
	res, err := svc.RunMLForecast(&ForecastRequest{Model: "auto-arima", Days: days, Confidence: 95, Seasonality: "auto", Uncertainty: true})
	if err != nil {
		t.Fatalf("RunMLForecast error: %v", err)
	}
	if res == nil {
		t.Fatalf("nil result")
	}
	if len(res.Predictions) != days {
		t.Fatalf("expected %d predictions got %d", days, len(res.Predictions))
	}
	if res.Accuracy < 0.80 || res.Accuracy > 0.99 {
		t.Fatalf("accuracy out of expected mock range: %f", res.Accuracy)
	}
	for _, p := range res.Predictions {
		if p.Value < 0 {
			t.Fatalf("prediction value negative: %+v", p)
		}
		if p.LowerCI < 0 {
			t.Fatalf("lower CI negative: %+v", p)
		}
		if p.LowerCI > p.UpperCI {
			t.Fatalf("lower CI > upper CI: %+v", p)
		}
	}
}

func TestDetectAnomalies(t *testing.T) {
	svc := NewAdvancedAnalyticsService()
	res, err := svc.DetectAnomalies(&AnomalyDetectionRequest{Method: "isolation-forest", Sensitivity: "medium", Threshold: 0.9, Window: "7d"})
	if err != nil {
		t.Fatalf("DetectAnomalies error: %v", err)
	}
	if res == nil {
		t.Fatalf("nil result")
	}
	if len(res.Anomalies) < 5 || len(res.Anomalies) > 20 { // 5 + safeRandomInt(15) => 5..19 (allow 20 defensively)
		t.Fatalf("unexpected anomalies count: %d", len(res.Anomalies))
	}
	for _, a := range res.Anomalies {
		if a.Score < 0.5 || a.Score > 1.0 {
			t.Fatalf("anomaly score out of range: %+v", a)
		}
		if a.Timestamp == "" {
			t.Fatalf("missing timestamp: %+v", a)
		}
	}
}

func TestRunAdvancedOptimization(t *testing.T) {
	svc := NewAdvancedAnalyticsService()
	res, err := svc.RunAdvancedOptimization(&OptimizationRequest{Algorithm: "genetic", Target: "cost", Iterations: 1000, Tolerance: 0.001})
	if err != nil {
		t.Fatalf("RunAdvancedOptimization error: %v", err)
	}
	if res == nil {
		t.Fatalf("nil result")
	}
	if res.ImprovementPercent < 5 || res.ImprovementPercent > 50 {
		t.Fatalf("improvement percent out of generous range: %f", res.ImprovementPercent)
	}
	if res.EstimatedSavings <= 0 {
		t.Fatalf("expected positive savings")
	}
	if len(res.Recommendations) < 3 || len(res.Recommendations) > 7 { // 3 + safeRandomInt(4) => 3..6 (allow 7 defensive)
		t.Fatalf("unexpected recommendations count: %d", len(res.Recommendations))
	}
	for _, r := range res.Recommendations {
		if r.Category == "" {
			t.Fatalf("empty recommendation category: %+v", r)
		}
		if r.Impact < 0 {
			t.Fatalf("negative impact: %+v", r)
		}
	}
}

func TestTrainCustomModel(t *testing.T) {
	svc := NewAdvancedAnalyticsService()
	res, err := svc.TrainCustomModel(&ModelTrainingRequest{ModelType: "lstm", DataFile: "data.csv", Features: "auto", ValidationSplit: 0.2, Epochs: 10, Name: "test-model"})
	if err != nil {
		t.Fatalf("TrainCustomModel error: %v", err)
	}
	if res == nil {
		t.Fatalf("nil result")
	}
	if res.ModelID == "" {
		t.Fatalf("expected model id assigned")
	}
	if res.TrainingAccuracy < 0.5 || res.TrainingAccuracy > 1.0 {
		t.Fatalf("training accuracy out of broad range: %f", res.TrainingAccuracy)
	}
	if res.ValidationAccuracy < 0.5 || res.ValidationAccuracy > 1.0 {
		t.Fatalf("validation accuracy out of broad range: %f", res.ValidationAccuracy)
	}
}

func TestStartStreamProcessing(t *testing.T) {
	svc := NewAdvancedAnalyticsService()
	res, err := svc.StartStreamProcessing(&StreamingRequest{Source: "kafka", Topic: "cost-events", Window: "5m", Aggregation: "sum", Persist: true})
	if err != nil {
		t.Fatalf("StartStreamProcessing error: %v", err)
	}
	if res == nil {
		t.Fatalf("nil result")
	}
	if res.EventsProcessed <= 0 {
		t.Fatalf("expected positive events processed")
	}
	if res.ProcessingRate <= 0 {
		t.Fatalf("expected positive processing rate")
	}
}

func TestRunCustomAnalytics(t *testing.T) {
	svc := NewAdvancedAnalyticsService()
	res, err := svc.RunCustomAnalytics(&CustomAnalyticsRequest{Script: "script.py", Environment: "python", GPU: false})
	if err != nil {
		t.Fatalf("RunCustomAnalytics error: %v", err)
	}
	if res == nil {
		t.Fatalf("nil result")
	}
	if !strings.Contains(strings.ToLower(res.Output), "execut") {
		t.Fatalf("unexpected output: %s", res.Output)
	}
	if res.Error != "" {
		t.Fatalf("expected empty error but got: %s", res.Error)
	}
}
