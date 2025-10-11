package analytics

import (
	"context"
	"strings"
	"testing"

	"local/costscope/internal/core/config"
	"local/costscope/internal/core/multitenant"
	"local/costscope/internal/database"
)

// capturingExecMulti reuses simple capturing executor logic (kept local to avoid test coupling).
type capturingExecMulti struct{ last string }

func (c *capturingExecMulti) ExecuteQuery(_ context.Context, q string) (*database.QueryResult, error) {
	c.last = q
	return &database.QueryResult{Columns: []string{}, Data: []map[string]interface{}{}}, nil
}

func TestFacade_TenantInjection_TableDriven(t *testing.T) {
	cases := []struct {
		name           string
		enabled        bool
		ctxTenant      string
		explicitTenant string
		expectSub      string // empty => expect no tenant clause
	}{
		{name: "flag_off_ignores_context", enabled: false, ctxTenant: "t-ctx", explicitTenant: "", expectSub: ""},
		{name: "explicit_overrides_context", enabled: true, ctxTenant: "t-ctx", explicitTenant: "t-explicit", expectSub: "tenant_id = 't-explicit'"},
		{name: "context_used_when_no_explicit", enabled: true, ctxTenant: "t-ctx", explicitTenant: "", expectSub: "tenant_id = 't-ctx'"},
		{name: "empty_context_no_filter", enabled: true, ctxTenant: "", explicitTenant: "", expectSub: ""},
		{name: "explicit_only", enabled: true, ctxTenant: "", explicitTenant: "t-explicit", expectSub: "tenant_id = 't-explicit'"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			exec := &capturingExecMulti{}
			cfg := &config.ConsolidatedConfig{MultiTenant: config.UnifiedMultiTenantConfig{Enabled: tc.enabled}}
			fac := NewFacadeWithConfig(exec, cfg)

			ctx := context.Background()
			if tc.ctxTenant != "" {
				ctx = multitenant.WithTenantToContext(ctx, tc.ctxTenant)
			}

			filters := &database.AnalyticsFilters{TenantID: tc.explicitTenant}
			if _, err := fac.CostSummary(ctx, filters); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tc.expectSub == "" {
				if strings.Contains(exec.last, "tenant_id =") {
					t.Fatalf("did not expect tenant filter, SQL=%s", exec.last)
				}
				return
			}
			if !strings.Contains(exec.last, tc.expectSub) {
				t.Fatalf("expected '%s' in SQL, got %s", tc.expectSub, exec.last)
			}
		})
	}
}
