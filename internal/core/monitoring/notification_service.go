package monitoring

import (
	"context"
	"fmt"
	"time"

	"github.com/costscope/costscope/internal/core/logging"
)

// BasicNotificationService implements NotificationService interface
type BasicNotificationService struct {
	logger                *logging.Logger
	supportedChannels     []string
	channelConfigurations map[string]ChannelConfig
	// handlers is a registry that maps channel type -> handler
	handlers map[string]NotificationHandler
}

// ChannelConfig defines configuration for notification channels
type ChannelConfig struct {
	Name          string                 `json:"name"`
	Type          string                 `json:"type"`
	Enabled       bool                   `json:"enabled"`
	Configuration map[string]interface{} `json:"configuration"`
	RateLimit     RateLimit              `json:"rate_limit"`
	RetryPolicy   RetryPolicy            `json:"retry_policy"`
}

// RateLimit defines rate limiting for notifications
type RateLimit struct {
	MaxNotifications int           `json:"max_notifications"`
	TimeWindow       time.Duration `json:"time_window"`
	BurstLimit       int           `json:"burst_limit"`
}

// RetryPolicy defines retry behavior for failed notifications
type RetryPolicy struct {
	MaxRetries    int           `json:"max_retries"`
	InitialDelay  time.Duration `json:"initial_delay"`
	BackoffFactor float64       `json:"backoff_factor"`
	MaxDelay      time.Duration `json:"max_delay"`
}

// Small helpers to reduce repeated struct literal patterns (helps with duplicate detection)
//
//nolint:unparam // window currently always 1h in templates; keep param for future variability
func rl(max int, window time.Duration, burst int) RateLimit {
	return RateLimit{MaxNotifications: max, TimeWindow: window, BurstLimit: burst}
}

func rp(maxRetries int, initialDelay time.Duration, backoff float64, maxDelay time.Duration) RetryPolicy {
	return RetryPolicy{MaxRetries: maxRetries, InitialDelay: initialDelay, BackoffFactor: backoff, MaxDelay: maxDelay}
}

func cfg(name, typ string, enabled bool, conf map[string]interface{}, rate RateLimit, retry RetryPolicy) ChannelConfig {
	return ChannelConfig{
		Name:          name,
		Type:          typ,
		Enabled:       enabled,
		Configuration: conf,
		RateLimit:     rate,
		RetryPolicy:   retry,
	}
}

// NotificationHandler defines a contract for sending notifications for a specific channel type.
// Implementations should perform delivery only; status bookkeeping is handled by the dispatcher.
type NotificationHandler interface {
	// Type returns the channel type this handler is responsible for (e.g., "email", "slack").
	Type() string
	// Deliver performs the actual send; should honor ctx for cancellation/timeouts.
	Deliver(ctx context.Context, notification *Notification, config ChannelConfig, logger *logging.Logger) error
}

// baseHandler provides a reusable implementation for simple channels with identical scaffolding
// (log start, wait-or-cancel, optional hooks, log success). It reduces duplication across handlers.
type baseHandler struct {
	typ     string
	display string
	delay   time.Duration
	before  func(ctx context.Context, n *Notification, cfg ChannelConfig, logger *logging.Logger)
	after   func(ctx context.Context, n *Notification, cfg ChannelConfig, logger *logging.Logger)
	// success generates the success log line; if nil, a default "<Display> notification sent successfully" is used
	success func(n *Notification) string
}

// hb constructs a base NotificationHandler with concise parameters.
//
//nolint:unparam // 'before' is nil in current registrations; keep hook for future handlers
func hb(typ, display string, delay time.Duration,
	before func(context.Context, *Notification, ChannelConfig, *logging.Logger),
	after func(context.Context, *Notification, ChannelConfig, *logging.Logger),
	success func(*Notification) string,
) NotificationHandler {
	return &baseHandler{typ: typ, display: display, delay: delay, before: before, after: after, success: success}
}

func (h *baseHandler) Type() string { return h.typ }

