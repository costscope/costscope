package multicloud

import (
	"context"
	"testing"
	"time"

	"github.com/costscope/costscope/internal/core/logging"
	"github.com/costscope/costscope/internal/providers"
)

// TestMulticloudService_StatusAndDisabledFlags covers GetServiceStatus and disabled feature branches.
func TestMulticloudService_StatusAndDisabledFlags(t *testing.T) {
	pm := providers.NewProviderManager()
	svc := NewMulticloudService(pm, logging.NewLogger(logging.LevelError))
	// Mutate config flags (read under lock in methods)
	svc.config.EnableOptimizations = false
	svc.config.EnableMigrations = false

	// Optimization disabled path
	if _, err := svc.AnalyzeOptimizations(context.Background(), &OptimizationRequest{Providers: []string{"aws"}, StartDate: time.Now().Add(-24 * time.Hour), EndDate: time.Now()}); err == nil {
		t.Fatalf("expected optimization disabled error")
	}
	// Migration disabled path
	if _, err := svc.EstimateMigrationCosts(context.Background(), &MigrationRequest{SourceProvider: "aws", TargetProvider: "gcp", MigrationTimeframe: 24 * time.Hour}); err == nil {
		t.Fatalf("expected migration disabled error")
	}

	st := svc.GetServiceStatus()
	if st == nil || st.ServiceName != "multicloud" {
		t.Fatalf("unexpected status: %+v", st)
	}
	if st.OptimizationsEnabled || st.MigrationsEnabled {
		t.Fatalf("expected flags disabled in status")
	}
}
