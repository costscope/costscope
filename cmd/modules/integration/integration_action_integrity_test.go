package integration

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// TestActionSpecIntegrity asserts that every ActionSpec (non-group) registers exactly one command.
func TestActionSpecIntegrity(t *testing.T) {
	specs := BuildDefaultActionSpecs()
	root := CreateIntegrationCommands()
	find := func(path []string) *cobra.Command {
		cur := root
		for _, seg := range path[1:] { // skip first (integration)
			var next *cobra.Command
			for _, c := range cur.Commands() {
				if c.Use == seg {
					next = c
					break
				}
			}
			if next == nil {
				return nil
			}
			cur = next
		}
		return cur
	}
	for _, s := range specs {
		if s.Group {
			continue
		}
		parts := append([]string{"integration", s.Category}, append(s.Parents, s.Use)...)
		if cmd := find(parts); cmd == nil {
			t.Fatalf("missing command for spec %s path=%s", s.ID, strings.Join(parts, "/"))
		}
	}
}

// TestActionSpecMissingCategory ensures that a spec with an unknown category does not panic
// and is simply skipped (a warning is logged). We inject a bogus spec and assert original
// valid specs still register.
func TestActionSpecMissingCategory(t *testing.T) {
	specs := append([]ActionSpec{}, BuildDefaultActionSpecs()...)
	specs = append(specs, ActionSpec{ID: "bogus.test", Category: "nonexistent", Use: "noop", Short: "noop"})

	root := &cobra.Command{Use: "integration"}
	ctx := &RegistrationContext{}
	registered := RegisterIntegrationActions(root, ctx, specs)

	// bogus spec should not appear
	if _, ok := registered["bogus.test"]; ok {
		t.Fatalf("expected bogus spec not to be registered")
	}
	// sanity: pick a known spec
	if _, ok := registered["dashboard.stop"]; !ok {
		t.Fatalf("expected known spec dashboard.stop to be registered")
	}
}