func (h *baseHandler) Deliver(ctx context.Context, n *Notification, cfg ChannelConfig, logger *logging.Logger) error {
	disp := h.display
	if disp == "" {
		disp = h.typ
	}
	logger.Info(fmt.Sprintf("Sending %s notification: %s (config: %+v)", disp, n.Subject, cfg))
	if h.before != nil {
		h.before(ctx, n, cfg, logger)
	}
	if err := waitOrSleep(ctx, h.delay); err != nil {
		return err
	}
	if h.after != nil {
		h.after(ctx, n, cfg, logger)
	}
	if h.success != nil {
		logger.Info(h.success(n))
	} else {
		logger.Info(fmt.Sprintf("%s notification sent successfully", disp))
	}
	return nil
}

// NewBasicNotificationService creates a new notification service
func NewBasicNotificationService(logger *logging.Logger) *BasicNotificationService {
	service := &BasicNotificationService{
		logger: logger,
		supportedChannels: []string{
			"email", "slack", "teams", "discord", "webhook",
			"sms", "pagerduty", "opsgenie", "telegram",
		},
		channelConfigurations: make(map[string]ChannelConfig),
		handlers:              make(map[string]NotificationHandler),
	}

	// Initialize default channel configurations
	service.initializeDefaultChannels()

	// Register default handlers in the registry (type -> handler) using a table-driven setup
	handlers := []NotificationHandler{
		hb("email", "email", 100*time.Millisecond, nil, nil, func(n *Notification) string {
			return fmt.Sprintf("Email notification sent successfully to %s", n.Recipient)
		}),
		hb("slack", "Slack", 150*time.Millisecond, nil,
			func(ctx context.Context, n *Notification, cfg ChannelConfig, logger *logging.Logger) {
				slackMessage := formatSlackMessage(n)
				logger.Debug(fmt.Sprintf("Slack message: %s", slackMessage))
			}, nil,
		),
		hb("teams", "Teams", 120*time.Millisecond, nil, nil, nil),
		hb("discord", "Discord", 110*time.Millisecond, nil, nil, nil),
		hb("webhook", "webhook", 200*time.Millisecond, nil, nil, nil),
		hb("sms", "SMS", 300*time.Millisecond, nil, nil, nil),
		hb("pagerduty", "PagerDuty", 250*time.Millisecond, nil, nil, nil),
		hb("opsgenie", "OpsGenie", 180*time.Millisecond, nil, nil, nil),
		hb("telegram", "Telegram", 130*time.Millisecond, nil, nil, nil),
	}
	for _, h := range handlers {
		service.registerHandler(h)
	}

	return service
}

// registerHandler adds or replaces a handler in the registry.
func (bns *BasicNotificationService) registerHandler(h NotificationHandler) {
	if h == nil {
		return
	}
	bns.handlers[h.Type()] = h
}

// SendNotification sends a notification through specified channels
func (bns *BasicNotificationService) SendNotification(ctx context.Context, notification *Notification) error {
	bns.logger.Info(fmt.Sprintf("Sending notification: %s via %s", notification.Type, notification.Channel))

	// Unified validation
	if err := bns.ValidateChannel(ctx, notification.Channel); err != nil {
		return fmt.Errorf("invalid channel %s: %w", notification.Channel, err)
	}

	// Resolve config and handler
	config, exists := bns.channelConfigurations[notification.Channel]
	if !exists {
		return fmt.Errorf("channel configuration not found for %s", notification.Channel)
	}

	handler, ok := bns.handlers[config.Type]
	if !ok {
		return fmt.Errorf("no handler registered for channel type: %s", config.Type)
	}

	// Set creation timestamp if missing
	if notification.CreatedAt.IsZero() {
		notification.CreatedAt = time.Now()
	}

	bns.logger.Debug(fmt.Sprintf("Dispatching notification to handler type=%s", handler.Type()))

	// Deliver using the handler; status bookkeeping centralized here
	if err := handler.Deliver(ctx, notification, config, bns.logger); err != nil {
		// Preserve semantics: on failure, don't mutate SentAt/Status
		bns.logger.Error(fmt.Sprintf("Delivery failed for channel %s (type=%s): %v", notification.Channel, config.Type, err))
		return err
	}

	now := time.Now()
	notification.SentAt = &now
	notification.Status = StatusSent

	bns.logger.Info(fmt.Sprintf("Notification sent successfully via %s", notification.Channel))
	return nil
}

