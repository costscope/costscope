package monitoring

import (
	"context"
	"fmt"
	"sync"
	"time"

	"local/costscope/internal/core/logging"
)

// BasicAlertManager implements AlertManager interface
type BasicAlertManager struct {
	logger            *logging.Logger
	monitoringService *BasicMonitoringService
	alertRules        map[string]*AlertRule
	escalationRules   map[string]EscalationRule

	mu sync.RWMutex
}

// EscalationRule defines how alerts should be escalated
type EscalationRule struct {
	ID               string            `json:"id"`
	AlertType        string            `json:"alert_type"`
	EscalationLevels []EscalationLevel `json:"escalation_levels"`
	MaxEscalations   int               `json:"max_escalations"`
}

// EscalationLevel defines a level of escalation
type EscalationLevel struct {
	Level                int      `json:"level"`
	DelayMinutes         int      `json:"delay_minutes"`
	NotificationChannels []string `json:"notification_channels"`
	Recipients           []string `json:"recipients"`
}

// NewBasicAlertManager creates a new basic alert manager
func NewBasicAlertManager(logger *logging.Logger, monitoringService *BasicMonitoringService) *BasicAlertManager {
	manager := &BasicAlertManager{
		logger:            logger,
		monitoringService: monitoringService,
		alertRules:        make(map[string]*AlertRule),
		escalationRules:   make(map[string]EscalationRule),
	}

	// Initialize default alert rules
	manager.initializeDefaultAlertRules()

	return manager
}

// TriggerAlert triggers a new alert
func (bam *BasicAlertManager) TriggerAlert(ctx context.Context, alert *Alert) error {
	bam.mu.Lock()
	defer bam.mu.Unlock()

	bam.logger.Info(fmt.Sprintf("Triggering alert: %s - %s", alert.Type, alert.Title))

	// Set baseline alert metadata (preserve previous semantics even if rule skips)
	if alert.CreatedAt.IsZero() {
		alert.CreatedAt = time.Now()
	}
	alert.UpdatedAt = time.Now()
	if alert.Status == "" {
		alert.Status = AlertStatusActive
	}

	// Check if we should trigger based on rules
	if rule, exists := bam.alertRules[alert.Type]; exists {
		if !rule.Enabled {
			bam.logger.Debug(fmt.Sprintf("Alert rule %s is disabled, skipping", alert.Type))
			return nil
		}

		// Apply rule configuration
		alert.Severity = rule.Severity
		if len(rule.Tags) > 0 {
			if alert.Tags == nil {
				alert.Tags = make(map[string]string)
			}
			for k, v := range rule.Tags {
				alert.Tags[k] = v
			}
		}
	}
	bam.emitAlert(alert)
	return nil
}

// EscalateAlert escalates an existing alert
func (bam *BasicAlertManager) EscalateAlert(ctx context.Context, alertID string) error {
	bam.mu.Lock()
	defer bam.mu.Unlock()

	bam.logger.Info(fmt.Sprintf("Escalating alert: %s", alertID))

	// Find the alert in the monitoring service
	alerts, err := bam.monitoringService.GetActiveAlerts(ctx)
	if err != nil {
		return fmt.Errorf("failed to get active alerts: %w", err)
	}

	var targetAlert *Alert
	for _, alert := range alerts {
		if alert.ID == alertID {
			targetAlert = alert
			break
		}
	}

	if targetAlert == nil {
		return fmt.Errorf("alert %s not found", alertID)
	}

	// Increment escalation level
	targetAlert.EscalationLevel++
	targetAlert.UpdatedAt = time.Now()

	// Check escalation rules
	if escalationRule, exists := bam.escalationRules[targetAlert.Type]; exists {
		if targetAlert.EscalationLevel >= escalationRule.MaxEscalations {
			bam.logger.Warn(fmt.Sprintf("Alert %s reached maximum escalation level", alertID))
			return nil
		}

		// Apply escalation logic
		for _, level := range escalationRule.EscalationLevels {
			if level.Level == targetAlert.EscalationLevel {
				bam.logger.Info(fmt.Sprintf("Escalating alert %s to level %d", alertID, level.Level))
				// Send escalated notifications
				for _, channel := range level.NotificationChannels {
					bam.logger.Info(fmt.Sprintf("Sending escalation notification to %s", channel))
				}
				break
			}
		}
	}

	bam.logger.Info(fmt.Sprintf("Alert %s escalated to level %d", alertID, targetAlert.EscalationLevel))
	return nil
}

