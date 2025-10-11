// Package types defines common types and data structures for CostScope framework
package types

import (
	"time"
)

// NOTE: Generic Result / PaginatedResult helper factories (NewResult/NewError/NewPaginatedResult)
// were removed as they were unused across the codebase and duplicated domain‑specific
// response patterns (reports/integration/focus/etc). This keeps a single source of
// truth for API / CLI response shapes and avoids drifting factories. If a unified
// response wrapper becomes necessary later, introduce it under a dedicated
// internal/api/response package and migrate handlers explicitly.

// Config represents configuration data
type Config struct {
	Database   DatabaseConfig   `json:"database" yaml:"database"`
	Server     ServerConfig     `json:"server" yaml:"server"`
	Logging    LoggingConfig    `json:"logging" yaml:"logging"`
	Plugins    PluginsConfig    `json:"plugins" yaml:"plugins"`
	Security   SecurityConfig   `json:"security" yaml:"security"`
	Features   FeaturesConfig   `json:"features" yaml:"features"`
	Providers  ProvidersConfig  `json:"providers" yaml:"providers"`
	Monitoring MonitoringConfig `json:"monitoring" yaml:"monitoring"`
}

// DatabaseConfig represents database configuration
type DatabaseConfig struct {
	Type        string `json:"type" yaml:"type"`
	Host        string `json:"host" yaml:"host"`
	Port        int    `json:"port" yaml:"port"`
	Database    string `json:"database" yaml:"database"`
	Username    string `json:"username" yaml:"username"`
	Password    string `json:"password" yaml:"password"`
	SSLMode     string `json:"ssl_mode" yaml:"ssl_mode"`
	MaxConns    int    `json:"max_conns" yaml:"max_conns"`
	MaxIdleTime string `json:"max_idle_time" yaml:"max_idle_time"`
}

// ServerConfig represents server configuration
type ServerConfig struct {
	Host         string        `json:"host" yaml:"host"`
	Port         int           `json:"port" yaml:"port"`
	ReadTimeout  time.Duration `json:"read_timeout" yaml:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout" yaml:"write_timeout"`
	IdleTimeout  time.Duration `json:"idle_timeout" yaml:"idle_timeout"`
	TLS          TLSConfig     `json:"tls" yaml:"tls"`
}

// TLSConfig represents TLS configuration
type TLSConfig struct {
	Enabled  bool   `json:"enabled" yaml:"enabled"`
	CertFile string `json:"cert_file" yaml:"cert_file"`
	KeyFile  string `json:"key_file" yaml:"key_file"`
}

// LoggingConfig represents logging configuration
type LoggingConfig struct {
	Level      string `json:"level" yaml:"level"`
	Format     string `json:"format" yaml:"format"`
	Output     string `json:"output" yaml:"output"`
	MaxSize    int    `json:"max_size" yaml:"max_size"`
	MaxBackups int    `json:"max_backups" yaml:"max_backups"`
	MaxAge     int    `json:"max_age" yaml:"max_age"`
	Compress   bool   `json:"compress" yaml:"compress"`
}

// PluginsConfig represents plugins configuration
type PluginsConfig struct {
	Directory string                 `json:"directory" yaml:"directory"`
	AutoLoad  bool                   `json:"auto_load" yaml:"auto_load"`
	Enabled   []string               `json:"enabled" yaml:"enabled"`
	Disabled  []string               `json:"disabled" yaml:"disabled"`
	Config    map[string]interface{} `json:"config" yaml:"config"`
}

// SecurityConfig represents security configuration
type SecurityConfig struct {
	Authentication AuthConfig      `json:"authentication" yaml:"authentication"`
	Authorization  AuthzConfig     `json:"authorization" yaml:"authorization"`
	RateLimit      RateLimitConfig `json:"rate_limit" yaml:"rate_limit"`
	CORS           CORSConfig      `json:"cors" yaml:"cors"`
}

// AuthConfig represents authentication configuration
type AuthConfig struct {
	Type      string                 `json:"type" yaml:"type"`
	Secret    string                 `json:"secret" yaml:"secret"`
	Expiry    time.Duration          `json:"expiry" yaml:"expiry"`
	Providers map[string]interface{} `json:"providers" yaml:"providers"`
}

