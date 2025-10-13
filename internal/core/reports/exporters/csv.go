package exporters

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/costscope/costscope/internal/core/reports/outputpath"
	"github.com/costscope/costscope/internal/core/reports/types"
)

// CSVExporter implements exporting reports to CSV format
type CSVExporter struct{}

// NewCSVExporter creates a new CSV exporter
func NewCSVExporter() *CSVExporter {
	return &CSVExporter{}
}

// Export exports a report to CSV format
func (e *CSVExporter) Export(ctx context.Context, report interface{}, format types.ExportFormat, output string) (int64, string, error) {
	if format != types.ExportFormatCSV {
		return 0, "", fmt.Errorf("unsupported format for CSV exporter: %s", format)
	}
	// Resolve destination and write via ObjectStore (supports fs, file://, s3://, gs://)
	dest, err := outputpath.ResolveOutputPath("", output, types.ExportFormatCSV)
	if err != nil {
		return 0, "", err
	}
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Convert report to CSV based on type
	if err := e.writeReportToCSV(writer, report); err != nil {
		return 0, "", fmt.Errorf("failed to write CSV data: %w", err)
	}

	// Flush before accessing buffer
	writer.Flush()
	if err := writer.Error(); err != nil {
		return 0, "", fmt.Errorf("csv writer error: %w", err)
	}

	store, _, err := NewObjectStore(ctx, dest)
	if err != nil {
		return 0, "", err
	}
	data := buf.Bytes()
	if err := store.Put(ctx, dest, bytes.NewReader(data)); err != nil {
		return 0, "", fmt.Errorf("failed to store CSV output: %w", err)
	}
	sum := sha256.Sum256(data)
	return int64(len(data)), hex.EncodeToString(sum[:]), nil
}

// GetSupportedFormats returns the formats supported by this exporter
func (e *CSVExporter) GetSupportedFormats() []types.ExportFormat {
	return []types.ExportFormat{types.ExportFormatCSV}
}

// writeReportToCSV writes report data to CSV writer
func (e *CSVExporter) writeReportToCSV(writer *csv.Writer, report interface{}) error {
	// Handle different report types
	switch r := report.(type) {
	case *types.CostAnalysisReport:
		return e.writeCostAnalysisToCSV(writer, r)
	case *types.UsageSummaryReport:
		return e.writeUsageSummaryToCSV(writer, r)
	case *types.TrendAnalysisReport:
		return e.writeTrendAnalysisToCSV(writer, r)
	case *types.AnomalyReport:
		return e.writeAnomalyReportToCSV(writer, r)
	case *types.ForecastReport:
		return e.writeForecastReportToCSV(writer, r)
	case *types.ExecutiveSummaryReport:
		return e.writeExecutiveSummaryToCSV(writer, r)
	default:
		return e.writeGenericToCSV(writer, report)
	}
}

// writeUsageSummaryToCSV writes usage summary report to CSV
func (e *CSVExporter) writeUsageSummaryToCSV(writer *csv.Writer, report *types.UsageSummaryReport) error {
	// Header
	if err := writer.Write([]string{
		"ID", "Title", "Generated At", "Resources Count", "Services Count", "Trends Count", "Capacity Items",
	}); err != nil {
		return err
	}
	// Row
	return writer.Write([]string{
		report.ID,
		report.Title,
		report.GeneratedAt.Format("2006-01-02 15:04:05"),
		strconv.Itoa(len(report.ResourceUtilization)),
		strconv.Itoa(len(report.ServiceUsage)),
		strconv.Itoa(len(report.UsageTrends)),
		strconv.Itoa(len(report.CapacityAnalysis)),
	})
}

// writeCostAnalysisToCSV writes cost analysis report to CSV
func (e *CSVExporter) writeCostAnalysisToCSV(writer *csv.Writer, report *types.CostAnalysisReport) error {
	// Write header
	if err := writer.Write([]string{
		"ID", "Title", "Generated At", "Total Cost", "Currency",
		"Cost By Service Count", "Cost By Region Count", "Top Cost Drivers Count",
	}); err != nil {
		return err
	}

	// Write data
	return writer.Write([]string{
		report.ID,
		report.Title,
		report.GeneratedAt.Format("2006-01-02 15:04:05"),
		strconv.FormatFloat(report.TotalCost, 'f', 2, 64),
		report.Currency,
	})
}

