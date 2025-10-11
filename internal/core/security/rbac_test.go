package security

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileRBACStore_LoadSaveRoundTrip(t *testing.T) {
	dir := t.TempDir()

	store := NewFileRBACStore(dir)

	// Add two roles and save
	rAdmin := Role{
		Name:        "admin",
		Description: "Administrator",
		Permissions: []Permission{{Resource: "*", Action: "*"}},
	}
	if err := store.AddRole(rAdmin); err != nil {
		t.Fatalf("AddRole(admin) err: %v", err)
	}
	rViewer := Role{
		Name:        "viewer",
		Description: "Read-only",
		Permissions: []Permission{{Resource: "reports", Action: "read"}},
	}
	if err := store.AddRole(rViewer); err != nil {
		t.Fatalf("AddRole(viewer) err: %v", err)
	}
	if err := store.Save(); err != nil {
		t.Fatalf("Save err: %v", err)
	}

	// Ensure file created
	if _, err := os.Stat(filepath.Join(dir, "roles.json")); err != nil {
		t.Fatalf("roles.json missing: %v", err)
	}

	// Load into a fresh store and verify content
	store2 := NewFileRBACStore(dir)
	if err := store2.Load(); err != nil {
		t.Fatalf("Load err: %v", err)
	}

	if got := len(store2.ListRoles()); got != 2 {
		t.Fatalf("want 2 roles, got %d", got)
	}

	if _, ok := store2.GetRole("admin"); !ok {
		t.Fatalf("expected to find role admin")
	}
	if _, ok := store2.GetRole("viewer"); !ok {
		t.Fatalf("expected to find role viewer")
	}
}

func TestRBACService_CreateRoleAndHasPermission(t *testing.T) {
	dir := t.TempDir()
	store := NewFileRBACStore(dir)
	svc := NewRBACService(store, nil) // logger optional

	// Create role
	perms := []Permission{{Resource: "focus", Action: "convert"}, {Resource: "reports", Action: "read"}}
	role, err := svc.CreateRole("operator", "Can convert and read reports", perms)
	if err != nil {
		t.Fatalf("CreateRole err: %v", err)
	}
	if role.Name != "operator" {
		t.Fatalf("unexpected role name: %s", role.Name)
	}

	// Permissions check
	if !svc.HasPermission("operator", "focus", "convert") {
		t.Fatalf("operator should have focus:convert")
	}
	if svc.HasPermission("operator", "focus", "delete") {
		t.Fatalf("operator should NOT have focus:delete")
	}
}

func TestRBACService_HasPermission_NegativeAndMissingRole(t *testing.T) {
	dir := t.TempDir()
	store := NewFileRBACStore(dir)
	svc := NewRBACService(store, nil)

	// Create a role with limited perms
	perms := []Permission{{Resource: "focus", Action: "convert"}}
	if _, err := svc.CreateRole("operator", "", perms); err != nil {
		t.Fatalf("CreateRole err: %v", err)
	}

	// Negative: wrong action
	if svc.HasPermission("operator", "focus", "delete") {
		t.Fatalf("operator should NOT have focus:delete")
	}

	// Negative: wrong resource
	if svc.HasPermission("operator", "reports", "read") {
		t.Fatalf("operator should NOT have reports:read")
	}

	// Negative: missing role
	if svc.HasPermission("unknown", "focus", "convert") {
		t.Fatalf("unknown role should NOT have any permissions")
	}
}

func TestRBACStore_AddRoleDuplicate(t *testing.T) {
	dir := t.TempDir()
	store := NewFileRBACStore(dir)

	r := Role{Name: "dupe", Permissions: []Permission{{Resource: "x", Action: "y"}}}
	if err := store.AddRole(r); err != nil {
		t.Fatalf("first AddRole err: %v", err)
	}
	if err := store.AddRole(r); err == nil {
		t.Fatalf("expected duplicate role error, got nil")
	}
}
