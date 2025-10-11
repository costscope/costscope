# mk/tools.mk - developer tooling & profiling

.PHONY: cpu-profile memory-profile size-analysis test-optimized tools-build tools-clean

.PHONY: lint-tools-install
lint-tools-install: ## Install pinned lint/security/analysis tools for local development
	@./scripts/install-lint-tools.sh

tools-build: ## Build all development tools
	@echo " Building development tools..."
	go build $(LDFLAGS) -o bin/code-optimizer ./scripts/tools/code-optimizer/
	go build $(LDFLAGS) -o bin/gc-test ./scripts/tools/gc-test/
	go build $(LDFLAGS) -o bin/performance-profiler ./scripts/tools/performance-profiler/
	go build $(LDFLAGS) -o bin/gc-benchmark ./scripts/tools/gc-benchmark/
	@echo " Tools built in bin/ directory"

tools-clean: ## Clean development tools binaries
	@echo " Cleaning development tools..."; rm -f bin/code-optimizer bin/gc-test bin/performance-profiler bin/gc-benchmark

memory-profile: ## Run memory profiling with benchmarks
	@echo " Running memory profiling..."; go test -memprofile=memory.prof -bench=. ./...; echo " Use: go tool pprof memory.prof"

cpu-profile: ## Run CPU profiling with benchmarks
	@echo " Running CPU profiling..."; go test -cpuprofile=cpu.prof -bench=. ./...; echo " Use: go tool pprof cpu.prof"

size-analysis: ## Analyze binary size and symbols
	@echo " Analyzing binary size..."; go build $(LDFLAGS) -o bin/analysis ./; echo "Binary size: $$(ls -lh bin/analysis | awk '{print $$5}')"; echo "Top 20 largest symbols:"; nm -S bin/analysis 2>/dev/null | sort -rn | head -20 || echo "Symbol analysis not available"; rm -f bin/analysis

test-optimized: ## Run tests with optimization flags
	@echo " Running optimized tests..."; go test -ldflags="-w -s" -gcflags="-l=4" ./...
