package analytics

import (
	"context"
	"strings"
	"testing"

	"local/costscope/internal/database"
)

// sqlCapturingExec captures the executed SQL for inspection.
type sqlCapturingExec struct {
	lastQuery string
	res       *database.QueryResult
}

func (e *sqlCapturingExec) ExecuteQuery(ctx context.Context, query string) (*database.QueryResult, error) {
	e.lastQuery = query
	if e.res != nil {
		return e.res, nil
	}
	return &database.QueryResult{Columns: []string{"service_name", "total_cost", "resource_count"}, Data: []map[string]interface{}{}}, nil
}

func TestFacadeTopServices_SQLShape(t *testing.T) {
	exec := &sqlCapturingExec{res: &database.QueryResult{Columns: []string{"service_name", "total_cost", "resource_count"}, Data: []map[string]interface{}{}}}
	fac := NewFacade(exec)
	_, err := fac.TopServices(context.Background(), nil, 7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	sql := exec.lastQuery
	mustContain := func(sub string) {
		if !strings.Contains(sql, sub) {
			t.Fatalf("expected SQL to contain %q. Full SQL: %s", sub, sql)
		}
	}
	mustContain("SELECT service_name, SUM(effective_cost) as total_cost, COUNT(*) as resource_count FROM focus_cost_data")
	mustContain("GROUP BY service_name")
	mustContain("ORDER BY total_cost DESC")
	mustContain("LIMIT 7")
}
