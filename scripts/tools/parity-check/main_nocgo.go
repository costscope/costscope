//go:build !cgo
// +build !cgo

package main

import (
	"encoding/json"
	"fmt"
	"os"
)

// minimal struct matching output fields so downstream consumers don't crash if invoked
type parityResult struct {
    Legacy          any     `json:"legacy,omitempty"`
    Unified         any     `json:"unified,omitempty"`
    EqualCost       bool    `json:"equal_cost"`
    EqualUsage      bool    `json:"equal_usage"`
    EqualRecords    bool    `json:"equal_records"`
    LiteHashLegacy  string  `json:"lite_hash_legacy,omitempty"`
    LiteHashUnified string  `json:"lite_hash_unified,omitempty"`
    EqualLiteHash   bool    `json:"equal_lite_hash"`
    DurationMs      int64   `json:"duration_ms"`
    Tolerance       float64 `json:"tolerance"`
    Timestamp       string  `json:"timestamp"`
    Skipped         string  `json:"skipped_reason"`
}

func main() {
    // Provide a deterministic, explicit message to indicate parity-check is skipped without cgo.
    res := parityResult{Skipped: "parity-check requires cgo-enabled DuckDB; build ran without cgo"}
    enc := json.NewEncoder(os.Stdout)
    enc.SetIndent("", "  ")
    _ = enc.Encode(res)
    // Exit 0 to avoid failing the build in environments without cgo.
    fmt.Fprintln(os.Stderr, "[parity-check] skipped: CGO disabled (no DuckDB)")
}
