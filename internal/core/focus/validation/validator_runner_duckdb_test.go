//go:build duckdb

package validation

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// helper to write a small JSON file readable by DuckDB read_json_auto
func writeTempJSONFocus(t *testing.T, dir string, name string, rows []map[string]interface{}) string { //nolint:gosec // test helper writing controlled JSON fixture
	t.Helper()
	path := filepath.Join(dir, name)
	f, err := os.Create(path) //nolint:gosec // path built from TempDir + fixed filename; test-controlled
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
	return path
}

func TestRunValidation_InvariantsNoBaseline(t *testing.T) {
	if testing.Short() {
		t.Skip("short")
	}
	dir := t.TempDir()
	data := []map[string]interface{}{
		{"effective_cost": 1.0, "list_cost": 1.2, "usage_quantity": 10.0, "charge_category": "Usage", "pricing_category": "OnDemand", "provider_name": "aws", "resource_id": "r1"},
		{"effective_cost": 2.0, "list_cost": 2.1, "usage_quantity": 5.0, "charge_category": "Usage", "pricing_category": "OnDemand", "provider_name": "aws", "resource_id": "r2"},
	}
	path := writeTempJSONFocus(t, dir, "focus.json", data)
	opts := ValidationOpts{InputPath: path, Spec: "v1.2", FormatHint: "json", RunSchema: true, RunQuality: true, RunPerformance: false, RunAnomalies: false, InvariantsEnabled: true}
	res, err := RunValidation(opts)
	if err != nil {
		t.Fatalf("RunValidation error: %v", err)
	}
	if res.Invariants == nil {
		t.Fatalf("expected invariants computed")
	}
	if res.Invariants.RowCount != 2 {
		t.Fatalf("rowcount= %d", res.Invariants.RowCount)
	}
}

func TestRunValidation_InvariantsBaselineViolation(t *testing.T) {
	if testing.Short() {
		t.Skip("short")
	}
	dir := t.TempDir()
	current := []map[string]interface{}{
		{"effective_cost": 10.0, "list_cost": 10.0, "usage_quantity": 1.0, "charge_category": "Usage", "pricing_category": "OnDemand", "provider_name": "aws", "resource_id": "r1"},
	}
	curPath := writeTempJSONFocus(t, dir, "cur.json", current)
	baselineMetrics := `{"row_count":1,"sum_effective_cost":1.0,"sum_list_cost":1.0,"sum_usage_quantity":1.0,"charge_category_distribution":{"Usage":100},"pricing_category_distribution":{"OnDemand":100},"provider_distribution":{"aws":100},"generated_at":"2024-01-01T00:00:00Z"}`
	basePath := filepath.Join(dir, "baseline.json")
	if err := os.WriteFile(basePath, []byte(baselineMetrics), 0600); err != nil {
		t.Fatalf("write baseline: %v", err)
	}
	opts := ValidationOpts{InputPath: curPath, Spec: "v1.2", FormatHint: "json", RunSchema: true, RunQuality: true, InvariantsEnabled: true, InvariantsBaseline: basePath, InvariantsTolerance: 0.0001}
	res, err := RunValidation(opts)
	if err == nil {
		t.Fatalf("expected violation error")
	}
	if res.Invariants == nil || len(res.Invariants.Violations) == 0 {
		t.Fatalf("expected violations recorded")
	}
}
