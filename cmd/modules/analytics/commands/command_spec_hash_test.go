package commands

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
)

// TestGeneratedSpecHashAnalytics ensures the generatedSpecHash constant matches the spec on disk.
func TestGeneratedSpecHashAnalytics(t *testing.T) {
	// Read spec relative to this package directory (skip if missing)
	p := filepath.Clean("command_spec.yaml")
	b, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			t.Skipf("spec %s not found", p)
		}
		t.Fatalf("read spec: %v", err)
	}
	h := sha256.Sum256(b)
	actual := hex.EncodeToString(h[:])
	if actual != generatedSpecHash {
		t.Fatalf("spec hash drift: generatedSpecHash=%s current=%s; run 'make gen-commands' and commit", generatedSpecHash, actual)
	}
}
