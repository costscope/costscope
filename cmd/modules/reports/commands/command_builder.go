package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"local/costscope/internal/core/logging"
	"local/costscope/internal/core/reports"
	"local/costscope/internal/core/reports/exporters"
	"local/costscope/internal/core/reports/types"
)

// ReportsCommands handles all report-related CLI commands
type ReportsCommands struct {
	reportService *reports.BasicReportService
	logger        *logging.Logger
}

// NewReportsCommands creates a new instance of ReportsCommands
func NewReportsCommands(
	reportService *reports.BasicReportService,
	logger *logging.Logger,
) *ReportsCommands {
	return &ReportsCommands{
		reportService: reportService,
		logger:        logger,
	}
}

// BuildReportsCommand builds the main reports command with all subcommands
func (rc *ReportsCommands) BuildReportsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reports",
		Short: "Generate and manage cost analysis reports",
		Long: `Generate comprehensive cost analysis reports in various formats.
Supports cost analysis, usage summaries, trend analysis, anomaly detection,
forecasting, and executive summaries.`,
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
		},
	}

	cmd.AddCommand(rc.buildGenerateCommand())
	cmd.AddCommand(rc.buildListCommand())
	cmd.AddCommand(rc.buildExportCommand())
	cmd.AddCommand(rc.buildDeleteCommand())
	cmd.AddCommand(rc.buildResolveOutputCommand())
	cmd.AddCommand(rc.buildMetadataListCommand())
	cmd.AddCommand(rc.buildIntegrityVerifyCommand())

	return cmd
}

// buildGenerateCommand builds the generate subcommand
func (rc *ReportsCommands) buildGenerateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate various types of reports",
		Long:  "Generate cost analysis, usage, trend, anomaly, forecast, or executive summary reports",
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
		},
	}

	// Spec-driven registration to avoid duplicated builders
	type spec struct {
		use       string
		short     string
		long      string
		rType     types.ReportType
		display   string
		nounLower string
		gen       func(ctx context.Context, opts *types.ReportOptions) (interface{}, error)
	}

	specs := []spec{
		{
			use:       "cost-analysis",
			short:     "Generate a comprehensive cost analysis report",
			long:      "Analyze costs by service, region, account and identify optimization opportunities",
			rType:     types.ReportTypeCostAnalysis,
			display:   "Cost Analysis Report",
			nounLower: "cost analysis report",
			gen: func(ctx context.Context, opts *types.ReportOptions) (interface{}, error) {
				return rc.reportService.GenerateCostAnalysisReport(ctx, opts)
			},
		},
		{
			use:       "usage-summary",
			short:     "Generate a resource usage summary report",
			long:      "Summarize resource utilization across services, regions, and accounts",
			rType:     types.ReportTypeUsageSummary,
			display:   "Usage Summary Report",
			nounLower: "usage summary report",
			gen: func(ctx context.Context, opts *types.ReportOptions) (interface{}, error) {
				return rc.reportService.GenerateUsageSummaryReport(ctx, opts)
			},
		},
		{
			use:       "trend-analysis",
			short:     "Generate a cost trend analysis report",
			long:      "Analyze cost trends over time to identify patterns and anomalies",
			rType:     types.ReportTypeTrendAnalysis,
			display:   "Trend Analysis Report",
			nounLower: "trend analysis report",
			gen: func(ctx context.Context, opts *types.ReportOptions) (interface{}, error) {
				return rc.reportService.GenerateTrendAnalysisReport(ctx, opts)
			},
		},
		{
			use:       "anomaly",
			short:     "Generate an anomaly detection report",
			long:      "Detect unusual spending patterns and potential cost anomalies",
			rType:     types.ReportTypeAnomaly,
			display:   "Anomaly Report",
			nounLower: "anomaly report",
			gen: func(ctx context.Context, opts *types.ReportOptions) (interface{}, error) {
				return rc.reportService.GenerateAnomalyReport(ctx, opts)
			},
		},
		{
			use:       "forecast",
			short:     "Generate a cost forecast report",
			long:      "Predict future costs based on historical data and trends",
			rType:     types.ReportTypeForecast,
			display:   "Forecast Report",
			nounLower: "forecast report",
			gen: func(ctx context.Context, opts *types.ReportOptions) (interface{}, error) {
				return rc.reportService.GenerateForecastReport(ctx, opts)
			},
		},
		{
			use:       "executive-summary",
			short:     "Generate an executive summary report",
			long:      "Create a high-level summary suitable for executives and management",
			rType:     types.ReportTypeExecutiveSummary,
			display:   "Executive Summary Report",
			nounLower: "executive summary report",
			gen: func(ctx context.Context, opts *types.ReportOptions) (interface{}, error) {
				return rc.reportService.GenerateExecutiveSummaryReport(ctx, opts)
			},
		},
	}

	for _, s := range specs {
		cmd.AddCommand(rc.buildReportGenCommand(s.use, s.short, s.long, s.rType, s.display, s.nounLower, s.gen))
	}

	return cmd
}

