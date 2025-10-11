package paths

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

// BaseOutputDir is the mandatory root directory for all Parquet (and other) user output files.
// All user supplied output paths are rewritten / validated to reside under this directory.
const BaseOutputDir = "costscope-data"

// ErrOutsideDataDir is returned when a user supplied output path resolves outside BaseOutputDir.
var ErrOutsideDataDir = errors.New("output path outside costscope-data")

// SanitizeOutput takes a user provided path (relative or absolute) and returns an absolute path
// guaranteed to reside within the repository root's costscope-data directory. The rules:
//   - Relative paths not starting with costscope-data/ are placed inside costscope-data/ (filename preserved)
//   - Relative paths already starting with costscope-data/ are kept (still validated)
//   - Absolute paths must already be within costscope-data/ (validation only)
//   - Any traversal attempts (.. segments that would escape) or absolute paths outside the base are rejected
//
// The function does not create any directories; it purely validates & resolves. A sentinel
// ErrOutsideDataDir is wrapped for policy violations so callers can test errors.Is / substring.
func SanitizeOutput(userPath string) (string, error) {
	if strings.TrimSpace(userPath) == "" {
		return "", fmt.Errorf("empty output path")
	}

	baseAbs, err := filepath.Abs(BaseOutputDir)
	if err != nil {
		return "", fmt.Errorf("resolve base dir: %w", err)
	}

	// Helper to verify final path is inside base
	ensureInside := func(final string) (string, error) {
		cleanFinal := filepath.Clean(final)
		rel, relErr := filepath.Rel(baseAbs, cleanFinal)
		if relErr != nil {
			return "", fmt.Errorf("rel calc: %w", relErr)
		}
		if rel == "." { // base itself is allowed but not very useful
			return cleanFinal, nil
		}
		if strings.HasPrefix(rel, "..") { // escaped
			return "", fmt.Errorf("%w: %s", ErrOutsideDataDir, cleanFinal)
		}
		return cleanFinal, nil
	}

	if filepath.IsAbs(userPath) {
		// Absolute path must already be under costscope-data
		return ensureInside(userPath)
	}

	cleaned := filepath.Clean(userPath)
	// Reject explicit traversal attempts up front
	if cleaned == ".." || strings.HasPrefix(cleaned, "../") || strings.Contains(cleaned, "/../") {
		return "", fmt.Errorf("%w: traversal not permitted (%s)", ErrOutsideDataDir, userPath)
	}

	// If user already prefixed costscope-data we keep it (still validated)
	if cleaned == BaseOutputDir || strings.HasPrefix(cleaned, BaseOutputDir+string(filepath.Separator)) {
		abs := filepath.Join(baseAbs, strings.TrimPrefix(cleaned[len(BaseOutputDir):], string(filepath.Separator)))
		return ensureInside(abs)
	}

	// Otherwise, place inside base
	final := filepath.Join(baseAbs, cleaned)
	return ensureInside(final)
}
