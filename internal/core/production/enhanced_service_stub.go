//go:build !enterprise

package production

import (
	"context"

	"local/costscope/internal/core/enterprise"
	"local/costscope/internal/core/integration"
	"local/costscope/internal/core/logging"
)

// Intentional enterprise stub for EnhancedProductionService.
// Community builds (without -tags enterprise) compile this file to preserve the
// public API while clearly signaling that enhanced production features are
// disabled. Enable the full implementation by rebuilding with:
//
//	go build -tags enterprise ./...
//
// Each method emits enterprise metrics (allowed=false) and returns a disabled
// sentinel error (enterprise.ErrEnterpriseFeatureDisabled wrapped with feature key).
// Do NOT add real logic here.
type EnhancedProductionService struct{}

// NewEnhancedProductionService keeps constructor parity with enterprise build.
// Intentional stub (enterprise gating): constructor returns disabled stub in community build.
func NewEnhancedProductionService(_ *BasicProductionService, _ integration.IntegrationService, _ *logging.Logger) *EnhancedProductionService {
	return &EnhancedProductionService{}
}

const enhancedProductionFeature = "production_enhanced"

// disabled helper centralizes metrics + error for stub calls.
func disabled() error {
	enterprise.ObserveInvocation(enhancedProductionFeature, false)
	enterprise.ObserveError(enhancedProductionFeature, "disabled")
	return enterprise.DisabledError(enhancedProductionFeature)
}

// Intentional stub (enterprise gating): always returns enterprise disabled error.
func (e *EnhancedProductionService) GetSystemStatusWithIntegrations(context.Context) (*EnhancedProductionSystemMetrics, error) {
	return nil, disabled()
}

// Intentional stub (enterprise gating): always returns enterprise disabled error.
func (e *EnhancedProductionService) RunIntegratedOptimization(context.Context, *EnhancedOptimizationOptions) (*IntegratedOptimizationReport, error) {
	return nil, disabled()
}

// Intentional stub (enterprise gating): always returns enterprise disabled error.
func (e *EnhancedProductionService) AssessIntegratedDeploymentReadiness(context.Context, string, *IntegratedDeploymentOptions) (*IntegratedDeploymentReadinessAssessment, error) {
	return nil, disabled()
}

// Intentional stub (enterprise gating): always returns enterprise disabled error.
func (e *EnhancedProductionService) GenerateEnhancedExecutiveReport(context.Context, *EnhancedReportOptions) (*EnhancedExecutiveReport, error) {
	return nil, disabled()
}
