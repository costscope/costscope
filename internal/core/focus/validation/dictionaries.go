package validation

// Lightweight dictionaries and normalization helpers for currencies and regions.
// This intentionally covers a practical subset to avoid heavy dependencies.

import (
	"strings"
)

// KnownCurrencies is a minimal ISO-4217 set commonly seen in cloud billing.
var KnownCurrencies = map[string]struct{}{
	"USD": {}, "EUR": {}, "GBP": {}, "JPY": {}, "CAD": {}, "AUD": {},
	"INR": {}, "CNY": {}, "CHF": {}, "SEK": {}, "NOK": {}, "DKK": {},
}

// NormalizeCurrency uppercases the input and validates it against KnownCurrencies.
// Returns normalized value and whether it is valid/known.
func NormalizeCurrency(in string) (string, bool) {
	if in == "" {
		return in, false
	}
	up := strings.ToUpper(strings.TrimSpace(in))
	_, ok := KnownCurrencies[up]
	return up, ok
}

// Region canonicalization across providers (sample subset only).
// Key: lowercased trimmed variant; Value: canonical code.
// (legacy regionAliases removed; provider-aware mapping handled in normalization package)

// Region normalization is performed only at conversion time (converter mappers).
// Validation layer deliberately does not mutate region values to avoid masking
// ingestion issues; previous NormalizeRegion shim removed as dead code.
