package common

import "strings"

// FirstNonEmpty returns the first non-empty (after TrimSpace) string from the list, or "".
// Centralized to avoid small duplicate helpers across provider packages.
func FirstNonEmpty(ss ...string) string {
	for _, s := range ss {
		if strings.TrimSpace(s) != "" {
			return s
		}
	}
	return ""
}
