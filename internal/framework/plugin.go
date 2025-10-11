// Package framework - Plugin Loader for extensible plugin architecture
package framework

import (
	"context"
	"fmt"
	"path/filepath"
	"plugin"
	"sync"
)

// Plugin represents a loadable plugin
type Plugin interface {
	Name() string
	Version() string
	Initialize(ctx context.Context, container *Container) error
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Health() string
}

// PluginInfo contains metadata about a plugin
type PluginInfo struct {
	Name        string                 `json:"name"`
	Version     string                 `json:"version"`
	Description string                 `json:"description"`
	Author      string                 `json:"author"`
	Metadata    map[string]interface{} `json:"metadata"`
	FilePath    string                 `json:"file_path"`
}

// LoadedPlugin represents a loaded plugin instance
type LoadedPlugin struct {
	Info     *PluginInfo
	Instance Plugin
	Plugin   *plugin.Plugin
	Loaded   bool
	Started  bool
}

// PluginLoader manages plugin loading and lifecycle
type PluginLoader struct {
	plugins     map[string]*LoadedPlugin
	pluginPaths []string
	mu          sync.RWMutex
	container   *Container
}

// NewPluginLoader creates a new plugin loader
func NewPluginLoader() *PluginLoader {
	return &PluginLoader{
		plugins:     make(map[string]*LoadedPlugin),
		pluginPaths: []string{"./plugins", "/opt/costscope/plugins"},
	}
}

// AddPluginPath adds a search path for plugins
func (pl *PluginLoader) AddPluginPath(path string) {
	pl.mu.Lock()
	defer pl.mu.Unlock()
	pl.pluginPaths = append(pl.pluginPaths, path)
}

// LoadPlugins loads all available plugins
func (pl *PluginLoader) LoadPlugins(ctx context.Context, container *Container) error {
	pl.mu.Lock()
	defer pl.mu.Unlock()

	pl.container = container

	// For now, register built-in plugins
	if err := pl.loadBuiltinPlugins(ctx); err != nil {
		return fmt.Errorf("failed to load builtin plugins: %w", err)
	}

	// TODO: Scan plugin directories and load .so files
	// This would be implemented for production use

	return nil
}

// loadBuiltinPlugins loads built-in plugins
func (pl *PluginLoader) loadBuiltinPlugins(ctx context.Context) error {
	// Register core analytics plugin
	analyticsPlugin := &AnalyticsPlugin{}
	if err := pl.registerPlugin(ctx, analyticsPlugin); err != nil {
		return fmt.Errorf("failed to register analytics plugin: %w", err)
	}

	// Register reporting plugin
	reportingPlugin := &ReportingPlugin{}
	if err := pl.registerPlugin(ctx, reportingPlugin); err != nil {
		return fmt.Errorf("failed to register reporting plugin: %w", err)
	}

	// Register monitoring plugin
	monitoringPlugin := &MonitoringPlugin{}
	if err := pl.registerPlugin(ctx, monitoringPlugin); err != nil {
		return fmt.Errorf("failed to register monitoring plugin: %w", err)
	}

	return nil
}

// registerPlugin registers a plugin instance
func (pl *PluginLoader) registerPlugin(ctx context.Context, pluginInstance Plugin) error {
	name := pluginInstance.Name()

	loadedPlugin := &LoadedPlugin{
		Info: &PluginInfo{
			Name:        name,
			Version:     pluginInstance.Version(),
			Description: fmt.Sprintf("Built-in %s plugin", name),
			Author:      "CostScope Team",
			Metadata:    make(map[string]interface{}),
		},
		Instance: pluginInstance,
		Loaded:   true,
	}

	// Initialize plugin
	if err := pluginInstance.Initialize(ctx, pl.container); err != nil {
		return fmt.Errorf("failed to initialize plugin '%s': %w", name, err)
	}

	// Start plugin
	if err := pluginInstance.Start(ctx); err != nil {
		return fmt.Errorf("failed to start plugin '%s': %w", name, err)
	}

	loadedPlugin.Started = true
	pl.plugins[name] = loadedPlugin

	return nil
}

// LoadPlugin loads a specific plugin by path
func (pl *PluginLoader) LoadPlugin(ctx context.Context, pluginPath string) error {
	pl.mu.Lock()
	defer pl.mu.Unlock()

	// Load the plugin file
	p, err := plugin.Open(pluginPath)
	if err != nil {
		return fmt.Errorf("failed to open plugin '%s': %w", pluginPath, err)
	}

	// Look for the NewPlugin symbol
	newPluginSymbol, err := p.Lookup("NewPlugin")
	if err != nil {
		return fmt.Errorf("plugin '%s' does not export NewPlugin function: %w", pluginPath, err)
	}

	// Cast to function
	newPlugin, ok := newPluginSymbol.(func() Plugin)
	if !ok {
		return fmt.Errorf("NewPlugin in '%s' has wrong signature", pluginPath)
	}

	// Create plugin instance
	pluginInstance := newPlugin()
	name := pluginInstance.Name()

	loadedPlugin := &LoadedPlugin{
		Info: &PluginInfo{
			Name:     name,
			Version:  pluginInstance.Version(),
			FilePath: pluginPath,
			Metadata: make(map[string]interface{}),
		},
		Instance: pluginInstance,
		Plugin:   p,
		Loaded:   true,
	}

	// Initialize plugin
	if err := pluginInstance.Initialize(ctx, pl.container); err != nil {
		return fmt.Errorf("failed to initialize plugin '%s': %w", name, err)
	}

	// Start plugin
	if err := pluginInstance.Start(ctx); err != nil {
		return fmt.Errorf("failed to start plugin '%s': %w", name, err)
	}

	loadedPlugin.Started = true
	pl.plugins[name] = loadedPlugin

	return nil
}