// AuthzConfig represents authorization configuration
type AuthzConfig struct {
	Type    string      `json:"type" yaml:"type"`
	Default string      `json:"default" yaml:"default"`
	Rules   []AuthzRule `json:"rules" yaml:"rules"`
}

// AuthzRule represents an authorization rule
type AuthzRule struct {
	Resource string   `json:"resource" yaml:"resource"`
	Actions  []string `json:"actions" yaml:"actions"`
	Roles    []string `json:"roles" yaml:"roles"`
	Effect   string   `json:"effect" yaml:"effect"`
}

// RateLimitConfig represents rate limiting configuration
type RateLimitConfig struct {
	Enabled bool   `json:"enabled" yaml:"enabled"`
	Rate    int    `json:"rate" yaml:"rate"`
	Burst   int    `json:"burst" yaml:"burst"`
	Window  string `json:"window" yaml:"window"`
}

// CORSConfig represents CORS configuration
type CORSConfig struct {
	Enabled     bool     `json:"enabled" yaml:"enabled"`
	Origins     []string `json:"origins" yaml:"origins"`
	Methods     []string `json:"methods" yaml:"methods"`
	Headers     []string `json:"headers" yaml:"headers"`
	Credentials bool     `json:"credentials" yaml:"credentials"`
}

// FeaturesConfig represents feature flags configuration
type FeaturesConfig struct {
	Analytics    bool `json:"analytics" yaml:"analytics"`
	Reporting    bool `json:"reporting" yaml:"reporting"`
	Alerts       bool `json:"alerts" yaml:"alerts"`
	Optimization bool `json:"optimization" yaml:"optimization"`
	Forecasting  bool `json:"forecasting" yaml:"forecasting"`
	Export       bool `json:"export" yaml:"export"`
}

// ProvidersConfig represents cloud providers configuration
type ProvidersConfig struct {
	AWS   AWSConfig   `json:"aws" yaml:"aws"`
	Azure AzureConfig `json:"azure" yaml:"azure"`
	GCP   GCPConfig   `json:"gcp" yaml:"gcp"`
}

// AWSConfig represents AWS configuration
type AWSConfig struct {
	Enabled         bool   `json:"enabled" yaml:"enabled"`
	Region          string `json:"region" yaml:"region"`
	AccessKeyID     string `json:"access_key_id" yaml:"access_key_id"`
	SecretAccessKey string `json:"secret_access_key" yaml:"secret_access_key"`
	SessionToken    string `json:"session_token" yaml:"session_token"`
	Profile         string `json:"profile" yaml:"profile"`
	S3Bucket        string `json:"s3_bucket" yaml:"s3_bucket"`
}

// AzureConfig represents Azure configuration
type AzureConfig struct {
	Enabled        bool   `json:"enabled" yaml:"enabled"`
	TenantID       string `json:"tenant_id" yaml:"tenant_id"`
	ClientID       string `json:"client_id" yaml:"client_id"`
	ClientSecret   string `json:"client_secret" yaml:"client_secret"`
	SubscriptionID string `json:"subscription_id" yaml:"subscription_id"`
}

// GCPConfig represents GCP configuration
type GCPConfig struct {
	Enabled     bool   `json:"enabled" yaml:"enabled"`
	ProjectID   string `json:"project_id" yaml:"project_id"`
	Credentials string `json:"credentials" yaml:"credentials"`
	Location    string `json:"location" yaml:"location"`
}

// MonitoringConfig represents monitoring configuration
type MonitoringConfig struct {
	Enabled   bool          `json:"enabled" yaml:"enabled"`
	Interval  time.Duration `json:"interval" yaml:"interval"`
	Endpoints []string      `json:"endpoints" yaml:"endpoints"`
	Alerts    AlertsConfig  `json:"alerts" yaml:"alerts"`
	Metrics   MetricsConfig `json:"metrics" yaml:"metrics"`
}

