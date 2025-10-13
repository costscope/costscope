package providers

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/costscope/costscope/internal/providers/types"
)

// MockProvider implements the CloudProvider interface for testing
type MockProvider struct {
	name                 string
	providerType         types.ProviderType
	shouldFailValidation bool
}

func (mp *MockProvider) ValidateCredentials(ctx context.Context, config map[string]string) error {
	if mp.shouldFailValidation {
		return fmt.Errorf("mock validation error")
	}
	return nil
}

func TestCreateProvider(t *testing.T) {
	pm := NewProviderManager()

	// Test AWS provider creation
	awsConfig := &types.ProviderConfig{
		Name: "test-aws",
		Type: types.ProviderTypeAWS,
		Credentials: map[string]string{
			"access_key": "test-key",
			"secret_key": "test-secret",
		},
	}

	provider, err := pm.CreateProvider(awsConfig)
	if err != nil {
		t.Errorf("Expected no error creating AWS provider, got: %v", err)
	}

	if provider.GetType() != types.ProviderTypeAWS {
		t.Errorf("Expected AWS provider type, got: %s", provider.GetType())
	}

	// Test Azure provider creation
	azureConfig := &types.ProviderConfig{
		Name: "test-azure",
		Type: types.ProviderTypeAzure,
		Credentials: map[string]string{
			"subscription_id": "12345678-1234-1234-1234-123456789012",
			"tenant_id":       "87654321-4321-4321-4321-210987654321",
			"client_id":       "abcdef12-3456-7890-abcd-ef1234567890",
			"client_secret":   "test-secret",
		},
	}

	provider, err = pm.CreateProvider(azureConfig)
	if err != nil {
		t.Errorf("Expected no error creating Azure provider, got: %v", err)
	}

	if provider.GetType() != types.ProviderTypeAzure {
		t.Errorf("Expected Azure provider type, got: %s", provider.GetType())
	}

	// Test GCP provider creation
	gcpConfig := &types.ProviderConfig{
		Name: "test-gcp",
		Type: types.ProviderTypeGCP,
		Credentials: map[string]string{
			"project_id":          "test-project",
			"service_account_key": `{"type":"service_account"}`,
		},
	}

	provider, err = pm.CreateProvider(gcpConfig)
	if err != nil {
		t.Errorf("Expected no error creating GCP provider, got: %v", err)
	}

	if provider.GetType() != types.ProviderTypeGCP {
		t.Errorf("Expected GCP provider type, got: %s", provider.GetType())
	}

	// Test unsupported provider type
	unsupportedConfig := &types.ProviderConfig{
		Name: "test-unsupported",
		Type: "unsupported",
	}

	_, err = pm.CreateProvider(unsupportedConfig)
	if err == nil {
		t.Error("Expected error for unsupported provider type, got nil")
	}
}

func TestRegisterNewProvider(t *testing.T) {
	pm := NewProviderManager()

	config := &types.ProviderConfig{
		Name: "test-azure-new",
		Type: types.ProviderTypeAzure,
		Credentials: map[string]string{
			"subscription_id": "12345678-1234-1234-1234-123456789012",
			"tenant_id":       "87654321-4321-4321-4321-210987654321",
			"client_id":       "abcdef12-3456-7890-abcd-ef1234567890",
			"client_secret":   "test-secret",
		},
	}

	err := pm.RegisterNewProvider(config)
	if err != nil {
		t.Errorf("Expected no error registering new provider, got: %v", err)
	}

	// Verify provider was registered
	provider, err := pm.GetProvider("test-azure-new")
	if err != nil {
		t.Errorf("Expected to find registered provider, got error: %v", err)
	}

	if provider.GetType() != types.ProviderTypeAzure {
		t.Errorf("Expected Azure provider type, got: %s", provider.GetType())
	}
}

