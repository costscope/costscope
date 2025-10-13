//go:build experimental

package focus

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/costscope/costscope/internal/core/logging"
)

// write minimal YAML including focus.use_unified_mapper_default
func writeYAML(t *testing.T, dir string, enabled bool) string {
	t.Helper()
	val := "false"
	if enabled {
		val = "true"
	}
	content := "environment: development\ncore:\n  app_name: a\n  version: v\n  log_level: info\n  data_directory: " + filepath.Join(dir, "d") + "\n  temp_directory: " + filepath.Join(dir, "t") + "\n  max_concurrency: 1\n  timeout: 1s\nproviders:\n  aws:\n    enabled: true\n    region: us-east-1\n    max_retries: 1\n    request_timeout: 1s\ndatabase:\n  type: sqlite\n  connection_string: x\n  max_connections: 1\n  max_idle_connections: 1\n  connection_timeout: 1s\n  query_timeout: 1s\n  migrations_path: " + filepath.Join(dir, "m") + "\n  auto_migrate: false\nstreaming:\n  max_concurrent_jobs: 1\n  default_workers: 1\n  default_memory: 128\n  job_timeout: 1s\n  progress_interval: 1s\n  checkpoint_enabled: false\n  checkpoint_dir: " + filepath.Join(dir, "c") + "\nsecurity:\n  encryption_enabled: false\n  tls_enabled: false\n  token_expiry: 1s\nfocus:\n  use_unified_mapper_default: " + val + "\n"
	p := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatalf("write yaml: %v", err)
	}
	return p
}

func TestResolveUseUnifiedMapper_Precedence(t *testing.T) {
	logger := logging.NewLogger(logging.LevelError)

	// 1) Request wins when provided
	t.Run("request_wins", func(t *testing.T) {
		tru := true
		opts := &ConvertOptions{UseUnifiedMapper: &tru}
		if got := resolveUseUnifiedMapper(opts, logger); !got {
			t.Fatalf("expected true from request pointer")
		}
	})

	// 2) YAML beats ENV when request omitted
	t.Run("yaml_beats_env", func(t *testing.T) {
		tmp := t.TempDir()
		cfg := writeYAML(t, tmp, true)
		// set both COSTSCOPE_CONFIG and env to false; YAML should make it true
		t.Setenv("COSTSCOPE_CONFIG", cfg)
		t.Setenv("COSTSCOPE_USE_UNIFIED_MAPPER", "false")
		if got := resolveUseUnifiedMapper(&ConvertOptions{}, logger); !got {
			t.Fatalf("expected true from YAML default over ENV false")
		}
	})

	// 3) ENV used when YAML missing
	t.Run("env_when_no_yaml", func(t *testing.T) {
		t.Setenv("COSTSCOPE_CONFIG", "")
		t.Setenv("COSTSCOPE_USE_UNIFIED_MAPPER", "true")
		if got := resolveUseUnifiedMapper(&ConvertOptions{}, logger); !got {
			t.Fatalf("expected true from ENV when YAML missing")
		}
	})
}
