package main

// Command Builders Code Generator
// Reads a YAML/JSON spec and generates Go code for Cobra commands.
// Usage:
//   go run ./scripts/tools/commandgen \
//     -package local/costscope/cmd/modules/analytics/commands \
//     -receiver AnalyticsCommands \
//     -spec cmd/modules/analytics/commands/command_spec.yaml \
//     -out cmd/modules/analytics/commands/zz_generated_command_builder.go
//
// Notes:
// - This tool focuses on reducing boilerplate by generating Use/Short/Long/Example and flags.
// - Handlers (RunE) should already exist as receiver methods like runXxx.
// - A simple hash of the spec is embedded for hash-guard tests.

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"
)

type FlagSpec struct {
	Name       string      `yaml:"name" json:"name"`
	Type       string      `yaml:"type" json:"type"` // string,int,bool,float,duration,stringSlice
	Shorthand  string      `yaml:"shorthand" json:"shorthand"`
	Default    interface{} `yaml:"default" json:"default"`
	Usage      string      `yaml:"usage" json:"usage"`
	Persistent bool        `yaml:"persistent" json:"persistent"`
	Required   bool        `yaml:"required" json:"required"`
	Bind       string      `yaml:"bind" json:"bind"` // Go selector relative to recv, e.g., flags.StartDate
}

type CommandNode struct {
	Use     string        `yaml:"use" json:"use"`
	Short   string        `yaml:"short" json:"short"`
	Long    string        `yaml:"long" json:"long"`
	Example string        `yaml:"example" json:"example"`
	Handler string        `yaml:"handler" json:"handler"`
	Flags   []FlagSpec    `yaml:"flags" json:"flags"`
	Sub     []CommandNode `yaml:"subcommands" json:"subcommands"`
}

type Spec struct {
	PackageImport string      `yaml:"package_import" json:"package_import"`
	PackageName   string      `yaml:"package_name" json:"package_name"`
	Receiver      string      `yaml:"receiver" json:"receiver"` // e.g., AnalyticsCommands or *AnalyticsCommands
	Root          CommandNode `yaml:"root" json:"root"`
}

// isAllowedSpecExt restricts spec files to YAML/JSON.
func isAllowedSpecExt(p string) bool {
	switch strings.ToLower(filepath.Ext(p)) {
	case ".yaml", ".yml", ".json":
		return true
	default:
		return false
	}
}

// mustReadSpec safely reads a spec file after basic validation to avoid G304 concerns.
func mustReadSpec(p string) []byte {
	clean := filepath.Clean(p)
	if !isAllowedSpecExt(clean) {
		log.Fatalf("unsupported spec extension for %q (allowed: .yaml, .yml, .json)", clean)
	}
	// Reject symlinks and ensure regular file
	if fi, err := os.Lstat(clean); err != nil {
		log.Fatalf("stat spec: %v", err)
	} else if fi.Mode()&os.ModeSymlink != 0 {
		log.Fatalf("spec path must not be a symlink: %s", clean)
	} else if !fi.Mode().IsRegular() {
		log.Fatalf("spec path must be a regular file: %s", clean)
	}
	b, err := os.ReadFile(clean)
	if err != nil {
		log.Fatalf("read spec: %v", err)
	}
	return b
}

