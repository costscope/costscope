package cmd

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

var (
	versionHuman bool
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the CostScope version and build metadata",
	Long:  `Show semantic version, commit, build date (RFC3339 UTC) and Go toolchain. JSON by default; pass --human for readable multi-line output.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		info := buildVersionInfo()
		if !versionHuman { // default JSON (acceptance requires JSON keys)
			b, err := json.Marshal(info)
			if err != nil {
				return err
			}
			fmt.Println(string(b))
			return nil
		}
		human := fmt.Sprintf("CostScope %s\ncommit: %s\nbuild_date: %s\ngo_version: %s", info.Version, info.Commit, info.BuildDate, info.GoVersion)
		if info.BuildDate == "unknown" || info.BuildDate == "" {
			human += "\n(note: build_date not stamped; use make build-release or set SOURCE_DATE_EPOCH)"
		} else if _, err := time.Parse(time.RFC3339, info.BuildDate); err != nil {
			human += "\n(warning: build_date not RFC3339)"
		}
		fmt.Println(human)
		return nil
	},
}

func initVersionCommand() {
	versionCmd.Flags().BoolVar(&versionHuman, "human", false, "human readable output instead of JSON (default JSON)")
}
