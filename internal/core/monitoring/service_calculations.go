package monitoring

// Helper calculation methods extracted from service.go (no behavior change)

func (bms *BasicMonitoringService) calculatePerformanceScore(resources *ResourceMetrics) int {
	cpuScore := 100 - int(resources.CPU.UsagePercent)
	memoryScore := 100 - int(resources.Memory.UsagePercent)
	diskScore := 100 - int(resources.Disk.UsagePercent)

	return (cpuScore + memoryScore + diskScore) / 3
}

func (bms *BasicMonitoringService) calculatePerformanceGrade(resources *ResourceMetrics) string {
	score := bms.calculatePerformanceScore(resources)

	switch {
	case score >= 90:
		return "A"
	case score >= 80:
		return "B"
	case score >= 70:
		return "C"
	case score >= 60:
		return "D"
	default:
		return "F"
	}
}

func (bms *BasicMonitoringService) calculateOverallHealthScore(resources *ResourceMetrics, app *ApplicationMetrics) int {
	performanceScore := bms.calculatePerformanceScore(resources)
	appScore := int(100 - app.ErrorRate*10) // Convert error rate to score

	return (performanceScore + appScore) / 2
}

func (bms *BasicMonitoringService) generateTrendIndicators() map[string]string {
	return map[string]string{
		"cpu":     "increasing",
		"memory":  "stable",
		"disk":    "stable",
		"network": "decreasing",
		"errors":  "stable",
		"latency": "improving",
	}
}
