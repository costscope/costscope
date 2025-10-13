//go:build cgo
// +build cgo

package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/costscope/costscope/internal/core/logging"

	_ "github.com/marcboeker/go-duckdb"
)

type aggregates struct {
	EffectiveCost float64 `json:"effective_cost_sum"`
	UsageQuantity float64 `json:"usage_quantity_sum"`
	Records       int64   `json:"record_count"`
	Path          string  `json:"path"`
}

type parityResult struct {
	Legacy          aggregates `json:"legacy"`
	Unified         aggregates `json:"unified"`
	EqualCost       bool       `json:"equal_cost"`
	EqualUsage      bool       `json:"equal_usage"`
	EqualRecords    bool       `json:"equal_records"`
	LiteHashLegacy  string     `json:"lite_hash_legacy,omitempty"`
	LiteHashUnified string     `json:"lite_hash_unified,omitempty"`
	EqualLiteHash   bool       `json:"equal_lite_hash"`
	DurationMs      int64      `json:"duration_ms"`
	Tolerance       float64    `json:"tolerance"`
	Timestamp       string     `json:"timestamp"`
}

func fetchAgg(ctx context.Context, db *sql.DB, path string) (aggregates, error) {
	if _, err := os.Stat(path); err != nil {
		return aggregates{}, fmt.Errorf("stat %s: %w", path, err)
	}
	agg := aggregates{Path: path}
	row := db.QueryRowContext(ctx, `SELECT sum(effective_cost)::DOUBLE, sum(usage_quantity)::DOUBLE, count(*)::BIGINT FROM read_parquet(?)`, path)
	if err := row.Scan(&agg.EffectiveCost, &agg.UsageQuantity, &agg.Records); err != nil {
		return agg, err
	}
	return agg, nil
}

func relEq(a, b, tol float64) bool {
	if a == b {
		return true
	}
	diff := a - b
	if diff < 0 {
		diff = -diff
	}
	denom := b
	if denom < 0 {
		denom = -denom
	}
	if denom == 0 {
		return diff < tol
	}
	return diff/denom <= tol
}

func latestRotated(base string) (string, error) {
	// Try pattern when caller passed a base without .parquet (e.g. "focus_fast") -> "focus_fast-*.parquet"
	matches, err := filepath.Glob(base + "-*.parquet")
	if err != nil {
		return "", err
	}
	if len(matches) > 0 {
		if s, ok := selectBestFromMatches(matches); ok {
			return s, nil
		}
	}

	// If caller passed a name that includes the .parquet extension (e.g. "focus_fast.parquet"),
	// try the rotated pattern using the name without extension: "focus_fast-*.parquet".
	if ext := filepath.Ext(base); ext != "" {
		without := strings.TrimSuffix(base, ext)
		matches, err = filepath.Glob(without + "-*.parquet")
		if err != nil {
			return "", err
		}
		if len(matches) > 0 {
			if s, ok := selectBestFromMatches(matches); ok {
				return s, nil
			}
		}
	}

	// Finally, accept a non-rotated exact file if present.
	if _, err2 := os.Stat(base); err2 == nil {
		return base, nil
	}
	// Also try adding .parquet when caller supplied a base without extension and no rotated files were found.
	if filepath.Ext(base) == "" {
		if _, err3 := os.Stat(base + ".parquet"); err3 == nil {
			return base + ".parquet", nil
		}
	}

	return "", fmt.Errorf("no parquet files found matching rotation pattern for input %q", base)
}

// selectBestFromMatches picks the most recent rotated parquet path from a list of matches.
// Returns (path, true) when a candidate exists, otherwise ("", false).
func selectBestFromMatches(matches []string) (string, bool) {
	// pattern: <prefix>-YYYYMMDD-HHMM-###.parquet
	rotatedRe := regexp.MustCompile(`^(.+)-(\d{8})-(\d{4})-(\d{3})\.parquet$`)
	type candidate struct {
		path string
		ts   time.Time
		seq  int
		ok   bool
	}
	var cand []candidate
	for _, p := range matches {
		name := filepath.Base(p)
		m := rotatedRe.FindStringSubmatch(name)
		if m != nil {
			tsStr := m[2] + m[3]
			if t, err := time.Parse("200601021504", tsStr); err == nil {
				if s, err := strconv.Atoi(m[4]); err == nil {
					cand = append(cand, candidate{path: p, ts: t, seq: s, ok: true})
					continue
				}
			}
		}
		cand = append(cand, candidate{path: p, ok: false})
	}
	var best *candidate
	for i := range cand {
		c := &cand[i]
		if best == nil {
			best = c
			continue
		}
		if best.ok && !c.ok {
			continue
		}
		if !best.ok && c.ok {
			best = c
			continue
		}
		if !best.ok && !c.ok {
			if c.path > best.path {
				best = c
			}
			continue
		}
		if c.ts.After(best.ts) {
			best = c
		} else if c.ts.Equal(best.ts) && c.seq > best.seq {
			best = c
		} else if c.ts.Equal(best.ts) && c.seq == best.seq && c.path > best.path {
			best = c
		}
	}
	if best == nil {
		return "", false
	}
	return best.path, true
}