// AcknowledgeAlert acknowledges an alert
func (bam *BasicAlertManager) AcknowledgeAlert(ctx context.Context, alertID string, user string) error {
	bam.mu.Lock()
	defer bam.mu.Unlock()

	bam.logger.Info(fmt.Sprintf("Acknowledging alert %s by user %s", alertID, user))

	// Find the alert in the monitoring service
	alerts, err := bam.monitoringService.GetActiveAlerts(ctx)
	if err != nil {
		return fmt.Errorf("failed to get active alerts: %w", err)
	}

	for _, alert := range alerts {
		if alert.ID == alertID {
			now := time.Now()
			alert.AcknowledgedAt = &now
			alert.AcknowledgedBy = user
			alert.UpdatedAt = now

			bam.logger.Info(fmt.Sprintf("Alert %s acknowledged by %s", alertID, user))
			return nil
		}
	}

	return fmt.Errorf("alert %s not found", alertID)
}

// CreateAlertRule creates a new alert rule
func (bam *BasicAlertManager) CreateAlertRule(ctx context.Context, rule *AlertRule) error {
	bam.mu.Lock()
	defer bam.mu.Unlock()

	bam.logger.Info(fmt.Sprintf("Creating alert rule: %s", rule.Name))

	// Validate rule
	if rule.ID == "" {
		rule.ID = fmt.Sprintf("rule_%d", time.Now().Unix())
	}

	if rule.CreatedAt.IsZero() {
		rule.CreatedAt = time.Now()
	}
	rule.UpdatedAt = time.Now()

	// Store the rule
	bam.alertRules[rule.ID] = rule

	bam.logger.Info(fmt.Sprintf("Alert rule created: %s (%s)", rule.Name, rule.ID))
	return nil
}

// UpdateAlertRule updates an existing alert rule
func (bam *BasicAlertManager) UpdateAlertRule(ctx context.Context, ruleID string, rule *AlertRule) error {
	bam.mu.Lock()
	defer bam.mu.Unlock()

	bam.logger.Info(fmt.Sprintf("Updating alert rule: %s", ruleID))

	if _, exists := bam.alertRules[ruleID]; !exists {
		return fmt.Errorf("alert rule %s not found", ruleID)
	}

	rule.ID = ruleID
	rule.UpdatedAt = time.Now()
	bam.alertRules[ruleID] = rule

	bam.logger.Info(fmt.Sprintf("Alert rule updated: %s", ruleID))
	return nil
}

// DeleteAlertRule deletes an alert rule
func (bam *BasicAlertManager) DeleteAlertRule(ctx context.Context, ruleID string) error {
	bam.mu.Lock()
	defer bam.mu.Unlock()

	bam.logger.Info(fmt.Sprintf("Deleting alert rule: %s", ruleID))

	if _, exists := bam.alertRules[ruleID]; !exists {
		return fmt.Errorf("alert rule %s not found", ruleID)
	}

	delete(bam.alertRules, ruleID)

	bam.logger.Info(fmt.Sprintf("Alert rule deleted: %s", ruleID))
	return nil
}

// ListAlertRules returns all alert rules
func (bam *BasicAlertManager) ListAlertRules(ctx context.Context) ([]*AlertRule, error) {
	bam.mu.RLock()
	defer bam.mu.RUnlock()

	rules := make([]*AlertRule, 0, len(bam.alertRules))
	for _, rule := range bam.alertRules {
		rules = append(rules, rule)
	}

	return rules, nil
}

// Helper methods

