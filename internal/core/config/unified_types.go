package config

import (
	"time"
)

// ConfigEnvironment represents the deployment environment
type ConfigEnvironment string

// Environment constants
const (
	Development ConfigEnvironment = "development"
	Staging     ConfigEnvironment = "staging"
	Production  ConfigEnvironment = "production"
	Testing     ConfigEnvironment = "testing"
)

// String returns the string representation of the environment
func (e ConfigEnvironment) String() string {
	return string(e)
}

// ConfigSection represents different configuration sections
type ConfigSection string

// Section constants
const (
	SectionCore      ConfigSection = "core"
	SectionProviders ConfigSection = "providers"
	SectionDatabase  ConfigSection = "database"
	SectionStreaming ConfigSection = "streaming"
	SectionSecurity  ConfigSection = "security"
	// SectionFocus contains FOCUS-related defaults for CLI/API operations
	SectionFocus ConfigSection = "focus"
)

// String returns the string representation of the section
func (s ConfigSection) String() string {
	return string(s)
}

// ConfigStatus represents the current status of the configuration
type ConfigStatus struct {
	Loaded      bool                   `json:"loaded"`
	Environment ConfigEnvironment      `json:"environment"`
	DataDir     string                 `json:"data_dir"`
	Sections    map[ConfigSection]bool `json:"sections"`
	LastError   string                 `json:"last_error,omitempty"`
}

// ConsolidatedConfig represents the complete application configuration
type ConsolidatedConfig struct {
	// Core application settings
	Core      UnifiedCoreConfig      `mapstructure:"core" yaml:"core"`
	Providers UnifiedProvidersConfig `mapstructure:"providers" yaml:"providers"`
	Database  UnifiedDatabaseConfig  `mapstructure:"database" yaml:"database"`
	Streaming UnifiedStreamingConfig `mapstructure:"streaming" yaml:"streaming"`
	Security  UnifiedSecurityConfig  `mapstructure:"security" yaml:"security"`
	Reports   UnifiedReportsConfig   `mapstructure:"reports" yaml:"reports"`
	// MultiTenant contains multi-tenant feature flag & future settings (skeleton)
	MultiTenant UnifiedMultiTenantConfig `mapstructure:"multi_tenant" yaml:"multi_tenant"`
	// Focus contains defaults for FOCUS operations (e.g., conversion)
	Focus UnifiedFocusConfig `mapstructure:"focus" yaml:"focus"`

	// Global settings
	DataDir     string            `mapstructure:"data_dir" yaml:"data_dir"`
	Environment ConfigEnvironment `mapstructure:"environment" yaml:"environment"`
	Version     string            `mapstructure:"version" yaml:"version"`
}

// UnifiedCoreConfig contains core application settings
type UnifiedCoreConfig struct {
	AppName        string        `mapstructure:"app_name" yaml:"app_name"`
	Version        string        `mapstructure:"version" yaml:"version"`
	Environment    string        `mapstructure:"environment" yaml:"environment"`
	LogLevel       string        `mapstructure:"log_level" yaml:"log_level"`
	DataDirectory  string        `mapstructure:"data_directory" yaml:"data_directory"`
	TempDirectory  string        `mapstructure:"temp_directory" yaml:"temp_directory"`
	MaxConcurrency int           `mapstructure:"max_concurrency" yaml:"max_concurrency"`
	Timeout        time.Duration `mapstructure:"timeout" yaml:"timeout"`
	MetricsEnabled bool          `mapstructure:"metrics_enabled" yaml:"metrics_enabled"`
}

// UnifiedProvidersConfig contains cloud provider settings
type UnifiedProvidersConfig struct {
	AWS   AWSProviderConfig   `mapstructure:"aws" yaml:"aws"`
	Azure AzureProviderConfig `mapstructure:"azure" yaml:"azure"`
	GCP   GCPProviderConfig   `mapstructure:"gcp" yaml:"gcp"`
}

// AWSProviderConfig contains AWS-specific settings
type AWSProviderConfig struct {
	Region         string            `mapstructure:"region" yaml:"region"`
	Profile        string            `mapstructure:"profile" yaml:"profile"`
	AccessKey      string            `mapstructure:"access_key" yaml:"access_key"`
	SecretKey      string            `mapstructure:"secret_key" yaml:"secret_key"`
	SessionToken   string            `mapstructure:"session_token" yaml:"session_token"`
	RoleARN        string            `mapstructure:"role_arn" yaml:"role_arn"`
	ExternalID     string            `mapstructure:"external_id" yaml:"external_id"`
	MaxRetries     int               `mapstructure:"max_retries" yaml:"max_retries"`
	RequestTimeout time.Duration     `mapstructure:"request_timeout" yaml:"request_timeout"`
	Enabled        bool              `mapstructure:"enabled" yaml:"enabled"`
	DefaultTags    map[string]string `mapstructure:"default_tags" yaml:"default_tags"`
}

