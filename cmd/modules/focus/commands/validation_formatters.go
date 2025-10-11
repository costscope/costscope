package commands

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"html/template"
	"local/costscope/internal/core/focus/validation"
	"path/filepath"
	"strings"
)

const (
	formatterNameJSON  = "json"
	formatterNameHTML  = "html"
	formatterNameCSV   = "csv"
	formatterNameTable = "table"
)

type Formatter interface {
	Format(*validation.ValidationFullResult) ([]byte, error)
	Name() string
}

type TableFormatter struct{}

func (TableFormatter) Name() string { return formatterNameTable }
func (TableFormatter) Format(r *validation.ValidationFullResult) ([]byte, error) {
	vr := r.Core
	b := &bytes.Buffer{}
	fmt.Fprintf(b, "\n═══════════════════════════════════════════════════════════════\n")
	fmt.Fprintf(b, "                    FOCUS VALIDATION REPORT\n")
	fmt.Fprintf(b, "═══════════════════════════════════════════════════════════════\n")
	fmt.Fprintf(b, " File: %s\n File Size: %.2f MB\n⏱️  Duration: %v\n Timestamp: %s\n\n", vr.FilePath, float64(vr.FileSize)/(1024*1024), r.Duration, vr.ValidationTime.Format("2006-01-02 15:04:05"))
	icon, txt := "", "VALID"
	if !vr.IsValid {
		icon, txt = "", "INVALID"
	}
	fmt.Fprintf(b, " OVERALL STATUS: %s %s (Score: %.1f/100)\n\n", icon, txt, vr.OverallScore)
	fmt.Fprintf(b, " VALIDATION DOMAINS:\n─────────────────────────────────────────────────────────────\n")
	fmt.Fprintf(b, "   Schema Validation: %s (%.1f/100)\n", passFail(vr.SchemaValidation.Valid), vr.SchemaValidation.Score)
	fmt.Fprintf(b, "   Data Quality: %s (%.1f/100)\n", passFail(vr.QualityAssessment.Valid), vr.QualityAssessment.Score)
	if vr.RanCompliance {
		fmt.Fprintf(b, "  ️  Compliance Check: %s (%.1f/100)\n", passFail(vr.ComplianceValidation.Valid), vr.ComplianceValidation.Score)
	} else {
		fmt.Fprintf(b, "  ️  Compliance Check: ⏭️  SKIPPED\n")
	}
	if vr.RanPerformance {
		fmt.Fprintf(b, "   Performance Analysis: %s (%.1f/100)\n", passFail(vr.PerformanceValidation.Valid), vr.PerformanceValidation.Score)
	} else {
		fmt.Fprintf(b, "   Performance Analysis: ⏭️  SKIPPED\n")
	}
	if vr.RanAnomalies && vr.AnomalyDetection != nil {
		fmt.Fprintf(b, "   Anomaly Detection: %s (%.1f/100)\n", passFail(vr.AnomalyDetection.Valid), vr.AnomalyDetection.Score)
	} else if !vr.RanAnomalies {
		fmt.Fprintf(b, "   Anomaly Detection: ⏭️  SKIPPED\n")
	}
	fmt.Fprintf(b, "\n")
	if len(vr.Issues) > 0 {
		fmt.Fprintf(b, "️  ISSUES FOUND (%d):\n─────────────────────────────────────────────────────────────\n", len(vr.Issues))
		for i, is := range vr.Issues {
			fmt.Fprintf(b, "%d. %s: %s\n", i+1, strings.ToUpper(is.Type), is.Message)
			if is.Suggestion != "" {
				fmt.Fprintf(b, "    Suggestion: %s\n", is.Suggestion)
			}
		}
		fmt.Fprintf(b, "\n")
	}
	if r.Invariants != nil {
		fmt.Fprintf(b, " Invariants: RowCount=%d SumEffectiveCost=%.2f Violations=%d\n", r.Invariants.RowCount, r.Invariants.SumEffectiveCost, len(r.Invariants.Violations))
	}
	fmt.Fprintf(b, "═══════════════════════════════════════════════════════════════\n")
	return b.Bytes(), nil
}

