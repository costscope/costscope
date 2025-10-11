package aws

import (
	"context"
	"testing"

	"local/costscope/internal/providers/testutils"
	"local/costscope/internal/providers/types"
)

// Helper function to create test AWS provider
func createTestAWSProvider(t *testing.T) *AWSProvider {
	config := testutils.CreateTestProviderConfig("test-aws", types.ProviderTypeAWS, testutils.GetAWSTestCredentials())
	provider, err := NewAWSProvider(config)
	testutils.AssertProviderCreation(t, provider, err, "test-aws")
	return provider
}

func TestNewAWSProvider(t *testing.T) {
	config := &types.ProviderConfig{
		Name: "test-aws",
		Type: types.ProviderTypeAWS,
		Credentials: map[string]string{
			"access_key": "test-access-key",
			"secret_key": "test-secret-key",
		},
	}

	provider, err := NewAWSProvider(config)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if provider == nil {
		t.Error("Expected provider to be created, got nil")
	}

	if provider.GetName() != "test-aws" {
		t.Errorf("Expected name 'test-aws', got: %s", provider.GetName())
	}

	if provider.GetType() != types.ProviderTypeAWS {
		t.Errorf("Expected type AWS, got: %s", provider.GetType())
	}
}

func TestNewAWSProviderMissingCredentials(t *testing.T) {
	config := &types.ProviderConfig{
		Name:        "test-aws",
		Type:        types.ProviderTypeAWS,
		Credentials: map[string]string{},
	}

	_, err := NewAWSProvider(config)
	if err == nil {
		t.Error("Expected error for missing credentials, got nil")
	}

	expectedError := "access_key is required in credentials"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got: '%s'", expectedError, err.Error())
	}
}

func TestAWSProviderGetProviderInfo(t *testing.T) {
	config := &types.ProviderConfig{
		Name: "test-aws",
		Type: types.ProviderTypeAWS,
		Credentials: map[string]string{
			"access_key": "test-access-key",
			"secret_key": "test-secret-key",
		},
	}

	provider, err := NewAWSProvider(config)
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	// Note: This test will fail with real AWS API calls since credentials are fake
	// We're testing that the method signature works properly
	_, err = provider.GetProviderInfo(context.Background())
	if err == nil {
		t.Log("Note: GetProviderInfo unexpectedly succeeded with fake credentials")
	} else {
		// Expected behavior with fake credentials
		t.Logf("GetProviderInfo failed as expected with fake credentials: %v", err)
	}

	// Test the basic methods that don't require AWS API calls
	if provider.GetName() != "test-aws" {
		t.Errorf("Expected name 'test-aws', got: %s", provider.GetName())
	}

	if provider.GetType() != types.ProviderTypeAWS {
		t.Errorf("Expected type AWS, got: %s", provider.GetType())
	}
}

func TestAWSProviderValidateCredentialsNoNetwork(t *testing.T) {
	// This test doesn't make actual AWS calls, just validates the method exists
	config := &types.ProviderConfig{
		Name: "test-aws",
		Type: types.ProviderTypeAWS,
		Credentials: map[string]string{
			"access_key": "test-access-key",
			"secret_key": "test-secret-key",
		},
	}

	provider, err := NewAWSProvider(config)
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	// Test with empty credentials
	err = provider.ValidateCredentials(context.Background(), map[string]string{})
	if err == nil {
		t.Error("Expected error for empty credentials, got nil")
	}

	// Test with missing secret key
	err = provider.ValidateCredentials(context.Background(), map[string]string{
		"access_key": "test",
	})
	if err == nil {
		t.Error("Expected error for missing secret_key, got nil")
	}
}

func TestAWSProviderGetCostData(t *testing.T) {
	provider := createTestAWSProvider(t)

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

func TestAWSProviderGetResourceData(t *testing.T) {
	provider := createTestAWSProvider(t)

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
