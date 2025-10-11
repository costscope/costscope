package providers_test

import (
	"testing"

	"local/costscope/internal/providers/aws"
	"local/costscope/internal/providers/azure"
	"local/costscope/internal/providers/gcp"
	"local/costscope/internal/providers/testutils"
	"local/costscope/internal/providers/types"
)

// TestTestutilsUsage provides an integration-style smoke test showing how the shared
// testutils helpers can be used to stand up provider instances with fake credentials
// in a uniform fashion. This also serves as living documentation so deadcode tooling
// recognizes the helpers as intentionally used.
func TestTestutilsUsage(t *testing.T) {
	cases := []struct {
		name    string
		ptype   types.ProviderType
		creds   map[string]string
		builder func(cfg *types.ProviderConfig) (interface{}, error)
	}{
		{
			name:  "aws",
			ptype: types.ProviderTypeAWS,
			creds: testutils.GetAWSTestCredentials(),
			builder: func(cfg *types.ProviderConfig) (interface{}, error) {
				return aws.NewAWSProvider(cfg)
			},
		},
		{
			name:  "gcp",
			ptype: types.ProviderTypeGCP,
			creds: testutils.GetGCPTestCredentials(),
			builder: func(cfg *types.ProviderConfig) (interface{}, error) {
				return gcp.NewGCPProvider(cfg)
			},
		},
		{
			name:  "azure",
			ptype: types.ProviderTypeAzure,
			creds: testutils.GetAzureTestCredentials(),
			builder: func(cfg *types.ProviderConfig) (interface{}, error) {
				return azure.NewAzureProvider(cfg)
			},
		},
	}

	for _, c := range cases {
		c := c
		cfg := testutils.CreateTestProviderConfig("test-"+c.name, c.ptype, c.creds)
		provider, err := c.builder(cfg)
		testutils.AssertProviderCreation(t, provider, err, "test-"+c.name)
	}
}
