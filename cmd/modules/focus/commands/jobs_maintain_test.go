package commands

import (
	"bytes"
	"fmt"
	"os"
	"testing"
	"time"

	store "github.com/costscope/costscope/internal/core/focus/conversion/store"
	"github.com/costscope/costscope/internal/core/focus/types"
)

// Test that the `jobs maintain --remove-older` command removes older entries from a Bolt DB
func TestJobsMaintain_RemoveOlder(t *testing.T) {
	tmp, err := os.CreateTemp("", "jobs-test-*.db")
	if err != nil {
		t.Fatalf("create temp db: %v", err)
	}
	path := tmp.Name()
	if err := tmp.Close(); err != nil {
		t.Fatalf("close temp file: %v", err)
	}
	defer func() {
		if err := os.Remove(path); err != nil {
			t.Logf("warning: failed to remove temp db: %v", err)
		}
	}()

	js, err := store.NewBoltJobStore(path)
	if err != nil {
		t.Fatalf("open bolt store: %v", err)
	}

	now := time.Now()

	old := &types.ConversionResult{ConversionId: "old", StartTime: now.Add(-72 * time.Hour), EndTime: now.Add(-48 * time.Hour)}

	if err := js.SaveResult(old); err != nil {
		t.Fatalf("save old result: %v", err)
	}

	recent := &types.ConversionResult{ConversionId: "recent", StartTime: now, EndTime: now}
	if err := js.SaveResult(recent); err != nil {
		t.Fatalf("save recent result: %v", err)
	}

	// close the store so the CLI command can open it (bbolt requires single writer handles)
	if err := js.Close(); err != nil {
		t.Fatalf("close bolt store: %v", err)
	}

	// Dry-run should report count without deleting (expect 1 old entry)
	cmdDry := NewJobsCommand()
	outDry := &bytes.Buffer{}
	cmdDry.SetOut(outDry)
	cmdDry.SetArgs([]string{"maintain", "--db-path", path, "--remove-older", "1d", "--dry-run"})
	if err := cmdDry.Execute(); err != nil {
		t.Fatalf("execute maintain dry-run command: %v", err)
	}
	// parse exact count from output: "Dry-run: %d entries would be removed older than %s\n"
	var parsedCount int
	var parsedTS string
	if n, _ := fmt.Sscanf(outDry.String(), "Dry-run: %d entries would be removed older than %s\n", &parsedCount, &parsedTS); n < 1 {
		t.Fatalf("failed to parse dry-run output: %q", outDry.String())
	}
	if parsedCount != 1 {
		t.Fatalf("expected dry-run count 1, got %d (output: %q)", parsedCount, outDry.String())
	}

	// Now execute the real prune to remove old entries
	cmd := NewJobsCommand()
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetArgs([]string{"maintain", "--db-path", path, "--remove-older", "1d"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute maintain command: %v", err)
	}

	// Reopen and assert only the recent entry remains
	js2, err := store.NewBoltJobStore(path)
	if err != nil {
		t.Fatalf("reopen bolt store: %v", err)
	}
	defer func() {
		if err := js2.Close(); err != nil {
			t.Fatalf("close bolt store: %v", err)
		}
	}()

	res := js2.ListResults(0)
	if len(res) != 1 {
		t.Fatalf("expected 1 result after prune, got %d", len(res))
	}
	if res[0].ConversionId != "recent" {
		t.Fatalf("expected remaining id 'recent', got '%s'", res[0].ConversionId)
	}
}

// Test that the `jobs maintain --compact` command runs and preserves entries
func TestJobsMaintain_Compact(t *testing.T) {
	tmp, err := os.CreateTemp("", "jobs-compact-*.db")
	if err != nil {
		t.Fatalf("create temp db: %v", err)
	}
	path := tmp.Name()
	if err := tmp.Close(); err != nil {
		t.Fatalf("close temp file: %v", err)
	}
	defer func() {
		if err := os.Remove(path); err != nil {
			t.Logf("warning: failed to remove temp db: %v", err)
		}
	}()

	js, err := store.NewBoltJobStore(path)
	if err != nil {
		t.Fatalf("open bolt store: %v", err)
	}

	now := time.Now()
	// seed a few entries
	for i := 0; i < 5; i++ {
		r := &types.ConversionResult{ConversionId: fmt.Sprintf("id-%s-%d", time.Now().Format("20060102150405"), i), StartTime: now, EndTime: now}
		if err := js.SaveResult(r); err != nil {
			t.Fatalf("save result: %v", err)
		}
		// tiny sleep to ensure different timestamps if needed
		time.Sleep(5 * time.Millisecond)
	}

	// close so the CLI can open the file
	if err := js.Close(); err != nil {
		t.Fatalf("close bolt store: %v", err)
	}

	cmd := NewJobsCommand()
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetArgs([]string{"maintain", "--db-path", path, "--compact"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute maintain compact command: %v", err)
	}

	// Dry-run for compact prints a message but does not modify
	cmd3 := NewJobsCommand()
	out3 := &bytes.Buffer{}
	cmd3.SetOut(out3)
	cmd3.SetArgs([]string{"maintain", "--db-path", path, "--compact", "--dry-run"})
	if err := cmd3.Execute(); err != nil {
		t.Fatalf("execute maintain compact dry-run command: %v", err)
	}
	if got := out3.String(); got == "" || !bytes.Contains([]byte(got), []byte("Dry-run:")) {
		t.Fatalf("expected dry-run output mentioning Dry-run for compact, got: %q", got)
	}

	// Reopen and assert entries still present
	js2, err := store.NewBoltJobStore(path)
	if err != nil {
		t.Fatalf("reopen bolt store: %v", err)
	}
	defer func() {
		if err := js2.Close(); err != nil {
			t.Fatalf("close bolt store: %v", err)
		}
	}()

	res := js2.ListResults(0)
	if len(res) == 0 {
		t.Fatalf("expected entries after compact, got none")
	}
}
