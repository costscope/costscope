package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	integration "local/costscope/cmd/modules/integration"
	"local/costscope/internal/cli/specs"
)

// gen-integration-cli-docs: prototype generator for integration ActionSpec summary.
// Future: add markdown generation + snapshot drift diff.
func main() {
	out := flag.String("out", "integration_commands.json", "output JSON file")
	mdOut := flag.String("md-out", "integration_commands.md", "output Markdown file")
	driftCheck := flag.Bool("drift-check", false, "fail with non-zero exit if existing JSON differs")
	lintOnly := flag.Bool("lint-only", false, "only run validation/lint (no files written unless drift-check) and exit")
	schemaVersion := "1.1"
	flag.Parse()

	cmds := buildCommands()
	// (Output struct now in encodeJSON helper scope)
	list, dupIDs := collectCommands(cmds)

	tree := buildTree(list)

	// JSON encode & optional drift check
	tmp, buf := encodeJSON(schemaVersion, list, tree, dupIDs)
	performDriftCheck(*driftCheck, *out, buf)
	writeJSON(*lintOnly, *out, buf)
	// Markdown
	md := buildMarkdown(schemaVersion, tmp.Checksum, list, tree, dupIDs)
	writeMarkdown(*lintOnly, *mdOut, md)
	// Table
	printTable(list)
	finalMessage(*lintOnly, len(list), schemaVersion, tmp.Checksum, *out, *mdOut)
}

// escapePipes escapes pipe characters for table cells
func escapePipes(s string) string { return strings.ReplaceAll(s, "|", "\\|") }

// isSafePath performs a minimal safety validation to satisfy gosec (G304) without over-restricting legitimate filenames.
func isSafePath(p string) bool {
	if p == "" {
		return false
	}
	if strings.Contains(p, "..") {
		return false
	}
	if strings.HasPrefix(p, "/") {
		return false
	}
	return true
}

// ---- helpers extracted to reduce cyclomatic complexity of main ----

type Flag struct {
	Name     string      `json:"name"`
	Required bool        `json:"required"`
	Type     string      `json:"type"`
	Default  interface{} `json:"default,omitempty"`
	Usage    string      `json:"usage"`
}
type Cmd struct {
	ID       string `json:"id"`
	Use      string `json:"use"`
	Path     string `json:"path"`
	Category string `json:"category"`
	Short    string `json:"short"`
	Long     string `json:"long,omitempty"`
	Flags    []Flag `json:"flags"`
	Group    bool   `json:"group"`
	NumFlags int    `json:"num_flags"`
	Example  string `json:"example,omitempty"`
}

func buildCommands() map[string]*cobra.Command {
	root := &cobra.Command{Use: "costscope"}
	ctx := &integration.RegistrationContext{}
	specs := integration.BuildDefaultActionSpecs()
	return integration.RegisterIntegrationActions(root, ctx, specs)
}

func collectCommands(cmds map[string]*cobra.Command) ([]Cmd, []string) {
	var list []Cmd
	idSeen := map[string]int{}
	for id, c := range cmds {
		path := commandPath(c)
		cat := c.Annotations["category"]
		var flagDetails []Flag
		c.Flags().VisitAll(func(f *pflag.Flag) {
			_, required := f.Annotations[cobra.BashCompOneRequiredFlag]
			var def interface{}
			if f.DefValue != "" {
				def = f.DefValue
			}
			flagDetails = append(flagDetails, Flag{Name: f.Name, Required: required, Type: f.Value.Type(), Default: def, Usage: f.Usage})
		})
		isGroup := c.RunE == nil && len(flagDetails) == 0 && len(c.Commands()) > 0
		list = append(list, Cmd{ID: id, Use: c.Use, Path: path, Category: cat, Short: c.Short, Long: c.Long, Flags: flagDetails, Group: isGroup, NumFlags: len(flagDetails), Example: c.Example})
		idSeen[id]++
	}
	// simple insertion sort by path (avoids pulling full sort import logic again)
	for i := 0; i < len(list)-1; i++ {
		for j := i + 1; j < len(list); j++ {
			if list[j].Path < list[i].Path {
				list[i], list[j] = list[j], list[i]
			}
		}
	}
	var dupIDs []string
	for k, v := range idSeen {
		if v > 1 {
			dupIDs = append(dupIDs, k)
		}
	}
	if len(dupIDs) > 0 {
		fmt.Fprintf(os.Stderr, "Duplicate ActionSpec IDs detected: %v\n", dupIDs)
	}
	// insertion sort dupIDs
	for i := 0; i < len(dupIDs)-1; i++ {
		for j := i + 1; j < len(dupIDs); j++ {
			if dupIDs[j] < dupIDs[i] {
				dupIDs[i], dupIDs[j] = dupIDs[j], dupIDs[i]
			}
		}
	}
	return list, dupIDs
}

