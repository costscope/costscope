//go:build integration

package azure

import (
	"context"
	"testing"

	"local/costscope/internal/providers/testutils"
	"local/costscope/internal/providers/types"
)

// TestAzureIntegrationCredentials ensures env credential plumbing works (placeholder provider).
func TestAzureIntegrationCredentials(t *testing.T) {
	creds := testutils.RequireAzureIntegrationCredentials(t)
	cfg := &types.ProviderConfig{
		Name:        "azure-int",
		Type:        types.ProviderTypeAzure,
		Credentials: creds,
	}
	p, err := NewAzureProvider(cfg)
	if err != nil {
		t.Fatalf("failed to construct Azure provider: %v", err)
	}
	if err := p.ValidateCredentials(context.Background(), creds); err != nil {
		t.Fatalf("ValidateCredentials failed: %v", err)
	}
	if _, err := p.GetProviderInfo(context.Background()); err != nil {
		t.Fatalf("GetProviderInfo failed: %v", err)
	}
}
