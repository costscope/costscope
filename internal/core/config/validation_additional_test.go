package config

import (
	"testing"
	"time"
)

// Additional table-driven tests to cover previously untested validation branches.

func TestValidateAllConfig_Nil(t *testing.T) {
	if err := ValidateAllConfig(nil); err == nil {
		t.Fatalf("expected error for nil config")
	}
}

func TestEnsureConfigDirectories_Nil(t *testing.T) {
	if err := EnsureConfigDirectories(nil); err == nil {
		t.Fatalf("expected error for nil config")
	}
}

func TestValidateProviders_AWS_Table(t *testing.T) {
	cases := []struct {
		name    string
		cfg     AWSProviderConfig
		wantErr bool
	}{
		{"missingRegion", AWSProviderConfig{Enabled: true, Profile: "p"}, true},
		{"profileOnly", AWSProviderConfig{Enabled: true, Region: "us-east-1", Profile: "p", RequestTimeout: time.Second}, false},
		{"roleOnly", AWSProviderConfig{Enabled: true, Region: "us-east-1", RoleARN: "arn:aws:iam::123:role/Test", RequestTimeout: time.Second}, false},
		{"keysOnly", AWSProviderConfig{Enabled: true, Region: "us-east-1", AccessKey: "A", SecretKey: "S", RequestTimeout: time.Second}, false},
		{"timeoutZero", AWSProviderConfig{Enabled: true, Region: "us-east-1", Profile: "p", RequestTimeout: 0}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			up := UnifiedProvidersConfig{AWS: tc.cfg}
			if err := ValidateProviders(&up); (err != nil) != tc.wantErr {
				t.Fatalf("unexpected error state: %v (wantErr=%v)", err, tc.wantErr)
			}
		})
	}
}

func TestValidateProviders_Azure_Table(t *testing.T) {
	base := AzureProviderConfig{Enabled: true, TenantID: "tid", ClientID: "cid", ClientSecret: "sec", SubscriptionID: "sub", RequestTimeout: time.Second}
	// Each case mutates one field to trigger an error.
	cases := []struct {
		name    string
		mutate  func(*AzureProviderConfig)
		wantErr bool
	}{
		{"valid", func(a *AzureProviderConfig) {}, false},
		{"missingTenant", func(a *AzureProviderConfig) { a.TenantID = "" }, true},
		{"missingClientID", func(a *AzureProviderConfig) { a.ClientID = "" }, true},
		{"missingClientSecret", func(a *AzureProviderConfig) { a.ClientSecret = "" }, true},
		{"missingSubscription", func(a *AzureProviderConfig) { a.SubscriptionID = "" }, true},
		{"timeoutZero", func(a *AzureProviderConfig) { a.RequestTimeout = 0 }, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := base
			tc.mutate(&cfg)
			up := UnifiedProvidersConfig{Azure: cfg}
			if err := ValidateProviders(&up); (err != nil) != tc.wantErr {
				t.Fatalf("unexpected error state: %v (wantErr=%v)", err, tc.wantErr)
			}
		})
	}
}

func TestValidateProviders_GCP_Table(t *testing.T) {
	cases := []struct {
		name    string
		cfg     GCPProviderConfig
		wantErr bool
	}{
		{"missingProject", GCPProviderConfig{Enabled: true, ServiceAccountKeyJSON: "{}", RequestTimeout: time.Second}, true},
		{"missingCredentials", GCPProviderConfig{Enabled: true, ProjectID: "p", RequestTimeout: time.Second}, true},
		{"keyJSON", GCPProviderConfig{Enabled: true, ProjectID: "p", ServiceAccountKeyJSON: "{}", RequestTimeout: time.Second}, false},
		{"timeoutZero", GCPProviderConfig{Enabled: true, ProjectID: "p", ServiceAccountKeyJSON: "{}", RequestTimeout: 0}, true},
		{"missingKeyPathFile", GCPProviderConfig{Enabled: true, ProjectID: "p", ServiceAccountKeyPath: "/nonexistent/file.json", RequestTimeout: time.Second}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			up := UnifiedProvidersConfig{GCP: tc.cfg}
			if err := ValidateProviders(&up); (err != nil) != tc.wantErr {
				t.Fatalf("unexpected error state: %v (wantErr=%v)", err, tc.wantErr)
			}
		})
	}
}

func TestValidateDatabase_Timeouts(t *testing.T) {
	d := UnifiedDatabaseConfig{Type: "sqlite", ConnectionString: "file::memory:?cache=shared", MaxConnections: 1, MaxIdleConnections: 0, ConnectionTimeout: time.Second, QueryTimeout: time.Second}
	if err := ValidateDatabase(&d); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	d.ConnectionTimeout = 0
	if err := ValidateDatabase(&d); err == nil {
		t.Fatalf("expected connection timeout error")
	}
	d.ConnectionTimeout = time.Second
	d.QueryTimeout = 0
	if err := ValidateDatabase(&d); err == nil {
		t.Fatalf("expected query timeout error")
	}
}

func TestValidateSecurity_TLSFileErrors(t *testing.T) {
	sec := UnifiedSecurityConfig{EncryptionEnabled: false, TLSEnabled: true, TLSCertPath: "cert.pem", TLSKeyPath: "", TokenExpiry: time.Second, JWTSecret: "s"}
	if err := ValidateSecurity(&sec); err == nil {
		t.Fatalf("expected key path error")
	}
	sec.TLSKeyPath = "key.pem"
	// Files do not exist, should now produce cert not found error first (order matters)
	if err := ValidateSecurity(&sec); err == nil {
		t.Fatalf("expected cert file not found error")
	}
}

func TestValidateFocus_NoOp(t *testing.T) {
	if err := ValidateFocus(&UnifiedFocusConfig{}); err != nil {
		t.Fatalf("expected nil error for no-op focus validator")
	}
}
