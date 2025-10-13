package multitenant

// TASK-MULTITENANT-SKEL: Helper wiring multi-tenant resolver into higher layers.
// This isolates the decision logic for when to inject tenant scoping into
// analytics/database filter objects. Future phases may expand with audit logging
// or error surfacing (currently silent per skeleton acceptance criteria).

import (
	"context"

	"github.com/costscope/costscope/internal/core/config"
	"github.com/costscope/costscope/internal/database"
)

// InjectTenantFilter mutates filters to set TenantID when ALL conditions hold:
//  1. cfg != nil AND multi-tenancy enabled.
//  2. filters != nil.
//  3. filters.TenantID empty (explicit user value always wins).
//  4. resolver resolves non-empty tenant id.
//
// Silent no-op otherwise (keeps callers simple and test friendly).
func InjectTenantFilter(ctx context.Context, cfg *config.ConsolidatedConfig, filters *database.AnalyticsFilters, r TenantResolver) {
	if cfg == nil || filters == nil || !IsEnabled(cfg) {
		return
	}
	if filters.TenantID != "" { // explicit override
		return
	}
	if r == nil {
		return
	}
	if id := EffectiveTenant(ctx, cfg, r); id != "" { // EffectiveTenant already double-checks flag
		filters.TenantID = id
	}
}
