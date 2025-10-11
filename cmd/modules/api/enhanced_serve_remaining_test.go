package api

import (
	"os"
	"strings"
	"testing"
	"time"
)

func TestDisplayServerURLs_WithEnvAndFlags(t *testing.T) {
	// Save & restore package vars
	prevDoc := enhancedAPIDocumentation
	prevGraphQL := enhancedAPIGraphQL
	prevMetrics := enhancedAPIMetrics
	defer func() {
		enhancedAPIDocumentation = prevDoc
		enhancedAPIGraphQL = prevGraphQL
		enhancedAPIMetrics = prevMetrics
	}()

	if err := os.Setenv("DOCS_BASE_URL", "https://example.com/base/"); err != nil {
		t.Fatalf("Setenv failed: %v", err)
	}
	defer func() {
		if err := os.Unsetenv("DOCS_BASE_URL"); err != nil {
			t.Fatalf("Unsetenv failed: %v", err)
		}
	}()

	enhancedAPIDocumentation = true
	enhancedAPIGraphQL = true
	enhancedAPIMetrics = true

	out := captureOutput(func() { displayServerURLs() })
	if !strings.Contains(out, "https://example.com/base/docs") {
		t.Fatalf("expected docs url in output, got: %s", out)
	}
	if !strings.Contains(out, "graphql/playground") {
		t.Fatalf("expected graphql playground in output, got: %s", out)
	}
	if !strings.Contains(out, "/metrics") {
		t.Fatalf("expected metrics url in output, got: %s", out)
	}
}

func TestDisplayServerURLs_DefaultBase(t *testing.T) {
	prevDoc := enhancedAPIDocumentation
	prevGraphQL := enhancedAPIGraphQL
	prevMetrics := enhancedAPIMetrics
	defer func() {
		enhancedAPIDocumentation = prevDoc
		enhancedAPIGraphQL = prevGraphQL
		enhancedAPIMetrics = prevMetrics
	}()

	if err := os.Unsetenv("DOCS_BASE_URL"); err != nil {
		t.Fatalf("Unsetenv failed: %v", err)
	}
	enhancedAPIDocumentation = false
	enhancedAPIGraphQL = false
	enhancedAPIMetrics = false

	out := captureOutput(func() { displayServerURLs() })
	if !strings.Contains(out, "http://localhost:8080") {
		t.Fatalf("expected default base url, got: %s", out)
	}
}

func TestInitializeComponents_MultiFlags(t *testing.T) {
	// Save & restore many package vars
	prevAdvanced := enhancedAPIAdvancedAuth
	prevRBAC := enhancedAPIRBAC
	prevML := enhancedAPIML
	prevMLModels := enhancedAPIMLModels
	prevGraphQL := enhancedAPIGraphQL
	prevWebSocket := enhancedAPIWebSocket
	prevStreaming := enhancedAPIStreaming
	prevCache := enhancedAPICache
	prevMetrics := enhancedAPIMetrics
	defer func() {
		enhancedAPIAdvancedAuth = prevAdvanced
		enhancedAPIRBAC = prevRBAC
		enhancedAPIML = prevML
		enhancedAPIMLModels = prevMLModels
		enhancedAPIGraphQL = prevGraphQL
		enhancedAPIWebSocket = prevWebSocket
		enhancedAPIStreaming = prevStreaming
		enhancedAPICache = prevCache
		enhancedAPIMetrics = prevMetrics
	}()

	enhancedAPIAdvancedAuth = true
	enhancedAPIRBAC = true
	enhancedAPIML = true
	enhancedAPIMLModels = 2
	enhancedAPIGraphQL = true
	enhancedAPIWebSocket = true
	enhancedAPIStreaming = true
	enhancedAPICache = true
	enhancedAPIMetrics = true

	out := captureOutput(func() {
		if err := initializeComponents(); err != nil {
			t.Fatalf("initializeComponents returned error: %v", err)
		}
	})

	if !strings.Contains(out, "Setting up JWT authentication") && !strings.Contains(out, "Setting up JWT authentication...") {
		// some variations of spacing; check for presence of 'authentication'
		if !strings.Contains(out, "authentication") {
			t.Fatalf("expected authentication setup text, got: %s", out)
		}
	}
	if !strings.Contains(out, "Loading 2 machine learning models") && !strings.Contains(out, "Loading 2 machine learning models...") {
		if !strings.Contains(out, "Loading 2") {
			t.Fatalf("expected ML loading text, got: %s", out)
		}
	}
}