func (mp *MockProvider) GetProviderInfo(ctx context.Context) (types.ProviderInfo, error) {
	return types.ProviderInfo{
		Name:             mp.name,
		Type:             mp.providerType,
		Version:          "1.0.0",
		SupportedRegions: []string{"us-east-1", "us-west-2"},
		Capabilities:     []string{"cost-data", "resource-data"},
		Metadata:         map[string]string{"test": "true"},
	}, nil
}

func (mp *MockProvider) GetCostData(ctx context.Context, params types.CostDataParams) ([]types.CostRecord, error) {
	return []types.CostRecord{
		{
			Date:       time.Now(),
			Amount:     100.50,
			Currency:   "USD",
			Service:    "EC2",
			Region:     "us-east-1",
			ResourceID: "i-1234567890abcdef0",
			Tags:       map[string]string{"Environment": "test"},
			Metadata:   map[string]interface{}{"instance_type": "t2.micro"},
		},
	}, nil
}

func (mp *MockProvider) GetResourceData(ctx context.Context, params types.ResourceDataParams) ([]types.ResourceRecord, error) {
	return []types.ResourceRecord{
		{
			ID:         "i-1234567890abcdef0",
			Name:       "test-instance",
			Type:       "EC2.Instance",
			Region:     "us-east-1",
			Status:     "running",
			Cost:       50.25,
			Currency:   "USD",
			Tags:       map[string]string{"Environment": "test"},
			Properties: map[string]interface{}{"instance_type": "t2.micro"},
			CreatedAt:  time.Now().Add(-24 * time.Hour),
			UpdatedAt:  time.Now(),
		},
	}, nil
}

func (mp *MockProvider) GetName() string {
	return mp.name
}

func (mp *MockProvider) GetType() types.ProviderType {
	return mp.providerType
}

func (mp *MockProvider) GetSupportedRegions() []string {
	return []string{"us-east-1", "us-west-2"}
}

func TestNewProviderManager(t *testing.T) {
	pm := NewProviderManager()
	if pm == nil {
		t.Fatal("NewProviderManager should not return nil")
	}

	if pm.providers == nil {
		t.Error("providers map should be initialized")
	}

	if pm.configs == nil {
		t.Error("configs map should be initialized")
	}

	if pm.statuses == nil {
		t.Error("statuses map should be initialized")
	}
}

