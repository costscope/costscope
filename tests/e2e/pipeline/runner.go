//go:build duckdb
// +build duckdb

package pipeline

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/marcboeker/go-duckdb"

	conv "github.com/costscope/costscope/internal/core/focus/conversion"
	"github.com/costscope/costscope/internal/core/focus/quality"
	focustypes "github.com/costscope/costscope/internal/core/focus/types"
)

// Report describes the end-to-end pipeline execution.
type Report struct {
	StartedAt          time.Time                `json:"started_at"`
	EndedAt            time.Time                `json:"ended_at"`
	DurationMs         int64                    `json:"duration_ms"`
	StageDurationsMs   map[string]int64         `json:"stage_durations_ms"`
	InputFiles         []string                 `json:"input_files"`
	OutputParquetFiles []string                 `json:"output_parquet_files"`
	AggregatesSQL      Aggregates               `json:"aggregates_sql"`
	AggregatesScan     Aggregates               `json:"aggregates_scan"`
	RelativeDrift      map[string]float64       `json:"relative_drift"`
	Invariants         quality.InvariantMetrics `json:"invariants"`
	BaselineInvariants quality.InvariantMetrics `json:"baseline_invariants,omitempty"`
	InvariantTolerance float64                  `json:"invariant_tolerance,omitempty"`
	DriftTolerance     float64                  `json:"drift_tolerance"`
	Passed             bool                     `json:"passed"`
	Notes              []string                 `json:"notes,omitempty"`
}

// Aggregates holds sum metrics for drift comparison.
type Aggregates struct {
	RowCount         int64   `json:"row_count"`
	SumEffectiveCost float64 `json:"sum_effective_cost"`
	SumListCost      float64 `json:"sum_list_cost"`
	SumUsageQuantity float64 `json:"sum_usage_quantity"`
}

// RunConfig configures the pipeline run.
type RunConfig struct {
	Provider           string
	InputFiles         []string
	WorkDir            string
	DriftTolerance     float64 // relative tolerance, default 0.001 (0.1%)
	ValidateOutput     bool
	BaselinePath       string  // optional path to baseline invariants JSON
	InvariantTolerance float64 // relative tolerance for invariants compare (default 0.01 = 1%)
}

