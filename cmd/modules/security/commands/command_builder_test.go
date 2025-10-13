package commands

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/spf13/cobra"

	"github.com/costscope/costscope/internal/core/logging"
	"github.com/costscope/costscope/internal/testutil"
)

func TestSecurityCommands_Build(t *testing.T) {
	s := NewSecurityCommands(logging.NewLogger(logging.LevelError))
	root := s.BuildSecurityCommand()
	if root == nil || len(root.Commands()) == 0 {
		t.Fatalf("expected subcommands")
	}
}

func TestSecurityCommands_RoleCreate(t *testing.T) {
	// Isolate RBAC store by switching to a temp working directory
	cwd := testutil.FindRepoRoot(t) // ensure repo root resolved (cleanup sets back to repo root)
	t.Cleanup(func() { _ = os.Chdir(cwd) })
	tmp := t.TempDir()
	// Prepare data/security structure in temp dir
	_ = os.MkdirAll(filepath.Join(tmp, "data", "security"), 0o750)
	_ = os.Chdir(tmp)

	s := NewSecurityCommands(logging.NewLogger(logging.LevelError))
	root := s.BuildSecurityCommand()

	// find rbac create-role
	var create *cobra.Command
	for _, c := range root.Commands() {
		if c.Use == "rbac" {
			for _, sc := range c.Commands() {
				if sc.Use == "create-role" {
					create = sc
				}
			}
		}
	}
	if create == nil {
		t.Fatalf("create-role not found")
	}

	// Execute via root so Cobra resolves subcommand path correctly
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SilenceUsage = true
	root.SilenceErrors = true
	// Use a unique role name per run to avoid collisions even within the temp store
	roleName := "viewer-" + time.Now().UTC().Format("150405.000")
	root.SetArgs([]string{"rbac", "create-role", "--name", roleName, "--description", "read-only", "--permission", "reports:read"})
	if err := root.Execute(); err != nil {
		t.Fatalf("execute failed: %v", err)
	}

	// ensure JSON output
	var v map[string]any
	if err := json.Unmarshal(buf.Bytes(), &v); err != nil {
		t.Fatalf("invalid json output: %v", err)
	}
}
