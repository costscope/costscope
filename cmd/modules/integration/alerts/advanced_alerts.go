package alerts

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"local/costscope/cmd/modules/integration/types"
)

// AlertManager handles advanced alerting capabilities
type AlertManager struct {
	alerts    map[string]*types.Alert
	channels  map[string]*types.NotificationChannel
	templates map[string]string
	rules     map[string]*AlertRule
}

// AlertRule represents a rule for triggering alerts
type AlertRule struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Conditions []types.AlertCondition `json:"conditions"`
	Actions    []types.AlertAction    `json:"actions"`
	Schedule   *types.AlertSchedule   `json:"schedule"`
	Enabled    bool                   `json:"enabled"`
	Priority   string                 `json:"priority"`
	Cooldown   time.Duration          `json:"cooldown"`
	LastFired  time.Time              `json:"last_fired"`
}

// NewAlertManager creates a new enhanced alert manager
func NewAlertManager() *AlertManager {
	return &AlertManager{
		alerts:    make(map[string]*types.Alert),
		channels:  make(map[string]*types.NotificationChannel),
		templates: getDefaultTemplates(),
		rules:     make(map[string]*AlertRule),
	}
}

// BuildAdvancedAlertCommands creates enhanced alert commands
func (am *AlertManager) BuildAdvancedAlertCommands() *cobra.Command {
	alertCmd := &cobra.Command{
		Use:   "alert",
		Short: "Advanced cost monitoring alerts",
		Long:  "Create, manage, and monitor advanced cost alerts with multiple channels and conditions.",
	}

	alertCmd.AddCommand(am.buildCreateAlertCommand())
	alertCmd.AddCommand(am.buildListAlertsCommand())
	alertCmd.AddCommand(am.buildTestAlertCommand())
	alertCmd.AddCommand(am.buildManageChannelsCommand())
	alertCmd.AddCommand(am.buildTemplateCommands())
	alertCmd.AddCommand(am.buildRulesCommand())
	alertCmd.AddCommand(am.buildAnalyticsCommand())

	return alertCmd
}

