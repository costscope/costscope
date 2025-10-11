# mk/test.mk - test & coverage targets

.PHONY: coverage-enforce coverage-runtime coverage-runtime-enforce coverage-pkg-enforce invariants-baseline test test-clean test-coverage test-duckdb test-matrix test-slim test-sqlite

test: test-slim ## Run tests (alias to slim)

test-slim: ## Run tests in slim mode (default; no build tags)
	@echo " Running tests (slim default)..."
	env -u GOROOT GOTOOLCHAIN=auto go test -v -race -cover ./...

test-sqlite: ## Run tests with SQLite-backed persistence (-tags sqlite; CGO required)
	@echo " Running tests (sqlite tag)..."
	CGO_ENABLED=1 env -u GOROOT GOTOOLCHAIN=auto go test -v -race -cover -tags sqlite ./...

test-duckdb: ## Run tests with DuckDB-enabled analytics (-tags duckdb; CGO required)
	@echo " Running tests (duckdb tag)..."
	CGO_ENABLED=1 env -u GOROOT GOTOOLCHAIN=auto go test -v -race -cover -tags duckdb ./...

test-matrix: ## Run slim, sqlite, and duckdb test suites sequentially
	@echo " Running test matrix: slim → sqlite → duckdb"
	$(MAKE) --no-print-directory test-slim
	$(MAKE) --no-print-directory test-sqlite
	$(MAKE) --no-print-directory test-duckdb

test-clean: ## Run tests with clean cache
	@echo " Running tests with clean cache..."
	go clean -testcache
	env -u GOROOT GOTOOLCHAIN=auto go test -v -race -cover ./...

# Coverage (global)
COVERAGE_MIN ?= 90.0

test-coverage: ## Run tests with coverage report
	@echo " Running tests with coverage..."
	@mkdir -p logs
	env -u GOROOT GOTOOLCHAIN=auto go test -v -race -coverprofile=logs/coverage.out ./...
	go tool cover -html=logs/coverage.out -o logs/coverage.html
	@echo "Coverage report generated: logs/coverage.html"

coverage-enforce: test-coverage ## Fail if coverage is below COVERAGE_MIN
	@cov=$$(go tool cover -func=coverage.out | awk '/total:/ {print $$3}' | sed 's/%//'); \
	if awk -v c="$$cov" -v min=$(COVERAGE_MIN) 'BEGIN{exit !(c+0 >= min)}'; then \
		echo " Coverage $$cov% >= $(COVERAGE_MIN)%"; \
	else \
		echo " Coverage $$cov% < $(COVERAGE_MIN)%"; \
		exit 1; \
	fi

# Runtime coverage subset
RUNTIME_PKG_REGEX ?= '^(local/costscope/internal/(core|providers))'
EXCLUDE_PKG_REGEX ?= '^(local/costscope/(scripts|_archive|examples|monitoring|charts)(/|$$))|^(local/costscope/scripts/tools)(/|$$)|^(local/costscope/internal/core/production)(/|$$)|^(local/costscope/internal/core/docs)(/|$$)|^(local/costscope/internal/providers/testutils)(/|$$)|^(local/costscope/internal/core/reports/types)(/|$$)'
PKGS_RUNTIME := $(shell go list ./... | grep -E $(RUNTIME_PKG_REGEX) | grep -Ev $(EXCLUDE_PKG_REGEX))

coverage-runtime: ## Run coverage for runtime packages only (excludes scripts/dev-tools)
	@echo " Running runtime-only coverage (excluding scripts/dev-tools)..."
	@if [ -z "$(PKGS_RUNTIME)" ]; then echo "No runtime packages detected"; exit 1; fi
	@PKGS="$(PKGS_RUNTIME)"; \
	 EXCL_FILE=$(COVERAGE_PKG_MIN_FILE); \
	 if [ -f "$$EXCL_FILE" ]; then \
	   echo " Using overrides from $$EXCL_FILE (ignore/skip entries excluded from total; prefixes allowed)"; \
	   TMP_PKGS=$$(mktemp); printf "%s\n" $$PKGS | tr ' ' '\n' > $$TMP_PKGS; \
	   TMP_EXCL=$$(mktemp); awk -F'=' '$$2 ~ /^(ignore|skip)$$/ {print $$1}' $$EXCL_FILE > $$TMP_EXCL; \
	   if [ -s $$TMP_EXCL ]; then \
	     BEFORE=$$(wc -l < $$TMP_PKGS | tr -d ' '); \
	     while IFS= read -r pref; do \
	       [ -n "$$pref" ] || continue; \
	       grep -Ev "^$${pref}(/|$$)" "$$TMP_PKGS" > "$$TMP_PKGS.new" || true; \
	       mv "$$TMP_PKGS.new" "$$TMP_PKGS"; \
	     done < "$$TMP_EXCL"; \
	     PKGS=$$(tr '\n' ' ' < "$$TMP_PKGS"); \
	     AFTER=$$(printf "%s\n" $$PKGS | tr ' ' '\n' | sed '/^$$/d' | wc -l | tr -d ' '); \
	     EXCL_LIST=$$(tr '\n' ',' < "$$TMP_EXCL" | sed 's/,$$//'); \
	     echo "   → Excluded prefixes: $$EXCL_LIST (kept $$AFTER of $$BEFORE)"; \
	   fi; \
	   rm -f $$TMP_PKGS $$TMP_EXCL; \
	 fi; \
	 if [ -z "$$PKGS" ]; then echo "No runtime packages left after exclusions"; exit 1; fi; \
	 echo " Effective runtime package count: $$(printf "%s" "$$PKGS" | tr ' ' '\n' | sed '/^$$/d' | wc -l)"; \
	 env -u GOROOT GOTOOLCHAIN=auto go test -v -race -coverprofile=coverage.out $$PKGS
	go tool cover -html=logs/coverage.out -o logs/coverage.html
	@echo "Runtime coverage report generated: logs/coverage.html"

