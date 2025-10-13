//go:build !enterprise

package database

import (
	"context"
	"testing"

	"github.com/costscope/costscope/internal/core/enterprise"
	"github.com/costscope/costscope/internal/core/logging"
)

func TestEnterpriseConnectionManagerStub_Disabled(t *testing.T) {
	logger := logging.NewLogger(logging.LevelError)
	mgr := NewEnterpriseConnectionManager(logger)
	if err := mgr.CreateConnectionPool(context.Background(), "pool1", "duckdb", "memory"); !enterprise.IsDisabled(err) {
		t.Fatalf("expected disabled error, got %v", err)
	}
}
