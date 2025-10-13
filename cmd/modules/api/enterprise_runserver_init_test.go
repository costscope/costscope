package api

import (
	"os"
	"strings"
	"testing"

	"github.com/costscope/costscope/internal/core/logging"
)

// TestRunEnterpriseAPIServer_InitBranches exercises initialization branches of runEnterpriseAPIServer
// without starting a real network server by using an invalid port which causes ListenAndServe to
// return quickly with an error. This is deterministic and avoids opening sockets.
func TestRunEnterpriseAPIServer_InitBranches(t *testing.T) {
	// Isolate HOME so config loading doesn't pick user files
	oldHome := os.Getenv("HOME")
	tmp := t.TempDir()
	if err := os.Setenv("HOME", tmp); err != nil {
		t.Fatalf("failed to set HOME: %v", err)
	}
	defer func() {
		if err := os.Setenv("HOME", oldHome); err != nil {
			t.Fatalf("failed to restore HOME: %v", err)
		}
	}()

	// Preserve and restore globals we mutate
	oldHost := enterpriseHost
	oldPort := enterprisePort
	oldJwt := enterpriseJwtSecret
	oldTls := enterpriseTlsEnabled
	oldCasbin := enterpriseCasbinEnabled
	oldJobWorkers := enterpriseJobWorkers

	enterpriseHost = "127.0.0.1"
	// Invalid negative port forces Listen error quickly
	enterprisePort = -1
	enterpriseJobWorkers = 1
	enterpriseJwtSecret = strings.Repeat("x", 40)
	enterpriseTlsEnabled = false
	enterpriseCasbinEnabled = false

	defer func() {
		enterpriseHost = oldHost
		enterprisePort = oldPort
		enterpriseJwtSecret = oldJwt
		enterpriseTlsEnabled = oldTls
		enterpriseCasbinEnabled = oldCasbin
		enterpriseJobWorkers = oldJobWorkers
	}()

	logger := logging.NewLogger(logging.LevelInfo)

	// Call and expect an error due to server listen failure; this ensures init code ran.
	if err := runEnterpriseAPIServer(nil, []string{}); err == nil {
		t.Fatalf("expected error from runEnterpriseAPIServer when Listen fails; got nil")
	} else {
		// Basic check: error should mention server
		if !strings.Contains(err.Error(), "server") {
			t.Fatalf("unexpected error from runEnterpriseAPIServer: %v", err)
		}
	}
	_ = logger // silence unused in case of build tags
}
