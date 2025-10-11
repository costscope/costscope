package commands

import (
	"context"
	"fmt"
	"strings"
	"time"

	"local/costscope/internal/core/production"
)

// Helper functions for production readiness assessment

// calculateOverallScore calculates overall readiness score
func calculateOverallScore(assessment *production.DeploymentReadinessAssessment, status *production.ProductionSystemMetrics) int {
	if assessment == nil || status == nil {
		return 0
	}

	// Weighted average of different scores
	weights := map[string]float64{
		"readiness":   0.3,
		"health":      0.25,
		"performance": 0.2,
		"security":    0.15,
		"integration": 0.1,
	}

	totalScore := 0.0
	totalScore += float64(assessment.ReadinessScore) * weights["readiness"]
	totalScore += float64(status.SystemHealth.HealthScore) * weights["health"]
	totalScore += float64(status.Performance.OptimizationScore) * weights["performance"]
	totalScore += float64(status.Security.SecurityScore) * weights["security"]
	totalScore += float64(status.Integration.IntegrationScore) * weights["integration"]

	return int(totalScore)
}

// determineReadinessLevel determines readiness level based on scores
func determineReadinessLevel(assessment *production.DeploymentReadinessAssessment, status *production.ProductionSystemMetrics) string {
	score := calculateOverallScore(assessment, status)

	switch {
	case score >= 90:
		return ReadinessLevelOptimized
	case score >= 75:
		return ReadinessLevelReady
	case score >= 50:
		return ReadinessLevelPartial
	default:
		return ReadinessLevelNotReady
	}
}

// combineCriticalIssues combines critical issues from assessment and status
func combineCriticalIssues(assessment *production.DeploymentReadinessAssessment, status *production.ProductionSystemMetrics) []string {
	var issues []string

	if assessment != nil {
		for _, issue := range assessment.CriticalIssues {
			issues = append(issues, issue.Title)
		}
	}

	if status != nil {
		issues = append(issues, status.CriticalIssues...)
	}

	return removeDuplicates(issues)
}

// getCheckedComponents returns list of checked components
func getCheckedComponents(healthChecks map[string]production.CheckResult) []string {
	var components []string
	for component := range healthChecks {
		components = append(components, component)
	}
	return components
}

// calculateOverallHealthStatus calculates overall health status
func calculateOverallHealthStatus(healthChecks map[string]production.CheckResult, validation *production.ValidationResult) string {
	if validation != nil && !validation.Valid {
		return HealthStatusCritical
	}

	hasWarnings := false
	for _, result := range healthChecks {
		if result.Status == "failed" || result.Status == "error" {
			return HealthStatusCritical
		}
		if result.Status == "warning" {
			hasWarnings = true
		}
	}

	if hasWarnings {
		return HealthStatusWarning
	}

	return HealthStatusHealthy
}

// countIssues counts total issues found
func countIssues(healthChecks map[string]production.CheckResult, validation *production.ValidationResult) int {
	count := 0

	for _, result := range healthChecks {
		if result.Status != "passed" && result.Status != "healthy" {
			count++
		}
	}

	if validation != nil && !validation.Valid {
		count += len(validation.Issues)
	}

	return count
}

// getReportSections returns sections for report type
func getReportSections(reportType string) []string {
	switch strings.ToLower(reportType) {
	case "executive":
		return []string{"overview", "summary", "recommendations", "roadmap"}
	case "technical":
		return []string{"overview", "metrics", "detailed_analysis", "recommendations", "appendix"}
	case "operational":
		return []string{"overview", "operational_metrics", "health_checks", "recommendations"}
	case "security":
		return []string{"security_overview", "compliance", "vulnerabilities", "recommendations"}
	case "cost":
		return []string{"cost_overview", "optimization", "savings", "recommendations"}
	default:
		return []string{"overview", "metrics", "recommendations"}
	}
}

// getDetailLevel returns detail level for audience
func getDetailLevel(audience string) string {
	switch strings.ToLower(audience) {
	case "executive":
		return "summary"
	case "technical":
		return "comprehensive"
	case "operational":
		return "detailed"
	default:
		return "detailed"
	}
}

// generateDeploymentPlan generates deployment plan for strategy
func generateDeploymentPlan(strategy string, assessment *production.DeploymentReadinessAssessment) *DeploymentPlan {
	steps := generateDeploymentSteps(strategy)
	rollout := generateRolloutStrategy(strategy)
	validation := generateValidationSteps()
	monitoring := generateMonitoringPlan()

	// Use assessment data to customize deployment steps timing
	if assessment != nil {
		// Adjust step durations based on readiness score and critical issues
		for i := range steps {
			if assessment.ReadinessScore < 70 || len(assessment.CriticalIssues) > 0 {
				steps[i].Duration = steps[i].Duration * 2 // Double time for lower readiness
			} else if assessment.ReadinessScore > 90 {
				steps[i].Duration = steps[i].Duration / 2 // Half time for high readiness
			}
		}
	}

	return &DeploymentPlan{
		Strategy:   strategy,
		Steps:      steps,
		Rollout:    rollout,
		Validation: validation,
		Monitoring: monitoring,
	}
}

