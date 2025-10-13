package testutils

import (
	"testing"

	"github.com/costscope/costscope/internal/providers/types"
)

// CreateTestProviderConfig creates a standardized test provider configuration
func CreateTestProviderConfig(name string, providerType types.ProviderType, credentials map[string]string) *types.ProviderConfig {
	return &types.ProviderConfig{
		Name:        name,
		Type:        providerType,
		Credentials: credentials,
	}
}

// AssertProviderCreation is a helper function to test provider creation
func AssertProviderCreation(t *testing.T, provider interface{}, err error, expectedName string) {
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}
	if provider == nil {
		t.Fatal("Provider should not be nil")
	}
}

// GetAWSTestCredentials returns test credentials for AWS
func GetAWSTestCredentials() map[string]string {
	return map[string]string{
		"access_key": "test-access-key",
		"secret_key": "test-secret-key",
	}
}

// GetGCPTestCredentials returns test credentials for GCP
func GetGCPTestCredentials() map[string]string {
	return map[string]string{
		"project_id":          "test-project-123456",
		"service_account_key": `{"type":"service_account","project_id":"test-project-123456","private_key_id":"key123","private_key":"-----BEGIN PRIVATE KEY-----\nMIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQC...\n-----END PRIVATE KEY-----\n","client_email":"test@test-project-123456.iam.gserviceaccount.com","client_id":"123456789012345678901","auth_uri":"https://accounts.google.com/o/oauth2/auth","token_uri":"https://oauth2.googleapis.com/token"}`,
	}
}

// GetAzureTestCredentials returns test credentials for Azure
func GetAzureTestCredentials() map[string]string {
	return map[string]string{
		"tenant_id":       "test-tenant-id",
		"client_id":       "test-client-id",
		"client_secret":   "test-client-secret",
		"subscription_id": "test-subscription-id",
	}
}
