package framework

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

// ---- Container extended tests ----

func TestContainer_GetAsAndUtilities(t *testing.T) {
	c := NewContainer()
	if h := c.Health(); h != HealthStatusEmpty { // empty path
		t.Fatalf("expected empty health, got %s", h)
	}

	c.RegisterSingleton("str", "value")
	// Has/List
	if !c.Has("str") {
		t.Fatalf("expected Has(str)=true")
	}
	if len(c.List()) != 1 {
		t.Fatalf("expected 1 service")
	}

	// GetAs pointer requirement
	var notPtr string
	if err := c.GetAs("str", notPtr); err == nil { // must be pointer
		t.Fatalf("expected error for non-pointer target")
	}
	var out string
	if err := c.GetAs("str", &out); err != nil || out != "value" {
		t.Fatalf("GetAs failed: %v out=%q", err, out)
	}

	// GetAs type mismatch
	var outInt int
	if err := c.GetAs("str", &outInt); err == nil {
		t.Fatalf("expected type mismatch error")
	}

	// Unregister + Clear
	c.Unregister("str")
	if c.Has("str") {
		t.Fatalf("service still present after Unregister")
	}
	c.RegisterSingleton("a", 1)
	c.RegisterSingleton("b", 2)
	c.Clear()
	if len(c.List()) != 0 {
		t.Fatalf("expected empty after Clear")
	}
}

// ---- Container Inject tag variants ----
type injDep struct{ V int }
type injTarget struct { // explicit name via field name when tag is true
	Dep  *injDep `inject:"true"`
	Also *injDep `inject:"Also"`
}

func TestContainer_InjectTagVariants(t *testing.T) {
	c := NewContainer()
	d := &injDep{V: 5}
	c.RegisterSingleton("Dep", d)
	c.RegisterSingleton("Also", d)
	var tgt injTarget
	if err := c.Inject(&tgt); err != nil {
		t.Fatalf("inject failed: %v", err)
	}
	if tgt.Dep == nil || tgt.Also == nil {
		t.Fatalf("expected both dependencies injected")
	}
}

// ---- EventBus extended tests ----

func TestEventBus_SubscribeOnceFilterSyncAndUnsubscribe(t *testing.T) {
	// Verify emit is no-op when not started
	ebIdle := NewEventBus()
	ebIdle.Emit("idle.event", nil) // should not panic

	eb := NewEventBus()
	ctx := context.Background()
	if err := eb.Start(ctx); err != nil {
		t.Fatalf("start err: %v", err)
	}
	t.Cleanup(func() { _ = eb.Stop(context.Background()) })

	// SubscribeOnce — wait for the handler via channel instead of sleeping
	onceCount := int32(0)
	doneOnce := make(chan struct{})
	eb.SubscribeOnce("once", func(Event) {
		atomic.AddInt32(&onceCount, 1)
		select {
		case <-doneOnce:
		default:
			close(doneOnce)
		}
	})
	eb.Emit("once", nil)
	eb.Emit("once", nil)
	select {
	case <-doneOnce:
		// handler fired once
	case <-time.After(500 * time.Millisecond):
		t.Fatalf("SubscribeOnce did not fire within timeout; count=%d", atomic.LoadInt32(&onceCount))
	}
	if atomic.LoadInt32(&onceCount) != 1 {
		t.Fatalf("SubscribeOnce delivered %d events", onceCount)
	}

	// Filter: allow only events with Data["ok"]==true
	filterCount := int32(0)
	// Filter: allow only events with Data["ok"]==true. Wait for a notification from handler.
	filterDone := make(chan struct{})
	eb.SubscribeWithOptions("flt", func(Event) {
		atomic.AddInt32(&filterCount, 1)
		select {
		case <-filterDone:
		default:
			close(filterDone)
		}
	}, SubscriptionOptions{Filter: func(e Event) bool { v, _ := e.Data["ok"].(bool); return v }})
	eb.Emit("flt", map[string]interface{}{"ok": false})
	eb.Emit("flt", map[string]interface{}{"ok": true})
	select {
	case <-filterDone:
		// ok
	case <-time.After(500 * time.Millisecond):
		t.Fatalf("filter handler did not fire within timeout; count=%d", atomic.LoadInt32(&filterCount))
	}
	if atomic.LoadInt32(&filterCount) != 1 {
		t.Fatalf("filter mismatch count=%d", filterCount)
	}

	// EmitSync immediate processing
	syncFired := int32(0)
	eb.Subscribe("sync", func(Event) { atomic.AddInt32(&syncFired, 1) })
	eb.EmitSync("sync", nil)
	if atomic.LoadInt32(&syncFired) != 1 {
		t.Fatalf("EmitSync not immediate")
	}

	// Unsubscribe: ensure count does not increase after unsubscribe by polling for a short timeout
	id := eb.Subscribe("gone", func(Event) { atomic.AddInt32(&syncFired, 1) })
	eb.Unsubscribe(id)
	before := atomic.LoadInt32(&syncFired)
	eb.Emit("gone", nil)
	// Poll for a short window to assert no additional increments
	deadline := time.Now().Add(300 * time.Millisecond)
	for time.Now().Before(deadline) {
		if atomic.LoadInt32(&syncFired) != before {
			t.Fatalf("unsubscribe ineffective: before=%d now=%d", before, atomic.LoadInt32(&syncFired))
		}
		time.Sleep(10 * time.Millisecond)
	}

	// Degraded health path: block workers so queue backs up
	blocker := make(chan struct{})
	eb.Subscribe("block", func(Event) { <-blocker })
	for i := 0; i < 900; i++ {
		eb.Emit("block", nil)
	}
	// Give workers time to pick up first events and block
	time.Sleep(30 * time.Millisecond)
	if h := eb.Health(); h != HealthStatusDegraded {
		t.Fatalf("expected degraded, got %s", h)
	}
	close(blocker)
	time.Sleep(20 * time.Millisecond)
}

