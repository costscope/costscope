package config

// EnvResolver has been fully superseded by Resolve*Field helpers (resolve_field.go).
// This file now only retains the optional YAML loader shim reused by helpers.

import (
	"local/costscope/internal/core/logging"
	"os"
	"path/filepath"
)

// LoadOptionalYAML loads YAML configuration from COSTSCOPE_CONFIG or ~/.costscope/config.yaml if present.
// Returns nil if not found or invalid. Used by Resolve*Field helpers.
func LoadOptionalYAML(logger *logging.Logger) *ConsolidatedConfig {
	cfgPath := os.Getenv("COSTSCOPE_CONFIG")
	if cfgPath == "" { // fallback to default path
		if home, err := os.UserHomeDir(); err == nil && home != "" {
			cfgPath = filepath.Join(home, ".costscope", "config.yaml")
		}
	}
	if cfgPath == "" || !filepath.IsAbs(cfgPath) {
		return nil
	}
	loader := NewConfigLoader(&resolverLoaderShim{l: logger})
	if cfg, err := loader.LoadFromFile(cfgPath); err == nil && cfg != nil {
		return cfg
	}
	return nil
}

// resolverLoaderShim adapts the unified logger to the config.Logger interface.
type resolverLoaderShim struct{ l *logging.Logger }

func (s *resolverLoaderShim) Infof(format string, args ...interface{})  { s.l.Infof(format, args...) }
func (s *resolverLoaderShim) Errorf(format string, args ...interface{}) { s.l.Errorf(format, args...) }
func (s *resolverLoaderShim) Debugf(format string, args ...interface{}) { s.l.Debugf(format, args...) }
func (s *resolverLoaderShim) Warnf(format string, args ...interface{})  { s.l.Warnf(format, args...) }
