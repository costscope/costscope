//go:build !experimental

package cmd

import "github.com/spf13/cobra"

// registerExperimental is a no-op in default builds; experimental commands are hidden unless -tags experimental is used.
func registerExperimental(root *cobra.Command) {}