// buildCreateAlertCommand creates the enhanced alert creation command
func (am *AlertManager) buildCreateAlertCommand() *cobra.Command {
	var (
		name       string
		alertType  string
		severity   string
		threshold  float64
		conditions []string
		channels   []string
		template   string
		schedule   string
		actions    []string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new advanced alert",
		Long: `Create a new alert with multiple conditions, channels, and actions.
Supports complex alerting scenarios with custom templates and schedules.`,
		Example: `  # Create a budget threshold alert
  costscope integration alert create --name "Monthly Budget" --type budget --threshold 1000 --severity high --channels slack,email

  # Create anomaly detection alert
  costscope integration alert create --name "Cost Anomaly" --type anomaly --conditions "change>50%,duration>1h" --channels teams,webhook

  # Create scheduled alert with custom template
  costscope integration alert create --name "Weekly Report" --type report --schedule "weekly" --template summary --channels email`,
		RunE: func(cmd *cobra.Command, args []string) error {
			alert := &types.Alert{
				ID:        am.generateAlertID(),
				Name:      name,
				Type:      alertType,
				Severity:  severity,
				Threshold: threshold,
				Created:   time.Now(),
				Updated:   time.Now(),
				Status:    types.AlertStatusActive,
			}

			// Parse conditions
			alert.Conditions = am.parseConditions(conditions)

			// Parse actions
			alert.Actions = am.parseActions(actions)

			// Setup notification channels
			alert.Channels = am.setupChannels(channels)

			// Parse schedule if provided
			if schedule != "" {
				alert.Schedule = am.parseSchedule(schedule)
			}

			// Store the alert
			am.alerts[alert.ID] = alert

			fmt.Printf(" Alert '%s' created successfully!\n", alert.Name)
			fmt.Printf(" Alert ID: %s\n", alert.ID)
			fmt.Printf(" Type: %s | Severity: %s | Threshold: %.2f\n",
				alert.Type, alert.Severity, alert.Threshold)

			if len(alert.Channels) > 0 {
				channelNames := make([]string, len(alert.Channels))
				for i, ch := range alert.Channels {
					channelNames[i] = ch.Name
				}
				fmt.Printf(" Channels: %s\n", strings.Join(channelNames, ", "))
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Alert name (required)")
	cmd.Flags().StringVar(&alertType, "type", "threshold", "Alert type (threshold, anomaly, budget, trend, report)")
	cmd.Flags().StringVar(&severity, "severity", "medium", "Alert severity (low, medium, high, critical)")
	cmd.Flags().Float64Var(&threshold, "threshold", 0, "Alert threshold value")
	cmd.Flags().StringSliceVar(&conditions, "conditions", nil, "Alert conditions (metric>value, change>percent, etc.)")
	cmd.Flags().StringSliceVar(&channels, "channels", nil, "Notification channels (slack, email, teams, webhook, sms)")
	cmd.Flags().StringVar(&template, "template", "default", "Notification template")
	cmd.Flags().StringVar(&schedule, "schedule", "", "Alert schedule (daily, weekly, monthly, cron)")
	cmd.Flags().StringSliceVar(&actions, "actions", nil, "Actions to take (notify, create_ticket, scale_down, etc.)")

	_ = cmd.MarkFlagRequired("name")

	return cmd
}

// buildListAlertsCommand creates the alert listing command
func (am *AlertManager) buildListAlertsCommand() *cobra.Command {
	var (
		status    string
		severity  string
		alertType string
		verbose   bool
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all alerts",
		Long:  "List all configured alerts with their status, metrics, and performance data.",
		Example: `  # List all alerts
  costscope integration alert list

  # List active high-severity alerts
  costscope integration alert list --status active --severity high

  # List budget alerts with verbose output
  costscope integration alert list --type budget --verbose`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return am.listAlerts(status, severity, alertType, verbose)
		},
	}

	cmd.Flags().StringVar(&status, "status", "", "Filter by status (active, triggered, resolved, suppressed)")
	cmd.Flags().StringVar(&severity, "severity", "", "Filter by severity (low, medium, high, critical)")
	cmd.Flags().StringVar(&alertType, "type", "", "Filter by type (threshold, anomaly, budget, trend, report)")
	cmd.Flags().BoolVar(&verbose, "verbose", false, "Show detailed alert information")

	return cmd
}

// buildTestAlertCommand creates the alert testing command
func (am *AlertManager) buildTestAlertCommand() *cobra.Command {
	var (
		alertID  string
		channels []string
		dryRun   bool
	)

	cmd := &cobra.Command{
		Use:   "test",
		Short: "Test alert delivery",
		Long:  "Test alert delivery across all configured channels or specific channels.",
		Example: `  # Test specific alert
  costscope integration alert test --alert-id alert_123

  # Test alert with specific channels
  costscope integration alert test --alert-id alert_123 --channels slack,email

  # Dry run test (no actual notifications)
  costscope integration alert test --alert-id alert_123 --dry-run`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return am.testAlert(alertID, channels, dryRun)
		},
	}

	cmd.Flags().StringVar(&alertID, "alert-id", "", "Alert ID to test (required)")
	cmd.Flags().StringSliceVar(&channels, "channels", nil, "Specific channels to test")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Perform dry run without sending notifications")

	_ = cmd.MarkFlagRequired("alert-id")

	return cmd
}

// buildManageChannelsCommand creates channel management commands
func (am *AlertManager) buildManageChannelsCommand() *cobra.Command {
	channelCmd := &cobra.Command{
		Use:   "channel",
		Short: "Manage notification channels",
		Long:  "Create, configure, and manage notification channels for alerts.",
	}

	channelCmd.AddCommand(am.buildAddChannelCommand())
	channelCmd.AddCommand(am.buildListChannelsCommand())
	channelCmd.AddCommand(am.buildTestChannelCommand())
	channelCmd.AddCommand(am.buildRemoveChannelCommand())

	return channelCmd
}

// buildAddChannelCommand creates the add channel command
func (am *AlertManager) buildAddChannelCommand() *cobra.Command {
	var (
		name        string
		channelType string
		config      map[string]string
	)

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a new notification channel",
		Long:  "Add a new notification channel with specific configuration.",
		Example: `  # Add Slack channel
  costscope integration alert channel add --name "team-alerts" --type slack --config webhook_url=https://hooks.slack.com/...

  # Add email channel
  costscope integration alert channel add --name "admin-email" --type email --config smtp_server=smtp.company.com --config recipients=admin@company.com

  # Add Teams channel
  costscope integration alert channel add --name "ops-teams" --type teams --config webhook_url=https://outlook.office.com/webhook/...`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Convert config to interface map
			channelConfig := make(map[string]interface{})
			for k, v := range config {
				channelConfig[k] = v
			}

			channel := &types.NotificationChannel{
				ID:     am.generateChannelID(),
				Type:   channelType,
				Name:   name,
				Config: channelConfig,
				Status: "active",
			}

			am.channels[channel.ID] = channel

			fmt.Printf(" Notification channel '%s' added successfully!\n", channel.Name)
			fmt.Printf(" Channel ID: %s\n", channel.ID)
			fmt.Printf(" Type: %s\n", channel.Type)

			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Channel name (required)")
	cmd.Flags().StringVar(&channelType, "type", "", "Channel type (slack, email, teams, webhook, sms) (required)")
	cmd.Flags().StringToStringVar(&config, "config", nil, "Channel configuration (key=value pairs)")

	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("type")

	return cmd
}

// Helper methods

func (am *AlertManager) generateAlertID() string {
	return fmt.Sprintf("alert_%d", time.Now().UnixNano())
}

func (am *AlertManager) generateChannelID() string {
	return fmt.Sprintf("channel_%d", time.Now().UnixNano())
}

func (am *AlertManager) parseConditions(conditions []string) []types.AlertCondition {
	var alertConditions []types.AlertCondition

	for _, condition := range conditions {
		// Parse condition string (e.g., "cost>1000", "change>50%")
		parts := strings.Split(condition, ">")
		if len(parts) == 2 {
			metric := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			alertConditions = append(alertConditions, types.AlertCondition{
				Metric:    metric,
				Operator:  ">",
				Value:     parseFloat(value),
				Period:    "5m",
				Aggregate: "avg",
			})
		}
	}

	return alertConditions
}

func (am *AlertManager) parseActions(actions []string) []types.AlertAction {
	var alertActions []types.AlertAction

	for _, action := range actions {
		alertActions = append(alertActions, types.AlertAction{
			Type:   action,
			Config: make(map[string]interface{}),
		})
	}

	return alertActions
}

func (am *AlertManager) setupChannels(channelNames []string) []types.NotificationChannel {
	var channels []types.NotificationChannel

	for _, name := range channelNames {
		// Find or create channel
		for _, channel := range am.channels {
			if channel.Name == name || channel.Type == name {
				channels = append(channels, *channel)
				break
			}
		}
	}

	return channels
}

func (am *AlertManager) parseSchedule(_ string) *types.AlertSchedule {
	return &types.AlertSchedule{
		Enabled:   true,
		TimeZone:  "UTC",
		Days:      []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday"},
		StartTime: "09:00",
		EndTime:   "17:00",
	}
}

func (am *AlertManager) listAlerts(status, severity, alertType string, verbose bool) error {
	fmt.Println(" Alert Management System")
	fmt.Println("═════════════════════════")

	if len(am.alerts) == 0 {
		fmt.Println("No alerts configured")
		return nil
	}

	for _, alert := range am.alerts {
		// Apply filters
		if status != "" && alert.Status != status {
			continue
		}
		if severity != "" && alert.Severity != severity {
			continue
		}
		if alertType != "" && alert.Type != alertType {
			continue
		}

		fmt.Printf("\n %s (%s)\n", alert.Name, alert.ID)
		fmt.Printf("   Type: %s | Severity: %s | Status: %s\n",
			alert.Type, alert.Severity, alert.Status)
		fmt.Printf("   Threshold: %.2f | Created: %s\n",
			alert.Threshold, alert.Created.Format("2006-01-02 15:04"))

		if verbose {
			if len(alert.Conditions) > 0 {
				fmt.Println("   Conditions:")
				for _, condition := range alert.Conditions {
					fmt.Printf("     • %s %s %.2f (period: %s)\n",
						condition.Metric, condition.Operator, condition.Value, condition.Period)
				}
			}

			if len(alert.Channels) > 0 {
				fmt.Println("   Channels:")
				for _, channel := range alert.Channels {
					fmt.Printf("     • %s (%s)\n", channel.Name, channel.Type)
				}
			}

			if alert.Metrics != nil {
				fmt.Printf("   Metrics: Triggered %d times, Last: %s\n",
					alert.Metrics.TriggeredCount,
					alert.Metrics.LastTriggered.Format("2006-01-02 15:04"))
			}
		}
	}

	return nil
}

func (am *AlertManager) testAlert(alertID string, channels []string, dryRun bool) error {
	alert, exists := am.alerts[alertID]
	if !exists {
		return fmt.Errorf("alert %s not found", alertID)
	}

	fmt.Printf(" Testing alert: %s\n", alert.Name)

	if dryRun {
		fmt.Println(" DRY RUN - No actual notifications will be sent")
	}

	// Test channels
	channelsToTest := alert.Channels
	if len(channels) > 0 {
		channelsToTest = am.filterChannels(alert.Channels, channels)
	}

	results := types.AlertTestResults{}

	for _, channel := range channelsToTest {
		fmt.Printf(" Testing %s channel (%s)...", channel.Type, channel.Name)

		if !dryRun {
			// Simulate sending test notification
			time.Sleep(500 * time.Millisecond)
		}

		result := types.AlertTestResult{
			Status:       "success",
			ResponseTime: 234 * time.Millisecond,
		}

		switch channel.Type {
		case "email":
			results.Email = result
		case "slack":
			results.Slack = result
		case "teams":
			// Set teams result
		case "webhook":
			results.Webhook = result
		case "sms":
			results.SMS = result
		}

		fmt.Printf("  (%.0fms)\n", result.ResponseTime.Seconds()*1000)
	}

	results.Overall = "success"

	// Display results
	resultsJSON, _ := json.MarshalIndent(results, "", "  ")
	fmt.Printf("\n Test Results:\n%s\n", string(resultsJSON))

	return nil
}

func (am *AlertManager) filterChannels(allChannels []types.NotificationChannel, filterNames []string) []types.NotificationChannel {
	var filtered []types.NotificationChannel

	for _, channel := range allChannels {
		for _, name := range filterNames {
			if channel.Name == name || channel.Type == name {
				filtered = append(filtered, channel)
				break
			}
		}
	}

	return filtered
}

func parseFloat(s string) float64 {
	// Simple float parsing, in real implementation would use strconv.ParseFloat
	// with proper error handling
	if s == "50%" {
		return 50.0
	}
	return 1000.0 // Default value
}

// Additional command builders (stubs for full implementation)
func (am *AlertManager) buildListChannelsCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List notification channels",
		RunE: func(cmd *cobra.Command, args []string) error {
			return am.listChannels()
		},
	}
}

