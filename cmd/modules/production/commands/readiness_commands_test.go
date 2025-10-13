package commands

import (
	"context"
	"testing"
	"time"

	"github.com/costscope/costscope/internal/core/logging"
	"github.com/costscope/costscope/internal/core/production"
	"github.com/costscope/costscope/internal/providers"
)

// TestProductionReadinessCommands tests the production readiness commands
func TestProductionReadinessCommands(t *testing.T) {
	logger := logging.NewLogger("test")
	providerManager := &providers.ProviderManager{}

	commands := NewProductionReadinessCommands(logger, providerManager)

	if commands == nil {
		t.Fatal("Expected commands to be created, got nil")
	}

	if commands.logger == nil {
		t.Fatal("Expected logger to be set")
	}

	if commands.providerManager == nil {
		t.Fatal("Expected provider manager to be set")
	}

	if commands.productionSvc == nil {
		t.Fatal("Expected production service to be set")
	}
}

// TestBuildProductionReadinessCommands tests command building
func TestBuildProductionReadinessCommands(t *testing.T) {
	logger := logging.NewLogger("test")
	providerManager := &providers.ProviderManager{}

	commands := NewProductionReadinessCommands(logger, providerManager)
	cmd := commands.BuildProductionReadinessCommands()

	if cmd == nil {
		t.Fatal("Expected command to be created, got nil")
	}

	if cmd.Use != "prod-readiness" {
		t.Errorf("Expected command use to be 'prod-readiness', got '%s'", cmd.Use)
	}

	// Check that subcommands are added
	subcommands := cmd.Commands()

	t.Logf("Found %d subcommands:", len(subcommands))
	for _, subCmd := range subcommands {
		t.Logf("  - %s", subCmd.Use)
	}

	// We expect at least some commands to be added
	if len(subcommands) == 0 {
		t.Error("Expected at least some subcommands to be added")
	}
}

// TestCalculateOverallScore tests the score calculation
func TestCalculateOverallScore(t *testing.T) {
	assessment := &production.DeploymentReadinessAssessment{
		ReadinessScore: 90,
	}

	status := &production.ProductionSystemMetrics{
		SystemHealth: production.SystemHealthStatus{
			HealthScore: 85,
		},
		Performance: production.PerformanceMetrics{
			OptimizationScore: 80,
		},
		Security: production.SecurityMetrics{
			SecurityScore: 75,
		},
		Integration: production.IntegrationMetrics{
			IntegrationScore: 70,
		},
	}

	score := calculateOverallScore(assessment, status)

	// Expected calculation:
	// 90 * 0.3 + 85 * 0.25 + 80 * 0.2 + 75 * 0.15 + 70 * 0.1
	// = 27 + 21.25 + 16 + 11.25 + 7 = 82.5 ≈ 82
	expected := 82

	if score != expected {
		t.Errorf("Expected score %d, got %d", expected, score)
	}
}

// TestDetermineReadinessLevel tests readiness level determination
func TestDetermineReadinessLevel(t *testing.T) {
	testCases := []struct {
		name          string
		score         int
		expectedLevel string
	}{
		{"Optimized", 95, ReadinessLevelOptimized},
		{"Ready", 80, ReadinessLevelReady},
		{"Partial", 60, ReadinessLevelPartial},
		{"Not Ready", 30, ReadinessLevelNotReady},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assessment := &production.DeploymentReadinessAssessment{
				ReadinessScore: tc.score,
			}
			status := &production.ProductionSystemMetrics{
				SystemHealth: production.SystemHealthStatus{HealthScore: tc.score},
				Performance:  production.PerformanceMetrics{OptimizationScore: tc.score},
				Security:     production.SecurityMetrics{SecurityScore: tc.score},
				Integration:  production.IntegrationMetrics{IntegrationScore: tc.score},
			}

			level := determineReadinessLevel(assessment, status)
			if level != tc.expectedLevel {
				t.Errorf("Expected level %s, got %s", tc.expectedLevel, level)
			}
		})
	}
}

