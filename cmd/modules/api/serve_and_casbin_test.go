package api

import (
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/costscope/costscope/internal/core/logging"
)

func TestRunAPIServer_TestModeSkipsStart(t *testing.T) {
	// ensure env restored
	old := os.Getenv("COSTSCOPE_TEST_MODE")
	if err := os.Setenv("COSTSCOPE_TEST_MODE", "1"); err != nil {
		t.Fatalf("failed to set COSTSCOPE_TEST_MODE: %v", err)
	}
	defer func() {
		if err := os.Setenv("COSTSCOPE_TEST_MODE", old); err != nil {
			t.Fatalf("failed to restore COSTSCOPE_TEST_MODE: %v", err)
		}
	}()

	// Build command and set a valid jwt-secret flag
	cmd := BuildAPICommand()
	if err := cmd.Flags().Set("jwt-secret", strings.Repeat("x", 32)); err != nil {
		t.Fatalf("failed to set flag: %v", err)
	}

	// Should return nil (skips server startup in test mode)
	if err := runAPIServer(cmd, []string{}); err != nil {
		t.Fatalf("expected runAPIServer to return nil in test mode, got: %v", err)
	}
}

func TestWrapServerWithCasbinIfEnabled_WithTempFiles_NoPanic(t *testing.T) {
	// create temporary model and policy files
	model := `[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = r.sub == p.sub && r.obj == p.obj && r.act == p.act
`
	policy := "p, alice, data1, read\n"

	mf, err := os.CreateTemp("", "casbin-model-*.conf")
	if err != nil {
		t.Fatalf("failed to create temp model: %v", err)
	}
	defer func() {
		if err := os.Remove(mf.Name()); err != nil {
			t.Fatalf("failed to remove temp model: %v", err)
		}
	}()

	if _, err := mf.WriteString(model); err != nil {
		t.Fatalf("failed to write model: %v", err)
	}
	if err := mf.Close(); err != nil {
		t.Fatalf("failed to close temp model file: %v", err)
	}

	pf, err := os.CreateTemp("", "casbin-policy-*.csv")
	if err != nil {
		t.Fatalf("failed to create temp policy: %v", err)
	}
	defer func() {
		if err := os.Remove(pf.Name()); err != nil {
			t.Fatalf("failed to remove temp policy: %v", err)
		}
	}()

	if _, err := pf.WriteString(policy); err != nil {
		t.Fatalf("failed to write policy: %v", err)
	}
	if err := pf.Close(); err != nil {
		t.Fatalf("failed to close temp policy file: %v", err)
	}

	// set globals and restore after
	prevEnabled := enterpriseCasbinEnabled
	prevModel := enterpriseCasbinModelPath
	prevPolicy := enterpriseCasbinPolicyPath
	enterpriseCasbinEnabled = true
	enterpriseCasbinModelPath = mf.Name()
	enterpriseCasbinPolicyPath = pf.Name()
	defer func() {
		enterpriseCasbinEnabled = prevEnabled
		enterpriseCasbinModelPath = prevModel
		enterpriseCasbinPolicyPath = prevPolicy
	}()

	// Build a dummy handler and pass its pointer
	var h http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) })
	logger := logging.NewLogger(logging.LevelInfo)
	// Should not panic
	wrapServerWithCasbinIfEnabled(&h, logger)
}
