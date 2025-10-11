package docs

import (
	"os"
	"strings"
)

const defaultBaseURL = "http://localhost:8080"

// GetBaseURL returns the base URL used for documentation/examples.
// It reads DOCS_BASE_URL env var (if set) otherwise falls back to defaultBaseURL.
// Trailing slashes are trimmed.
func GetBaseURL() string {
	v := strings.TrimSpace(os.Getenv("DOCS_BASE_URL"))
	if v == "" {
		return defaultBaseURL
	}
	for strings.HasSuffix(v, "/") && len(v) > 1 { // trim trailing slashes
		v = strings.TrimSuffix(v, "/")
	}
	return v
}

// GetWSBaseURL derives the websocket base URL (ws:// or wss://) from the HTTP(S) base.
func GetWSBaseURL() string {
	base := GetBaseURL()
	if strings.HasPrefix(base, "https://") {
		return "wss://" + strings.TrimPrefix(base, "https://")
	}
	if strings.HasPrefix(base, "http://") {
		return "ws://" + strings.TrimPrefix(base, "http://")
	}
	// If scheme missing, assume http and map to ws
	return "ws://" + strings.TrimPrefix(base, "//")
}
