//go:build !duckdb
// +build !duckdb

// Package duckdb provides DuckDB analytics integration when built with the
// 'duckdb' build tag. This stub exists to allow building and testing the
// repository without pulling in DuckDB/Arrow/Thrift dependencies.
package duckdb

// Note: No-op stub. All functional implementations live in files guarded by
// //go:build duckdb. Importers that need DuckDB should also be behind the
// same build tag.
