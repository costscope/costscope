package multitenant

import (
	"context"
	"local/costscope/internal/core/config"
	"testing"
)

func TestIsEnabled_DefaultFalse(t *testing.T) {
	cfg := &config.ConsolidatedConfig{MultiTenant: config.UnifiedMultiTenantConfig{Enabled: false}}
	if IsEnabled(cfg) {
		// should be false by default
		// (ensures feature flag off yields no activation)
		// This acts as regression guard for TASK-MULTITENANT-SKEL acceptance criteria.
		// Behavior: disabled must not alter system logic; here we only check flag.
		// Additional integration tests will be added when enforcement implemented.
		// If this fails, unintended enablement occurred.
		t.Fatalf("expected multi-tenancy disabled by default")
	}
}

func TestEffectiveTenant_WhenDisabled_ReturnsEmpty(t *testing.T) {
	cfg := &config.ConsolidatedConfig{MultiTenant: config.UnifiedMultiTenantConfig{Enabled: false}}
	id := EffectiveTenant(context.Background(), cfg, &NoopResolver{})
	if id != "" {
		t.Fatalf("expected empty tenant id when feature disabled, got %q", id)
	}
}

func TestNoopResolverAlwaysEmpty(t *testing.T) {
	res := &NoopResolver{}
	id, err := res.Resolve(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "" {
		t.Fatalf("expected empty id from noop resolver")
	}
}

func TestEffectiveTenant_WithEnabled_UsesContextResolver(t *testing.T) {
	cfg := &config.ConsolidatedConfig{MultiTenant: config.UnifiedMultiTenantConfig{Enabled: true}}
	// Attach tenant to context
	base := context.Background()
	ctx := WithTenantToContext(base, "tenant-xyz")
	// Use context resolver with default key
	r := ContextResolver{Key: ContextKeyTenantID}
	got := EffectiveTenant(ctx, cfg, r)
	if got != "tenant-xyz" {
		t.Fatalf("expected tenant-xyz, got %q", got)
	}
}