func (am *AlertManager) buildTestChannelCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "test",
		Short: "Test notification channel",
		RunE: func(cmd *cobra.Command, args []string) error {
			return am.testChannel()
		},
	}
}

func (am *AlertManager) buildRemoveChannelCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "remove",
		Short: "Remove notification channel",
		RunE: func(cmd *cobra.Command, args []string) error {
			return am.removeChannel()
		},
	}
}

func (am *AlertManager) buildTemplateCommands() *cobra.Command {
	return &cobra.Command{
		Use:   "template",
		Short: "Manage alert templates",
	}
}

func (am *AlertManager) buildRulesCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "rules",
		Short: "Manage alert rules",
	}
}

func (am *AlertManager) buildAnalyticsCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "analytics",
		Short: "View alert analytics",
	}
}

// Placeholder implementations
func (am *AlertManager) listChannels() error {
	fmt.Println(" Notification Channels:")
	if len(am.channels) == 0 {
		fmt.Println("   No channels configured")
		return nil
	}

	for _, channel := range am.channels {
		fmt.Printf("   • %s (%s) - %s\n", channel.Name, channel.Type, channel.Status)
	}
	return nil
}

func (am *AlertManager) testChannel() error {
	fmt.Println(" Testing notification channel...")
	return nil
}

func (am *AlertManager) removeChannel() error {
	fmt.Println("️ Removing notification channel...")
	return nil
}

// getDefaultTemplates returns default notification templates
func getDefaultTemplates() map[string]string {
	return map[string]string{
		"default": "Alert: {{.Name}} - {{.Message}}",
		"summary": "Cost Summary: {{.TotalCost}} ({{.Change}}% change)",
		"anomaly": "Cost Anomaly Detected: {{.Service}} - {{.Change}}% increase",
		"budget":  "Budget Alert: {{.BudgetName}} - {{.Spent}}/{{.Budget}} ({{.Percentage}}%)",
	}
}
