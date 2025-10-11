//go:build enterprise

package diagnostics

import (
	diagcmd "local/costscope/cmd/modules/diagnostics/commands"
	"local/costscope/internal/core/logging"
	"local/costscope/internal/providers"
)

// CreateDiagnosticsCommands is a small factory used by tests or external wiring
//
//nolint:deadcode // kept for DI convenience; used by external wiring/tests in some builds
func CreateDiagnosticsCommands(logger *logging.Logger, pm *providers.ProviderManager) *diagcmd.DiagnosticsCommands {
	return diagcmd.NewDiagnosticsCommands(logger, pm)
}