func commandPath(c *cobra.Command) string {
	path := c.Name()
	parent := c.Parent()
	for parent != nil && parent.Name() != "costscope" {
		path = parent.Name() + " " + path
		parent = parent.Parent()
	}
	return path
}

func buildTree(list []Cmd) []string {
	var tree []string
	for _, c := range list {
		depth := strings.Count(c.Path, " ") + 1
		tree = append(tree, fmt.Sprintf("%s%s", strings.Repeat("  ", depth-1), c.Path))
	}
	return tree
}

// --- extracted helpers to reduce main complexity ---
func encodeJSON(schemaVersion string, list []Cmd, tree []string, dupIDs []string) (struct{ Checksum string }, *bytes.Buffer) {
	tmp := struct {
		SchemaVersion string   `json:"schema_version"`
		CommandCount  int      `json:"command_count"`
		Checksum      string   `json:"checksum"`
		Commands      []Cmd    `json:"commands"`
		Tree          []string `json:"tree"`
		DuplicateIDs  []string `json:"duplicate_ids,omitempty"`
	}{SchemaVersion: schemaVersion, CommandCount: len(list), Commands: list, Tree: tree, DuplicateIDs: dupIDs}
	raw, err := json.Marshal(tmp)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error encoding: %v\n", err)
		os.Exit(1)
	}
	tmp.Checksum = specs.ComputeChecksumHex(raw)
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	enc.SetIndent("", "  ")
	if err := enc.Encode(tmp); err != nil {
		fmt.Fprintf(os.Stderr, "error encoding: %v\n", err)
		os.Exit(1)
	}
	return struct{ Checksum string }{Checksum: tmp.Checksum}, buf
}

func performDriftCheck(enabled bool, out string, buf *bytes.Buffer) {
	if !enabled {
		return
	}
	if !isSafePath(out) {
		fmt.Fprintf(os.Stderr, "unsafe output path: %s\n", out)
		os.Exit(1)
	}
	existing, err := os.ReadFile(out) // #nosec G304 path validated by isSafePath
	if err != nil {
		fmt.Fprintf(os.Stderr, " Drift check failed: existing file %s not found\n", out)
		os.Exit(2)
	}
	if !bytes.Equal(bytes.TrimSpace(existing), bytes.TrimSpace(buf.Bytes())) {
		fmt.Fprintf(os.Stderr, " Drift detected for %s (run make gen-integration-cli-docs)\n", out)
		os.Exit(2)
	}
}

func writeJSON(lintOnly bool, out string, buf *bytes.Buffer) {
	if lintOnly {
		return
	}
	if !isSafePath(out) {
		fmt.Fprintf(os.Stderr, "unsafe output path: %s\n", out)
		os.Exit(1)
	}
	if err := os.WriteFile(out, buf.Bytes(), 0o600); err != nil {
		fmt.Fprintf(os.Stderr, "error writing output: %v\n", err)
		os.Exit(1)
	}
}

