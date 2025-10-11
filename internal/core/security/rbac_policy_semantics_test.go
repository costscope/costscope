package security

import (
	"testing"
	"time"
)

// TestRBACPolicySemantics validates baseline semantics requested:
// - allow admin wildcard (action=* resource=*)
// - allow read-only: list/get health & analytics read (simulated with analytics:trends + focus:validate as read ops)
// - deny unknown role
// - deny escalation: read-only cannot perform mutation (POST/DELETE style -> simulate focus:convert and reports:generate)
func TestRBACPolicySemantics(t *testing.T) {
	// Setup in-memory store
	dir := t.TempDir()
	store := NewFileRBACStore(dir)
	if err := store.Load(); err != nil {
		t.Fatalf("load store: %v", err)
	}
	svc := NewRBACService(store, nil)

	// Admin role with wildcard perms
	admin := Role{Name: "admin", Permissions: []Permission{{Resource: "*", Action: "*"}}, CreatedAt: time.Now().UTC()}
	if err := store.AddRole(admin); err != nil {
		t.Fatalf("add admin: %v", err)
	}

	// Reader role – allow limited read operations only
	readerPerms := []Permission{
		{Resource: ResourceAnalytics, Action: ActionTrends}, // analytics read example
		{Resource: ResourceFocus, Action: ActionValidate},   // focus validate (read-like)
	}
	reader := Role{Name: "reader", Permissions: readerPerms, CreatedAt: time.Now().UTC()}
	if err := store.AddRole(reader); err != nil {
		t.Fatalf("add reader: %v", err)
	}

	type check struct {
		role     string
		resource string
		action   string
		allowed  bool
		desc     string
	}
	tests := []check{
		{"admin", ResourceFocus, ActionConvert, true, "admin wildcard convert"},
		{"admin", ResourceAnalytics, ActionForecast, true, "admin wildcard analytics forecast"},
		{"reader", ResourceAnalytics, ActionTrends, true, "reader allowed analytics trends"},
		{"reader", ResourceFocus, ActionValidate, true, "reader allowed focus validate"},
		{"reader", ResourceFocus, ActionConvert, false, "reader cannot convert (mutation)"},
		{"reader", ResourceReports, ActionGenerate, false, "reader cannot generate reports (mutation)"},
		{"unknown", ResourceFocus, ActionValidate, false, "unknown role denied"},
	}

	for _, tc := range tests {
		got := svc.HasPermission(tc.role, tc.resource, tc.action)
		if got != tc.allowed {
			t.Fatalf("%s: expected allowed=%v got %v (role=%s resource=%s action=%s)", tc.desc, tc.allowed, got, tc.role, tc.resource, tc.action)
		}
	}
}
