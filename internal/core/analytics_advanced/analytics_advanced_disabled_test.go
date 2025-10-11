//go:build !experimental
// +build !experimental

package analytics_advanced

import "testing"

// TestDisabledService verifies that the non-experimental stub returns ErrDisabled
// and emits the expected sentinel error across all interface methods. This guards
// against accidental behavior changes when refactoring the stub or metrics layer.
func TestDisabledService_AllMethods(t *testing.T) {
	svc := NewAdvancedAnalyticsService()
	if svc == nil {
		t.Fatalf("expected service instance")
	}
	if _, err := svc.RunMLForecast(&ForecastRequest{Days: 7}); err == nil || err != ErrDisabled {
		t.Fatalf("RunMLForecast expected ErrDisabled got %v", err)
	}
	if _, err := svc.DetectAnomalies(&AnomalyDetectionRequest{}); err == nil || err != ErrDisabled {
		t.Fatalf("DetectAnomalies expected ErrDisabled got %v", err)
	}
	if _, err := svc.RunAdvancedOptimization(&OptimizationRequest{}); err == nil || err != ErrDisabled {
		t.Fatalf("RunAdvancedOptimization expected ErrDisabled got %v", err)
	}
	if _, err := svc.TrainCustomModel(&ModelTrainingRequest{}); err == nil || err != ErrDisabled {
		t.Fatalf("TrainCustomModel expected ErrDisabled got %v", err)
	}
	if _, err := svc.StartStreamProcessing(&StreamingRequest{}); err == nil || err != ErrDisabled {
		t.Fatalf("StartStreamProcessing expected ErrDisabled got %v", err)
	}
	if _, err := svc.RunCustomAnalytics(&CustomAnalyticsRequest{}); err == nil || err != ErrDisabled {
		t.Fatalf("RunCustomAnalytics expected ErrDisabled got %v", err)
	}
}
