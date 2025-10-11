package providers

import (
	"context"
	"fmt"
	"sync"

	"local/costscope/internal/core/logging"
	"local/costscope/internal/core/monitoring/telemetry"
	_ "local/costscope/internal/providers/aws"   // side-effect registration
	_ "local/costscope/internal/providers/azure" // side-effect registration
	_ "local/costscope/internal/providers/gcp"   // side-effect registration
	"local/costscope/internal/providers/registry"
	"local/costscope/internal/providers/types"
)

// ProviderManager manages all cloud providers through unified interfaces
type ProviderManager struct {
	providers map[string]types.CloudProvider
	configs   map[string]*types.ProviderConfig
	statuses  map[string]*types.ProviderStatus
	logger    *logging.Logger
	mutex     sync.RWMutex
}

// NewProviderManager creates a new provider manager
func NewProviderManager() *ProviderManager {
	return &ProviderManager{
		providers: make(map[string]types.CloudProvider),
		configs:   make(map[string]*types.ProviderConfig),
		statuses:  make(map[string]*types.ProviderStatus),
		logger:    logging.NewLogger(logging.LevelInfo),
	}
}

// RegisterProvider registers a cloud provider with the manager
func (pm *ProviderManager) RegisterProvider(name string, provider types.CloudProvider, config *types.ProviderConfig) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	if _, exists := pm.providers[name]; exists {
		return fmt.Errorf("provider %s is already registered", name)
	}

	pm.providers[name] = provider
	pm.configs[name] = config
	pm.statuses[name] = &types.ProviderStatus{
		Name:         name,
		Type:         config.Type,
		IsConnected:  false,
		HealthStatus: "unknown",
	}

	pm.logger.Info(fmt.Sprintf("Registered provider: %s (%s)", name, config.Type))
	return nil
}

// GetProvider retrieves a provider by name
func (pm *ProviderManager) GetProvider(name string) (types.CloudProvider, error) {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	provider, exists := pm.providers[name]
	if !exists {
		return nil, fmt.Errorf("provider %s not found", name)
	}

	return provider, nil
}

// GetAllProviders returns all registered providers
func (pm *ProviderManager) GetAllProviders() map[string]types.CloudProvider {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	result := make(map[string]types.CloudProvider)
	for name, provider := range pm.providers {
		result[name] = provider
	}

	return result
}

// getProviderData is a generic helper for retrieving provider data with existence check
func (pm *ProviderManager) getProviderData(name string, dataMap interface{}, itemType string) (interface{}, error) {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	switch m := dataMap.(type) {
	case map[string]*types.ProviderConfig:
		if config, exists := m[name]; exists {
			return config, nil
		}
	case map[string]*types.ProviderStatus:
		if status, exists := m[name]; exists {
			return status, nil
		}
	}

	return nil, fmt.Errorf("provider %s %s not found", itemType, name)
}

// GetProviderConfig retrieves provider configuration by name
func (pm *ProviderManager) GetProviderConfig(name string) (*types.ProviderConfig, error) {
	data, err := pm.getProviderData(name, pm.configs, "config")
	if err != nil {
		return nil, err
	}
	return data.(*types.ProviderConfig), nil
}

// GetProviderStatus retrieves provider status by name
func (pm *ProviderManager) GetProviderStatus(name string) (*types.ProviderStatus, error) {
	data, err := pm.getProviderData(name, pm.statuses, "status")
	if err != nil {
		return nil, err
	}
	return data.(*types.ProviderStatus), nil
}

// GetAllStatuses returns status for all registered providers
func (pm *ProviderManager) GetAllStatuses() map[string]*types.ProviderStatus {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	result := make(map[string]*types.ProviderStatus)
	for name, status := range pm.statuses {
		result[name] = status
	}

	return result
}

// ValidateProvider validates a provider's credentials and connectivity
func (pm *ProviderManager) ValidateProvider(ctx context.Context, name string) error {
	provider, err := pm.GetProvider(name)
	if err != nil {
		return err
	}

	config, err := pm.GetProviderConfig(name)
	if err != nil {
		return err
	}

	pm.logger.Info(fmt.Sprintf("Validating provider: %s", name))

	err = provider.ValidateCredentials(ctx, config.Credentials)
	if err != nil {
		pm.updateProviderStatus(name, false, err.Error())
		return fmt.Errorf("validation failed for provider %s: %w", name, err)
	}

	pm.updateProviderStatus(name, true, "")
	pm.logger.Info(fmt.Sprintf("Provider %s validated successfully", name))
	return nil
}

// ValidateAllProviders validates all registered providers
func (pm *ProviderManager) ValidateAllProviders(ctx context.Context) map[string]error {
	pm.mutex.RLock()
	providerNames := make([]string, 0, len(pm.providers))
	for name := range pm.providers {
		providerNames = append(providerNames, name)
	}
	pm.mutex.RUnlock()

	results := make(map[string]error)
	for _, name := range providerNames {
		results[name] = pm.ValidateProvider(ctx, name)
	}

	return results
}

// RemoveProvider removes a provider from the manager
func (pm *ProviderManager) RemoveProvider(name string) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	if _, exists := pm.providers[name]; !exists {
		return fmt.Errorf("provider %s not found", name)
	}

	delete(pm.providers, name)
	delete(pm.configs, name)
	delete(pm.statuses, name)

	pm.logger.Info(fmt.Sprintf("Removed provider: %s", name))
	return nil
}

// ListProviders returns a list of all registered provider names and types
func (pm *ProviderManager) ListProviders() map[string]types.ProviderType {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	result := make(map[string]types.ProviderType)
	for name, config := range pm.configs {
		result[name] = config.Type
	}

	return result
}

// CreateProvider creates a new provider instance based on the config
func (pm *ProviderManager) CreateProvider(config *types.ProviderConfig) (types.CloudProvider, error) {
	if regProv, err := registry.Get(string(config.Type), config); err == nil && regProv != nil {
		if cp, ok := regProv.(types.CloudProvider); ok {
			return cp, nil
		}
	}
	// Registry miss: increment fallback metric (should remain zero; switch removed)
	telemetry.ProviderRegistryFallbacks.WithLabelValues(string(config.Type)).Inc()
	return nil, fmt.Errorf("unsupported provider type (registry miss): %s", config.Type)
}

// RegisterNewProvider creates and registers a new provider from config
func (pm *ProviderManager) RegisterNewProvider(config *types.ProviderConfig) error {
	provider, err := pm.CreateProvider(config)
	if err != nil {
		return fmt.Errorf("failed to create provider: %w", err)
	}

	return pm.RegisterProvider(config.Name, provider, config)
}

// updateProviderStatus updates the status of a provider (internal method)
func (pm *ProviderManager) updateProviderStatus(name string, connected bool, errorMsg string) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	if status, exists := pm.statuses[name]; exists {
		status.IsConnected = connected
		status.ErrorMessage = errorMsg
		if connected {
			status.HealthStatus = "healthy"
		} else {
			status.HealthStatus = "error"
		}
	}
}
