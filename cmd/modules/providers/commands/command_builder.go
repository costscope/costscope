package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"local/costscope/internal/core/logging"
	"local/costscope/internal/providers"
	"local/costscope/internal/providers/types"
	verify "local/costscope/internal/providers/verify"
)

// ProviderCommands manages all provider-related commands
type ProviderCommands struct {
	manager *providers.ProviderManager
	logger  *logging.Logger
}

// NewProviderCommands creates a new instance of ProviderCommands
func NewProviderCommands() *ProviderCommands {
	return &ProviderCommands{
		manager: providers.NewProviderManager(),
		logger:  logging.NewLogger(logging.LevelInfo),
	}
}

// BuildProvidersCommand creates the main providers command with all subcommands
func (pc *ProviderCommands) BuildProvidersCommand() *cobra.Command {
	providersCmd := &cobra.Command{
		Use:   "providers",
		Short: "Manage cloud providers",
		Long: `Manage cloud providers for cost analysis and resource monitoring.

The providers command allows you to configure and manage connections to multiple
cloud providers including AWS, Azure, and Google Cloud Platform.

Features:
  • Connect to cloud providers with credentials
  • List all configured providers and their status
  • Validate provider credentials and connectivity
  • Get cost and resource data from providers
  • Manage provider configurations`,
		Example: `  # List all configured providers
  costscope providers list

  # Check status of all providers
  costscope providers status

  # Validate provider credentials
  costscope providers validate --name aws-prod

  # Get provider information
  costscope providers info --name aws-prod`,
	}

	// Add subcommands
	providersCmd.AddCommand(pc.buildListCommand())
	providersCmd.AddCommand(pc.buildStatusCommand())
	providersCmd.AddCommand(pc.buildValidateCommand())
	providersCmd.AddCommand(pc.buildInfoCommand())
	providersCmd.AddCommand(pc.buildVerifyCommand())

	return providersCmd
}

// buildListCommand creates the list command
func (pc *ProviderCommands) buildListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all configured providers",
		Long:  `List all configured cloud providers and their types`,
		Run: func(cmd *cobra.Command, args []string) {
			pc.listProviders()
		},
	}
}

// buildStatusCommand creates the status command
func (pc *ProviderCommands) buildStatusCommand() *cobra.Command {
	return pc.buildCommandWithNameAndBool(
		"status",
		"Check provider status",
		`Check the connection status and health of cloud providers`,
		"verbose", "v", "Show detailed status information",
		pc.showStatus,
	)
}

// buildValidateCommand creates the validate command
func (pc *ProviderCommands) buildValidateCommand() *cobra.Command {
	return pc.buildCommandWithNameAndBool(
		"validate",
		"Validate provider credentials",
		`Validate cloud provider credentials and connectivity`,
		"all", "a", "Validate all providers",
		pc.validateProviders,
	)
}

