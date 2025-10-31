//go:build casbinpoc
// +build casbinpoc

package security

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/costscope/costscope/internal/testutil"
)

// helper to create a temp policy file with provided content
func writeTempPolicy(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "policy-*.csv")
	if err != nil {
		t.Fatalf("failed to create temp policy: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("failed to write policy: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("failed to close policy: %v", err)
	}
	return f.Name()
}

func TestCasbinMiddleware_PathBased_Allow(t *testing.T) {
	// Use example path-based model
	root := testutil.FindRepoRoot(t)
	modelPath := filepath.Join(root, filepath.FromSlash("configs/rbac_model.conf.example"))
	if _, err := os.Stat(modelPath); err != nil {
		t.Skipf("model file not found: %v", err)
	}
	// Note: the example model uses g(r.sub, p.sub), so we need a role assignment
	// mapping for the subject. Since we use a role-like subject value ("role:admin"),
	// assign it to itself to satisfy g().
	policy := "p, role:admin, /api/v1/*, (GET|POST)\n" +
		"g, role:admin, role:admin\n"
	policyPath := writeTempPolicy(t, policy)

	e, err := NewEnforcerFromFiles(modelPath, policyPath)
	if err != nil {
		t.Fatalf("enforcer load err: %v", err)
	}

	// Next handler returns 200 OK
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, "ok")
	})

	// Use header subject extractor and set role:admin
	mw := CasbinHTTPMiddleware(next, e, HeaderSubjectExtractor())

	req := httptest.NewRequest(http.MethodPost, "/api/v1/focus/convert", nil)
	req.Header.Set("X-Subject", "role:admin")
	rr := httptest.NewRecorder()

	mw.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestCasbinMiddleware_PathBased_Deny(t *testing.T) {
	root := testutil.FindRepoRoot(t)
	modelPath := filepath.Join(root, filepath.FromSlash("configs/rbac_model.conf.example"))
	if _, err := os.Stat(modelPath); err != nil {
		t.Skipf("model file not found: %v", err)
	}
	// Only admin is granted; viewer is not assigned to a policy allowing access.
	policy := "p, role:admin, /api/v1/*, (GET|POST)\n" +
		"g, role:admin, role:admin\n"
	policyPath := writeTempPolicy(t, policy)

	e, err := NewEnforcerFromFiles(modelPath, policyPath)
	if err != nil {
		t.Fatalf("enforcer load err: %v", err)
	}

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, "ok")
	})

	mw := CasbinHTTPMiddleware(next, e, HeaderSubjectExtractor())

	req := httptest.NewRequest(http.MethodPost, "/api/v1/focus/convert", nil)
	req.Header.Set("X-Subject", "role:viewer")
	rr := httptest.NewRecorder()

	mw.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rr.Code)
	}
}
