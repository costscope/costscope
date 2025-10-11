package performance

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestExecuteParallel_Basic(t *testing.T) {
	pe := NewPerformanceEngine(DefaultPerformanceConfig())
	if err := pe.Start(context.Background()); err != nil {
		t.Fatalf("start: %v", err)
	}
	defer func() { _ = pe.Stop() }()
	jobs := []Job{{ID: "a", Data: 2, Processor: func(v interface{}) (interface{}, error) { return v.(int) * 3, nil }}, {ID: "b", Data: 5, Processor: func(v interface{}) (interface{}, error) { return v.(int) + 7, nil }}}
	res, err := pe.ExecuteParallel(context.Background(), jobs)
	if err != nil {
		t.Fatalf("ExecuteParallel error: %v", err)
	}
	if len(res) != 2 {
		t.Fatalf("expected 2 results got %d", len(res))
	}
	if res[0].Error != nil || res[1].Error != nil {
		t.Fatalf("unexpected job errors: %+v", res)
	}
}

func TestExecuteParallel_TimeoutAndNil(t *testing.T) {
	pe := NewPerformanceEngine(DefaultPerformanceConfig())
	if err := pe.Start(context.Background()); err != nil {
		t.Fatalf("start: %v", err)
	}
	defer func() { _ = pe.Stop() }()
	jobs := []Job{
		{ID: "timeout", Timeout: 10 * time.Millisecond, Processor: func(interface{}) (interface{}, error) {
			time.Sleep(50 * time.Millisecond)
			return 1, nil
		}},
		{ID: "nilproc"}, // nil processor
	}
	res, err := pe.ExecuteParallel(context.Background(), jobs)
	if err != nil { // context error only surfaces as err; per-job errors embedded
		t.Fatalf("unexpected top-level error: %v", err)
	}
	if len(res) != 2 {
		t.Fatalf("expected 2 results got %d", len(res))
	}
	foundNil := false
	foundTimeout := false
	for _, r := range res {
		if r.JobID == "nilproc" && r.Error != nil && strings.Contains(r.Error.Error(), "nil processor") {
			foundNil = true
		}
		if r.JobID == "timeout" && r.Error != nil && (errors.Is(r.Error, context.DeadlineExceeded) || strings.Contains(r.Error.Error(), "timeout")) {
			foundTimeout = true
		}
	}
	if !foundNil || !foundTimeout {
		t.Fatalf("expected nil processor and timeout errors, got: %+v", res)
	}
}
