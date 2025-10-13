package commands

import (
	"strings"
	"testing"

	"github.com/costscope/costscope/internal/providers"
)

// TestMulticloudCommands_CommandSurfaceSnapshot ensures newly added advanced commands stay stable (basic help surface)
func TestMulticloudCommands_CommandSurfaceSnapshot(t *testing.T) {
	pm := &providers.ProviderManager{}
	cmds := NewMulticloudCommands(pm)
	root := cmds.BuildMulticloudCommand()
	cmds.AttachEnhancedSubcommands(root)

	expected := []string{"recommendations", "inventory", "migration-plan", "feasibility"}
	for _, name := range expected {
		c, _, err := root.Find([]string{name, "--help"})
		if err != nil {
			// Cobra help path differs; fallback to manual lookup
			for _, sc := range root.Commands() {
				if sc.Use == name {
					c = sc
					break
				}
			}
		}
		if c == nil {
			t.Fatalf("command %s not found", name)
		}
		// Very lightweight snapshot: ensure short description keyword present
		short := strings.ToLower(c.Short)
		if name == "recommendations" && !strings.Contains(short, "recommend") {
			t.Errorf("recommendations short help mismatch: %s", c.Short)
		}
		if name == "inventory" && !strings.Contains(short, "inventory") {
			t.Errorf("inventory short help mismatch: %s", c.Short)
		}
		if name == "migration-plan" && !strings.Contains(short, "migration") {
			t.Errorf("migration-plan short help mismatch: %s", c.Short)
		}
		if name == "feasibility" && !strings.Contains(short, "feasibility") {
			t.Errorf("feasibility short help mismatch: %s", c.Short)
		}
	}
}
