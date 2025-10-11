package gcp

import (
	"local/costscope/internal/core/focus/conversion/common"
	"local/costscope/internal/core/focus/types"
	"local/costscope/internal/core/monitoring/telemetry"
	"local/costscope/internal/framework/mapping"
)

// legacyCSVMapper uses FullRowMapper and increments classifier metrics per record.
func (g *GCPConverter) legacyCSVMapper(pathLabel string) func([]string, [][]string) ([]types.FocusRecord, int) {
	return func(headers []string, rows [][]string) ([]types.FocusRecord, int) { //nolint:cyclop
		mapper := NewFullRowMapper(headers)
		out := make([]types.FocusRecord, 0, len(rows))
		errs := 0
		for _, r := range rows {
			fr, _, _, err := mapper.Map(r)
			if err != nil {
				errs++
				continue
			}
			telemetry.ClassifierDecisions.WithLabelValues("gcp", pathLabel, fr.ChargeCategory).Inc()
			out = append(out, fr)
		}
		if n := len(out); n > 0 {
			telemetry.MapperRowsTotal.WithLabelValues("gcp", pathLabel).Add(float64(n))
		}
		return out, errs
	}
}

// unifiedCSVMapper maps CSV using FieldMapper + provider post-map enrichment, then emits metrics.
func (g *GCPConverter) unifiedCSVMapper(config *types.ConversionConfig, fm *mapping.FieldMapper, pathLabel string) func([]string, [][]string) ([]types.FocusRecord, int) {
	return func(headers []string, chunk [][]string) ([]types.FocusRecord, int) { //nolint:cyclop
		// Build index once
		idx := make(map[int]string, len(headers))
		for i, h := range headers {
			idx[i] = h
		}
		out := make([]types.FocusRecord, 0, len(chunk))
		errs := 0
		for _, rec := range chunk {
			// Build object for unified mapper
			obj := make(map[string]interface{}, len(rec))
			for i, v := range rec {
				if name, ok := idx[i]; ok {
					obj[name] = v
				}
			}
			fr, merr := fm.MapToFOCUS(obj)
			if merr != nil {
				errs++
				continue
			}
			// Provider-specific unified post-map + enrichment
			ApplyUnifiedPostMapGCP(fr, obj)
			EnsureUsageUnit(fr, obj)
			EnrichUnified(fr, obj)
			// Unified enum/region/unit normalization (ensures canonicalization + cache metrics)
			common.ApplyUnifiedNormalization("gcp", fr)

			telemetry.ClassifierDecisions.WithLabelValues("gcp", pathLabel, fr.ChargeCategory).Inc()
			out = append(out, *fr)
		}
		if n := len(out); n > 0 {
			telemetry.MapperRowsTotal.WithLabelValues("gcp", pathLabel).Add(float64(n))
		}
		// SourceFileName is added by ProcessCSV before write
		_ = config // keep parity with Azure mapper signature
		return out, errs
	}
}

// mapObjectLegacy builds a JSON object mapper using FullJSONMapper with enrichment parity.
func (g *GCPConverter) mapObjectLegacy(_ *types.ConversionConfig) func(map[string]interface{}) *types.FocusRecord { //nolint:unparam
	jm := NewFullJSONMapper()
	return func(obj map[string]interface{}) *types.FocusRecord {
		fr, _, _ := jm.Map(obj)
		// FullJSONMapper already classifies and fills fields sufficiently for legacy parity.
		return fr
	}
}

// mapObjectUnified builds a JSON object mapper using FieldMapper and provider post-map + enrichment.
func (g *GCPConverter) mapObjectUnified(fm *mapping.FieldMapper, _ *types.ConversionConfig) func(map[string]interface{}) *types.FocusRecord { //nolint:unparam
	return func(obj map[string]interface{}) *types.FocusRecord {
		fr, err := fm.MapToFOCUS(obj)
		if err != nil {
			// In unified path we skip errors; caller counts errorRecords via ProcessJSON when mapping fails (here we just return an empty record to be ignored).
			// However ProcessJSON expects a non-nil FocusRecord, so create minimal with default Usage category.
			fr = &types.FocusRecord{ChargeCategory: types.ChargeCategories.Usage}
		}
		ApplyUnifiedPostMapGCP(fr, obj)
		EnsureUsageUnit(fr, obj)
		EnrichUnified(fr, obj)
		// Unified enum/region/unit normalization (ensures canonicalization + cache metrics)
		common.ApplyUnifiedNormalization("gcp", fr)
		return fr
	}
}

// wrapJSONMapper wraps a JSON mapper to emit classifier metrics per record and return the record.
func (g *GCPConverter) wrapJSONMapper(pathLabel string, inner func(map[string]interface{}) *types.FocusRecord) func(map[string]interface{}) *types.FocusRecord {
	return func(obj map[string]interface{}) *types.FocusRecord {
		fr := inner(obj)
		if fr != nil {
			telemetry.ClassifierDecisions.WithLabelValues("gcp", pathLabel, fr.ChargeCategory).Inc()
		}
		return fr
	}
}

// ensure import of common for lint (used via alias)
var _ = common.FirstNonEmpty
