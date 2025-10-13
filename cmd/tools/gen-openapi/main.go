package main

// Go-based OpenAPI spec generator:
//  * Replaces placeholders {{BASE_URL}} and {{WS_BASE_URL}} using helpers in internal/core/docs
//  * Optionally injects integration path fragment (integration.paths.fragment.yaml) identical to prior shell script
//  * Always generates standard spec; generates enterprise spec when template exists
//  * Single source of truth (removes duplicated ws base URL derivation logic from shell script)
//
// Env:
//   DOCS_BASE_URL (optional) – consumed by docs.GetBaseURL / GetWSBaseURL
//   OUTPUT_DIR (optional) – defaults to internal/api/docs
//
// Exit codes: 0 success; non-zero on any IO or template error.

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	docsb "github.com/costscope/costscope/internal/core/docs"
	"github.com/costscope/costscope/internal/testutil"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "openapi generation failed: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	repoRoot, err := testutil.RepoRoot()
	if err != nil {
		return err
	}

	outDir := os.Getenv("OUTPUT_DIR")
	if outDir == "" {
		outDir = filepath.Join(repoRoot, "internal", "api", "docs")
	}
	// Directory creation with restrictive permissions (<= 0750 per security guidance)
	if err := os.MkdirAll(outDir, 0o750); err != nil {
		return fmt.Errorf("create out dir: %w", err)
	}

	base := docsb.GetBaseURL()
	ws := docsb.GetWSBaseURL()

	// Standard spec
	stdTpl := filepath.Join(outDir, "openapi.template.yaml")
	stdOut := filepath.Join(outDir, "openapi.yaml")
	if err := processTemplate(stdTpl, stdOut, base, ws, repoRoot); err != nil {
		return fmt.Errorf("standard spec: %w", err)
	}

	// Enterprise spec (generate only if template exists)
	entTpl := filepath.Join(outDir, "enterprise-openapi.template.yaml")
	if fileExists(entTpl) {
		entOut := filepath.Join(outDir, "enterprise-openapi.yaml")
		if err := processTemplate(entTpl, entOut, base, ws, repoRoot); err != nil {
			return fmt.Errorf("enterprise spec: %w", err)
		}
	}

	if _, err := fmt.Fprintf(os.Stdout, "Rendered OpenAPI specs (BASE_URL=%s WS_BASE_URL=%s)\n", base, ws); err != nil {
		return fmt.Errorf("print success message: %w", err)
	}
	return nil
}

func processTemplate(input, output, base, ws, repoRoot string) error {
	// Validate template path stays within docs directory to avoid unintended inclusion (#gosec G304)
	docsDir := filepath.Join(repoRoot, "internal", "api", "docs") + string(os.PathSeparator)
	cleanInput := filepath.Clean(input) + string(os.PathSeparator) // add sep to avoid prefix false positives where names share prefix
	if !strings.HasPrefix(cleanInput, docsDir) {
		return fmt.Errorf("template path outside docs directory: %s", input)
	}
	b, err := os.ReadFile(input) // #nosec G304 - path validated above
	if err != nil {
		return fmt.Errorf("read template %s: %w", input, err)
	}
	rendered := bytes.ReplaceAll(b, []byte("{{BASE_URL}}"), []byte(base))
	rendered = bytes.ReplaceAll(rendered, []byte("{{WS_BASE_URL}}"), []byte(ws))

	// Inject integration fragment if present
	fragPath := filepath.Join(repoRoot, "internal", "api", "docs", "integration.paths.fragment.yaml")
	if fileExists(fragPath) {
		// #nosec G304 - controlled, fixed path in repo
		rf, err := os.Open(fragPath)
		if err != nil {
			return fmt.Errorf("open fragment: %w", err)
		}
		fragContent, err := stripFirstLineAndIndent(rf, 2)
		_ = rf.Close()
		if err != nil {
			return fmt.Errorf("process fragment: %w", err)
		}
		// Insert before first 'components:' occurrence
		lines := strings.Split(string(rendered), "\n")
		inserted := false
		var outLines []string
		for _, l := range lines {
			if !inserted && strings.HasPrefix(l, "components:") {
				outLines = append(outLines, fragContent...)
				inserted = true
			}
			outLines = append(outLines, l)
		}
		rendered = []byte(strings.Join(outLines, "\n"))
	}

	// Safety: ensure no unreplaced placeholders remain
	if bytes.Contains(rendered, []byte("{{BASE_URL}}")) || bytes.Contains(rendered, []byte("{{WS_BASE_URL}}")) {
		return errors.New("unreplaced placeholders remain after processing")
	}
	// Restrictive file permissions (<= 0600) for generated specs (readable by owner only)
	if err := os.WriteFile(output, rendered, 0o600); err != nil { // #gosec G306
		return fmt.Errorf("write %s: %w", output, err)
	}
	return nil
}

func stripFirstLineAndIndent(r io.Reader, spaces int) ([]string, error) {
	scanner := bufio.NewScanner(r)
	var lines []string
	first := true
	indent := strings.Repeat(" ", spaces)
	for scanner.Scan() {
		if first { // skip the first line (e.g., 'paths:')
			first = false
			continue
		}
		lines = append(lines, indent+scanner.Text())
	}
	return lines, scanner.Err()
}

// local findRepoRoot removed in favor of testutil.RepoRoot

func fileExists(p string) bool {
	info, err := os.Stat(p)
	return err == nil && !info.IsDir()
}
