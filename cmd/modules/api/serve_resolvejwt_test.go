package api

import (
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/costscope/costscope/internal/core/logging"
)

func TestResolveJWTSecret_FlagPrecedence(t *testing.T) {
	// isolate config loading by overriding HOME
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

	// set package var and set flag on a fresh command so the explicit branch is picked
	prev := jwtSecret
	jwtSecret = strings.Repeat("f", 40)
	defer func() { jwtSecret = prev }()

	cmd := &cobra.Command{}
	cmd.Flags().String("jwt-secret", "", "")
	if err := cmd.Flags().Set("jwt-secret", jwtSecret); err != nil {
		t.Fatalf("set flag failed: %v", err)
	}
	logger := logging.NewLogger(logging.LevelInfo)
	v, err := resolveJWTSecret(cmd, logger)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if v != jwtSecret {
		t.Fatalf("expected %q got %q", jwtSecret, v)
	}
}

func TestResolveJWTSecret_EnvPrecedence(t *testing.T) {
	// isolate config loading by overriding HOME
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

	// clear flag and package var
	prev := jwtSecret
	jwtSecret = ""
	defer func() { jwtSecret = prev }()

	// ensure no flag is set on a fresh command
	cmd := &cobra.Command{}
	cmd.Flags().String("jwt-secret", "", "")

	// set env
	old := os.Getenv("COSTSCOPE_JWT_SECRET")
	if err := os.Setenv("COSTSCOPE_JWT_SECRET", "env-secret-123456789012345678901234567890"); err != nil {
		t.Fatalf("failed to set COSTSCOPE_JWT_SECRET: %v", err)
	}
	defer func() {
		if err := os.Setenv("COSTSCOPE_JWT_SECRET", old); err != nil {
			t.Fatalf("failed to restore COSTSCOPE_JWT_SECRET: %v", err)
		}
	}()

	logger := logging.NewLogger(logging.LevelInfo)
	v, err := resolveJWTSecret(cmd, logger)
	if err != nil {
		t.Fatalf("expected no error from env, got %v", err)
	}
	if v == "" {
		t.Fatalf("expected non-empty secret from env")
	}
}

func TestResolveJWTSecret_ErrorWhenMissing(t *testing.T) {
	// isolate config loading by overriding HOME
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

	// ensure env cleared and no flag/value
	old := os.Getenv("COSTSCOPE_JWT_SECRET")
	if err := os.Unsetenv("COSTSCOPE_JWT_SECRET"); err != nil {
		t.Fatalf("failed to unset COSTSCOPE_JWT_SECRET: %v", err)
	}
	defer func() {
		if err := os.Setenv("COSTSCOPE_JWT_SECRET", old); err != nil {
			t.Fatalf("failed to restore COSTSCOPE_JWT_SECRET: %v", err)
		}
	}()

	prev := jwtSecret
	jwtSecret = ""
	defer func() { jwtSecret = prev }()

	cmd := &cobra.Command{}
	cmd.Flags().String("jwt-secret", "", "")
	logger := logging.NewLogger(logging.LevelInfo)
	_, err := resolveJWTSecret(cmd, logger)
	if err == nil {
		t.Fatalf("expected error when jwt secret missing")
	}
}
