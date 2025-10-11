package types

import (
	"testing"
	"time"
)

func TestAnalyticsStatusString(t *testing.T) {
	tests := []struct {
		status   AnalyticsStatus
		expected string
	}{
		{AnalyticsStatusPending, "pending"},
		{AnalyticsStatusRunning, "running"},
		{AnalyticsStatusCompleted, "completed"},
		{AnalyticsStatusFailed, "failed"},
		{AnalyticsStatusCancelled, "cancelled"},
	}

	for _, test := range tests {
		if test.status.String() != test.expected {
			t.Errorf("Expected %s.String() to be '%s', got '%s'",
				test.status, test.expected, test.status.String())
		}
	}
}

func TestAnalyticsOptions(t *testing.T) {
	opts := &AnalyticsOptions{
		TableName:              "cost_data",
		Currency:               "USD",
		GroupByFields:          []string{"service", "region"},
		SortOrder:              "desc",
		Filters:                map[string]interface{}{"region": "us-east-1"},
		TransformationRules:    map[string]string{"currency": "normalize"},
		EnableML:               true,
		EnableCaching:          true,
		StrictTypes:            false,
		EnableParallel:         true,
		ForecastDays:           30,
		EnableAnomalyDetection: true,
		EnableTrendAnalysis:    true,
		EnablePredictions:      true,
		MaxConcurrency:         10,
		CacheTTL:               time.Hour,
		TimeFormat:             "2006-01-02",
	}

	// Test that all fields are properly set
	if opts.TableName != "cost_data" {
		t.Errorf("Expected TableName 'cost_data', got '%s'", opts.TableName)
	}

	if opts.Currency != "USD" {
		t.Errorf("Expected Currency 'USD', got '%s'", opts.Currency)
	}

	if len(opts.GroupByFields) != 2 {
		t.Errorf("Expected 2 GroupByFields, got %d", len(opts.GroupByFields))
	}

	if opts.GroupByFields[0] != "service" || opts.GroupByFields[1] != "region" {
		t.Errorf("Expected GroupByFields ['service', 'region'], got %v", opts.GroupByFields)
	}

	if opts.SortOrder != "desc" {
		t.Errorf("Expected SortOrder 'desc', got '%s'", opts.SortOrder)
	}

	if len(opts.Filters) != 1 {
		t.Errorf("Expected 1 filter, got %d", len(opts.Filters))
	}

	if opts.Filters["region"] != "us-east-1" {
		t.Errorf("Expected filter region 'us-east-1', got %v", opts.Filters["region"])
	}

	if len(opts.TransformationRules) != 1 {
		t.Errorf("Expected 1 transformation rule, got %d", len(opts.TransformationRules))
	}

	if opts.TransformationRules["currency"] != "normalize" {
		t.Errorf("Expected transformation rule currency 'normalize', got %v", opts.TransformationRules["currency"])
	}

	if !opts.EnableML {
		t.Error("Expected EnableML to be true")
	}

	if !opts.EnableCaching {
		t.Error("Expected EnableCaching to be true")
	}

	if opts.StrictTypes {
		t.Error("Expected StrictTypes to be false")
	}

	if !opts.EnableParallel {
		t.Error("Expected EnableParallel to be true")
	}

	if opts.ForecastDays != 30 {
		t.Errorf("Expected ForecastDays 30, got %d", opts.ForecastDays)
	}

	if !opts.EnableAnomalyDetection {
		t.Error("Expected EnableAnomalyDetection to be true")
	}

	if !opts.EnableTrendAnalysis {
		t.Error("Expected EnableTrendAnalysis to be true")
	}

	if !opts.EnablePredictions {
		t.Error("Expected EnablePredictions to be true")
	}

	if opts.MaxConcurrency != 10 {
		t.Errorf("Expected MaxConcurrency 10, got %d", opts.MaxConcurrency)
	}

	if opts.CacheTTL != time.Hour {
		t.Errorf("Expected CacheTTL 1h, got %v", opts.CacheTTL)
	}

	if opts.TimeFormat != "2006-01-02" {
		t.Errorf("Expected TimeFormat '2006-01-02', got '%s'", opts.TimeFormat)
	}
}