// writeTrendAnalysisToCSV writes trend analysis report to CSV
func (e *CSVExporter) writeTrendAnalysisToCSV(writer *csv.Writer, report *types.TrendAnalysisReport) error {
	// Write header
	if err := writer.Write([]string{
		"ID", "Title", "Generated At", "Trends Count", "Forecasts Count", "ML Insights Count",
	}); err != nil {
		return err
	}

	// Write data
	return writer.Write([]string{
		report.ID,
		report.Title,
		report.GeneratedAt.Format("2006-01-02 15:04:05"),
		strconv.Itoa(len(report.Trends)),
		strconv.Itoa(len(report.Forecasts)),
		strconv.Itoa(len(report.MLInsights)),
	})
}

// writeAnomalyReportToCSV writes anomaly report to CSV
func (e *CSVExporter) writeAnomalyReportToCSV(writer *csv.Writer, report *types.AnomalyReport) error {
	// Write header
	if err := writer.Write([]string{
		"ID", "Title", "Generated At", "Anomalies Count", "Alerts Count", "Risk Level",
	}); err != nil {
		return err
	}

	// Write data
	return writer.Write([]string{
		report.ID,
		report.Title,
		report.GeneratedAt.Format("2006-01-02 15:04:05"),
		strconv.Itoa(len(report.Anomalies)),
		strconv.Itoa(len(report.Alerts)),
		report.RiskLevel,
	})
}

// writeForecastReportToCSV writes forecast report to CSV
func (e *CSVExporter) writeForecastReportToCSV(writer *csv.Writer, report *types.ForecastReport) error {
	// Write header
	if err := writer.Write([]string{
		"ID", "Title", "Generated At", "Forecasts Count", "Scenarios Count", "Confidence",
	}); err != nil {
		return err
	}

	// Write data
	return writer.Write([]string{
		report.ID,
		report.Title,
		report.GeneratedAt.Format("2006-01-02 15:04:05"),
		strconv.Itoa(len(report.Forecasts)),
		strconv.Itoa(len(report.Scenarios)),
		strconv.FormatFloat(report.Confidence, 'f', 2, 64),
	})
}

// writeExecutiveSummaryToCSV writes executive summary report to CSV
func (e *CSVExporter) writeExecutiveSummaryToCSV(writer *csv.Writer, report *types.ExecutiveSummaryReport) error {
	// Write header
	if err := writer.Write([]string{
		"ID", "Title", "Generated At", "Total Spend", "Spend Change", "Top Cost Driver",
		"Optimization Opportunity", "Risk Level", "Recommendations Count",
	}); err != nil {
		return err
	}

	// Write data
	return writer.Write([]string{
		report.ID,
		report.Title,
		report.GeneratedAt.Format("2006-01-02 15:04:05"),
		strconv.FormatFloat(report.ExecutiveSummary.TotalSpend, 'f', 2, 64),
		strconv.FormatFloat(report.ExecutiveSummary.SpendChange, 'f', 2, 64),
		report.ExecutiveSummary.TopCostDriver,
		strconv.FormatFloat(report.ExecutiveSummary.OptimizationOpp, 'f', 2, 64),
		report.ExecutiveSummary.RiskLevel,
		strconv.Itoa(report.ExecutiveSummary.Recommendations),
	})
}

// writeGenericToCSV writes generic data to CSV using reflection
func (e *CSVExporter) writeGenericToCSV(writer *csv.Writer, data interface{}) error {
	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return fmt.Errorf("unsupported data type for CSV export: %T", data)
	}

	t := v.Type()

	// Write header
	var headers []string
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		if field.IsExported() {
			headers = append(headers, field.Name)
		}
	}

	if err := writer.Write(headers); err != nil {
		return err
	}

	// Write data
	var values []string
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		if field.IsExported() {
			fieldValue := v.Field(i)
			values = append(values, e.formatValue(fieldValue))
		}
	}

	return writer.Write(values)
}

// formatValue formats a reflect.Value to string for CSV
func (e *CSVExporter) formatValue(v reflect.Value) string {
	switch v.Kind() {
	case reflect.String:
		return v.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(v.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(v.Float(), 'f', 2, 64)
	case reflect.Bool:
		return strconv.FormatBool(v.Bool())
	case reflect.Slice:
		var items []string
		for i := 0; i < v.Len(); i++ {
			items = append(items, e.formatValue(v.Index(i)))
		}
		return strings.Join(items, ";")
	case reflect.Map:
		return fmt.Sprintf("map[%d items]", v.Len())
	case reflect.Struct:
		return fmt.Sprintf("struct[%s]", v.Type().Name())
	case reflect.Ptr:
		if v.IsNil() {
			return ""
		}
		return e.formatValue(v.Elem())
	default:
		return fmt.Sprintf("%v", v.Interface())
	}
}
