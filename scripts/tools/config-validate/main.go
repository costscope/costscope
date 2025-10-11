package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	cfg "local/costscope/internal/core/config"

	"gopkg.in/yaml.v3"
)

// Result structures for JSON/text output
type fileResult struct {
	File       string       `json:"file"`
	Valid      bool         `json:"valid"`
	DurationMS int64        `json:"duration_ms"`
	Errors     []errorEntry `json:"errors,omitempty"`
}

type errorEntry struct {
	Section string `json:"section"`
	Key     string `json:"key"`
	Message string `json:"message"`
}

type report struct {
	Files   []fileResult `json:"files"`
	Summary summary      `json:"summary"`
}

type summary struct {
	Total  int `json:"total"`
	Failed int `json:"failed"`
}

const (
	exitOK          = 0
	exitValidation  = 2
	exitInternalErr = 3
)

func main() {
	format := flag.String("format", "json", "output format: json|text")
	root := flag.String("configs", "configs", "configs directory (default 'configs')")
	flag.Parse()

	rep, err := run(*root)
	if err != nil {
		// internal error – write to stderr in minimal form
		fmt.Fprintf(os.Stderr, "internal error: %v\n", err)
		os.Exit(exitInternalErr)
	}

	switch strings.ToLower(*format) {
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(rep)
	case "text":
		for _, f := range rep.Files {
			if f.Valid {
				fmt.Printf("[OK] %s (%dms)\n", f.File, f.DurationMS)
				continue
			}
			fmt.Printf("[FAIL] %s (%dms)\n", f.File, f.DurationMS)
			for _, e := range f.Errors {
				if e.Section != "" {
					fmt.Printf("  - %s.%s: %s\n", e.Section, e.Key, e.Message)
				} else {
					fmt.Printf("  - %s: %s\n", e.Key, e.Message)
				}
			}
		}
		fmt.Printf("Summary: total=%d failed=%d\n", rep.Summary.Total, rep.Summary.Failed)
	default:
		fmt.Fprintf(os.Stderr, "unsupported format: %s\n", *format)
		os.Exit(exitInternalErr)
	}

	if rep.Summary.Failed > 0 {
		os.Exit(exitValidation)
	}
	os.Exit(exitOK)
}

// run performs discovery & validation (extracted for tests)
func run(configDir string) (*report, error) {
	var files []string
	err := filepath.WalkDir(configDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil { // permission or other error
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(d.Name(), ".yaml") && !strings.HasSuffix(d.Name(), ".yml") {
			return nil
		}
		if strings.Contains(d.Name(), ".example") { // skip examples
			return nil
		}
		files = append(files, path)
		return nil
	})
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	sort.Strings(files)

	rep := &report{}
	for _, fpath := range files {
		start := time.Now()
		fr := fileResult{File: fpath, Valid: true}
		content, err := os.ReadFile(fpath) // #nosec G304 (paths are discovered under controlled 'configs' directory)
		if err != nil {
			fr.Valid = false
			fr.Errors = append(fr.Errors, errorEntry{Message: fmt.Sprintf("read error: %v", err)})
			fr.DurationMS = time.Since(start).Milliseconds()
			rep.Files = append(rep.Files, fr)
			continue
		}
		var cfgStruct cfg.ConsolidatedConfig
		if err := yaml.Unmarshal(content, &cfgStruct); err != nil {
			fr.Valid = false
			fr.Errors = append(fr.Errors, errorEntry{Message: fmt.Sprintf("yaml parse error: %v", err)})
			fr.DurationMS = time.Since(start).Milliseconds()
			rep.Files = append(rep.Files, fr)
			continue
		}
		if err := cfg.ValidateAllConfig(&cfgStruct); err != nil {
			fr.Valid = false
			// unwrap chain until first *ConfigError (should be direct) capturing section/key/message
			if ce, ok := err.(*cfg.ConfigError); ok {
				fr.Errors = append(fr.Errors, errorEntry{Section: ce.Section.String(), Key: ce.Key, Message: ce.Message})
			} else {
				fr.Errors = append(fr.Errors, errorEntry{Message: err.Error()})
			}
		}
		fr.DurationMS = time.Since(start).Milliseconds()
		rep.Files = append(rep.Files, fr)
	}

	failed := 0
	for _, f := range rep.Files {
		if !f.Valid {
			failed++
		}
	}
	rep.Summary = summary{Total: len(rep.Files), Failed: failed}
	return rep, nil
}
