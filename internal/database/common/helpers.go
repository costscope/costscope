package common

// NOTE: These lightweight clause helpers are intentionally kept in a tiny
// shared package so multiple query builder implementations (FOCUS, DuckDB,
// future engines) can delegate formatting concerns consistently without
// duplicating simple string assembly logic. Only helpers that are *actually*
// consumed by active builders should live here. When adding new helpers prefer
// first implementing the feature directly inside a builder; promote to this
// package only if the exact logic is required by more than one builder.
//
// If a helper becomes unused (verify with `make deadcode`), remove it to keep
// the surface minimal. All helpers MUST have explicit unit tests in
// helpers_test.go and at least one integration style assertion in a builder
// test demonstrating end-to-end SQL assembly.

import (
	"fmt"
	"strings"
)

// EqOrIn builds an equality predicate or IN list for one or more values.
// Returns empty string when values slice is empty. Values are assumed pre-sanitized.
// Rationale: kept as the sole shared helper because both focus & duckdb builders
// need the identical branching logic (0 / 1 / many) and prior duplication caused
// divergence + test redundancy. Simpler one-liner helpers (order by, left join,
// cost threshold) were inlined to reduce false deadcode reports in slim builds
// (where extended JOIN paths are not compiled) and shrink the shared surface.
func EqOrIn(column string, values []string) string {
	switch len(values) {
	case 0:
		return ""
	case 1:
		return fmt.Sprintf("%s = '%s'", column, values[0])
	default:
		quoted := "'" + strings.Join(values, "','") + "'"
		return fmt.Sprintf("%s IN (%s)", column, quoted)
	}
}
