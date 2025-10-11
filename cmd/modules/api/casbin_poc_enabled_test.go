//go:build casbinpoc
// +build casbinpoc

package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/golang-jwt/jwt/v5"

	"local/costscope/internal/core/logging"
)

func TestWrapServerWithCasbinIfEnabled_SuccessPath(t *testing.T) {
	// Prepare temporary model and policy files
	dir := t.TempDir()
	modelPath := filepath.Join(dir, "model.conf")
	policyPath := filepath.Join(dir, "policy.csv")
	model := `[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = r.sub == p.sub && r.obj == p.obj && r.act == p.act
`
	if err := os.WriteFile(modelPath, []byte(model), 0o600); err != nil {
		t.Fatalf("write model: %v", err)
	}
	// Policy permits role:admin to GET /allowed
	policy := "p, role:admin, /allowed, GET\n"
	if err := os.WriteFile(policyPath, []byte(policy), 0o600); err != nil {
		t.Fatalf("write policy: %v", err)
	}

	// Toggle package vars
	prevEnabled := enterpriseCasbinEnabled
	prevModel := enterpriseCasbinModelPath
	prevPolicy := enterpriseCasbinPolicyPath
	prevSecret := enterpriseJwtSecret
	enterpriseCasbinEnabled = true
	enterpriseCasbinModelPath = modelPath
	enterpriseCasbinPolicyPath = policyPath
	enterpriseJwtSecret = "test-secret-which-is-long-enough-0123456789"
	defer func() {
		enterpriseCasbinEnabled = prevEnabled
		enterpriseCasbinModelPath = prevModel
		enterpriseCasbinPolicyPath = prevPolicy
		enterpriseJwtSecret = prevSecret
	}()

	// Base handler that writes OK
	var h http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = io.WriteString(w, "ok")
	})

	logger := logging.NewLogger(logging.LevelInfo)
	// Wrap with casbin
	wrapServerWithCasbinIfEnabled(&h, logger)

	// 1) No auth -> should be 401
	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/allowed", nil)
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for missing auth, got %d", rr.Code)
	}

	// 2) With JWT bearing role 'admin' -> allowed
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"roles": []string{"admin"},
	})
	tokStr, err := token.SignedString([]byte(enterpriseJwtSecret))
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}
	rr2 := httptest.NewRecorder()
	req2 := httptest.NewRequest("GET", "/allowed", nil)
	req2.Header.Set("Authorization", "Bearer "+tokStr)
	h.ServeHTTP(rr2, req2)
	if rr2.Code != 200 {
		t.Fatalf("expected 200 for authorized request, got %d; body=%q", rr2.Code, rr2.Body.String())
	}
}