func (rc *ReportsCommands) buildListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List available reports",
		Long:  "List all generated reports with their metadata",
		RunE: func(cmd *cobra.Command, args []string) error {
			reportType, _ := cmd.Flags().GetString("type")
			limit, _ := cmd.Flags().GetInt("limit")

			filters := &types.ReportFilters{
				Limit: limit,
			}
			if reportType != "" {
				rt := types.ReportType(reportType)
				filters.ReportType = &rt
			}

			reports, err := rc.reportService.ListReports(context.Background(), filters)
			if err != nil {
				return fmt.Errorf("failed to list reports: %w", err)
			}

			return rc.displayReportList(reports)
		},
	}
	cmd.Flags().String("type", "", "Filter by report type")
	cmd.Flags().Int("limit", 10, "Maximum number of reports to list")
	return cmd
}

func (rc *ReportsCommands) buildExportCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export [report-id]",
		Short: "Export a report to a file",
		Long:  "Export an existing report to JSON, CSV, YAML, Parquet, or send via HTTP",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			reportID := args[0]
			format, _ := cmd.Flags().GetString("format")
			outputPath, _ := cmd.Flags().GetString("output")
			baseDir, _ := cmd.Flags().GetString("base-dir")
			includeContent, _ := cmd.Flags().GetBool("include-content")

			// HTTP exporter requires a destination URL; don't synthesize a filename.
			if outputPath == "" {
				if strings.EqualFold(format, string(types.ExportFormatHTTP)) {
					return fmt.Errorf("output URL is required for http export (e.g., https://example/api/reports)")
				}
				// Preserve legacy naming scheme (report_<id>) to avoid breaking expectations even though resolver uses timestamp naming.
				outputPath = fmt.Sprintf("report_%s.%s", reportID, format)
			}

			exportFormat := types.ExportFormat(format)

			// Inject optional auth headers for HTTP exporter from environment, per README
			// COSTSCOPE_HTTP_BEARER_TOKEN -> Authorization: Bearer <token>
			// COSTSCOPE_HTTP_API_KEY      -> X-API-Key: <key>
			ctx := context.Background()
			if includeContent {
				ctx = reports.WithIncludeContent(ctx, true)
			}
			if exportFormat == types.ExportFormatHTTP {
				if tok := strings.TrimSpace(os.Getenv("COSTSCOPE_HTTP_BEARER_TOKEN")); tok != "" {
					ctx = exporters.WithBearerToken(ctx, tok)
				}
				if key := strings.TrimSpace(os.Getenv("COSTSCOPE_HTTP_API_KEY")); key != "" {
					ctx = exporters.WithAPIKey(ctx, key)
				}
				ctx = reports.WithOutputDirSource(ctx, "explicit")
			} else {
				// Classify precedence source using resolver without altering chosen legacy output filename.
				resolver := reports.NewReportPathResolver(rc.logger)
				// Only treat output as explicit if user set --output (Cobra flag changed).
				var explicitForResolver string
				if f := cmd.Flags().Lookup("output"); f != nil && f.Changed {
					explicitForResolver = outputPath
				}
				if res, err := resolver.Resolve(context.Background(), &reports.ResolveRequest{BaseDir: baseDir, ExplicitPath: explicitForResolver, Format: exportFormat}); err == nil {
					ctx = reports.WithOutputDirSource(ctx, res.Source)
				} else {
					rc.logger.Warn(fmt.Sprintf("failed to resolve output precedence for export: %v", err))
				}
			}

			err := rc.reportService.ExportReport(ctx, reportID, exportFormat, outputPath)
			if err != nil {
				return fmt.Errorf("failed to export report: %w", err)
			}
			fmt.Printf(" Report exported successfully to: %s\n", outputPath)
			return nil
		},
	}
	cmd.Flags().String("format", "json", "Export format (json, csv, yaml, parquet, http)")
	cmd.Flags().String("output", "", "Output file path")
	cmd.Flags().String("base-dir", "", "Explicit base output directory (used for precedence classification when --output omitted)")
	cmd.Flags().Bool("include-content", false, "Include full report content where available (otherwise exports metadata)")
	return cmd
}

