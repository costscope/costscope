package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"local/costscope/internal/core/logging"
	"local/costscope/internal/core/production"
	"local/costscope/internal/providers"
)

// DiagnosticsCommands provides system diagnostics CLI
type DiagnosticsCommands struct {
	logger          *logging.Logger
	providerManager *providers.ProviderManager
	productionSvc   production.ProductionService
}

// NewDiagnosticsCommands constructs diagnostics CLI with default production service (factory/DI)
func NewDiagnosticsCommands(logger *logging.Logger, providerManager *providers.ProviderManager) *DiagnosticsCommands {
	svc := production.NewBasicProductionService(providerManager, logger)
	return &DiagnosticsCommands{logger: logger, providerManager: providerManager, productionSvc: svc}
}

// NewDiagnosticsCommandsWithService allows injecting a custom production service (for tests)
//
//nolint:deadcode // used by tests and alternative wiring in enterprise builds
func NewDiagnosticsCommandsWithService(logger *logging.Logger, providerManager *providers.ProviderManager, svc production.ProductionService) *DiagnosticsCommands {
	return &DiagnosticsCommands{logger: logger, providerManager: providerManager, productionSvc: svc}
}

// self-reference to satisfy static deadcode analysis in default builds while retaining
// the constructor for tests and enterprise wiring. No runtime effect.
var _ = NewDiagnosticsCommandsWithService

// BuildDiagnosticsCommand creates the root diagnostics command
func (dc *DiagnosticsCommands) BuildDiagnosticsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "diagnostics",
		Short: "System diagnostics and status",
		Long: `Run quick system diagnostics and view readiness indicators.

This command queries core production services to gather health, performance,
security, and integration metrics, producing a concise status report.`,
		Example: `  # Show concise status in a human-friendly table
  costscope diagnostics status

  # Output JSON for automation
  costscope diagnostics status --output json

  # Require a minimum readiness score (0-100)
  costscope diagnostics status --min-score 80`,
	}

	// Subcommands
	cmd.AddCommand(dc.buildStatusCommand())
	return cmd
}

// buildStatusCommand creates the diagnostics status command
func (dc *DiagnosticsCommands) buildStatusCommand() *cobra.Command {
	var (
		output   string
		minScore int
		detailed bool
	)

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show system readiness status",
		Long:  "Collects system metrics from core services and prints readiness status.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			// Validate flags deterministically
			if err := validateOutputFlag(output, []string{"table", "json", "yaml"}); err != nil {
				return err
			}
			if minScore < 0 || minScore > 100 {
				return friendlyFlagError("min-score", "must be between 0 and 100")
			}

			ctx := context.Background()
			dc.logger.Info("Collecting diagnostics status")
			status, err := dc.productionSvc.GetSystemStatus(ctx)
			if err != nil {
				return fmt.Errorf("failed to collect system status: %w", err)
			}

			// Enforce threshold
			if minScore > 0 && status.ReadinessScore < minScore {
				return fmt.Errorf("readiness score %d is below required minimum %d", status.ReadinessScore, minScore)
			}

			switch output {
			case "json":
				payload := buildStatusPayload(status, detailed)
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(payload)
			case "yaml":
				// Avoid extra dependency; simple YAML-like output for clarity
				printYAMLPayload(cmd, buildStatusPayload(status, detailed))
				return nil
			default: // table
				printTablePayload(cmd, status, detailed)
				return nil
			}
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "table", "Output format (table, json, yaml)")
	cmd.Flags().IntVar(&minScore, "min-score", 0, "Minimum acceptable readiness score (0-100)")
	cmd.Flags().BoolVarP(&detailed, "detailed", "d", false, "Show detailed component metrics")

	return cmd
}

// --- helpers ---

func validateOutputFlag(v string, allowed []string) error {
	for _, a := range allowed {
		if v == a {
			return nil
		}
	}
	return friendlyFlagError("output", fmt.Sprintf("must be one of: %s", strings.Join(allowed, ", ")))
}

func friendlyFlagError(name, msg string) error {
	return fmt.Errorf("invalid --%s: %s", name, msg)
}

type statusPayload struct {
	Timestamp       time.Time                         `json:"timestamp"`
	ReadinessScore  int                               `json:"readiness_score"`
	ProductionReady bool                              `json:"production_ready"`
	Health          production.SystemHealthStatus     `json:"health"`
	Performance     production.PerformanceMetrics     `json:"performance"`
	Security        production.SecurityMetrics        `json:"security"`
	Integration     production.IntegrationMetrics     `json:"integration"`
	Details         map[string]map[string]interface{} `json:"details,omitempty"`
}

func buildStatusPayload(status *production.ProductionSystemMetrics, detailed bool) *statusPayload {
	p := &statusPayload{
		Timestamp:       status.Timestamp,
		ReadinessScore:  status.ReadinessScore,
		ProductionReady: status.ProductionReady,
		Health:          status.SystemHealth,
		Performance:     status.Performance,
		Security:        status.Security,
		Integration:     status.Integration,
	}
	if detailed {
		p.Details = map[string]map[string]interface{}{
			"health": {
				"component_health": status.SystemHealth.ComponentHealth,
			},
			"security": {
				"compliance_status": status.Security.ComplianceStatus,
			},
			"integration": {
				"integration_health": status.Integration.IntegrationHealth,
			},
		}
	}
	return p
}

func printTablePayload(cmd *cobra.Command, status *production.ProductionSystemMetrics, detailed bool) {
	out := cmd.OutOrStdout()
	_, _ = fmt.Fprintf(out, "Diagnostics Status\n")
	_, _ = fmt.Fprintf(out, "===================\n")
	_, _ = fmt.Fprintf(out, "Readiness Score: %d\n", status.ReadinessScore)
	_, _ = fmt.Fprintf(out, "Production Ready: %t\n", status.ProductionReady)
	_, _ = fmt.Fprintf(out, "Health: %s (score: %d)\n", status.SystemHealth.Status, status.SystemHealth.HealthScore)
	_, _ = fmt.Fprintf(out, "Performance: grade %s (score: %d)\n", status.Performance.PerformanceGrade, status.Performance.OptimizationScore)
	_, _ = fmt.Fprintf(out, "Security: grade %s (score: %d)\n", status.Security.SecurityGrade, status.Security.SecurityScore)
	_, _ = fmt.Fprintf(out, "Integration: %s (score: %d)\n", status.Integration.OperationalMaturity, status.Integration.IntegrationScore)
	if detailed {
		_, _ = fmt.Fprintln(out, "\nComponents:")
		// Print limited deterministic subset
		keys := []string{"providers", "analytics", "reports"}
		for _, k := range keys {
			if v, ok := status.SystemHealth.ComponentHealth[k]; ok {
				_, _ = fmt.Fprintf(out, "  - %s: %s\n", k, v)
			}
		}
	}
}

func printYAMLPayload(cmd *cobra.Command, p *statusPayload) {
	out := cmd.OutOrStdout()
	// Minimal, deterministic YAML-like output
	_, _ = fmt.Fprintf(out, "readiness_score: %d\n", p.ReadinessScore)
	_, _ = fmt.Fprintf(out, "production_ready: %t\n", p.ProductionReady)
	_, _ = fmt.Fprintf(out, "health_status: %s\n", p.Health.Status)
	_, _ = fmt.Fprintf(out, "performance_grade: %s\n", p.Performance.PerformanceGrade)
	_, _ = fmt.Fprintf(out, "security_grade: %s\n", p.Security.SecurityGrade)
}
