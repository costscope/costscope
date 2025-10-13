package azure

import (
	"context"
	"testing"

	"github.com/costscope/costscope/internal/providers/testutils"
	"github.com/costscope/costscope/internal/providers/types"
)

// Helper function to create test Azure provider
func createTestAzureProvider(t *testing.T) *AzureProvider {
	t.Helper()
	config := testutils.CreateTestProviderConfig("test-azure", types.ProviderTypeAzure, testutils.GetAzureTestCredentials())
	provider, err := NewAzureProvider(config)
	testutils.AssertProviderCreation(t, provider, err, "test-azure")
	return provider
}

func TestNewAzureProvider(t *testing.T) {
	provider := createTestAzureProvider(t)
	if provider.GetName() != "test-azure" {
		t.Errorf("Expected name 'test-azure', got: %s", provider.GetName())
	}
	if provider.GetType() != types.ProviderTypeAzure {
		t.Errorf("Expected type Azure, got: %s", provider.GetType())
	}
}

func TestNewAzureProviderMissingCredentials(t *testing.T) {
	config := &types.ProviderConfig{Name: "test-azure", Type: types.ProviderTypeAzure, Credentials: map[string]string{}}
	_, err := NewAzureProvider(config)
	if err == nil {
		t.Error("Expected error for missing credentials, got nil")
	}
	expectedError := "subscription_id is required in credentials"
	if err != nil && err.Error() != expectedError { // err non-nil guaranteed here but guard defensively
		t.Errorf("Expected error '%s', got: '%s'", expectedError, err.Error())
	}
}

func TestAzureTestutilsCredentialsHelper(t *testing.T) {
	creds := testutils.GetAzureTestCredentials()
	required := []string{"subscription_id", "tenant_id", "client_id", "client_secret"}
	for _, k := range required {
		if _, ok := creds[k]; !ok {
			t.Fatalf("expected key %s in azure test credentials", k)
		}
	}
}

func TestAzureProviderGetProviderInfo(t *testing.T) {
	provider := createTestAzureProvider(t)

	info, err := provider.GetProviderInfo(context.Background())
	if err != nil {
		t.Fatalf("Failed to get provider info: %v", err)
	}

	if info.Name != "test-azure" {
		t.Errorf("Expected name 'test-azure', got: %s", info.Name)
	}

	if info.Type != types.ProviderTypeAzure {
		t.Errorf("Expected type Azure, got: %s", info.Type)
	}

	if len(info.SupportedRegions) == 0 {
		t.Error("Expected supported regions, got empty list")
	}
}

func TestAzureProviderValidateCredentials(t *testing.T) {
	provider := createTestAzureProvider(t)

	// Test with empty credentials
	err := provider.ValidateCredentials(context.Background(), map[string]string{})
	if err == nil {
		t.Error("Expected error for empty credentials, got nil")
	}

	// Test with missing client_secret
	err = provider.ValidateCredentials(context.Background(), map[string]string{
		"subscription_id": "12345678-1234-1234-1234-123456789012",
		"tenant_id":       "87654321-4321-4321-4321-210987654321",
		"client_id":       "abcdef12-3456-7890-abcd-ef1234567890",
	})
	if err == nil {
		t.Error("Expected error for missing client_secret, got nil")
	}

	// Test with valid credentials
	err = provider.ValidateCredentials(context.Background(), map[string]string{
		"subscription_id": "12345678-1234-1234-1234-123456789012",
		"tenant_id":       "87654321-4321-4321-4321-210987654321",
		"client_id":       "abcdef12-3456-7890-abcd-ef1234567890",
		"client_secret":   "test-client-secret",
	})
	if err != nil {
		t.Errorf("Expected no error for valid credentials, got: %v", err)
	}
}

func TestAzureProviderGetCostData(t *testing.T) {
	provider := createTestAzureProvider(t)

	// Test placeholder functionality - currently returns sample data
	costs, err := provider.GetCostData(context.Background(), types.CostDataParams{})
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if costs == nil {
		t.Error("Expected costs data, got nil")
	}

	// Currently returns sample data, not empty
	if len(costs) == 0 {
		t.Error("Expected sample cost data, got empty slice")
	}
}

func TestAzureProviderGetResourceData(t *testing.T) {
	provider := createTestAzureProvider(t)

	// Test placeholder functionality - currently returns sample data
	resources, err := provider.GetResourceData(context.Background(), types.ResourceDataParams{})
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if resources == nil {
		t.Error("Expected resources data, got nil")
	}

	// Currently returns sample data, not empty
	if len(resources) == 0 {
		t.Error("Expected sample resource data, got empty slice")
	}
}

func TestAzureProviderGetSupportedRegions(t *testing.T) {
	provider := createTestAzureProvider(t)

	regions := provider.GetSupportedRegions()
	if len(regions) == 0 {
		t.Error("Expected supported regions, got empty list")
	}

	// Check for some known Azure regions
	foundEastUS := false
	foundWestEurope := false
	for _, region := range regions {
		if region == "East US" {
			foundEastUS = true
		}
		if region == "West Europe" {
			foundWestEurope = true
		}
	}

	if !foundEastUS {
		t.Error("Expected to find 'East US' in supported regions")
	}
	if !foundWestEurope {
		t.Error("Expected to find 'West Europe' in supported regions")
	}
}
