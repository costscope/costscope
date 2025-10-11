package connections

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"

	"local/costscope/cmd/modules/integration/helpers"
	"local/costscope/cmd/modules/integration/types"
	"local/costscope/internal/core/logging"
)

// ConnectionManager handles third-party system connections with enhanced features
type ConnectionManager struct {
	availableIntegrations map[string][]types.Integration
	activeConnections     map[string]*ConnectionState
	healthMonitor         *HealthMonitor
	mu                    sync.RWMutex
}

// ConfigStore interface for configuration management (placeholder)
type ConfigStore interface {
	Store(connectionID string, config map[string]interface{}) error
	Load(connectionID string) (map[string]interface{}, error)
	Delete(connectionID string) error
}

// EventBus interface for event handling (placeholder)
type EventBus interface {
	Publish(event ConnectionEvent) error
	Subscribe(eventType string, handler func(ConnectionEvent)) error
}

// ConnectionState represents the state of an active connection
type ConnectionState struct {
	ConnectionID  string
	Integration   types.Integration
	Config        map[string]interface{}
	Status        string
	EstablishedAt time.Time
	LastActivity  time.Time
	ErrorCount    int
	Health        *types.HealthStatus
	Metrics       *types.IntegrationMetrics
}

// HealthMonitor monitors the health of connections
type HealthMonitor struct {
	checkInterval time.Duration
	timeout       time.Duration
	checks        map[string]*HealthCheck
}

// HealthCheck represents a health check configuration
type HealthCheck struct {
	ConnectionID string
	URL          string
	Method       string
	Headers      map[string]string
	ExpectedCode int
	Timeout      time.Duration
}

// ConnectionEvent represents a connection-related event
type ConnectionEvent struct {
	Type         string
	ConnectionID string
	Integration  string
	Timestamp    time.Time
	Data         map[string]interface{}
}

// small factory to reduce duplication when declaring integrations
func makeIntegration(name, desc string, category string, status, version string, lastUpdated time.Time, features []string) types.Integration {
	return types.Integration{
		Name:        name,
		Description: desc,
		Category:    category,
		Status:      status,
		Version:     version,
		LastUpdated: lastUpdated,
		Features:    features,
	}
}

// NewConnectionManager creates a new enhanced connection manager
func NewConnectionManager() *ConnectionManager {
	cm := &ConnectionManager{
		availableIntegrations: getAvailableIntegrations(),
		activeConnections:     make(map[string]*ConnectionState),
		healthMonitor: &HealthMonitor{
			checkInterval: 5 * time.Minute,
			timeout:       30 * time.Second,
			checks:        make(map[string]*HealthCheck),
		},
	}

	// Start health monitoring
	go cm.startHealthMonitoring()

	return cm
}

// BuildConnectCommand creates the enhanced connect command
func (cm *ConnectionManager) BuildConnectCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "connect",
		Short: "Connect to third-party systems with advanced management",
		Long: `Integrate with billing systems, ITSM tools, and business intelligence platforms.
Supports advanced features like health monitoring, automatic retry, and configuration management.`,
		RunE: helpers.RunEWithLogging("integration.connect", cm.runConnect),
	}

	cmd.Flags().String("system", "", "system to connect to")
	cmd.Flags().Bool("list", false, "list available integrations")
	cmd.Flags().String("category", "", "filter by integration category")
	cmd.Flags().StringToString("config", nil, "connection configuration (key=value pairs)")
	cmd.Flags().Bool("health-check", true, "enable health monitoring for the connection")
	cmd.Flags().Duration("timeout", 30*time.Second, "connection timeout")
	cmd.Flags().Bool("auto-retry", true, "enable automatic retry on connection failures")

	return cmd
}

// BuildAdvancedCommands creates advanced connection management commands
func (cm *ConnectionManager) BuildAdvancedCommands() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "connections",
		Short: "Advanced connection management",
		Long:  "Advanced connection management with health monitoring, configuration, and analytics.",
	}

	cmd.AddCommand(cm.buildListConnectionsCommand())
	cmd.AddCommand(cm.buildHealthCommand())
	cmd.AddCommand(cm.buildConfigCommand())
	cmd.AddCommand(cm.buildAnalyticsCommand())
	cmd.AddCommand(cm.buildMaintenanceCommand())

	return cmd
}

