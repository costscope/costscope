//go:build !devcli

package cmd

import "github.com/spf13/cobra"

// registerDevCLI is a no-op unless built with -tags devcli.
func registerDevCLI(root *cobra.Command) {}