// buildResolveOutputCommand adds a helper to preview the resolved output path precedence
// without generating a report (MVP exposure of dead code path).
func (rc *ReportsCommands) buildResolveOutputCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resolve-output",
		Short: "Resolve final report output path given flags and config precedence",
		RunE: func(cmd *cobra.Command, args []string) error {
			formatStr, _ := cmd.Flags().GetString("format")
			baseDir, _ := cmd.Flags().GetString("base-dir")
			explicitPath, _ := cmd.Flags().GetString("output")
			if formatStr == "" {
				formatStr = string(types.ExportFormatJSON)
			}
			format := types.ExportFormat(formatStr)
			resolver := reports.NewReportPathResolver(rc.logger)
			res, err := resolver.Resolve(context.Background(), &reports.ResolveRequest{BaseDir: baseDir, ExplicitPath: explicitPath, Format: format})
			if err != nil {
				return err
			}
			if _, err := fmt.Fprintf(cmd.OutOrStdout(), "%s\n", res.Path); err != nil {
				return err
			}
			return nil
		},
	}
	cmd.Flags().String("format", string(types.ExportFormatJSON), "Export format (json|csv|yaml|pdf|excel|parquet|http)")
	cmd.Flags().String("base-dir", "", "Explicit output directory (overrides config dir; ignored if --output set)")
	cmd.Flags().String("output", "", "Explicit full output file path (highest precedence)")
	return cmd
}

func (rc *ReportsCommands) buildDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete [report-id]",
		Short: "Delete a report",
		Long:  "Delete an existing report from storage",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			reportID := args[0]
			err := rc.reportService.DeleteReport(context.Background(), reportID)
			if err != nil {
				return fmt.Errorf("failed to delete report: %w", err)
			}
			fmt.Printf(" Report %s deleted successfully\n", reportID)
			return nil
		},
	}
	return cmd
}

// buildMetadataListCommand lists export metadata (exported files) with optional filters/pagination.
func (rc *ReportsCommands) buildMetadataListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "exports",
		Short: "List exported report artifacts (metadata store)",
		RunE: func(cmd *cobra.Command, args []string) error {
			format, _ := cmd.Flags().GetString("format")
			afterStr, _ := cmd.Flags().GetString("after")
			beforeStr, _ := cmd.Flags().GetString("before")
			limit, _ := cmd.Flags().GetInt("limit")
			offset, _ := cmd.Flags().GetInt("offset")
			jsonOut, _ := cmd.Flags().GetBool("json")
			var afterPtr *time.Time
			if afterStr != "" {
				if t, err := time.Parse(time.RFC3339, afterStr); err == nil {
					afterPtr = &t
				}
			}
			var beforePtr *time.Time
			if beforeStr != "" {
				if t, err := time.Parse(time.RFC3339, beforeStr); err == nil {
					beforePtr = &t
				}
			}
			opts := &reports.MetadataListOptions{Format: format, CreatedAfter: afterPtr, CreatedBefore: beforePtr, Limit: limit, Offset: offset}
			list, err := rc.reportService.ListReportMetadataOptions(context.Background(), opts)
			if err != nil {
				return err
			}
			if jsonOut {
				// JSON envelope (minimal, scripting-friendly). No total count (would require full scan when limit/offset used).
				envelope := struct {
					Success bool                      `json:"success"`
					Count   int                       `json:"count"`
					Limit   int                       `json:"limit"`
					Offset  int                       `json:"offset"`
					Items   []*reports.ReportMetadata `json:"items"`
				}{Success: true, Count: len(list), Limit: limit, Offset: offset, Items: list}
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				if err := enc.Encode(envelope); err != nil {
					return err
				}
				return nil
			}
			if len(list) == 0 {
				if _, err := fmt.Fprintln(cmd.OutOrStdout(), "No exported reports found."); err != nil { // errcheck
					return err
				}
				return nil
			}
			for _, md := range list {
				if _, err := fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\t%s\t%d bytes\t%s\n", md.ID, md.Format, md.CreatedAt.Format(time.RFC3339), md.SizeBytes, md.Path); err != nil { // errcheck
					return err
				}
			}
			return nil
		},
	}
	cmd.Flags().String("format", "", "Filter by export format (json|csv|yaml|parquet|pdf|excel|http)")
	cmd.Flags().String("after", "", "Created after (RFC3339)")
	cmd.Flags().String("before", "", "Created before (RFC3339)")
	cmd.Flags().Int("limit", 50, "Limit number of results")
	cmd.Flags().Int("offset", 0, "Offset for pagination")
	cmd.Flags().Bool("json", false, "Output JSON envelope (scripting)")
	return cmd
}

