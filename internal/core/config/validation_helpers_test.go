package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func minimalValidConfig(tmp string) *ConsolidatedConfig {
	return &ConsolidatedConfig{
		Core: UnifiedCoreConfig{
			AppName:        "costscope",
			Version:        "1.0.0",
			LogLevel:       "info",
			MaxConcurrency: 2,
			Timeout:        time.Second,
		},
		Providers: UnifiedProvidersConfig{ // disabled by default
			AWS:   AWSProviderConfig{},
			Azure: AzureProviderConfig{},
			GCP:   GCPProviderConfig{},
		},
		Database: UnifiedDatabaseConfig{
			Type:               "sqlite",
			ConnectionString:   "file::memory:?cache=shared",
			MaxConnections:     1,
			MaxIdleConnections: 0,
			ConnectionTimeout:  time.Second,
			QueryTimeout:       time.Second,
		},
		Streaming: UnifiedStreamingConfig{
			MaxConcurrentJobs: 1,
			DefaultWorkers:    1,
			DefaultMemory:     128,
			JobTimeout:        time.Second,
			ProgressInterval:  time.Second,
		},
		Security: UnifiedSecurityConfig{
			EncryptionEnabled: false,
			TLSEnabled:        false,
			TokenExpiry:       time.Minute,
			JWTSecret:         "abc123secure",
		},
		Focus:       UnifiedFocusConfig{},
		Reports:     UnifiedReportsConfig{},
		MultiTenant: UnifiedMultiTenantConfig{},
		DataDir:     tmp,
		Environment: Development,
		Version:     "1.0.0",
	}
}

func TestValidateAllConfig_Success(t *testing.T) {
	tmp := t.TempDir()
	cfg := minimalValidConfig(tmp)
	if err := ValidateAllConfig(cfg); err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
}

func TestValidateAllConfig_GlobalFailures(t *testing.T) {
	cfg := minimalValidConfig(" ")
	cfg.DataDir = "" // invalid
	if err := ValidateAllConfig(cfg); err == nil {
		t.Fatalf("expected error for empty data dir")
	}
	cfg.DataDir = "ok" // still relative but empty check passes; environment invalid next
	cfg.Environment = ConfigEnvironment("weird")
	if err := ValidateAllConfig(cfg); err == nil {
		t.Fatalf("expected environment error")
	}
}

func TestValidateCore_Errors(t *testing.T) {
	c := UnifiedCoreConfig{AppName: "", Version: "1", LogLevel: "info", MaxConcurrency: 1, Timeout: time.Second}
	if err := ValidateCore(&c); err == nil {
		t.Fatalf("expected app_name error")
	}
	c.AppName = "x"
	c.Version = ""
	if err := ValidateCore(&c); err == nil {
		t.Fatalf("expected version error")
	}
	c.Version = "1"
	c.LogLevel = "verbose"
	if err := ValidateCore(&c); err == nil {
		t.Fatalf("expected log level error")
	}
	c.LogLevel = "info"
	c.MaxConcurrency = 0
	if err := ValidateCore(&c); err == nil {
		t.Fatalf("expected max_concurrency error")
	}
	c.MaxConcurrency = 1
	c.Timeout = 0
	if err := ValidateCore(&c); err == nil {
		t.Fatalf("expected timeout error")
	}
}

func TestValidateProviders_AWSPaths(t *testing.T) {
	p := UnifiedProvidersConfig{AWS: AWSProviderConfig{Enabled: true, Region: "", Profile: "p"}}
	if err := ValidateProviders(&p); err == nil {
		t.Fatalf("expected region error")
	}
	p.AWS.Region = "us-east-1"
	p.AWS.Profile = ""
	p.AWS.AccessKey = "A" // missing secret key
	if err := ValidateProviders(&p); err == nil {
		t.Fatalf("expected credentials error (missing secret key)")
	}
	p.AWS.SecretKey = "S"
	p.AWS.RequestTimeout = time.Second
	if err := ValidateProviders(&p); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	p.AWS.MaxRetries = -1
	if err := ValidateProviders(&p); err == nil {
		t.Fatalf("expected max retries error")
	}
}

func TestValidateProviders_AzureAndGCP(t *testing.T) {
	p := UnifiedProvidersConfig{Azure: AzureProviderConfig{Enabled: true, TenantID: ""}}
	if err := ValidateProviders(&p); err == nil {
		t.Fatalf("expected azure tenant error")
	}
	p.Azure.TenantID = "tid"
	p.Azure.ClientID = "cid"
	p.Azure.ClientSecret = "sec"
	p.Azure.SubscriptionID = "sub"
	p.Azure.RequestTimeout = time.Second
	if err := ValidateProviders(&p); err != nil {
		t.Fatalf("unexpected azure success error: %v", err)
	}

	p.GCP.Enabled = true
	p.GCP.ProjectID = ""
	if err := ValidateProviders(&p); err == nil {
		t.Fatalf("expected gcp project id error")
	}
	p.GCP.ProjectID = "proj"
	p.GCP.ServiceAccountKeyJSON = "{ }"
	p.GCP.RequestTimeout = time.Second
	if err := ValidateProviders(&p); err != nil {
		t.Fatalf("unexpected gcp success error: %v", err)
	}
}

