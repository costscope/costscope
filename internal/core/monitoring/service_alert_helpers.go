package monitoring

// Utility and alert helper methods extracted from service.go (no behavior change)

func (bms *BasicMonitoringService) countAlertsBySeverity(severity string) int {
	count := 0
	for _, alert := range bms.activeAlerts {
		if alert.Severity == severity {
			count++
		}
	}
	return count
}

func (bms *BasicMonitoringService) getRecentAlerts(limit int) []Alert {
	alerts := make([]Alert, 0)
	count := 0

	for i := len(bms.activeAlerts) - 1; i >= 0 && count < limit; i-- {
		alerts = append(alerts, *bms.activeAlerts[i])
		count++
	}

	return alerts
}

func (bms *BasicMonitoringService) getStatusFromValue(value, warning, critical float64) string {
	if value >= critical {
		return "critical"
	} else if value >= warning {
		return "warning"
	}
	return HealthyStatus
}