// buildIntegrityVerifyCommand verifies checksum of a stored export.
func (rc *ReportsCommands) buildIntegrityVerifyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "verify-export [report-id]",
		Short: "Verify integrity (checksum) of an exported report file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			match, actual, err := rc.reportService.VerifyReportIntegrity(context.Background(), id)
			if err != nil {
				return err
			}
			if match {
				if _, err := fmt.Fprintf(cmd.OutOrStdout(), "Checksum OK (%s)\n", actual); err != nil { // errcheck
					return err
				}
			} else {
				if _, err := fmt.Fprintf(cmd.OutOrStdout(), "Checksum mismatch (actual %s)\n", actual); err != nil { // errcheck
					return err
				}
			}
			return nil
		},
	}
	return cmd
}

func (rc *ReportsCommands) addReportFlags(cmd *cobra.Command) {
	cmd.Flags().String("title", "", "Report title")
	cmd.Flags().String("description", "", "Report description")
	cmd.Flags().String("start-date", "", "Start date (YYYY-MM-DD)")
	cmd.Flags().String("end-date", "", "End date (YYYY-MM-DD)")
	cmd.Flags().StringSlice("providers", []string{}, "Cloud providers to include")
	cmd.Flags().StringSlice("regions", []string{}, "Regions to include")
	cmd.Flags().StringSlice("services", []string{}, "Services to include")
	cmd.Flags().StringSlice("accounts", []string{}, "Accounts to include")
	cmd.Flags().String("currency", "USD", "Currency for cost calculations")
	cmd.Flags().StringSlice("group-by", []string{}, "Group results by fields")
	cmd.Flags().Bool("include-ml", false, "Include machine learning insights")
	cmd.Flags().String("detail-level", "standard", "Detail level (basic, standard, detailed)")
	cmd.Flags().String("output", "", "Output file path")
	cmd.Flags().Bool("dry-run-output-resolution", false, "Only resolve and print final output path (no generation)")
	// These mirror standalone resolve-output flags so dry-run resolution can reuse identical semantics.
	cmd.Flags().String("format", string(types.ExportFormatJSON), "Output format for path resolution (json|csv|yaml|pdf|excel|parquet|http)")
	cmd.Flags().String("base-dir", "", "Explicit output directory (ignored if --output set)")
}

// buildReportGenCommand is a generic helper to build a report generation subcommand.
// It wires common flags and execution flow, delegating the actual generation to gen.
func (rc *ReportsCommands) buildReportGenCommand(
	use string,
	short string,
	long string,
	rType types.ReportType,
	displayTitle string,
	nounLower string,
	gen func(ctx context.Context, opts *types.ReportOptions) (interface{}, error),
) *cobra.Command {
	cmd := &cobra.Command{
		Use:   use,
		Short: short,
		Long:  long,
		RunE: func(cmd *cobra.Command, args []string) error {
			options := rc.parseReportOptions(cmd, rType)
			if dry, _ := cmd.Flags().GetBool("dry-run-output-resolution"); dry {
				formatStr, _ := cmd.Flags().GetString("format")
				format := types.ExportFormat(formatStr)
				baseDir, _ := cmd.Flags().GetString("base-dir")
				resolver := reports.NewReportPathResolver(rc.logger)
				res, err := resolver.Resolve(context.Background(), &reports.ResolveRequest{BaseDir: baseDir, ExplicitPath: options.OutputPath, Format: format})
				if err != nil {
					return err
				}
				if _, err := fmt.Fprintln(cmd.OutOrStdout(), res.Path); err != nil {
					return err
				}
				return nil
			}
			rc.logger.Info("Generating " + nounLower)
			report, err := gen(context.Background(), options)
			if err != nil {
				return fmt.Errorf("failed to generate %s: %w", nounLower, err)
			}
			return rc.displayReport(displayTitle, report)
		},
	}
	rc.addReportFlags(cmd)
	return cmd
}

