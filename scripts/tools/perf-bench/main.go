// Unified benchmark tool (perf build tag removed). Always available; legacy invocations
// using `-tags perf` continue to work but are unnecessary.

package main

// Performance benchmark & regression guard for legacy vs unified mapper.
// TASK-PERF-BENCH implementation.
// Runs conversion scenarios (chunk sizes 10k/50k/100k) with rotation on/off
// for both legacy (optimized) and unified mapper paths, then compares ratios.
// Fails (exit 1) if unified exceeds duration or allocation ratio thresholds.
// Output: bench_results.json (default) containing full scenario metrics.

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"sort"
	"strconv"
	"time"

	awsc "local/costscope/internal/core/focus/conversion/aws"
	azc "local/costscope/internal/core/focus/conversion/azure"
	gcpp "local/costscope/internal/core/focus/conversion/gcp"
	ftypes "local/costscope/internal/core/focus/types"
	"local/costscope/internal/database/performance"
)

type runMetrics struct {
	DurationMS  int64   `json:"duration_ms"`
	Records     int64   `json:"records"`
	RPS         float64 `json:"rps"`
	AllocBytes  uint64  `json:"alloc_bytes"`
	AllocMB     float64 `json:"alloc_mb"`
	OutputFiles int     `json:"output_files"`
}

type scenarioResult struct {
	ChunkSize  int              `json:"chunk_size"`
	Rotation   string           `json:"rotation"` // on|off
	Legacy     runMetrics       `json:"legacy"`
	Unified    runMetrics       `json:"unified"`
	Ratios     ratioResult      `json:"ratios"`
	Thresholds thresholds       `json:"thresholds"`
	Baseline   *baselineMetrics `json:"baseline,omitempty"`
	Status     string           `json:"status"` // pass|fail
	Notes      []string         `json:"notes,omitempty"`
}

// Status constants to satisfy goconst and avoid magic strings.
const (
	statusPass = "pass"
	statusFail = "fail"
)

type baselineMetrics struct {
	LegacyDurationMS        int64   `json:"legacy_duration_ms"`
	UnifiedDurationMS       int64   `json:"unified_duration_ms"`
	LegacyAllocBytes        uint64  `json:"legacy_alloc_bytes"`
	UnifiedAllocBytes       uint64  `json:"unified_alloc_bytes"`
	LegacyDurationDeltaPct  float64 `json:"legacy_duration_delta_pct"`
	UnifiedDurationDeltaPct float64 `json:"unified_duration_delta_pct"`
	LegacyAllocDeltaPct     float64 `json:"legacy_alloc_delta_pct"`
	UnifiedAllocDeltaPct    float64 `json:"unified_alloc_delta_pct"`
}

type ratioResult struct {
	Duration float64 `json:"duration"`
	Alloc    float64 `json:"alloc"`
}

type thresholds struct {
	DurationMax float64 `json:"duration_max"`
	AllocMax    float64 `json:"alloc_max"`
}

type benchReport struct {
	GeneratedAt   time.Time        `json:"generated_at"`
	Provider      string           `json:"provider"`
	InputFile     string           `json:"input_file"`
	DurationMax   float64          `json:"duration_ratio_max"`
	AllocMax      float64          `json:"alloc_ratio_max"`
	Scenarios     []scenarioResult `json:"scenarios"`
	OverallStatus string           `json:"overall_status"`
	Version       string           `json:"version"`
	BaselineFile  string           `json:"baseline_file,omitempty"`
	PerfEngine    *perfEngineBench `json:"perf_engine,omitempty"`
}

// perfEngineBench holds optional micro-benchmarks for the database performance layer (TASK: integrate unused layer)
type perfEngineBench struct {
	Memory struct {
		DurationMS    int64   `json:"duration_ms"`
		Operations    int64   `json:"operations"`
		OpsPerSec     float64 `json:"ops_per_sec"`
		MemoryDeltaMB int64   `json:"memory_delta_mb"`
	} `json:"memory"`
	Cache struct {
		DurationMS int64   `json:"duration_ms"`
		Operations int64   `json:"operations"`
		OpsPerSec  float64 `json:"ops_per_sec"`
		HitRate    float64 `json:"hit_rate"`
	} `json:"cache"`
	Parallel struct {
		DurationMS int64   `json:"duration_ms"`
		Jobs       int64   `json:"jobs"`
		JobsPerSec float64 `json:"jobs_per_sec"`
	} `json:"parallel"`
	Notes []string `json:"notes,omitempty"`
}

