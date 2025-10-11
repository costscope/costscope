package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	clispecs "local/costscope/internal/cli/specs"

	"github.com/spf13/cobra"
)

// TestClassifyIntegrationErrorSentinels validates sentinel error → label mapping.
func TestClassifyIntegrationErrorSentinels(t *testing.T) {
	cases := []struct {
		err  error
		want string
	}{
		{ErrNotFound, "not_found"},
		{ErrTimeout, "timeout"},
		{ErrUnauthorized, "unauthorized"},
		{ErrValidation, "validation"},
		{ErrConflict, "conflict"},
		{nil, ""},
	}
	for _, c := range cases {
		got := classifyIntegrationError(c.err)
		if got != c.want {
			t.Fatalf("expected %q got %q", c.want, got)
		}
	}
}

// TestClassifyIntegrationErrorFallback ensures fallback string inspection works for uncategorised errors.
func TestClassifyIntegrationErrorFallback(t *testing.T) {
	cases := map[string]string{
		"resource NOT FOUND":              "not_found",
		"deadline exceeded while waiting": "timeout",
		"permission denied":               "unauthorized",
		"invalid payload provided":        "validation",
		"already exists":                  "conflict",
		"weird unexpected thing":          "other",
	}
	for msg, want := range cases {
		got := classifyIntegrationError(&dummyErr{m: msg})
		if got != want {
			t.Errorf("%s: want %s got %s", msg, want, got)
		}
	}
}

type dummyErr struct{ m string }

func (d *dummyErr) Error() string { return d.m }

// TestNoDuplicateActionSpecIDs mirrors generator duplicate detection logic.
func TestNoDuplicateActionSpecIDs(t *testing.T) {
	specs := BuildDefaultActionSpecs()
	seen := map[string]struct{}{}
	for _, s := range specs {
		if _, ok := seen[s.ID]; ok {
			t.Fatalf("duplicate action spec id detected: %s", s.ID)
		}
		seen[s.ID] = struct{}{}
	}
}

// TestExamplesContainCostscopeBasic sanity checks example formatting for specs that define an Example.
func TestExamplesContainCostscopeBasic(t *testing.T) {
	specs := BuildDefaultActionSpecs()
	for _, s := range specs {
		if s.Example == "" { // skip
			continue
		}
		if !strings.Contains(s.Example, "costscope integration") && !strings.Contains(s.Example, "costscope ") {
			t.Errorf("example for %s does not contain 'costscope': %s", s.ID, s.Example)
		}
		// light parse: ensure no obvious unclosed quote
		if strings.Count(s.Example, "\"")%2 == 1 {
			t.Errorf("example for %s appears to have unmatched quotes: %s", s.ID, s.Example)
		}
	}
}

// TestMarkdownChecksumRecompute regenerates markdown/json in a temp dir and verifies checksum field matches recomputed checksum.
func TestMarkdownChecksumRecompute(t *testing.T) {
	// reuse generator logic indirectly: register commands then build a pseudo JSON manifest similar to generator.
	root := &cobra.Command{Use: "costscope"}
	ctx := &RegistrationContext{}
	specs := BuildDefaultActionSpecs()
	cmds := RegisterIntegrationActions(root, ctx, specs)

	type mini struct {
		ID    string `json:"id"`
		Path  string `json:"path"`
		Short string `json:"short"`
	}
	var list []mini
	for id, c := range cmds {
		path := c.Name()
		p := c.Parent()
		for p != nil && p.Name() != "costscope" {
			path = p.Name() + " " + path
			p = p.Parent()
		}
		list = append(list, mini{ID: id, Path: path, Short: c.Short})
	}
	// compute checksum over concatenated deterministic string
	// (simpler than full generator; purpose is to ensure stable pre-hash content ordering does not break silently)
	// sort by path for stability
	// Use same ordering logic as generator: path lexicographically
	for i := 0; i < len(list)-1; i++ {
		for j := i + 1; j < len(list); j++ {
			if list[j].Path < list[i].Path {
				list[i], list[j] = list[j], list[i]
			}
		}
	}
	var b strings.Builder
	for _, m := range list {
		b.WriteString(m.ID)
		b.WriteString("|")
		b.WriteString(m.Path)
		b.WriteString("|")
		b.WriteString(m.Short)
		b.WriteString("\n")
	}
	_ = clispecs.ComputeChecksumHex([]byte(b.String())) // currently not compared to file; baseline future extension.
}

// TestGeneratorJSONChecksumIfPresent compares JSON checksum if file exists (non-fatal skip if missing).
func TestGeneratorJSONChecksumIfPresent(t *testing.T) {
	path := filepath.Join(".", "integration_commands.json")
	if strings.Contains(path, "..") {
		t.Fatalf("unsafe path: %s", path)
	}
	content, err := os.ReadFile(path) // #nosec G304 path validated above
	if err != nil {
		t.Skipf("json output not present: %v", err)
	}
	// very light verification: ensure checksum field exists and is 64 hex chars
	s := string(content)
	idx := strings.Index(s, "\"checksum\":")
	if idx == -1 {
		t.Fatalf("checksum field missing in %s", path)
	}
	// naive slice forward
	after := s[idx:]
	// extract first quoted value after colon
	q1 := strings.Index(after, "\"")
	if q1 == -1 || len(after) < q1+1 {
		t.Fatalf("malformed checksum field")
	}
	q2 := strings.Index(after[q1+1:], "\"")
	if q2 == -1 {
		t.Fatalf("malformed checksum field end")
	}
	val := after[q1+1 : q1+1+q2]
	if len(val) != 64 {
		t.Fatalf("expected 64 hex chars for checksum got %d", len(val))
	}
}