func TestValidateDatabase(t *testing.T) {
	d := UnifiedDatabaseConfig{Type: "bogus"}
	if err := ValidateDatabase(&d); err == nil {
		t.Fatalf("expected invalid type error")
	}
	d.Type = "sqlite"
	if err := ValidateDatabase(&d); err == nil {
		t.Fatalf("expected missing connection string error")
	}
	d.ConnectionString = "file::memory:"
	d.MaxConnections = 0
	d.MaxIdleConnections = 0
	d.ConnectionTimeout = time.Second
	d.QueryTimeout = time.Second
	if err := ValidateDatabase(&d); err == nil {
		t.Fatalf("expected max connections error")
	}
	d.MaxConnections = 1
	if err := ValidateDatabase(&d); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	d.MaxIdleConnections = -1
	if err := ValidateDatabase(&d); err == nil {
		t.Fatalf("expected negative idle error")
	}
}

func TestValidateStreaming(t *testing.T) {
	s := UnifiedStreamingConfig{}
	if err := ValidateStreaming(&s); err == nil {
		t.Fatalf("expected max_concurrent_jobs error")
	}
	s.MaxConcurrentJobs = 1
	s.DefaultWorkers = 0
	if err := ValidateStreaming(&s); err == nil {
		t.Fatalf("expected default_workers error")
	}
	s.DefaultWorkers = 1
	s.DefaultMemory = 0
	if err := ValidateStreaming(&s); err == nil {
		t.Fatalf("expected default_memory error")
	}
	s.DefaultMemory = 1
	s.JobTimeout = 0
	if err := ValidateStreaming(&s); err == nil {
		t.Fatalf("expected job_timeout error")
	}
	s.JobTimeout = time.Second
	s.ProgressInterval = 0
	if err := ValidateStreaming(&s); err == nil {
		t.Fatalf("expected progress_interval error")
	}
	s.ProgressInterval = time.Second
	if err := ValidateStreaming(&s); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
}

func TestValidateSecurity(t *testing.T) {
	sec := UnifiedSecurityConfig{EncryptionEnabled: true, EncryptionKey: "", TLSEnabled: false, TokenExpiry: time.Second, JWTSecret: "s"}
	if err := ValidateSecurity(&sec); err == nil {
		t.Fatalf("expected encryption key error")
	}
	sec.EncryptionKey = "k"
	sec.TLSEnabled = true
	if err := ValidateSecurity(&sec); err == nil {
		t.Fatalf("expected missing TLS cert path error")
	}
	sec.TLSCertPath = filepath.Join(t.TempDir(), "cert.pem")
	sec.TLSKeyPath = filepath.Join(t.TempDir(), "key.pem")
	// create files
	if err := os.WriteFile(sec.TLSCertPath, []byte("cert"), 0o600); err != nil {
		t.Fatalf("write cert: %v", err)
	}
	if err := os.WriteFile(sec.TLSKeyPath, []byte("key"), 0o600); err != nil {
		t.Fatalf("write key: %v", err)
	}
	if err := ValidateSecurity(&sec); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	sec.TokenExpiry = 0
	if err := ValidateSecurity(&sec); err == nil {
		t.Fatalf("expected token expiry error")
	}
	sec.TokenExpiry = time.Second
	sec.JWTSecret = ""
	if err := ValidateSecurity(&sec); err == nil {
		t.Fatalf("expected jwt secret error")
	}
}

func TestEnsureConfigDirectories(t *testing.T) {
	base := t.TempDir()
	cfg := minimalValidConfig(base)
	cfg.Core.DataDirectory = filepath.Join(base, "data")
	cfg.Core.TempDirectory = filepath.Join(base, "tmp")
	cfg.Streaming.CheckpointEnabled = true
	cfg.Streaming.CheckpointDir = filepath.Join(base, "cp")
	cfg.Database.MigrationsPath = filepath.Join(base, "migrations")
	if err := EnsureConfigDirectories(cfg); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	// basic existence checks
	dirs := []string{cfg.DataDir, cfg.Core.DataDirectory, cfg.Core.TempDirectory, cfg.Streaming.CheckpointDir, cfg.Database.MigrationsPath}
	for _, d := range dirs {
		if fi, err := os.Stat(d); err != nil || !fi.IsDir() {
			t.Fatalf("expected dir %s", d)
		}
	}
}
