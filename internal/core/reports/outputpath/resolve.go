package outputpath

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"local/costscope/internal/core/config"
	"local/costscope/internal/core/config/precedence"
	"local/costscope/internal/core/logging"
	rtypes "local/costscope/internal/core/reports/types"
)

const defaultReportsDir = "costscope-data/reports"

var filenameSanitizer = regexp.MustCompile(`[^a-zA-Z0-9_.\-]`)

// ResolveOutputPath determines the final path for a report export.
// Precedence for selecting the base directory (reports.output_dir):
//  1. Explicit base argument (if non-empty)
//  2. YAML config (reports.output_dir) when loaded
//  3. Environment variable COSTSCOPE_REPORTS_DIR
//  4. Default constant (costscope-data/reports)
//
// A single structured log line `config_precedence_resolved` is emitted via precedence.LogResolved.
// When explicit output path is provided it is honored verbatim (extension auto-added if local path
// without extension); for object storage schemes (s3://, gs://, file://) no mkdir is attempted.
func ResolveOutputPath(base, explicit string, format rtypes.ExportFormat) (string, error) {
	ext := extensionForFormat(format)
	if explicit != "" {
		if !isObjectStorePath(explicit) && filepath.Ext(explicit) == "" {
			explicit += ext
		}
		if !isObjectStorePath(explicit) {
			if err := os.MkdirAll(filepath.Dir(explicit), 0o750); err != nil {
				return "", fmt.Errorf("create reports dir: %w", err)
			}
		}
		return explicit, nil
	}

	// Resolve base directory using unified precedence + single structured log line
	// field: reports.output_dir, order: explicit(base arg) > YAML > ENV > default
	logger := logging.GetLogger()
	var explicitPtr *string
	if base != "" {
		explicitPtr = &base
	}
	var yamlPtr *string
	if cfg := config.LoadOptionalYAML(logger); cfg != nil {
		yamlPtr = &cfg.Reports.OutputDir
	}
	res := precedence.ResolveString(explicitPtr, yamlPtr, "COSTSCOPE_REPORTS_DIR", defaultReportsDir)
	precedence.LogResolved(logger, "reports.output_dir", res)
	baseDir := res.Value

	ts := time.Now().UTC().Format("20060102-150405")
	filename := fmt.Sprintf("%s-report%s", ts, ext)
	filename = filenameSanitizer.ReplaceAllString(filename, "_")

	if isObjectStorePath(baseDir) {
		if !strings.HasSuffix(baseDir, "/") {
			baseDir += "/"
		}
		return baseDir + filename, nil
	}

	if err := os.MkdirAll(baseDir, 0o750); err != nil {
		return "", fmt.Errorf("create reports dir: %w", err)
	}
	return filepath.Join(baseDir, filename), nil
}

func extensionForFormat(format rtypes.ExportFormat) string {
	switch format {
	case rtypes.ExportFormatParquet:
		return ".parquet"
	case rtypes.ExportFormatJSON:
		return ".json"
	case rtypes.ExportFormatCSV:
		return ".csv"
	case rtypes.ExportFormatYAML:
		return ".yaml"
	case rtypes.ExportFormatPDF:
		return ".pdf"
	case rtypes.ExportFormatExcel:
		return ".xlsx"
	default:
		return ""
	}
}

func isObjectStorePath(p string) bool {
	return strings.HasPrefix(p, "s3://") || strings.HasPrefix(p, "gs://") || strings.HasPrefix(p, "file://")
}
