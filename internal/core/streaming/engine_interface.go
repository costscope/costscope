package streaming

import "context"

// StreamingEngine defines the minimal public surface used by HTTP handlers and tests.
// Enterprise builds may provide extra methods (Pause, Stop) not required by the
// public HTTP layer; those are intentionally excluded to minimize stub surface.
type StreamingEngine interface {
	StartStreamingOperation(ctx context.Context, req *StreamingOperationRequest) (*StreamingOperation, error)
	GetStreamingOperation(id string) (*StreamingOperation, error)
	ResumeStreamingOperation(id string) error
	CancelStreamingOperation(id string) error
	ListActiveOperations() []*StreamingOperation
}
