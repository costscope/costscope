package store

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	focustypes "local/costscope/internal/core/focus/types"
)

func TestBoltJobStore_SaveFinalizeList(t *testing.T) {
	// create temp db file
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "jobs.db")

	s, err := NewBoltJobStore(dbPath)
	if err != nil {
		t.Fatalf("NewBoltJobStore failed: %v", err)
	}
	defer func() {
		if err := s.Close(); err != nil {
			t.Fatalf("close store: %v", err)
		}
		if err := os.Remove(dbPath); err != nil {
			t.Logf("warning: failed to remove db file: %v", err)
		}
	}()

	base := time.Now()
	res := &focustypes.ConversionResult{ConversionId: "bconv-1", Success: true, StartTime: base}
	if err := s.SaveResult(res); err != nil {
		t.Fatalf("SaveResult failed: %v", err)
	}

	// finalize
	end := base.Add(5 * time.Second)
	dur := 5 * time.Second
	if err := s.FinalizeResultTiming("bconv-1", end, dur); err != nil {
		t.Fatalf("FinalizeResultTiming failed: %v", err)
	}

	list := s.ListResults(0)
	if len(list) != 1 {
		t.Fatalf("expected 1 result, got %d", len(list))
	}
	got := list[0]
	if got.ConversionId != "bconv-1" {
		t.Fatalf("unexpected id: %s", got.ConversionId)
	}
	if !got.EndTime.Equal(end) {
		t.Fatalf("end time mismatch: expected %v got %v", end, got.EndTime)
	}
	if got.Duration != dur {
		t.Fatalf("duration mismatch: expected %v got %v", dur, got.Duration)
	}
}
