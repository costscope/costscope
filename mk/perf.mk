# mk/perf.mk - performance & parity related targets

# Parquet rotation size used during test conversions (can be overridden)
PARQUET_ROTATE_SIZE ?= 10000000000

.PHONY: bench data-parity-guard data-parity-smoke invariants-guard invariants-update-baseline parity-check parity-json parity-smoke perf-bench perf-bench-full perf-bench-synth perf-bench-update-baseline perf-gen-synth perf-gen-smoke perf-guard perf-parity perf-short

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

perf-gen-smoke: ## Generate small synthetic dataset for smoke guard (fast)
	@echo " Generating small synthetic dataset (50 rows)..."
	go run ./scripts/tools/perf-gen-synth -rows 50 -out tests/perf/aws-cur-smoke.csv.gz
	@echo " Smoke dataset ready: tests/perf/aws-cur-smoke.csv.gz"

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

perf-short: ## Run short perf bench (stable) on synthetic dataset (if present)
		@echo " Running short perf bench (stable iterations)..."
		@if [ ! -f tests/perf/aws-cur-synth.csv.gz ]; then $(MAKE) perf-gen-synth; fi
		# Configure warm-up and measured iterations with sensible defaults to reduce jitter
		@WARM=$${PERF_SHORT_WARMUP:-3}; MEAS=$${PERF_SHORT_ITERS:-7}; \
		DUR=$${PERF_SHORT_DURATION_MAX:-$${PERF_BENCH_DURATION_MAX:-1.15}}; \
		ALLOC=$${PERF_SHORT_ALLOC_MAX:-$${PERF_BENCH_ALLOC_MAX:-1.20}}; \
		echo "  Using warmup=$$WARM measured=$$MEAS thresholds: duration_max=$$DUR alloc_max=$$ALLOC"; \
		if [ "$$WARM" -gt 0 ]; then \
			for i in $$(seq 1 $$WARM); do \
				go run ./scripts/tools/perf-bench -input tests/perf/aws-cur-synth.csv.gz -iterations 1 -output /tmp/costscope_perfwarm.json >/dev/null 2>&1 || true; \
			done; \
		fi; \
		PERF_BENCH_DURATION_MAX=$$DUR PERF_BENCH_ALLOC_MAX=$$ALLOC \
			go run ./scripts/tools/perf-bench -input tests/perf/aws-cur-synth.csv.gz -iterations $$MEAS -output bench_results.json $(EXTRA_ARGS) || (echo " Short perf bench failed" && exit 1)
		@echo " Short perf bench complete"

parity-check: build-slim ## Generate legacy & unified parquet outputs and compare aggregate parity
	@echo " Using slim binary for parquet conversion (no CGO)"
	BIN=bin/costscope; OUT_DIR=$$(pwd)/.cache/parquet; \
	 mkdir -p $$OUT_DIR; \
	 echo " Generating legacy parquet output..."; \
	 $$BIN convert --provider aws --input tests/perf/aws-cur-synth.csv.gz --output $$OUT_DIR/aws_legacy.parquet --streaming --rotate-size $(PARQUET_ROTATE_SIZE) || { echo " legacy convert failed"; exit 1; }; \
	 echo " Generating unified parquet output..."; \
	 COSTSCOPE_USE_UNIFIED_MAPPER=true $$BIN convert --provider aws --input tests/perf/aws-cur-synth.csv.gz --output $$OUT_DIR/aws_unified.parquet --streaming --rotate-size $(PARQUET_ROTATE_SIZE) || { echo " unified convert failed"; exit 1; }; \
	 echo " Comparing aggregate parity (effective_cost, usage_quantity, records)..."; \
	 go run ./scripts/tools/parity-check --legacy $$OUT_DIR/aws_legacy.parquet --unified $$OUT_DIR/aws_unified.parquet || { echo " Parity mismatch"; exit 1; }; \
	 echo " Parity aggregates match"

perf-parity: perf-short parity-check ## Run short perf bench then parity check (aggregates)
	@echo " Perf + Parity sequence complete"

parity-json: prepare-parity-binaries ## Generate fast & unified parquet outputs and write parity.json (fails on mismatch)
	@bash scripts/guards/parity_guard.sh || exit $$?

parity-smoke: prepare-parity-binaries ## Run parity guard on small dataset (fast)
	@echo " Running parity smoke guard (small dataset)"
	@if [ ! -f tests/perf/aws-cur-smoke.csv.gz ]; then $(MAKE) perf-gen-smoke; fi
	INPUT=tests/perf/aws-cur-smoke.csv.gz bash scripts/guards/parity_guard.sh || exit $$?

invariants-guard: prepare-parity-binaries ## Run invariants drift guard against baseline
	@bash scripts/guards/invariants_guard.sh || exit $$?

data-parity-guard: parity-json invariants-guard ## Run combined fast/unified parity + invariants drift guards
	@echo "️ Data parity + invariants guard complete"

data-parity-smoke: parity-smoke ## Run fast smoke parity guard on small dataset
	@echo "️ Data parity smoke guard complete"

invariants-update-baseline: build-optimized-duckdb ## Recompute invariants baseline JSON from current fast path output
	@echo "️  Recomputing invariants baseline (synthetic dataset)"
	BASE=$${INVARIANTS_BASELINE:-tests/fixtures/quality/baseline_invariants.json}; BIN=bin/costscope-optimized-duckdb; DBG=bin/costscope-duckdb-debug; \
	echo " Converting (fast path) with rotate-size=$(PARQUET_ROTATE_SIZE)"; \
	bin/costscope convert --provider aws --input tests/perf/aws-cur-synth.csv.gz --output focus_fast.parquet --streaming --rotate-size $(PARQUET_ROTATE_SIZE) >/dev/null 2>&1 || { echo " fast convert failed"; exit 1; }; \
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