// buildCommandWithNameAndBool creates a command that shares the common pattern of
// having a required provider name (optional) and one boolean flag, then delegates
// to the provided run function with the resolved values.
func (pc *ProviderCommands) buildCommandWithNameAndBool(
	use string,
	short string,
	long string,
	boolFlagName string,
	boolFlagShorthand string,
	boolFlagUsage string,
	run func(name string, flag bool),
) *cobra.Command {
	var name string
	var bflag bool

	cmd := &cobra.Command{
		Use:   use,
		Short: short,
		Long:  long,
		Run: func(cmd *cobra.Command, args []string) {
			run(name, bflag)
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "Provider name")
	cmd.Flags().BoolVarP(&bflag, boolFlagName, boolFlagShorthand, false, boolFlagUsage)

	return cmd
}

// buildInfoCommand creates the info command
func (pc *ProviderCommands) buildInfoCommand() *cobra.Command {
	var name string
	var format string

	infoCmd := &cobra.Command{
		Use:   "info",
		Short: "Get provider information",
		Long:  `Get detailed information about a cloud provider`,
		Run: func(cmd *cobra.Command, args []string) {
			pc.showProviderInfo(name, format)
		},
	}

	infoCmd.Flags().StringVarP(&name, "name", "n", "", "Provider name")
	infoCmd.Flags().StringVarP(&format, "format", "f", "table", "Output format (table, json)")
	_ = infoCmd.MarkFlagRequired("name")

	return infoCmd
}

// listProviders lists all configured providers
func (pc *ProviderCommands) listProviders() {
	pc.logger.Info("Listing all configured providers")

	providers := pc.manager.ListProviders()

	if len(providers) == 0 {
		fmt.Println("No providers configured")
		return
	}

	fmt.Println("Configured providers:")
	fmt.Printf("%-20s %-10s\n", "NAME", "TYPE")
	fmt.Printf("%-20s %-10s\n", "----", "----")

	for name, providerType := range providers {
		fmt.Printf("%-20s %-10s\n", name, providerType)
	}

	pc.logger.Info("Listed all providers successfully")
}

// showStatus shows provider status information
func (pc *ProviderCommands) showStatus(name string, verbose bool) {
	pc.logger.Info("Checking provider status")

	if name != "" {
		// Show status for specific provider
		status, err := pc.manager.GetProviderStatus(name)
		if err != nil {
			fmt.Printf("Error getting status for provider %s: %v\n", name, err)
			pc.logger.Error(fmt.Sprintf("Failed to get status for provider %s: %v", name, err))
			return
		}

		pc.printProviderStatus(status, verbose)
	} else {
		// Show status for all providers
		statuses := pc.manager.GetAllStatuses()

		if len(statuses) == 0 {
			fmt.Println("No providers configured")
			return
		}

		fmt.Println("Provider Status:")
		if !verbose {
			fmt.Printf("%-20s %-10s %-12s %-15s\n", "NAME", "TYPE", "STATUS", "HEALTH")
			fmt.Printf("%-20s %-10s %-12s %-15s\n", "----", "----", "------", "------")
		}

		for _, status := range statuses {
			pc.printProviderStatus(status, verbose)
		}
	}

	pc.logger.Info("Status check completed")
}

// printProviderStatus prints status information for a single provider
func (pc *ProviderCommands) printProviderStatus(status *types.ProviderStatus, verbose bool) {
	connectionStatus := "disconnected"
	if status.IsConnected {
		connectionStatus = "connected"
	}

	if verbose {
		fmt.Printf("\nProvider: %s\n", status.Name)
		fmt.Printf("  Type: %s\n", status.Type)
		fmt.Printf("  Status: %s\n", connectionStatus)
		fmt.Printf("  Health: %s\n", status.HealthStatus)
		if status.ErrorMessage != "" {
			fmt.Printf("  Error: %s\n", status.ErrorMessage)
		}
		if !status.LastSyncTime.IsZero() {
			fmt.Printf("  Last Sync: %s\n", status.LastSyncTime.Format("2006-01-02 15:04:05"))
		}
		fmt.Printf("  Metrics: %d\n", status.MetricsCount)
		fmt.Printf("  Resources: %d\n", status.ResourcesCount)
	} else {
		fmt.Printf("%-20s %-10s %-12s %-15s\n",
			status.Name,
			status.Type,
			connectionStatus,
			status.HealthStatus)
	}
}

// validateProviders validates provider credentials
func (pc *ProviderCommands) validateProviders(name string, all bool) {
	pc.logger.Info("Validating provider credentials")
	ctx := context.Background()

	if name != "" {
		// Validate specific provider
		err := pc.manager.ValidateProvider(ctx, name)
		if err != nil {
			fmt.Printf("Validation failed for provider %s: %v\n", name, err)
			pc.logger.Error(fmt.Sprintf("Validation failed for provider %s: %v", name, err))
			return
		}

		fmt.Printf("Provider %s validated successfully\n", name)
	} else if all {
		// Validate all providers
		results := pc.manager.ValidateAllProviders(ctx)

		fmt.Println("Validation Results:")
		for providerName, err := range results {
			if err != nil {
				fmt.Printf("  %s: FAILED - %v\n", providerName, err)
			} else {
				fmt.Printf("  %s: PASSED\n", providerName)
			}
		}
	} else {
		fmt.Println("Error: Must specify either --name or --all flag")
		return
	}

	pc.logger.Info("Validation completed")
}

// showProviderInfo shows detailed information about a provider
func (pc *ProviderCommands) showProviderInfo(name, format string) {
	pc.logger.Info(fmt.Sprintf("Getting information for provider: %s", name))

	provider, err := pc.manager.GetProvider(name)
	if err != nil {
		fmt.Printf("Error getting provider %s: %v\n", name, err)
		pc.logger.Error(fmt.Sprintf("Failed to get provider %s: %v", name, err))
		return
	}

	ctx := context.Background()
	info, err := provider.GetProviderInfo(ctx)
	if err != nil {
		fmt.Printf("Error getting provider info: %v\n", err)
		pc.logger.Error(fmt.Sprintf("Failed to get provider info: %v", err))
		return
	}

	switch format {
	case "json":
		output, err := json.MarshalIndent(info, "", "  ")
		if err != nil {
			fmt.Printf("Error formatting output: %v\n", err)
			return
		}
		fmt.Println(string(output))

	case "table":
		fallthrough
	default:
		fmt.Printf("Provider Information: %s\n", info.Name)
		fmt.Printf("  Type: %s\n", info.Type)
		fmt.Printf("  Version: %s\n", info.Version)
		fmt.Printf("  Supported Regions: %v\n", info.SupportedRegions)
		fmt.Printf("  Capabilities: %v\n", info.Capabilities)
		if len(info.Metadata) > 0 {
			fmt.Printf("  Metadata:\n")
			for key, value := range info.Metadata {
				fmt.Printf("    %s: %s\n", key, value)
			}
		}
	}

	pc.logger.Info("Provider information retrieved successfully")
}

// buildVerifyCommand creates the 'verify' command (skeleton implementation)
// Usage: costscope providers verify <file> --provider aws [flags]
func (pc *ProviderCommands) buildVerifyCommand() *cobra.Command {
	var provider string
	var limit int
	var useUnified bool
	var invariants bool
	var invariantsBaseline string
	var invariantsTolerance float64
	var invariantsReport string
	var failOnInvariants bool
	var format string
	var stopAfter string
	var errorThreshold int

	cmd := &cobra.Command{
		Use:   "verify <file>",
		Short: "Parse, map, (optionally) compute invariants, and validate a raw billing file",
		Long: `Perform an offline verification pipeline on a raw provider billing export:
Stages: parse -> map -> (invariants) -> validate.
No Parquet is written; this is a pre-flight correctness & drift check before full conversion.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			file := args[0]
			opts := verify.Options{
				Provider:         provider,
				File:             file,
				Limit:            limit,
				StopAfter:        stopAfter,
				UseUnified:       useUnified,
				EnableInvariants: invariants,
				BaselinePath:     invariantsBaseline,
				Tolerance:        invariantsTolerance,
				ErrorThreshold:   errorThreshold,
				Format:           format,
			}
			if err := opts.Validate(); err != nil {
				pc.logger.Error(fmt.Sprintf("providers.verify.validate_options error: %v", err))
				return err
			}
			pc.logger.Info(fmt.Sprintf("providers.verify.start provider=%s file=%s", provider, file))
			sum, err := verify.Process(opts)
			if err != nil {
				pc.logger.Error(fmt.Sprintf("providers.verify.internal_error: %v", err))
				return err
			}
			// Optional invariants report file
			if invariantsReport != "" && invariants {
				_ = os.WriteFile(invariantsReport, []byte(sum.JSON()), 0o600)
			}
			// Output formatting
			switch format {
			case "json":
				fmt.Println(sum.JSON())
			case "yaml":
				// Simple YAML emulation (minimal) – future: proper marshal
				fmt.Println("summary:")
				fmt.Printf("  provider: %s\n", sum.Provider)
				fmt.Printf("  file: %s\n", sum.File)
				fmt.Printf("  exit_code: %d\n", sum.ExitCode)
				fmt.Printf("  overall_status: %s\n", sum.OverallStatus)
				fmt.Println("  stages:")
				for k, st := range sum.Stages {
					fmt.Printf("    %s: {status: %s, processed: %d}\n", k, st.Status, st.ProcessedRows)
				}
			default:
				printVerifyTable(sum)
			}
			pc.logger.Info(fmt.Sprintf("providers.verify.done exit_code=%d status=%s", sum.ExitCode, sum.OverallStatus))
			// Non-zero exit codes represent pipeline-level conditions; Cobra will propagate error if we return one.
			if sum.ExitCode != 0 {
				return fmt.Errorf("verify pipeline exited with code %d (%s)", sum.ExitCode, sum.OverallStatus)
			}
			return nil
		},
		Example: `  # Basic verify
  costscope providers verify cur.csv.gz --provider aws

  # With invariants and baseline drift detection
  costscope providers verify cur.csv.gz --provider aws --invariants --invariants-baseline baseline.json --fail-on-invariants

  # Unified mapper + stop after map stage
  costscope providers verify export.csv --provider azure --use-unified-mapper --stop-after map`,
	}

	cmd.Flags().StringVar(&provider, "provider", "", "Provider key (aws, azure, gcp, ...) or 'auto'")
	_ = cmd.MarkFlagRequired("provider")
	cmd.Flags().IntVar(&limit, "limit", 10000, "Maximum mapped records to process (0 = no limit)")
	cmd.Flags().BoolVar(&useUnified, "use-unified-mapper", false, "Force unified experimental mapper path")
	cmd.Flags().BoolVar(&invariants, "invariants", false, "Compute lightweight streaming invariants")
	cmd.Flags().StringVar(&invariantsBaseline, "invariants-baseline", "", "Baseline invariants JSON for drift comparison")
	cmd.Flags().Float64Var(&invariantsTolerance, "invariants-tolerance", 0.01, "Relative tolerance for drift detection")
	cmd.Flags().StringVar(&invariantsReport, "invariants-report", "", "Write current invariants JSON report")
	cmd.Flags().BoolVar(&failOnInvariants, "fail-on-invariants", false, "Exit non-zero on invariants drift violations")
	cmd.Flags().StringVar(&format, "format", "table", "Output format: table|json|yaml")
	cmd.Flags().StringVar(&stopAfter, "stop-after", "", "Stop after stage: parse|map|invariants|validate")
	cmd.Flags().IntVar(&errorThreshold, "error-threshold", 100, "Maximum parse/map errors before abort (0 = fail fast)")

	return cmd
}

// printVerifyTable prints a compact table summary of the verify stages.
func printVerifyTable(sum *verify.Summary) {
	fmt.Printf("Verify Summary (provider=%s file=%s exit=%d status=%s)\n", sum.Provider, sum.File, sum.ExitCode, sum.OverallStatus)
	fmt.Printf("%-12s %-10s %-10s %-10s %-8s\n", "STAGE", "STATUS", "ROWS", "SAMPLED", "MS")
	fmt.Printf("%-12s %-10s %-10s %-10s %-8s\n", "-----", "------", "----", "-------", "--")
	order := []verify.Stage{verify.StageParse, verify.StageMap, verify.StageInvariants, verify.StageValidate}
	for _, stg := range order {
		res, ok := sum.Stages[stg]
		if !ok {
			continue
		}
		fmt.Printf("%-12s %-10s %-10d %-10d %-8d\n", stg, res.Status, res.ProcessedRows, res.SampledRows, res.DurationMs)
	}
}
