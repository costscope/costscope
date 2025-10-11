package analytics

import (
	"context"
	"testing"

	"local/costscope/internal/database"
)

// fake executor captures the last query
type capturingExec struct{ last string }

func (f *capturingExec) ExecuteQuery(ctx context.Context, query string) (*database.QueryResult, error) {
	f.last = query
	return &database.QueryResult{Columns: []string{}, Data: []map[string]interface{}{}}, nil
}

func TestFacade_TenantFilterApplied(t *testing.T) {
	exec := &capturingExec{}
	f := NewFacade(exec)
	tenant := "tenant-qwe"
	filters := &database.AnalyticsFilters{TenantID: tenant}
	if _, err := f.CostSummary(context.Background(), filters); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exec.last == "" {
		t.Fatalf("expected a query to be executed")
	}
	if !contains(exec.last, "tenant_id = '"+tenant+"'") {
		t.Fatalf("expected tenant filter in query, got: %s", exec.last)
	}
}

func contains(h, n string) bool {
	if len(n) == 0 {
		return true
	}
	for i := 0; i+len(n) <= len(h); i++ {
		if h[i:i+len(n)] == n {
			return true
		}
	}
	return false
}
