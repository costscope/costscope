package helpers

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"local/costscope/internal/core/logging"
)

// RunEWithLogging wraps a cobra RunE with standardized start/finish logging and duration.
// It preserves the original signature and return semantics.
func RunEWithLogging(name string, fn func(cmd *cobra.Command, args []string) error) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		l := logging.GetLogger().WithFields(map[string]interface{}{
			"cmd": name,
		})
		start := time.Now()
		l.Info("command start")
		err := fn(cmd, args)
		duration := time.Since(start)
		if err != nil {
			l.ErrorWithFields("command error", map[string]interface{}{
				"cmd":      name,
				"duration": duration.String(),
				"error":    err.Error(),
			})
			return err
		}
		l.InfoWithFields("command success", map[string]interface{}{
			"cmd":      name,
			"duration": duration.String(),
		})
		return nil
	}
}

// GenerateID returns a short, random, hex-encoded identifier with the given prefix.
// Example: prefix "conn" -> "conn_1a2b3c4d5e6f7a8b".
func GenerateID(prefix string) string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%s_%s", prefix, hex.EncodeToString(b))
}

// PrintHeader prints a simple section header line.
func PrintHeader(title string) {
	fmt.Println(title)
}

// PrintTip prints a small tip line with a lightbulb marker.
func PrintTip(tip string) {
	fmt.Printf(" %s\n", tip)
}

// PrintKV prints a key/value line with basic formatting.
func PrintKV(key string, value interface{}) {
	fmt.Printf("   %s: %v\n", key, value)
}

// AddVerboseFlag adds a --verbose boolean flag bound to the provided pointer.
// NOTE: previously had AddVerboseFlag / AddStatusFilterFlag convenience helpers.
// They were never adopted outside this package and duplicate the declarative
// CLI flag registration approach used elsewhere. Removed as dead code to reduce
// surface area; commands should register flags directly or via the shared
// action spec registrar. If reintroduced, ensure they add unique value beyond
// thin wrappers around cobra flag methods.
