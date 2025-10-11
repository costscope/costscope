package api

import (
	"os"
	"testing"

	"github.com/spf13/cobra"

	"local/costscope/internal/core/logging"
)

func TestResolveJWTSecret_FlagWins(t *testing.T) {
	// isolate HOME to avoid config YAML interference
	oldHome := os.Getenv("HOME")
	tmp := t.TempDir()
	if err := os.Setenv("HOME", tmp); err != nil {
		t.Fatalf("failed to set HOME: %v", err)
	}
	defer func() { _ = os.Setenv("HOME", oldHome) }()

	prev := jwtSecret
	jwtSecret = "flag-secret-abcdefghijklmnopqrstuvwxyz123456"
	defer func() { jwtSecret = prev }()

	cmd := &cobra.Command{}
	cmd.Flags().String("jwt-secret", "", "")
	if err := cmd.Flags().Set("jwt-secret", jwtSecret); err != nil {
		t.Fatalf("failed to set flag: %v", err)
	}
	logger := logging.NewLogger(logging.LevelInfo)
	v, err := resolveJWTSecret(cmd, logger)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if v != jwtSecret {
		t.Fatalf("expected flag value, got %q", v)
	}
}

func TestResolveJWTSecret_EnvWins(t *testing.T) {
	oldHome := os.Getenv("HOME")
	tmp := t.TempDir()
	if err := os.Setenv("HOME", tmp); err != nil {
		t.Fatalf("failed to set HOME: %v", err)
	}
	defer func() { _ = os.Setenv("HOME", oldHome) }()

	// ensure no flag or package var
	prev := jwtSecret
	jwtSecret = ""
	defer func() { jwtSecret = prev }()

	cmd := &cobra.Command{}
	cmd.Flags().String("jwt-secret", "", "")

	old := os.Getenv("COSTSCOPE_JWT_SECRET")
	if err := os.Setenv("COSTSCOPE_JWT_SECRET", "env-secret-abcdefghijklmnopqrstuvwxyz1234"); err != nil {
		t.Fatalf("failed to set env: %v", err)
	}
	defer func() { _ = os.Setenv("COSTSCOPE_JWT_SECRET", old) }()

	logger := logging.NewLogger(logging.LevelInfo)
	v, err := resolveJWTSecret(cmd, logger)
	if err != nil {
		t.Fatalf("expected no error from env, got %v", err)
	}
	if v == "" {
		t.Fatalf("expected non-empty secret from env")
	}
}
