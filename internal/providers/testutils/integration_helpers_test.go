package testutils

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Integration credential helpers are now test-only (in *_test.go) to keep production
// surface minimal while preserving optional integration smoke tests.

func RequireAWSIntegrationCredentials(t *testing.T) map[string]string {
	t.Helper()
	access := firstNonEmpty(os.Getenv("COSTSCOPE_AWS_ACCESS_KEY_ID"), os.Getenv("AWS_ACCESS_KEY_ID"))
	secret := firstNonEmpty(os.Getenv("COSTSCOPE_AWS_SECRET_ACCESS_KEY"), os.Getenv("AWS_SECRET_ACCESS_KEY"))
	region := firstNonEmpty(os.Getenv("COSTSCOPE_AWS_REGION"), os.Getenv("AWS_REGION"))
	if region == "" {
		region = "us-east-1"
	}
	if access == "" || secret == "" {
		t.Skip("skipping AWS integration test (missing env credentials)")
	}
	return map[string]string{"access_key": access, "secret_key": secret, "region": region}
}

func RequireGCPIntegrationCredentials(t *testing.T) map[string]string {
	t.Helper()
	project := os.Getenv("COSTSCOPE_GCP_PROJECT_ID")
	keyInline := os.Getenv("COSTSCOPE_GCP_SERVICE_ACCOUNT_KEY")
	if keyInline == "" {
		if p := os.Getenv("COSTSCOPE_GCP_SERVICE_ACCOUNT_KEY_FILE"); p != "" {
			if !filepath.IsAbs(p) || strings.Contains(p, "..") {
				t.Skipf("skipping GCP integration (unsafe path %s)", p)
			} else {
				// #nosec G304 -- test-only path read: validated as absolute, no traversal ('..') and controlled via env for integration testing
				data, err := os.ReadFile(p)
				if err == nil {
					keyInline = string(data)
				} else {
					t.Skipf("skipping GCP integration (read error %v)", err)
				}
			}
		}
	}
	if project == "" || keyInline == "" || !isLikelyJSON(keyInline) {
		t.Skip("skipping GCP integration test (missing project/key)")
	}
	return map[string]string{"project_id": project, "service_account_key": keyInline}
}

func RequireAzureIntegrationCredentials(t *testing.T) map[string]string {
	t.Helper()
	sub := os.Getenv("COSTSCOPE_AZURE_SUBSCRIPTION_ID")
	tenant := os.Getenv("COSTSCOPE_AZURE_TENANT_ID")
	client := os.Getenv("COSTSCOPE_AZURE_CLIENT_ID")
	secret := os.Getenv("COSTSCOPE_AZURE_CLIENT_SECRET")
	if sub == "" || tenant == "" || client == "" || secret == "" {
		t.Skip("skipping Azure integration test (missing env)")
	}
	return map[string]string{"subscription_id": sub, "tenant_id": tenant, "client_id": client, "client_secret": secret}
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

func isLikelyJSON(s string) bool {
	if len(s) < 2 || s[0] != '{' || s[len(s)-1] != '}' {
		return false
	}
	var tmp map[string]interface{}
	return json.Unmarshal([]byte(s), &tmp) == nil
}