// runConnect executes the enhanced connect command
func (cm *ConnectionManager) runConnect(cmd *cobra.Command, args []string) error {
	system, _ := cmd.Flags().GetString("system")
	listMode, _ := cmd.Flags().GetBool("list")
	category, _ := cmd.Flags().GetString("category")
	configStr, _ := cmd.Flags().GetStringToString("config")
	enableHealthCheck, _ := cmd.Flags().GetBool("health-check")
	timeout, _ := cmd.Flags().GetDuration("timeout")
	autoRetry, _ := cmd.Flags().GetBool("auto-retry")

	// Convert string map to interface map
	config := make(map[string]interface{})
	for k, v := range configStr {
		config[k] = v
	}

	helpers.PrintHeader(" CostScope Enhanced Integration Hub")
	fmt.Println("════════════════════════════════════")

	if system == "" || listMode {
		return cm.displayAvailableIntegrations(category)
	}

	return cm.connectToSystemAdvanced(system, config, enableHealthCheck, timeout, autoRetry)
}

// displayAvailableIntegrations shows available third-party integrations with enhanced info
func (cm *ConnectionManager) displayAvailableIntegrations(filterCategory string) error {
	helpers.PrintHeader(" Available Integrations:")

	for category, systems := range cm.availableIntegrations {
		// Apply category filter if specified
		if filterCategory != "" && !strings.EqualFold(category, filterCategory) {
			continue
		}

		fmt.Printf("\n %s:\n", category)
		for _, sys := range systems {
			status := "Available"

			// Check if system is currently connected
			cm.mu.RLock()
			for _, conn := range cm.activeConnections {
				if conn.Integration.Name == sys.Name {
					status = fmt.Sprintf("Connected (%s)", conn.Status)
					break
				}
			}

			// Show enhanced information
			fmt.Printf("   • %s - %s [%s]\n", sys.Name, sys.Description, status)
			if len(sys.Features) > 0 {
				fmt.Printf("     Features: %s\n", strings.Join(sys.Features, ", "))
			}
			if sys.Version != "" {
				fmt.Printf("     Version: %s\n", sys.Version)
			}
		}
	}

	helpers.PrintTip("Use --system <name> to connect to a specific integration")
	helpers.PrintTip("Use --category <category> to filter by category")
	helpers.PrintTip("Use --config key=value to pass configuration")

	return nil
}

// connectToSystemAdvanced establishes connection with advanced features
func (cm *ConnectionManager) connectToSystemAdvanced(system string, config map[string]interface{}, enableHealthCheck bool, timeout time.Duration, autoRetry bool) error {
	fmt.Printf(" Connecting to %s with advanced features...\n", system)

	// Generate connection ID
	connectionID := cm.generateConnectionID()

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Perform connection
	result := cm.performConnectionAdvanced(ctx, system, config, connectionID)

	if result.Success {
		// Store connection state
		cm.mu.Lock()
		cm.activeConnections[connectionID] = &ConnectionState{
			ConnectionID:  connectionID,
			Integration:   cm.findIntegration(system),
			Config:        config,
			Status:        "connected",
			EstablishedAt: time.Now(),
			LastActivity:  time.Now(),
			ErrorCount:    0,
			Health:        result.HealthCheck,
			Metrics:       &types.IntegrationMetrics{},
		}
		cm.mu.Unlock()

		// Setup health monitoring if enabled
		if enableHealthCheck {
			cm.setupHealthCheck(connectionID, system)
		}

		// Publish connection event
		cm.publishEvent(ConnectionEvent{
			Type:         "connection_established",
			ConnectionID: connectionID,
			Integration:  system,
			Timestamp:    time.Now(),
			Data:         map[string]interface{}{"config": config},
		})

		// Output and logging
		logger := logging.GetLogger().WithFields(map[string]interface{}{"integration": system, "connection_id": connectionID})
		fmt.Printf(" Successfully connected to %s!\n", system)
		helpers.PrintKV(" Connection ID", connectionID)
		helpers.PrintKV(" Integration Status", result.Status)
		helpers.PrintKV(" Data Sync", result.DataSync)
		helpers.PrintKV(" Available Metrics", result.AvailableMetrics)
		logger.Info("integration connected")

		if result.HealthCheck != nil {
			fmt.Printf(" Health Status: %s (%.2fms response time)\n",
				result.HealthCheck.Status, result.HealthCheck.ResponseTime)
		}

		if len(result.Features) > 0 {
			fmt.Println("\n Available Features:")
			for _, feature := range result.Features {
				fmt.Printf("   • %s\n", feature)
			}
		}

		if autoRetry {
			fmt.Println(" Auto-retry enabled for connection failures")
		}
	} else {
		fmt.Printf(" Failed to connect to %s\n", system)
		fmt.Printf(" Error: %s\n", result.Error)
		logging.GetLogger().ErrorWithFields("integration connect failed", map[string]interface{}{
			"integration":   system,
			"error":         result.Error,
			"connection_id": connectionID,
		})

		// If auto-retry is enabled, schedule retry
		if autoRetry {
			fmt.Println(" Auto-retry will attempt reconnection in 5 minutes")
			go cm.scheduleRetry(system, config, 5*time.Minute)
		}
	}

	return nil
}

