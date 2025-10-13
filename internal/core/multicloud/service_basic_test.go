package multicloud

import (
	"testing"

	"github.com/costscope/costscope/internal/core/logging"
	"github.com/costscope/costscope/internal/providers"
)

func TestNewMulticloudService_Defaults(t *testing.T) {
	logger := logging.NewLogger(logging.LevelInfo)
	pm := providers.NewProviderManager()
	svc := NewMulticloudService(pm, logger)
	if svc == nil || svc.config == nil {
		t.Fatalf("expected service & config")
	}
	if svc.config.DefaultCurrency != "USD" {
		t.Fatalf("unexpected default currency: %s", svc.config.DefaultCurrency)
	}
	if svc.optimizationEngine == nil || svc.migrationEngine == nil || svc.discoveryEngine == nil {
		t.Fatalf("expected engines to be initialized")
	}
}
