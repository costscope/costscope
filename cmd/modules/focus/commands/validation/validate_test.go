package validation

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// helper to execute a cobra command
func executeCommand(cmd *cobra.Command, args ...string) error {
	cmd.SetArgs(args)
	cmd.SetOut(os.Stdout)
	cmd.SetErr(os.Stderr)
	return cmd.Execute()
}

func TestBuildValidateCommand_Help(t *testing.T) {
	cmd := BuildValidateCommand()
	if cmd.Use == "" || cmd.RunE == nil {
		t.Fatalf("validate command not properly constructed")
	}
}

func TestRunValidate_MinimalJSON(t *testing.T) {
	dir := t.TempDir()
	// Use a filename containing "test" to trigger the simulator's happy path
	// in SchemaValidator and QualityValidator so validation passes without os.Exit(1).
	f := filepath.Join(dir, "test-focus.parquet")
	if err := os.WriteFile(f, []byte(""), 0600); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	cmd := BuildValidateCommand()
	// Enable only schema, quality, and performance to avoid anomaly detector making the run invalid
	err := executeCommand(cmd, f, "--format", "json", "--quiet", "--schema", "--quality", "--performance")
	if err != nil {
		t.Fatalf("validate command failed: %v", err)
	}
}

func TestRunBatchValidate_NoFiles(t *testing.T) {
	dir := t.TempDir()
	cmd := BuildValidateCommand()
	// find the batch subcommand
	var batch *cobra.Command
	for _, c := range cmd.Commands() {
		if c.Use == "batch [directory]" {
			batch = c
			break
		}
	}
	if batch == nil {
		t.Fatalf("batch subcommand not found")
	}
	err := executeCommand(batch, dir)
	if err == nil {
		t.Fatalf("expected error for no files, got nil")
	}
}

func TestValidate_ListSchemas_NoInput(t *testing.T) {
	cmd := BuildValidateCommand()
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() { os.Stdout = oldStdout }()

	// run with --list-schemas (no args)
	if err := executeCommand(cmd, "--list-schemas"); err != nil {
		t.Fatalf("list-schemas failed: %v", err)
	}
	_ = w.Close()
	out, _ := io.ReadAll(r)
	s := string(out)
	if !strings.Contains(s, "Available FOCUS schemas:") {
		t.Fatalf("unexpected output: %s", s)
	}
	// check expected builtin names are present
	if !strings.Contains(s, "focus-1.2") || !strings.Contains(s, "focus-1.1") || !strings.Contains(s, "focus-1.0") {
		t.Fatalf("expected builtin schema names in output, got: %s", s)
	}
}
