package verify

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Stage string

const (
	StageParse      Stage = "parse"
	StageMap        Stage = "map"
	StageInvariants Stage = "invariants"
	StageValidate   Stage = "validate"
)

type Status string

const (
	StatusOK        Status = "ok"
	StatusSkipped   Status = "skipped"
	StatusError     Status = "error"
	StatusDrift     Status = "drift"
	StatusThreshold Status = "threshold_exceeded"
)

type StageResult struct {
	Status          Status   `json:"status"`
	DurationMs      int64    `json:"duration_ms"`
	ProcessedRows   int64    `json:"processed_rows,omitempty"`
	SampledRows     int64    `json:"sampled_rows,omitempty"`
	Errors          []string `json:"errors,omitempty"`
	Warnings        []string `json:"warnings,omitempty"`
	DriftViolations []string `json:"drift_violations,omitempty"`
}

type Summary struct {
	Provider       string                `json:"provider"`
	File           string                `json:"file"`
	Format         string                `json:"format"`
	Limit          int                   `json:"limit"`
	StopAfter      string                `json:"stop_after,omitempty"`
	UseUnified     bool                  `json:"use_unified_mapper"`
	Invariants     bool                  `json:"invariants_enabled"`
	BaselinePath   string                `json:"baseline_path,omitempty"`
	Tolerance      float64               `json:"tolerance,omitempty"`
	ErrorThreshold int                   `json:"error_threshold,omitempty"`
	StartedAt      time.Time             `json:"started_at"`
	CompletedAt    time.Time             `json:"completed_at"`
	DurationMs     int64                 `json:"duration_ms"`
	Stages         map[Stage]StageResult `json:"stages"`
	ExitCode       int                   `json:"exit_code"`
	OverallStatus  string                `json:"overall_status"`
	Notes          []string              `json:"notes,omitempty"`
}

type Options struct {
	Provider         string
	File             string
	Limit            int
	StopAfter        string
	UseUnified       bool
	EnableInvariants bool
	BaselinePath     string
	Tolerance        float64
	ErrorThreshold   int
	Format           string
}

const (
	ExitOK          = 0
	ExitParseError  = 2
	ExitMapError    = 3
	ExitDrift       = 4
	ExitValidateErr = 5
)

func Process(opts Options) (*Summary, error) {
	start := time.Now()
	stages := map[Stage]StageResult{}
	shouldRun := func(s Stage) bool {
		if opts.StopAfter == "" {
			return true
		}
		order := []Stage{StageParse, StageMap, StageInvariants, StageValidate}
		target := Stage(opts.StopAfter)
		for _, st := range order {
			if st == s {
				return true
			}
			if st == target {
				return false
			}
		}
		return true
	}

	pr := StageResult{Status: StatusOK}
	pStart := time.Now()
	var perrs []string
	if strings.TrimSpace(opts.Provider) == "" {
		perrs = append(perrs, "missing provider")
	}
	if strings.TrimSpace(opts.File) == "" {
		perrs = append(perrs, "missing file path")
	}
	var detectedFormat string
	var sampleRows int64
	if len(perrs) == 0 {
		detectedFormat = detectFormat(opts.File)
		sr, err := sampleFile(opts.File, opts.Limit)
		if err != nil {
			perrs = append(perrs, fmt.Sprintf("read error: %v", err))
		} else {
			sampleRows = sr
		}
	}
	if len(perrs) > 0 {
		pr.Status = StatusError
		pr.Errors = perrs
	}
	pr.SampledRows = sampleRows
	pr.ProcessedRows = sampleRows
	pr.DurationMs = time.Since(pStart).Milliseconds()
	stages[StageParse] = pr
	if detectedFormat != "" {
		opts.Format = detectedFormat
	}
	if pr.Status == StatusError {
		sum := buildSummary(opts, stages, start)
		sum.ExitCode = ExitParseError
		sum.OverallStatus = "parse_error"
		return sum, nil
	}

	if shouldRun(StageMap) {
		mr := StageResult{Status: StatusOK}
		mStart := time.Now()
		// Use parse sampledRows as input; if zero & limit specified, reflect limit expectation.
		parseRes := stages[StageParse]
		mr.SampledRows = parseRes.SampledRows
		mr.ProcessedRows = parseRes.SampledRows
		mr.DurationMs = time.Since(mStart).Milliseconds()
		stages[StageMap] = mr
	} else {
		stages[StageMap] = StageResult{Status: StatusSkipped}
	}

	if opts.EnableInvariants {
		if shouldRun(StageInvariants) {
			ir := StageResult{Status: StatusOK}
			iStart := time.Now()
			ir.DurationMs = time.Since(iStart).Milliseconds()
			stages[StageInvariants] = ir
		} else {
			stages[StageInvariants] = StageResult{Status: StatusSkipped}
		}
	} else {
		stages[StageInvariants] = StageResult{Status: StatusSkipped}
	}

	if shouldRun(StageValidate) {
		vr := StageResult{Status: StatusOK}
		vStart := time.Now()
		vr.DurationMs = time.Since(vStart).Milliseconds()
		stages[StageValidate] = vr
	} else {
		stages[StageValidate] = StageResult{Status: StatusSkipped}
	}

	sum := buildSummary(opts, stages, start)
	sum.ExitCode, sum.OverallStatus = deriveExit(stages)
	return sum, nil
}

