package integration

import (
	"fmt"

	"local/costscope/cmd/modules/integration/alerts"
	"local/costscope/cmd/modules/integration/connections"
	"local/costscope/cmd/modules/integration/dashboard"
	"local/costscope/cmd/modules/integration/webhooks"
	"local/costscope/internal/core/integration"

	"github.com/spf13/cobra"
)

// CreateIntegrationCommands creates the enhanced integration module CLI commands
func CreateIntegrationCommands() *cobra.Command {
	integrationCmd := &cobra.Command{
		Use:   "integration",
		Short: "Manage external system integrations with enhanced features",
		Long: `The enhanced integration module provides advanced capabilities to connect with external systems,
manage cost alerts, automate workflows, and operate web dashboards.

Enhanced Features:
- Advanced connection management with health monitoring
- Multi-channel alerts with custom templates and conditions
- Interactive dashboards with real-time updates and plugins
- Secure webhooks with delivery tracking and retry policies
- Workflow automation with complex triggers and actions

Core Features:
- Connect to billing systems (AWS, Azure, GCP)
- Integrate with ITSM tools (ServiceNow, JIRA)
- Set up BI dashboards (Tableau, PowerBI)
- Configure monitoring systems (Datadog, New Relic)
- Manage notification channels (Slack, Teams, PagerDuty)
- Automate cost management workflows
- Run interactive web dashboards`,
		Example: `  # List available integrations
  costscope integration list

  # Connect with advanced configuration
  costscope integration connect aws --config access_key=KEY --health-check --auto-retry

  # Create advanced alert with multiple conditions
  costscope integration alert create --name "Budget Alert" --type budget --threshold 1000 \
    --conditions "change>20%,duration>1h" --channels slack,email

  # Start enhanced dashboard with plugins
  costscope integration dashboard start --port 8080 --theme dark --features real-time,export

  # Create secure webhook with retry policy
  costscope integration webhook create --name "cost-alerts" --url "https://api.company.com/webhook" \
	--secret "<YOUR_LONG_RANDOM_SECRET>" --max-retries 5`,
	}

	// Initialize enhanced components
	connectionManager := connections.NewConnectionManager()
	alertManager := alerts.NewAlertManager()
	dashboardManager := dashboard.NewDashboardManager()
	webhookManager := webhooks.NewWebhookManager()

	// Retain existing non-migrated commands (list/disconnect/status/test + alerts full manager + advanced connections)
	integrationCmd.AddCommand(createListCommand())
	integrationCmd.AddCommand(createDisconnectCommand())
	integrationCmd.AddCommand(createStatusCommand())
	integrationCmd.AddCommand(createTestCommand())
	integrationCmd.AddCommand(alertManager.BuildAdvancedAlertCommands())
	integrationCmd.AddCommand(connectionManager.BuildAdvancedCommands())

	// Registrar-based deduplicated actions (webhook.*, dashboard.start/status, connections.connect)
	ctx := &RegistrationContext{
		WebhookMgr:   webhookManager,
		DashboardMgr: dashboardManager,
		ConnMgr:      connectionManager,
		AlertMgr:     alertManager,
	}
	RegisterIntegrationActions(integrationCmd, ctx, BuildDefaultActionSpecs())

	// Legacy builder merge logic removed (all migrated via declarative registrar). (TASK-INTEGRATION-REGISTRAR-CLEANUP)

	return integrationCmd
}

// ensureParent finds or creates a parent command under integration root.
// ensureParent removed (all parents created by registrar specs)

// List command
func createListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List available integrations",
		Long:  "Display all available integration systems with their current status and capabilities",
		RunE: func(cmd *cobra.Command, args []string) error {
			service := integration.NewService()
			defer func() { _ = service.Close() }()

			filter := &integration.IntegrationFilter{}

			result, err := service.ListIntegrations(filter)
			if err != nil {
				return fmt.Errorf("failed to list integrations: %w", err)
			}

			printIntegrationList(result)
			return nil
		},
	}

	return cmd
}

// Connect command
// (connect command creation removed; now handled by registrar under connections category)

