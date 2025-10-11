package monitoring

import (
	"context"
	"time"
)

// collectSystemHealth was extracted from service.go to reduce LOC without changing behavior.
func (bms *BasicMonitoringService) collectSystemHealth(ctx context.Context) *SystemHealthStatus {
	components := []string{"api", "database", "cache", "storage", "messaging", "monitoring"}

	componentHealth := make(map[string]string)
	healthyComponents := make([]string, 0)
	degradedComponents := make([]string, 0)
	failedComponents := make([]string, 0)
	criticalComponents := make([]string, 0)

	overallScore := 0

	for _, component := range components {
		health, err := bms.GetComponentHealth(ctx, component)
		if err != nil || health == nil {
			componentHealth[component] = UnknownStatus
			failedComponents = append(failedComponents, component)
			continue
		}

		componentHealth[component] = health.Status

		switch health.Status {
		case HealthyStatus:
			healthyComponents = append(healthyComponents, component)
			overallScore += 100
		case DegradedStatus:
			degradedComponents = append(degradedComponents, component)
			overallScore += 70
		case SeverityCritical:
			criticalComponents = append(criticalComponents, component)
			overallScore += 30
		default:
			failedComponents = append(failedComponents, component)
		}
	}

	if len(components) > 0 {
		overallScore = overallScore / len(components)
	}

	overallHealth := HealthyStatus
	if overallScore < 50 {
		overallHealth = SeverityCritical
	} else if overallScore < 80 {
		overallHealth = DegradedStatus
	}

	return &SystemHealthStatus{
		OverallHealth:      overallHealth,
		HealthScore:        overallScore,
		ComponentHealth:    componentHealth,
		CriticalComponents: criticalComponents,
		HealthyComponents:  healthyComponents,
		DegradedComponents: degradedComponents,
		FailedComponents:   failedComponents,
		LastHealthCheck:    time.Now(),
		UptimeHours:        72.5,
		SystemLoad:         1.2,
		MemoryPressure:     0.65,
		DiskPressure:       0.78,
	}
}
