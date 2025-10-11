package api

import (
	"testing"
	"time"
)

func TestDisplayFunctions_NoPanic(t *testing.T) {
	// Toggle several globals to hit conditional branches
	enhancedAPIAnalytics = true
	enhancedAPIML = true
	enhancedAPIOptimization = true
	enhancedAPIForecasting = true
	enhancedAPIRealtime = true
	enhancedAPIRBAC = true
	enhancedAPIAdvancedAuth = true
	enhancedAPIAuditLog = true
	enhancedAPICache = true
	enhancedAPICompression = true
	enhancedAPIMetrics = true
	enhancedAPIDocumentation = true
	enhancedAPIMLModels = 2
	enhancedAPIWorkers = 5

	// Call display helpers which print to stdout; ensure they run without panic
	displayCoreCapabilities()
	displayCommunicationProtocols()
	displaySecurityFeatures()
	displayAdditionalFeatures()
	setupServerEndpoints()
	displayServerURLs()
	// Provide a past start time to make startup summary print non-zero durations
	displayStartupSummary(time.Now().Add(-1500 * time.Millisecond))
}