COVERAGE_RUNTIME_MIN ?= 70.0
coverage-runtime-enforce: coverage-runtime ## Enforce minimum runtime-only coverage (COVERAGE_RUNTIME_MIN)
	@cov=$$(go tool cover -func=coverage.out | awk '/total:/ {print $$3}' | sed 's/%//'); \
	if awk -v c="$$cov" -v min=$(COVERAGE_RUNTIME_MIN) 'BEGIN{exit !(c+0 >= min)}'; then \
	  echo " Runtime coverage $$cov% >= $(COVERAGE_RUNTIME_MIN)%"; \
	else \
	  echo " Runtime coverage $$cov% < $(COVERAGE_RUNTIME_MIN)%"; \
	  echo "   Hint: excludes scripts/_archive/examples/monitoring/charts"; \
	  exit 1; \
	fi

COVERAGE_PKG_MIN ?= 70.0
COVERAGE_PKG_MIN_FILE ?= coverage.min
coverage-pkg-enforce: ## Enforce per-package minimum coverage for runtime packages
	@echo " Enforcing per-package coverage >= $(COVERAGE_PKG_MIN)% for runtime packages..."
	@if [ -z "$(PKGS_RUNTIME)" ]; then echo "No runtime packages detected"; exit 1; fi
	@fail=0; \
	for p in $(PKGS_RUNTIME); do \
	  out=$$(echo $$p | sed 's#[^a-zA-Z0-9_/.-]##g; s#/#_#g'); \
	  echo " → $$p"; \
	  env -u GOROOT GOTOOLCHAIN=auto go test -coverprofile=tmp.cover.$$out -covermode=atomic $$p >/tmp/cover.$$out.log 2>&1 || true; \
	  pct=$$(grep -Eo 'coverage: [0-9]+\.?[0-9]*% of statements' /tmp/cover.$$out.log | awk '{print $$2}' | tr -d '%'); \
	  if [ -z "$$pct" ]; then pct=0; fi; \
	  MIN=$(COVERAGE_PKG_MIN); \
	  if [ -f "$(COVERAGE_PKG_MIN_FILE)" ]; then \
	    OV=$$(awk -F'=' -v pkg="$$p" '$$1==pkg {print $$2}' $(COVERAGE_PKG_MIN_FILE)); \
	    if [ "$$OV" != "" ]; then MIN=$$OV; fi; \
	  fi; \
	  if awk -v c="$$pct" -v min=$$MIN 'BEGIN{exit !(c+0 >= min)}'; then \
	    echo "     $$pct%"; \
	  else \
	    echo "     $$pct% (< $${MIN}%)"; \
	    fail=1; \
	  fi; \
	  rm -f tmp.cover.$$out /tmp/cover.$$out.log || true; \
	done; \
	if [ $$fail -ne 0 ]; then echo " Per-package coverage gate failed"; exit 2; else echo " Per-package coverage gate passed"; fi

# Invariants baseline regeneration (kept minimal here)
invariants-baseline: ## Regenerate invariants baseline (usage: make invariants-baseline file=... out=... tol=0.01)
	@if [ -z "$(file)" ]; then echo "file= (input FOCUS file) required"; exit 1; fi
	@if [ -z "$(out)" ]; then echo "out= (output baseline JSON) required"; exit 1; fi
	@echo "Generating invariants baseline from $(file) -> $(out) (tolerance=$(tol))"
	@./costscope invariants regenerate $(file) --output $(out) $(if $(tol),--tolerance $(tol),) --force
	@echo "Baseline written to $(out)"
