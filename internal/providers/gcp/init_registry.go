package gcp

import (
	"github.com/costscope/costscope/internal/providers/registry"
	"github.com/costscope/costscope/internal/providers/types"
)

func init() {
	registry.Register("gcp", func(opts ...interface{}) (registry.Provider, error) {
		if len(opts) == 0 {
			return nil, nil
		}
		if cfg, ok := opts[0].(*types.ProviderConfig); ok {
			return NewGCPProvider(cfg)
		}
		return nil, nil
	})
}
