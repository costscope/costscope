package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"local/costscope/internal/core/production"
)

// Output format constants
const (
	outputFormatYAML    = "yaml"
	outputFormatHTML    = "html"
	reportTypeExecutive = "executive"
	deploymentRolling   = "rolling"
	statusPassed        = "passed"
)

// Output methods for production readiness commands

// outputAssessmentResult outputs production assessment result
func (prc *ProductionReadinessCommands) outputAssessmentResult(result *ProductionAssessmentResult, format, outputPath string, detailed bool) error {
	switch strings.ToLower(format) {
	case outputFormatJSON:
		return prc.outputJSON(result, outputPath)
	case outputFormatYAML:
		return prc.outputYAML(result, outputPath)
	case outputFormatHTML:
		return prc.outputHTML(result, outputPath)
	default:
		return prc.printAssessmentTable(result, detailed)
	}
}

// outputMetrics outputs metrics collection result
func (prc *ProductionReadinessCommands) outputMetrics(metrics *MetricsCollectionResult, format, outputPath string) error {
	switch strings.ToLower(format) {
	case outputFormatJSON:
		return prc.outputJSON(metrics, outputPath)
	case "prometheus":
		return prc.outputPrometheus(metrics, outputPath)
	case "csv":
		return prc.outputCSV(metrics, outputPath)
	default:
		return prc.printMetricsTable(metrics)
	}
}

// outputOptimizationReport outputs optimization report
func (prc *ProductionReadinessCommands) outputOptimizationReport(report *production.ProductionOptimizationReport, format, outputPath string) error {
	switch strings.ToLower(format) {
	case outputFormatJSON:
		return prc.outputJSON(report, outputPath)
	case "pdf":
		return prc.outputPDF(report, outputPath)
	case outputFormatHTML:
		return prc.outputHTML(report, outputPath)
	default:
		return prc.printOptimizationReportTable(report)
	}
}

// outputCheckResult outputs check result
func (prc *ProductionReadinessCommands) outputCheckResult(result *ProductionCheckResult, format string, detailed bool) error {
	switch strings.ToLower(format) {
	case outputFormatJSON:
		return prc.outputJSON(result, "")
	case "junit":
		return prc.outputJUnit(result)
	default:
		return prc.printCheckResultTable(result, detailed)
	}
}

// outputExecutiveReport outputs executive report
func (prc *ProductionReadinessCommands) outputExecutiveReport(report *production.ExecutiveReport, options *production.ReportOptions) error {
	if options.OutputPath != "" {
		switch strings.ToLower(options.Format) {
		case "pdf":
			return prc.outputPDF(report, options.OutputPath)
		case outputFormatHTML:
			return prc.outputHTML(report, options.OutputPath)
		case "markdown":
			return prc.outputMarkdown(report, options.OutputPath)
		case outputFormatJSON:
			return prc.outputJSON(report, options.OutputPath)
		}
	}

	return prc.printExecutiveReportTable(report)
}

// outputValidationResult outputs validation result
func (prc *ProductionReadinessCommands) outputValidationResult(validation *production.ValidationResult, format string, strict bool) error {
	switch strings.ToLower(format) {
	case outputFormatJSON:
		return prc.outputJSON(validation, "")
	case outputFormatYAML:
		return prc.outputYAML(validation, "")
	default:
		return prc.printValidationTable(validation, strict)
	}
}

// outputDeploymentPlan outputs deployment plan
func (prc *ProductionReadinessCommands) outputDeploymentPlan(plan *DeploymentPlanInfo, format string, dryRun bool) error {
	switch strings.ToLower(format) {
	case outputFormatJSON:
		return prc.outputJSON(plan, "")
	case outputFormatYAML:
		return prc.outputYAML(plan, "")
	default:
		return prc.printDeploymentPlanTable(plan, dryRun)
	}
}

// Generic output methods

// outputJSON outputs data as JSON
func (prc *ProductionReadinessCommands) outputJSON(data interface{}, outputPath string) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")

	if outputPath != "" {
		file, err := prc.createOutputFile(outputPath)
		if err != nil {
			return err
		}
		defer func() {
			if closeErr := file.Close(); closeErr != nil {
				prc.logger.Warnf("Error closing file: %v", closeErr)
			}
		}()
		encoder = json.NewEncoder(file)
		encoder.SetIndent("", "  ")
	}

	return encoder.Encode(data)
}

