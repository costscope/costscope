package validation

import (
	"github.com/spf13/cobra"

	parent "local/costscope/cmd/modules/focus/commands"
)

// BuildValidateCommand returns the canonical validate command from the parent commands package.
// This thin wrapper preserves import/backward compatibility while delegating to the single source of truth.
func BuildValidateCommand() *cobra.Command {
	return parent.BuildValidateCommand()
}

// keep a local reference so static deadcode tools recognize usage in default builds
// without altering runtime behavior. This preserves the import path shim while cleanly
// passing deadcode checks for the CLI modules scope.
var _ = BuildValidateCommand
