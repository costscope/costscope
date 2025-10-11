package dashboard

import (
	"bytes"
	"os"
	"strings"
	"testing"
	"time"
)

// TestStartWithOptionsUpdatesConfigAndStarts validates that StartWithOptions applies
// provided options (including security) and transitions the manager to running state.
func TestStartWithOptionsUpdatesConfigAndStarts(t *testing.T) {
	dm := NewDashboardManager()

	opts := StartOptions{
		Port:       9090,
		Theme:      "enterprise-dark",
		AutoOpen:   false,
		Features:   []string{"real-time-updates", "drill-down"},
		Auth:       true,
		AllowedIPs: []string{"10.0.0.0/24", "192.168.1.10"},
	}

	start := time.Now()
	if err := dm.StartWithOptions(opts); err != nil {
		t.Fatalf("StartWithOptions returned error: %v", err)
	}
	if dur := time.Since(start); dur > 3*time.Second { // safety guard if future sleep grows unexpectedly
		t.Fatalf("dashboard start took unexpectedly long: %v", dur)
	}

	if dm.config.Port != opts.Port {
		t.Errorf("expected port %d got %d", opts.Port, dm.config.Port)
	}
	if dm.config.Theme != opts.Theme {
		t.Errorf("expected theme %s got %s", opts.Theme, dm.config.Theme)
	}
	if dm.config.AutoOpen != opts.AutoOpen {
		t.Errorf("expected AutoOpen %v got %v", opts.AutoOpen, dm.config.AutoOpen)
	}
	if dm.config.Security == nil || !dm.config.Security.Enabled {
		t.Fatalf("expected security enabled")
	}
	if len(dm.config.Security.AllowedIPs) != len(opts.AllowedIPs) {
		t.Errorf("expected %d allowed IPs got %d", len(opts.AllowedIPs), len(dm.config.Security.AllowedIPs))
	}
	if !dm.isRunning {
		t.Fatalf("expected dashboard to be running after start")
	}
}

// TestStartWithOptionsPreservesFeaturesWhenEmpty ensures providing an empty Features slice
// does not overwrite the default feature list.
func TestStartWithOptionsPreservesFeaturesWhenEmpty(t *testing.T) {
	dm := NewDashboardManager()
	defaultFeatures := append([]string(nil), dm.config.Features...) // copy

	opts := StartOptions{Port: 8081, Theme: "modern", AutoOpen: true, Features: nil}
	if err := dm.StartWithOptions(opts); err != nil {
		t.Fatalf("StartWithOptions error: %v", err)
	}

	if len(dm.config.Features) != len(defaultFeatures) {
		t.Fatalf("expected default features preserved; got %d want %d", len(dm.config.Features), len(defaultFeatures))
	}
	// spot check one known feature
	found := false
	for _, f := range dm.config.Features {
		if f == "real-time-updates" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected to find 'real-time-updates' in features after start")
	}
}

// TestStatusVerboseOutputIncludesMetrics captures verbose status output and asserts
// that performance metrics section is printed when dashboard running.
func TestStatusVerboseOutputIncludesMetrics(t *testing.T) {
	dm := NewDashboardManager()
	if err := dm.StartWithOptions(StartOptions{Port: 8082, Theme: "modern", AutoOpen: false}); err != nil {
		t.Fatalf("start error: %v", err)
	}

	// capture stdout
	var buf bytes.Buffer
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := dm.Status(true)
	if cerr := w.Close(); cerr != nil {
		t.Fatalf("close pipe writer: %v", cerr)
	}
	os.Stdout = old
	if err != nil {
		t.Fatalf("Status returned error: %v", err)
	}
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf("reading captured output: %v", err)
	}
	out := buf.String()

	// Basic assertions
	if !strings.Contains(out, "Dashboard Status") {
		t.Errorf("expected 'Dashboard Status' header in output")
	}
	if !strings.Contains(out, "Performance Metrics") {
		t.Errorf("expected verbose performance metrics section in output")
	}
	if !strings.Contains(out, "Load Time:") {
		t.Errorf("expected load time metric in output")
	}
}
