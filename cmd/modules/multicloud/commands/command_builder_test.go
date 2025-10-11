package commands

import (
	"bytes"
	"testing"

	"local/costscope/internal/providers"
)

// TestMulticloudCommands_BuildMulticloudCommand tests the main command building
func TestMulticloudCommands_BuildMulticloudCommand(t *testing.T) {
	// Create mock provider manager
	providerManager := &providers.ProviderManager{}

	// Create command builder
	commands := NewMulticloudCommands(providerManager)

	// Build command and attach enhanced subcommands
	cmd := commands.BuildMulticloudCommand()
	commands.AttachEnhancedSubcommands(cmd)

	// Basic verification
	if cmd.Use != "multicloud" {
		t.Errorf("Expected command use to be 'multicloud', got '%s'", cmd.Use)
	}

	if cmd.Short != "Multi-cloud cost optimization and management" {
		t.Errorf("Expected short description mismatch")
	}

	// Verify subcommands exist
	subcommands := cmd.Commands()
	expectedCount := 13 // base 5 + 4 enhanced + 4 newly added (recommendations, inventory, migration-plan, feasibility)

	if len(subcommands) != expectedCount {
		t.Errorf("Expected %d subcommands, got %d", expectedCount, len(subcommands))
	}

	// Ensure new advanced surfacing commands exist
	found := map[string]bool{}
	for _, c := range subcommands {
		found[c.Use] = true
	}
	for _, name := range []string{"recommendations", "inventory", "migration-plan", "feasibility"} {
		if !found[name] {
			t.Errorf("expected subcommand %s present", name)
		}
	}

	// Verify global flags exist
	if cmd.PersistentFlags().Lookup("providers") == nil {
		t.Error("Expected 'providers' flag to exist")
	}

	if cmd.PersistentFlags().Lookup("start-date") == nil {
		t.Error("Expected 'start-date' flag to exist")
	}

	if cmd.PersistentFlags().Lookup("format") == nil {
		t.Error("Expected 'format' flag to exist")
	}

	// Help snapshot (lightweight): ensure --help prints something and includes a key phrase
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--help"})
	_ = cmd.Execute()
	out := buf.String()
	if out == "" || !contains(out, "Multi-cloud commands") {
		t.Errorf("expected help output to contain module description, got: %q", out)
	}
}

// TestMulticloudCommands_ParseDateRange tests date range parsing
func TestMulticloudCommands_ParseDateRange(t *testing.T) {
	providerManager := &providers.ProviderManager{}
	commands := NewMulticloudCommands(providerManager)

	// Valid range
	start, end, err := commands.parseDateRange("2025-01-01", "2025-01-31")
	if err != nil {
		t.Fatalf("Expected no error for valid date range, got: %v", err)
	}
	if !start.Before(end) {
		t.Error("Expected start date to be before end date")
	}

	// Empty dates -> defaults
	start2, end2, err2 := commands.parseDateRange("", "")
	if err2 != nil {
		t.Fatalf("Expected defaults without error, got: %v", err2)
	}
	if start2.IsZero() || end2.IsZero() {
		t.Error("Expected non-zero default dates")
	}

	// Invalid format
	if _, _, err3 := commands.parseDateRange("invalid-date", "2025-01-31"); err3 == nil {
		t.Error("Expected error for invalid date format")
	}

	// Start after end
	if _, _, err4 := commands.parseDateRange("2025-02-01", "2025-01-31"); err4 == nil {
		t.Error("Expected error when start date is after end date")
	}
}

