//go:build !experimental
// +build !experimental

package analytics_advanced

// Disabled (non-experimental) stub.
// Intentional stub (experimental gating): provides a compile-time compatible
// implementation of AdvancedAnalyticsService when the "experimental" build tag
// is NOT supplied. All methods emit disabled metrics and return ErrDisabled.
// The richer implementation lives in build-tagged files guarded by
// //go:build experimental. This keeps community builds lean while preserving a
// stable internal API surface. When promoting features to GA this file and the
// gating may be removed.

import (
	"errors"

	"github.com/costscope/costscope/internal/core/enterprise"
)

// We duplicate the interface + types here because the canonical versions are
// only compiled in experimental builds. Keep fields minimal to satisfy callers.

type AdvancedAnalyticsService interface {
	RunMLForecast(request *ForecastRequest) (*ForecastResult, error)
	DetectAnomalies(request *AnomalyDetectionRequest) (*AnomalyDetectionResult, error)
	RunAdvancedOptimization(request *OptimizationRequest) (*OptimizationResult, error)
	TrainCustomModel(request *ModelTrainingRequest) (*ModelTrainingResult, error)
	StartStreamProcessing(request *StreamingRequest) (*StreamingResult, error)
	RunCustomAnalytics(request *CustomAnalyticsRequest) (*CustomAnalyticsResult, error)
}

type ForecastRequest struct{ Days int }
type ForecastResult struct{}

type AnomalyDetectionRequest struct{}
type AnomalyDetectionResult struct{}

type OptimizationRequest struct{}
type OptimizationResult struct{}

type ModelTrainingRequest struct{}
type ModelTrainingResult struct{}

type StreamingRequest struct{}
type StreamingResult struct{}

type CustomAnalyticsRequest struct{}
type CustomAnalyticsResult struct{}

// ErrDisabled is a sentinel error returned by all stub methods so tests and
// callers can use errors.Is for reliable detection.
var ErrDisabled = errors.New("advanced analytics disabled (build without -tags experimental)")

type disabledService struct{}

func NewAdvancedAnalyticsService() AdvancedAnalyticsService { return &disabledService{} }

func (d *disabledService) disabledErr() error {
	enterprise.ObserveInvocation("analytics_advanced", false)
	enterprise.ObserveError("analytics_advanced", "disabled")
	return ErrDisabled
}

func (d *disabledService) RunMLForecast(*ForecastRequest) (*ForecastResult, error) {
	return nil, d.disabledErr()
}
func (d *disabledService) DetectAnomalies(*AnomalyDetectionRequest) (*AnomalyDetectionResult, error) {
	return nil, d.disabledErr()
}
func (d *disabledService) RunAdvancedOptimization(*OptimizationRequest) (*OptimizationResult, error) {
	return nil, d.disabledErr()
}
func (d *disabledService) TrainCustomModel(*ModelTrainingRequest) (*ModelTrainingResult, error) {
	return nil, d.disabledErr()
}
func (d *disabledService) StartStreamProcessing(*StreamingRequest) (*StreamingResult, error) {
	return nil, d.disabledErr()
}
func (d *disabledService) RunCustomAnalytics(*CustomAnalyticsRequest) (*CustomAnalyticsResult, error) {
	return nil, d.disabledErr()
}
