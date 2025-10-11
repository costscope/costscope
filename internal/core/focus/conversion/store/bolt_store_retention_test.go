package store

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	focustypes "local/costscope/internal/core/focus/types"
)

func TestBoltJobStore_RemoveOlderThan(t *testing.T) {
	d := t.TempDir()
	path := filepath.Join(d, "test.db")
	bs, err := NewBoltJobStore(path)
	if err != nil {
		t.Fatalf("open bolt store: %v", err)
	}
	defer func() {
		if err := os.Remove(path); err != nil {
			t.Logf("warning: failed to remove test db: %v", err)
		}
	}()
	defer func() {
		if err := bs.Close(); err != nil {
			t.Fatalf("close bolt store: %v", err)
		}
	}()

	now := time.Now()
	old := &focustypes.ConversionResult{ConversionId: "old", StartTime: now.Add(-48 * time.Hour), EndTime: now.Add(-48 * time.Hour)}
	mid := &focustypes.ConversionResult{ConversionId: "mid", StartTime: now.Add(-24 * time.Hour), EndTime: now.Add(-24 * time.Hour)}
	new := &focustypes.ConversionResult{ConversionId: "new", StartTime: now, EndTime: now}

	if err := bs.SaveResult(old); err != nil {
		t.Fatalf("save old: %v", err)
	}
	if err := bs.SaveResult(mid); err != nil {
		t.Fatalf("save mid: %v", err)
	}
	if err := bs.SaveResult(new); err != nil {
		t.Fatalf("save new: %v", err)
	}

	removed, err := bs.RemoveOlderThan(now.Add(-36 * time.Hour))
	if err != nil {
		t.Fatalf("remove older: %v", err)
	}
	if removed != 1 {
		t.Fatalf("expected removed=1 got=%d", removed)
	}

	all := bs.ListResults(0)
	if len(all) != 2 {
		t.Fatalf("expected 2 remaining, got %d", len(all))
	}
}

func TestBoltJobStore_Compact(t *testing.T) {
	d := t.TempDir()
	path := filepath.Join(d, "compact.db")
	bs, err := NewBoltJobStore(path)
	if err != nil {
		t.Fatalf("open bolt store: %v", err)
	}
	defer func() {
		if err := bs.Close(); err != nil {
			t.Fatalf("close bolt store: %v", err)
		}
	}()

	// add a few entries
	for i := 0; i < 10; i++ {
		id := fmt.Sprintf("id-%d", i)
		res := &focustypes.ConversionResult{ConversionId: id, StartTime: time.Now()}
		if err := bs.SaveResult(res); err != nil {
			t.Fatalf("save %s: %v", id, err)
		}
	}

	// compact should succeed
	if err := bs.Compact(); err != nil {
		t.Fatalf("compact failed: %v", err)
	}

	// reopen and ensure entries are present
	all := bs.ListResults(0)
	if len(all) != 10 {
		t.Fatalf("expected 10 entries after compact, got %d", len(all))
	}
}
