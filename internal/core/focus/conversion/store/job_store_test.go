package store

import (
	"fmt"
	"testing"
	"time"

	focustypes "github.com/costscope/costscope/internal/core/focus/types"
)

func TestInMemoryJobStore_SaveAndList(t *testing.T) {
	s := NewInMemoryJobStore()

	now := time.Now()
	res := &focustypes.ConversionResult{ConversionId: "conv-1", Success: true, StartTime: now}

	if err := s.SaveResult(res); err != nil {
		t.Fatalf("SaveResult failed: %v", err)
	}

	list := s.ListResults(0)
	if len(list) != 1 {
		t.Fatalf("expected 1 result, got %d", len(list))
	}
	if list[0].ConversionId != "conv-1" {
		t.Fatalf("unexpected conversion id: %s", list[0].ConversionId)
	}
}

func TestInMemoryJobStore_FinalizeResultTiming(t *testing.T) {
	s := NewInMemoryJobStore()

	start := time.Now().Add(-time.Second)
	res := &focustypes.ConversionResult{ConversionId: "conv-2", Success: true, StartTime: start}
	if err := s.SaveResult(res); err != nil {
		t.Fatalf("SaveResult failed: %v", err)
	}

	end := time.Now()
	dur := end.Sub(start)
	if err := s.FinalizeResultTiming("conv-2", end, dur); err != nil {
		t.Fatalf("FinalizeResultTiming failed: %v", err)
	}

	list := s.ListResults(0)
	if len(list) != 1 {
		t.Fatalf("expected 1 result after finalize, got %d", len(list))
	}
	got := list[0]
	if !got.EndTime.Equal(end) {
		t.Fatalf("end time mismatch: expected %v got %v", end, got.EndTime)
	}
	if got.Duration != dur {
		t.Fatalf("duration mismatch: expected %v got %v", dur, got.Duration)
	}
}

// Integration-style test: save multiple results, finalize one, and verify ListResults limit
func TestInMemoryJobStore_ListLimitAndFinalize(t *testing.T) {
	s := NewInMemoryJobStore()

	base := time.Now()
	for i := 1; i <= 5; i++ {
		id := fmt.Sprintf("conv-%d", i)
		res := &focustypes.ConversionResult{ConversionId: id, Success: true, StartTime: base.Add(time.Duration(i) * time.Second)}
		if err := s.SaveResult(res); err != nil {
			t.Fatalf("SaveResult failed for %s: %v", id, err)
		}
	}

	// finalize conv-3
	end := base.Add(10 * time.Second)
	dur := 10 * time.Second
	if err := s.FinalizeResultTiming("conv-3", end, dur); err != nil {
		t.Fatalf("FinalizeResultTiming failed: %v", err)
	}

	// request only 2 most recent results
	list := s.ListResults(2)
	if len(list) != 2 {
		t.Fatalf("expected 2 results, got %d", len(list))
	}

	// verify finalized item exists when listing all
	all := s.ListResults(0)
	found := false
	for _, r := range all {
		if r.ConversionId == "conv-3" {
			found = true
			if !r.EndTime.Equal(end) || r.Duration != dur {
				t.Fatalf("finalize values mismatch: got end=%v dur=%v", r.EndTime, r.Duration)
			}
		}
	}
	if !found {
		t.Fatalf("conv-3 not found in all results")
	}
}
