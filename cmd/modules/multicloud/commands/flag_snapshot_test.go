package commands

import (
	_ "embed"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"local/costscope/internal/providers"

	"github.com/spf13/pflag"
)

//go:embed testdata/command_flags_snapshot.json
var flagSnapshotBaseline []byte

// TestMulticloudCommands_FlagSnapshot captures the flag surface (names only) for each multicloud command.
// Update by setting UPDATE_SNAPSHOT=1 env var and re-running this test if intentional additive changes occur.
func TestMulticloudCommands_FlagSnapshot(t *testing.T) {
	pm := &providers.ProviderManager{}
	cmds := NewMulticloudCommands(pm)
	root := cmds.BuildMulticloudCommand()
	cmds.AttachEnhancedSubcommands(root)

	snapshot := make(map[string][]string)

	// Root (persistent flags)
	root.Flags().VisitAll(func(f *pflag.Flag) {}) // no-op placeholder to ensure compilation if refactoring occurs

	// Collect persistent flags explicitly
	var rootFlagNames []string
	root.PersistentFlags().VisitAll(func(f *pflag.Flag) { rootFlagNames = append(rootFlagNames, f.Name) })
	sort.Strings(rootFlagNames)
	snapshot[root.Use] = rootFlagNames

	for _, c := range root.Commands() {
		var names []string
		c.Flags().VisitAll(func(f *pflag.Flag) { names = append(names, f.Name) })
		sort.Strings(names)
		snapshot[c.Use] = names
	}

	// Load embedded snapshot baseline (avoids G304 variable path read warning)
	path := filepath.Join("testdata", "command_flags_snapshot.json") // used only when updating snapshot
	if len(flagSnapshotBaseline) == 0 {
		t.Fatalf("embedded snapshot baseline is empty")
	}
	var want map[string][]string
	if err := json.Unmarshal(flagSnapshotBaseline, &want); err != nil {
		t.Fatalf("invalid embedded snapshot JSON: %v", err)
	}

	// Normalize ordering in expected snapshot (older snapshots may be unsorted)
	for _, v := range want {
		sort.Strings(v)
	}

	// If update requested, rewrite snapshot
	if os.Getenv("UPDATE_SNAPSHOT") == "1" {
		out, _ := json.MarshalIndent(snapshot, "", "  ")
		// Restrictive permissions (0600) to satisfy gosec G306
		if err := os.WriteFile(path, out, 0600); err != nil {
			t.Fatalf("failed writing snapshot: %v", err)
		}
		// Note: we don't update the embedded copy automatically; run `go generate` with embedding if changed
		t.Skip("snapshot updated on disk; re-run tests after embedding refresh")
	}

	// Compare
	if diff := diffFlagSnapshots(want, snapshot); diff != "" {
		t.Fatalf("flag snapshot mismatch:\n%s", diff)
	}
}

// diffFlagSnapshots produces a deterministic diff between two snapshots.
func diffFlagSnapshots(a, b map[string][]string) string {
	// Simple textual diff: list commands added/removed or flag set changes.
	var out string
	// Collect union of command names
	seen := map[string]struct{}{}
	for k := range a {
		seen[k] = struct{}{}
	}
	for k := range b {
		seen[k] = struct{}{}
	}
	var cmds []string
	for k := range seen {
		cmds = append(cmds, k)
	}
	sort.Strings(cmds)
	for _, c := range cmds {
		av, aok := a[c]
		bv, bok := b[c]
		if !aok {
			out += "+ command added: " + c + "\n"
			continue
		}
		if !bok {
			out += "- command removed: " + c + "\n"
			continue
		}
		// Compare flag slices
		if !equalStringSlices(av, bv) {
			out += "~ flags changed for command " + c + "\n  want: " + sliceToString(av) + "\n  got : " + sliceToString(bv) + "\n"
		}
	}
	return out
}

func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func sliceToString(s []string) string {
	if len(s) == 0 {
		return "[]"
	}
	out := "["
	for i, v := range s {
		if i > 0 {
			out += ", "
		}
		out += v
	}
	out += "]"
	return out
}