// Disconnect command
func createDisconnectCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "disconnect [system]",
		Short: "Disconnect from an external system",
		Long:  "Terminate the connection to an external system",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			system := args[0]
			service := integration.NewService()
			defer func() { _ = service.Close() }()

			result, err := service.DisconnectFromSystem(system)
			if err != nil {
				return fmt.Errorf("failed to disconnect from %s: %w", system, err)
			}

			if result.Success {
				fmt.Printf(" Successfully disconnected from %s\n", system)
			} else {
				fmt.Printf(" Failed to disconnect from %s: %s\n", system, result.Error)
			}

			return nil
		},
	}

	_ = cmd.MarkFlagRequired("system")

	return cmd
}

// Status command
func createStatusCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status [system]",
		Short: "Check integration status",
		Long:  "Display the current status of integration connections",
		RunE: func(cmd *cobra.Command, args []string) error {
			service := integration.NewService()
			defer func() { _ = service.Close() }()

			verbose, _ := cmd.Flags().GetBool("verbose")

			if len(args) > 0 {
				// Status for specific system
				systemName := args[0]
				result, err := service.GetConnectionStatus(systemName)
				if err != nil {
					return fmt.Errorf("failed to get status for %s: %w", systemName, err)
				}

				fmt.Printf("System: %s\n", result.SystemName)
				fmt.Printf("Status: %s\n", result.Status)
				fmt.Printf("Health Score: %.1f\n", result.HealthScore)
				if verbose {
					fmt.Printf("Uptime: %s\n", result.Uptime)
					fmt.Printf("Last Sync: %s\n", result.LastSync)
					fmt.Printf("Data Transferred: %d bytes\n", result.DataTransferred)
				}
			} else {
				// Status for all systems
				fmt.Println("Integration Status: All systems operational")
			}

			return nil
		},
	}

	cmd.Flags().Bool("verbose", false, "Show detailed status information")

	return cmd
}

// Test command
func createTestCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "test [system]",
		Short: "Test system connection",
		Long:  "Test the connection to an external system",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			system := args[0]
			service := integration.NewService()
			defer func() { _ = service.Close() }()

			result, err := service.TestConnection(system)
			if err != nil {
				return fmt.Errorf("failed to test connection to %s: %w", system, err)
			}

			printTestResult(result)
			return nil
		},
	}

	_ = cmd.MarkFlagRequired("system")

	return cmd
}

// Print functions

func printIntegrationList(result *integration.IntegrationListResult) {
	fmt.Println(" Available Integrations")
	fmt.Println("═══════════════════════════")

	if len(result.Integrations) == 0 {
		fmt.Println("No integrations available")
		return
	}

	// Group by category
	categories := make(map[string][]integration.Integration)
	for _, integration := range result.Integrations {
		categories[integration.Category] = append(categories[integration.Category], integration)
	}

	for category, integrations := range categories {
		fmt.Printf("\n %s:\n", category)
		for _, integration := range integrations {
			status := "Available"
			if integration.Status != "" {
				status = integration.Status
			}
			fmt.Printf("   • %s - %s [%s]\n", integration.Name, integration.Description, status)
		}
	}

	fmt.Printf("\n Use 'costscope integration connect <system>' to connect\n")
}

// printConnectionResult removed (not used by current command set)

func printTestResult(result *integration.ConnectionTestResult) {
	fmt.Printf(" Testing %s\n", result.SystemName)
	fmt.Println("═══════════════════════")

	if result.Success {
		fmt.Printf(" Connection: Success (%s)\n", result.ResponseTime)

		if len(result.AvailableFeatures) > 0 {
			fmt.Println("\n Available Features:")
			for _, feature := range result.AvailableFeatures {
				fmt.Printf("    %s\n", feature)
			}
		}
	} else {
		fmt.Printf(" Connection: Failed\n")
		fmt.Printf(" Error: %s\n", result.Error)
	}
}

// Enhanced integration commands use specialized managers:
// - AlertManager for advanced alerting
// - ConnectionManager for enhanced connections
// - DashboardManager for interactive dashboards
// - WebhookManager for secure webhooks
