//go:build enterprise

package streaming

import (
	"context"
	crand "crypto/rand"
	"errors"
	"math"
	"sync"
	"sync/atomic"
	"time"

	"github.com/costscope/costscope/internal/core/logging"
	"github.com/costscope/costscope/internal/core/monitoring/telemetry"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

// Message represents a single unit of work flowing through the pipeline.
type Message struct {
	ID        string
	Key       string // used for partitioning or per-key semantics
	Data      []byte
	Attempt   int
	Enqueued  time.Time
	LastError error
}

// Consumer processes a message. Return nil on success; non-nil to trigger retry.
type Consumer func(ctx context.Context, msg *Message) error

// RetryPolicy controls retry behavior.
type RetryPolicy struct {
	MaxAttempts int           // total attempts including the first; 0/1 means no retry
	BaseDelay   time.Duration // base backoff delay (e.g., 10ms)
	MaxDelay    time.Duration // maximum backoff delay cap
	Jitter      bool          // add small jitter to spread retries
}

// PipelineOptions configures the stream pipeline.
type PipelineOptions struct {
	BufferSize       int  // bounded channel capacity; backpressure when full
	Workers          int  // number of concurrent consumers
	EnableIdempotent bool // drop duplicate IDs before processing
	MetricsInterval  time.Duration
	Retry            RetryPolicy
}

// Metrics exposes counters and gauges for tests/monitoring.
type Metrics struct {
	Produced     uint64
	Enqueued     uint64
	Consumed     uint64
	Succeeded    uint64
	Failed       uint64
	Retried      uint64
	DedupDropped uint64
	DLQ          uint64
	InFlight     int64 // gauge
	// Latency aggregates (nanoseconds) for simple checks
	TotalProcLatency int64
}

// Pipeline implements a bounded, worker-based streaming processor with backpressure, retries, idempotency and an in-memory dead-letter queue.
type Pipeline struct {
	opts PipelineOptions
	log  *logging.Logger
	cons Consumer

	inCh    chan *Message
	dlqMu   sync.Mutex
	dlq     []*Message
	started int32
	closed  int32

	ctx    context.Context // used for metrics and global lifecycle
	cancel context.CancelFunc
	wg     sync.WaitGroup

	metrics     Metrics
	metricsMu   sync.RWMutex
	onMetricsCb func(Metrics)

	// idempotency store (best-effort, in-memory)
	dedup sync.Map // map[string]struct{}
}

// NewPipeline creates a new Pipeline.
func NewPipeline(logger *logging.Logger, opts PipelineOptions, consumer Consumer) *Pipeline {
	if logger == nil {
		logger = logging.NewLogger(logging.LevelInfo)
	}
	if opts.Workers <= 0 {
		opts.Workers = 1
	}
	if opts.BufferSize <= 0 {
		opts.BufferSize = 1024
	}
	if opts.MetricsInterval <= 0 {
		opts.MetricsInterval = 2 * time.Second
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &Pipeline{
		opts:   opts,
		log:    logger,
		cons:   consumer,
		inCh:   make(chan *Message, opts.BufferSize),
		ctx:    ctx,
		cancel: cancel,
	}
}

// Note: metrics callback was unused across the codebase; setter removed to reduce surface.
// Tests access metrics via SnapshotMetrics(). Existing field is kept for potential future wiring.

// Start launches worker goroutines and a metrics reporter.
func (p *Pipeline) Start() error {
	if !atomic.CompareAndSwapInt32(&p.started, 0, 1) {
		return errors.New("pipeline already started")
	}
	// Workers
	for i := 0; i < p.opts.Workers; i++ {
		p.wg.Add(1)
		go p.worker()
	}
	// Metrics ticker (lifecycle controlled by ctx)
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		ticker := time.NewTicker(p.opts.MetricsInterval)
		defer ticker.Stop()
		for {
			select {
			case <-p.ctx.Done():
				return
			case <-ticker.C:
				p.emitMetrics()
			}
		}
	}()
	p.log.Info("streaming pipeline started")
	return nil
}

// Publish enqueues a message or blocks when backpressure applies. Context can cancel the attempt.
func (p *Pipeline) Publish(ctx context.Context, m *Message) error {
	if atomic.LoadInt32(&p.closed) == 1 {
		return errors.New("pipeline closed")
	}
	if m.Enqueued.IsZero() {
		m.Enqueued = time.Now()
	}
	atomic.AddUint64(&p.metrics.Produced, 1)
	for {
		// First, try fast non-blocking enqueue; if full, record backpressure metric once then block.
		select {
		case p.inCh <- m:
			atomic.AddUint64(&p.metrics.Enqueued, 1)
			atomic.AddInt64(&p.metrics.InFlight, 1)
			return nil
		default:
			// channel full -> backpressure event (count once per loop iteration before blocking)
			telemetry.StreamingBackpressure.Inc()
		}
		// Now block respecting contexts
		select {
		case <-p.ctx.Done():
			return p.ctx.Err()
		case <-ctx.Done():
			return ctx.Err()
		case p.inCh <- m:
			atomic.AddUint64(&p.metrics.Enqueued, 1)
			atomic.AddInt64(&p.metrics.InFlight, 1)
			return nil
		}
	}
}

// Shutdown gracefully stops intake and waits for workers to finish within the timeout context.
func (p *Pipeline) Shutdown(ctx context.Context) error {
	if !atomic.CompareAndSwapInt32(&p.closed, 0, 1) {
		// already closed
		return nil
	}
	// Drain-first shutdown: close intake, let workers finish remaining items
	close(p.inCh)
	ch := make(chan struct{})
	go func() {
		defer close(ch)
		// Wait for workers and metrics goroutine as well; we'll cancel metrics after workers
		p.wg.Wait()
	}()
	// Wait for workers to drain within timeout by monitoring InFlight and input closed
	drained := make(chan struct{})
	go func() {
		for atomic.LoadInt64(&p.metrics.InFlight) != 0 {
			time.Sleep(5 * time.Millisecond)
		}
		close(drained)
	}()
	select {
	case <-ctx.Done():
		// If caller's deadline exceeded, still cancel metrics and return
		p.cancel()
		<-ch
		return ctx.Err()
	case <-drained:
		// Now cancel metrics goroutine and wait all
		p.cancel()
		<-ch
		p.emitMetrics()
		p.log.Info("streaming pipeline stopped")
		return nil
	}
}

// DeadLetters returns a snapshot of the DLQ.
func (p *Pipeline) DeadLetters() []*Message {
	p.dlqMu.Lock()
	defer p.dlqMu.Unlock()
	out := make([]*Message, len(p.dlq))
	copy(out, p.dlq)
	return out
}

// SnapshotMetrics returns a copy of metrics.
func (p *Pipeline) SnapshotMetrics() Metrics {
	p.metricsMu.RLock()
	defer p.metricsMu.RUnlock()
	return p.metrics
}

func (p *Pipeline) worker() {
	defer p.wg.Done()
	for msg := range p.inCh {
		if msg == nil {
			continue
		}
		wctx, span := otel.Tracer("costscope.streaming").Start(p.ctx, "streaming.batch")
		// idempotent drop
		if p.opts.EnableIdempotent {
			if msg.ID != "" {
				if _, loaded := p.dedup.LoadOrStore(msg.ID, struct{}{}); loaded {
					atomic.AddUint64(&p.metrics.DedupDropped, 1)
					atomic.AddInt64(&p.metrics.InFlight, -1)
					span.SetAttributes(attribute.String("msg.id", truncateID(msg.ID)), attribute.Bool("dropped", true))
					span.End()
					continue
				}
			}
		}
		atomic.AddUint64(&p.metrics.Consumed, 1)
		start := time.Now()
		// process with retry
		if p.processWithRetry(msg) {
			atomic.AddUint64(&p.metrics.Succeeded, 1)
		} else {
			atomic.AddUint64(&p.metrics.Failed, 1)
			// add to DLQ
			p.dlqMu.Lock()
			p.dlq = append(p.dlq, msg)
			p.dlqMu.Unlock()
			atomic.AddUint64(&p.metrics.DLQ, 1)
		}
		// latency
		d := time.Since(start)
		if d < 0 {
			d = 0
		}
		atomic.AddInt64(&p.metrics.TotalProcLatency, int64(d))
		atomic.AddInt64(&p.metrics.InFlight, -1)
		span.SetAttributes(
			attribute.String("msg.id", truncateID(msg.ID)),
			attribute.Int("attempt", msg.Attempt),
			attribute.Bool("success", msg.LastError == nil),
			attribute.Int64("proc.nanos", d.Nanoseconds()),
		)
		span.End()
		_ = wctx
	}
}

// truncateID returns first 12 chars (or full if shorter) for span attribute brevity.
func truncateID(id string) string {
	if len(id) > 12 {
		return id[:12]
	}
	return id
}

func (p *Pipeline) processWithRetry(msg *Message) bool {
	attempts := 1
	if p.opts.Retry.MaxAttempts > 0 {
		attempts = p.opts.Retry.MaxAttempts
	}
	// Always attempt at least once
	for a := 1; a <= attempts; a++ {
		msg.Attempt = a
		if err := p.cons(p.ctx, msg); err == nil {
			return true
		} else {
			msg.LastError = err
		}
		if a < attempts {
			atomic.AddUint64(&p.metrics.Retried, 1)
			// backoff
			delay := p.backoffDelay(a)
			time.Sleep(delay)
		}
	}
	return false
}

func (p *Pipeline) backoffDelay(attempt int) time.Duration {
	base := p.opts.Retry.BaseDelay
	if base <= 0 {
		base = 10 * time.Millisecond
	}
	max := p.opts.Retry.MaxDelay
	if max <= 0 {
		max = 500 * time.Millisecond
	}
	// exponential: base * 2^(attempt-1)
	mult := math.Pow(2, float64(attempt-1))
	delay := time.Duration(float64(base) * mult)
	if delay > max {
		delay = max
	}
	if p.opts.Retry.Jitter {
		// +/- 20% jitter using crypto/rand
		var b [1]byte
		if _, err := crand.Read(b[:]); err == nil {
			f := float64(b[0]) / 255.0 // [0,1]
			j := 0.2 * (f*2 - 1)       // [-0.2, 0.2]
			delay = time.Duration(float64(delay) * (1 + j))
			if delay < 0 {
				delay = base
			}
		}
	}
	return delay
}

func (p *Pipeline) emitMetrics() {
	if p.onMetricsCb == nil {
		return
	}
	p.metricsMu.RLock()
	snapshot := p.metrics
	p.metricsMu.RUnlock()
	// deliver snapshot
	p.onMetricsCb(snapshot)
}
