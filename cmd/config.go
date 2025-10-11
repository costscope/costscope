package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"local/costscope/internal/core/config"
	"local/costscope/internal/core/logging"

	"github.com/spf13/cobra"
)

var (
	configVerbose bool
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configuration management commands",
	Long:  `Manage CostScope configuration settings and validation`,
}

var configValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate configuration",
	Long:  `Validate CostScope configuration files and settings`,
	Run: func(cmd *cobra.Command, args []string) {
		logger := logging.NewLogger(logging.LevelInfo)
		logger.Info("Validating configuration...")

		configManager := config.NewConfigManager()

		if err := configManager.Validate(); err != nil {
			logger.Error(fmt.Sprintf("Configuration validation failed: %v", err))
			return
		}

		fmt.Println("Configuration validation passed")
		logger.Info("Configuration validation completed successfully")
	},
}

var configVersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show configuration version",
	Long:  `Display the current configuration version`,
	Run: func(cmd *cobra.Command, args []string) {
		logger := logging.NewLogger(logging.LevelInfo)
		logger.Info("Getting configuration version...")

		configManager := config.NewConfigManager()
		version := configManager.GetVersion()

		fmt.Printf("Configuration version: %s\n", version)
		logger.Info("Configuration version retrieved successfully")
	},
}

func initConfigCommands() {
	configCmd.AddCommand(configValidateCmd)
	configCmd.AddCommand(configVersionCmd)

	// Add flags
	configCmd.PersistentFlags().BoolVarP(&configVerbose, "verbose", "v", false, "verbose output")

	// Cleanup legacy: physically remove cmd/modules/config_advanced
	var repoRoot string
	cleanupCmd := &cobra.Command{
		Use:   "cleanup-legacy",
		Short: "Physically remove legacy modules (one-off maintenance)",
		Long:  "Removes the obsolete cmd/modules/config_advanced directory from a local clone. Provide --repo-root to point at the repository root.",
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := logging.NewLogger(logging.LevelInfo)

			if repoRoot == "" {
				repoRoot = "."
			}
			absRoot, err := filepath.Abs(repoRoot)
			if err != nil {
				logger.Error(fmt.Sprintf("failed to resolve repo root: %v", err))
				return err
			}
			target := filepath.Join(absRoot, "cmd", "modules", "config_advanced")
			info, statErr := os.Stat(target)
			if statErr != nil {
				if os.IsNotExist(statErr) {
					logger.Info(fmt.Sprintf("legacy path already absent: %s", target))
					return nil
				}
				logger.Error(fmt.Sprintf("failed to stat legacy path: %v", statErr))
				return statErr
			}
			if !info.IsDir() {
				logger.Error(fmt.Sprintf("not a directory: %s", target))
				return fmt.Errorf("not a directory: %s", target)
			}
			logger.Info(fmt.Sprintf("removing legacy directory: %s", target))
			if err := os.RemoveAll(target); err != nil {
				logger.Error(fmt.Sprintf("failed to remove legacy directory: %v", err))
				return err
			}
			logger.Info("legacy directory removed successfully")
			return nil
		},
	}
	// Hide from public help to avoid shipping developer-only utility in UX
	cleanupCmd.Hidden = true
	cleanupCmd.Flags().StringVar(&repoRoot, "repo-root", ".", "path to repository root (default: current dir)")
	configCmd.AddCommand(cleanupCmd)
}