func (bam *BasicAlertManager) initializeDefaultAlertRules() {
	defaultRules := []*AlertRule{
		bam.buildAlertRule(
			"cpu_critical",
			"CPU Critical Usage",
			"CPU usage exceeded critical threshold",
			"cpu_usage",
			">=",
			90.0,
			5*time.Minute,
			"critical",
			[]string{"email", "slack", "pagerduty"},
			map[string]string{"component": "system", "priority": "high"},
		),
		bam.buildAlertRule(
			"cpu_warning",
			"CPU Warning Usage",
			"CPU usage exceeded warning threshold",
			"cpu_usage",
			">=",
			70.0,
			10*time.Minute,
			"warning",
			[]string{"email", "slack"},
			map[string]string{"component": "system", "priority": "medium"},
		),
		bam.buildAlertRule(
			"memory_critical",
			"Memory Critical Usage",
			"Memory usage exceeded critical threshold",
			"memory_usage",
			">=",
			95.0,
			3*time.Minute,
			"critical",
			[]string{"email", "slack", "pagerduty"},
			map[string]string{"component": "system", "priority": "high"},
		),
		bam.buildAlertRule(
			"memory_warning",
			"Memory Warning Usage",
			"Memory usage exceeded warning threshold",
			"memory_usage",
			">=",
			80.0,
			15*time.Minute,
			"warning",
			[]string{"email", "slack"},
			map[string]string{"component": "system", "priority": "medium"},
		),
		bam.buildAlertRule(
			"disk_critical",
			"Disk Critical Usage",
			"Disk usage exceeded critical threshold",
			"disk_usage",
			">=",
			95.0,
			5*time.Minute,
			"critical",
			[]string{"email", "slack", "pagerduty"},
			map[string]string{"component": "storage", "priority": "high"},
		),
		bam.buildAlertRule(
			"error_rate_critical",
			"Error Rate Critical",
			"Application error rate exceeded critical threshold",
			"error_rate",
			">=",
			5.0,
			2*time.Minute,
			"critical",
			[]string{"email", "slack", "pagerduty"},
			map[string]string{"component": "application", "priority": "high"},
		),
		bam.buildAlertRule(
			"latency_critical",
			"Response Time Critical",
			"Response time exceeded critical threshold",
			"response_time",
			">=",
			500.0,
			5*time.Minute,
			"critical",
			[]string{"email", "slack"},
			map[string]string{"component": "application", "priority": "high"},
		),
		bam.buildAlertRule(
			"integration_failure",
			"Integration Failure",
			"Integration system failure detected",
			"integration_health",
			"<",
			50.0,
			1*time.Minute,
			"critical",
			[]string{"email", "slack", "pagerduty"},
			map[string]string{"component": "integration", "priority": "high"},
		),
	}

	// Add default rules to the manager
	for _, rule := range defaultRules {
		bam.alertRules[rule.ID] = rule
	}

	// Initialize escalation rules
	bam.escalationRules["cpu_critical"] = bam.buildEscalationRule(
		"cpu_critical",
		15, 30,
		[]string{"email", "slack"}, []string{"team-lead", "ops-manager"},
		[]string{"email", "slack", "pagerduty"}, []string{"engineering-manager", "cto"},
		2,
	)

	bam.escalationRules["memory_critical"] = bam.buildEscalationRule(
		"memory_critical",
		10, 20,
		[]string{"email", "slack"}, []string{"team-lead", "ops-manager"},
		[]string{"email", "slack", "pagerduty"}, []string{"engineering-manager", "cto"},
		2,
	)

	bam.logger.Info(fmt.Sprintf("Initialized %d default alert rules and %d escalation rules",
		len(defaultRules), len(bam.escalationRules)))
}

// buildAlertRule creates a fully-initialized AlertRule with standard defaults and timestamps.
func (bam *BasicAlertManager) buildAlertRule(
	id, name, description, metric, operator string,
	threshold float64,
	duration time.Duration,
	severity string,
	channels []string,
	tags map[string]string,
) *AlertRule {
	now := time.Now()
	// defensive copies to avoid aliasing external slices/maps
	chCopy := make([]string, len(channels))
	copy(chCopy, channels)
	tagCopy := make(map[string]string, len(tags))
	for k, v := range tags {
		tagCopy[k] = v
	}
	return &AlertRule{
		ID:                   id,
		Name:                 name,
		Description:          description,
		Metric:               metric,
		Operator:             operator,
		Threshold:            threshold,
		Duration:             duration,
		Severity:             severity,
		NotificationChannels: chCopy,
		Tags:                 tagCopy,
		Enabled:              true,
		CreatedAt:            now,
		UpdatedAt:            now,
	}
}

// emitAlert finalizes and logs an alert emission without changing semantics.
func (bam *BasicAlertManager) emitAlert(alert *Alert) {
	// finalize metadata
	now := time.Now()
	if alert.CreatedAt.IsZero() {
		alert.CreatedAt = now
	}
	alert.UpdatedAt = now
	if alert.Status == "" {
		alert.Status = AlertStatusActive
	}

	bam.logger.Info(fmt.Sprintf("Alert triggered successfully: %s", alert.ID))
}

// buildEscalationRule constructs an EscalationRule with two levels matching existing defaults.
func (bam *BasicAlertManager) buildEscalationRule(
	alertType string,
	level1DelayMinutes int, level2DelayMinutes int,
	level1Channels []string, level1Recipients []string,
	level2Channels []string, level2Recipients []string,
	maxEscalations int,
) EscalationRule {
	// defensive copies to avoid aliasing external slices
	l1Ch := make([]string, len(level1Channels))
	copy(l1Ch, level1Channels)
	l1Rcpt := make([]string, len(level1Recipients))
	copy(l1Rcpt, level1Recipients)
	l2Ch := make([]string, len(level2Channels))
	copy(l2Ch, level2Channels)
	l2Rcpt := make([]string, len(level2Recipients))
	copy(l2Rcpt, level2Recipients)

	return EscalationRule{
		ID:        fmt.Sprintf("%s_escalation", alertType),
		AlertType: alertType,
		EscalationLevels: []EscalationLevel{
			{
				Level:                1,
				DelayMinutes:         level1DelayMinutes,
				NotificationChannels: l1Ch,
				Recipients:           l1Rcpt,
			},
			{
				Level:                2,
				DelayMinutes:         level2DelayMinutes,
				NotificationChannels: l2Ch,
				Recipients:           l2Rcpt,
			},
		},
		MaxEscalations: maxEscalations,
	}
}
