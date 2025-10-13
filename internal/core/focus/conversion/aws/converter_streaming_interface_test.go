package aws

import (
	"testing"

	"github.com/costscope/costscope/internal/core/focus/types"
)

// Compile-time assertion that aws.Converter implements types.StreamingConverter.
func TestAWSConverterImplementsStreaming(t *testing.T) {
	var _ types.StreamingConverter = (*Converter)(nil)
}
