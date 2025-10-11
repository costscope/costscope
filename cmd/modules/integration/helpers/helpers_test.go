package helpers

import (
	"errors"
	"testing"
	"time"

	"github.com/spf13/cobra"
)

func TestGenerateID(t *testing.T) {
	id := GenerateID("conn")
	if len(id) == 0 || id[:5] != "conn_" {
		t.Fatalf("unexpected id: %s", id)
	}
}

func TestRunEWithLoggingSuccess(t *testing.T) {
	ran := false
	cmd := &cobra.Command{Use: "test"}
	wrapper := RunEWithLogging("test", func(cmd *cobra.Command, args []string) error {
		ran = true
		return nil
	})
	if err := wrapper(cmd, []string{"a"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ran {
		t.Fatal("wrapped function did not run")
	}
}

func TestRunEWithLoggingError(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	want := errors.New("boom")
	wrapper := RunEWithLogging("test", func(cmd *cobra.Command, args []string) error {
		time.Sleep(10 * time.Millisecond)
		return want
	})
	if err := wrapper(cmd, nil); err == nil {
		t.Fatal("expected error, got nil")
	}
}
