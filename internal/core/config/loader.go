package config

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// ConfigLoader handles loading configuration from files and environment variables
type ConfigLoader struct {
	logger Logger
}

// NewConfigLoader creates a new configuration loader
func NewConfigLoader(logger Logger) *ConfigLoader {
	return &ConfigLoader{
		logger: logger,
	}
}

// LoadFromFile loads configuration from a YAML file
func (cl *ConfigLoader) LoadFromFile(configPath string) (*ConsolidatedConfig, error) {
	cl.logger.Infof("Loading configuration from file: %s", configPath)

	// Check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, NewConfigError("", "file", fmt.Sprintf("config file not found: %s", configPath), err)
	}

	// Validate config path to prevent directory traversal attacks
	cleanPath := filepath.Clean(configPath)
	if strings.Contains(cleanPath, "..") || !filepath.IsAbs(cleanPath) {
		return nil, NewConfigError("", "file", "invalid config path", nil)
	}

	// Read file
	data, err := os.ReadFile(cleanPath) // #nosec G304 - path validated above
	if err != nil {
		return nil, NewConfigError("", "file", fmt.Sprintf("failed to read config file: %s", cleanPath), err)
	}

	// Parse YAML strictly to detect unknown fields
	var config ConsolidatedConfig
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	if err := dec.Decode(&config); err != nil {
		return nil, NewConfigError("", "parse", "failed to parse YAML config", err)
	}

	cl.logger.Infof("Configuration loaded from file successfully")
	return &config, nil
}

// LoadFromEnv loads configuration from environment variables
// NOTE: Bulk ENV loading removed; precedence-based per-field resolution supersedes this logic.
