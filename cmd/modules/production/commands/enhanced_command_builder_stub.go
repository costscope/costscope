//go:build !enterprise

package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

// BuildEnhancedProductionCommands is unavailable without the enterprise tag; returns a placeholder command.
// Intentional stub (enterprise gating): command placeholder in community build.
func BuildEnhancedProductionCommands() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "production-enhanced",
		Short: "Enhanced production commands (enterprise-only)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("enterprise-only: build with -tags enterprise to enable enhanced production commands")
		},
	}
	return cmd
}
