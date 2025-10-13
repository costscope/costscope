package multicloud

import (
	"context"
	"testing"
	"time"

	"github.com/costscope/costscope/internal/core/logging"
	"github.com/costscope/costscope/internal/providers"
	"github.com/costscope/costscope/internal/providers/types"
)

// stubProvider implements types.CloudProvider for tests
type stubProvider struct{ name string }

func (s *stubProvider) ValidateCredentials(ctx context.Context, config map[string]string) error {
	return nil
}
func (s *stubProvider) GetProviderInfo(ctx context.Context) (types.ProviderInfo, error) {
	return types.ProviderInfo{Name: s.name, Type: types.ProviderType(s.name), SupportedRegions: []string{"us-east-1"}}, nil
}
func (s *stubProvider) GetCostData(ctx context.Context, params types.CostDataParams) ([]types.CostRecord, error) {
	return nil, nil
}
func (s *stubProvider) GetResourceData(ctx context.Context, params types.ResourceDataParams) ([]types.ResourceRecord, error) {
	return nil, nil
}
func (s *stubProvider) GetName() string               { return s.name }
func (s *stubProvider) GetType() types.ProviderType   { return types.ProviderType(s.name) }
func (s *stubProvider) GetSupportedRegions() []string { return []string{"us-east-1"} }

// helper to build service with registered providers
func newTestService(t *testing.T) *MulticloudService {
	pm := providers.NewProviderManager()
	logger := logging.NewLogger(logging.LevelError)
	for _, p := range []string{"aws", "azure", "gcp"} {
		if err := pm.RegisterProvider(p, &stubProvider{name: p}, &types.ProviderConfig{Name: p, Type: types.ProviderType(p), Credentials: map[string]string{}}); err != nil {
			t.Fatalf("register provider %s: %v", p, err)
		}
	}
	return NewMulticloudService(pm, logger)
}

func TestMulticloudService_OptimizationAndRecommendations(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()

	optReq := &OptimizationRequest{Providers: []string{"aws", "azure"}, StartDate: time.Now().AddDate(0, 0, -7), EndDate: time.Now()}
	optRes, err := svc.AnalyzeOptimizations(ctx, optReq)
	if err != nil {
		t.Fatalf("AnalyzeOptimizations error: %v", err)
	}
	if len(optRes.Providers) != 2 {
		t.Errorf("expected 2 providers, got %d", len(optRes.Providers))
	}

	recReq := &RecommendationRequest{Providers: []string{"aws"}, MaxRecommendations: 5}
	recRes, err := svc.GetRecommendations(ctx, recReq)
	if err != nil {
		t.Fatalf("GetRecommendations error: %v", err)
	}
	if recRes.TotalRecommendations == 0 {
		t.Errorf("expected recommendations >0")
	}
}

func TestMulticloudService_Migration(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()
	migReq := &MigrationRequest{SourceProvider: "aws", TargetProvider: "azure", MigrationTimeframe: 30 * 24 * time.Hour}
	if _, err := svc.EstimateMigrationCosts(ctx, migReq); err != nil {
		t.Fatalf("EstimateMigrationCosts error: %v", err)
	}
	if _, err := svc.GenerateMigrationPlan(ctx, migReq); err != nil {
		t.Fatalf("GenerateMigrationPlan error: %v", err)
	}
}

func TestMulticloudService_DiscoveryInventory(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()
	discReq := &DiscoveryRequest{Providers: []string{"aws", "gcp"}, IncludeMetadata: true}
	discRes, err := svc.DiscoverResources(ctx, discReq)
	if err != nil {
		t.Fatalf("DiscoverResources error: %v", err)
	}
	if discRes.TotalResources == 0 {
		t.Errorf("expected discovered resources >0")
	}
	inv, err := svc.GetUnifiedInventory(ctx, nil)
	if err != nil {
		t.Fatalf("GetUnifiedInventory error: %v", err)
	}
	if len(inv.Providers) == 0 {
		t.Errorf("expected inventory providers >0")
	}
}

func TestMulticloudService_CostComparison(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()
	cmpReq := &CostComparisonRequest{Providers: []string{"aws", "azure"}, StartDate: time.Now().AddDate(0, 0, -1), EndDate: time.Now(), Currency: "USD"}
	res, err := svc.CompareCosts(ctx, cmpReq)
	if err != nil {
		t.Fatalf("CompareCosts error: %v", err)
	}
	if len(res.Providers) == 0 {
		t.Errorf("expected provider cost data")
	}
}
