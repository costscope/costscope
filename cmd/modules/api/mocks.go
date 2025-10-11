package api

import (
	"local/costscope/internal/core/logging"
)

// =====================================================================================
// Mock Managers - Temporary implementations for API development
// =====================================================================================

// MockConversionManager is a mock implementation of conversion manager
type MockConversionManager struct {
	logger *logging.Logger
}

// NewMockConversionManager creates a new mock conversion manager
func NewMockConversionManager(logger *logging.Logger) *MockConversionManager {
	return &MockConversionManager{
		logger: logger,
	}
}

// MockAnalysisManager is a mock implementation of analysis manager
type MockAnalysisManager struct {
	logger *logging.Logger
}

// NewMockAnalysisManager creates a new mock analysis manager
func NewMockAnalysisManager(logger *logging.Logger) *MockAnalysisManager {
	return &MockAnalysisManager{
		logger: logger,
	}
}

// MockComparisonManager is a mock implementation of comparison manager
type MockComparisonManager struct {
	logger *logging.Logger
}

// NewMockComparisonManager creates a new mock comparison manager
func NewMockComparisonManager(logger *logging.Logger) *MockComparisonManager {
	return &MockComparisonManager{
		logger: logger,
	}
}

// MockValidationManager is a mock implementation of validation manager
type MockValidationManager struct {
	logger *logging.Logger
}

// NewMockValidationManager creates a new mock validation manager
func NewMockValidationManager(logger *logging.Logger) *MockValidationManager {
	return &MockValidationManager{
		logger: logger,
	}
}
