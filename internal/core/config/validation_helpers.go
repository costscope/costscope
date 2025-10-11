package config

// Canonical stateless configuration validation entrypoints.
// This file supersedes the legacy object-based ConfigValidator (removed during
// early development) and is the single supported path going forward:
//   1) ValidateAllConfig(cfg)    – pure validation, no side effects
//   2) EnsureConfigDirectories() – explicit, idempotent directory creation
// Section validators (*ValidateX*) remain exported while they are still
// directly invoked in some startup / test flows; if future refactors remove
// such direct calls they can be unexported (grep before changing). No
// // Deprecated or //nolint markers are required because legacy symbols were
// fully removed (internal-only) and public CLI/API surfaces were unaffected.
// Rationale: pure functions = simpler tests, deterministic ordering, and
// explicit side‑effects.

import (
	"fmt"
	"os"
	"strings"
)

// ValidateAllConfig runs all section validators and global checks in deterministic order.
// It does NOT create directories; callers may follow with EnsureConfigDirectories if desired.
func ValidateAllConfig(cfg *ConsolidatedConfig) error {
	if cfg == nil {
		return NewConfigError("", "", "configuration is nil", nil)
	}
	if err := ValidateGlobalSettings(cfg); err != nil {
		return err
	}
	if err := ValidateCore(&cfg.Core); err != nil {
		return err
	}
	if err := ValidateProviders(&cfg.Providers); err != nil {
		return err
	}
	if err := ValidateDatabase(&cfg.Database); err != nil {
		return err
	}
	if err := ValidateStreaming(&cfg.Streaming); err != nil {
		return err
	}
	if err := ValidateSecurity(&cfg.Security); err != nil {
		return err
	}
	if err := ValidateFocus(&cfg.Focus); err != nil {
		return err
	}
	return nil
}

// EnsureConfigDirectories performs idempotent directory creation for paths previously validated.
func EnsureConfigDirectories(cfg *ConsolidatedConfig) error {
	if cfg == nil {
		return NewConfigError("", "", "configuration is nil", nil)
	}
	if cfg.DataDir != "" {
		if err := os.MkdirAll(cfg.DataDir, 0o750); err != nil {
			return NewConfigError("", "data_dir", fmt.Sprintf("failed to create data directory: %s", cfg.DataDir), err)
		}
	}
	if d := cfg.Core.DataDirectory; d != "" {
		if err := os.MkdirAll(d, 0o750); err != nil {
			return NewConfigError(SectionCore, "data_directory", fmt.Sprintf("failed to create data directory: %s", d), err)
		}
	}
	if t := cfg.Core.TempDirectory; t != "" {
		if err := os.MkdirAll(t, 0o750); err != nil {
			return NewConfigError(SectionCore, "temp_directory", fmt.Sprintf("failed to create temp directory: %s", t), err)
		}
	}
	if cfg.Streaming.CheckpointEnabled && cfg.Streaming.CheckpointDir != "" {
		if err := os.MkdirAll(cfg.Streaming.CheckpointDir, 0o750); err != nil {
			return NewConfigError(SectionStreaming, "checkpoint_dir", fmt.Sprintf("failed to create checkpoint directory: %s", cfg.Streaming.CheckpointDir), err)
		}
	}
	if mp := cfg.Database.MigrationsPath; mp != "" {
		if err := os.MkdirAll(mp, 0o750); err != nil {
			return NewConfigError(SectionDatabase, "migrations_path", fmt.Sprintf("failed to create migrations directory: %s", mp), err)
		}
	}
	return nil
}

func ValidateGlobalSettings(cfg *ConsolidatedConfig) error {
	validEnv := map[ConfigEnvironment]bool{Development: true, Staging: true, Production: true, Testing: true}
	if !validEnv[cfg.Environment] {
		return NewConfigError("", "environment", fmt.Sprintf("invalid environment: %s", cfg.Environment), nil)
	}
	if strings.TrimSpace(cfg.DataDir) == "" {
		return NewConfigError("", "data_dir", "data directory cannot be empty", nil)
	}
	return nil
}

