package security

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/costscope/costscope/internal/testutil"
)

// TestCasbinExamplesPresent validates example model/policy files exist and have basic markers.
func TestCasbinExamplesPresent(t *testing.T) {
	root := testutil.FindRepoRoot(t)
	model := filepath.Join(root, "configs", "rbac_model.conf.example")
	policy := filepath.Join(root, "configs", "rbac_policy.csv.example")

	if _, err := os.Stat(model); err != nil {
		t.Fatalf("missing model example: %s (%v)", model, err)
	}
	if _, err := os.Stat(policy); err != nil {
		t.Fatalf("missing policy example: %s (%v)", policy, err)
	}

	// #nosec G304 - test reads a known repo-local example file path
	b, err := os.ReadFile(model)
	if err != nil {
		t.Fatalf("read model example: %v", err)
	}
	if !strings.Contains(string(b), "[request_definition]") {
		t.Fatalf("model example missing request_definition section")
	}

	// #nosec G304 - test reads a known repo-local example file path
	pb, err := os.ReadFile(policy)
	if err != nil {
		t.Fatalf("read policy example: %v", err)
	}
	if !strings.Contains(string(pb), "role:admin") {
		t.Fatalf("policy example missing role:admin row")
	}
}

// TestMultiTenantDefaultsDisabled checks that shipped YAML configs have multi_tenant.enabled=false.
func TestMultiTenantDefaultsDisabled(t *testing.T) {
	root := testutil.FindRepoRoot(t)
	files := []string{"development.yaml", "staging.yaml", "production.yaml"}
	for _, name := range files {
		p := filepath.Join(root, "configs", name)
		// #nosec G304 - test reads repo-local config file
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatalf("read %s: %v", p, err)
		}
		s := string(b)
		if !strings.Contains(s, "multi_tenant:") {
			t.Fatalf("%s missing multi_tenant section", p)
		}
		if !strings.Contains(s, "enabled: false") {
			t.Fatalf("%s should set multi_tenant.enabled: false by default", p)
		}
	}
}

// findRepoRoot helper is defined in test_helpers_test.go
