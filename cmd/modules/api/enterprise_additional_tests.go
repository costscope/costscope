package api

import "testing"

func TestRunAPIServer_ReturnsErrorWhenJWTMissing(t *testing.T) {
	// Ensure no COSTSCOPE_TEST_MODE is set so runAPIServer will attempt resolveJWTSecret and fail
	// by returning an error when secret is missing.
	cmd := BuildAPICommand()
	// Clear any jwtSecret package var to force missing-secret path
	prev := jwtSecret
	jwtSecret = ""
	defer func() { jwtSecret = prev }()

	if err := runAPIServer(cmd, []string{}); err == nil {
		t.Fatalf("expected error from runAPIServer when jwt secret missing, got nil")
	}
}