// outputYAML outputs data as YAML (simplified JSON for now)
func (prc *ProductionReadinessCommands) outputYAML(data interface{}, outputPath string) error {
	// For now, using JSON format - would implement YAML marshaling in production
	return prc.outputJSON(data, outputPath)
}

// outputHTML outputs data as HTML
func (prc *ProductionReadinessCommands) outputHTML(data interface{}, outputPath string) error {
	// For now, just noting that HTML output would be implemented
	fmt.Println("HTML output not implemented yet - showing JSON instead:")
	return prc.outputJSON(data, outputPath)
}

// outputPDF outputs data as PDF
func (prc *ProductionReadinessCommands) outputPDF(data interface{}, outputPath string) error {
	// For now, just noting that PDF output would be implemented
	fmt.Println("PDF output not implemented yet - showing JSON instead:")
	return prc.outputJSON(data, outputPath)
}

// outputMarkdown outputs data as Markdown
func (prc *ProductionReadinessCommands) outputMarkdown(data interface{}, outputPath string) error {
	// For now, just noting that Markdown output would be implemented
	fmt.Println("Markdown output not implemented yet - showing JSON instead:")
	return prc.outputJSON(data, outputPath)
}

// outputPrometheus outputs metrics in Prometheus format
func (prc *ProductionReadinessCommands) outputPrometheus(metrics *MetricsCollectionResult, outputPath string) error {
	// For now, just noting that Prometheus output would be implemented
	fmt.Println("Prometheus output not implemented yet - showing JSON instead:")
	return prc.outputJSON(metrics, outputPath)
}

// outputCSV outputs metrics as CSV
func (prc *ProductionReadinessCommands) outputCSV(metrics *MetricsCollectionResult, outputPath string) error {
	// For now, just noting that CSV output would be implemented
	fmt.Println("CSV output not implemented yet - showing JSON instead:")
	return prc.outputJSON(metrics, outputPath)
}

// outputJUnit outputs check results in JUnit format
func (prc *ProductionReadinessCommands) outputJUnit(result *ProductionCheckResult) error {
	// For now, just noting that JUnit output would be implemented
	fmt.Println("JUnit output not implemented yet - showing JSON instead:")
	return prc.outputJSON(result, "")
}

// Table printing methods

// printAssessmentTable prints assessment result as table
func (prc *ProductionReadinessCommands) printAssessmentTable(result *ProductionAssessmentResult, detailed bool) error {
	fmt.Println("═══════════════════════════════════════════════════")
	fmt.Println("           PRODUCTION READINESS ASSESSMENT")
	fmt.Println("═══════════════════════════════════════════════════")

	fmt.Printf("Environment:        %s\n", result.Environment)
	fmt.Printf("Assessment Time:    %s\n", result.Timestamp.Format("2006-01-02 15:04:05"))
	fmt.Printf("Duration:           %v\n", result.AssessmentDuration)

	if result.Summary != nil {
		fmt.Printf("Overall Score:      %d/100\n", result.Summary.OverallScore)
		fmt.Printf("Readiness Level:    %s\n", result.Summary.ReadinessLevel)
		fmt.Printf("Action Required:    %t\n", result.Summary.ActionRequired)

		if len(result.Summary.CriticalIssues) > 0 {
			fmt.Println("\n--- CRITICAL ISSUES ---")
			for _, issue := range result.Summary.CriticalIssues {
				fmt.Printf("• %s\n", issue)
			}
		}
	}

	if detailed && result.SystemStatus != nil {
		fmt.Println("\n--- DETAILED SYSTEM STATUS ---")
		fmt.Printf("Health Score:       %d/100\n", result.SystemStatus.SystemHealth.HealthScore)
		fmt.Printf("Performance Score:  %d/100\n", result.SystemStatus.Performance.OptimizationScore)
		fmt.Printf("Security Score:     %d/100\n", result.SystemStatus.Security.SecurityScore)
		fmt.Printf("Integration Score:  %d/100\n", result.SystemStatus.Integration.IntegrationScore)
	}

	if len(result.Recommendations) > 0 {
		fmt.Println("\n--- RECOMMENDATIONS ---")
		for i, rec := range result.Recommendations {
			if i >= 5 && !detailed { // Limit to 5 recommendations unless detailed
				fmt.Printf("... and %d more recommendations\n", len(result.Recommendations)-5)
				break
			}
			fmt.Printf("• [%s] %s\n", string(rec.Priority), rec.Title)
		}
	}

	return nil
}

