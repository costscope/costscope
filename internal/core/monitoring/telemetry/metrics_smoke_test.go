package telemetry

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func TestExporterMetricsInc(t *testing.T) {
	ExporterBytes.WithLabelValues("test-target", "json").Add(42)
	ExporterDuration.WithLabelValues("test-target", "json").Observe(0.001)
}

// TestMetricsExposure ensures that the minimal observability metrics are registered and exposed.
func TestMetricsExposure(t *testing.T) {
	// Register may panic if double-called; ignore subsequent calls in tests by recover
	defer func() { _ = recover() }()
	Register()
	// Force initial samples so histogram/counter appear
	ConversionDurationSimple.WithLabelValues("aws", "legacy").Observe(0)
	MapperRowsTotal.WithLabelValues("aws", "legacy").Add(0)
	// Ensure cache-related metrics are present
	NormalizationCacheHits.WithLabelValues("region", "any").Add(0)
	EnumCacheHits.WithLabelValues("unit", "any").Add(0)
	rr := httptest.NewRecorder()
	promhttp.Handler().ServeHTTP(rr, httptest.NewRequest("GET", "/metrics", nil))
	body := rr.Body.String()
	// Check for required metric names
	required := []string{
		"costscope_conversion_conversion_duration_seconds",
		"costscope_conversion_mapper_rows_total",
		"costscope_conversion_active_jobs",
		"costscope_normalization_cache_hits_total",
		"costscope_enum_cache_hits_total",
		"costscope_health_readiness",
	}
	for _, name := range required {
		present := strings.Contains(body, name+"_bucket") || strings.Contains(body, name+" ") || strings.Contains(body, name+"{")
		if !present {
			t.Logf("metrics body:\n%s", body)
			t.Fatalf("expected metric %s to be exposed", name)
		}
	}
}
