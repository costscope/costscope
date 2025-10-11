package api

import (
	"strings"
	"testing"
)

func TestDisplayAdditionalFeatures_showsAll(t *testing.T) {
	// preserve
	prevML := enhancedAPIML
	prevCache := enhancedAPICache
	prevCompression := enhancedAPICompression
	prevMetrics := enhancedAPIMetrics
	prevDocs := enhancedAPIDocumentation
	prevModels := enhancedAPIMLModels
	prevCacheSize := enhancedAPICacheSize
	defer func() {
		enhancedAPIML = prevML
		enhancedAPICache = prevCache
		enhancedAPICompression = prevCompression
		enhancedAPIMetrics = prevMetrics
		enhancedAPIDocumentation = prevDocs
		enhancedAPIMLModels = prevModels
		enhancedAPICacheSize = prevCacheSize
	}()

	enhancedAPIML = true
	enhancedAPIMLModels = 3
	enhancedAPICache = true
	enhancedAPICacheSize = "128MB"
	enhancedAPICompression = true
	enhancedAPIMetrics = true
	enhancedAPIDocumentation = true

	out := captureOutput(func() {
		displayAdditionalFeatures()
	})

	if !strings.Contains(out, "ML Models: 3 active models") {
		t.Fatalf("expected ML models line; got %q", out)
	}
	if !strings.Contains(out, "Caching: Redis (128MB)") && !strings.Contains(out, "Caching: Redis") {
		t.Fatalf("expected cache info; got %q", out)
	}
	if !strings.Contains(out, "Compression") {
		t.Fatalf("expected Compression line; got %q", out)
	}
	if !strings.Contains(out, "Metrics: Prometheus/Grafana") {
		t.Fatalf("expected Metrics line; got %q", out)
	}
	if !strings.Contains(out, "Documentation: Swagger/OpenAPI") {
		t.Fatalf("expected Documentation line; got %q", out)
	}
}

func TestInitializeComponents_lightPaths(t *testing.T) {
	prevAuth := enhancedAPIAdvancedAuth
	prevRBAC := enhancedAPIRBAC
	prevML := enhancedAPIML
	prevGraph := enhancedAPIGraphQL
	prevWS := enhancedAPIWebSocket
	prevStreaming := enhancedAPIStreaming
	prevCache := enhancedAPICache
	prevMetrics := enhancedAPIMetrics
	defer func() {
		enhancedAPIAdvancedAuth = prevAuth
		enhancedAPIRBAC = prevRBAC
		enhancedAPIML = prevML
		enhancedAPIGraphQL = prevGraph
		enhancedAPIWebSocket = prevWS
		enhancedAPIStreaming = prevStreaming
		enhancedAPICache = prevCache
		enhancedAPIMetrics = prevMetrics
	}()

	enhancedAPIAdvancedAuth = true
	enhancedAPIRBAC = true
	enhancedAPIML = true
	enhancedAPIGraphQL = true
	enhancedAPIWebSocket = true
	enhancedAPIStreaming = true
	enhancedAPICache = true
	enhancedAPIMetrics = true
	enhancedAPIMLModels = 1

	out := captureOutput(func() {
		_ = initializeComponents()
	})

	// Just ensure initialization messages for enabled features appear
	if !strings.Contains(out, "Initializing API server components") {
		t.Fatalf("expected init header; got %q", out)
	}
	if !strings.Contains(out, "Loading 1 machine learning models") && !strings.Contains(out, "Loading 1 machine") {
		t.Fatalf("expected ML loading; got %q", out)
	}
	if !strings.Contains(out, "Setting up JWT authentication") && !strings.Contains(out, "JWT authentication") {
		t.Fatalf("expected JWT setup message; got %q", out)
	}
	if !strings.Contains(out, "Setting up GraphQL schema") && !strings.Contains(out, "GraphQL") {
		t.Fatalf("expected GraphQL init; got %q", out)
	}
}

func TestSetupServerEndpoints_printsEndpoints(t *testing.T) {
	prevAnalytics := enhancedAPIAnalytics
	prevML := enhancedAPIML
	prevOpt := enhancedAPIOptimization
	prevForecast := enhancedAPIForecasting
	defer func() {
		enhancedAPIAnalytics = prevAnalytics
		enhancedAPIML = prevML
		enhancedAPIOptimization = prevOpt
		enhancedAPIForecasting = prevForecast
	}()

	enhancedAPIAnalytics = true
	enhancedAPIML = true
	enhancedAPIOptimization = true
	enhancedAPIForecasting = true

	out := captureOutput(func() {
		setupServerEndpoints()
	})

	if !strings.Contains(out, "/api/v2/analytics") {
		t.Fatalf("expected analytics endpoint; got %q", out)
	}
	if !strings.Contains(out, "/api/v2/ml") {
		t.Fatalf("expected ml endpoint; got %q", out)
	}
	if !strings.Contains(out, "/api/v2/optimization") {
		t.Fatalf("expected optimization endpoint; got %q", out)
	}
	if !strings.Contains(out, "/api/v2/forecasting") {
		t.Fatalf("expected forecasting endpoint; got %q", out)
	}
}

func TestDisplayServerConfig_includesHeader(t *testing.T) {
	out := captureOutput(func() {
		displayServerConfig()
	})
	if !strings.Contains(out, "Enhanced CostScope API Server") {
		t.Fatalf("expected header; got %q", out)
	}
}
