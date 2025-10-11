package logging

import (
	"testing"
)

func TestNewLogger(t *testing.T) {
	logger := NewLogger(LevelInfo)
	if logger == nil {
		t.Fatal("NewLogger should not return nil")
	}
	if logger.level != LevelInfo {
		t.Errorf("Expected level %s, got %s", LevelInfo, logger.level)
	}
}

func TestIsEnabledFor(t *testing.T) {
	logger := NewLogger(LevelWarn)

	tests := []struct {
		level    LogLevel
		expected bool
	}{
		{LevelDebug, false},
		{LevelInfo, false},
		{LevelWarn, true},
		{LevelError, true},
		{LevelFatal, true},
	}

	for _, test := range tests {
		result := logger.IsEnabledFor(test.level)
		if result != test.expected {
			t.Errorf("IsEnabledFor(%s) = %v, expected %v", test.level, result, test.expected)
		}
	}
}

func TestLogLevels(t *testing.T) {
	logger := NewLogger(LevelDebug)

	// These should not panic
	logger.Debug("debug message")
	logger.Info("info message")
	logger.Warn("warn message")
	logger.Error("error message")

	// Test that these methods exist and can be called
	if logger.IsEnabledFor(LevelDebug) != true {
		t.Error("Debug logger should be enabled for debug level")
	}
}
