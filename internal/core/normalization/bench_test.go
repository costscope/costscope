package normalization

import (
	"testing"
)

func BenchmarkNormalizeRegionCacheHit(b *testing.B) {
	PreWarm() // ensure cache pre-populated
	// Choose a hot key
	for i := 0; i < b.N; i++ {
		if _, ok := NormalizeRegion("aws", "us-east-1"); !ok {
			b.Fatalf("unexpected miss")
		}
	}
}

func BenchmarkNormalizeRegionMiss(b *testing.B) {
	// Use value unlikely in dictionary; vary suffix to limit cache benefit
	for i := 0; i < b.N; i++ {
		_, _ = NormalizeRegion("aws", "unknown-region-"+string('a'+rune(i%26)))
	}
}

func BenchmarkNormalizeUnit(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NormalizeUnit("gib")
	}
}
