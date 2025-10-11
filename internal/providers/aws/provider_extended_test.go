package aws

import (
	"context"
	"errors"
	"strings"
	"testing"

	prtypes "local/costscope/internal/providers/types"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

// fakeSTS implements minimal GetCallerIdentity for deterministic tests without network.
type fakeSTS struct{ fail bool }

func (f fakeSTS) GetCallerIdentity(ctx context.Context, in *sts.GetCallerIdentityInput, optFns ...func(*sts.Options)) (*sts.GetCallerIdentityOutput, error) {
	if f.fail {
		return nil, errors.New("simulated sts failure")
	}
	acc := "123456789012"
	arn := "arn:aws:iam::123456789012:user/test"
	return &sts.GetCallerIdentityOutput{Account: &acc, Arn: &arn, UserId: aws.String("AIDAEXAMPLE")}, nil
}

// injectProviderIdentity patches the provider config with a supplied STS client (via config replacement).
// Since provider uses p.config directly, we create a new provider then override its config's credential provider
// and call GetProviderInfo through a shim that swaps the global STS constructor. To avoid invasive code changes,
// we wrap logic in a helper calling internal method pattern (kept minimal to avoid altering prod code).

func newProviderWithStaticCreds(t *testing.T) *AWSProvider {
	cfg := &prtypes.ProviderConfig{Name: "cov-aws", Type: prtypes.ProviderTypeAWS, Credentials: map[string]string{"access_key": "AKIAFAKE", "secret_key": "SECRETFAKE"}}
	p, err := NewAWSProvider(cfg)
	if err != nil {
		t.Fatalf("NewAWSProvider: %v", err)
	}
	return p
}

func TestGetProviderInfo_SuccessPathSimulated(t *testing.T) {
	p := newProviderWithStaticCreds(t)
	p.stsClient = fakeSTS{fail: false}
	info, err := p.GetProviderInfo(context.Background())
	if err != nil {
		t.Fatalf("GetProviderInfo error: %v", err)
	}
	if info.Metadata["account_id"] != "123456789012" {
		t.Fatalf("expected account 123456789012, got %s", info.Metadata["account_id"])
	}
	if info.Metadata["region"] == "" {
		t.Fatalf("expected region set")
	}
}

func TestValidateCredentials_ErrorBranches(t *testing.T) {
	cfg := &prtypes.ProviderConfig{Name: "cov-aws", Type: prtypes.ProviderTypeAWS, Credentials: map[string]string{"access_key": "AKIAFAKE", "secret_key": "SECRETFAKE"}}
	p, _ := NewAWSProvider(cfg)
	// missing both
	if err := p.ValidateCredentials(context.Background(), map[string]string{}); err == nil || !strings.Contains(err.Error(), "access_key") {
		t.Fatalf("expected access_key missing error, got %v", err)
	}
	// missing secret
	if err := p.ValidateCredentials(context.Background(), map[string]string{"access_key": "A"}); err == nil || !strings.Contains(err.Error(), "secret_key") {
		t.Fatalf("expected secret_key missing error, got %v", err)
	}
}

func TestAWSDataHelpers_SampleSizes(t *testing.T) {
	p := newProviderWithStaticCreds(t)
	costs, _ := p.GetCostData(context.Background(), prtypes.CostDataParams{})
	if len(costs) != 2 {
		t.Fatalf("expected 2 sample cost records, got %d", len(costs))
	}
	resources, _ := p.GetResourceData(context.Background(), prtypes.ResourceDataParams{})
	if len(resources) != 2 {
		t.Fatalf("expected 2 sample resource records, got %d", len(resources))
	}
}