// generateDeploymentSteps generates deployment steps for strategy
func generateDeploymentSteps(strategy string) []DeploymentStep {
	switch strings.ToLower(strategy) {
	case "blue-green":
		return []DeploymentStep{
			{Order: 1, Name: "Prepare Green Environment", Duration: 10 * time.Minute},
			{Order: 2, Name: "Deploy to Green", Duration: 15 * time.Minute},
			{Order: 3, Name: "Run Smoke Tests", Duration: 5 * time.Minute},
			{Order: 4, Name: "Switch Traffic", Duration: 2 * time.Minute},
			{Order: 5, Name: "Monitor", Duration: 10 * time.Minute},
			{Order: 6, Name: "Cleanup Blue", Duration: 5 * time.Minute},
		}
	case deploymentRolling:
		return []DeploymentStep{
			{Order: 1, Name: "Rolling Update Start", Duration: 5 * time.Minute},
			{Order: 2, Name: "Update Batch 1", Duration: 10 * time.Minute},
			{Order: 3, Name: "Health Check", Duration: 2 * time.Minute},
			{Order: 4, Name: "Update Batch 2", Duration: 10 * time.Minute},
			{Order: 5, Name: "Final Health Check", Duration: 5 * time.Minute},
		}
	case "canary":
		return []DeploymentStep{
			{Order: 1, Name: "Deploy Canary", Duration: 10 * time.Minute},
			{Order: 2, Name: "Route 5% Traffic", Duration: 2 * time.Minute},
			{Order: 3, Name: "Monitor Canary", Duration: 15 * time.Minute},
			{Order: 4, Name: "Scale to 50%", Duration: 5 * time.Minute},
			{Order: 5, Name: "Full Deployment", Duration: 10 * time.Minute},
		}
	default:
		return []DeploymentStep{
			{Order: 1, Name: "Basic Deployment", Duration: 20 * time.Minute},
			{Order: 2, Name: "Health Check", Duration: 5 * time.Minute},
		}
	}
}

// generateRolloutStrategy generates rollout strategy
func generateRolloutStrategy(strategy string) *RolloutStrategy {
	switch strings.ToLower(strategy) {
	case deploymentRolling:
		return &RolloutStrategy{
			Type:           "RollingUpdate",
			BatchSize:      2,
			BatchDelay:     30 * time.Second,
			MaxSurge:       "25%",
			MaxUnavailable: "25%",
		}
	default:
		return &RolloutStrategy{
			Type:           "Recreate",
			BatchSize:      1,
			BatchDelay:     0,
			MaxSurge:       "100%",
			MaxUnavailable: "0",
		}
	}
}

// generateValidationSteps generates validation steps
func generateValidationSteps() *ValidationSteps {
	return &ValidationSteps{
		PreDeployment: []string{
			"Check system resources",
			"Validate configuration",
			"Backup current state",
			"Verify connectivity",
		},
		PostDeployment: []string{
			"Health endpoint check",
			"Database connectivity",
			"Integration tests",
			"Performance benchmarks",
		},
		HealthChecks: []string{
			"Application health",
			"Database health",
			"External service connectivity",
			"Resource utilization",
		},
		Smoke: []string{
			"Basic functionality test",
			"API endpoint tests",
			"Authentication test",
			"Core feature validation",
		},
	}
}

// generateMonitoringPlan generates monitoring plan
func generateMonitoringPlan() *MonitoringPlan {
	return &MonitoringPlan{
		Metrics: []string{
			"response_time",
			"error_rate",
			"throughput",
			"resource_utilization",
		},
		Alerts: []string{
			"high_error_rate",
			"slow_response_time",
			"resource_exhaustion",
			"service_unavailable",
		},
		Dashboards: []string{
			"application_overview",
			"performance_metrics",
			"error_tracking",
			"infrastructure_health",
		},
		SLIs: []string{
			"availability",
			"latency",
			"throughput",
			"error_budget",
		},
	}
}

// assessDeploymentRisks assesses deployment risks
func assessDeploymentRisks(assessment *production.DeploymentReadinessAssessment) *RiskAssessment {
	riskFactors := []RiskFactor{
		{
			Name:        "Configuration Changes",
			Level:       RiskLevelMedium,
			Impact:      "Service disruption",
			Probability: "Medium",
			Mitigation:  "Thorough configuration validation and rollback plan",
		},
		{
			Name:        "Database Migration",
			Level:       RiskLevelHigh,
			Impact:      "Data loss or corruption",
			Probability: "Low",
			Mitigation:  "Comprehensive backup and migration testing",
		},
	}

	overallRisk := RiskLevelMedium
	if assessment != nil && assessment.ReadinessScore < 70 {
		overallRisk = RiskLevelHigh
	}

	return &RiskAssessment{
		OverallRisk: overallRisk,
		RiskFactors: riskFactors,
		Mitigations: []string{
			"Comprehensive testing in staging environment",
			"Gradual rollout with monitoring",
			"Prepared rollback procedures",
			"24/7 monitoring during deployment",
		},
		ContingencyPlan: "Immediate rollback if error rate exceeds 1% or response time increases by 50%",
	}
}

