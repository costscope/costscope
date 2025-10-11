package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// helper to write a temp yaml file
func writeTempConfig(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write temp config: %v", err)
	}
	return path
}

func TestRun_SuccessAndFailure(t *testing.T) {
	tmp := t.TempDir()
	// valid minimal config
	valid := `environment: development
data_dir: ./data
core:
  app_name: costscope
  version: 1.0.0
  log_level: debug
  data_directory: ./data
  temp_directory: ./temp
  max_concurrency: 1
  timeout: 1s
providers:
  aws:
    enabled: false
  azure:
    enabled: false
  gcp:
    enabled: false
database:
  type: sqlite
  connection_string: ./dev.db
  max_connections: 1
  max_idle_connections: 0
  connection_timeout: 1s
  query_timeout: 1s
streaming:
  max_concurrent_jobs: 1
  default_workers: 1
  default_memory: 1
  job_timeout: 1s
  progress_interval: 1s
security:
  encryption_enabled: false
  tls_enabled: false
  token_expiry: 1s
  jwt_secret: secret
`
	invalid := `environment: development
data_dir: "" # invalid empty
core:
  app_name: costscope
  version: 1.0.0
  log_level: debug
  data_directory: ./data
  temp_directory: ./temp
  max_concurrency: 1
  timeout: 1s
providers:
  aws:
    enabled: false
  azure:
    enabled: false
  gcp:
    enabled: false
database:
  type: sqlite
  connection_string: ./dev.db
  max_connections: 1
  max_idle_connections: 0
  connection_timeout: 1s
  query_timeout: 1s
streaming:
  max_concurrent_jobs: 1
  default_workers: 1
  default_memory: 1
  job_timeout: 1s
  progress_interval: 1s
security:
  encryption_enabled: false
  tls_enabled: false
  token_expiry: 1s
  jwt_secret: secret
`

	validPath := writeTempConfig(t, tmp, "valid.yaml", valid)
	invalidPath := writeTempConfig(t, tmp, "invalid.yaml", invalid)
	if validPath == invalidPath { // impossible, but ensures both return values are referenced for static analysis
		t.Fatalf("unexpected identical temp paths: %s", validPath)
	}

	rep, err := run(tmp)
	if err != nil {
		t.Fatalf("run returned error: %v", err)
	}
	if rep.Summary.Total != 2 {
		b, _ := json.Marshal(rep)
		t.Fatalf("expected 2 files, got %d: %s", rep.Summary.Total, string(b))
	}
	if rep.Summary.Failed != 1 {
		b, _ := json.Marshal(rep)
		t.Fatalf("expected 1 failed, got %d: %s", rep.Summary.Failed, string(b))
	}
	// ensure invalid contains expected error about data_dir
	found := false
	for _, f := range rep.Files {
		if !f.Valid {
			for _, e := range f.Errors {
				if e.Key == "data_dir" {
					found = true
				}
			}
		}
	}
	if !found {
		t.Fatalf("expected data_dir error not found: %+v", rep.Files)
	}
}