func main() {
	logger := logging.GetLogger().WithFields(map[string]interface{}{"tool": "parity-check"})
	legacy := flag.String("legacy", "", "Path (prefix if rotation) to legacy parquet output (e.g. /tmp/aws_legacy.parquet)")
	unified := flag.String("unified", "", "Path (prefix if rotation) to unified parquet output")
	out := flag.String("out", "", "Optional path to write JSON parity report")
	tolerance := flag.Float64("tolerance", 1e-9, "Relative tolerance for float equality (cost & usage)")
	includeLiteHash := flag.Bool("include-lite-hash", true, "Compute lite parity hash over key fields (may add minor overhead)")
	flag.Parse()
	if *legacy == "" || *unified == "" {
		logger.FatalWithFields("--legacy and --unified parquet file paths required", nil)
	}

	// Support passing the non-rotated base name (tool will pick the most recent rotated file)
	legacyPath, err := latestRotated(*legacy)
	if err != nil {
		logger.FatalWithFields("resolve legacy path", map[string]interface{}{"error": err.Error(), "input": *legacy})
	}
	unifiedPath, err := latestRotated(*unified)
	if err != nil {
		logger.FatalWithFields("resolve unified path", map[string]interface{}{"error": err.Error(), "input": *unified})
	}

	start := time.Now()
	db, err := sql.Open("duckdb", ":memory:")
	if err != nil {
		logger.FatalWithFields("open duckdb", map[string]interface{}{"error": err.Error()})
	}
	defer func() { _ = db.Close() }()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	legacyAgg, err := fetchAgg(ctx, db, legacyPath)
	if err != nil {
		logger.FatalWithFields("legacy agg", map[string]interface{}{"error": err.Error()})
	}
	unifiedAgg, err := fetchAgg(ctx, db, unifiedPath)
	if err != nil {
		logger.FatalWithFields("unified agg", map[string]interface{}{"error": err.Error()})
	}

	res := parityResult{
		Legacy:       legacyAgg,
		Unified:      unifiedAgg,
		EqualCost:    relEq(legacyAgg.EffectiveCost, unifiedAgg.EffectiveCost, *tolerance),
		EqualUsage:   relEq(legacyAgg.UsageQuantity, unifiedAgg.UsageQuantity, *tolerance),
		EqualRecords: legacyAgg.Records == unifiedAgg.Records,
		DurationMs:   time.Since(start).Milliseconds(),
		Tolerance:    *tolerance,
		Timestamp:    time.Now().UTC().Format(time.RFC3339),
	}

	if *includeLiteHash {
		// Compute lite hashes order-independently via DuckDB to avoid loading full records into memory.
		lhLegacy, err := computeLiteHash(ctx, db, legacyPath)
		if err != nil {
			logger.WarnWithFields("lite hash legacy", map[string]interface{}{"error": err.Error()})
		} else {
			res.LiteHashLegacy = lhLegacy
		}
		lhUnified, err := computeLiteHash(ctx, db, unifiedPath)
		if err != nil {
			logger.WarnWithFields("lite hash unified", map[string]interface{}{"error": err.Error()})
		} else {
			res.LiteHashUnified = lhUnified
		}
		if res.LiteHashLegacy != "" && res.LiteHashUnified != "" {
			res.EqualLiteHash = (res.LiteHashLegacy == res.LiteHashUnified)
		}
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(res); err != nil {
		logger.FatalWithFields("encode", map[string]interface{}{"error": err.Error()})
	}
	if *out != "" {
		f, err := os.Create(*out)
		if err != nil {
			logger.FatalWithFields("create report", map[string]interface{}{"error": err.Error(), "path": *out})
		}
		if err := json.NewEncoder(f).Encode(res); err != nil {
			logger.FatalWithFields("write report", map[string]interface{}{"error": err.Error()})
		}
		if err := f.Close(); err != nil {
			logger.WarnWithFields("closing report file", map[string]interface{}{"error": err.Error()})
		}
	}
	if !res.EqualCost || !res.EqualUsage || !res.EqualRecords || (*includeLiteHash && !res.EqualLiteHash) {
		os.Exit(2)
	}
}

// computeLiteHash builds a stable hash over key parity fields (effective_cost, usage_quantity,
// provider_name, service_name, charge_category) without materializing all rows into a slice.
// It mirrors the formatting logic in conversion.HashFocusLite for parity consistency.
func computeLiteHash(ctx context.Context, db *sql.DB, path string) (string, error) {
	// Order-independent: fetch rows, build slice of strings, sort, then hash.
	// We pull into memory only required short strings.
	rows, err := db.QueryContext(ctx, `SELECT effective_cost, usage_quantity, coalesce(provider_name,''), coalesce(service_name,''), coalesce(charge_category,'') FROM read_parquet(?)`, path)
	if err != nil {
		return "", err
	}
	defer func() { _ = rows.Close() }()
	parts := make([]string, 0, 1024)
	for rows.Next() {
		var eff, usage float64
		var prov, svc, cat string
		if err := rows.Scan(&eff, &usage, &prov, &svc, &cat); err != nil {
			return "", err
		}
		parts = append(parts, fmt.Sprintf("%.6f|%.6f|%s|%s|%s", eff, usage, prov, svc, cat))
	}
	if err := rows.Err(); err != nil {
		return "", err
	}
	if len(parts) == 0 {
		return "", nil
	}
	sort.Strings(parts)
	h := sha256.New()
	for _, p := range parts {
		_, _ = h.Write([]byte(p))
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
