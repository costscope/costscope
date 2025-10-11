package parity

import (
	"context"
	"testing"

	conversion "local/costscope/internal/core/focus/conversion"
	"local/costscope/internal/core/focus/types"
)

// TestUniversalMinimalInvalidConfig ensures Convert returns error on invalid config
func TestUniversalMinimalInvalidConfig(t *testing.T) {
	t.Parallel()
	cm := conversion.NewConversionManager(2)
	uc := cm.GetConverter()
	_, err := uc.Convert(context.Background(), &types.ConversionConfig{Provider: "", InputPath: "", OutputPath: ""})
	if err == nil {
		t.Fatalf("expected error for invalid config")
	}
}