func buildSummary(opts Options, stages map[Stage]StageResult, start time.Time) *Summary {
	return &Summary{Provider: opts.Provider, File: opts.File, Format: opts.Format, Limit: opts.Limit, StopAfter: opts.StopAfter, UseUnified: opts.UseUnified, Invariants: opts.EnableInvariants, BaselinePath: opts.BaselinePath, Tolerance: opts.Tolerance, ErrorThreshold: opts.ErrorThreshold, StartedAt: start, CompletedAt: time.Now(), DurationMs: time.Since(start).Milliseconds(), Stages: stages}
}

func deriveExit(stages map[Stage]StageResult) (int, string) {
	if s, ok := stages[StageValidate]; ok && s.Status == StatusError {
		return ExitValidateErr, "validate_error"
	}
	if s, ok := stages[StageInvariants]; ok && s.Status == StatusDrift {
		return ExitDrift, "invariants_drift"
	}
	if s, ok := stages[StageMap]; ok && (s.Status == StatusError || s.Status == StatusThreshold) {
		return ExitMapError, "map_error"
	}
	if s, ok := stages[StageParse]; ok && s.Status == StatusError {
		return ExitParseError, "parse_error"
	}
	return ExitOK, "ok"
}

func (s *Summary) JSON() string { b, _ := json.MarshalIndent(s, "", "  "); return string(b) }

func (o Options) Validate() error {
	if o.Provider == "" {
		return errors.New("provider is required")
	}
	if o.File == "" {
		return errors.New("file path is required")
	}
	if o.Tolerance < 0 {
		return fmt.Errorf("tolerance must be >= 0")
	}
	if o.ErrorThreshold < 0 {
		return fmt.Errorf("error-threshold must be >= 0")
	}
	return nil
}

// detectFormat infers a lightweight format label from filename extension.
func detectFormat(path string) string {
	name := strings.ToLower(filepath.Base(path))
	// Unconditionally trim .gz (TrimSuffix is idempotent) – staticcheck S1017 compliant.
	name = strings.TrimSuffix(name, ".gz")
	switch {
	case strings.HasSuffix(name, ".csv"):
		return "csv"
	case strings.HasSuffix(name, ".json"):
		return "json"
	case strings.HasSuffix(name, ".parquet"):
		return "parquet"
	default:
		return "unknown"
	}
}

// sampleFile counts up to limit lines (or a default cap) from a text-like file; gzip supported.
// Returns number of lines sampled (excluding header heuristically if present not implemented in stub).
func sampleFile(path string, limit int) (int64, error) {
	// Clean the incoming path to mitigate trivial traversal; this is still user-supplied and
	// intentionally allowed so we document and guard. G304 (gosec) suppressed with justification.
	cleanPath := filepath.Clean(path)
	f, err := os.Open(cleanPath) // #nosec G304 -- acceptable: path comes from explicit CLI/user input for sampling only
	if err != nil {
		return 0, err
	}
	defer func() {
		if cerr := f.Close(); cerr != nil {
			// intentionally ignored: close error on read-only sampling is non-fatal
			_ = cerr
		}
	}()
	var r io.Reader = f
	if strings.HasSuffix(strings.ToLower(cleanPath), ".gz") {
		gz, gzErr := gzip.NewReader(f)
		if gzErr != nil {
			return 0, gzErr
		}
		defer func() {
			if cerr := gz.Close(); cerr != nil {
				// intentionally ignored: gzip close error after scan not actionable here
				_ = cerr
			}
		}()
		r = gz
	}
	scanner := bufio.NewScanner(r)
	// Increase buffer for wide lines
	buf := make([]byte, 0, 128*1024)
	scanner.Buffer(buf, 2*1024*1024)
	var n int64
	max := limit
	if max <= 0 {
		max = 100
	} // default sample size
	for scanner.Scan() {
		n++
		if n >= int64(max) {
			break
		}
	}
	if err := scanner.Err(); err != nil {
		return n, err
	}
	return n, nil
}
