package config

import (
	"testing"
)

func TestNewConfigManager(t *testing.T) {
	cm := NewConfigManager()
	if cm == nil {
		t.Error("NewConfigManager should not return nil")
	}
}

func TestConfigManagerGetVersion(t *testing.T) {
	cm := NewConfigManager()
	version := cm.GetVersion()
	if version == "" {
		t.Error("GetVersion should not return empty string")
	}
	expected := "1.0.0"
	if version != expected {
		t.Errorf("Expected version %s, got %s", expected, version)
	}
}

func TestConfigManagerValidate(t *testing.T) {
	cm := NewConfigManager()
	err := cm.Validate()
	if err != nil {
		t.Errorf("Validate should not return error, got: %v", err)
	}
}
