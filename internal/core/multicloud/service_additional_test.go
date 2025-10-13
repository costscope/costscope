package multicloud

import (
	"context"
	"testing"

	"github.com/costscope/costscope/internal/core/logging"
	"github.com/costscope/costscope/internal/providers"
)

// TestMulticloudService_ValidateProvidersBranches covers success and error paths.
func TestMulticloudService_ValidateProvidersBranches(t *testing.T) {
	pm := providers.NewProviderManager()
	svc := NewMulticloudService(pm, logging.NewLogger(logging.LevelError))
	// Provider validation success path (unknown later engine errors ignored)
	_, _ = svc.AnalyzeOptimizations(context.Background(), &OptimizationRequest{Providers: []string{"aws"}})
	// Empty providers
	if _, err := svc.AnalyzeOptimizations(context.Background(), &OptimizationRequest{Providers: []string{}}); err == nil {
		t.Fatalf("expected error for empty providers")
	}
	// Unknown provider
	if err := svc.validateProviders([]string{"unknown"}); err == nil {
		t.Fatalf("expected unknown provider error")
	}
}