// Run executes the ingestion->conversion->validation->export pipeline.
func Run(ctx context.Context, cfg RunConfig) (*Report, error) {
	if len(cfg.InputFiles) == 0 {
		return nil, errors.New("no input files provided")
	}
	if cfg.WorkDir == "" {
		wd, err := os.MkdirTemp("", "pipeline-e2e-*")
		if err != nil {
			return nil, err
		}
		cfg.WorkDir = wd
	}
	if cfg.DriftTolerance <= 0 {
		cfg.DriftTolerance = 0.001
	}

	if cfg.InvariantTolerance <= 0 {
		cfg.InvariantTolerance = 0.01
	}
	rep := &Report{StartedAt: time.Now().UTC(), StageDurationsMs: map[string]int64{}, InputFiles: append([]string{}, cfg.InputFiles...), DriftTolerance: cfg.DriftTolerance, RelativeDrift: map[string]float64{}, InvariantTolerance: cfg.InvariantTolerance}

	stage := func(name string, fn func() error) error {
		start := time.Now()
		err := fn()
		rep.StageDurationsMs[name] = time.Since(start).Milliseconds()
		return err
	}

	var convResults []*focustypes.ConversionResult
	uc := conv.NewConversionManager(4).GetConverter()

	// Conversion stage (ingestion included)
	if err := stage("conversion", func() error {
		for _, in := range cfg.InputFiles {
			out := filepath.Join(cfg.WorkDir, filepath.Base(strings.TrimSuffix(in, filepath.Ext(in)))+".parquet")
			c := &focustypes.ConversionConfig{Provider: cfg.Provider, InputPath: in, OutputPath: out, OutputFormat: "parquet", Streaming: true, ValidateInput: false, ValidateOutput: cfg.ValidateOutput, Parquet: focustypes.ParquetOptions{CompressionCodec: "snappy", RotateSizeBytes: -1}}
			res, err := uc.Convert(ctx, c)
			if err != nil {
				return fmt.Errorf("convert %s: %w", in, err)
			}
			// Some legacy non-streaming paths may omit OutputFile; ensure we set & verify.
			if res.OutputFile == "" {
				res.OutputFile = out
			}
			if _, statErr := os.Stat(res.OutputFile); statErr != nil {
				return fmt.Errorf("expected parquet output not found: %s (hint: streaming path writes files; ensure converter supports streaming)", res.OutputFile)
			}
			convResults = append(convResults, res)
			rep.OutputParquetFiles = append(rep.OutputParquetFiles, res.OutputFile)
		}
		return nil
	}); err != nil {
		rep.Notes = append(rep.Notes, err.Error())
		finalize(rep)
		return rep, err
	}

	// Validation stage intentionally skipped (converter may have already validated output)
	if cfg.ValidateOutput {
		_ = stage("validation", func() error { return nil })
	}

	// Aggregates via DuckDB SQL
	var agSQL Aggregates
	if err := stage("aggregate_sql", func() error {
		db, err := sql.Open("duckdb", ":memory:")
		if err != nil {
			return err
		}
		defer func() { _ = db.Close() }()
		placeholders := make([]string, 0, len(rep.OutputParquetFiles))
		args := make([]interface{}, 0, len(rep.OutputParquetFiles))
		for _, f := range rep.OutputParquetFiles {
			placeholders = append(placeholders, "SELECT * FROM read_parquet(?)")
			args = append(args, f)
		}
		union := strings.Join(placeholders, " UNION ALL ")
		// The union string is constructed from static fragments; still avoid fmt string concatenation for the full query.
		// The union clause is built exclusively from static fragments "SELECT * FROM read_parquet(?)" joined by " UNION ALL " and
		// parameters are still bound via QueryRow arguments; no user input influences SQL structure. #nosec G202
		baseQ := "SELECT COUNT(*) as rc, SUM(effective_cost), SUM(list_cost), SUM(usage_quantity) FROM (" + union + ")" //nolint:gosec
		row := db.QueryRow(baseQ, args...)
		if err := row.Scan(&agSQL.RowCount, &agSQL.SumEffectiveCost, &agSQL.SumListCost, &agSQL.SumUsageQuantity); err != nil {
			return err
		}
		rep.AggregatesSQL = agSQL
		return nil
	}); err != nil {
		rep.Notes = append(rep.Notes, err.Error())
		finalize(rep)
		return rep, err
	}

	// Scan records path & invariants
	var agScan Aggregates
	var allRecords []focustypes.FocusRecord
	if err := stage("scan_records", func() error {
		db, err := sql.Open("duckdb", ":memory:")
		if err != nil {
			return err
		}
		defer func() { _ = db.Close() }()
		for _, f := range rep.OutputParquetFiles {
			rows, err := db.Query("SELECT effective_cost, list_cost, usage_quantity, charge_category, pricing_category, provider_name, resource_id FROM read_parquet(?)", f)
			if err != nil {
				return err
			}
			for rows.Next() {
				var ec, lc, uq float64
				var cc, pc, pn, rid string
				if err := rows.Scan(&ec, &lc, &uq, &cc, &pc, &pn, &rid); err != nil {
					return err
				}
				allRecords = append(allRecords, focustypes.FocusRecord{EffectiveCost: ec, ListCost: lc, UsageQuantity: uq, ChargeCategory: cc, PricingCategory: pc, ProviderName: pn, ResourceId: rid})
				agScan.RowCount++
				agScan.SumEffectiveCost += ec
				agScan.SumListCost += lc
				agScan.SumUsageQuantity += uq
			}
			if err := rows.Err(); err != nil {
				return err
			}
			_ = rows.Close()
		}
		rep.AggregatesScan = agScan
		rep.Invariants = quality.ComputeInvariants(allRecords)
		// Optional baseline comparison for regression detection
		if cfg.BaselinePath != "" {
			base, berr := quality.LoadBaseline(cfg.BaselinePath)
			if berr != nil {
				rep.Notes = append(rep.Notes, fmt.Sprintf("baseline_load_error:%v", berr))
			} else {
				rep.BaselineInvariants = base
				quality.CompareInvariants(&rep.Invariants, base, cfg.InvariantTolerance)
			}
		}
		return nil
	}); err != nil {
		rep.Notes = append(rep.Notes, err.Error())
		finalize(rep)
		return rep, err
	}

	// Drift comparison
	rep.RelativeDrift["sum_effective_cost"] = relDiff(agScan.SumEffectiveCost, agSQL.SumEffectiveCost)
	rep.RelativeDrift["sum_list_cost"] = relDiff(agScan.SumListCost, agSQL.SumListCost)
	rep.RelativeDrift["sum_usage_quantity"] = relDiff(agScan.SumUsageQuantity, agSQL.SumUsageQuantity)
	rep.RelativeDrift["row_count"] = relDiff(float64(agScan.RowCount), float64(agSQL.RowCount))
	pass := true
	for k, v := range rep.RelativeDrift {
		if v > cfg.DriftTolerance {
			pass = false
			rep.Notes = append(rep.Notes, fmt.Sprintf("drift %s=%.6f > tol %.6f", k, v, cfg.DriftTolerance))
		}
	}
	if len(rep.Invariants.Violations) > 0 {
		pass = false
	}
	rep.Passed = pass

	finalize(rep)
	return rep, nil
}

func relDiff(a, b float64) float64 {
	if b == 0 {
		if a == 0 {
			return 0
		}
		return 1
	}
	if a == b {
		return 0
	}
	if a < 0 && b < 0 {
		a, b = -a, -b
	}
	d := a - b
	if d < 0 {
		d = -d
	}
	if b < 0 {
		b = -b
	}
	return d / b
}

// Save writes report JSON.
func (r *Report) Save(path string) error {
	if path == "" {
		return errors.New("empty path")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return err
	}
	b, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o600)
}

func finalize(r *Report) {
	r.EndedAt = time.Now().UTC()
	r.DurationMs = r.EndedAt.Sub(r.StartedAt).Milliseconds()
}
