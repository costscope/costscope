package jobs

import (
	"sync/atomic"
	"testing"
	"time"

	"local/costscope/internal/core/logging"
)

type mockBroadcaster struct {
	progress int64
	status   int64
}

func (m *mockBroadcaster) BroadcastJobProgress(jobID string, progress *Progress) {
	atomic.AddInt64(&m.progress, 1)
}

func (m *mockBroadcaster) BroadcastJobStatus(jobID string, status JobStatus, result map[string]interface{}, err string) {
	atomic.AddInt64(&m.status, 1)
}

func TestManagerBroadcasterIntegration(t *testing.T) {
	logger := logging.NewLogger(logging.LevelError)
	mgr := NewManager(logger, 1)
	if err := mgr.Start(); err != nil {
		t.Fatalf("failed to start manager: %v", err)
	}
	// Stop method removed; manager lifetime ends at test completion.

	mb := &mockBroadcaster{}
	mgr.SetBroadcaster(mb)

	cfg := &JobConfig{
		ID:       "job-bcast-1",
		Type:     "test",
		Priority: PriorityNormal,
		Timeout:  5 * time.Second,
	}
	job, err := mgr.SubmitJob(cfg, nil)
	if err != nil {
		t.Fatalf("submit job failed: %v", err)
	}

	// Wait for completion (MockTask takes ~3s)
	deadline := time.Now().Add(8 * time.Second)
	for time.Now().Before(deadline) {
		j, _ := mgr.GetJob(job.ID)
		if j.Status == StatusCompleted || j.Status == StatusFailed {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	if atomic.LoadInt64(&mb.progress) == 0 {
		t.Fatalf("expected progress broadcasts > 0, got %d", mb.progress)
	}
	if atomic.LoadInt64(&mb.status) == 0 {
		t.Fatalf("expected status broadcasts > 0, got %d", mb.status)
	}
}
