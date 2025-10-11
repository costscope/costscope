//go:build !enterprise

package database

import (
	"context"

	"local/costscope/internal/core/enterprise"
	"local/costscope/internal/core/logging"
)

// Intentional stub (enterprise gating): minimal surface for feature detection.
// Community build exposes ONLY CreateConnectionPool (and Stop for forward
// compatibility) so downstream code can detect the disabled enterprise feature
// without pulling a wide API that invites accidental dependencies.
// Full implementation (with pooling, metrics, optimization) is compiled in
// with: `go build -tags enterprise ./...`
//
// Rationale for minimalism:
//   - Shrinks public surface & deadcode noise.
//   - Avoids false dependencies on enterprise-only metrics/types.
//   - Keeps upgrade path explicit: adding new gated capability requires
//     deliberate API expansion instead of latent no-op methods.
type EnterpriseConnectionManager struct{}

const connectionManagerFeature = "connection_manager"

func disabledConnMgr() error {
	enterprise.ObserveInvocation(connectionManagerFeature, false)
	enterprise.ObserveError(connectionManagerFeature, "disabled")
	return enterprise.DisabledError(connectionManagerFeature)
}

// NewEnterpriseConnectionManager returns a disabled stub instance.
// Intentional stub (enterprise gating): constructor returns disabled stub; real pooling behind -tags enterprise.
func NewEnterpriseConnectionManager(_ *logging.Logger) *EnterpriseConnectionManager { //nolint:revive // Parity with enterprise constructor
	return &EnterpriseConnectionManager{}
}

// CreateConnectionPool always returns the enterprise disabled sentinel error in
// community builds. Presence of this method allows callers to probe for
// enterprise capability without linking to wider pool/metrics APIs.
// Intentional stub (enterprise gating): always returns enterprise disabled error.
func (ecm *EnterpriseConnectionManager) CreateConnectionPool(context.Context, string, string, string) error {
	return disabledConnMgr()
}