func TestRegisterProvider(t *testing.T) {
	pm := NewProviderManager()

	mockProvider := &MockProvider{
		name:         "test-aws",
		providerType: types.ProviderTypeAWS,
	}

	config := &types.ProviderConfig{
		Name:        "test-aws",
		Type:        types.ProviderTypeAWS,
		Credentials: map[string]string{"access_key": "test", "secret_key": "test"},
		Settings:    map[string]interface{}{"region": "us-east-1"},
		Regions:     []string{"us-east-1"},
		IsDefault:   true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := pm.RegisterProvider("test-aws", mockProvider, config)
	if err != nil {
		t.Errorf("RegisterProvider should not return error, got: %v", err)
	}

	// Test duplicate registration
	err = pm.RegisterProvider("test-aws", mockProvider, config)
	if err == nil {
		t.Error("RegisterProvider should return error for duplicate registration")
	}
}

func TestGetProvider(t *testing.T) {
	pm := NewProviderManager()

	mockProvider := &MockProvider{
		name:         "test-aws",
		providerType: types.ProviderTypeAWS,
	}

	config := &types.ProviderConfig{
		Name: "test-aws",
		Type: types.ProviderTypeAWS,
	}

	err := pm.RegisterProvider("test-aws", mockProvider, config)
	if err != nil {
		t.Fatalf("Failed to register provider: %v", err)
	}

	// Test getting existing provider
	provider, err := pm.GetProvider("test-aws")
	if err != nil {
		t.Errorf("GetProvider should not return error, got: %v", err)
	}

	if provider == nil {
		t.Error("GetProvider should return provider")
	}

	// Test getting non-existing provider
	_, err = pm.GetProvider("non-existing")
	if err == nil {
		t.Error("GetProvider should return error for non-existing provider")
	}
}

func TestValidateProvider(t *testing.T) {
	pm := NewProviderManager()
	ctx := context.Background()

	// Test with provider that passes validation
	mockProvider := &MockProvider{
		name:                 "test-aws",
		providerType:         types.ProviderTypeAWS,
		shouldFailValidation: false,
	}

	config := &types.ProviderConfig{
		Name:        "test-aws",
		Type:        types.ProviderTypeAWS,
		Credentials: map[string]string{"access_key": "test"},
	}

	err := pm.RegisterProvider("test-aws", mockProvider, config)
	if err != nil {
		t.Fatalf("Failed to register provider: %v", err)
	}

	err = pm.ValidateProvider(ctx, "test-aws")
	if err != nil {
		t.Errorf("ValidateProvider should not return error, got: %v", err)
	}

	// Check status was updated
	status, err := pm.GetProviderStatus("test-aws")
	if err != nil {
		t.Errorf("GetProviderStatus should not return error, got: %v", err)
	}

	if !status.IsConnected {
		t.Error("Provider should be connected after successful validation")
	}

	// Test with provider that fails validation
	mockProviderFail := &MockProvider{
		name:                 "test-aws-fail",
		providerType:         types.ProviderTypeAWS,
		shouldFailValidation: true,
	}

	configFail := &types.ProviderConfig{
		Name:        "test-aws-fail",
		Type:        types.ProviderTypeAWS,
		Credentials: map[string]string{"access_key": "invalid"},
	}

	err = pm.RegisterProvider("test-aws-fail", mockProviderFail, configFail)
	if err != nil {
		t.Fatalf("Failed to register provider: %v", err)
	}

	err = pm.ValidateProvider(ctx, "test-aws-fail")
	if err == nil {
		t.Error("ValidateProvider should return error for invalid provider")
	}
}

func TestListProviders(t *testing.T) {
	pm := NewProviderManager()

	// Initially should be empty
	providers := pm.ListProviders()
	if len(providers) != 0 {
		t.Error("ListProviders should return empty map initially")
	}

	// Register a provider
	mockProvider := &MockProvider{
		name:         "test-aws",
		providerType: types.ProviderTypeAWS,
	}

	config := &types.ProviderConfig{
		Name: "test-aws",
		Type: types.ProviderTypeAWS,
	}

	err := pm.RegisterProvider("test-aws", mockProvider, config)
	if err != nil {
		t.Fatalf("Failed to register provider: %v", err)
	}

	providers = pm.ListProviders()
	if len(providers) != 1 {
		t.Error("ListProviders should return one provider")
	}

	if providers["test-aws"] != types.ProviderTypeAWS {
		t.Error("Provider type should match")
	}
}

func TestRemoveProvider(t *testing.T) {
	pm := NewProviderManager()

	mockProvider := &MockProvider{
		name:         "test-aws",
		providerType: types.ProviderTypeAWS,
	}

	config := &types.ProviderConfig{
		Name: "test-aws",
		Type: types.ProviderTypeAWS,
	}

	err := pm.RegisterProvider("test-aws", mockProvider, config)
	if err != nil {
		t.Fatalf("Failed to register provider: %v", err)
	}

	// Verify provider exists
	providers := pm.ListProviders()
	if len(providers) != 1 {
		t.Error("Should have one provider before removal")
	}

	// Remove provider
	err = pm.RemoveProvider("test-aws")
	if err != nil {
		t.Errorf("RemoveProvider should not return error, got: %v", err)
	}

	// Verify provider is removed
	providers = pm.ListProviders()
	if len(providers) != 0 {
		t.Error("Should have no providers after removal")
	}

	// Test removing non-existing provider
	err = pm.RemoveProvider("non-existing")
	if err == nil {
		t.Error("RemoveProvider should return error for non-existing provider")
	}
}
