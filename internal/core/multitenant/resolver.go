package multitenant

//nolint:unused // Resolver types exercised only in multi-tenant tests until full enforcement phase rolls out.
// TASK-MULTITENANT-SKEL: Initial skeleton for future multi-tenant support.
// This package intentionally contains minimal wiring so that enabling the feature
// later (isolation, scoping, authz integration) will not require invasive refactors.

import (
	"context"
	"local/costscope/internal/core/config"
)

// TenantResolver resolves a tenant identifier from a request context (e.g. auth token, headers).
// Future implementations may inspect JWT claims or API key metadata.
// Returning an empty string means: no tenant resolved / single-tenant mode.
type TenantResolver interface {
	Resolve(ctx context.Context) (string, error)
}

// NoopResolver is the default resolver used when multi-tenancy is disabled.
// It always returns an empty tenant id and no error.
type NoopResolver struct{}

// Resolve implements TenantResolver.
func (n *NoopResolver) Resolve(ctx context.Context) (string, error) { return "", nil }

// IsEnabled returns whether multi-tenancy feature flag is enabled via config.
// Environment variable: COSTSCOPE_MULTI_TENANT_ENABLED (bool)
// YAML key: multi_tenant.enabled
// Default: false
func IsEnabled(cfg *config.ConsolidatedConfig) bool {
	if cfg == nil {
		return false
	}
	return cfg.MultiTenant.Enabled
}

// EffectiveTenant returns resolved tenant id only when feature flag is enabled.
// If disabled or resolver returns empty, an empty string is returned.
func EffectiveTenant(ctx context.Context, cfg *config.ConsolidatedConfig, r TenantResolver) string {
	if !IsEnabled(cfg) || r == nil {
		return ""
	}
	id, _ := r.Resolve(ctx) // ignore error for skeleton; future phases will handle
	return id
}

// ----------------------------------------------------------------------------
// Context helpers and a generic resolver for context-scoped tenant propagation
// ----------------------------------------------------------------------------

// ctxKey is an unexported type to avoid collisions in context values.
type ctxKey string

// ContextKeyTenantID is the key used for storing tenant id in request contexts.
// Exported for middleware and handler layers to share consistently.
var ContextKeyTenantID ctxKey = "tenant_id"

// WithTenantToContext attaches the provided tenant id to the context.
func WithTenantToContext(ctx context.Context, tenantID string) context.Context {
	if tenantID == "" {
		return ctx
	}
	return context.WithValue(ctx, ContextKeyTenantID, tenantID)
}

// ContextResolver resolves tenant id from a context value under the provided key.
// If the value is not present or not a string, it returns an empty id.
type ContextResolver struct {
	Key interface{}
}

// Resolve implements TenantResolver by reading from context.Value(Key).
func (r ContextResolver) Resolve(ctx context.Context) (string, error) {
	if ctx == nil {
		return "", nil
	}
	v := ctx.Value(r.Key)
	if s, ok := v.(string); ok {
		return s, nil
	}
	return "", nil
}
