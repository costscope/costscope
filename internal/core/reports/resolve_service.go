package reports

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/costscope/costscope/internal/core/config"
	"github.com/costscope/costscope/internal/core/config/precedence"
	"github.com/costscope/costscope/internal/core/logging"
	"github.com/costscope/costscope/internal/core/monitoring/telemetry"
	"github.com/costscope/costscope/internal/core/reports/outputpath"
	"github.com/costscope/costscope/internal/core/reports/types"
)

// ResolveRequest expresses inputs for resolving final output path.
type ResolveRequest struct {
	BaseDir      string
	ExplicitPath string
	Format       types.ExportFormat
}

// ResolveResult returns resolved data.
type ResolveResult struct {
	Path   string
	Source string // explicit|yaml|env|default (captured from structured log by caller if needed)
	// Collision detection intentionally omitted in MVP (added when persistence present)
}

// ReportPathResolver wraps logic; BasicReportService will embed/compose later if expanded.
type ReportPathResolver struct {
	logger *logging.Logger
}

func NewReportPathResolver(log *logging.Logger) *ReportPathResolver {
	return &ReportPathResolver{logger: log}
}

// Resolve performs precedence-based resolution and records metrics.
func (r *ReportPathResolver) Resolve(ctx context.Context, req *ResolveRequest) (*ResolveResult, error) {
	start := time.Now()
	src := precedence.SourceDefault
	if req.ExplicitPath != "" {
		src = precedence.SourceExplicit
	} else if req.BaseDir != "" { // base dir explicit override
		src = precedence.SourceExplicit
	} else if cfg := config.LoadOptionalYAML(r.logger); cfg != nil && cfg.Reports.OutputDir != "" && os.Getenv("COSTSCOPE_REPORTS_DIR") == "" {
		src = precedence.SourceYAML
	} else if v := os.Getenv("COSTSCOPE_REPORTS_DIR"); v != "" {
		src = precedence.SourceEnv
	}
	out, err := outputpath.ResolveOutputPath(req.BaseDir, req.ExplicitPath, req.Format)
	telemetry.ReportsResolveRequests.WithLabelValues(req.Format.String()).Inc()
	if err != nil {
		return nil, fmt.Errorf("resolve output path: %w", err)
	}
	// Collision metric: if file exists and non-empty, increment.
	if fi, statErr := os.Stat(out); statErr == nil && !fi.IsDir() && fi.Size() > 0 {
		telemetry.ReportsResolveCollisions.Inc()
	}
	telemetry.ReportsResolveLatency.WithLabelValues(req.Format.String(), string(src)).Observe(time.Since(start).Seconds())
	r.logger.Info(fmt.Sprintf("report_output_resolved path=%s format=%s source=%s", out, req.Format.String(), src))
	return &ResolveResult{Path: out, Source: string(src)}, nil
}