func TestAnalyticsResults(t *testing.T) {
	now := time.Now()
	results := &AnalyticsResults{
		Timestamp:     now,
		AnalyticsType: "basic_analysis",
		TableName:     "cost_data",
		FiltersCount:  1,
		AnalysisResult: map[string]interface{}{
			"total_cost": 1234.56,
			"currency":   "USD",
		},
		ComparisonResult: map[string]interface{}{
			"change_percent": 12.2,
			"trend":          "increasing",
		},
		ForecastResult: map[string]interface{}{
			"predicted_cost": 1456.78,
			"confidence":     0.85,
		},
		TypeSafetyDemo: map[string]interface{}{
			"demo_field": "demo_value",
		},
	}

	if results.Timestamp != now {
		t.Errorf("Expected Timestamp %v, got %v", now, results.Timestamp)
	}

	if results.AnalyticsType != "basic_analysis" {
		t.Errorf("Expected AnalyticsType 'basic_analysis', got '%s'", results.AnalyticsType)
	}

	if results.TableName != "cost_data" {
		t.Errorf("Expected TableName 'cost_data', got '%s'", results.TableName)
	}

	if results.FiltersCount != 1 {
		t.Errorf("Expected FiltersCount 1, got %d", results.FiltersCount)
	}

	if len(results.AnalysisResult) != 2 {
		t.Errorf("Expected 2 analysis results, got %d", len(results.AnalysisResult))
	}

	if results.AnalysisResult["total_cost"] != 1234.56 {
		t.Errorf("Expected total_cost 1234.56, got %v", results.AnalysisResult["total_cost"])
	}

	if len(results.ComparisonResult) != 2 {
		t.Errorf("Expected 2 comparison results, got %d", len(results.ComparisonResult))
	}

	if len(results.ForecastResult) != 2 {
		t.Errorf("Expected 2 forecast results, got %d", len(results.ForecastResult))
	}

	if len(results.TypeSafetyDemo) != 1 {
		t.Errorf("Expected 1 type safety demo field, got %d", len(results.TypeSafetyDemo))
	}
}

func TestAnalyticsJob(t *testing.T) {
	now := time.Now()
	endTime := now.Add(time.Minute)

	job := &AnalyticsJob{
		ID:     "test-job-001",
		Status: AnalyticsStatusCompleted,
		Options: &AnalyticsOptions{
			TableName: "cost_data",
			Currency:  "USD",
		},
		Results: &AnalyticsResults{
			Timestamp:     now,
			AnalyticsType: "basic_analysis",
			TableName:     "cost_data",
			FiltersCount:  0,
		},
		Error:     "",
		StartTime: now,
		EndTime:   &endTime,
		Progress:  100.0,
		Metadata: map[string]interface{}{
			"worker_id": "worker-001",
			"priority":  "high",
		},
	}

	if job.ID != "test-job-001" {
		t.Errorf("Expected ID 'test-job-001', got '%s'", job.ID)
	}

	if job.Status != AnalyticsStatusCompleted {
		t.Errorf("Expected Status %s, got %s", AnalyticsStatusCompleted, job.Status)
	}

	if job.Options == nil {
		t.Error("Options should not be nil")
	}

	if job.Results == nil {
		t.Error("Results should not be nil")
	}

	if job.Error != "" {
		t.Errorf("Expected empty Error, got '%s'", job.Error)
	}

	if job.StartTime != now {
		t.Errorf("Expected StartTime %v, got %v", now, job.StartTime)
	}

	if job.EndTime == nil {
		t.Error("EndTime should not be nil")
	} else if *job.EndTime != endTime {
		t.Errorf("Expected EndTime %v, got %v", endTime, *job.EndTime)
	}

	if job.Progress != 100.0 {
		t.Errorf("Expected Progress 100.0, got %f", job.Progress)
	}

	if len(job.Metadata) != 2 {
		t.Errorf("Expected 2 metadata fields, got %d", len(job.Metadata))
	}

	if job.Metadata["worker_id"] != "worker-001" {
		t.Errorf("Expected worker_id 'worker-001', got %v", job.Metadata["worker_id"])
	}
}
