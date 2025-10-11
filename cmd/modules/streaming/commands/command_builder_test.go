package commands

import (
	"testing"

	"local/costscope/cmd/modules/streaming/types"
)

func TestNewStreamingCommands(t *testing.T) {
	commands := NewStreamingCommands()

	if commands == nil {
		t.Fatal("NewStreamingCommands returned nil")
	}

	if commands.flags == nil {
		t.Error("StreamingCommands.flags is nil")
	}

	if commands.logger == nil {
		t.Error("StreamingCommands.logger is nil")
	}
}

func TestBuildMainCommand(t *testing.T) {
	commands := NewStreamingCommands()
	cmd := commands.BuildMainCommand()

	if cmd == nil {
		t.Fatal("BuildMainCommand returned nil")
	}

	if cmd.Use != "streaming" {
		t.Errorf("Expected command use to be 'streaming', got '%s'", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Command short description is empty")
	}

	// Check if subcommands are added
	subcommands := cmd.Commands()
	expectedSubcommands := []string{"history", "list", "pause", "resume", "start", "status", "stop"}

	if len(subcommands) != len(expectedSubcommands) {
		t.Errorf("Expected %d subcommands, got %d", len(expectedSubcommands), len(subcommands))
	}

	// Commands are sorted alphabetically by cobra
	for i, subcmd := range subcommands {
		if i < len(expectedSubcommands) && subcmd.Use != expectedSubcommands[i] {
			t.Errorf("Expected subcommand %d to be '%s', got '%s'", i, expectedSubcommands[i], subcmd.Use)
		}
	}
}

func TestStreamingFlagsIntegration(t *testing.T) {
	flags := &types.StreamingFlags{}

	flags.Provider = "aws"
	flags.Workers = 8
	flags.Verbose = true

	if flags.Provider != "aws" {
		t.Errorf("Expected provider to be 'aws', got '%s'", flags.Provider)
	}

	if flags.Workers != 8 {
		t.Errorf("Expected workers to be 8, got %d", flags.Workers)
	}

	if !flags.Verbose {
		t.Error("Expected verbose to be true")
	}
}
