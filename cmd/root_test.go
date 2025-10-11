package cmd

import (
	"testing"
)

func TestRootCommand(t *testing.T) {
	// Test that root command can be executed
	if rootCmd == nil {
		t.Error("rootCmd should not be nil")
	}

	if rootCmd.Use != "costscope" {
		t.Errorf("Expected 'costscope', got '%s'", rootCmd.Use)
	}
}

func TestCommandsRegistered(t *testing.T) {
	// Check that basic commands are registered
	commands := rootCmd.Commands()

	commandNames := make(map[string]bool)
	for _, cmd := range commands {
		commandNames[cmd.Use] = true
	}

	expectedCommands := []string{"config", "version"}
	for _, expected := range expectedCommands {
		if !commandNames[expected] {
			t.Errorf("Expected command '%s' to be registered", expected)
		}
	}
}