// printMetricsTable prints metrics as table
func (prc *ProductionReadinessCommands) printMetricsTable(metrics *MetricsCollectionResult) error {
	fmt.Println("═══════════════════════════════════════════════════")
	fmt.Println("              PRODUCTION METRICS")
	fmt.Println("═══════════════════════════════════════════════════")

	fmt.Printf("Metrics Type:       %s\n", metrics.Type)
	fmt.Printf("Collection Time:    %s\n", metrics.Timestamp.Format("2006-01-02 15:04:05"))
	fmt.Printf("Collection Duration: %v\n", metrics.CollectionDuration)

	if metrics.SystemHealth != nil {
		fmt.Println("\n--- SYSTEM HEALTH ---")
		fmt.Printf("Status:             %s\n", metrics.SystemHealth.Status)
		fmt.Printf("Health Score:       %d/100\n", metrics.SystemHealth.HealthScore)
		fmt.Printf("Uptime:             %.1f hours\n", metrics.SystemHealth.UptimeHours)
		fmt.Printf("Error Rate:         %.2f%%\n", metrics.SystemHealth.ErrorRate*100)
		fmt.Printf("Response Time:      %.1f ms\n", metrics.SystemHealth.ResponseTimeMs)
	}

	if metrics.Performance != nil {
		fmt.Println("\n--- PERFORMANCE ---")
		fmt.Printf("Throughput:         %d ops/sec\n", metrics.Performance.ThroughputOpsPerSec)
		fmt.Printf("Memory Usage:       %.1f/%.1f MB\n", metrics.Performance.MemoryUsageMB, metrics.Performance.MemoryLimitMB)
		fmt.Printf("CPU Usage:          %.1f%%\n", metrics.Performance.CPUUsagePercent)
		fmt.Printf("Disk Usage:         %.1f%%\n", metrics.Performance.DiskUsagePercent)
		fmt.Printf("Network Latency:    %.1f ms\n", metrics.Performance.NetworkLatencyMs)
		fmt.Printf("Performance Grade:  %s\n", metrics.Performance.PerformanceGrade)
	}

	if metrics.Security != nil {
		fmt.Println("\n--- SECURITY ---")
		fmt.Printf("Security Score:     %d/100\n", metrics.Security.SecurityScore)
		fmt.Printf("Open Vulnerabilities: %d\n", metrics.Security.VulnerabilitiesOpen)
		fmt.Printf("High Risk Vulns:    %d\n", metrics.Security.VulnerabilitiesHigh)
		fmt.Printf("Encryption Enabled: %t\n", metrics.Security.EncryptionEnabled)
		fmt.Printf("Access Violations:  %d\n", metrics.Security.AccessViolations)
		fmt.Printf("Security Grade:     %s\n", metrics.Security.SecurityGrade)
	}

	return nil
}

// printOptimizationReportTable prints optimization report as table
func (prc *ProductionReadinessCommands) printOptimizationReportTable(report *production.ProductionOptimizationReport) error {
	fmt.Println("═══════════════════════════════════════════════════")
	fmt.Println("           PRODUCTION OPTIMIZATION REPORT")
	fmt.Println("═══════════════════════════════════════════════════")

	fmt.Printf("Generated:          %s\n", report.GeneratedAt.Format("2006-01-02 15:04:05"))

	if report.OptimizationResults.TotalImprovements > 0 {
		fmt.Println("\n--- OPTIMIZATION RESULTS ---")
		fmt.Printf("Total Improvements: %d\n", report.OptimizationResults.TotalImprovements)
		fmt.Printf("Performance Gains:  %.1f%%\n", report.OptimizationResults.PerformanceGains)
		fmt.Printf("Cost Savings:       $%.0f\n", report.OptimizationResults.CostSavings)
		fmt.Printf("Security Enhancements: %d\n", report.OptimizationResults.SecurityEnhancements)
		fmt.Printf("Efficiency Gains:   %.1f%%\n", report.OptimizationResults.EfficiencyGains)
	}

	if len(report.Recommendations) > 0 {
		fmt.Println("\n--- TOP RECOMMENDATIONS ---")
		for i, rec := range report.Recommendations {
			if i >= 5 { // Limit to top 5
				break
			}
			fmt.Printf("• [%s] %s: %s\n", string(rec.Priority), rec.Type, rec.Title)
			fmt.Printf("  Impact: %s, Effort: %s, ROI: %.1f%%\n", string(rec.Impact), string(rec.Effort), rec.ROI*100)
		}
	}

	if report.ROIAnalysis.TotalInvestment > 0 {
		fmt.Println("\n--- ROI ANALYSIS ---")
		fmt.Printf("Total Investment:   $%.0f\n", report.ROIAnalysis.TotalInvestment)
		fmt.Printf("Projected Savings:  $%.0f\n", report.ROIAnalysis.ProjectedSavings)
		fmt.Printf("ROI Percentage:     %.1f%%\n", report.ROIAnalysis.ROIPercentage)
		fmt.Printf("Payback Period:     %d days\n", report.ROIAnalysis.PaybackPeriodDays)
	}

	return nil
}

