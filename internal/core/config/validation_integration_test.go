package config

import (
	"path/filepath"
	"testing"
	"time"
)

// TestStartupValidationAndDirectories simulates an application startup sequence
// invoking ValidateAllConfig followed by EnsureConfigDirectories to ensure
// they compose without side-effects or race conditions.
func TestStartupValidationAndDirectories(t *testing.T) {
	base := t.TempDir()
	cfg := minimalValidConfig(base)

	// Add directories that should be created during EnsureConfigDirectories
	cfg.Core.DataDirectory = filepath.Join(base, "coredata")
	cfg.Core.TempDirectory = filepath.Join(base, "tmp")
	cfg.Streaming.CheckpointEnabled = true
	cfg.Streaming.CheckpointDir = filepath.Join(base, "checkpoint")
	cfg.Database.MigrationsPath = filepath.Join(base, "migrations")

	// Simulate a small variation: enable AWS provider with profile credentials
	cfg.Providers.AWS.Enabled = true
	cfg.Providers.AWS.Region = "us-east-1"
	cfg.Providers.AWS.Profile = "default"
	cfg.Providers.AWS.RequestTimeout = 5 * time.Second

	if err := ValidateAllConfig(cfg); err != nil {
		// Fail fast like a real startup would
		if ce, ok := err.(*ConfigError); ok {
			// Ensure we got structured context
			t.Fatalf("validation failed section=%s key=%s msg=%s underlying=%v", ce.Section, ce.Key, ce.Message, ce.Err)
		}
		panic(err)
	}
	if err := EnsureConfigDirectories(cfg); err != nil {
		if ce, ok := err.(*ConfigError); ok {
			t.Fatalf("ensure directories failed section=%s key=%s msg=%s underlying=%v", ce.Section, ce.Key, ce.Message, ce.Err)
		}
		panic(err)
	}
}
