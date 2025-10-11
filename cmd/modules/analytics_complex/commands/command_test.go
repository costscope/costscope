//go:build experimental

package commands

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnalyticsComplexCommands(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected string
		wantErr  bool
	}{
		{
			name:     "analytics-complex help",
			args:     []string{"--help"},
			expected: "Advanced Type-Safe Analytics provides enterprise-grade cost analysis",
			wantErr:  false,
		},
		{
			name:     "analytics-complex analyze basic",
			args:     []string{"analyze", "--service", "ec2,rds", "--region", "us-east-1"},
			expected: "", // Command logs to logger, not stdout
			wantErr:  false,
		},
		{
			name:     "analytics-complex forecast",
			args:     []string{"forecast", "--model", "auto-arima", "--days", "30"},
			expected: "", // Command logs to logger, not stdout
			wantErr:  false,
		},
		{
			name:     "analytics-complex detect",
			args:     []string{"detect", "--method", "isolation-forest", "--sensitivity", "high"},
			expected: "", // Command logs to logger, not stdout
			wantErr:  false,
		},
		{
			name:     "analytics-complex transform",
			args:     []string{"transform", "--type", "aggregate", "--target", "cost"},
			expected: "", // Command logs to logger, not stdout
			wantErr:  false,
		},
		{
			name:     "analytics-complex optimize",
			args:     []string{"optimize", "--algorithm", "genetic", "--target", "cost"},
			expected: "", // Command logs to logger, not stdout
			wantErr:  false,
		},
		{
			name:     "analytics-complex custom query",
			args:     []string{"custom", "--query", "SELECT service, SUM(cost) FROM costs GROUP BY service", "--validate"},
			expected: "", // Command logs to logger, not stdout
			wantErr:  false,
		},
		{
			name:     "analytics-complex custom dry-run",
			args:     []string{"custom", "--script", "analytics.py", "--dry-run"},
			expected: "", // Command logs to logger, not stdout
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create command
			acc := NewAnalyticsComplexCommands()
			cmd := acc.BuildAnalyticsComplexCommand()

			// Capture output
			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)

			// Set args
			cmd.SetArgs(tt.args)

			// Execute
			err := cmd.Execute()

			// Check results
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			output := buf.String()
			if tt.expected != "" {
				assert.Contains(t, output, tt.expected)
			}
		})
	}
}

func TestTypeFilterCreation(t *testing.T) {
	t.Run("string slice filter", func(t *testing.T) {
		services := []string{"ec2", "rds", "s3"}
		filter := createFilterValue(services)

		require.NotNil(t, filter)
		assert.Equal(t, services, filter.Value)
		assert.Equal(t, "[]string", filter.Type)
		assert.Equal(t, "eq", filter.Operator)
		assert.True(t, filter.Validated)
	})

	t.Run("float64 filter", func(t *testing.T) {
		threshold := 100.5
		filter := createFilterValue(threshold)

		require.NotNil(t, filter)
		assert.Equal(t, threshold, filter.Value)
		assert.Equal(t, "float64", filter.Type)
		assert.Equal(t, "eq", filter.Operator)
		assert.True(t, filter.Validated)
	})
}

func TestAnalyticsComplexSubcommands(t *testing.T) {
	acc := NewAnalyticsComplexCommands()
	cmd := acc.BuildAnalyticsComplexCommand()

	// Check that all expected subcommands are present
	expectedSubcommands := []string{"analyze", "forecast", "detect", "transform", "optimize"}

	for _, expectedCmd := range expectedSubcommands {
		subCmd, _, err := cmd.Find([]string{expectedCmd})
		require.NoError(t, err, "Subcommand %s should exist", expectedCmd)
		assert.Equal(t, expectedCmd, subCmd.Name())
	}
}

func TestAnalyzeCommandFlags(t *testing.T) {
	acc := NewAnalyticsComplexCommands()
	cmd := acc.buildAnalyzeCommand()

	// Test that required flags are present
	expectedFlags := []string{
		"service", "region", "account", "cost-threshold", "date-range",
		"ml-enabled", "anomaly-detection", "forecast-enabled", "forecast-periods",
		"parallel", "workers", "cache-enabled", "output", "detailed",
	}

	for _, flagName := range expectedFlags {
		flag := cmd.Flags().Lookup(flagName)
		assert.NotNil(t, flag, "Flag %s should exist", flagName)
	}
}

func TestForecastCommandFlags(t *testing.T) {
	acc := NewAnalyticsComplexCommands()
	cmd := acc.buildForecastCommand()

	// Test that required flags are present
	expectedFlags := []string{
		"model", "days", "confidence", "features", "seasonality",
		"uncertainty", "validation", "output",
	}

	for _, flagName := range expectedFlags {
		flag := cmd.Flags().Lookup(flagName)
		assert.NotNil(t, flag, "Flag %s should exist", flagName)
	}

	// Test default values
	assert.Equal(t, "auto-arima", cmd.Flags().Lookup("model").DefValue)
	assert.Equal(t, "30", cmd.Flags().Lookup("days").DefValue)
	assert.Equal(t, "95", cmd.Flags().Lookup("confidence").DefValue)
}

func TestDetectCommandFlags(t *testing.T) {
	acc := NewAnalyticsComplexCommands()
	cmd := acc.buildDetectCommand()

	// Test that required flags are present
	expectedFlags := []string{
		"method", "sensitivity", "real-time", "alerts", "threshold", "window", "output",
	}

	for _, flagName := range expectedFlags {
		flag := cmd.Flags().Lookup(flagName)
		assert.NotNil(t, flag, "Flag %s should exist", flagName)
	}

	// Test default values
	assert.Equal(t, "isolation-forest", cmd.Flags().Lookup("method").DefValue)
	assert.Equal(t, "medium", cmd.Flags().Lookup("sensitivity").DefValue)
	assert.Equal(t, "0.95", cmd.Flags().Lookup("threshold").DefValue)
}

func TestTransformCommandFlags(t *testing.T) {
	acc := NewAnalyticsComplexCommands()
	cmd := acc.buildTransformCommand()

	// Test that required flags are present
	expectedFlags := []string{
		"type", "target", "method", "group-by", "rows", "columns",
		"condition", "optimize-memory", "output",
	}

	for _, flagName := range expectedFlags {
		flag := cmd.Flags().Lookup(flagName)
		assert.NotNil(t, flag, "Flag %s should exist", flagName)
	}

	// Test default values
	assert.Equal(t, "aggregate", cmd.Flags().Lookup("type").DefValue)
	assert.Equal(t, "cost", cmd.Flags().Lookup("target").DefValue)
	assert.Equal(t, "sum", cmd.Flags().Lookup("method").DefValue)
}

func TestOptimizeCommandFlags(t *testing.T) {
	acc := NewAnalyticsComplexCommands()
	cmd := acc.buildOptimizeCommand()

	// Test that required flags are present
	expectedFlags := []string{
		"algorithm", "target", "constraints", "iterations", "tolerance",
		"multi-objective", "output",
	}

	for _, flagName := range expectedFlags {
		flag := cmd.Flags().Lookup(flagName)
		assert.NotNil(t, flag, "Flag %s should exist", flagName)
	}

	// Test default values
	assert.Equal(t, "genetic", cmd.Flags().Lookup("algorithm").DefValue)
	assert.Equal(t, "cost", cmd.Flags().Lookup("target").DefValue)
	assert.Equal(t, "1000", cmd.Flags().Lookup("iterations").DefValue)
}
