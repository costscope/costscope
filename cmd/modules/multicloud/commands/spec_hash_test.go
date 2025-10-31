package commands

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"testing"
)

// TestSpecHashMatchesMulticloud ensures the generated code matches the current spec (hash guard).
func TestSpecHashMatchesMulticloud(t *testing.T) {
	// Read spec relative to this package directory
	path := "command_spec.yaml"
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read spec: %v", err)
	}
	sum := sha256.Sum256(canonicalizeSpec(b))
	got := hex.EncodeToString(sum[:])
	if got != generatedSpecHash {
		t.Fatalf("spec hash mismatch: generated=%s current=%s (run commandgen)", generatedSpecHash, got)
	}
}
