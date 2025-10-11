package monitoring

import (
	"context"
	"fmt"
	"time"

	"local/costscope/internal/core/logging"
)

// DefaultHealthChecker provides component and system health evaluations.
type DefaultHealthChecker struct {
	logger *logging.Logger
}

func NewDefaultHealthChecker(logger *logging.Logger) *DefaultHealthChecker {
	if logger == nil {
		logger = logging.GetLogger()
	}
	return &DefaultHealthChecker{logger: logger.WithFields(map[string]interface{}{"component": "monitoring", "subcomponent": "health"})}
}

func (hc *DefaultHealthChecker) System(ctx context.Context) *SystemHealthStatus {
	// Mirror existing BasicMonitoringService.collectSystemHealth behavior via delegation for step 1.
	// For minimal-risk refactor we can call into a lightweight shim that uses Component for a fixed set.
	components := []string{"api", "database", "cache", "storage", "messaging", "monitoring"}

	componentHealth := make(map[string]string)
	healthyComponents := make([]string, 0)
	degradedComponents := make([]string, 0)
	failedComponents := make([]string, 0)
	criticalComponents := make([]string, 0)

	overallScore := 0

	for _, c := range components {
		ch, _ := hc.Component(ctx, c)
		if ch == nil {
			componentHealth[c] = "unknown"
			failedComponents = append(failedComponents, c)
			continue
		}
		componentHealth[c] = ch.Status
		switch ch.Status {
		case HealthyStatus:
			healthyComponents = append(healthyComponents, c)
			overallScore += 100
		case DegradedStatus:
			degradedComponents = append(degradedComponents, c)
			overallScore += 70
		case SeverityCritical:
			criticalComponents = append(criticalComponents, c)
			overallScore += 30
		default:
			failedComponents = append(failedComponents, c)
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

func (hc *DefaultHealthChecker) Component(ctx context.Context, component string) (*ComponentHealth, error) {
	hc.logger.Info(fmt.Sprintf("Getting health status for component: %s", component))
	health := &ComponentHealth{
		ComponentName: component,
		Status:        HealthyStatus,
		HealthScore:   85,
		LastChecked:   time.Now(),
		ResponseTime:  25.5,
		ErrorRate:     0.1,
		Dependencies:  []string{"database", "cache"},
		Metrics: map[string]interface{}{
			"requests_per_second": 125.5,
			"success_rate":        99.9,
			"memory_usage":        "512MB",
		},
		Issues:          []string{},
		Recommendations: []string{"Consider optimizing query performance"},
	}
	switch component {
	case "api":
		health.Status = HealthyStatus
		health.HealthScore = 92
		health.ResponseTime = 15.2
	case "database":
		health.Status = HealthyStatus
		health.HealthScore = 88
		health.ResponseTime = 5.8
	case "cache":
		health.Status = DegradedStatus
		health.HealthScore = 75
		health.ErrorRate = 0.5
		health.Issues = []string{"High cache miss rate"}
	case "storage":
		health.Status = HealthyStatus
		health.HealthScore = 90
	case "messaging":
		health.Status = SeverityCritical
		health.HealthScore = 45
		health.ErrorRate = 3.2
		health.Issues = []string{"Message queue backlog", "Connection timeouts"}
	default:
		health.Status = "unknown"
		health.HealthScore = 0
		health.Issues = []string{"Component not found"}
	}
	return health, nil
}

func (hc *DefaultHealthChecker) Run(ctx context.Context, components []string) (*HealthCheckResults, error) {
	hc.logger.Info(fmt.Sprintf("Running health checks for %d components", len(components)))

	startTime := time.Now()
	results := make(map[string]bool)
	failedChecks := make([]string, 0)
	warningChecks := make([]string, 0)

	passedChecks := 0
	for _, component := range components {
		health, err := hc.Component(ctx, component)
		if err != nil {
			results[component] = false
			failedChecks = append(failedChecks, component)
			continue
		}
		switch health.Status {
		case HealthyStatus:
			results[component] = true
			passedChecks++
		case DegradedStatus:
			results[component] = true
			passedChecks++
			warningChecks = append(warningChecks, component)
		default:
			results[component] = false
			failedChecks = append(failedChecks, component)
		}
	}

	duration := time.Since(startTime)
	overallScore := 0
	if len(components) > 0 {
		overallScore = (passedChecks * 100) / len(components)
	}

	overallHealth := HealthyStatus
	if overallScore < 70 {
		overallHealth = SeverityCritical
	} else if overallScore < 90 {
		overallHealth = DegradedStatus
	}

	return &HealthCheckResults{
		OverallHealth:    overallHealth,
		OverallScore:     overallScore,
		ComponentResults: results,
		FailedChecks:     failedChecks,
		CheckTimestamp:   time.Now(),
		CheckDuration:    duration,
		TotalChecks:      len(components),
		PassedChecks:     passedChecks,
		FailedCheckCount: len(failedChecks),
		WarningChecks:    warningChecks,
	}, nil
}