func TestDisplaySecurityFeatures_Various(t *testing.T) {
	prevAdv := enhancedAPIAdvancedAuth
	prevRBAC := enhancedAPIRBAC
	prevAudit := enhancedAPIAuditLog
	defer func() {
		enhancedAPIAdvancedAuth = prevAdv
		enhancedAPIRBAC = prevRBAC
		enhancedAPIAuditLog = prevAudit
	}()

	enhancedAPIAdvancedAuth = true
	enhancedAPIRBAC = true
	enhancedAPIAuditLog = true

	out := captureOutput(func() { displaySecurityFeatures() })
	if !strings.Contains(out, "Advanced Authentication") {
		t.Fatalf("expected Advanced Authentication, got: %s", out)
	}
	if !strings.Contains(out, "Role-Based Access Control") {
		t.Fatalf("expected RBAC text, got: %s", out)
	}
	if !strings.Contains(out, "Audit Logging") {
		t.Fatalf("expected Audit Logging text, got: %s", out)
	}
}

func TestDisplayStartupSummary_SecurityModes(t *testing.T) {
	prevWorkers := enhancedAPIWorkers
	prevCache := enhancedAPICache
	prevCompression := enhancedAPICompression
	prevRBAC := enhancedAPIRBAC
	prevAuth := enhancedAPIAdvancedAuth
	prevGraphQL := enhancedAPIGraphQL
	prevWebSocket := enhancedAPIWebSocket
	defer func() {
		enhancedAPIWorkers = prevWorkers
		enhancedAPICache = prevCache
		enhancedAPICompression = prevCompression
		enhancedAPIRBAC = prevRBAC
		enhancedAPIAdvancedAuth = prevAuth
		enhancedAPIGraphQL = prevGraphQL
		enhancedAPIWebSocket = prevWebSocket
	}()

	enhancedAPIWorkers = 8
	enhancedAPICache = true
	enhancedAPICompression = true
	enhancedAPIRBAC = true
	enhancedAPIAdvancedAuth = true
	enhancedAPIGraphQL = true
	enhancedAPIWebSocket = true

	out := captureOutput(func() { displayStartupSummary(time.Now().Add(-1500 * time.Millisecond)) })
	if !strings.Contains(out, "Workers: 8") && !strings.Contains(out, "Workers: 8\n") {
		if !strings.Contains(out, "Workers: 8") {
			t.Fatalf("expected workers count in summary, got: %s", out)
		}
	}
	if !strings.Contains(out, "Enterprise (RBAC + OAuth)") {
		t.Fatalf("expected Enterprise security string, got: %s", out)
	}
	if !strings.Contains(out, "+ GraphQL") {
		t.Fatalf("expected GraphQL in protocols line, got: %s", out)
	}
}

func TestStartEnhancedServer_PrintsExpectedLines(t *testing.T) {
	prevAnalytics := enhancedAPIAnalytics
	prevML := enhancedAPIML
	defer func() {
		enhancedAPIAnalytics = prevAnalytics
		enhancedAPIML = prevML
	}()

	enhancedAPIAnalytics = true
	enhancedAPIML = true

	out := captureOutput(func() {
		if err := startEnhancedServer(time.Now().Add(-2*time.Second), nil); err != nil {
			t.Fatalf("startEnhancedServer returned error: %v", err)
		}
	})

	if !strings.Contains(out, "Enhanced API server is ready for requests") {
		t.Fatalf("expected ready message, got: %s", out)
	}
	if !strings.Contains(out, "Configuring API endpoints") {
		t.Fatalf("expected endpoints configuration text, got: %s", out)
	}
}