// TestCombineCriticalIssues tests critical issues combination
func TestCombineCriticalIssues(t *testing.T) {
	assessment := &production.DeploymentReadinessAssessment{
		CriticalIssues: []production.Issue{
			{Title: "Database connection issue"},
			{Title: "Security vulnerability"},
		},
	}

	status := &production.ProductionSystemMetrics{
		CriticalIssues: []string{
			"High memory usage",
			"Security vulnerability", // Duplicate
		},
	}

	issues := combineCriticalIssues(assessment, status)

	// Should have 3 unique issues
	expected := 3
	if len(issues) != expected {
		t.Errorf("Expected %d issues, got %d", expected, len(issues))
	}

	// Verify deduplication
	issueMap := make(map[string]bool)
	for _, issue := range issues {
		if issueMap[issue] {
			t.Errorf("Duplicate issue found: %s", issue)
		}
		issueMap[issue] = true
	}
}

// TestCalculateOverallHealthStatus tests health status calculation
func TestCalculateOverallHealthStatus(t *testing.T) {
	testCases := []struct {
		name           string
		healthChecks   map[string]production.CheckResult
		validation     *production.ValidationResult
		expectedStatus string
	}{
		{
			name: "All healthy",
			healthChecks: map[string]production.CheckResult{
				"api": {Status: "passed"},
				"db":  {Status: "passed"},
			},
			validation:     &production.ValidationResult{Valid: true},
			expectedStatus: HealthStatusHealthy,
		},
		{
			name: "Some warnings",
			healthChecks: map[string]production.CheckResult{
				"api": {Status: "passed"},
				"db":  {Status: "warning"},
			},
			validation:     &production.ValidationResult{Valid: true},
			expectedStatus: HealthStatusWarning,
		},
		{
			name: "Critical failure",
			healthChecks: map[string]production.CheckResult{
				"api": {Status: "failed"},
				"db":  {Status: "passed"},
			},
			validation:     &production.ValidationResult{Valid: true},
			expectedStatus: HealthStatusCritical,
		},
		{
			name: "Invalid validation",
			healthChecks: map[string]production.CheckResult{
				"api": {Status: "passed"},
			},
			validation:     &production.ValidationResult{Valid: false},
			expectedStatus: HealthStatusCritical,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			status := calculateOverallHealthStatus(tc.healthChecks, tc.validation)
			if status != tc.expectedStatus {
				t.Errorf("Expected status %s, got %s", tc.expectedStatus, status)
			}
		})
	}
}

// TestGenerateDeploymentSteps tests deployment steps generation
func TestGenerateDeploymentSteps(t *testing.T) {
	testCases := []struct {
		strategy      string
		expectedSteps int
	}{
		{"blue-green", 6},
		{"rolling", 5},
		{"canary", 5},
		{"default", 2},
	}

	for _, tc := range testCases {
		t.Run(tc.strategy, func(t *testing.T) {
			steps := generateDeploymentSteps(tc.strategy)
			if len(steps) != tc.expectedSteps {
				t.Errorf("Expected %d steps for %s strategy, got %d",
					tc.expectedSteps, tc.strategy, len(steps))
			}

			// Verify order is sequential
			for i, step := range steps {
				if step.Order != i+1 {
					t.Errorf("Expected step order %d, got %d", i+1, step.Order)
				}
			}
		})
	}
}

// TestMetricsCollectionResult tests metrics collection result structure
func TestMetricsCollectionResult(t *testing.T) {
	now := time.Now()
	duration := 100 * time.Millisecond

	result := &MetricsCollectionResult{
		Type:      "health",
		Timestamp: now,
		SystemHealth: &production.SystemHealthStatus{
			Status:      "healthy",
			HealthScore: 95,
		},
		CollectionDuration: duration,
	}

	if result.Type != "health" {
		t.Errorf("Expected type 'health', got '%s'", result.Type)
	}

	if !result.Timestamp.Equal(now) {
		t.Errorf("Expected timestamp %v, got %v", now, result.Timestamp)
	}

	if result.SystemHealth == nil {
		t.Error("Expected system health to be set")
	}

	if result.SystemHealth.HealthScore != 95 {
		t.Errorf("Expected health score 95, got %d", result.SystemHealth.HealthScore)
	}

	if result.CollectionDuration != duration {
		t.Errorf("Expected collection duration %v, got %v", duration, result.CollectionDuration)
	}
}