// AzureProviderConfig contains Azure-specific settings
type AzureProviderConfig struct {
	TenantID        string        `mapstructure:"tenant_id" yaml:"tenant_id"`
	ClientID        string        `mapstructure:"client_id" yaml:"client_id"`
	ClientSecret    string        `mapstructure:"client_secret" yaml:"client_secret"`
	SubscriptionID  string        `mapstructure:"subscription_id" yaml:"subscription_id"`
	Environment     string        `mapstructure:"environment" yaml:"environment"`
	RequestTimeout  time.Duration `mapstructure:"request_timeout" yaml:"request_timeout"`
	Enabled         bool          `mapstructure:"enabled" yaml:"enabled"`
	DefaultLocation string        `mapstructure:"default_location" yaml:"default_location"`
}

// GCPProviderConfig contains GCP-specific settings
type GCPProviderConfig struct {
	ProjectID             string            `mapstructure:"project_id" yaml:"project_id"`
	ServiceAccountKeyPath string            `mapstructure:"service_account_key_path" yaml:"service_account_key_path"`
	ServiceAccountKeyJSON string            `mapstructure:"service_account_key_json" yaml:"service_account_key_json"`
	Region                string            `mapstructure:"region" yaml:"region"`
	Zone                  string            `mapstructure:"zone" yaml:"zone"`
	RequestTimeout        time.Duration     `mapstructure:"request_timeout" yaml:"request_timeout"`
	Enabled               bool              `mapstructure:"enabled" yaml:"enabled"`
	DefaultLabels         map[string]string `mapstructure:"default_labels" yaml:"default_labels"`
}

// UnifiedDatabaseConfig contains database settings
type UnifiedDatabaseConfig struct {
	Type               string        `mapstructure:"type" yaml:"type"`
	ConnectionString   string        `mapstructure:"connection_string" yaml:"connection_string"`
	MaxConnections     int           `mapstructure:"max_connections" yaml:"max_connections"`
	MaxIdleConnections int           `mapstructure:"max_idle_connections" yaml:"max_idle_connections"`
	ConnectionTimeout  time.Duration `mapstructure:"connection_timeout" yaml:"connection_timeout"`
	QueryTimeout       time.Duration `mapstructure:"query_timeout" yaml:"query_timeout"`
	MigrationsPath     string        `mapstructure:"migrations_path" yaml:"migrations_path"`
	AutoMigrate        bool          `mapstructure:"auto_migrate" yaml:"auto_migrate"`
}

// UnifiedStreamingConfig contains streaming job settings
type UnifiedStreamingConfig struct {
	MaxConcurrentJobs int           `mapstructure:"max_concurrent_jobs" yaml:"max_concurrent_jobs"`
	DefaultWorkers    int           `mapstructure:"default_workers" yaml:"default_workers"`
	DefaultMemory     int           `mapstructure:"default_memory" yaml:"default_memory"`
	JobTimeout        time.Duration `mapstructure:"job_timeout" yaml:"job_timeout"`
	ProgressInterval  time.Duration `mapstructure:"progress_interval" yaml:"progress_interval"`
	CheckpointEnabled bool          `mapstructure:"checkpoint_enabled" yaml:"checkpoint_enabled"`
	CheckpointDir     string        `mapstructure:"checkpoint_dir" yaml:"checkpoint_dir"`
}

// UnifiedSecurityConfig contains security settings
type UnifiedSecurityConfig struct {
	EncryptionEnabled bool          `mapstructure:"encryption_enabled" yaml:"encryption_enabled"`
	EncryptionKey     string        `mapstructure:"encryption_key" yaml:"encryption_key"`
	TLSEnabled        bool          `mapstructure:"tls_enabled" yaml:"tls_enabled"`
	TLSCertPath       string        `mapstructure:"tls_cert_path" yaml:"tls_cert_path"`
	TLSKeyPath        string        `mapstructure:"tls_key_path" yaml:"tls_key_path"`
	TokenExpiry       time.Duration `mapstructure:"token_expiry" yaml:"token_expiry"`
	APIKeyRequired    bool          `mapstructure:"api_key_required" yaml:"api_key_required"`
	JWTSecret         string        `mapstructure:"jwt_secret" yaml:"jwt_secret"`
}