func ValidateCore(c *UnifiedCoreConfig) error {
	if strings.TrimSpace(c.AppName) == "" {
		return NewConfigError(SectionCore, "app_name", "application name cannot be empty", nil)
	}
	if strings.TrimSpace(c.Version) == "" {
		return NewConfigError(SectionCore, "version", "version cannot be empty", nil)
	}
	validLog := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	if !validLog[strings.ToLower(c.LogLevel)] {
		return NewConfigError(SectionCore, "log_level", fmt.Sprintf("invalid log level: %s", c.LogLevel), nil)
	}
	if c.MaxConcurrency <= 0 {
		return NewConfigError(SectionCore, "max_concurrency", "max concurrency must be greater than 0", nil)
	}
	if c.Timeout <= 0 {
		return NewConfigError(SectionCore, "timeout", "timeout must be greater than 0", nil)
	}
	return nil
}

func ValidateProviders(p *UnifiedProvidersConfig) error {
	if p.AWS.Enabled {
		if err := validateAWS(&p.AWS); err != nil {
			return err
		}
	}
	if p.Azure.Enabled {
		if err := validateAzure(&p.Azure); err != nil {
			return err
		}
	}
	if p.GCP.Enabled {
		if err := validateGCP(&p.GCP); err != nil {
			return err
		}
	}
	return nil
}

func validateAWS(a *AWSProviderConfig) error {
	if strings.TrimSpace(a.Region) == "" {
		return NewConfigError(SectionProviders, "aws.region", "AWS region cannot be empty", nil)
	}
	hasProfile := a.Profile != ""
	hasKeys := a.AccessKey != "" && a.SecretKey != ""
	hasRole := a.RoleARN != ""
	if !hasProfile && !hasKeys && !hasRole {
		return NewConfigError(SectionProviders, "aws.credentials", "AWS requires either profile, access keys, or role ARN", nil)
	}
	if a.MaxRetries < 0 {
		return NewConfigError(SectionProviders, "aws.max_retries", "max retries cannot be negative", nil)
	}
	if a.RequestTimeout <= 0 {
		return NewConfigError(SectionProviders, "aws.request_timeout", "request timeout must be greater than 0", nil)
	}
	return nil
}

func validateAzure(az *AzureProviderConfig) error {
	if az.TenantID == "" {
		return NewConfigError(SectionProviders, "azure.tenant_id", "Azure tenant ID cannot be empty", nil)
	}
	if az.ClientID == "" {
		return NewConfigError(SectionProviders, "azure.client_id", "Azure client ID cannot be empty", nil)
	}
	if az.ClientSecret == "" {
		return NewConfigError(SectionProviders, "azure.client_secret", "Azure client secret cannot be empty", nil)
	}
	if az.SubscriptionID == "" {
		return NewConfigError(SectionProviders, "azure.subscription_id", "Azure subscription ID cannot be empty", nil)
	}
	if az.RequestTimeout <= 0 {
		return NewConfigError(SectionProviders, "azure.request_timeout", "request timeout must be greater than 0", nil)
	}
	return nil
}

func validateGCP(g *GCPProviderConfig) error {
	if g.ProjectID == "" {
		return NewConfigError(SectionProviders, "gcp.project_id", "GCP project ID cannot be empty", nil)
	}
	hasKeyPath := g.ServiceAccountKeyPath != ""
	hasKeyJSON := g.ServiceAccountKeyJSON != ""
	if !hasKeyPath && !hasKeyJSON {
		return NewConfigError(SectionProviders, "gcp.credentials", "GCP requires either service account key path or key JSON", nil)
	}
	if hasKeyPath {
		if _, err := os.Stat(g.ServiceAccountKeyPath); os.IsNotExist(err) {
			return NewConfigError(SectionProviders, "gcp.service_account_key_path", fmt.Sprintf("service account key file not found: %s", g.ServiceAccountKeyPath), err)
		}
	}
	if g.RequestTimeout <= 0 {
		return NewConfigError(SectionProviders, "gcp.request_timeout", "request timeout must be greater than 0", nil)
	}
	return nil
}