// TestGetReportSections tests report sections retrieval
func TestGetReportSections(t *testing.T) {
	testCases := []struct {
		reportType       string
		expectedSections []string
	}{
		{
			"executive",
			[]string{"overview", "summary", "recommendations", "roadmap"},
		},
		{
			"technical",
			[]string{"overview", "metrics", "detailed_analysis", "recommendations", "appendix"},
		},
		{
			"security",
			[]string{"security_overview", "compliance", "vulnerabilities", "recommendations"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.reportType, func(t *testing.T) {
			sections := getReportSections(tc.reportType)
			if len(sections) != len(tc.expectedSections) {
				t.Errorf("Expected %d sections, got %d",
					len(tc.expectedSections), len(sections))
			}

			for i, expected := range tc.expectedSections {
				if sections[i] != expected {
					t.Errorf("Expected section '%s', got '%s'", expected, sections[i])
				}
			}
		})
	}
}

// BenchmarkCalculateOverallScore benchmarks score calculation
func BenchmarkCalculateOverallScore(b *testing.B) {
	assessment := &production.DeploymentReadinessAssessment{
		ReadinessScore: 90,
	}

	status := &production.ProductionSystemMetrics{
		SystemHealth: production.SystemHealthStatus{HealthScore: 85},
		Performance:  production.PerformanceMetrics{OptimizationScore: 80},
		Security:     production.SecurityMetrics{SecurityScore: 75},
		Integration:  production.IntegrationMetrics{IntegrationScore: 70},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		calculateOverallScore(assessment, status)
	}
}

// MockProductionService for testing (would be in a separate test file in production)
type MockProductionService struct{}

func (m *MockProductionService) GetSystemStatus(ctx context.Context) (*production.ProductionSystemMetrics, error) {
	return &production.ProductionSystemMetrics{
		ReadinessScore: 85,
		SystemHealth:   production.SystemHealthStatus{HealthScore: 85},
		Performance:    production.PerformanceMetrics{OptimizationScore: 80},
		Security:       production.SecurityMetrics{SecurityScore: 75},
		Integration:    production.IntegrationMetrics{IntegrationScore: 70},
	}, nil
}

func (m *MockProductionService) RunOptimization(ctx context.Context, options *production.OptimizationOptions) (*production.ProductionOptimizationReport, error) {
	return &production.ProductionOptimizationReport{
		GeneratedAt: time.Now(),
		OptimizationResults: production.OptimizationResults{
			TotalImprovements: 5,
		},
		Recommendations: []production.ProductionRecommendation{},
	}, nil
}

func (m *MockProductionService) AssessDeploymentReadiness(ctx context.Context, environment string) (*production.DeploymentReadinessAssessment, error) {
	return &production.DeploymentReadinessAssessment{
		ReadinessScore:  90,
		ReadinessStatus: "ready",
	}, nil
}

func (m *MockProductionService) GenerateExecutiveReport(ctx context.Context, options *production.ReportOptions) (*production.ExecutiveReport, error) {
	return &production.ExecutiveReport{
		GeneratedAt: time.Now(),
	}, nil
}

func (m *MockProductionService) ValidateProductionConfiguration(ctx context.Context) (*production.ValidationResult, error) {
	return &production.ValidationResult{
		Valid: true,
		Score: 95,
	}, nil
}

func (m *MockProductionService) GetHealthChecks(ctx context.Context) (map[string]production.CheckResult, error) {
	return map[string]production.CheckResult{
		"api": {Status: "passed", Message: "API healthy"},
		"db":  {Status: "passed", Message: "Database healthy"},
	}, nil
}