// UnloadPlugins unloads all plugins
func (pl *PluginLoader) UnloadPlugins(ctx context.Context) error {
	pl.mu.Lock()
	defer pl.mu.Unlock()

	var errors []error

	for name, loadedPlugin := range pl.plugins {
		if loadedPlugin.Started {
			if err := loadedPlugin.Instance.Stop(ctx); err != nil {
				errors = append(errors, fmt.Errorf("failed to stop plugin '%s': %w", name, err))
			}
		}
	}

	pl.plugins = make(map[string]*LoadedPlugin)

	if len(errors) > 0 {
		return fmt.Errorf("errors during plugin unloading: %v", errors)
	}

	return nil
}

// UnloadPlugin unloads a specific plugin
func (pl *PluginLoader) UnloadPlugin(ctx context.Context, pluginName string) error {
	pl.mu.Lock()
	defer pl.mu.Unlock()

	loadedPlugin, exists := pl.plugins[pluginName]
	if !exists {
		return fmt.Errorf("plugin '%s' not found", pluginName)
	}

	if loadedPlugin.Started {
		if err := loadedPlugin.Instance.Stop(ctx); err != nil {
			return fmt.Errorf("failed to stop plugin '%s': %w", pluginName, err)
		}
	}

	delete(pl.plugins, pluginName)
	return nil
}

// GetPlugin returns a loaded plugin by name
func (pl *PluginLoader) GetPlugin(name string) (Plugin, error) {
	pl.mu.RLock()
	defer pl.mu.RUnlock()

	loadedPlugin, exists := pl.plugins[name]
	if !exists {
		return nil, fmt.Errorf("plugin '%s' not found", name)
	}

	return loadedPlugin.Instance, nil
}

// ListPlugins returns information about all loaded plugins
func (pl *PluginLoader) ListPlugins() []*PluginInfo {
	pl.mu.RLock()
	defer pl.mu.RUnlock()

	plugins := make([]*PluginInfo, 0, len(pl.plugins))
	for _, loadedPlugin := range pl.plugins {
		plugins = append(plugins, loadedPlugin.Info)
	}

	return plugins
}

// IsPluginLoaded checks if a plugin is loaded
func (pl *PluginLoader) IsPluginLoaded(name string) bool {
	pl.mu.RLock()
	defer pl.mu.RUnlock()

	loadedPlugin, exists := pl.plugins[name]
	return exists && loadedPlugin.Loaded
}

// IsPluginStarted checks if a plugin is started
func (pl *PluginLoader) IsPluginStarted(name string) bool {
	pl.mu.RLock()
	defer pl.mu.RUnlock()

	loadedPlugin, exists := pl.plugins[name]
	return exists && loadedPlugin.Started
}

// Health returns the health status of the plugin loader
func (pl *PluginLoader) Health() string {
	pl.mu.RLock()
	defer pl.mu.RUnlock()

	totalPlugins := len(pl.plugins)
	if totalPlugins == 0 {
		return "empty"
	}

	unhealthyCount := 0
	for _, loadedPlugin := range pl.plugins {
		if !loadedPlugin.Started || loadedPlugin.Instance.Health() != "healthy" {
			unhealthyCount++
		}
	}

	if unhealthyCount == 0 {
		return HealthStatusHealthy
	} else if unhealthyCount < totalPlugins {
		return HealthStatusDegraded
	} else {
		return "unhealthy"
	}
}

// Stats returns statistics about loaded plugins
func (pl *PluginLoader) Stats() map[string]interface{} {
	pl.mu.RLock()
	defer pl.mu.RUnlock()

	loaded := 0
	started := 0
	healthy := 0

	for _, loadedPlugin := range pl.plugins {
		if loadedPlugin.Loaded {
			loaded++
		}
		if loadedPlugin.Started {
			started++
		}
		if loadedPlugin.Instance.Health() == "healthy" {
			healthy++
		}
	}

	return map[string]interface{}{
		"total_plugins":   len(pl.plugins),
		"loaded_plugins":  loaded,
		"started_plugins": started,
		"healthy_plugins": healthy,
		"plugin_paths":    pl.pluginPaths,
	}
}

// DiscoverPlugins scans plugin directories for available plugins
func (pl *PluginLoader) DiscoverPlugins() ([]*PluginInfo, error) {
	var plugins []*PluginInfo

	for _, path := range pl.pluginPaths {
		matches, err := filepath.Glob(filepath.Join(path, "*.so"))
		if err != nil {
			continue
		}

		for _, match := range matches {
			info := &PluginInfo{
				Name:     filepath.Base(match),
				FilePath: match,
				Metadata: make(map[string]interface{}),
			}
			plugins = append(plugins, info)
		}
	}

	return plugins, nil
}
