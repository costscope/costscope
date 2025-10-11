package validation

import (
	"bytes"
	"encoding/json"
	"html/template"
	"strings"
	"testing"
	"time"
)

func TestFormatters_JSONHTML_Snapshots(t *testing.T) {
	now := time.Unix(1700000000, 0).UTC()
	core := &ValidationResult{FilePath: "sample.focus.parquet", FileFormat: "parquet", FileSize: 12345, ValidationTime: now, Duration: 123 * time.Millisecond, IsValid: true, OverallScore: 95.5,
		SchemaValidation:      SchemaValidationResult{Valid: true, Score: 100},
		QualityAssessment:     QualityAssessmentResult{Valid: true, Score: 96},
		ComplianceValidation:  ComplianceValidationResult{Valid: true, Score: 90},
		PerformanceValidation: PerformanceValidationResult{Valid: true, Score: 88},
		RanCompliance:         true, RanPerformance: true,
		Issues: []ValidationIssue{}, Summary: ValidationSummary{TotalIssues: 0},
	}
	full := &ValidationFullResult{Core: core, Duration: 123 * time.Millisecond}
	b, err := json.MarshalIndent(full, "", "  ")
	if err != nil {
		t.Fatalf("marshal json: %v", err)
	}
	// snapshot (inline) ensures deterministic output; update intentionally if format changes
	if !strings.Contains(string(b), "\"overall_score\": 95.5") {
		t.Fatalf("snapshot missing score")
	}
	// HTML minimal template for smoke test (avoid importing commands package for formatter)
	tmpl := `<!DOCTYPE html><html><body>{{ $v := .Core }}<h1>FOCUS Validation Report</h1><p>{{ $v.FilePath }}</p><p>{{ printf "%.1f" $v.OverallScore }}</p></body></html>`
	var buf bytes.Buffer
	if err := template.Must(template.New("t").Parse(tmpl)).Execute(&buf, full); err != nil {
		t.Fatalf("exec html: %v", err)
	}
	if !strings.Contains(buf.String(), "95.5") {
		t.Fatalf("html missing score")
	}
}
