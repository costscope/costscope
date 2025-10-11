package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"runtime/pprof"
	"time"

	awsc "local/costscope/internal/core/focus/conversion/aws"
	"local/costscope/internal/core/focus/types"
	"local/costscope/internal/core/logging"
)

func main() {
	input := flag.String("input", "demo/focus-conversion/demo-cur-data.csv", "AWS CUR CSV path (supports .gz)")
	output := flag.String("output", "tmp/perf-focus.parquet", "Output focus parquet path")
	chunk := flag.Int("chunk", 10000, "Streaming chunk size")
	cpu := flag.String("cpu", "cpu.prof", "CPU profile output file")
	mem := flag.String("mem", "mem.prof", "Heap profile output file")
	flag.Parse()

	logger := logging.GetLogger().WithFields(map[string]interface{}{"tool": "perf-convert-aws"})
	if err := os.MkdirAll("tmp", 0o750); err != nil {
		logger.FatalWithFields("mkdir tmp failed", map[string]interface{}{"error": err.Error()})
	}

	// CPU profile
	f, err := os.Create(*cpu)
	if err != nil {
		logger.FatalWithFields("create cpu profile failed", map[string]interface{}{"error": err.Error()})
	}
	if err := pprof.StartCPUProfile(f); err != nil {
		logger.FatalWithFields("start cpu profile failed", map[string]interface{}{"error": err.Error()})
	}

	ctx := context.Background()
	conv := awsc.NewAWSConverter()
	cfg := &types.ConversionConfig{
		Provider:     "aws",
		InputPath:    *input,
		OutputPath:   *output,
		Streaming:    true,
		ChunkSize:    *chunk,
		Workers:      1,
		Compression:  true,
		ConversionId: fmt.Sprintf("perf_%d", time.Now().Unix()),
	}

	res, err := conv.ConvertStream(ctx, cfg, nil)
	pprof.StopCPUProfile()
	_ = f.Close()
	if err != nil {
		logger.FatalWithFields("convert failed", map[string]interface{}{"error": err.Error()})
	}

	// Heap profile
	mf, err := os.Create(*mem)
	if err != nil {
		logger.FatalWithFields("create mem profile failed", map[string]interface{}{"error": err.Error()})
	}
	// runtime.GC() // optional: reflect peak usage by avoiding forced GC
	if err := pprof.WriteHeapProfile(mf); err != nil {
		logger.FatalWithFields("write heap profile failed", map[string]interface{}{"error": err.Error()})
	}
	_ = mf.Close()

	fmt.Printf("Done. Records: %d, duration: %v, rps: %.1f\n", res.OutputRecords, res.Duration, res.RecordsPerSecond)
	fmt.Println("Profiles generated:", *cpu, *mem)
}
