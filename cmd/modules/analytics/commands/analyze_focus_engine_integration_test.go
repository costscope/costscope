package commands

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/costscope/costscope/internal/core/analytics"
	"github.com/costscope/costscope/internal/core/logging"

	"github.com/spf13/cobra"
)

// helper to build the generated analytics command (subset) for tests
func buildTestAnalyticsCmd() *cobra.Command {
	logger := logging.NewLogger("error")
	svc := analytics.NewBasicService(&analytics.Config{MLEnabled: true, AnomalyDetection: true, TrendAnalysis: true, EnablePredictions: true}, logger)
	ac := NewAnalyticsCommands(logger, svc)
	return ac.BuildAnalyticsCommand()
}

// TestFocusEngineExtendedPresence ensures the extended block is only present when --use-focus-engine is set.
func TestFocusEngineExtendedPresence(t *testing.T) {
	tmp, err := os.CreateTemp(t.TempDir(), "focus-dataset-*.parquet")
	if err != nil {
		t.Fatalf("temp file: %v", err)
	}
	_ = tmp.Close()

	capture := func(fn func()) string {
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w
		fn()
		if err := w.Close(); err != nil {
			t.Fatalf("close pipe writer: %v", err)
		}
		os.Stdout = old
		var b bytes.Buffer
		_, _ = io.Copy(&b, r)
		return b.String()
	}

	// Without flag
	outNoFlag := capture(func() {
		root := buildTestAnalyticsCmd()
		root.SetArgs([]string{"analyze", "--table", tmp.Name(), "--currency", "USD"})
		if err := root.Execute(); err != nil {
			t.Fatalf("execute analyze default: %v", err)
		}
	})
	if strings.Contains(outNoFlag, "\"extended\"") {
		t.Fatalf("extended section unexpectedly present without --use-focus-engine")
	}

	// With flag
	outWithFlag := capture(func() {
		root2 := buildTestAnalyticsCmd()
		root2.SetArgs([]string{"analyze", "--table", tmp.Name(), "--use-focus-engine", "--focus-forecast-days", "5"})
		if err := root2.Execute(); err != nil {
			t.Fatalf("execute analyze focus engine: %v", err)
		}
	})
	if !strings.Contains(outWithFlag, "\"extended\"") {
		t.Fatalf("extended section missing with --use-focus-engine; output=%s", outWithFlag)
	}
	// New fields: trends & optimizations
	if !strings.Contains(outWithFlag, "\"trends\"") {
		t.Fatalf("trends field missing in extended output: %s", outWithFlag)
	}
	if !strings.Contains(outWithFlag, "\"optimizations\"") {
		t.Fatalf("optimizations field missing in extended output: %s", outWithFlag)
	}
}
