# mk/perf.mk - performance & parity related targets

.PHONY: bench data-parity-guard invariants-guard invariants-update-baseline parity-check parity-json perf-bench perf-bench-full perf-bench-synth perf-bench-update-baseline perf-gen-synth perf-guard perf-parity perf-short

bench: ## Run benchmarks
	@echo " Running benchmarks..."
	go test -bench=. -benchmem ./...

perf-guard: ## Run targeted legacy vs unified benchmark guard (M11)
	@echo "️  Running perf guard (legacy vs unified mapper)..."
	bash scripts/perf/perf-guard.sh || (echo " Perf guard failed" && exit 1)
	@echo " Perf guard passed"

perf-bench: ## Run legacy vs unified performance regression guard (TASK-PERF-BENCH)
	@echo " Running performance regression guard (legacy vs unified)..."
	# Compute GO_OUT_ARGS at runtime to avoid make conditional directives being interpreted
	# as shell commands in some execution environments (act / sh wrappers).
	@if [ -n "$(GITHUB_WORKSPACE)" ]; then \
		OUT="-output \"$(GITHUB_WORKSPACE)/bench_results.json\""; \
	else \
		OUT="-output bench_results.json"; \
	fi; \
	go run ./scripts/tools/perf-bench -input demo/focus-conversion/demo-cur-data.csv $$OUT $(EXTRA_ARGS) || (echo " Performance regression detected" && exit 1)
	@echo " Performance benchmark complete (see bench_results.json)"

perf-gen-synth: ## Generate synthetic AWS CUR-like dataset for perf tests
	@echo " Generating synthetic dataset (20k rows)..."
	go run ./scripts/tools/perf-gen-synth -rows 20000 -out tests/perf/aws-cur-synth.csv.gz
	@echo " Synthetic dataset ready: tests/perf/aws-cur-synth.csv.gz"

perf-bench-synth: perf-gen-synth ## Run perf bench against synthetic dataset
	@echo " Running perf bench on synthetic dataset..."
			# Call the stable wrapper script (avoids complex heredoc quoting issues under various shells)
			@bash scripts/tools/run_perf_wrapper.sh tests/perf/aws-cur-synth.csv.gz -iterations 3 $(EXTRA_ARGS)
	@echo " Perf bench complete (see bench_results.json, perf_metrics.prom)"

perf-bench-full: ## Run expanded perf bench (demo or synthetic) with unified regression guard script
	@echo " Running expanded performance benchmark (legacy vs unified mapper)..."
	bash scripts/perf/run_bench.sh -i ${INPUT:-demo/focus-conversion/demo-cur-data.csv} -o bench_results.json -p perf_metrics.prom || (echo " Performance regression detected" && exit 1)
	@echo " Expanded performance benchmark passed"

perf-bench-update-baseline: perf-gen-synth ## Regenerate baseline guard JSON for perf bench
	@echo " Regenerating performance baseline (5 iterations)..."
	PERF_BENCH_DURATION_MAX=$${PERF_BENCH_DURATION_MAX:-1.15} \
	PERF_BENCH_ALLOC_MAX=$${PERF_BENCH_ALLOC_MAX:-1.20} \
	go run ./scripts/tools/perf-bench -input tests/perf/aws-cur-synth.csv.gz -iterations 5 -output tests/perf/baseline_bench_results.json || (echo " Baseline regeneration failed" && exit 1)
	@echo " Baseline updated: tests/perf/baseline_bench_results.json"

perf-short: ## Run short perf bench (3 iterations) on synthetic dataset (if present)
	@echo " Running short perf bench (3 iterations)..."
	@if [ ! -f tests/perf/aws-cur-synth.csv.gz ]; then $(MAKE) perf-gen-synth; fi
	go run ./scripts/tools/perf-bench -input tests/perf/aws-cur-synth.csv.gz -iterations 3 -output bench_results.json $(EXTRA_ARGS) || (echo " Short perf bench failed" && exit 1)
	@echo " Short perf bench complete"

parity-check: build-slim ## Generate legacy & unified parquet outputs and compare aggregate parity
	@echo " Using slim binary for parquet conversion (no CGO)"
	BIN=bin/costscope; OUT_DIR=$$(pwd)/.cache/parquet; \
	 mkdir -p $$OUT_DIR; \
	 echo " Generating legacy parquet output..."; \
	 $$BIN convert --provider aws --input tests/perf/aws-cur-synth.csv.gz --output $$OUT_DIR/aws_legacy.parquet --streaming --rotate-size 10000000000 || { echo " legacy convert failed"; exit 1; }; \
	 echo " Generating unified parquet output..."; \
	 COSTSCOPE_USE_UNIFIED_MAPPER=true $$BIN convert --provider aws --input tests/perf/aws-cur-synth.csv.gz --output $$OUT_DIR/aws_unified.parquet --streaming --rotate-size 10000000000 || { echo " unified convert failed"; exit 1; }; \
	 echo " Comparing aggregate parity (effective_cost, usage_quantity, records)..."; \
	 go run ./scripts/tools/parity-check --legacy $$OUT_DIR/aws_legacy.parquet --unified $$OUT_DIR/aws_unified.parquet || { echo " Parity mismatch"; exit 1; }; \
	 echo " Parity aggregates match"

perf-parity: perf-short parity-check ## Run short perf bench then parity check (aggregates)
	@echo " Perf + Parity sequence complete"

