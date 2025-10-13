package security

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/costscope/costscope/internal/core/monitoring/telemetry"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// TestCheckPermissionMetrics exercises the tracing/metrics path of CheckPermission to improve coverage.
func TestCheckPermissionMetrics(t *testing.T) {
	// Register metrics (ignore panic if already registered in larger test run)
	func() { defer func() { _ = recover() }(); telemetry.Register() }()

	dir := t.TempDir()
	store := NewFileRBACStore(dir)
	svc := NewRBACService(store, nil)
	// Add a role with a single permission
	if err := store.AddRole(Role{Name: "tester", Permissions: []Permission{{Resource: ResourceFocus, Action: ActionConvert}}}); err != nil {
		t.Fatalf("add role: %v", err)
	}

	// Allowed check
	if !svc.CheckPermission(context.Background(), "tester", ResourceFocus, ActionConvert) {
		t.Fatalf("expected tester to be allowed focus:convert")
	}
	// Denied check (action mismatch) to exercise outcome label
	if svc.CheckPermission(context.Background(), "tester", ResourceFocus, ActionValidate) {
		t.Fatalf("expected tester to be denied focus:validate")
	}

	// Scrape metrics and ensure both allowed and denied counters present
	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/metrics", nil)
	promhttp.Handler().ServeHTTP(rr, req)
	body := rr.Body.String()
	if !containsAll(body, []string{"costscope_rbac_checks_total{action=\"convert\",allowed=\"allowed\",resource=\"focus\"}", "costscope_rbac_checks_total{action=\"validate\",allowed=\"denied\",resource=\"focus\"}"}) {
		t.Fatalf("expected allowed+denied RBAC metrics, got: %s", body)
	}
}

// containsAll simple helper (kept local to avoid pulling extra util dependencies)
func containsAll(s string, subs []string) bool {
	for _, sub := range subs {
		if !contains(s, sub) {
			return false
		}
	}
	return true
}

func contains(s, sub string) bool {
	return len(sub) == 0 || (len(s) >= len(sub) && indexOf(s, sub) >= 0)
}

// indexOf naive substring search (avoid strings import to keep file minimal & focused)
func indexOf(haystack, needle string) int {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return i
		}
	}
	return -1
}
