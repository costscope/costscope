package specs

// Package specs provides small helpers shared across CLI spec generation and tests.

import (
	"crypto/sha256"
	"encoding/hex"
)

// ComputeChecksumHex returns the lowercase hex-encoded SHA-256 of the provided data.
// It is used by the integration CLI docs generator and related tests to ensure
// consistent checksum calculation across tools and packages.
func ComputeChecksumHex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
