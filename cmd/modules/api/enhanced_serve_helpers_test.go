package api

import (
	"io"
	"os"
	"strings"
	"testing"
	"time"
)

// captureOutput runs f while capturing stdout and returns the captured string.
func captureOutput(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	_ = w.Close()
	out, _ := io.ReadAll(r)
	os.Stdout = old
	return string(out)
}

func TestBuildEnhancedAPICommand_basic(t *testing.T) {
	cmd := BuildEnhancedAPICommand()
	if cmd == nil {
		t.Fatal("expected non-nil command")
	}
	if cmd.Use != "enhanced" {
		t.Fatalf("unexpected command use: %s", cmd.Use)
	}
	if cmd.RunE == nil {
		t.Fatalf("expected RunE to be set")
	}
}

func TestDisplayCommunicationAndCoreCapabilities(t *testing.T) {
	// preserve original values
	prevGraphQL := enhancedAPIGraphQL
	prevWebSocket := enhancedAPIWebSocket
	prevStreaming := enhancedAPIStreaming
	prevAnalytics := enhancedAPIAnalytics
	prevML := enhancedAPIML
	defer func() {
		enhancedAPIGraphQL = prevGraphQL
		enhancedAPIWebSocket = prevWebSocket
		enhancedAPIStreaming = prevStreaming
		enhancedAPIAnalytics = prevAnalytics
		enhancedAPIML = prevML
	}()

	enhancedAPIGraphQL = true
	enhancedAPIWebSocket = true
	enhancedAPIStreaming = true
	enhancedAPIAnalytics = true
	enhancedAPIML = true

	out := captureOutput(func() {
		displayCommunicationProtocols()
		displayCoreCapabilities()
	})

	if !strings.Contains(out, "GraphQL API") {
		t.Fatalf("expected GraphQL in output; got %q", out)
	}
	if !strings.Contains(out, "WebSocket Streaming") {
		t.Fatalf("expected WebSocket in output; got %q", out)
	}
	if !strings.Contains(out, "Data Streaming") {
		t.Fatalf("expected Data Streaming in output; got %q", out)
	}
	if !strings.Contains(out, "Analytics API") {
		t.Fatalf("expected Analytics API in output; got %q", out)
	}
	if !strings.Contains(out, "Machine Learning API") {
		t.Fatalf("expected Machine Learning API in output; got %q", out)
	}
}

func TestDisplaySecurityFeatures_showsEnabled(t *testing.T) {
	prevAuth := enhancedAPIAdvancedAuth
	prevRBAC := enhancedAPIRBAC
	prevAudit := enhancedAPIAuditLog
	defer func() {
		enhancedAPIAdvancedAuth = prevAuth
		enhancedAPIRBAC = prevRBAC
		enhancedAPIAuditLog = prevAudit
	}()

	enhancedAPIAdvancedAuth = true
	enhancedAPIRBAC = true
	enhancedAPIAuditLog = true

	out := captureOutput(func() {
		displaySecurityFeatures()
	})

	if !strings.Contains(out, "Advanced Authentication") {
		t.Fatalf("expected Advanced Authentication; got %q", out)
	}
	if !strings.Contains(out, "Role-Based Access Control") {
		t.Fatalf("expected RBAC; got %q", out)
	}
	if !strings.Contains(out, "Audit Logging") {
		t.Fatalf("expected Audit Logging; got %q", out)
	}
}

func TestDisplayStartupSummary_basic(t *testing.T) {
	prevWorkers := enhancedAPIWorkers
	prevCache := enhancedAPICache
	prevCompression := enhancedAPICompression
	prevRBAC := enhancedAPIRBAC
	prevAuth := enhancedAPIAdvancedAuth
	defer func() {
		enhancedAPIWorkers = prevWorkers
		enhancedAPICache = prevCache
		enhancedAPICompression = prevCompression
		enhancedAPIRBAC = prevRBAC
		enhancedAPIAdvancedAuth = prevAuth
	}()

	enhancedAPIWorkers = 2
	enhancedAPICache = true
	enhancedAPICacheSize = "128MB"
	enhancedAPICompression = true
	enhancedAPIRBAC = false
	enhancedAPIAdvancedAuth = false

	out := captureOutput(func() {
		// pass a start time ~1s ago to produce a readable startup time
		displayStartupSummary(time.Now().Add(-1 * time.Second))
	})

	if !strings.Contains(out, "Server Configuration Summary") {
		t.Fatalf("expected summary header; got %q", out)
	}
	if !strings.Contains(out, "Workers: 2") {
		t.Fatalf("expected workers line; got %q", out)
	}
	if !strings.Contains(out, "Cache: Redis (128MB)") {
		t.Fatalf("expected cache info; got %q", out)
	}
	if !strings.Contains(out, "Security: Standard") {
		t.Fatalf("expected Security: Standard; got %q", out)
	}
}
