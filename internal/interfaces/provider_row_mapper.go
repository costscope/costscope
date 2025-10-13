package interfaces

import (
	"github.com/costscope/costscope/internal/core/focus/types"
)

// ProviderRowMapper defines a minimal, provider-agnostic contract for mapping
// a raw provider row (CSV/JSON fields materialized as []string) into a FOCUS
// v1.2 FocusRecord. Implementations should perform provider-specific field
// extraction, normalization, and classification, returning a fully-populated
// FocusRecord or an error describing a row-level mapping issue.
//
// Notes:
//   - Errors returned here should not generally abort the entire pipeline; callers
//     may count and continue (parity with existing converters).
//   - Implementations SHOULD avoid side effects; this interface is designed for
//     isolated, table-driven unit tests across providers.
type ProviderRowMapper interface {
	Map(row []string) (types.FocusRecord, error)
}