// ---- Lifecycle tests ----

type testComponent struct {
	name     string
	health   string
	startErr error
	stopErr  error
	started  bool
}

func (c *testComponent) Name() string { return c.name }
func (c *testComponent) Start(ctx context.Context) error {
	if c.startErr != nil {
		return c.startErr
	}
	c.started = true
	return nil
}
func (c *testComponent) Stop(ctx context.Context) error {
	if c.stopErr != nil {
		return c.stopErr
	}
	c.started = false
	return nil
}
func (c *testComponent) Health() string { return c.health }

func TestLifecycle_Paths(t *testing.T) {
	lc := NewLifecycle()
	a := &testComponent{name: "a", health: HealthStatusHealthy}
	b := &testComponent{name: "b", health: HealthStatusHealthy}
	lc.Register("b", b, 5)
	lc.Register("a", a, 1) // higher priority (lower number)
	if err := lc.Start(context.Background()); err != nil {
		t.Fatalf("start err: %v", err)
	}
	if !lc.IsRunning() || lc.GetPhase() != PhaseRunning {
		t.Fatalf("expected running phase")
	}
	stats := lc.Stats()
	if stats["started_components"].(int) != 2 {
		t.Fatalf("stats mismatch %#v", stats)
	}
	// Start again idempotent
	if err := lc.Start(context.Background()); err != nil {
		t.Fatalf("second start err: %v", err)
	}
	// Degrade path: change component health
	b.health = HealthStatusDegraded
	if h := lc.Health(); h != HealthStatusDegraded {
		t.Fatalf("expected degraded health")
	}
	// GetComponent success + error
	if _, err := lc.GetComponent("a"); err != nil {
		t.Fatalf("GetComponent a err: %v", err)
	}
	if _, err := lc.GetComponent("missing"); err == nil {
		t.Fatalf("expected missing component error")
	}
	if err := lc.Stop(context.Background()); err != nil {
		t.Fatalf("stop err: %v", err)
	}
	if lc.Health() != HealthStatusStopped {
		t.Fatalf("expected stopped health")
	}
	// Stop again no-op
	if err := lc.Stop(context.Background()); err != nil {
		t.Fatalf("second stop err: %v", err)
	}
}

func TestLifecycle_StartErrorPropagates(t *testing.T) {
	lc := NewLifecycle()
	failing := &testComponent{name: "fail", startErr: errors.New("boom"), health: HealthStatusHealthy}
	lc.Register("fail", failing, 1)
	if err := lc.Start(context.Background()); err == nil {
		t.Fatalf("expected start error")
	}
	if lc.Health() != HealthStatusStopped {
		t.Fatalf("health after start failure should be stopped")
	}
}

// ---- PluginLoader edge tests ----
func TestPluginLoader_Edge(t *testing.T) {
	pl := NewPluginLoader()
	c := NewContainer()
	ctx := context.Background()
	if err := pl.LoadPlugins(ctx, c); err != nil {
		t.Fatalf("load err: %v", err)
	}
	if !pl.IsPluginLoaded("analytics") || !pl.IsPluginStarted("analytics") {
		t.Fatalf("expected analytics loaded+started")
	}
	if _, err := pl.GetPlugin("missing"); err == nil {
		t.Fatalf("expected missing plugin error")
	}
	// Unload single plugin
	if err := pl.UnloadPlugin(ctx, "analytics"); err != nil {
		t.Fatalf("unload analytics err: %v", err)
	}
	// Unload all
	if err := pl.UnloadPlugins(ctx); err != nil {
		t.Fatalf("unload all err: %v", err)
	}
	if h := pl.Health(); h != HealthStatusEmpty {
		t.Fatalf("expected empty health after unload, got %s", h)
	}
}
