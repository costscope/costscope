package commands

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

// Local copy of generator structs for deterministic hash in tests.
type testFlagSpec struct {
	Name       string      `yaml:"name" json:"name"`
	Type       string      `yaml:"type" json:"type"`
	Shorthand  string      `yaml:"shorthand" json:"shorthand"`
	Default    interface{} `yaml:"default" json:"default"`
	Usage      string      `yaml:"usage" json:"usage"`
	Persistent bool        `yaml:"persistent" json:"persistent"`
	Required   bool        `yaml:"required" json:"required"`
	Bind       string      `yaml:"bind" json:"bind"`
}

type testCommandNode struct {
	Use     string            `yaml:"use" json:"use"`
	Short   string            `yaml:"short" json:"short"`
	Long    string            `yaml:"long" json:"long"`
	Example string            `yaml:"example" json:"example"`
	Handler string            `yaml:"handler" json:"handler"`
	Flags   []testFlagSpec    `yaml:"flags" json:"flags"`
	Sub     []testCommandNode `yaml:"subcommands" json:"subcommands"`
}

type testSpec struct {
	PackageImport string          `yaml:"package_import" json:"package_import"`
	PackageName   string          `yaml:"package_name" json:"package_name"`
	Receiver      string          `yaml:"receiver" json:"receiver"`
	Root          testCommandNode `yaml:"root" json:"root"`
}

// TestGeneratedSpecHashMulticloud ensures the generatedSpecHash constant matches the spec on disk.
func TestGeneratedSpecHashMulticloud(t *testing.T) {
	p := filepath.Clean("command_spec.yaml")
	b, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			t.Skipf("spec %s not found", p)
		}
		t.Fatalf("read spec: %v", err)
	}
	var s testSpec
	if err := yaml.Unmarshal(b, &s); err != nil {
		t.Fatalf("parse spec: %v", err)
	}
	jb, _ := json.Marshal(s)
	h := sha256.Sum256(jb)
	actual := hex.EncodeToString(h[:])
	if actual != generatedSpecHash {
		t.Fatalf("spec hash drift: generatedSpecHash=%s current=%s; run 'make gen-commands' and commit", generatedSpecHash, actual)
	}
}