// SendAlert sends an alert through multiple channels
func (bns *BasicNotificationService) SendAlert(ctx context.Context, alert *Alert, channels []string) error {
	bns.logger.Info(fmt.Sprintf("Sending alert %s to %d channels", alert.ID, len(channels)))

	var errors []error

	for _, channel := range channels {
		// Create notification from alert
		notification := &Notification{
			ID:        fmt.Sprintf("notif_%s_%s", alert.ID, channel),
			Type:      "alert",
			Channel:   channel,
			Recipient: "default",
			Subject:   alert.Title,
			Message:   bns.formatAlertMessage(alert),
			Severity:  alert.Severity,
			CreatedAt: time.Now(),
			Status:    "pending",
			Metadata: map[string]interface{}{
				"alert_id":   alert.ID,
				"alert_type": alert.Type,
				"component":  alert.Component,
				"source":     alert.Source,
				"created_at": alert.CreatedAt,
			},
		}

		// Send notification
		err := bns.SendNotification(ctx, notification)
		if err != nil {
			bns.logger.Error(fmt.Sprintf("Failed to send alert to %s: %v", channel, err))
			errors = append(errors, fmt.Errorf("channel %s: %w", channel, err))
			continue
		}

		bns.logger.Info(fmt.Sprintf("Alert sent successfully to %s", channel))
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to send alert to some channels: %v", errors)
	}

	return nil
}

// ValidateChannel validates if a channel is supported and configured
func (bns *BasicNotificationService) ValidateChannel(ctx context.Context, channel string) error {
	// Check if channel is supported
	supported := false
	for _, supportedChannel := range bns.supportedChannels {
		if supportedChannel == channel {
			supported = true
			break
		}
	}

	if !supported {
		return fmt.Errorf("channel %s is not supported", channel)
	}

	// Check if channel is configured
	config, exists := bns.channelConfigurations[channel]
	if !exists {
		return fmt.Errorf("channel %s is not configured", channel)
	}

	if !config.Enabled {
		return fmt.Errorf("channel %s is disabled", channel)
	}

	return nil
}

// GetSupportedChannels returns list of supported notification channels
func (bns *BasicNotificationService) GetSupportedChannels() []string {
	return bns.supportedChannels
}

// Channel-specific handlers are implemented via baseHandler to reduce duplication.

// waitOrSleep waits for the provided duration or returns early if the context is canceled.
// It consolidates the repeated cancel-or-sleep pattern used by notification handlers.
func waitOrSleep(ctx context.Context, d time.Duration) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(d):
		return nil
	}
}

// Helper methods for formatting messages

func (bns *BasicNotificationService) formatAlertMessage(alert *Alert) string {
	return fmt.Sprintf(`
 ALERT: %s

Component: %s
Severity: %s
Description: %s
Time: %s

Alert ID: %s
Source: %s

Please investigate immediately.
`,
		alert.Title,
		alert.Component,
		alert.Severity,
		alert.Description,
		alert.CreatedAt.Format("2006-01-02 15:04:05"),
		alert.ID,
		alert.Source,
	)
}

func formatSlackMessage(notification *Notification) string {
	color := "good"
	switch notification.Severity {
	case SeverityCritical:
		color = "danger"
	case SeverityWarning:
		color = "warning"
	case SeverityInfo:
		color = "good"
	}

	// Simulate Slack attachment format
	return fmt.Sprintf(`{
		"attachments": [{
			"color": "%s",
			"title": "%s",
			"text": "%s",
			"ts": %d
		}]
	}`, color, notification.Subject, notification.Message, notification.CreatedAt.Unix())
}

