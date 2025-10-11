package api

import (
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestRunAPIServer_TestMode_SkipsStartup(t *testing.T) {
	// isolate config loading
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

	oldTest := os.Getenv("COSTSCOPE_TEST_MODE")
	if err := os.Setenv("COSTSCOPE_TEST_MODE", "1"); err != nil {
		t.Fatalf("failed to set COSTSCOPE_TEST_MODE: %v", err)
	}
	defer func() {
		if err := os.Setenv("COSTSCOPE_TEST_MODE", oldTest); err != nil {
			t.Fatalf("failed to restore COSTSCOPE_TEST_MODE: %v", err)
		}
	}()

	// Build command and set explicit jwt-secret flag so resolveJWTSecret uses explicit value
	cmd := BuildAPICommand()
	secret := strings.Repeat("x", 40)
	if err := cmd.Flags().Set("jwt-secret", secret); err != nil {
		t.Fatalf("failed to set flag: %v", err)
	}

	if err := runAPIServer(cmd, []string{}); err != nil {
		t.Fatalf("expected runAPIServer to return nil in test mode, got %v", err)
	}
}

func TestRunAPIServer_MissingJWT_ReturnsError(t *testing.T) {
	// isolate config loading
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

	oldTest := os.Getenv("COSTSCOPE_TEST_MODE")
	if err := os.Setenv("COSTSCOPE_TEST_MODE", "1"); err != nil {
		t.Fatalf("failed to set COSTSCOPE_TEST_MODE: %v", err)
	}
	defer func() {
		if err := os.Setenv("COSTSCOPE_TEST_MODE", oldTest); err != nil {
			t.Fatalf("failed to restore COSTSCOPE_TEST_MODE: %v", err)
		}
	}()

	// Ensure no jwt flag or env is set
	cmd := &cobra.Command{}
	cmd.Flags().String("jwt-secret", "", "")

	if err := runAPIServer(cmd, []string{}); err == nil {
		t.Fatalf("expected error when jwt secret missing")
	}
}
