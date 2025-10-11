package azure

import (
	"fmt"
	"time"

	"local/costscope/internal/core/focus/types"
)

// De-duplicated constants (kept local to avoid import cycles during migration)
const (
	providerAzure  = "azure"
	spotStr        = "spot"
	kwReservation  = "reservation"
	kwReserved     = "reserved"
	kwSavingsPlan1 = "savingsplan"
	kwSavingsPlan2 = "savings plan"
	kwCommitment   = "commitment"
	kwAmortized    = "amortized"
	liTypeDiscount = "Discount"
	tokenDiscount  = "discount"
	tokenUsageDisc = "usage-discount"
)

// NOTE: The former lite RowMapper (MapRaw) was removed to eliminate duplication.
// Parity tests now project a reduced view from FullRowMapper output (see PR note).

// FullRowMapper performs full FocusRecord mapping using header index and injected helpers.
type FullRowMapper struct {
	idx            *HeaderIndex
	classify       func(chargeType, billingType string, effectiveCost float64, billedCost *float64, candidateValues []string, provider string) string
	applyBenefits  func(idx *HeaderIndex, row []string, fr *types.FocusRecord)
	parseTags      func(s string) types.Tags
	ensureDiscount func(chargeType, billingType string, fr *types.FocusRecord) bool
	now            func() time.Time
	// new split components (wired from existing deps to preserve behavior)
	fm FieldMapper
	cf Classifier
	nz Normalizer
}

// NewFullRowMapperWithDeps creates a FullRowMapper with explicit dependencies supplied by the root package.
func NewFullRowMapperWithDeps(
	idx *HeaderIndex,
	classify func(chargeType, billingType string, effectiveCost float64, billedCost *float64, candidateValues []string, provider string) string,
	applyBenefits func(idx *HeaderIndex, row []string, fr *types.FocusRecord),
	parseTags func(s string) types.Tags,
	ensureDiscount func(chargeType, billingType string, fr *types.FocusRecord) bool,
	now func() time.Time,
) *FullRowMapper {
	if now == nil {
		now = time.Now
	}
	// wire the split components using the same deps to preserve behavior
	fm := newFieldMapper(idx, applyBenefits, parseTags, now)
	cf := newClassifier(idx, classify)
	nz := newNormalizer(idx)
	return &FullRowMapper{idx: idx, classify: classify, applyBenefits: applyBenefits, parseTags: parseTags, ensureDiscount: ensureDiscount, now: now, fm: fm, cf: cf, nz: nz}
}

// Map maps a CSV row to a full FocusRecord, matching legacy behavior.
func (m *FullRowMapper) Map(row []string) (types.FocusRecord, error) { //nolint:cyclop
	if m.idx == nil {
		return types.FocusRecord{}, fmt.Errorf("azure full row mapper: nil header index")
	}
	// 1) pure field mapping
	fr, err := m.fm.MapFields(row)
	if err != nil {
		return types.FocusRecord{}, err
	}
	// 2) classification (initial + fallback)
	if m.classify != nil {
		// use component to replicate embedded logic
		m.cf.Classify(row, &fr)
	} else {
		fr.ChargeCategory = types.ChargeCategories.Usage
	}
	// 3) discount normalization
	if m.ensureDiscount != nil {
		m.nz.Normalize(row, &fr)
	}
	return fr, nil
}

func rowValue(r []string, i int) string { return Get(r, i) }

func stringPtr(s string) *string { return &s }

// collectCandidateValuesAzure mirrors the legacy helper used in root conversion to keep parity.
func collectCandidateValuesAzure(row []string, primary ...interface{}) []string { // variadic for indices + slices
	out := []string{}
	for _, p := range primary {
		switch v := p.(type) {
		case int:
			if v >= 0 && v < len(row) {
				out = append(out, row[v])
			}
		case []int:
			for _, i := range v {
				if i >= 0 && i < len(row) {
					out = append(out, row[i])
				}
			}
		}
	}
	// Limit additional scan to first 20 columns to cap allocations; legacy path scanned whole row.
	limit := len(row)
	if limit > 20 {
		limit = 20
	}
	for i := 0; i < limit; i++ {
		out = append(out, row[i])
	}
	return out
}