// AlertsConfig represents alerts configuration
type AlertsConfig struct {
	Enabled  bool            `json:"enabled" yaml:"enabled"`
	Channels []string        `json:"channels" yaml:"channels"`
	Rules    []AlertRule     `json:"rules" yaml:"rules"`
	Webhooks []WebhookConfig `json:"webhooks" yaml:"webhooks"`
}

// AlertRule represents an alert rule
type AlertRule struct {
	Name      string                 `json:"name" yaml:"name"`
	Condition string                 `json:"condition" yaml:"condition"`
	Threshold float64                `json:"threshold" yaml:"threshold"`
	Duration  time.Duration          `json:"duration" yaml:"duration"`
	Severity  string                 `json:"severity" yaml:"severity"`
	Metadata  map[string]interface{} `json:"metadata" yaml:"metadata"`
}

// WebhookConfig represents webhook configuration
type WebhookConfig struct {
	Name    string            `json:"name" yaml:"name"`
	URL     string            `json:"url" yaml:"url"`
	Method  string            `json:"method" yaml:"method"`
	Headers map[string]string `json:"headers" yaml:"headers"`
	Timeout time.Duration     `json:"timeout" yaml:"timeout"`
}

// MetricsConfig represents metrics configuration
type MetricsConfig struct {
	Enabled   bool              `json:"enabled" yaml:"enabled"`
	Namespace string            `json:"namespace" yaml:"namespace"`
	Labels    map[string]string `json:"labels" yaml:"labels"`
	Export    ExportConfig      `json:"export" yaml:"export"`
}

// ExportConfig represents metrics export configuration
type ExportConfig struct {
	Type     string            `json:"type" yaml:"type"`
	Endpoint string            `json:"endpoint" yaml:"endpoint"`
	Interval time.Duration     `json:"interval" yaml:"interval"`
	Timeout  time.Duration     `json:"timeout" yaml:"timeout"`
	Headers  map[string]string `json:"headers" yaml:"headers"`
}

// CostData represents cost analysis data
type CostData struct {
	Provider string                 `json:"provider"`
	Account  string                 `json:"account"`
	Service  string                 `json:"service"`
	Region   string                 `json:"region"`
	Resource string                 `json:"resource"`
	Amount   float64                `json:"amount"`
	Currency string                 `json:"currency"`
	Date     time.Time              `json:"date"`
	Tags     map[string]string      `json:"tags"`
	Metadata map[string]interface{} `json:"metadata"`
}

// UsageData represents usage metrics data
type UsageData struct {
	Provider   string            `json:"provider"`
	Service    string            `json:"service"`
	Resource   string            `json:"resource"`
	Metric     string            `json:"metric"`
	Value      float64           `json:"value"`
	Unit       string            `json:"unit"`
	Timestamp  time.Time         `json:"timestamp"`
	Tags       map[string]string `json:"tags"`
	Dimensions map[string]string `json:"dimensions"`
}

// Recommendation represents an optimization recommendation
type Recommendation struct {
	ID          string            `json:"id"`
	Type        string            `json:"type"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Impact      ImpactData        `json:"impact"`
	Confidence  float64           `json:"confidence"`
	Resource    ResourceInfo      `json:"resource"`
	Actions     []Action          `json:"actions"`
	CreatedAt   time.Time         `json:"created_at"`
	ExpiresAt   time.Time         `json:"expires_at"`
	Tags        map[string]string `json:"tags"`
}

// ImpactData represents the impact of a recommendation
type ImpactData struct {
	CostSavings float64 `json:"cost_savings"`
	Currency    string  `json:"currency"`
	Period      string  `json:"period"`
	Percentage  float64 `json:"percentage"`
	RiskLevel   string  `json:"risk_level"`
}

// ResourceInfo represents resource information
type ResourceInfo struct {
	Provider   string                 `json:"provider"`
	Type       string                 `json:"type"`
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Region     string                 `json:"region"`
	Account    string                 `json:"account"`
	Tags       map[string]string      `json:"tags"`
	Properties map[string]interface{} `json:"properties"`
}

// Action represents a recommended action
type Action struct {
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
	Automated   bool                   `json:"automated"`
	RiskLevel   string                 `json:"risk_level"`
}