func buildMarkdown(schemaVersion, checksum string, list []Cmd, tree, dupIDs []string) *bytes.Buffer {
	var md bytes.Buffer
	md.WriteString("# Integration Commands\n\n")
	md.WriteString("Auto-generated summary of declarative integration ActionSpecs. Do not edit manually.\n\n")
	md.WriteString(fmt.Sprintf("Schema Version: %s  \\n+Checksum: `%s`  \\n+Commands: %d\n\n", schemaVersion, checksum, len(list)))
	if len(dupIDs) > 0 {
		md.WriteString(fmt.Sprintf("**WARNING:** Duplicate IDs detected: %v\n\n", dupIDs))
	}
	md.WriteString("## Summary Table\n\n")
	md.WriteString("| Path | Category | Description | Flags |\n|------|----------|-------------|-------|\n")
	for _, c := range list {
		desc := c.Short
		if desc == "" && c.Group {
			desc = "Group command"
		}
		_, _ = md.WriteString(fmt.Sprintf("| `%s` | %s | %s | %d |\n", c.Path, c.Category, escapePipes(desc), c.NumFlags))
	}
	md.WriteString("\n## Command Tree\n\n````\n")
	for _, line := range tree {
		_, _ = md.WriteString(line + "\n")
	}
	md.WriteString("````\n\n## Command Details\n\n")
	for _, c := range list {
		_, _ = md.WriteString(fmt.Sprintf("### `%s`\n\n", c.Path))
		if c.Short != "" {
			_, _ = md.WriteString(c.Short + "\n\n")
		}
		if c.Long != "" && c.Long != c.Short {
			_, _ = md.WriteString(c.Long + "\n\n")
		}
		if c.Group {
			_, _ = md.WriteString("_Group command (no direct execution)._\n\n")
		}
		if c.Example != "" {
			_, _ = md.WriteString("**Example:**\\n\n````bash\n" + c.Example + "\n````\n\n")
		}
		if len(c.Flags) == 0 {
			_, _ = md.WriteString("No flags.\n\n")
			continue
		}
		_, _ = md.WriteString("| Flag | Type | Required | Default | Usage |\n|------|------|----------|---------|-------|\n")
		for _, f := range c.Flags {
			def := fmt.Sprintf("%v", f.Default)
			if def == "" {
				def = "-"
			}
			req := "no"
			if f.Required {
				req = "yes"
			}
			_, _ = md.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %s |\n", f.Name, f.Type, req, def, escapePipes(f.Usage)))
		}
		_, _ = md.WriteString("\n")
	}
	return &md
}

func writeMarkdown(lintOnly bool, mdOut string, md *bytes.Buffer) {
	if lintOnly {
		return
	}
	if !isSafePath(mdOut) {
		fmt.Fprintf(os.Stderr, "unsafe output path: %s\n", mdOut)
		os.Exit(1)
	}
	if err := os.WriteFile(mdOut, md.Bytes(), 0o600); err != nil {
		fmt.Fprintf(os.Stderr, "error writing markdown: %v\n", err)
		os.Exit(1)
	}
}

func printTable(list []Cmd) {
	tw := tabwriter.NewWriter(os.Stdout, 2, 4, 2, ' ', 0)
	_, _ = fmt.Fprintln(tw, "PATH\tCATEGORY\tFLAGS")
	for _, c := range list {
		flagNames := make([]string, len(c.Flags))
		for i, f := range c.Flags {
			flagNames[i] = f.Name
		}
		_, _ = fmt.Fprintf(tw, "%s\t%s\t%s\n", c.Path, c.Category, strings.Join(flagNames, ","))
	}
	_ = tw.Flush()
}

func finalMessage(lintOnly bool, count int, schemaVersion, checksum, out, mdOut string) {
	if lintOnly {
		fmt.Printf("Lint-only: %d commands processed (schema %s) checksum %s\n", count, schemaVersion, checksum)
		return
	}
	fmt.Printf("\nWrote %d integration commands to %s and %s (schema %s, checksum %s)\n", count, out, mdOut, schemaVersion, checksum)
}
