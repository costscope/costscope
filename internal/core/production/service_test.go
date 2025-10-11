package production

import (
	"context"
	"testing"
	"time"

	"local/costscope/internal/core/logging"

	"github.com/stretchr/testify/assert"
)

// Test_GetSystemStatus_Smoke reuses existing fakes in service_basic_test.go to
// exercise GetSystemStatus without external dependencies.
func Test_GetSystemStatus_Smoke(t *testing.T) {
	logger := logging.NewLogger(logging.LevelError)
	svc := &BasicProductionService{logger: logger, cache: make(map[string]interface{})}

	// reuse fakes defined in service_basic_test.go
	svc.metricsCollector = fakeMetricsCollector{}
	svc.optimizationEngine = fakeOptimizationEngine{}
	svc.deploymentAssessor = fakeDeploymentAssessor{}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	metrics, err := svc.GetSystemStatus(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, metrics)
	// basic sanity checks
	assert.GreaterOrEqual(t, metrics.ReadinessScore, 0)
}
