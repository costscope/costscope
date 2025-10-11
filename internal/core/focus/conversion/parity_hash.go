package conversion

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
)

// HashFocusLite computes a stable, order‑independent hash over key parity fields for a slice
// of FocusRecordLite (effective_cost, usage_quantity, provider_name, service_name, charge_category).
//
// Rationale / Duplication Note:
// The parity-check script (scripts/tools/parity-check) contains a similar function (computeLiteHash)
// which derives the hash directly from a DuckDB parquet scan without materializing FocusRecordLite.
// This helper is retained inside the conversion package for unit / mapper parity tests where we already
// hold in-memory lite projections and wish to avoid spinning up DuckDB. Keeping both implementations
// minimizes test friction while allowing the CLI tool to remain self-contained. Should we standardize
// later, we can export a streaming variant and de-duplicate; until then, any change to the formatting
// string MUST be mirrored in the tool (and vice versa) to preserve hash comparability in CI.
//
// Risk if modified:
// - Divergent formatting between this helper and computeLiteHash would yield false parity failures.
// - Added / reordered fields would break historical comparisons and cached baseline artifacts.
// Mitigation: keep the exact formatting token sequence and precision (%.6f) aligned.
func HashFocusLite(recs []FocusRecordLite) string {
	parts := make([]string, 0, len(recs))
	for _, r := range recs {
		parts = append(parts, fmt.Sprintf("%.6f|%.6f|%s|%s|%s", r.EffectiveCost, r.UsageQuantity, r.ProviderName, r.ServiceName, r.ChargeCategory))
	}
	sort.Strings(parts)
	h := sha256.New()
	for _, p := range parts {
		_, _ = h.Write([]byte(p))
	}
	return hex.EncodeToString(h.Sum(nil))
}