// TestMulticloudCommands_ParseOptimizationTypes tests optimization types parsing
func TestMulticloudCommands_ParseOptimizationTypes(t *testing.T) {
	providerManager := &providers.ProviderManager{}
	commands := NewMulticloudCommands(providerManager)

	// Test empty input (should return all types)
	result1 := commands.parseOptimizationTypes([]string{})
	if len(result1) == 0 {
		t.Error("Expected non-empty result for empty input")
	}

	// Test single type
	result2 := commands.parseOptimizationTypes([]string{"right_sizing"})
	if len(result2) != 1 {
		t.Errorf("Expected 1 result, got %d", len(result2))
	}

	// Test multiple types
	result3 := commands.parseOptimizationTypes([]string{"right_sizing", "spot_instances"})
	if len(result3) != 2 {
		t.Errorf("Expected 2 results, got %d", len(result3))
	}

	// Test unknown type (should be ignored)
	result4 := commands.parseOptimizationTypes([]string{"unknown_type"})
	if len(result4) != 0 {
		t.Errorf("Expected 0 results for unknown type, got %d", len(result4))
	}
}

// TestMulticloudCommands_GetProviders tests provider selection
func TestMulticloudCommands_GetProviders(t *testing.T) {
	providerManager := &providers.ProviderManager{}
	commands := NewMulticloudCommands(providerManager)

	// Test with specific providers
	commands.flags.Providers = []string{"aws", "azure"}
	result1 := commands.getProviders()

	if len(result1) != 2 {
		t.Errorf("Expected 2 providers, got %d", len(result1))
	}

	if result1[0] != "aws" || result1[1] != "azure" {
		t.Errorf("Expected [aws, azure], got %v", result1)
	}

	// Test with empty providers (should return defaults)
	commands.flags.Providers = []string{}
	result2 := commands.getProviders()

	if len(result2) == 0 {
		t.Error("Expected non-empty default providers")
	}
}

// TestMulticloudCommands_CommandCreation tests individual command creation
func TestMulticloudCommands_CommandCreation(t *testing.T) {
	providerManager := &providers.ProviderManager{}
	commands := NewMulticloudCommands(providerManager)

	root := commands.BuildMulticloudCommand()
	if root == nil {
		t.Fatal("multicloud root is nil")
	}

	// Find optimize
	foundOptimize := false
	var optimizeCmd = &root.Commands()[0] // placeholder to avoid unused var
	_ = optimizeCmd
	for _, c := range root.Commands() {
		if c.Use == "optimize" {
			// Verify flags
			if c.Flags().Lookup("types") == nil {
				t.Error("Expected 'types' flag on optimize command")
			}
			if c.Flags().Lookup("risk-tolerance") == nil {
				t.Error("Expected 'risk-tolerance' flag on optimize command")
			}
			foundOptimize = true
		}
	}
	if !foundOptimize {
		t.Fatal("optimize subcommand not found")
	}

	// Find migrate and verify flags
	foundMigrate := false
	for _, c := range root.Commands() {
		if c.Use == "migrate" {
			if c.Flags().Lookup("source") == nil {
				t.Error("Expected 'source' flag on migrate command")
			}
			if c.Flags().Lookup("target") == nil {
				t.Error("Expected 'target' flag on migrate command")
			}
			foundMigrate = true
		}
	}
	if !foundMigrate {
		t.Fatal("migrate subcommand not found")
	}

	// Find discover and verify flags
	foundDiscover := false
	for _, c := range root.Commands() {
		if c.Use == "discover" {
			if c.Flags().Lookup("resource-types") == nil {
				t.Error("Expected 'resource-types' flag on discover command")
			}
			if c.Flags().Lookup("include-costs") == nil {
				t.Error("Expected 'include-costs' flag on discover command")
			}
			foundDiscover = true
		}
	}
	if !foundDiscover {
		t.Fatal("discover subcommand not found")
	}
}

// BenchmarkMulticloudCommands_BuildCommand benchmarks command building
func BenchmarkMulticloudCommands_BuildCommand(b *testing.B) {
	providerManager := &providers.ProviderManager{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		commands := NewMulticloudCommands(providerManager)
		_ = commands.BuildMulticloudCommand()
	}
}

// small helper to avoid extra imports
func contains(s, sub string) bool { return bytes.Contains([]byte(s), []byte(sub)) }