func ValidateDatabase(d *UnifiedDatabaseConfig) error {
	if !map[string]bool{"sqlite": true, "postgres": true, "mysql": true}[d.Type] {
		return NewConfigError(SectionDatabase, "type", fmt.Sprintf("invalid database type: %s", d.Type), nil)
	}
	if d.ConnectionString == "" {
		return NewConfigError(SectionDatabase, "connection_string", "database connection string cannot be empty", nil)
	}
	if d.MaxConnections <= 0 {
		return NewConfigError(SectionDatabase, "max_connections", "max connections must be greater than 0", nil)
	}
	if d.MaxIdleConnections < 0 {
		return NewConfigError(SectionDatabase, "max_idle_connections", "max idle connections cannot be negative", nil)
	}
	if d.ConnectionTimeout <= 0 {
		return NewConfigError(SectionDatabase, "connection_timeout", "connection timeout must be greater than 0", nil)
	}
	if d.QueryTimeout <= 0 {
		return NewConfigError(SectionDatabase, "query_timeout", "query timeout must be greater than 0", nil)
	}
	return nil
}

func ValidateStreaming(s *UnifiedStreamingConfig) error {
	if s.MaxConcurrentJobs <= 0 {
		return NewConfigError(SectionStreaming, "max_concurrent_jobs", "max concurrent jobs must be greater than 0", nil)
	}
	if s.DefaultWorkers <= 0 {
		return NewConfigError(SectionStreaming, "default_workers", "default workers must be greater than 0", nil)
	}
	if s.DefaultMemory <= 0 {
		return NewConfigError(SectionStreaming, "default_memory", "default memory must be greater than 0", nil)
	}
	if s.JobTimeout <= 0 {
		return NewConfigError(SectionStreaming, "job_timeout", "job timeout must be greater than 0", nil)
	}
	if s.ProgressInterval <= 0 {
		return NewConfigError(SectionStreaming, "progress_interval", "progress interval must be greater than 0", nil)
	}
	return nil
}

func ValidateSecurity(sec *UnifiedSecurityConfig) error {
	if sec.EncryptionEnabled && sec.EncryptionKey == "" {
		return NewConfigError(SectionSecurity, "encryption_key", "encryption key cannot be empty when encryption is enabled", nil)
	}
	if sec.TLSEnabled {
		if sec.TLSCertPath == "" {
			return NewConfigError(SectionSecurity, "tls_cert_path", "TLS certificate path cannot be empty when TLS is enabled", nil)
		}
		if sec.TLSKeyPath == "" {
			return NewConfigError(SectionSecurity, "tls_key_path", "TLS key path cannot be empty when TLS is enabled", nil)
		}
		if _, err := os.Stat(sec.TLSCertPath); os.IsNotExist(err) {
			return NewConfigError(SectionSecurity, "tls_cert_path", fmt.Sprintf("TLS certificate file not found: %s", sec.TLSCertPath), err)
		}
		if _, err := os.Stat(sec.TLSKeyPath); os.IsNotExist(err) {
			return NewConfigError(SectionSecurity, "tls_key_path", fmt.Sprintf("TLS key file not found: %s", sec.TLSKeyPath), err)
		}
	}
	if sec.TokenExpiry <= 0 {
		return NewConfigError(SectionSecurity, "token_expiry", "token expiry must be greater than 0", nil)
	}
	if sec.JWTSecret == "" {
		return NewConfigError(SectionSecurity, "jwt_secret", "jwt secret cannot be empty", nil)
	}
	return nil
}

func ValidateFocus(*UnifiedFocusConfig) error { return nil }
