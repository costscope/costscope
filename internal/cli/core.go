// Package cli provides enhanced CLI framework with auto-discovery and dependency injection.
package cli

import (
	"context"
	"fmt"
	"reflect"
	"sort"

	"local/costscope/internal/framework"

	"github.com/spf13/cobra"
)

// CommandProvider defines the interface for command providers
type CommandProvider interface {
	GetCommands() []*cobra.Command
	Priority() int
}

// EnhancedCLI represents the enhanced CLI framework
type EnhancedCLI struct {
	rootCmd   *cobra.Command
	framework *framework.Framework
	providers map[string]CommandProvider
	commands  map[string]*cobra.Command
}

// NewEnhancedCLI creates a new enhanced CLI instance
func NewEnhancedCLI(fw *framework.Framework) *EnhancedCLI {
	rootCmd := &cobra.Command{
		Use:   "costscope",
		Short: "CostScope - Enterprise Cloud Cost Analysis Tool",
		Long: `CostScope is a comprehensive enterprise-grade tool for analyzing and optimizing cloud costs.
It supports multiple cloud providers, FOCUS standard compliance, and provides advanced analytics capabilities.`,
		Version: "1.0.0",
	}

	cli := &EnhancedCLI{
		rootCmd:   rootCmd,
		framework: fw,
		providers: make(map[string]CommandProvider),
		commands:  make(map[string]*cobra.Command),
	}

	// Setup global flags
	cli.setupGlobalFlags()

	// Register built-in command providers
	cli.registerBuiltinProviders()

	return cli
}

// setupGlobalFlags sets up global CLI flags
func (cli *EnhancedCLI) setupGlobalFlags() {
	flags := cli.rootCmd.PersistentFlags()

	flags.Bool("debug", false, "Enable debug mode")
	flags.String("config", "", "Config file path")
	flags.String("log-level", "info", "Log level (debug, info, warn, error)")
	flags.Bool("quiet", false, "Suppress non-essential output")
}

// registerBuiltinProviders registers built-in command providers
func (cli *EnhancedCLI) registerBuiltinProviders() {
	// Register core command providers
	cli.RegisterProvider("analyze", NewAnalyzeCommandProvider())
	cli.RegisterProvider("report", NewReportCommandProvider())
	cli.RegisterProvider("config", NewConfigCommandProvider())
	cli.RegisterProvider("plugin", NewPluginCommandProvider())
	cli.RegisterProvider("server", NewServerCommandProvider())
}

// RegisterProvider registers a command provider
func (cli *EnhancedCLI) RegisterProvider(name string, provider CommandProvider) {
	cli.providers[name] = provider
}

// DiscoverAndRegisterCommands discovers and registers all commands from providers
func (cli *EnhancedCLI) DiscoverAndRegisterCommands() error {
	// Sort providers by priority
	var providers []CommandProvider
	for _, provider := range cli.providers {
		providers = append(providers, provider)
	}

	sort.Slice(providers, func(i, j int) bool {
		return providers[i].Priority() < providers[j].Priority()
	})

	// Register commands from each provider
	for _, provider := range providers {
		commands := provider.GetCommands()
		for _, cmd := range commands {
			// Inject dependencies into command
			if err := cli.injectDependencies(cmd); err != nil {
				return fmt.Errorf("failed to inject dependencies for command '%s': %w", cmd.Use, err)
			}

			// Register command
			cli.rootCmd.AddCommand(cmd)
			cli.commands[cmd.Use] = cmd
		}
	}

	return nil
}

// injectDependencies injects framework services into command handlers
// nolint:unparam // kept to preserve method signature compatibility
func (cli *EnhancedCLI) injectDependencies(cmd *cobra.Command) error {
	// Get the command's RunE function
	if cmd.RunE == nil {
		return nil
	}

	// Use reflection to analyze the function signature
	runFunc := reflect.ValueOf(cmd.RunE)
	runType := runFunc.Type()

	// Check if it's a function
	if runType.Kind() != reflect.Func {
		return nil
	}

	// Create a wrapper that injects dependencies
	originalRun := cmd.RunE
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		// Create a context with framework access and attach it to the command
		ctx := context.WithValue(context.Background(), contextKeyFramework, cli.framework)
		// Container no longer injected via dedicated helper; access through framework if needed.
		cmd.SetContext(ctx)

		// Call the original function with the enhanced context
		return originalRun(cmd, args)
	}

	return nil
}

// Execute executes the CLI
// GetRootCommand returns the root command
func (cli *EnhancedCLI) GetRootCommand() *cobra.Command {
	return cli.rootCmd
}

// Helper functions for getting framework services from context
// context keys for injection
type contextKey int

const (
	contextKeyFramework contextKey = iota
)

func GetFramework(cmd *cobra.Command) *framework.Framework {
	ctx := cmd.Context()
	if ctx == nil {
		return nil
	}

	fw, ok := ctx.Value(contextKeyFramework).(*framework.Framework)
	if !ok {
		return nil
	}

	return fw
}

// Utility functions for CLI development
func IsDebugMode(cmd *cobra.Command) bool {
	debug, _ := cmd.Flags().GetBool("debug")
	return debug
}

func IsJSONOutput(cmd *cobra.Command) bool {
	json, _ := cmd.Flags().GetBool("json")
	return json
}

func IsQuietMode(cmd *cobra.Command) bool {
	quiet, _ := cmd.Flags().GetBool("quiet")
	return quiet
}

func GetLogLevel(cmd *cobra.Command) string {
	level, _ := cmd.Flags().GetString("log-level")
	if level == "" {
		level = "info"
	}
	return level
}

// (Removed GetConfigPath helper – direct flag access kept where needed.)

// PrintInfo prints informational message unless quiet mode is enabled
func PrintInfo(cmd *cobra.Command, format string, args ...interface{}) {
	if !IsQuietMode(cmd) {
		fmt.Printf(format+"\n", args...)
	}
}

// CommandContext provides context for command execution
type CommandContext struct {
	Framework *framework.Framework
	Container *framework.Container
	EventBus  *framework.EventBus
	Debug     bool
	JSONMode  bool
	QuietMode bool
	LogLevel  string
}

// NewCommandContext creates a new command context from cobra command
func NewCommandContext(cmd *cobra.Command) *CommandContext {
	fw := GetFramework(cmd)

	ctx := &CommandContext{
		Framework: fw,
		Debug:     IsDebugMode(cmd),
		JSONMode:  IsJSONOutput(cmd),
		QuietMode: IsQuietMode(cmd),
		LogLevel:  GetLogLevel(cmd),
	}

	if fw != nil {
		ctx.Container = fw.Container()
		ctx.EventBus = fw.EventBus()
	}

	return ctx
}
