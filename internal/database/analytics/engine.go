//go:build deprecated_analytics

package analytics

// This file intentionally contains no executable code.
// The legacy AnalyticsEngine implementation has been retired in favor of the
// thin analytics Facade (see facade.go) built directly on the FOCUS QueryBuilder + DuckDB.
// This file is excluded from normal builds via the 'deprecated_analytics' build tag to
// prevent dead code from being compiled while preserving history and blame.