func main() {
	var (
		input        = flag.String("input", "demo/focus-conversion/demo-cur-data.csv", "Input billing CSV/CSV.gz path (AWS sample)")
		provider     = flag.String("provider", "aws", "Provider (aws|azure|gcp) - current benchmark optimized for aws sample")
		output       = flag.String("output", "bench_results.json", "Benchmark results JSON output path")
		durationMax  = flag.Float64("duration-max", 1.15, "Max allowed unified/legacy duration ratio (env PERF_BENCH_DURATION_MAX overrides)")
		allocMax     = flag.Float64("alloc-max", 1.20, "Max allowed unified/legacy allocation ratio (env PERF_BENCH_ALLOC_MAX overrides)")
		iterations   = flag.Int("iterations", 1, "Iterations per scenario (median taken if >1)")
		rotateSize   = flag.Int64("rotate-size-bytes", 2048, "Rotation size bytes when rotation is 'on' (small to force rotation)")
		workers      = flag.Int("workers", 1, "Workers to use for streaming conversion")
		baselinePath = flag.String("baseline", "", "Optional baseline bench_results.json to compare raw metric deltas")
		promOut      = flag.String("prom-output", "", "Optional Prometheus metrics output file (text format)")
		perfInclude  = flag.Bool("include-perf-engine", false, "Include internal performance engine micro-benchmarks (memory/cache/parallel)")
		perfOps      = flag.Int("perf-ops", 5000, "Operations for memory/cache benchmarks (parallel uses perf-ops/10 jobs, min 100)")
		// Noise guards for tiny demo inputs: when absolute durations are very small,
		// ratio spikes are often measurement noise. If both unified and legacy are
		// below this floor (in ms), treat the scenario as pass. Also allow a small
		// absolute delta (in ms) to be ignored.
		noiseFloorMS = flag.Int64("duration-noise-floor-ms", 20, "If both durations < this (ms), ignore ratio and pass scenario")
		deltaFloorMS = flag.Int64("duration-delta-floor-ms", 10, "If abs(unified-legacy) <= this (ms) and both under noise floor, pass")
	)
	flag.Parse()

	if *iterations < 1 {
		fmt.Fprintf(os.Stderr, "iterations must be >=1\n")
		os.Exit(2)
	}

	// Environment variable overrides (matrix-friendly)
	applyEnvOverrides(durationMax, allocMax, noiseFloorMS, deltaFloorMS)

	// Ensure reproducibility baseline: reduce GC noise & disable GC tuning changes.
	debug.SetGCPercent(100)

	chunkSizes := []int{10000, 50000, 100000}
	rotations := []string{"off", "on"}

	report := benchReport{
		GeneratedAt: time.Now().UTC(),
		Provider:    *provider,
		InputFile:   *input,
		DurationMax: *durationMax,
		AllocMax:    *allocMax,
		Version:     "1",
	}

	var baseline map[string]scenarioResult
	if *baselinePath != "" {
		if b, err := loadBaseline(*baselinePath); err == nil {
			baseline = b
			report.BaselineFile = *baselinePath
		} else {
			fmt.Fprintf(os.Stderr, "baseline load failed: %v\n", err)
		}
	}

	for _, cs := range chunkSizes {
		for _, rot := range rotations {
			scen := runScenarioBatch(*provider, *input, cs, rot, *iterations, *rotateSize, *workers, *durationMax, *allocMax, *noiseFloorMS, *deltaFloorMS, baseline)
			report.Scenarios = append(report.Scenarios, scen)
		}
	}

	// Optional performance engine micro benchmarks (additive, does not affect pass/fail)
	if *perfInclude {
		peRes, err := runPerformanceEngineBenches(*perfOps)
		if err != nil {
			// Non-fatal: note error
			peRes = &perfEngineBench{Notes: []string{fmt.Sprintf("perf engine bench error: %v", err)}}
		}
		report.PerfEngine = peRes
	}

	// overall status
	report.OverallStatus = statusPass
	for _, s := range report.Scenarios {
		if s.Status != statusPass {
			report.OverallStatus = statusFail
			break
		}
	}

	if err := writeJSON(*output, report); err != nil {
		fmt.Fprintf(os.Stderr, "write results: %v\n", err)
		os.Exit(1)
	}

	_ = maybeWriteProm(*promOut, &report)

	if report.OverallStatus != statusPass {
		fmt.Fprintf(os.Stderr, "Regression detected (see %s)\n", *output)
		os.Exit(1)
	}
	fmt.Printf("Benchmark complete. Results: %s (status=%s)\n", *output, report.OverallStatus)
}

