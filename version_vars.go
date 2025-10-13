package main

import "github.com/costscope/costscope/cmd"

// Build metadata variables stamped via -ldflags (-X main.Version, etc.).
// They default to placeholders for dev builds.
var (
	Version   = "dev"
	Commit    = "none"
	BuildDate = "unknown"
	GoVersion = "unknown"
)

// Bridge values into cmd package so the version command can read them without
// importing main (which is disallowed).
func init() {
	cmd.Version = Version
	cmd.Commit = Commit
	cmd.BuildDate = BuildDate
	if GoVersion != "unknown" {
		cmd.GoVersion = GoVersion
	}
}
