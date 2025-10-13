package cmd

import (
	"github.com/costscope/costscope/cmd/modules/streaming/commands"
	"github.com/costscope/costscope/internal/core/logging"

	"github.com/spf13/cobra"
)

// StreamingManager orchestrates streaming management operations
type StreamingManager struct {
	commands *commands.StreamingCommands
	logger   *logging.Logger
}

// NewStreamingManager creates a new instance of StreamingManager
func NewStreamingManager() *StreamingManager {
	return &StreamingManager{
		commands: commands.NewStreamingCommands(),
		logger:   logging.NewLogger(logging.LevelInfo),
	}
}

// BuildStreamingCommand creates the main streaming command
func (sm *StreamingManager) BuildStreamingCommand() *cobra.Command {
	sm.logger.Info("Building streaming command")
	return sm.commands.BuildMainCommand()
}
