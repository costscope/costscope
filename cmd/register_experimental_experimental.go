//go:build experimental

package cmd

import (
	advanced "local/costscope/cmd/modules/analytics_advanced/commands"
	complex "local/costscope/cmd/modules/analytics_complex/commands"

	"github.com/spf13/cobra"
)

// registerExperimental wires experimental command groups when built with -tags experimental.
func registerExperimental(root *cobra.Command) {
	// analytics-advanced
	aac := advanced.NewAdvancedAnalyticsCommands()
	root.AddCommand(aac.BuildAdvancedAnalyticsCommand())

	// analytics-complex (type-safe ML analytics)
	acc := complex.NewAnalyticsComplexCommands()
	root.AddCommand(acc.BuildAnalyticsComplexCommand())
}
