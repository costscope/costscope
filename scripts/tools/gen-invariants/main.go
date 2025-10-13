//go:build tools
// +build tools

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"path/filepath"
	"time"

	pipeline "github.com/costscope/costscope/tests/e2e/pipeline"
)

// Helper to print invariant metrics JSON for a provider's current test fixtures.
// Example:
//
//	go run ./scripts/tools/gen-invariants -provider azure > tests/fixtures/quality/baseline_azure_invariants.json
func main() {
	provider := flag.String("provider", "", "cloud provider: aws|azure|gcp")
	deterministic := flag.Bool("deterministic", false, "emit fixed generated_at timestamp for reproducible baselines")
	flag.Parse()
	if *provider == "" {
		log.Fatal("-provider required")
	}
	fixturesDir := filepath.Join("tests", "fixtures", *provider)
	var inputs []string
	switch *provider {
	case "azure":
		inputs = []string{
			filepath.Join(fixturesDir, "usage.csv"),
			filepath.Join(fixturesDir, "tax_refund.csv"),
			filepath.Join(fixturesDir, "reservation_credit.csv"),
		}
	case "gcp":
		inputs = []string{
			filepath.Join(fixturesDir, "usage_minimal.csv"),
			filepath.Join(fixturesDir, "credit_cud.csv"),
			filepath.Join(fixturesDir, "credit_spot.csv"),
			filepath.Join(fixturesDir, "credit_sustained_promo.csv"),
		}
	case "aws":
		inputs = []string{filepath.Join(fixturesDir, "cur_baseline_sample.csv")}
	default:
		log.Fatalf("unsupported provider: %s", *provider)
	}
	rep, err := pipeline.Run(context.Background(), pipeline.RunConfig{Provider: *provider, InputFiles: inputs, DriftTolerance: 0.001, ValidateOutput: true})
	if err != nil {
		log.Fatalf("run error: %v", err)
	}
	// Clear violations for baseline; embed metadata marker
	rep.Invariants.Violations = nil
	if rep.Invariants.Metadata == nil {
		rep.Invariants.Metadata = map[string]string{}
	}
	rep.Invariants.Metadata["baseline_source"] = "generated"
	if *deterministic {
		// Force generated_at to a stable value (Unix epoch) for CI diff stability
		rep.Invariants.GeneratedAt = time.Unix(0, 0).UTC()
	}
	b, _ := json.MarshalIndent(rep.Invariants, "", "  ")
	fmt.Println(string(b))
}
