//go:build !duckdb
// +build !duckdb

package duckdb

// Linked reports whether this binary was built with DuckDB support.
func Linked() bool { return false }

// QuickPing is unavailable without DuckDB; it returns nil to avoid failing slim builds.
// Callers should rely on Linked() to report the linkage state.
func QuickPing() error { return nil }
