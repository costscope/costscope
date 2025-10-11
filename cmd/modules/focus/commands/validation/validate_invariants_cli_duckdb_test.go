//go:build duckdb

package validation

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// writeTempJSONFocus writes NDJSON focus rows (minimal fields used by invariants)
func writeTempJSONFocus(t *testing.T, dir, name string, rows []map[string]interface{}) string { //nolint:gosec // test helper, controlled input
	t.Helper()
	p := filepath.Join(dir, name)
	f, err := os.Create(p)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	enc := json.NewEncoder(f)
	for _, r := range rows {
		if err := enc.Encode(r); err != nil {
			t.Fatalf("encode: %v", err)
		}
	}
	_ = f.Close()
	return p
}

func TestValidateCLI_InvariantsNoBaseline(t *testing.T) {
	if testing.Short() {
		t.Skip("short")
	}
	dir := t.TempDir()
	rows := []map[string]interface{}{
		{"effective_cost": 1.0, "list_cost": 1.1, "usage_quantity": 10.0, "charge_category": "Usage", "pricing_category": "OnDemand", "provider_name": "aws", "resource_id": "r1"},
		{"effective_cost": 2.0, "list_cost": 2.2, "usage_quantity": 5.0, "charge_category": "Usage", "pricing_category": "OnDemand", "provider_name": "aws", "resource_id": "r2"},
	}
	// filename includes "test" to keep core validators happy as per existing CLI tests
	input := writeTempJSONFocus(t, dir, "test-focus.json", rows)
	cmd := BuildValidateCommand()
	// Run schema+quality only, enable invariants; quiet to reduce noise
	if err := executeCommand(cmd, input, "--format", "json", "--schema", "--quality", "--invariants", "--quiet"); err != nil {
		t.Fatalf("validate cli failed: %v", err)
	}
}

func TestValidateCLI_InvariantsBaselineViolation(t *testing.T) {
	if testing.Short() {
		t.Skip("short")
	}
	dir := t.TempDir()
	current := []map[string]interface{}{
		{"effective_cost": 10.0, "list_cost": 10.0, "usage_quantity": 1.0, "charge_category": "Usage", "pricing_category": "OnDemand", "provider_name": "aws", "resource_id": "r1"},
	}
	input := writeTempJSONFocus(t, dir, "test-current.json", current)
	baselineMetrics := `{"row_count":1,"sum_effective_cost":1.0,"sum_list_cost":1.0,"sum_usage_quantity":1.0,"charge_category_distribution":{"Usage":100},"pricing_category_distribution":{"OnDemand":100},"provider_distribution":{"aws":100},"generated_at":"2024-01-01T00:00:00Z"}`
	basePath := filepath.Join(dir, "baseline.json")
	if err := os.WriteFile(basePath, []byte(baselineMetrics), 0600); err != nil {
		t.Fatalf("write baseline: %v", err)
	}
	cmd := BuildValidateCommand()
	// Expect an error due to invariants drift vs baseline
	if err := executeCommand(cmd, input, "--format", "json", "--schema", "--quality", "--invariants", "--invariants-baseline", basePath, "--invariants-tolerance", "0.0001", "--quiet"); err == nil {
		t.Fatalf("expected invariants drift error, got nil")
	}
}
