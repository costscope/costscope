//go:build experimental

package analytics_complex

import (
	"local/costscope/cmd/modules/analytics_complex/commands"

	"github.com/spf13/cobra"
)

// RegisterCommands adds analytics-complex commands to the root command
func RegisterCommands(rootCmd *cobra.Command) {
	analyticsComplexCommands := commands.NewAnalyticsComplexCommands()
	rootCmd.AddCommand(analyticsComplexCommands.BuildAnalyticsComplexCommand())
}
