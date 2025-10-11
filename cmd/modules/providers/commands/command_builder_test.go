package commands

import "testing"

// Clean, minimal tests for provider command builders. Duplicate / malformed
// legacy test code was removed for stability; tests now assert presence and
// required flags without being brittle about total counts.

func TestNewProviderCommands(t *testing.T) {
	pc := NewProviderCommands()
	if pc == nil {
		t.Fatal("NewProviderCommands should not return nil")
	}
	if pc.manager == nil {
		t.Error("manager should be initialized")
	}
	if pc.logger == nil {
		t.Error("logger should be initialized")
	}
}

func TestBuildProvidersCommand(t *testing.T) {
	pc := NewProviderCommands()
	cmd := pc.BuildProvidersCommand()
	if cmd == nil {
		t.Fatal("BuildProvidersCommand should not return nil")
	}
	if cmd.Use != "providers" {
		t.Errorf("expected 'providers', got %s", cmd.Use)
	}
	expected := []string{"list", "status", "validate", "info", "verify <file>"}
	present := map[string]bool{}
	for _, sc := range cmd.Commands() {
		present[sc.Use] = true
	}
	missing := []string{}
	for _, e := range expected {
		if !present[e] {
			missing = append(missing, e)
		}
	}
	if len(missing) > 0 {
		found := []string{}
		for k := range present {
			found = append(found, k)
		}
		t.Fatalf("missing expected subcommands: %v (found=%v)", missing, found)
	}
}

func TestBuildListCommand(t *testing.T) {
	pc := NewProviderCommands()
	cmd := pc.buildListCommand()
	if cmd == nil {
		t.Fatal("buildListCommand should not return nil")
	}
	if cmd.Use != "list" {
		t.Errorf("expected 'list', got %s", cmd.Use)
	}
	if cmd.Run == nil {
		t.Error("list command must have Run closure")
	}
}

func TestBuildStatusCommand(t *testing.T) {
	pc := NewProviderCommands()
	cmd := pc.buildStatusCommand()
	if cmd == nil {
		t.Fatal("buildStatusCommand should not return nil")
	}
	if cmd.Use != "status" {
		t.Errorf("expected 'status', got %s", cmd.Use)
	}
	if cmd.Flag("name") == nil {
		t.Error("missing --name flag")
	}
	if cmd.Flag("verbose") == nil {
		t.Error("missing --verbose flag")
	}
}

func TestBuildValidateCommand(t *testing.T) {
	pc := NewProviderCommands()
	cmd := pc.buildValidateCommand()
	if cmd == nil {
		t.Fatal("buildValidateCommand should not return nil")
	}
	if cmd.Use != "validate" {
		t.Errorf("expected 'validate', got %s", cmd.Use)
	}
	if cmd.Flag("name") == nil {
		t.Error("missing --name flag")
	}
	if cmd.Flag("all") == nil {
		t.Error("missing --all flag")
	}
}

func TestBuildInfoCommand(t *testing.T) {
	pc := NewProviderCommands()
	cmd := pc.buildInfoCommand()
	if cmd == nil {
		t.Fatal("buildInfoCommand should not return nil")
	}
	if cmd.Use != "info" {
		t.Errorf("expected 'info', got %s", cmd.Use)
	}
	if cmd.Flag("name") == nil {
		t.Error("missing --name flag")
	}
	if cmd.Flag("format") == nil {
		t.Error("missing --format flag")
	}
}

func TestBuildVerifyCommand(t *testing.T) {
	pc := NewProviderCommands()
	cmd := pc.buildVerifyCommand()
	if cmd == nil {
		t.Fatal("buildVerifyCommand should not return nil")
	}
	if cmd.Use != "verify <file>" {
		t.Errorf("expected 'verify <file>', got %s", cmd.Use)
	}
	flags := []string{"provider", "limit", "use-unified-mapper", "invariants", "invariants-baseline", "invariants-tolerance", "invariants-report", "fail-on-invariants", "format", "stop-after", "error-threshold"}
	for _, f := range flags {
		if cmd.Flag(f) == nil {
			t.Errorf("missing --%s flag", f)
		}
	}
}