func passFail(ok bool) string {
	if ok {
		return " PASSED"
	}
	return " FAILED"
}

type JSONFormatter struct{}

func (JSONFormatter) Name() string { return formatterNameJSON }
func (JSONFormatter) Format(r *validation.ValidationFullResult) ([]byte, error) {
	return json.MarshalIndent(r, "", "  ")
}

type HTMLFormatter struct{}

func (HTMLFormatter) Name() string { return formatterNameHTML }
func (HTMLFormatter) Format(r *validation.ValidationFullResult) ([]byte, error) {
	tmpl := `<!DOCTYPE html><html><head><meta charset="utf-8"><title>FOCUS Validation Report</title><style>body{font-family:Arial;margin:20px}.status-valid{color:green}.status-invalid{color:red}table{border-collapse:collapse;margin-top:16px}td,th{border:1px solid #ccc;padding:4px 8px}</style></head><body>{{ $v := .Core }}<h1>FOCUS Validation Report</h1><p><strong>File:</strong> {{ $v.FilePath }}</p><p><strong>Timestamp:</strong> {{ $v.ValidationTime.Format "2006-01-02 15:04:05" }}</p><p class="status-{{ if $v.IsValid }}valid{{ else }}invalid{{ end }}">Score: {{ printf "%.1f" $v.OverallScore }}/100</p><table><tr><th>Domain</th><th>Status</th><th>Score</th></tr><tr><td>Schema</td><td>{{ if $v.SchemaValidation.Valid }}PASSED{{ else }}FAILED{{ end }}</td><td>{{ printf "%.1f" $v.SchemaValidation.Score }}</td></tr><tr><td>Quality</td><td>{{ if $v.QualityAssessment.Valid }}PASSED{{ else }}FAILED{{ end }}</td><td>{{ printf "%.1f" $v.QualityAssessment.Score }}</td></tr><tr><td>Compliance</td><td>{{ if $v.RanCompliance }}{{ if $v.ComplianceValidation.Valid }}PASSED{{ else }}FAILED{{ end }}{{ else }}SKIPPED{{ end }}</td><td>{{ if $v.RanCompliance }}{{ printf "%.1f" $v.ComplianceValidation.Score }}{{ else }}-{{ end }}</td></tr></table>{{ if .Invariants }}<p>Invariants RowCount={{ .Invariants.RowCount }} Violations={{ len .Invariants.Violations }}</p>{{ end }}</body></html>`
	t, err := template.New("validation-html").Parse(tmpl)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, r); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

type CSVFormatter struct{}

func (CSVFormatter) Name() string { return formatterNameCSV }
func (CSVFormatter) Format(r *validation.ValidationFullResult) ([]byte, error) {
	buf := &bytes.Buffer{}
	w := csv.NewWriter(buf)
	_ = w.Write([]string{"file", "valid", "score", "timestamp", "issues"})
	vr := r.Core
	_ = w.Write([]string{vr.FilePath, fmt.Sprintf("%t", vr.IsValid), fmt.Sprintf("%.1f", vr.OverallScore), vr.ValidationTime.Format("2006-01-02 15:04:05"), fmt.Sprintf("%d", len(vr.Issues))})
	w.Flush()
	return buf.Bytes(), nil
}

func SelectFormatter(name, outputPath string) Formatter {
	if name == "" && outputPath != "" {
		switch strings.ToLower(filepath.Ext(outputPath)) {
		case ".json":
			name = formatterNameJSON
		case ".html":
			name = formatterNameHTML
		case ".csv":
			name = formatterNameCSV
		}
	}
	switch strings.ToLower(name) {
	case formatterNameJSON:
		return JSONFormatter{}
	case formatterNameHTML:
		return HTMLFormatter{}
	case formatterNameCSV:
		return CSVFormatter{}
	default:
		return TableFormatter{}
	}
}
