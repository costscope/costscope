package gcp

import (
	"context"
	"testing"

	"github.com/costscope/costscope/internal/providers/testutils"
	"github.com/costscope/costscope/internal/providers/types"
)

// Helper function to create test GCP provider
func createTestGCPProvider(t *testing.T) *GCPProvider {
	config := testutils.CreateTestProviderConfig("test-gcp", types.ProviderTypeGCP, testutils.GetGCPTestCredentials())
	provider, err := NewGCPProvider(config)
	testutils.AssertProviderCreation(t, provider, err, "test-gcp")
	return provider
}

func TestNewGCPProvider(t *testing.T) {
	provider := createTestGCPProvider(t)

	if provider.GetName() != "test-gcp" {
		t.Errorf("Expected name 'test-gcp', got: %s", provider.GetName())
	}

	if provider.GetType() != types.ProviderTypeGCP {
		t.Errorf("Expected type GCP, got: %s", provider.GetType())
	}
}

func TestNewGCPProviderMissingCredentials(t *testing.T) {
	config := &types.ProviderConfig{
		Name:        "test-gcp",
		Type:        types.ProviderTypeGCP,
		Credentials: map[string]string{},
	}

	_, err := NewGCPProvider(config)
	if err == nil {
		t.Error("Expected error for missing credentials, got nil")
	}

	expectedError := "project_id is required in credentials"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got: '%s'", expectedError, err.Error())
	}
}

func TestGCPProviderGetProviderInfo(t *testing.T) {
	provider := createTestGCPProvider(t)

	info, err := provider.GetProviderInfo(context.Background())
	if err != nil {
		t.Fatalf("Failed to get provider info: %v", err)
	}

	if info.Name != "test-gcp" {
		t.Errorf("Expected name 'test-gcp', got: %s", info.Name)
	}

	if info.Type != types.ProviderTypeGCP {
		t.Errorf("Expected type GCP, got: %s", info.Type)
	}

	if len(info.SupportedRegions) == 0 {
		t.Error("Expected supported regions, got empty list")
	}
}

func TestGCPProviderValidateCredentials(t *testing.T) {
	provider := createTestGCPProvider(t)

	// Test with empty credentials
	err := provider.ValidateCredentials(context.Background(), map[string]string{})
	if err == nil {
		t.Error("Expected error for empty credentials, got nil")
	}

	// Test with missing service_account_key
	err = provider.ValidateCredentials(context.Background(), map[string]string{
		"project_id": "test-project-123456",
	})
	if err == nil {
		t.Error("Expected error for missing service_account_key, got nil")
	}

	// Test with valid credentials
	err = provider.ValidateCredentials(context.Background(), map[string]string{
		"project_id":          "test-project-123456",
		"service_account_key": `{"type":"service_account","project_id":"test-project-123456"}`,
	})
	if err != nil {
		t.Errorf("Expected no error for valid credentials, got: %v", err)
	}
}

func TestGCPProviderGetCostData(t *testing.T) {
	provider := createTestGCPProvider(t)

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

func TestGCPProviderGetResourceData(t *testing.T) {
	provider := createTestGCPProvider(t)

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

func TestGCPProviderGetSupportedRegions(t *testing.T) {
	provider := createTestGCPProvider(t)

	regions := provider.GetSupportedRegions()
	if len(regions) == 0 {
		t.Error("Expected supported regions, got empty list")
	}

	// Check for some known GCP regions
	foundUSCentral1 := false
	foundEuropeWest1 := false
	for _, region := range regions {
		if region == "us-central1" {
			foundUSCentral1 = true
		}
		if region == "europe-west1" {
			foundEuropeWest1 = true
		}
	}

	if !foundUSCentral1 {
		t.Error("Expected to find 'us-central1' in supported regions")
	}
	if !foundEuropeWest1 {
		t.Error("Expected to find 'europe-west1' in supported regions")
	}
}