func (rc *ReportsCommands) parseReportOptions(cmd *cobra.Command, reportType types.ReportType) *types.ReportOptions {
	options := &types.ReportOptions{
		ReportType: reportType,
	}
	if title, _ := cmd.Flags().GetString("title"); title != "" {
		options.Title = title
	} else {
		options.Title = fmt.Sprintf("%s Report", strings.ReplaceAll(string(reportType), "_", " "))
	}
	if desc, _ := cmd.Flags().GetString("description"); desc != "" {
		options.Description = desc
	}
	now := time.Now()
	startDate := now.AddDate(0, -1, 0)
	endDate := now
	if startStr, _ := cmd.Flags().GetString("start-date"); startStr != "" {
		if parsed, err := time.Parse("2006-01-02", startStr); err == nil {
			startDate = parsed
		}
	}
	if endStr, _ := cmd.Flags().GetString("end-date"); endStr != "" {
		if parsed, err := time.Parse("2006-01-02", endStr); err == nil {
			endDate = parsed
		}
	}
	options.DateRange = types.DateRange{
		StartDate: startDate,
		EndDate:   endDate,
	}
	options.Providers, _ = cmd.Flags().GetStringSlice("providers")
	options.Regions, _ = cmd.Flags().GetStringSlice("regions")
	options.Services, _ = cmd.Flags().GetStringSlice("services")
	options.Accounts, _ = cmd.Flags().GetStringSlice("accounts")
	options.Currency, _ = cmd.Flags().GetString("currency")
	options.GroupBy, _ = cmd.Flags().GetStringSlice("group-by")
	options.IncludeML, _ = cmd.Flags().GetBool("include-ml")
	options.DetailLevel, _ = cmd.Flags().GetString("detail-level")
	options.OutputPath, _ = cmd.Flags().GetString("output")
	return options
}

func (rc *ReportsCommands) displayReport(title string, report interface{}) error {
	fmt.Printf("\n %s\n", title)
	fmt.Println(strings.Repeat("=", len(title)+4))
	switch r := report.(type) {
	case *types.CostAnalysisReport:
		rc.displayCostAnalysisReport(r)
	case *types.UsageSummaryReport:
		rc.displayUsageSummaryReport(r)
	case *types.TrendAnalysisReport:
		rc.displayTrendAnalysisReport(r)
	case *types.AnomalyReport:
		rc.displayAnomalyReport(r)
	case *types.ForecastReport:
		rc.displayForecastReport(r)
	case *types.ExecutiveSummaryReport:
		rc.displayExecutiveSummaryReport(r)
	default:
		fmt.Printf("Report ID: %v\n", report)
	}
	fmt.Println()
	return nil
}

func (rc *ReportsCommands) displayCostAnalysisReport(report *types.CostAnalysisReport) {
	fmt.Printf(" Report ID: %s\n", report.ID)
	fmt.Printf(" Title: %s\n", report.Title)
	fmt.Printf(" Total Cost: %.2f %s\n", report.TotalCost, report.Currency)
	fmt.Printf(" Generated: %s\n", report.GeneratedAt.Format(time.RFC3339))
	if len(report.CostByService) > 0 {
		fmt.Println("\n Top Services:")
		for i, service := range report.CostByService {
			if i >= 5 {
				break
			}
			fmt.Printf("  - %s: %.2f %s\n", service.ServiceName, service.Cost, report.Currency)
		}
	}
}