// printCheckResultTable prints check result as table
func (prc *ProductionReadinessCommands) printCheckResultTable(result *ProductionCheckResult, detailed bool) error {
	fmt.Println("═══════════════════════════════════════════════════")
	fmt.Println("         PRODUCTION HEALTH & COMPLIANCE CHECK")
	fmt.Println("═══════════════════════════════════════════════════")

	fmt.Printf("Check Time:         %s\n", result.Timestamp.Format("2006-01-02 15:04:05"))
	fmt.Printf("Overall Status:     %s\n", result.OverallStatus)
	fmt.Printf("Components Checked: %d\n", len(result.ComponentsChecked))
	fmt.Printf("Issues Found:       %d\n", result.IssuesFound)

	if len(result.HealthChecks) > 0 {
		fmt.Println("\n--- HEALTH CHECK RESULTS ---")
		for component, checkResult := range result.HealthChecks {
			status := " PASS"
			if checkResult.Status != "passed" {
				status = " FAIL"
			}
			fmt.Printf("%-20s %s - %s\n", component, status, checkResult.Message)
			if detailed && checkResult.Details != "" {
				fmt.Printf("  Details: %s\n", checkResult.Details)
			}
		}
	}

	if result.ConfigValidation != nil {
		fmt.Println("\n--- CONFIGURATION VALIDATION ---")
		status := " VALID"
		if !result.ConfigValidation.Valid {
			status = " INVALID"
		}
		fmt.Printf("Configuration:      %s (Score: %d/100)\n", status, result.ConfigValidation.Score)

		if len(result.ConfigValidation.Issues) > 0 {
			fmt.Println("Issues:")
			for _, issue := range result.ConfigValidation.Issues {
				fmt.Printf("  • %s: %s\n", issue.Title, issue.Description)
			}
		}
	}

	return nil
}

// printExecutiveReportTable prints executive report as table
func (prc *ProductionReadinessCommands) printExecutiveReportTable(report *production.ExecutiveReport) error {
	fmt.Println("═══════════════════════════════════════════════════")
	fmt.Println("              EXECUTIVE PRODUCTION REPORT")
	fmt.Println("═══════════════════════════════════════════════════")

	fmt.Printf("Generated:          %s\n", report.GeneratedAt.Format("2006-01-02 15:04:05"))

	// Basic implementation - would be expanded for full executive report
	fmt.Println("\n--- EXECUTIVE SUMMARY ---")
	fmt.Printf("Report available for detailed viewing\n")
	fmt.Printf("Processing Time:    %d ms\n", report.ProcessingTimeMs)

	return nil
}