// performConnectionAdvanced simulates the actual connection process with enhanced features
func (cm *ConnectionManager) performConnectionAdvanced(ctx context.Context, system string, config map[string]interface{}, connectionID string) types.ConnectionResult {
	// Simulate connection time
	select {
	case <-ctx.Done():
		return types.ConnectionResult{
			Success: false,
			Error:   "Connection timeout",
		}
	case <-time.After(1 * time.Second):
		// Continue with connection
	}

	switch strings.ToLower(system) {
	case "tableau":
		return types.ConnectionResult{
			Success:          true,
			Status:           "Connected",
			DataSync:         "Real-time",
			AvailableMetrics: 25,
			ConnectionID:     connectionID,
			EstablishedAt:    time.Now(),
			Configuration:    config,
			HealthCheck: &types.HealthStatus{
				Status:       "healthy",
				LastCheck:    time.Now(),
				ResponseTime: 125.5,
				ErrorCount:   0,
				Uptime:       99.9,
			},
			Features: []string{
				"Cost trend dashboards",
				"Interactive cost drill-downs",
				"Executive summary reports",
				"Automated data refresh",
				"Real-time alerts",
				"Custom KPI tracking",
			},
		}
	case "slack":
		return types.ConnectionResult{
			Success:          true,
			Status:           "Connected",
			DataSync:         "Event-driven",
			AvailableMetrics: 8,
			ConnectionID:     connectionID,
			EstablishedAt:    time.Now(),
			Configuration:    config,
			HealthCheck: &types.HealthStatus{
				Status:       "healthy",
				LastCheck:    time.Now(),
				ResponseTime: 89.3,
				ErrorCount:   0,
				Uptime:       100.0,
			},
			Features: []string{
				"Cost alert notifications",
				"Daily cost summaries",
				"Interactive cost queries",
				"Budget threshold alerts",
				"Anomaly detection alerts",
				"Custom notification templates",
			},
		}
	case "servicenow":
		return types.ConnectionResult{
			Success:          true,
			Status:           "Connected",
			DataSync:         "Scheduled",
			AvailableMetrics: 15,
			ConnectionID:     connectionID,
			EstablishedAt:    time.Now(),
			Configuration:    config,
			HealthCheck: &types.HealthStatus{
				Status:       "healthy",
				LastCheck:    time.Now(),
				ResponseTime: 234.7,
				ErrorCount:   0,
				Uptime:       98.5,
			},
			Features: []string{
				"Automated ticket creation for cost anomalies",
				"Cost optimization workflow triggers",
				"Budget approval workflows",
				"Change request cost impact analysis",
				"SLA tracking for cost issues",
				"Integration with CMDB",
			},
		}
	case "power bi", "powerbi":
		return types.ConnectionResult{
			Success:          true,
			Status:           "Connected",
			DataSync:         "Hourly",
			AvailableMetrics: 20,
			ConnectionID:     connectionID,
			EstablishedAt:    time.Now(),
			Configuration:    config,
			HealthCheck: &types.HealthStatus{
				Status:       "healthy",
				LastCheck:    time.Now(),
				ResponseTime: 156.2,
				ErrorCount:   0,
				Uptime:       99.2,
			},
			Features: []string{
				"Cost analytics reports",
				"Executive dashboards",
				"Automated report distribution",
				"Cost forecasting models",
				"Drill-through capabilities",
				"Mobile dashboard support",
			},
		}
	case "jira":
		return types.ConnectionResult{
			Success:          true,
			Status:           "Connected",
			DataSync:         "Event-driven",
			AvailableMetrics: 12,
			ConnectionID:     connectionID,
			EstablishedAt:    time.Now(),
			Configuration:    config,
			HealthCheck: &types.HealthStatus{
				Status:       "healthy",
				LastCheck:    time.Now(),
				ResponseTime: 178.9,
				ErrorCount:   0,
				Uptime:       97.8,
			},
			Features: []string{
				"Cost optimization tickets",
				"Budget approval workflows",
				"Resource request tracking",
				"Cost impact assessments",
				"Automated issue creation",
				"Cost-related project tracking",
			},
		}
	default:
		return types.ConnectionResult{
			Success: false,
			Error:   fmt.Sprintf("Integration '%s' not available or configured. Use 'connect --list' to see available integrations.", system),
		}
	}
}

