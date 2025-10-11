package common

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"local/costscope/internal/core/multicloud"
)

// ParseDateRange parses start and end dates in YYYY-MM-DD format.
// Defaults: start = 1 month ago, end = today when empty.
func ParseDateRange(startDate, endDate string) (time.Time, time.Time, error) {
	if startDate == "" {
		startDate = time.Now().AddDate(0, -1, 0).Format("2006-01-02")
	}
	if endDate == "" {
		endDate = time.Now().Format("2006-01-02")
	}

	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid start date: %w", err)
	}

	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid end date: %w", err)
	}

	if start.After(end) {
		return time.Time{}, time.Time{}, fmt.Errorf("start date must be before end date")
	}

	return start, end, nil
}

// LoadMulticloudConfig returns default config when path is empty, otherwise reads the file.
// Security: validates path to avoid traversal and only allows .json/.yaml/.yml suffixes.
// Note: YAML suffix is accepted but the current implementation uses JSON unmarshal to preserve existing behavior.
func LoadMulticloudConfig(configFile string) (*multicloud.MulticloudConfig, error) {
	if configFile == "" {
		return &multicloud.MulticloudConfig{
			DefaultCurrency:      "USD",
			DefaultTimeout:       30 * time.Minute,
			MaxConcurrentScans:   5,
			CacheEnabled:         true,
			CacheTTL:             1 * time.Hour,
			EnableOptimizations:  true,
			EnableMigrations:     true,
			DefaultRiskTolerance: multicloud.RiskLevelMedium,
		}, nil
	}

	// Validate config file path to prevent path traversal attacks
	if strings.Contains(configFile, "..") || (!strings.HasSuffix(configFile, ".json") && !strings.HasSuffix(configFile, ".yaml") && !strings.HasSuffix(configFile, ".yml")) {
		return nil, fmt.Errorf("invalid config file path: %s", configFile)
	}

	data, err := os.ReadFile(filepath.Clean(configFile)) // #nosec G304 - path is validated above
	if err != nil {
		return nil, err
	}

	var config multicloud.MulticloudConfig
	if err := json.Unmarshal(data, &config); err != nil { // preserve prior behavior
		return nil, err
	}

	return &config, nil
}
