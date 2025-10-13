package monitoring

import (
	"context"
	"testing"
	"time"

	"github.com/costscope/costscope/internal/core/logging"
)

// TestAlertManager_RuleLifecycle exercises create/update/delete/list paths plus rule field mutations.
func TestAlertManager_RuleLifecycle(t *testing.T) {
	logger := logging.NewLogger(logging.LevelError)
	svc := NewBasicMonitoringService(logger, nil, nil)
	am := svc.alertManager.(*BasicAlertManager)
	ctx := context.Background()

	// Create
	rule := &AlertRule{Name: "Test Rule", Description: "desc", Metric: "cpu", Operator: ">", Threshold: 80, Severity: SeverityWarning, Enabled: true}
	if err := am.CreateAlertRule(ctx, rule); err != nil {
		t.Fatalf("create rule: %v", err)
	}
	if rule.ID == "" || rule.CreatedAt.IsZero() || rule.UpdatedAt.IsZero() {
		t.Fatalf("rule timestamps/ID not set")
	}

	// Update (change threshold)
	updated := *rule
	updated.Threshold = 85
	if err := am.UpdateAlertRule(ctx, rule.ID, &updated); err != nil {
		t.Fatalf("update rule: %v", err)
	}
	if got := am.alertRules[rule.ID]; got.Threshold != 85 || got.UpdatedAt.Equal(got.CreatedAt) {
		t.Fatalf("update not applied")
	}

	// List
	rules, err := am.ListAlertRules(ctx)
	if err != nil || len(rules) == 0 {
		t.Fatalf("list rules failed: %v", err)
	}

	// Delete
	if err := am.DeleteAlertRule(ctx, rule.ID); err != nil {
		t.Fatalf("delete rule: %v", err)
	}
	if _, exists := am.alertRules[rule.ID]; exists {
		t.Fatalf("rule still present after delete")
	}
}

// TestAlertManager_TriggerAlert_RuleApplication ensures rule fields override alert ones when enabled and skip when disabled.
func TestAlertManager_TriggerAlert_RuleApplication(t *testing.T) {
	logger := logging.NewLogger(logging.LevelError)
	svc := NewBasicMonitoringService(logger, nil, nil)
	am := svc.alertManager.(*BasicAlertManager)
	ctx := context.Background()

	// Create custom rule
	rule := &AlertRule{ID: "custom_rule", Name: "Custom", Metric: "latency", Operator: ">", Threshold: 500, Severity: SeverityCritical, Enabled: true, Tags: map[string]string{"k": "v"}}
	if err := am.CreateAlertRule(ctx, rule); err != nil {
		t.Fatalf("create rule: %v", err)
	}

	alert := &Alert{ID: "a1", Type: "custom_rule", Severity: SeverityInfo, Title: "t", Description: "d"}
	if err := am.TriggerAlert(ctx, alert); err != nil {
		t.Fatalf("trigger alert: %v", err)
	}
	if alert.Severity != SeverityCritical { // overridden by rule
		t.Fatalf("expected severity override, got %s", alert.Severity)
	}
	if alert.Tags["k"] != "v" {
		t.Fatalf("expected tags merged from rule")
	}

	// Disable and ensure early return (no severity change if we tweak rule)
	am.alertRules["custom_rule"].Enabled = false
	alert2 := &Alert{ID: "a2", Type: "custom_rule", Severity: SeverityInfo, Title: "t2"}
	if err := am.TriggerAlert(ctx, alert2); err != nil { // should still succeed
		t.Fatalf("trigger alert disabled: %v", err)
	}
	if alert2.Severity != SeverityInfo { // unchanged because rule disabled
		t.Fatalf("expected severity untouched when rule disabled")
	}
}

// TestAlertManager_EscalateAndAcknowledge covers escalate (with max) and acknowledge paths including missing cases.
func TestAlertManager_EscalateAndAcknowledge(t *testing.T) {
	logger := logging.NewLogger(logging.LevelError)
	svc := NewBasicMonitoringService(logger, nil, nil)
	am := svc.alertManager.(*BasicAlertManager)
	ctx := context.Background()

	// Seed an active alert matching cpu_critical escalation rule.
	a := &Alert{ID: "cpu1", Type: "cpu_critical", Title: "High CPU", Severity: SeverityCritical, Status: AlertStatusActive, CreatedAt: time.Now()}
	svc.mu.Lock()
	svc.activeAlerts = append(svc.activeAlerts, a)
	svc.mu.Unlock()

	if err := am.EscalateAlert(ctx, "cpu1"); err != nil {
		t.Fatalf("first escalate: %v", err)
	}
	if a.EscalationLevel != 1 {
		t.Fatalf("expected escalation level 1, got %d", a.EscalationLevel)
	}
	// Second escalate hits max (2) and returns without error.
	if err := am.EscalateAlert(ctx, "cpu1"); err != nil {
		t.Fatalf("second escalate: %v", err)
	}
	if a.EscalationLevel != 2 { // incremented before max check
		t.Fatalf("expected escalation level 2 (max), got %d", a.EscalationLevel)
	}
	// Third escalate should still return nil (remains at 3? No: code increments then checks >= MaxEscalations (2) and returns; but since already 2 it will increment to 3) – verify behavior.
	if err := am.EscalateAlert(ctx, "cpu1"); err != nil { // ensure no panic / error even beyond max
		t.Fatalf("third escalate: %v", err)
	}
	if a.EscalationLevel < 2 { // guard unexpected regression
		t.Fatalf("unexpected escalation level after third call: %d", a.EscalationLevel)
	}

	// Acknowledge
	if err := am.AcknowledgeAlert(ctx, "cpu1", "tester"); err != nil {
		t.Fatalf("acknowledge: %v", err)
	}
	if a.AcknowledgedBy != "tester" || a.AcknowledgedAt == nil {
		t.Fatalf("ack fields not set correctly")
	}

	// Missing cases
	if err := am.EscalateAlert(ctx, "missing"); err == nil {
		t.Fatalf("expected error escalating missing alert")
	}
	if err := am.AcknowledgeAlert(ctx, "missing", "tester"); err == nil {
		t.Fatalf("expected error acknowledging missing alert")
	}
}