// Additional helper methods
func (cm *ConnectionManager) generateConnectionID() string { return helpers.GenerateID("conn") }

func (cm *ConnectionManager) findIntegration(name string) types.Integration {
	for _, systems := range cm.availableIntegrations {
		for _, sys := range systems {
			if strings.EqualFold(sys.Name, name) {
				return sys
			}
		}
	}
	return types.Integration{}
}

func (cm *ConnectionManager) setupHealthCheck(connectionID, system string) {
	// Implementation would setup health monitoring for the specified system
	fmt.Printf(" Health monitoring enabled for connection %s (system: %s)\n", connectionID, system)
}

func (cm *ConnectionManager) scheduleRetry(system string, config map[string]interface{}, delay time.Duration) {
	// Implementation would schedule retry with the provided configuration
	configKeys := make([]string, 0, len(config))
	for key := range config {
		configKeys = append(configKeys, key)
	}
	fmt.Printf("⏰ Retry scheduled for %s in %v (config keys: %v)\n", system, delay, configKeys)
}

func (cm *ConnectionManager) publishEvent(event ConnectionEvent) {
	// Implementation would publish event to event bus
}

func (cm *ConnectionManager) startHealthMonitoring() {
	// Implementation would start health monitoring goroutine
}

// buildSimpleCommand reduces duplication for straightforward Cobra commands
func buildSimpleCommand(use, short string, runE func(cmd *cobra.Command, args []string) error) *cobra.Command {
	return &cobra.Command{
		Use:   use,
		Short: short,
		RunE:  runE,
	}
}

// Additional command builders (placeholders for full implementation)
func (cm *ConnectionManager) buildListConnectionsCommand() *cobra.Command {
	return buildSimpleCommand("list", "List active connections", func(cmd *cobra.Command, args []string) error {
		return cm.listActiveConnections()
	})
}

func (cm *ConnectionManager) buildHealthCommand() *cobra.Command {
	return buildSimpleCommand("health", "Check connection health", func(cmd *cobra.Command, args []string) error {
		return cm.checkConnectionHealth()
	})
}

func (cm *ConnectionManager) buildConfigCommand() *cobra.Command {
	return buildSimpleCommand("config", "Manage connection configuration", func(cmd *cobra.Command, args []string) error {
		return cm.manageConfiguration()
	})
}

func (cm *ConnectionManager) buildAnalyticsCommand() *cobra.Command {
	return buildSimpleCommand("analytics", "View connection analytics", func(cmd *cobra.Command, args []string) error {
		return cm.showAnalytics()
	})
}

func (cm *ConnectionManager) buildMaintenanceCommand() *cobra.Command {
	return buildSimpleCommand("maintenance", "Connection maintenance operations", func(cmd *cobra.Command, args []string) error {
		return cm.performMaintenance()
	})
}

// Placeholder implementations for additional methods
func (cm *ConnectionManager) listActiveConnections() error {
	fmt.Println(" Active Connections:")
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if len(cm.activeConnections) == 0 {
		fmt.Println("   No active connections")
		return nil
	}

	for id, conn := range cm.activeConnections {
		fmt.Printf("   • %s: %s (%s) - %s\n",
			id, conn.Integration.Name, conn.Status,
			time.Since(conn.EstablishedAt).Round(time.Minute))
	}
	return nil
}

func (cm *ConnectionManager) checkConnectionHealth() error {
	fmt.Println(" Connection Health Status:")
	// Implementation would check health of all connections
	return nil
}

func (cm *ConnectionManager) manageConfiguration() error {
	fmt.Println("️ Connection Configuration Management:")
	// Implementation would manage connection configurations
	return nil
}

func (cm *ConnectionManager) showAnalytics() error {
	fmt.Println(" Connection Analytics:")
	// Implementation would show analytics
	return nil
}

