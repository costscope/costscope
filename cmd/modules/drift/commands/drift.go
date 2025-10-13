package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	focquality "github.com/costscope/costscope/internal/core/focus/quality"
	"github.com/costscope/costscope/internal/core/quality/drift"

	"github.com/spf13/cobra"
)

// BuildDriftCommand root command for advanced drift checks.
func BuildDriftCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "drift", Short: "Advanced semantic drift checks (chi-square, buckets, trends)"}
	cmd.AddCommand(buildAnalyzeCommand())
	return cmd
}

func buildAnalyzeCommand() *cobra.Command {
	var (
		baselinePath        string
		currentPath         string
		historyPaths        []string
		alpha               float64
		outputJSON          string
		bucketSchemaStr     string
		baselineDataset     string
		currentDataset      string
		percentilesStr      string
		percentileThreshold float64
	)
	c := &cobra.Command{Use: "analyze", Short: "Run advanced drift analysis producing JSON report", RunE: func(cmd *cobra.Command, args []string) error {
		if baselinePath == "" || currentPath == "" {
			return fmt.Errorf("--baseline and --current required")
		}
		// Parse schema & percentiles early
		schema, err := parseFloatList(bucketSchemaStr)
		if err != nil {
			return err
		}
		percentiles, err := parseFloatList(percentilesStr)
		if err != nil {
			return err
		}
		rep, err := runDriftAnalyze(runParams{
			BaselinePath:        baselinePath,
			CurrentPath:         currentPath,
			History:             historyPaths,
			Alpha:               alpha,
			BucketSchema:        schema,
			BaselineData:        baselineDataset,
			CurrentData:         currentDataset,
			Percentiles:         percentiles,
			PercentileThreshold: percentileThreshold,
		})
		if err != nil {
			return err
		}
		drift.RecordMetrics(rep)
		if outputJSON != "" {
			if err := os.MkdirAll("./", 0o750); err != nil {
				return err
			}
			b, _ := json.MarshalIndent(rep, "", "  ")
			if err := os.WriteFile(outputJSON, b, 0o600); err != nil {
				return err
			}
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(rep)
	}}
	c.Flags().StringVar(&baselinePath, "baseline", "", "Baseline invariants JSON (required)")
	c.Flags().StringVar(&currentPath, "current", "", "Current invariants JSON (required)")
	c.Flags().StringSliceVar(&historyPaths, "history", nil, "Optional list of older invariants JSON for trend slope computation (oldest->newest)")
	c.Flags().Float64Var(&alpha, "alpha", 0.01, "Significance threshold for chi-square")
	c.Flags().StringVar(&outputJSON, "output", "", "Optional path to write drift report JSON")
	c.Flags().StringVar(&bucketSchemaStr, "bucket-schema", "", "Comma-separated bucket boundaries (e.g. 0.01,0.1,1,10,100)")
	c.Flags().StringVar(&baselineDataset, "baseline-dataset", "", "Optional baseline dataset (parquet/csv/json) for precise buckets")
	c.Flags().StringVar(&currentDataset, "current-dataset", "", "Optional current dataset (parquet/csv/json) for precise buckets")
	c.Flags().StringVar(&percentilesStr, "percentiles", "", "Comma-separated percentiles (0-1) e.g. 0.5,0.9,0.95,0.99")
	c.Flags().Float64Var(&percentileThreshold, "percentile-threshold", 0.01, "Relative percentile drift threshold (default 0.01 = 1%)")
	return c
}

// ---- refactored helpers (keep simple to reduce cyclomatic complexity of main command) ----
type runParams struct {
	BaselinePath        string
	CurrentPath         string
	History             []string
	Alpha               float64
	BucketSchema        []float64
	BaselineData        string
	CurrentData         string
	Percentiles         []float64
	PercentileThreshold float64
}

func runDriftAnalyze(p runParams) (drift.Report, error) { //nolint:funlen
	baseInvVal, err := focquality.LoadBaseline(p.BaselinePath)
	if err != nil {
		return drift.Report{}, fmt.Errorf("load baseline invariants: %w", err)
	}
	curInvVal, err := focquality.LoadBaseline(p.CurrentPath)
	if err != nil {
		return drift.Report{}, fmt.Errorf("load current invariants: %w", err)
	}
	baseInv, curInv := &baseInvVal, &curInvVal
	baseCharge, curCharge, basePricing, curPricing := buildDistributions(baseInv, curInv)
	baseEffVals, baseUseVals, curEffVals, curUseVals, err := collectValueSlices(p, baseInv, curInv)
	if err != nil {
		return drift.Report{}, err
	}
	baseBuckets, _ := drift.BuildCostBuckets(baseEffVals, baseUseVals, p.BucketSchema)
	curBuckets, _ := drift.BuildCostBuckets(curEffVals, curUseVals, p.BucketSchema)
	hist := buildHistorySnapshots(p.History)
	curSnap := drift.Snapshot{TimestampUnix: curInv.GeneratedAt.Unix(), RowCount: int64(curInv.RowCount), SumEffective: curInv.SumEffectiveCost, SumList: curInv.SumListCost, SumUsage: curInv.SumUsageQuantity}
	return drift.Run(drift.Config{Alpha: p.Alpha, BucketSchema: p.BucketSchema, Percentiles: p.Percentiles, PercentileDriftThreshold: p.PercentileThreshold},
		baseCharge, curCharge, basePricing, curPricing, baseBuckets, curBuckets, hist, curSnap,
		baseEffVals, curEffVals, baseUseVals, curUseVals,
	)
}

func buildDistributions(baseInv, curInv *focquality.InvariantMetrics) (map[string]float64, map[string]float64, map[string]float64, map[string]float64) {
	baseCharge := map[string]float64{}
	curCharge := map[string]float64{}
	basePricing := map[string]float64{}
	curPricing := map[string]float64{}
	for k, v := range baseInv.ChargeCategoryDistribution {
		baseCharge[k] = v * float64(baseInv.RowCount) / 100
	}
	for k, v := range curInv.ChargeCategoryDistribution {
		curCharge[k] = v * float64(curInv.RowCount) / 100
	}
	for k, v := range baseInv.PricingCategoryDistribution {
		basePricing[k] = v * float64(baseInv.RowCount) / 100
	}
	for k, v := range curInv.PricingCategoryDistribution {
		curPricing[k] = v * float64(curInv.RowCount) / 100
	}
	return baseCharge, curCharge, basePricing, curPricing
}

func collectValueSlices(p runParams, baseInv, curInv *focquality.InvariantMetrics) ([]float64, []float64, []float64, []float64, error) {
	var baseEffVals, baseUseVals, curEffVals, curUseVals []float64
	if p.BaselineData != "" {
		be, bu, err := loadValues(p.BaselineData)
		if err != nil {
			return nil, nil, nil, nil, fmt.Errorf("baseline dataset load: %w", err)
		}
		baseEffVals, baseUseVals = be, bu
	}
	if p.CurrentData != "" {
		ce, cu, err := loadValues(p.CurrentData)
		if err != nil {
			return nil, nil, nil, nil, fmt.Errorf("current dataset load: %w", err)
		}
		curEffVals, curUseVals = ce, cu
	}
	if len(baseEffVals) == 0 { // averages fallback
		avgEff, avgUse := averages(baseInv)
		baseEffVals, baseUseVals = []float64{avgEff}, []float64{avgUse}
	}
	if len(curEffVals) == 0 {
		avgEff, avgUse := averages(curInv)
		curEffVals, curUseVals = []float64{avgEff}, []float64{avgUse}
	}
	return baseEffVals, baseUseVals, curEffVals, curUseVals, nil
}

func averages(inv *focquality.InvariantMetrics) (float64, float64) {
	if inv.RowCount == 0 {
		return 0, 0
	}
	return inv.SumEffectiveCost / float64(inv.RowCount), inv.SumUsageQuantity / float64(inv.RowCount)
}

func buildHistorySnapshots(paths []string) []drift.Snapshot {
	out := make([]drift.Snapshot, 0, len(paths))
	for _, p := range paths {
		inv, err := focquality.LoadBaseline(p)
		if err != nil {
			continue
		}
		out = append(out, drift.Snapshot{TimestampUnix: inv.GeneratedAt.Unix(), RowCount: int64(inv.RowCount), SumEffective: inv.SumEffectiveCost, SumList: inv.SumListCost, SumUsage: inv.SumUsageQuantity})
	}
	return out
}

func parseFloatList(csv string) ([]float64, error) {
	if strings.TrimSpace(csv) == "" {
		return nil, nil
	}
	parts := strings.Split(csv, ",")
	out := make([]float64, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		f, err := strconv.ParseFloat(p, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid float %q: %w", p, err)
		}
		out = append(out, f)
	}
	return out, nil
}

// Dataset loader helpers are provided in tagged files (drift_load_duckdb.go and drift_load_noduckdb.go)