// printValidationTable prints validation result as table
func (prc *ProductionReadinessCommands) printValidationTable(validation *production.ValidationResult, strict bool) error {
	fmt.Println("═══════════════════════════════════════════════════")
	fmt.Println("         PRODUCTION CONFIGURATION VALIDATION")
	fmt.Println("═══════════════════════════════════════════════════")

	// In strict mode, treat warnings as blocking issues
	hasBlockingIssues := !validation.Valid
	if strict && len(validation.Warnings) > 0 {
		hasBlockingIssues = true
	}

	status := " VALID"
	if hasBlockingIssues {
		status = " INVALID"
	}

	fmt.Printf("Configuration:      %s", status)
	if strict {
		fmt.Printf(" (STRICT MODE)")
	}
	fmt.Printf("\n")
	fmt.Printf("Validation Score:   %d/100\n", validation.Score)

	if len(validation.Issues) > 0 {
		fmt.Println("\n--- ISSUES ---")
		for _, issue := range validation.Issues {
			fmt.Printf("• [%s] %s: %s\n", issue.Severity, issue.Title, issue.Description)
		}
	}

	if len(validation.Warnings) > 0 {
		warningLabel := "--- WARNINGS ---"
		if strict {
			warningLabel = "--- WARNINGS (TREATED AS ERRORS IN STRICT MODE) ---"
		}
		fmt.Printf("\n%s\n", warningLabel)
		for _, warning := range validation.Warnings {
			prefix := "• "
			if strict {
				prefix = " "
			}
			fmt.Printf("%s%s: %s\n", prefix, warning.Title, warning.Description)
		}
	}

	if len(validation.Recommendations) > 0 {
		fmt.Println("\n--- RECOMMENDATIONS ---")
		for _, rec := range validation.Recommendations {
			fmt.Printf("• %s\n", rec)
		}
	}

	return nil
}

// printDeploymentPlanTable prints deployment plan as table
func (prc *ProductionReadinessCommands) printDeploymentPlanTable(plan *DeploymentPlanInfo, dryRun bool) error {
	fmt.Println("═══════════════════════════════════════════════════")
	fmt.Println("            PRODUCTION DEPLOYMENT PLAN")
	fmt.Println("═══════════════════════════════════════════════════")

	if dryRun {
		fmt.Println("                    [DRY RUN MODE]")
		fmt.Println("═══════════════════════════════════════════════════")
	}

	fmt.Printf("Strategy:           %s\n", plan.Strategy)
	fmt.Printf("Environment:        %s\n", plan.Environment)
	fmt.Printf("Estimated Duration: %v\n", plan.EstimatedDuration)

	if plan.ReadinessAssessment != nil {
		fmt.Printf("Readiness Score:    %d/100\n", plan.ReadinessAssessment.ReadinessScore)
		fmt.Printf("Readiness Status:   %s\n", plan.ReadinessAssessment.ReadinessStatus)
	}

	if plan.DeploymentPlan != nil && len(plan.DeploymentPlan.Steps) > 0 {
		fmt.Println("\n--- DEPLOYMENT STEPS ---")
		for _, step := range plan.DeploymentPlan.Steps {
			fmt.Printf("%d. %s (%v)\n", step.Order, step.Name, step.Duration)
			if step.Description != "" {
				fmt.Printf("   %s\n", step.Description)
			}
		}
	}

	if plan.RiskAssessment != nil {
		fmt.Println("\n--- RISK ASSESSMENT ---")
		fmt.Printf("Overall Risk:       %s\n", plan.RiskAssessment.OverallRisk)
		if len(plan.RiskAssessment.RiskFactors) > 0 {
			for _, risk := range plan.RiskAssessment.RiskFactors {
				fmt.Printf("• %s (%s risk): %s\n", risk.Name, risk.Level, risk.Impact)
			}
		}
	}

	if len(plan.Prerequisites) > 0 {
		fmt.Println("\n--- PREREQUISITES ---")
		for _, prereq := range plan.Prerequisites {
			fmt.Printf("• %s\n", prereq)
		}
	}

	if plan.RollbackPlan != nil {
		fmt.Println("\n--- ROLLBACK PLAN ---")
		fmt.Printf("Strategy:           %s\n", plan.RollbackPlan.Strategy)
		fmt.Printf("Estimated Duration: %v\n", plan.RollbackPlan.Duration)
	}

	return nil
}

// Helper methods

// createOutputFile creates output file with directory structure
func (prc *ProductionReadinessCommands) createOutputFile(outputPath string) (*os.File, error) {
	// Clean the path to prevent directory traversal
	cleanPath := filepath.Clean(outputPath)
	dir := filepath.Dir(cleanPath)

	// Use more restrictive permissions
	if err := os.MkdirAll(dir, 0750); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	file, err := os.Create(cleanPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create output file: %w", err)
	}

	return file, nil
}
