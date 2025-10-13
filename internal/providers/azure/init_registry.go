package azure

import (
	"github.com/costscope/costscope/internal/providers/registry"
	"github.com/costscope/costscope/internal/providers/types"
)

func init() {
	registry.Register("azure", func(opts ...interface{}) (registry.Provider, error) {
		if len(opts) == 0 {
			return nil, nil
		}
		if cfg, ok := opts[0].(*types.ProviderConfig); ok {
			return NewAzureProvider(cfg)
		}
		return nil, nil
	})
}