// runPerformanceEngineBenches executes lightweight micro benchmarks of the internal performance layer
// to keep the code exercised and provide visibility when integrating or deciding deprecation.
func runPerformanceEngineBenches(ops int) (*perfEngineBench, error) {
	if ops < 100 { // enforce a lower bound to reduce timing noise
		ops = 100
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	eng := performance.NewPerformanceEngine(performance.DefaultPerformanceConfig())
	if err := eng.Start(ctx); err != nil {
		return nil, fmt.Errorf("start performance engine: %w", err)
	}
	defer func() { _ = eng.Stop() }()

	bench := performance.NewPerformanceBenchmark(eng)

	memBr := bench.RunMemoryBenchmark(ctx, ops)
	cacheBr := bench.RunCacheBenchmark(ops)
	parallelJobs := ops / 10
	if parallelJobs < 100 { // ensure reasonable job count
		parallelJobs = 100
	}
	parBr := bench.RunParallelBenchmark(ctx, parallelJobs)

	res := &perfEngineBench{}
	res.Memory.DurationMS = memBr.Duration.Milliseconds()
	res.Memory.Operations = memBr.OperationsCount
	res.Memory.OpsPerSec = memBr.OperationsPerSec
	res.Memory.MemoryDeltaMB = memBr.MemoryUsedMB

	res.Cache.DurationMS = cacheBr.Duration.Milliseconds()
	res.Cache.Operations = cacheBr.OperationsCount
	res.Cache.OpsPerSec = cacheBr.OperationsPerSec
	res.Cache.HitRate = cacheBr.CacheHitRate

	res.Parallel.DurationMS = parBr.Duration.Milliseconds()
	res.Parallel.Jobs = parBr.OperationsCount
	res.Parallel.JobsPerSec = parBr.OperationsPerSec

	return res, nil
}

// applyEnvOverrides applies environment overrides for thresholds and noise-floor knobs.
func applyEnvOverrides(durationMax, allocMax *float64, noiseFloorMS, deltaFloorMS *int64) {
	if v := os.Getenv("PERF_BENCH_DURATION_MAX"); v != "" {
		if f, err := parseFloatStrict(v); err == nil {
			*durationMax = f
		} else {
			fmt.Fprintf(os.Stderr, "WARN: invalid PERF_BENCH_DURATION_MAX=%s (%v)\n", v, err)
		}
	}
	if v := os.Getenv("PERF_BENCH_ALLOC_MAX"); v != "" {
		if f, err := parseFloatStrict(v); err == nil {
			*allocMax = f
		} else {
			fmt.Fprintf(os.Stderr, "WARN: invalid PERF_BENCH_ALLOC_MAX=%s (%v)\n", v, err)
		}
	}
	if v := os.Getenv("PERF_BENCH_DURATION_NOISE_FLOOR_MS"); v != "" {
		if f, err := parseInt64Strict(v); err == nil {
			*noiseFloorMS = f
		} else {
			fmt.Fprintf(os.Stderr, "WARN: invalid PERF_BENCH_DURATION_NOISE_FLOOR_MS=%s (%v)\n", v, err)
		}
	}
	if v := os.Getenv("PERF_BENCH_DURATION_DELTA_FLOOR_MS"); v != "" {
		if f, err := parseInt64Strict(v); err == nil {
			*deltaFloorMS = f
		} else {
			fmt.Fprintf(os.Stderr, "WARN: invalid PERF_BENCH_DURATION_DELTA_FLOOR_MS=%s (%v)\n", v, err)
		}
	}
}

// runScenarioBatch executes a single (chunkSize, rotation) scenario for both legacy and unified
// paths, aggregates metrics across iterations, evaluates thresholds, and returns the result.
func runScenarioBatch(provider, input string, cs int, rot string, iterations int, rotateSize int64, workers int, durationMax, allocMax float64, noiseFloorMS, deltaFloorMS int64, baseline map[string]scenarioResult) scenarioResult {
	scen := scenarioResult{ChunkSize: cs, Rotation: rot, Thresholds: thresholds{DurationMax: durationMax, AllocMax: allocMax}}
	var legacyRuns, unifiedRuns []runMetrics
	for i := 0; i < iterations; i++ {
		lm, err := runConversionScenario(provider, input, cs, rot == "on", rotateSize, workers, false)
		if err != nil {
			scen.Status = statusFail
			scen.Notes = append(scen.Notes, fmt.Sprintf("legacy error(iter %d): %v", i, err))
			return scen
		}
		legacyRuns = append(legacyRuns, lm)
		um, err := runConversionScenario(provider, input, cs, rot == "on", rotateSize, workers, true)
		if err != nil {
			scen.Status = statusFail
			scen.Notes = append(scen.Notes, fmt.Sprintf("unified error(iter %d): %v", i, err))
			return scen
		}
		unifiedRuns = append(unifiedRuns, um)
	}
	if len(legacyRuns) != iterations || len(unifiedRuns) != iterations {
		return scen
	}

	scen.Legacy = aggregateMedian(legacyRuns)
	scen.Unified = aggregateMedian(unifiedRuns)
	scen.Ratios = ratioResult{
		Duration: safeDiv(float64(scen.Unified.DurationMS), float64(scen.Legacy.DurationMS)),
		Alloc:    safeDiv(float64(scen.Unified.AllocBytes), float64(scen.Legacy.AllocBytes)),
	}

	// Baseline deltas if present
	if baseline != nil {
		key := baselineKey(cs, rot)
		if bsc, ok := baseline[key]; ok {
			bm := &baselineMetrics{
				LegacyDurationMS:        bsc.Legacy.DurationMS,
				UnifiedDurationMS:       bsc.Unified.DurationMS,
				LegacyAllocBytes:        bsc.Legacy.AllocBytes,
				UnifiedAllocBytes:       bsc.Unified.AllocBytes,
				LegacyDurationDeltaPct:  pctDelta(float64(scen.Legacy.DurationMS), float64(bsc.Legacy.DurationMS)),
				UnifiedDurationDeltaPct: pctDelta(float64(scen.Unified.DurationMS), float64(bsc.Unified.DurationMS)),
				LegacyAllocDeltaPct:     pctDelta(float64(scen.Legacy.AllocBytes), float64(bsc.Legacy.AllocBytes)),
				UnifiedAllocDeltaPct:    pctDelta(float64(scen.Unified.AllocBytes), float64(bsc.Unified.AllocBytes)),
			}
			scen.Baseline = bm
			if math.Abs(bm.UnifiedDurationDeltaPct) > 10 {
				scen.Notes = append(scen.Notes, fmt.Sprintf("unified duration delta %.1f%% vs baseline", bm.UnifiedDurationDeltaPct))
			}
		}
	}

	// Noise-floor guard
	absDelta := math.Abs(float64(scen.Unified.DurationMS - scen.Legacy.DurationMS))
	if float64(scen.Unified.DurationMS) < float64(noiseFloorMS) && float64(scen.Legacy.DurationMS) < float64(noiseFloorMS) && absDelta <= float64(deltaFloorMS) {
		scen.Status = statusPass
		scen.Notes = append(scen.Notes, fmt.Sprintf("under noise floor (%dms), absΔ=%.0fms; ratio ignored", noiseFloorMS, absDelta))
		return scen
	}

	if scen.Ratios.Duration <= durationMax && scen.Ratios.Alloc <= allocMax {
		scen.Status = statusPass
	} else {
		scen.Status = statusFail
		if scen.Ratios.Duration > durationMax {
			scen.Notes = append(scen.Notes, fmt.Sprintf("duration ratio %.3f > %.2f", scen.Ratios.Duration, durationMax))
		}
		if scen.Ratios.Alloc > allocMax {
			scen.Notes = append(scen.Notes, fmt.Sprintf("alloc ratio %.3f > %.2f", scen.Ratios.Alloc, allocMax))
		}
	}
	return scen
}

// maybeWriteProm writes Prometheus metrics if a non-empty path is provided.
func maybeWriteProm(promPath string, report *benchReport) error {
	if promPath == "" {
		return nil
	}
	if err := writePromMetrics(promPath, report); err != nil {
		fmt.Fprintf(os.Stderr, "write prom metrics failed: %v\n", err)
		return err
	}
	return nil
}

func runConversionScenario(provider, input string, chunkSize int, rotation bool, rotateSize int64, workers int, unified bool) (runMetrics, error) {
	// Temp dir for outputs per run
	base := filepath.Join(os.TempDir(), "costscope-perf-bench")
	_ = os.MkdirAll(base, 0o750)
	mode := "legacy"
	if unified {
		mode = "unified"
	}
	rotLabel := "off"
	rotateBytes := int64(-1)
	if rotation {
		rotLabel = "on"
		rotateBytes = rotateSize
	}
	outDir := filepath.Join(base, fmt.Sprintf("%s_%s_cs%d_%d", mode, rotLabel, chunkSize, time.Now().UnixNano()))
	if err := os.MkdirAll(outDir, 0o750); err != nil {
		return runMetrics{}, err
	}
	outFile := filepath.Join(outDir, "output.parquet")

	cfg := &ftypes.ConversionConfig{
		Provider:    provider,
		InputPath:   input,
		OutputPath:  outFile,
		Streaming:   true,
		ChunkSize:   chunkSize,
		Workers:     workers,
		Compression: true,
		Parquet: ftypes.ParquetOptions{
			CompressionCodec:  "snappy",
			RotateSizeBytes:   rotateBytes,
			RowGroupSizeBytes: 128 * 1024 * 1024,
			PageSizeBytes:     8 * 1024,
			RotateInterval:    "",
			FilePrefix:        "bench",
		},
		UseUnifiedMapper: unified,
		ConversionId:     fmt.Sprintf("bench_%s_%d_%s", mode, chunkSize, rotLabel),
		CreatedAt:        time.Now().UTC(),
		CreatedBy:        "perf-bench",
	}

	// Select converter (currently only aws dataset is guaranteed present)
	// Use concrete converter types directly (they implement ConvertStream but may not satisfy StreamingConverter interface fully in this context).
	var conv interface {
		ConvertStream(context.Context, *ftypes.ConversionConfig, ftypes.ProgressCallback) (*ftypes.ConversionResult, error)
	}
	switch provider {
	case "aws":
		conv = awsc.NewAWSConverter()
	case "azure":
		conv = azc.NewAzureConverter()
	case "gcp":
		conv = gcpp.NewGCPConverter()
	default:
		return runMetrics{}, fmt.Errorf("unsupported provider: %s", provider)
	}

	// GC & memory snapshot
	runtime.GC()
	var mBefore, mAfter runtime.MemStats
	runtime.ReadMemStats(&mBefore)

	start := time.Now()
	res, err := conv.ConvertStream(context.Background(), cfg, nil)
	duration := time.Since(start)
	runtime.ReadMemStats(&mAfter)
	allocBytes := mAfter.TotalAlloc - mBefore.TotalAlloc
	if err != nil {
		return runMetrics{}, err
	}

	// Count output files (may be rotated)
	files := 0
	if walkErr := filepath.WalkDir(outDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && filepath.Ext(path) == ".parquet" {
			files++
		}
		return nil
	}); walkErr != nil {
		return runMetrics{}, walkErr
	}

	rm := runMetrics{
		DurationMS:  duration.Milliseconds(),
		Records:     res.OutputRecords,
		RPS:         safeDiv(float64(res.OutputRecords), duration.Seconds()),
		AllocBytes:  allocBytes,
		AllocMB:     float64(allocBytes) / (1024 * 1024),
		OutputFiles: files,
	}
	return rm, nil
}

