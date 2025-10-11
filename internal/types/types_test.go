package types

import "testing"

// Placeholder test retained after removal of legacy generic Result helpers.
// The original factories (NewResult, NewError, NewPaginatedResult) were unused
// across the codebase and have been removed to prevent divergence from the
// unified API response envelope (internal/api/response). This test remains to
// keep package test scaffolding (can be expanded with future type-specific tests).
func TestTypes_Placeholder(t *testing.T) {
	// no-op
}