// generateRollbackPlan generates rollback plan
func generateRollbackPlan(strategy string) *RollbackPlan {
	// Customize rollback strategy based on deployment strategy
	rollbackStrategy := "automated_rollback"
	duration := 5 * time.Minute

	switch strategy {
	case "blue_green":
		rollbackStrategy = "switch_traffic"
		duration = 2 * time.Minute
	case deploymentRolling:
		rollbackStrategy = "progressive_rollback"
		duration = 10 * time.Minute
	case "canary":
		rollbackStrategy = "stop_canary_rollback"
		duration = 3 * time.Minute
	}

	return &RollbackPlan{
		Strategy: rollbackStrategy,
		TriggerConditions: []string{
			"Error rate > 1%",
			"Response time > 5s",
			"Health check failures",
			"Manual trigger",
		},
		Steps: []string{
			"Stop new deployments",
			"Revert to previous version",
			"Verify system health",
			"Notify stakeholders",
		},
		Duration:     duration,
		DataRecovery: "Restore from backup if data corruption detected",
	}
}

// estimateDeploymentDuration estimates deployment duration
func estimateDeploymentDuration(strategy string, assessment *production.DeploymentReadinessAssessment) time.Duration {
	baseDuration := map[string]time.Duration{
		"blue-green": 45 * time.Minute,
		"rolling":    30 * time.Minute,
		"canary":     60 * time.Minute,
	}

	duration, exists := baseDuration[strategy]
	if !exists {
		duration = 25 * time.Minute
	}

	// Adjust based on readiness score
	if assessment != nil && assessment.ReadinessScore < 70 {
		duration = time.Duration(float64(duration) * 1.5)
	}

	return duration
}

// listDeploymentPrerequisites lists deployment prerequisites
func listDeploymentPrerequisites(assessment *production.DeploymentReadinessAssessment) []string {
	prerequisites := []string{
		"Production environment prepared",
		"Configuration validated",
		"Database migrations tested",
		"Integration tests passed",
		"Security scan completed",
		"Performance benchmarks established",
		"Monitoring systems active",
		"Rollback procedures documented",
	}

	if assessment != nil && assessment.ReadinessScore < 80 {
		prerequisites = append(prerequisites, "Address critical issues before deployment")
	}

	return prerequisites
}

// removeDuplicates removes duplicate strings from slice
func removeDuplicates(slice []string) []string {
	keys := make(map[string]bool)
	var result []string

	for _, item := range slice {
		if !keys[item] {
			keys[item] = true
			result = append(result, item)
		}
	}

	return result
}

// Continuous metrics collection
func (prc *ProductionReadinessCommands) runContinuousMetricsCollection(ctx context.Context, metricsType string, interval time.Duration, outputFormat, outputPath string) error {
	prc.logger.Info(fmt.Sprintf("Starting continuous metrics collection (interval: %v)", interval))

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			metrics, err := prc.collectSpecificMetrics(ctx, metricsType)
			if err != nil {
				prc.logger.Warn(fmt.Sprintf("Failed to collect metrics: %v", err))
				continue
			}

			if err := prc.outputMetrics(metrics, outputFormat, outputPath); err != nil {
				prc.logger.Warn(fmt.Sprintf("Failed to output metrics: %v", err))
			}
		}
	}
}

// collectSpecificMetrics collects specific type of metrics
func (prc *ProductionReadinessCommands) collectSpecificMetrics(ctx context.Context, metricsType string) (*MetricsCollectionResult, error) {
	startTime := time.Now()
	result := &MetricsCollectionResult{
		Type:      metricsType,
		Timestamp: startTime,
	}

	switch strings.ToLower(metricsType) {
	case "health":
		health, err := prc.productionSvc.GetSystemStatus(ctx)
		if err != nil {
			return nil, err
		}
		result.SystemHealth = &health.SystemHealth

	case "performance":
		status, err := prc.productionSvc.GetSystemStatus(ctx)
		if err != nil {
			return nil, err
		}
		result.Performance = &status.Performance

	case "security":
		status, err := prc.productionSvc.GetSystemStatus(ctx)
		if err != nil {
			return nil, err
		}
		result.Security = &status.Security

	case "integration":
		status, err := prc.productionSvc.GetSystemStatus(ctx)
		if err != nil {
			return nil, err
		}
		result.Integration = &status.Integration

	case "analytics":
		status, err := prc.productionSvc.GetSystemStatus(ctx)
		if err != nil {
			return nil, err
		}
		result.Analytics = &status.Analytics

	case "all":
		status, err := prc.productionSvc.GetSystemStatus(ctx)
		if err != nil {
			return nil, err
		}
		result.SystemHealth = &status.SystemHealth
		result.Performance = &status.Performance
		result.Security = &status.Security
		result.Integration = &status.Integration
		result.Analytics = &status.Analytics

	default:
		return nil, fmt.Errorf("unknown metrics type: %s", metricsType)
	}

	result.CollectionDuration = time.Since(startTime)
	return result, nil
}
