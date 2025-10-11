package multitenant

import (
	"context"
	"local/costscope/internal/core/config"
	"local/costscope/internal/database"
	"testing"
)

// testResolver implements TenantResolver returning a fixed id.
type testResolver struct{ id string }

func (t testResolver) Resolve(ctx context.Context) (string, error) { return t.id, nil }

func TestInjectTenantFilter_Scenarios(t *testing.T) {
	cases := []struct {
		name       string
		enabled    bool
		existingID string
		resolverID string
		expectID   string // expected filters.TenantID after injection
		nilFilters bool
		useNilCfg  bool
	}{
		{name: "feature_disabled_noop", enabled: false, resolverID: "tenant-a", expectID: ""},
		{name: "nil_filters_noop", enabled: true, resolverID: "tenant-a", nilFilters: true, expectID: ""},
		{name: "nil_cfg_noop", enabled: true, resolverID: "tenant-a", useNilCfg: true, expectID: ""},
		{name: "explicit_override", enabled: true, existingID: "explicit", resolverID: "tenant-a", expectID: "explicit"},
		{name: "resolver_injected", enabled: true, resolverID: "tenant-a", expectID: "tenant-a"},
		{name: "empty_resolver_id", enabled: true, resolverID: "", expectID: ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var cfg *config.ConsolidatedConfig
			if !tc.useNilCfg {
				cfg = &config.ConsolidatedConfig{MultiTenant: config.UnifiedMultiTenantConfig{Enabled: tc.enabled}}
			}
			var filters *database.AnalyticsFilters
			if !tc.nilFilters {
				filters = &database.AnalyticsFilters{TenantID: tc.existingID}
			}
			InjectTenantFilter(context.Background(), cfg, filters, testResolver{id: tc.resolverID})
			if filters == nil { // nothing to assert
				return
			}
			if filters.TenantID != tc.expectID {
				t.Fatalf("expected TenantID %q, got %q", tc.expectID, filters.TenantID)
			}
		})
	}
}
