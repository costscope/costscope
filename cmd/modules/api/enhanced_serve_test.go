package api

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"local/costscope/internal/core/logging"
)

// captureStdout captures stdout while fn runs and returns the output.
func captureStdout(fn func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	outC := make(chan string)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		outC <- buf.String()
	}()
	fn()
	_ = w.Close()
	os.Stdout = old
	return <-outC
}

func TestEnhancedServe_DisplayAndInit(t *testing.T) {
	// enable some feature flags to hit branches in display helpers
	enhancedAPIAnalytics = true
	enhancedAPIML = true
	enhancedAPIMLModels = 2
	enhancedAPIDocumentation = true
	enhancedAPIGraphQL = true

	// Capture stdout to avoid test noise; ensure functions run without panic
	_ = captureStdout(func() {
		displayServerConfig()
		displayCommunicationProtocols()
		displaySecurityFeatures()
		displayAdditionalFeatures()
		displayServerURLs()
	})

	// initializeComponents should return nil
	if err := initializeComponents(); err != nil {
		t.Fatalf("initializeComponents failed: %v", err)
	}

	// startEnhancedServer should not error
	if err := startEnhancedServer(time.Now(), logging.NewLogger(logging.LevelInfo)); err != nil {
		t.Fatalf("startEnhancedServer returned error: %v", err)
	}
}

func TestDisplayServerURLsAndCoreCapabilities(t *testing.T) {
	// Ensure environment affects base URL selection
	t.Setenv("DOCS_BASE_URL", "https://example.com/")

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Exercise functions that only print
	enhancedAPIAnalytics = true
	enhancedAPIDocumentation = true
	enhancedAPIGraphQL = true
	enhancedAPIMetrics = true
	enhancedAPICache = true
	enhancedAPIML = true
	enhancedAPIMLModels = 2

	displayCoreCapabilities()
	displayServerURLs()

	// Close and read
	if err := w.Close(); err != nil {
		t.Fatalf("pipe close failed: %v", err)
	}
	// read captured output
	bs, _ := io.ReadAll(r)
	os.Stdout = old

	out := string(bs)
	if !strings.Contains(out, "Server:") {
		t.Fatalf("expected Server URL in output, got: %s", out)
	}
	if !strings.Contains(out, "Analytics API") && !strings.Contains(out, "Machine Learning API") {
		t.Fatalf("expected capability lines, got: %s", out)
	}

	// quick startup summary invocation path (uses time since)
	displayStartupSummary(time.Now().Add(-time.Second))
}
