package commands

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"testing"

	"gopkg.in/yaml.v3"
)

// Local typed struct for stable hashing (matches generator schema subset)
type shFlagSpec struct {
	Name       string      `yaml:"name" json:"name"`
	Type       string      `yaml:"type" json:"type"`
	Shorthand  string      `yaml:"shorthand" json:"shorthand"`
	Default    interface{} `yaml:"default" json:"default"`
	Usage      string      `yaml:"usage" json:"usage"`
	Persistent bool        `yaml:"persistent" json:"persistent"`
	Required   bool        `yaml:"required" json:"required"`
	Bind       string      `yaml:"bind" json:"bind"`
}
type shCommandNode struct {
	Use     string          `yaml:"use" json:"use"`
	Short   string          `yaml:"short" json:"short"`
	Long    string          `yaml:"long" json:"long"`
	Example string          `yaml:"example" json:"example"`
	Handler string          `yaml:"handler" json:"handler"`
	Flags   []shFlagSpec    `yaml:"flags" json:"flags"`
	Sub     []shCommandNode `yaml:"subcommands" json:"subcommands"`
}
type shSpec struct {
	PackageImport string        `yaml:"package_import" json:"package_import"`
	PackageName   string        `yaml:"package_name" json:"package_name"`
	Receiver      string        `yaml:"receiver" json:"receiver"`
	Root          shCommandNode `yaml:"root" json:"root"`
}

// TestSpecHashMatchesMulticloud ensures the generated code matches the current spec (hash guard).
func TestSpecHashMatchesMulticloud(t *testing.T) {
	path := "command_spec.yaml"
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read spec: %v", err)
	}
	var s shSpec
	if err := yaml.Unmarshal(b, &s); err != nil {
		t.Fatalf("parse spec: %v", err)
	}
	jb, _ := json.Marshal(s)
	sum := sha256.Sum256(jb)
	got := hex.EncodeToString(sum[:])
	if got != generatedSpecHash {
		t.Fatalf("spec hash mismatch: generated=%s current=%s (run commandgen)", generatedSpecHash, got)
	}
}
