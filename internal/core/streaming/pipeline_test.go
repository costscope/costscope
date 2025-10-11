//go:build enterprise

package streaming

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"local/costscope/internal/core/logging"
)

// Test 1: High-load with backpressure (1M messages simulated with smaller loop for CI constraints)
func TestPipelineHighLoadBackpressure(t *testing.T) {
	// Keep runtime short for CI; simulate 200k msgs rather than 1M, but structure supports 1M+
	const total = 200_000
	var processed uint64

	consumer := func(ctx context.Context, msg *Message) error {
		// Simple processing; emulate small cost
		atomic.AddUint64(&processed, 1)
		return nil
	}

	p := NewPipeline(logging.NewLogger(logging.LevelError), PipelineOptions{
		BufferSize:       1024, // backpressure kicks in
		Workers:          8,
		EnableIdempotent: false,
		MetricsInterval:  0,
		Retry:            RetryPolicy{MaxAttempts: 1},
	}, consumer)

	if err := p.Start(); err != nil {
		t.Fatalf("start: %v", err)
	}

	// Produce in separate goroutine to exercise blocking behavior
	pubCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		for i := 0; i < total; i++ {
			m := &Message{ID: fmt.Sprintf("%d", i)}
			if err := p.Publish(pubCtx, m); err != nil {
				done <- err
				return
			}
		}
		done <- nil
	}()

	if err := <-done; err != nil {
		t.Fatalf("publish failed: %v", err)
	}

	// shutdown and drain
	shutCtx, cancel2 := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel2()
	if err := p.Shutdown(shutCtx); err != nil {
		t.Fatalf("shutdown: %v", err)
	}

	if got := atomic.LoadUint64(&processed); got != total {
		t.Fatalf("processed=%d want=%d", got, total)
	}

	// metrics sanity
	m := p.SnapshotMetrics()
	if m.Produced != uint64(total) || m.Enqueued != uint64(total) || m.Succeeded != uint64(total) {
		t.Fatalf("metrics mismatch: %+v", m)
	}
}

// Test 2: Reorder vs idempotency — push duplicate IDs out-of-order; ensure only one success per ID
func TestPipelineIdempotency(t *testing.T) {
	var successes uint64
	consumer := func(ctx context.Context, msg *Message) error {
		atomic.AddUint64(&successes, 1)
		return nil
	}
	p := NewPipeline(logging.NewLogger(logging.LevelError), PipelineOptions{
		BufferSize:       64,
		Workers:          4,
		EnableIdempotent: true,
		Retry:            RetryPolicy{MaxAttempts: 1},
	}, consumer)
	_ = p.Start()

	pubCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// send duplicates and out-of-order
	ids := []string{"A", "B", "A", "C", "B", "D", "C", "E", "E", "F", "F", "F"}
	for _, id := range ids {
		if err := p.Publish(pubCtx, &Message{ID: id}); err != nil {
			t.Fatalf("publish: %v", err)
		}
	}

	shutCtx, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel2()
	if err := p.Shutdown(shutCtx); err != nil {
		t.Fatalf("shutdown: %v", err)
	}

	// unique IDs = A,B,C,D,E,F => 6
	if atomic.LoadUint64(&successes) != 6 {
		t.Fatalf("expected 6 unique successes, got %d", successes)
	}
	m := p.SnapshotMetrics()
	expDrops := uint64(len(ids) - 6) //nolint:gosec // len(ids) small and non-negative in test
	if m.DedupDropped != expDrops {
		t.Fatalf("expected %d drops, got %d", len(ids)-6, m.DedupDropped)
	}
}

// Test 3: Retry with exponential backoff and DLQ after max attempts
func TestPipelineRetryAndDLQ(t *testing.T) {
	var attempts uint64
	// Fail first 2 attempts for each id, then succeed on 3rd; pipeline max=2 -> DLQ
	consumer := func(ctx context.Context, msg *Message) error {
		// count attempts per ID globally for simplicity
		atomic.AddUint64(&attempts, 1)
		return errors.New("fail")
	}

	p := NewPipeline(logging.NewLogger(logging.LevelError), PipelineOptions{
		BufferSize:       16,
		Workers:          2,
		EnableIdempotent: false,
		Retry:            RetryPolicy{MaxAttempts: 2, BaseDelay: time.Millisecond, MaxDelay: 3 * time.Millisecond, Jitter: false},
	}, consumer)
	_ = p.Start()

	pubCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Send 10 messages -> each should fail twice and go to DLQ
	for i := 0; i < 10; i++ {
		if err := p.Publish(pubCtx, &Message{ID: fmt.Sprintf("x-%d", i)}); err != nil {
			t.Fatalf("publish: %v", err)
		}
	}

	shutCtx, cancel2 := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel2()
	if err := p.Shutdown(shutCtx); err != nil {
		t.Fatalf("shutdown: %v", err)
	}

	// Each item attempted twice
	if attempts < 20 {
		t.Fatalf("attempts=%d want>=20", attempts)
	}
	if len(p.DeadLetters()) != 10 {
		t.Fatalf("dlq=%d want=10", len(p.DeadLetters()))
	}
	m := p.SnapshotMetrics()
	if m.DLQ != 10 || m.Failed != 10 || m.Retried < 10 {
		t.Fatalf("metrics: %+v", m)
	}
}

// Test 4: Graceful shutdown drains channels and no deadlock
func TestPipelineGracefulShutdown(t *testing.T) {
	consumer := func(ctx context.Context, msg *Message) error {
		// simulate some work
		time.Sleep(200 * time.Microsecond)
		return nil
	}
	p := NewPipeline(logging.NewLogger(logging.LevelError), PipelineOptions{
		BufferSize:       256,
		Workers:          4,
		EnableIdempotent: false,
		Retry:            RetryPolicy{MaxAttempts: 1},
	}, consumer)
	_ = p.Start()

	pubCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	for i := 0; i < 1000; i++ {
		if err := p.Publish(pubCtx, &Message{ID: fmt.Sprintf("g-%d", i)}); err != nil {
			t.Fatalf("publish: %v", err)
		}
	}

	// Now shutdown with deadline; should complete without hanging
	shutCtx, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel2()
	if err := p.Shutdown(shutCtx); err != nil {
		t.Fatalf("shutdown: %v", err)
	}

	m := p.SnapshotMetrics()
	if m.Succeeded != 1000 || m.InFlight != 0 {
		t.Fatalf("unexpected metrics after shutdown: %+v", m)
	}
}