func safeDiv(a, b float64) float64 {
	if b == 0 {
		return math.NaN()
	}
	return a / b
}

func writeJSON(path string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

// aggregateMedian returns a runMetrics whose numeric fields are median across runs (by DurationMS & AllocBytes & RPS).
func aggregateMedian(runs []runMetrics) runMetrics {
	if len(runs) == 0 {
		return runMetrics{}
	}
	durs := make([]int64, len(runs))
	allocs := make([]uint64, len(runs))
	rps := make([]float64, len(runs))
	recs := make([]int64, len(runs))
	files := make([]int, len(runs))
	for i, r := range runs {
		durs[i] = r.DurationMS
		allocs[i] = r.AllocBytes
		rps[i] = r.RPS
		recs[i] = r.Records
		files[i] = r.OutputFiles
	}
	sort.Slice(durs, func(i, j int) bool { return durs[i] < durs[j] })
	sort.Slice(allocs, func(i, j int) bool { return allocs[i] < allocs[j] })
	sort.Slice(rps, func(i, j int) bool { return rps[i] < rps[j] })
	sort.Slice(recs, func(i, j int) bool { return recs[i] < recs[j] })
	sort.Slice(files, func(i, j int) bool { return files[i] < files[j] })
	mid := len(runs) / 2
	med := runMetrics{
		DurationMS:  durs[mid],
		AllocBytes:  allocs[mid],
		RPS:         rps[mid],
		Records:     recs[mid],
		OutputFiles: files[mid],
	}
	med.AllocMB = float64(med.AllocBytes) / (1024 * 1024)
	return med
}

func loadBaseline(path string) (map[string]scenarioResult, error) {
	// #nosec G304 path is user-controlled for benchmarking, acceptable risk
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var rep benchReport
	if err := json.Unmarshal(b, &rep); err != nil {
		return nil, err
	}
	out := make(map[string]scenarioResult)
	for _, s := range rep.Scenarios {
		out[baselineKey(s.ChunkSize, s.Rotation)] = s
	}
	return out, nil
}

func baselineKey(chunk int, rotation string) string {
	return fmt.Sprintf("cs%d_%s", chunk, rotation)
}

func pctDelta(cur, base float64) float64 {
	if base == 0 {
		if cur == 0 {
			return 0
		}
		return 100
	}
	return (cur - base) / base * 100
}

// writePromMetrics emits simple text-format Prometheus metrics for dashboard scraping.
func writePromMetrics(path string, rep *benchReport) error {
	// #nosec G304 output path supplied by caller of tool
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	// Metric definitions
	_, _ = fmt.Fprintln(f, "# HELP costscope_perf_bench_duration_ratio Unified vs legacy duration ratio")
	_, _ = fmt.Fprintln(f, "# TYPE costscope_perf_bench_duration_ratio gauge")
	_, _ = fmt.Fprintln(f, "# HELP costscope_perf_bench_alloc_ratio Unified vs legacy allocation ratio")
	_, _ = fmt.Fprintln(f, "# TYPE costscope_perf_bench_alloc_ratio gauge")
	for _, s := range rep.Scenarios {
		_, _ = fmt.Fprintf(f, "costscope_perf_bench_duration_ratio{chunk=\"%d\",rotation=\"%s\"} %.6f\n", s.ChunkSize, s.Rotation, s.Ratios.Duration)
		_, _ = fmt.Fprintf(f, "costscope_perf_bench_alloc_ratio{chunk=\"%d\",rotation=\"%s\"} %.6f\n", s.ChunkSize, s.Rotation, s.Ratios.Alloc)
		if s.Baseline != nil {
			_, _ = fmt.Fprintln(f, "# Baseline deltas (percent)")
			_, _ = fmt.Fprintf(f, "costscope_perf_bench_unified_duration_delta_pct{chunk=\"%d\",rotation=\"%s\"} %.2f\n", s.ChunkSize, s.Rotation, s.Baseline.UnifiedDurationDeltaPct)
			_, _ = fmt.Fprintf(f, "costscope_perf_bench_unified_alloc_delta_pct{chunk=\"%d\",rotation=\"%s\"} %.2f\n", s.ChunkSize, s.Rotation, s.Baseline.UnifiedAllocDeltaPct)
		}
	}
	_, _ = fmt.Fprintf(f, "costscope_perf_bench_scenarios_total %d\n", len(rep.Scenarios))
	if rep.OverallStatus == statusPass {
		_, _ = fmt.Fprintln(f, "costscope_perf_bench_status 1")
	} else {
		_, _ = fmt.Fprintln(f, "costscope_perf_bench_status 0")
	}
	return nil
}

func parseFloatStrict(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}

func parseInt64Strict(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}