parity-json: ## Generate fast & unified parquet outputs and write parity.json (fails on mismatch)
	@echo " Converting (fast path) → focus_fast.parquet"
	bin/costscope convert --provider aws --input tests/perf/aws-cur-synth.csv.gz --output focus_fast.parquet --streaming --rotate-size 10000000000 >/dev/null 2>&1 || { echo " fast convert failed"; exit 1; }
	@echo " Converting (unified mapper) → focus_unified.parquet"
	COSTSCOPE_USE_UNIFIED_MAPPER=1 bin/costscope convert --provider aws --input tests/perf/aws-cur-synth.csv.gz --output focus_unified.parquet --streaming --rotate-size 10000000000 >/dev/null 2>&1 || { echo " unified convert failed"; exit 1; }
	@echo " Running parity-check (lite hash enabled)"
	go run ./scripts/tools/parity-check --legacy focus_fast.parquet -unified focus_unified.parquet -tolerance $${PARITY_TOLERANCE:-1e-9} -out parity.json || { echo " Parity mismatch"; exit 2; }
	@echo " Parity guard passed (parity.json)"

invariants-guard: build-optimized-duckdb build-duckdb-debug ## Run invariants drift guard against baseline
	@echo " Using DuckDB binary for invariants guard"
	DBG=bin/costscope-duckdb-debug; OPT=bin/costscope-optimized-duckdb; BIN=$$DBG; BASE=$${INVARIANTS_BASELINE:-tests/fixtures/quality/baseline_invariants.json}; \
	if [ ! -f $$BASE ]; then echo " Baseline not found: $$BASE"; exit 1; fi; \
	echo " Converting (fast path only – invariants via regenerate) vs baseline=$$BASE (preferred debug=$$BIN)"; \
	if ! $$BIN convert --provider aws --input tests/perf/aws-cur-synth.csv.gz --output focus_fast.parquet --streaming --rotate-size 10000000000 >/dev/null 2>&1; then \
	  echo "️  Debug convert failed; trying optimized binary"; BIN=$$OPT; \
	  if ! $$BIN convert --provider aws --input tests/perf/aws-cur-synth.csv.gz --output focus_fast.parquet --streaming --rotate-size 10000000000 >/dev/null 2>&1; then \
	    echo " Conversion failed with both binaries"; exit 2; \
	  fi; \
	fi; \
	LATEST=$$(ls -1t focus_fast*.parquet 2>/dev/null | head -n1); [ -n "$$LATEST" ] || { echo " No parquet output"; exit 2; }; \
	echo " Regenerating current invariants from $$LATEST"; \
	if ! $$BIN invariants regenerate "$$LATEST" --output invariants_current.json --force --tolerance $${INVARIANTS_TOLERANCE:-0.01} >/dev/null 2>&1; then \
	  echo "️  Regenerate failed with $$BIN; trying fallback binary"; ALT=$$( [ $$BIN = $$DBG ] && echo $$OPT || echo $$DBG ); \
	  if ! $$ALT invariants regenerate "$$LATEST" --output invariants_current.json --force --tolerance $${INVARIANTS_TOLERANCE:-0.01} >/dev/null 2>&1; then \
	    echo " Invariants regenerate failed with both binaries"; exit 2; \
	  fi; \
	fi; \
	echo " Diffing invariants (current vs baseline)"; \
	if ! $$BIN invariants diff invariants_current.json --baseline $$BASE --tolerance $${INVARIANTS_TOLERANCE:-0.01} --report invariants.json >/dev/null 2>&1; then \
	  echo " Invariants drift detected"; cat invariants.json | head -n120; exit 3; \
	fi; \
	echo " Invariants guard passed (no drift) via $$BIN"; echo $$BIN > invariants_engine.txt

data-parity-guard: parity-json invariants-guard ## Run combined fast/unified parity + invariants drift guards
	@echo "️ Data parity + invariants guard complete"

invariants-update-baseline: build-optimized-duckdb ## Recompute invariants baseline JSON from current fast path output
	@echo "️  Recomputing invariants baseline (synthetic dataset)"
	BASE=$${INVARIANTS_BASELINE:-tests/fixtures/quality/baseline_invariants.json}; BIN=bin/costscope-optimized-duckdb; DBG=bin/costscope-duckdb-debug; \
	echo " Converting (fast path) with large rotate-size to minimize rotation"; \
	bin/costscope convert --provider aws --input tests/perf/aws-cur-synth.csv.gz --output focus_fast.parquet --streaming --rotate-size 10000000000 >/dev/null 2>&1 || { echo " fast convert failed"; exit 1; }; \
	LATEST=$$(ls -1t focus_fast*.parquet 2>/dev/null | head -n1); [ -n "$$LATEST" ] || { echo " No focus_fast parquet produced"; exit 2; }; \
	echo " Regenerating baseline from $$LATEST (preferred binary: $$BIN)"; \
	if $$BIN invariants regenerate "$$LATEST" --output $$BASE --force --tolerance $${INVARIANTS_TOLERANCE:-0.01} > /tmp/inv_regenerate.log 2>&1; then \
	  echo " Baseline regenerate succeeded with optimized binary"; \
	else \
	  echo "️  Optimized binary failed, retrying with debug binary"; \
	  if $$DBG invariants regenerate "$$LATEST" --output $$BASE --force --tolerance $${INVARIANTS_TOLERANCE:-0.01} >> /tmp/inv_regenerate.log 2>&1; then \
	    echo " Baseline regenerate succeeded with debug binary"; \
	  else \
	    echo " baseline regenerate failed (all binaries)"; cat /tmp/inv_regenerate.log; exit 2; \
	  fi; \
	fi; \
	grep -i "baseline" /tmp/inv_regenerate.log >/dev/null 2>&1 || true; \
	echo " Baseline updated: $$BASE (source=$$LATEST)"