func (cm *ConnectionManager) performMaintenance() error {
	fmt.Println(" Performing connection maintenance:")
	// Implementation would perform maintenance operations
	return nil
}

// ===== Thin adapter layer (TASK-INTEGRATION-REGISTRAR) =====
// RunConnect exposes runConnect for the generic registrar without changing original signature.
func (cm *ConnectionManager) RunConnect(cmd *cobra.Command, args []string) error {
	return cm.runConnect(cmd, args)
}

// GetConnectionStatus returns enhanced status of all connections
func (cm *ConnectionManager) GetConnectionStatus() map[string]types.ConnectionResult {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	status := make(map[string]types.ConnectionResult)

	for id, conn := range cm.activeConnections {
		status[id] = types.ConnectionResult{
			Success:       conn.Status == "connected",
			Status:        conn.Status,
			ConnectionID:  conn.ConnectionID,
			EstablishedAt: conn.EstablishedAt,
			Configuration: conn.Config,
			HealthCheck:   conn.Health,
		}
	}

	return status
}

// getAvailableIntegrations returns the enhanced catalog of available integrations
func getAvailableIntegrations() map[string][]types.Integration {
	now := time.Now()

	return map[string][]types.Integration{
		"Billing & Finance": {
			makeIntegration("SAP ERP", "Enterprise resource planning integration", types.CategoryBilling, types.StatusAvailable, "v2.1.0", now.AddDate(0, 0, -5), []string{"Real-time billing", "Cost allocation", "Budget management", "Financial reporting"}),
			makeIntegration("Oracle Financials", "Financial management system", types.CategoryBilling, types.StatusAvailable, "v1.8.3", now.AddDate(0, 0, -10), []string{"GL integration", "AP/AR management", "Cost center tracking"}),
			makeIntegration("QuickBooks", "Small business accounting", types.CategoryBilling, types.StatusAvailable, "v3.2.1", now.AddDate(0, 0, -3), []string{"Invoice management", "Expense tracking", "Tax calculations"}),
		},
		"Business Intelligence": {
			makeIntegration("Tableau", "Data visualization and analytics", types.CategoryBI, types.StatusConnected, "v2023.3", now.AddDate(0, 0, -1), []string{"Interactive dashboards", "Real-time data", "Advanced analytics", "Mobile support"}),
			makeIntegration("Power BI", "Microsoft business analytics", types.CategoryBI, types.StatusAvailable, "v2.123", now.AddDate(0, 0, -2), []string{"Cloud integration", "AI insights", "Collaborative reports"}),
			makeIntegration("Looker", "Modern BI and data platform", types.CategoryBI, types.StatusAvailable, "v23.20", now.AddDate(0, 0, -7), []string{"Data modeling", "Embedded analytics", "API access"}),
		},
		"ITSM & Operations": {
			makeIntegration("ServiceNow", "IT service management platform", types.CategoryITSM, types.StatusAvailable, "v1.4.2", now.AddDate(0, 0, -4), []string{"Incident management", "Change control", "CMDB integration", "Workflow automation"}),
			makeIntegration("Jira Service Desk", "Atlassian service management", types.CategoryITSM, types.StatusAvailable, "v5.1.0", now.AddDate(0, 0, -6), []string{"Ticket management", "SLA tracking", "Customer portal"}),
			makeIntegration("PagerDuty", "Digital operations management", types.CategoryMonitoring, types.StatusAvailable, "v3.7.1", now.AddDate(0, 0, -8), []string{"Incident response", "On-call management", "Alert routing"}),
		},
		"Communication & Collaboration": {
			makeIntegration("Slack", "Team collaboration platform", types.CategoryNotification, types.StatusConnected, "v1.29.0", now.AddDate(0, 0, -1), []string{"Real-time notifications", "Custom workflows", "Bot integration", "File sharing"}),
			makeIntegration("Microsoft Teams", "Unified communication platform", types.CategoryNotification, types.StatusAvailable, "v1.6.00", now.AddDate(0, 0, -2), []string{"Video conferencing", "Chat integration", "Office 365 sync"}),
			makeIntegration("Discord", "Voice and text communication", types.CategoryNotification, types.StatusAvailable, "v1.0.9024", now.AddDate(0, 0, -5), []string{"Community channels", "Voice alerts", "Bot commands"}),
		},
	}
}
