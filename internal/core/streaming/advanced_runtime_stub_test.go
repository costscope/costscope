//go:build !enterprise

package streaming

import (
	"testing"

	streamingTypes "github.com/costscope/costscope/cmd/modules/streaming/types"
)

// Test ensuring advanced streaming runtime constructors yield disabled behavior in non-enterprise builds.
func TestAdvancedStreamingRuntimeStub(t *testing.T) {
	jm := NewJobManager(nil, "")
	if _, err := jm.StartJob(&streamingTypes.StreamingJobConfig{}); err == nil {
		// We expect disabled error; nil = test failure
		// Using string contains instead of type unwrap to remain stable to future enterprise error wrapping changes
		// but minimal: just assert non-nil
		// Intentional failure path: early exit
		// Adjust message for clarity
		// (Keep test short & deterministic.)
		//
		// NOTE: This branch should never run under !enterprise build.
		// Fail fast if disabled error contract is broken.
		t.Fatalf("expected disabled error from StartJob in stub")
	}
	p := NewPipeline(nil, PipelineOptions{}, nil)
	if err := p.Start(); err == nil {
		// The stub Start should return disabled error (not nil)
		// If semantics change later (e.g., Start becomes no-op), update this test accordingly.
		t.Fatalf("expected disabled error from pipeline Start in stub")
	}
}