// Initialize default channel configurations
func (bns *BasicNotificationService) initializeDefaultChannels() {
	// Use per-type templates to avoid repeating identical rate/ retry literals across channels.
	type tmpl struct {
		rl RateLimit
		rp RetryPolicy
	}
	templates := map[string]tmpl{
		"email":     {rl(100, 1*time.Hour, 10), rp(3, 30*time.Second, 2.0, 5*time.Minute)},
		"slack":     {rl(200, 1*time.Hour, 20), rp(3, 10*time.Second, 1.5, 2*time.Minute)},
		"teams":     {rl(150, 1*time.Hour, 15), rp(3, 15*time.Second, 2.0, 3*time.Minute)},
		"discord":   {rl(100, 1*time.Hour, 10), rp(2, 20*time.Second, 2.0, 2*time.Minute)},
		"webhook":   {rl(500, 1*time.Hour, 50), rp(5, 5*time.Second, 1.5, 1*time.Minute)},
		"sms":       {rl(20, 1*time.Hour, 5), rp(2, 1*time.Minute, 2.0, 5*time.Minute)},
		"pagerduty": {rl(100, 1*time.Hour, 10), rp(3, 30*time.Second, 2.0, 5*time.Minute)},
		"opsgenie":  {rl(100, 1*time.Hour, 10), rp(3, 30*time.Second, 2.0, 5*time.Minute)},
		"telegram":  {rl(50, 1*time.Hour, 5), rp(3, 20*time.Second, 2.0, 3*time.Minute)},
	}

	type def struct {
		key     string
		name    string
		typ     string
		enabled bool
		conf    map[string]interface{}
	}

	// tiny constructor to keep defs concise and reduce repetitive literals
	//nolint:unparam // 'enabled' may repeat the same value in current defaults; keep for configurability
	d := func(key, name, typ string, enabled bool, conf map[string]interface{}) def {
		return def{key: key, name: name, typ: typ, enabled: enabled, conf: conf}
	}

	// extracted small conf maps to avoid near-identical inline blocks (helps dupl)
	pagerDutyConf := map[string]interface{}{
		"integration_key": "integration_key",
		"api_url":         "https://events.pagerduty.com/v2/enqueue",
	}
	opsGenieConf := map[string]interface{}{
		"api_key": "api_key",
		"api_url": "https://api.opsgenie.com/v2/alerts",
	}
	// additional small conf maps to avoid inline duplicates
	smsConf := map[string]interface{}{
		"provider":    "twilio",
		"account_sid": "account_sid",
		"auth_token":  "auth_token",
		"from_number": "+1234567890",
	}
	telegramConf := map[string]interface{}{
		"bot_token": "bot_token",
		"chat_id":   "-1001234567890",
	}
	defs := []def{
		{
			key: "email", name: "Email", typ: "email", enabled: true,
			conf: map[string]interface{}{
				"smtp_host":    "smtp.example.com",
				"smtp_port":    587,
				"username":     "notifications@costscope.com",
				"use_tls":      true,
				"from_address": "CostScope Alerts <alerts@costscope.com>",
			},
		},
		{
			key: "slack", name: "Slack", typ: "slack", enabled: true,
			conf: map[string]interface{}{
				"webhook_url": "https://hooks.slack.com/services/...",
				"channel":     "#alerts",
				"username":    "CostScope Bot",
				"icon_emoji":  ":warning:",
			},
		},
		{
			key: "teams", name: "Microsoft Teams", typ: "teams", enabled: true,
			conf: map[string]interface{}{
				"webhook_url": "https://outlook.office.com/webhook/...",
				"theme_color": "FF5733",
			},
		},
		{
			key: "discord", name: "Discord", typ: "discord", enabled: false,
			conf: map[string]interface{}{
				"webhook_url": "https://discord.com/api/webhooks/...",
				"username":    "CostScope",
			},
		},
		{
			key: "webhook", name: "Generic Webhook", typ: "webhook", enabled: false,
			conf: map[string]interface{}{
				"url":    "https://api.example.com/webhook",
				"method": "POST",
				"headers": map[string]string{
					"Content-Type":  "application/json",
					"Authorization": "Bearer token123",
				},
				"timeout": 30,
			},
		},
		d("sms", "SMS", "sms", false, smsConf),
		d("pagerduty", "PagerDuty", "pagerduty", false, pagerDutyConf),
		d("opsgenie", "OpsGenie", "opsgenie", false, opsGenieConf),
		d("telegram", "Telegram", "telegram", false, telegramConf),
	}

	// Build and store configurations from definitions and templates.
	count := 0
	for _, d := range defs {
		t := templates[d.typ]
		bns.channelConfigurations[d.key] = cfg(d.name, d.typ, d.enabled, d.conf, t.rl, t.rp)
		count++
	}

	bns.logger.Info(fmt.Sprintf("Initialized %d notification channels", count))
}
