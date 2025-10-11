package cmd

import "runtime"

// Version metadata variables. These are ultimately populated via build-time
// ldflags applied to the main package; a bridge init in package main copies
// values into these variables so the command can access them without importing
// package main (which is disallowed).
var (
	Version   = "dev"
	Commit    = "none"
	BuildDate = "unknown"
	GoVersion = runtime.Version()
)

// VersionInfo DTO for JSON/human output.
type VersionInfo struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildDate string `json:"build_date"`
	GoVersion string `json:"go_version"`
	Dirty     bool   `json:"dirty,omitempty"`
}

func buildVersionInfo() VersionInfo {
	vi := VersionInfo{Version: Version, Commit: Commit, BuildDate: BuildDate, GoVersion: GoVersion}
	if hasDirtySuffix(Version) {
		vi.Dirty = true
	}
	return vi
}

func hasDirtySuffix(v string) bool {
	if len(v) < 6 {
		return false
	}
	return (len(v) >= 6 && (suffixMatch(v, "-dirty") || suffixMatch(v, "+dirty")))
}

func suffixMatch(s, suf string) bool { return len(s) >= len(suf) && s[len(s)-len(suf):] == suf }
