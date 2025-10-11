//go:build enterprise

package database

import (
	"context"
	"testing"

	"local/costscope/internal/core/logging"
)

func TestEnterpriseConnectionManager_PoolMetrics(t *testing.T) {
	logger := logging.NewLogger(logging.LevelInfo)
	mgr := NewEnterpriseConnectionManager(logger)
	ctx := context.Background()

	if err := mgr.CreateConnectionPool(ctx, "testpool", "duckdb", "file:memory"); err != nil {
		t.Fatalf("CreateConnectionPool failed: %v", err)
	}

	metrics, err := mgr.GetPoolMetrics("testpool")
	if err != nil {
		t.Fatalf("GetPoolMetrics failed: %v", err)
	}
	if metrics == nil {
		t.Fatalf("expected metrics, got nil")
	}

	// Stop removed: resources released via GC; no explicit shutdown required.
}

func TestEnterpriseConnectionManager_ManagerMetrics(t *testing.T) { // enterprise-only health accessor test
	logger := logging.NewLogger(logging.LevelInfo)
	mgr := NewEnterpriseConnectionManager(logger)

	metrics := mgr.GetManagerMetrics()
	if metrics == nil { // Should always return a struct even before pools are created
		t.Fatalf("expected non-nil manager metrics")
	}

	// Create a pool and ensure the metrics reflect it
	ctx := context.Background()
	if err := mgr.CreateConnectionPool(ctx, "pool2", "duckdb", "file:memory"); err != nil {
		t.Fatalf("CreateConnectionPool failed: %v", err)
	}

	updated := mgr.GetManagerMetrics()
	if updated.TotalPools < 1 {
		t.Fatalf("expected at least 1 pool, got %d", updated.TotalPools)
	}
	// Stop removed: resources released via GC; no explicit shutdown required.
}