func (rc *ReportsCommands) displayUsageSummaryReport(report *types.UsageSummaryReport) {
	fmt.Printf(" Report ID: %s\n", report.ID)
	fmt.Printf(" Title: %s\n", report.Title)
	fmt.Printf(" Generated: %s\n", report.GeneratedAt.Format(time.RFC3339))
	if len(report.ResourceUtilization) > 0 {
		fmt.Println("\n Resource Utilization:")
		for i, resource := range report.ResourceUtilization {
			if i >= 5 {
				break
			}
			fmt.Printf("  - %s (%s): %.1f%%\n", resource.ResourceID, resource.ResourceType, resource.Utilization)
		}
	}
}

func (rc *ReportsCommands) displayTrendAnalysisReport(report *types.TrendAnalysisReport) {
	fmt.Printf(" Report ID: %s\n", report.ID)
	fmt.Printf(" Title: %s\n", report.Title)
	fmt.Printf(" Generated: %s\n", report.GeneratedAt.Format(time.RFC3339))
	if len(report.Trends) > 0 {
		fmt.Println("\n Trend Analysis:")
		for i, trend := range report.Trends {
			if i >= 5 {
				break
			}
			fmt.Printf("  - %s: %s\n", trend.Metric, trend.Direction)
		}
	}
}

func (rc *ReportsCommands) displayAnomalyReport(report *types.AnomalyReport) {
	fmt.Printf(" Report ID: %s\n", report.ID)
	fmt.Printf(" Title: %s\n", report.Title)
	fmt.Printf("️  Risk Level: %s\n", report.RiskLevel)
	fmt.Printf(" Generated: %s\n", report.GeneratedAt.Format(time.RFC3339))
	if len(report.Anomalies) > 0 {
		fmt.Println("\n Detected Anomalies:")
		for i, anomaly := range report.Anomalies {
			if i >= 5 {
				break
			}
			fmt.Printf("  - %s: %.2f (%.1f deviation)\n",
				anomaly.Metric, anomaly.Actual, anomaly.Deviation)
		}
	}
}

func (rc *ReportsCommands) displayForecastReport(report *types.ForecastReport) {
	fmt.Printf(" Report ID: %s\n", report.ID)
	fmt.Printf(" Title: %s\n", report.Title)
	fmt.Printf(" Confidence: %.1f%%\n", report.Confidence*100)
	fmt.Printf(" Generated: %s\n", report.GeneratedAt.Format(time.RFC3339))
	if len(report.Forecasts) > 0 {
		fmt.Println("\n Forecast Data:")
		for i, forecast := range report.Forecasts {
			if i >= 3 {
				break
			}
			fmt.Printf("  - %s: %.2f (%.2f-%.2f)\n",
				forecast.Date.Format("2006-01"), forecast.Value,
				forecast.ConfidenceLow, forecast.ConfidenceHigh)
		}
	}
}

func (rc *ReportsCommands) displayExecutiveSummaryReport(report *types.ExecutiveSummaryReport) {
	fmt.Printf(" Report ID: %s\n", report.ID)
	fmt.Printf(" Title: %s\n", report.Title)
	fmt.Printf(" Generated: %s\n", report.GeneratedAt.Format(time.RFC3339))
	exec := report.ExecutiveSummary
	fmt.Printf(" Total Spend: %.2f\n", exec.TotalSpend)
	fmt.Printf(" Spend Change: %.1f%%\n", exec.SpendChange)
	fmt.Printf(" Top Cost Driver: %s\n", exec.TopCostDriver)
	fmt.Printf("️  Risk Level: %s\n", exec.RiskLevel)
	if len(exec.KeyInsights) > 0 {
		fmt.Println("\n Key Insights:")
		for i, insight := range exec.KeyInsights {
			if i >= 3 {
				break
			}
			fmt.Printf("  - %s\n", insight)
		}
	}
}

func (rc *ReportsCommands) displayReportList(reports []*types.ReportInfo) error {
	if len(reports) == 0 {
		fmt.Println("No reports found.")
		return nil
	}
	fmt.Printf("\n Found %d report(s):\n\n", len(reports))
	for _, report := range reports {
		fmt.Printf(" %s\n", report.ID)
		fmt.Printf("   Title: %s\n", report.Title)
		fmt.Printf("   Type: %s\n", report.ReportType)
		fmt.Printf("   Created: %s\n", report.CreatedAt.Format(time.RFC3339))
		fmt.Printf("   Size: %d bytes\n", report.Size)
		fmt.Println()
	}
	return nil
}
