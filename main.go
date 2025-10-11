package main

import (
	"local/costscope/cmd"
	"local/costscope/internal/core/logging"
)

// main is the entry point for the CostScope application
func main() {
	if err := cmd.Execute(); err != nil {
		// Route fatal startup errors through unified logger
		logger := logging.GetLogger()
		logger.FatalWithFields("cli_execute_failed", map[string]interface{}{"error": err.Error()})
	}
}
