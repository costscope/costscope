//go:build integration

package aws

import (
	"context"
	"testing"

	"local/costscope/internal/providers/testutils"
	"local/costscope/internal/providers/types"
)

// TestAWSIntegrationCredentialsAndSTS validates that when real credentials are supplied via env
// the provider can perform an STS identity call (ValidateCredentials already exercises this path).
// Skips automatically when env is missing.
func TestAWSIntegrationCredentialsAndSTS(t *testing.T) {
	creds := testutils.RequireAWSIntegrationCredentials(t)

	cfg := &types.ProviderConfig{
		Name:        "aws-int",
		Type:        types.ProviderTypeAWS,
		Credentials: map[string]string{"access_key": creds["access_key"], "secret_key": creds["secret_key"]},
		Settings:    map[string]interface{}{"region": creds["region"]},
	}
	p, err := NewAWSProvider(cfg)
	if err != nil {
		t.Fatalf("failed to construct AWS provider with integration creds: %v", err)
	}
	// Validate explicitly (network STS call)
	if err := p.ValidateCredentials(context.Background(), cfg.Credentials); err != nil {
		t.Fatalf("ValidateCredentials failed: %v", err)
	}
	if _, err := p.GetProviderInfo(context.Background()); err != nil {
		t.Fatalf("GetProviderInfo failed: %v", err)
	}
}
