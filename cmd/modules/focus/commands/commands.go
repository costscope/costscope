package commands

import (
	"github.com/spf13/cobra"
)

// BuildFocusCommand builds the top-level 'focus' command and attaches subcommands like convert and validate.
// Diff has been removed from the public CLI surface; keep convert/validate only.
func BuildFocusCommand() *cobra.Command {
	root := &cobra.Command{
		Use:   "focus",
		Short: "FOCUS operations (convert, validate, jobs)",
		RunE:  func(cmd *cobra.Command, _ []string) error { return cmd.Help() },
	}

	// Existing convert command from this package
	root.AddCommand(BuildConvertCommand())

	// Attach validate command
	root.AddCommand(BuildValidateCommand())

	// Attach jobs command (asynchronous conversion job management in local CLI mode)
	root.AddCommand(NewJobsCommand())

	return root
}
