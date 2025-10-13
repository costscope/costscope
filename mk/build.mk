# mk/build.mk - build related targets

.PHONY: build build-clean build-duckdb-debug build-enterprise build-optimized build-optimized-duckdb build-production build-release build-slim duckdb-smoke

build: build-slim ## Default: build slim binary (no DuckDB/SQLite; CGO disabled)

build-slim: ## Build slim binary without DuckDB/Arrow/Thrift/SQLite (default)
	@echo " Building slim binary (no DuckDB/SQLite)..."
	@mkdir -p .cache/go-build .cache/go-tmp
	GOTMPDIR=$$(pwd)/.cache/go-tmp GOCACHE=$$(pwd)/.cache/go-build CGO_ENABLED=0 go build $(LDFLAGS) -o bin/costscope ./

build-clean: ## Build with clean cache to catch compilation issues (keeps strip/trimpath flags)
	@echo " Building with clean cache..."
	go clean -cache
	go build $(LDFLAGS) -o bin/costscope ./

build-release: ## Build reproducible release (stamps Version, Commit, BuildDate, GoVersion)
		@# Skip clean-repo enforcement under ACT or when emulated via nektos/act (GITHUB_ACTOR)
		@if [ "${IS_ACT:-}" = "true" ] || [ "${GITHUB_ACTOR:-}" = "nektos/act" ]; then \
				echo "[build-release] Skipping clean-repo check (ACT detected)"; \
			else \
				./scripts/check-clean-repo.sh; \
			fi
	@echo " Building CostScope (release)..."
	@VERSION_FLAG=$(shell git describe --tags --always --dirty 2>/dev/null || echo dev); \
	 COMMIT_FLAG=$(shell git rev-parse --short=12 HEAD 2>/dev/null || echo none); \
	 BUILD_DATE_FLAG=$(shell if [ -n "$$SOURCE_DATE_EPOCH" ]; then date -u -d @$$SOURCE_DATE_EPOCH +%Y-%m-%dT%H:%M:%SZ; else date -u +%Y-%m-%dT%H:%M:%SZ; fi); \
	 GO_VERSION_FLAG=$(shell go version | awk '{print $$3}'); \
	 CGO_ENABLED=0 GOOS=linux go build -trimpath -buildvcs=false -ldflags="-s -w -X main.Version=$$VERSION_FLAG -X main.Commit=$$COMMIT_FLAG -X main.BuildDate=$$BUILD_DATE_FLAG -X main.GoVersion=$$GO_VERSION_FLAG -X github.com/costscope/costscope/cmd.Version=$$VERSION_FLAG -X github.com/costscope/costscope/cmd.Commit=$$COMMIT_FLAG -X github.com/costscope/costscope/cmd.BuildDate=$$BUILD_DATE_FLAG -X github.com/costscope/costscope/cmd.GoVersion=$$GO_VERSION_FLAG" -o costscope ./; \
	 echo " Built release binary: costscope"; echo "Version: $$VERSION_FLAG"; echo "Commit: $$COMMIT_FLAG"; echo "BuildDate: $$BUILD_DATE_FLAG"; echo "Go: $$GO_VERSION_FLAG";

build-optimized: ## Build optimized binary with maximum performance flags (no DuckDB)
	@echo " Building optimized binary..."
	go build $(LDFLAGS) $(GCFLAGS) -o bin/costscope-optimized ./
	@echo " Optimized build completed"

build-optimized-duckdb: ## Build optimized binary with DuckDB-enabled features (adds size; includes Arrow/Thrift)
	@echo " Building optimized binary (with DuckDB, CGO enabled)..."
	@mkdir -p .cache/go-build .cache/go-tmp
	GOTMPDIR=$$(pwd)/.cache/go-tmp GOCACHE=$$(pwd)/.cache/go-build CGO_ENABLED=1 go build $(LDFLAGS) -tags duckdb -o bin/costscope-optimized-duckdb ./ || { echo " DuckDB build failed"; exit 1; }
	@echo " Optimized DuckDB build completed"

build-duckdb-debug: ## Build unstripped DuckDB debug binary (no LDFLAGS/GCFLAGS) for diagnostics
	@echo "  Building DuckDB debug binary (unstripped)..."
	@mkdir -p .cache/go-build .cache/go-tmp
	GOTMPDIR=$$(pwd)/.cache/go-tmp GOCACHE=$$(pwd)/.cache/go-build CGO_ENABLED=1 go build -tags duckdb -o bin/costscope-duckdb-debug ./ || { echo " DuckDB debug build failed"; exit 1; }
	@echo " DuckDB debug build: bin/costscope-duckdb-debug"

duckdb-smoke: build-duckdb-debug ## Run a minimal DuckDB driver smoke test (SELECT 42)
	@echo " Running DuckDB smoke test..."
	go run -tags duckdb ./scripts/tools/duckdb-smoketest || { echo " DuckDB smoke test failed"; exit 1; }
	@echo " DuckDB smoke test passed"

build-production: ## Build production binary for Linux/AMD64
	@echo " Building production binary..."
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(LDFLAGS) $(GCFLAGS) -o bin/costscope-production ./
	@echo " Production build completed"

build-enterprise: ## Build enterprise binary with enterprise features
	@echo " Building enterprise binary..."
	CGO_ENABLED=0 go build $(LDFLAGS) $(GCFLAGS) -tags enterprise -o bin/costscope-enterprise ./
	@echo " Enterprise build completed"