func toHash(b []byte) string {
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

// emitFlag emits a pflag binding call, choosing Var/VarP vs non-binding forms.
func emitFlag(buf *strings.Builder, api, methodBase, bind, name, sh, defStr, usage string) {
	if bind != "" {
		if sh != "" {
			fmt.Fprintf(buf, "\t%s.%sVarP(&recv.%s, %q, %q, %s, %q)\n", api, methodBase, bind, name, sh, defStr, usage)
		} else {
			fmt.Fprintf(buf, "\t%s.%sVar(&recv.%s, %q, %s, %q)\n", api, methodBase, bind, name, defStr, usage)
		}
	} else {
		if sh != "" {
			fmt.Fprintf(buf, "\t%s.%sP(%q, %q, %s, %q)\n", api, methodBase, name, sh, defStr, usage)
		} else {
			fmt.Fprintf(buf, "\t%s.%s(%q, %s, %q)\n", api, methodBase, name, defStr, usage)
		}
	}
}

func emitFlagBinding(buf *strings.Builder, cmdVar string, f FlagSpec) {
	api := cmdVar + ".Flags()"
	if f.Persistent {
		api = cmdVar + ".PersistentFlags()"
	}
	def := f.Default
	bind := strings.TrimSpace(f.Bind)
	sh := strings.TrimSpace(f.Shorthand)
	switch f.Type {
	case "string":
		if def == nil {
			def = ""
		}
		emitFlag(buf, api, "String", bind, f.Name, sh, fmt.Sprintf("%q", def), f.Usage)
	case "stringSlice":
		emitFlag(buf, api, "StringSlice", bind, f.Name, sh, "[]string{}", f.Usage)
	case "bool":
		bdef := false
		if v, ok := def.(bool); ok {
			bdef = v
		}
		emitFlag(buf, api, "Bool", bind, f.Name, sh, fmt.Sprintf("%v", bdef), f.Usage)
	case "int":
		ival := 0
		if v, ok := def.(int); ok {
			ival = v
		}
		emitFlag(buf, api, "Int", bind, f.Name, sh, fmt.Sprintf("%d", ival), f.Usage)
	case "float":
		fval := 0.0
		if v, ok := def.(float64); ok {
			fval = v
		}
		emitFlag(buf, api, "Float64", bind, f.Name, sh, fmt.Sprintf("%f", fval), f.Usage)
	case "duration":
		emitFlag(buf, api, "Duration", bind, f.Name, sh, "0", f.Usage)
	default:
		fmt.Fprintf(buf, "\t// unsupported flag type: %s for %s\n", f.Type, f.Name)
	}
	if f.Required {
		fmt.Fprintf(buf, "\t_ = %s.MarkFlagRequired(\"%s\")\n", cmdVar, f.Name)
	}
}

func emitCommand(buf *strings.Builder, node CommandNode, varName string) {
	fmt.Fprintf(buf, "\t%s := &cobra.Command{\n", varName)
	fmt.Fprintf(buf, "\t\tUse:   %q,\n", node.Use)
	fmt.Fprintf(buf, "\t\tShort: %q,\n", node.Short)
	if strings.TrimSpace(node.Long) != "" {
		fmt.Fprintf(buf, "\t\tLong: %q,\n", node.Long)
	}
	if strings.TrimSpace(node.Example) != "" {
		fmt.Fprintf(buf, "\t\tExample: %q,\n", node.Example)
	}
	if node.Handler != "" {
		fmt.Fprintf(buf, "\t\tRunE: recv.%s,\n", node.Handler)
	} else {
		fmt.Fprintf(buf, "\t\tRunE: func(cmd *cobra.Command, args []string) error { return cmd.Help() },\n")
	}
	fmt.Fprintf(buf, "\t}\n\n")

	for _, fl := range node.Flags {
		emitFlagBinding(buf, varName, fl)
	}

	for i, sub := range node.Sub {
		childVar := fmt.Sprintf("%s_%d", varName, i)
		emitCommand(buf, sub, childVar)
		fmt.Fprintf(buf, "\t%s.AddCommand(%s)\n", varName, childVar)
	}
}

func main() {
	receiver := flag.String("receiver", "", "Receiver type name (e.g., AnalyticsCommands)")
	specPath := flag.String("spec", "", "Spec file path (YAML or JSON)")
	out := flag.String("out", "", "Output file path for generated code")
	flag.Parse()

	if *specPath == "" || *receiver == "" || *out == "" {
		log.Fatalf("missing required flags: -spec, -receiver, -out")
	}

	raw := mustReadSpec(*specPath)
	var s Spec
	if strings.HasSuffix(strings.ToLower(*specPath), ".json") {
		if err := json.Unmarshal(raw, &s); err != nil {
			log.Fatalf("parse JSON: %v", err)
		}
	} else {
		if err := yaml.Unmarshal(raw, &s); err != nil {
			log.Fatalf("parse YAML: %v", err)
		}
	}

	specHash := toHash(raw)

	var b strings.Builder
	b.WriteString("// Code generated by commandgen; DO NOT EDIT.\n")
	b.WriteString("// Spec: ")
	b.WriteString(filepath.Base(filepath.Clean(*specPath)))
	b.WriteString("\n")
	b.WriteString("// SpecHash: ")
	b.WriteString(specHash)
	b.WriteString("\n\n")
	pkgName := strings.TrimSpace(s.PackageName)
	if pkgName == "" {
		pkgName = "commands"
	}
	b.WriteString("package ")
	b.WriteString(pkgName)
	b.WriteString("\n\n")
	b.WriteString("import (\n\t\"github.com/spf13/cobra\"\n)\n\n")
	fmt.Fprintf(&b, "const generatedSpecHash = %q\n\n", specHash)

	recv := strings.TrimPrefix(*receiver, "*")
	// Root builder name: Build<Title(Use)>Command
	caser := cases.Title(language.English)
	titleUse := caser.String(s.Root.Use)
	fmt.Fprintf(&b, "// Build%sCommand builds the '%s' command and its subcommands (generated)\n", titleUse, s.Root.Use)
	fmt.Fprintf(&b, "func (recv *%s) Build%sCommand() *cobra.Command {\n", recv, titleUse)
	emitCommand(&b, s.Root, "root")
	b.WriteString("\n\treturn root\n}\n")

	// Use 0600 to satisfy gosec G306 (restrictive permissions for generated file)
	if err := os.WriteFile(filepath.Clean(*out), []byte(b.String()), 0600); err != nil {
		log.Fatalf("write output: %v", err)
	}
}
