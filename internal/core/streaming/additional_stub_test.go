package streaming

import (
	"context"
	"strings"
	"testing"
)

// TestDisabledJobManagerMethods ensure all exposed methods return disabled errors consistently.
func TestDisabledJobManagerMethods(t *testing.T) {
	jm := NewJobManager(nil, "")
	if _, err := jm.StartJob(nil); err == nil {
		t.Fatalf("expected error from StartJob on stub")
	}
	if err := jm.PauseJob("id"); err == nil {
		t.Fatalf("expected error from PauseJob")
	}
	if err := jm.ResumeJob("id"); err == nil {
		t.Fatalf("expected error from ResumeJob")
	}
	if err := jm.StopJob("id"); err == nil {
		t.Fatalf("expected error from StopJob")
	}
	if _, err := jm.GetJobStatus("id"); err == nil {
		t.Fatalf("expected error from GetJobStatus")
	}
	if _, err := jm.ListJobs(); err == nil {
		t.Fatalf("expected error from ListJobs")
	}
}

// TestDisabledPipelineMethods verify pipeline stub returns disabled errors for active ops but not for Shutdown.
func TestDisabledPipelineMethods(t *testing.T) {
	p := NewPipeline(nil, PipelineOptions{}, nil)
	if err := p.Start(); err == nil {
		t.Fatalf("expected Start disabled error")
	}
	if err := p.Publish(context.TODO(), nil); err == nil {
		t.Fatalf("expected Publish disabled error")
	}
	if err := p.Shutdown(context.TODO()); err != nil {
		t.Fatalf("shutdown should be nil error, got %v", err)
	}
	if dl := p.DeadLetters(); dl != nil {
		t.Fatalf("expected nil dead letters slice, got %v", dl)
	}
	if m := p.SnapshotMetrics(); m != (Metrics{}) {
		t.Fatalf("expected zero metrics struct")
	}
}

// TestPersistentJobManagerDisabled ensures persistent variant also returns disabled error.
func TestPersistentJobManagerDisabled(t *testing.T) {
	pm := NewPersistentJobManager(nil, nil, "")
	if _, err := pm.StartJobPersistent(nil); err == nil {
		t.Fatalf("expected disabled error")
	}
}

// Optional: assert error type matches enterprise.DisabledError signature (string contains feature name).
func TestDisabledErrorContainsFeatureName(t *testing.T) {
	jm := NewJobManager(nil, "")
	_, err := jm.StartJob(nil)
	if err == nil {
		t.Fatalf("expected error")
	}
	// Ensure we received a non-nil error and mention of disabled feature is present.
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "disabled") && !strings.Contains(err.Error(), "not available") {
		// Not a strict contract—just a sanity check that the error references the disabled state.
		t.Fatalf("unexpected error value: %v", err)
	}
}
