package testutil

import "testing"

func TestFindRepoRoot_NotEmpty(t *testing.T) {
	if got := FindRepoRoot(t); got == "" {
		t.Fatalf("FindRepoRoot returned empty string")
	}
}

func TestRepoRoot_Succeeds(t *testing.T) {
	if got, err := RepoRoot(); err != nil || got == "" {
		t.Fatalf("RepoRoot failed: got=%q err=%v", got, err)
	}
}
