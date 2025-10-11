package testutil

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// RepoRoot returns the directory containing go.mod, walking up from CWD.
// This is a non-test API suitable for use by commands or other non-test code.
func RepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("getwd: %w", err)
	}
	for i := 0; i < 12; i++ {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fmt.Errorf("could not locate repo root with go.mod")
}

// FindRepoRoot is a test-friendly wrapper around RepoRoot which fails the test
// on error and returns the discovered path.
func FindRepoRoot(t *testing.T) string {
	t.Helper()
	r, err := RepoRoot()
	if err != nil {
		t.Fatalf("locate repo root: %v", err)
	}
	return r
}