// UnifiedReportsConfig contains report generation/export defaults
type UnifiedReportsConfig struct {
	// OutputDir is the base directory/prefix for generated reports when an explicit output path is not provided.
	// Supports local filesystem paths as well as object storage prefixes (s3://, gs://). For object storage
	// prefixes no local directory creation is attempted.
	OutputDir string `mapstructure:"output_dir" yaml:"output_dir"`
	// MetadataStorePath optionally enables durable export metadata persistence (JSON lines file). When empty the
	// report service uses only in-memory metadata (lost on restart). Supports local filesystem paths only. This is an
	// MVP persistence layer; future adapters (e.g. sqlite/postgres) may supersede it while keeping field optional.
	MetadataStorePath string `mapstructure:"metadata_store_path" yaml:"metadata_store_path"`
	// MetadataRetentionMaxRecords optionally caps stored metadata records. 0 = unlimited.
	MetadataRetentionMaxRecords int `mapstructure:"metadata_retention_max_records" yaml:"metadata_retention_max_records"`
	// MetadataRetentionMaxAge defines maximum age (e.g. 168h, 720h) before pruning. Zero duration = disabled.
	MetadataRetentionMaxAge time.Duration `mapstructure:"metadata_retention_max_age" yaml:"metadata_retention_max_age"`
}

// UnifiedFocusConfig contains FOCUS module defaults
type UnifiedFocusConfig struct {
	// UseUnifiedMapperDefault controls default mapping engine for convert operations when not specified by CLI/API
	UseUnifiedMapperDefault bool `mapstructure:"use_unified_mapper_default" yaml:"use_unified_mapper_default"`
	// EngineForecastDaysDefault provides an optional YAML default for the FOCUS analysis engine forecast horizon
	// when the CLI flag and ENV variable (COSTSCOPE_FOCUS_ENGINE_FORECAST_DAYS) are not supplied. Zero = fallback.
	EngineForecastDaysDefault int `mapstructure:"engine_forecast_days_default" yaml:"engine_forecast_days_default"`
	// EnginePhaseTimeout sets a soft per-phase timeout (e.g. "1500ms", "2s") for extended phases when not
	// provided via flag or env (COSTSCOPE_FOCUS_ENGINE_TIMEOUT). Zero = fallback (2s currently in resolver).
	EnginePhaseTimeout time.Duration `mapstructure:"engine_phase_timeout" yaml:"engine_phase_timeout"`
	// InvariantsEnabledDefault allows enabling streaming invariants collection globally without passing --invariants.
	// Precedence: explicit flag > YAML (this field) > ENV (COSTSCOPE_INVARIANTS_ENABLED) > default(false).
	InvariantsEnabledDefault bool `mapstructure:"invariants_enabled_default" yaml:"invariants_enabled_default"`
	// InvariantsToleranceDefault supplies a default relative tolerance for drift comparisons (e.g. 0.01 = 1%) when
	// --invariants-tolerance not provided. Precedence: explicit flag > YAML > ENV (COSTSCOPE_INVARIANTS_TOLERANCE) > 0.01 fallback.
	InvariantsToleranceDefault float64 `mapstructure:"invariants_tolerance_default" yaml:"invariants_tolerance_default"`
	// InvariantsBaselineDefault optionally specifies a baseline JSON path used when --invariants-baseline not passed.
	// Precedence: explicit flag > YAML > ENV (COSTSCOPE_INVARIANTS_BASELINE) > empty.
	InvariantsBaselineDefault string `mapstructure:"invariants_baseline_default" yaml:"invariants_baseline_default"`
}

// UnifiedMultiTenantConfig contains skeleton multi-tenancy feature gating.
// TASK-MULTITENANT-SKEL: This is an initial placeholder; future phases will expand with
// tenant scoping, isolation policies, and quota controls (see ADR TODO file).
type UnifiedMultiTenantConfig struct {
	Enabled bool `mapstructure:"enabled" yaml:"enabled"`
	// MaxJobsPerTenant caps the total number of jobs ever created for a tenant during
	// the current process lifetime (0 = unlimited). Future phases may persist counts.
	MaxJobsPerTenant int `mapstructure:"max_jobs_per_tenant" yaml:"max_jobs_per_tenant"`
	// MaxActiveJobsPerTenant caps concurrently active (pending+running) jobs per tenant
	// (0 = unlimited).
	MaxActiveJobsPerTenant int `mapstructure:"max_active_jobs_per_tenant" yaml:"max_active_jobs_per_tenant"`
}
