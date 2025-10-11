//go:build devcli

package cmd

import "github.com/spf13/cobra"

// registerDevCLI wires lightweight developer / internal utility commands when built with -tags devcli.
// Intentionally minimal surface: aids rapid local iteration without polluting standard builds.
func registerDevCLI(root *cobra.Command) {
	dev := &cobra.Command{
		Use:   "dev",
		Short: "Developer utilities (build with -tags devcli)",
		Long:  "Ephemeral developer helpers – NOT for production. Hidden in standard builds.",
	}

	dev.AddCommand(&cobra.Command{
		Use:   "echo",
		Short: "Echo arguments (sanity check that devcli tag is active)",
		Run: func(cmd *cobra.Command, args []string) {
			// Keep extremely simple (no logger dependency to avoid side‑effects)
			for _, a := range args {
				cmd.Println(a)
			}
		},
	})

	root.AddCommand(dev)
}
