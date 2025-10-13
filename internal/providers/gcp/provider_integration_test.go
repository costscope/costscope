//go:build integration

package gcp

import (
	"context"
	"testing"

	"github.com/costscope/costscope/internal/providers/testutils"
	"github.com/costscope/costscope/internal/providers/types"
)

// TestGCPIntegrationCredentials ensures env credential plumbing works (placeholder provider).
func TestGCPIntegrationCredentials(t *testing.T) {
	creds := testutils.RequireGCPIntegrationCredentials(t)
	cfg := &types.ProviderConfig{
		Name:        "gcp-int",
		Type:        types.ProviderTypeGCP,
		Credentials: creds,
	}
	p, err := NewGCPProvider(cfg)
	if err != nil {
		t.Fatalf("failed to construct GCP provider: %v", err)
	}
	if err := p.ValidateCredentials(context.Background(), creds); err != nil {
		t.Fatalf("ValidateCredentials failed: %v", err)
	}
	if _, err := p.GetProviderInfo(context.Background()); err != nil {
		t.Fatalf("GetProviderInfo failed: %v", err)
	}
}
