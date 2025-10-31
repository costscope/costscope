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

// canonicalizeSpec replicates the generator's hash canonicalization to avoid
// environment-specific drift (e.g., line endings or YAML formatting).
func canonicalizeSpec(raw []byte) []byte {
	var v any
	// Attempt YAML first (spec is YAML here too); fall back to JSON; else raw
	if err := yaml.Unmarshal(raw, &v); err == nil {
		if jb, err := json.Marshal(v); err == nil {
			return jb
		}
	}
	if err := json.Unmarshal(raw, &v); err == nil {
		if jb, err := json.Marshal(v); err == nil {
			return jb
		}
	}
	return raw
}

// TestGeneratedSpecHashMulticloud ensures the generatedSpecHash constant matches the spec on disk.
func TestGeneratedSpecHashMulticloud(t *testing.T) {
	// Read spec relative to this package directory to avoid GOPATH/WD ambiguity
	p := filepath.Clean("command_spec.yaml")
	b, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			t.Skipf("spec %s not found", p)
		}
		t.Fatalf("read spec: %v", err)
	}
	h := sha256.Sum256(canonicalizeSpec(b))
	actual := hex.EncodeToString(h[:])
	if actual != generatedSpecHash {
		t.Fatalf("spec hash drift: generatedSpecHash=%s current=%s; run 'make gen-commands' and commit", generatedSpecHash, actual)
	}
}
