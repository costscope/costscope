package types

import "time"

// ConfigProfile represents a configuration profile with environment settings
type ConfigProfile struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Environment string                 `json:"environment"`
	Settings    map[string]interface{} `json:"settings"`
	Active      bool                   `json:"active"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// ConfigValidationResult contains the results of configuration validation
type ConfigValidationResult struct {
	IsValid     bool     `json:"is_valid"`
	Errors      []string `json:"errors"`
	Warnings    []string `json:"warnings"`
	Suggestions []string `json:"suggestions"`
	ConfigPath  string   `json:"config_path"`
	Profile     string   `json:"profile"`
}

// ConfigTemplate represents a configuration template
type ConfigTemplate struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Template    map[string]interface{} `json:"template"`
	Variables   []string               `json:"variables"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}
