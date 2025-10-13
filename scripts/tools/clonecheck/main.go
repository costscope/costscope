package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/costscope/costscope/internal/testutil"
)

type snippet struct {
	file      string
	startLine int
	endLine   int
	text      string
}

type group struct {
	count   int
	snips   []snippet
	fullDup bool
}

var (
	groupHeader = regexp.MustCompile(`^found (\d+) clones:`)
	locRe       = regexp.MustCompile(`^\s+([^:]+):(\d+),(\d+)`)
)

// safeJoin joins base and rel and ensures the result remains within the base directory.
// It prevents path traversal outside of the intended workspace root.
func safeJoin(base, rel string) (string, error) {
	baseAbs, err := filepath.Abs(base)
	if err != nil {
		return "", err
	}
	// Clean relative path and join
	candidate := filepath.Join(baseAbs, filepath.Clean(rel))
	candAbs, err := filepath.Abs(candidate)
	if err != nil {
		return "", err
	}
	// Ensure candAbs is within baseAbs
	if candAbs == baseAbs {
		return candAbs, nil
	}
	// Add path separator to avoid prefix tricks (e.g., /baseX)
	baseWithSep := baseAbs + string(os.PathSeparator)
	if strings.HasPrefix(candAbs, baseWithSep) {
		return candAbs, nil
	}
	return "", fmt.Errorf("unsafe path outside base: %s", rel)
}

func readLines(path string, start, end int) (string, error) {
	// #nosec G304 -- Path validated via safeJoin in caller; not user-controlled beyond that.
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()
	s := bufio.NewScanner(f)
	var b strings.Builder
	line := 0
	for s.Scan() {
		line++
		if line < start {
			continue
		}
		if line > end {
			break
		}
		b.WriteString(s.Text())
		b.WriteByte('\n')
	}
	return b.String(), s.Err()
}

func main() {
	report := "logs/report_duplicates.txt"
	if len(os.Args) > 1 {
		report = os.Args[1]
	}
	r, err := os.ReadFile(report)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read report: %v\n", err)
		os.Exit(1)
	}
	lines := strings.Split(string(r), "\n")
	cwd, err := testutil.RepoRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to locate repo root: %v\n", err)
		os.Exit(1)
	}
	var groups []group
	for i := 0; i < len(lines); i++ {
		m := groupHeader.FindStringSubmatch(lines[i])
		if m == nil {
			continue
		}
		cnt, _ := strconv.Atoi(m[1])
		g := group{count: cnt}
		for j := 0; j < cnt && i+1 < len(lines); j++ {
			i++
			lm := locRe.FindStringSubmatch(lines[i])
			if lm == nil {
				continue
			}
			file := filepath.Clean(lm[1])
			start, _ := strconv.Atoi(lm[2])
			end, _ := strconv.Atoi(lm[3])
			// Validate and resolve to an absolute path under cwd
			safePath, jerr := safeJoin(cwd, file)
			if jerr != nil {
				g.snips = append(g.snips, snippet{file, start, end, fmt.Sprintf("<ERR %v>", jerr)})
				continue
			}
			text, err := readLines(safePath, start, end)
			if err != nil {
				text = fmt.Sprintf("<ERR %v>", err)
			}
			g.snips = append(g.snips, snippet{file, start, end, text})
		}
		// determine if snippets are exactly identical (full clone)
		g.fullDup = true
		if len(g.snips) > 0 {
			base := g.snips[0].text
			for k := 1; k < len(g.snips); k++ {
				if g.snips[k].text != base {
					g.fullDup = false
					break
				}
			}
		}
		groups = append(groups, g)
	}

	// Print report in a compact machine-friendly way
	for gi, g := range groups {
		status := "partial"
		if g.fullDup {
			status = "full"
		}
		fmt.Printf("group %d: %s (%d clones)\n", gi+1, status, g.count)
		for _, s := range g.snips {
			fmt.Printf("  %s:%d,%d\n", s.file, s.startLine, s.endLine)
		}
	}
}
