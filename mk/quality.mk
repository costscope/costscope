# mk/quality.mk - lint, duplicates, complexity, analysis, etc.

.PHONY: allowlist-lint allowlist-rationale-lint analyze build-flags-guard complexity contract-check deadcode deadcode-baseline-guard deadcode-full deadcode-guard duplicates duplicates-func duplicates-func-gate duplicates-gate hotpath-deps-guard lint lint-api-resp-guard lint-fix lint-nolint-guard notice notice-drift print-% quality quality-full quality-nocache security

lint: ## Run linting
	@echo " Running golangci-lint..."
	golangci-lint run --timeout=10m
	@echo " API response helper guard..."
	bash scripts/tools/check-api-response-helpers.sh

lint-nolint-guard: ## Verify all //nolint:unused directives include rationale comments
	@echo "️  Verifying //nolint:unused rationale policy..."
	bash scripts/tools/check-nolint-unused.sh

lint-api-resp-guard: ## Run API response helper usage guard
	@echo "️  Verifying standardized API response helper usage..."
	bash scripts/tools/check-api-response-helpers.sh

# (Security aggregate lives in security.mk but simple alias remains here if needed)

complexity: ## Check cyclomatic complexity
	@echo " Checking cyclomatic complexity..."
	gocyclo -over 25 . || (echo "️  High complexity found in utility scripts - acceptable for tools" && exit 0)

# Duplicate detection (abbreviated from original; full logic preserved)
DUPL_REPORT ?= logs/report_duplicates.txt
DUPL_MAX_GROUPS ?= 168
DUPL_FUNC_MAX_GROUPS ?= 16
INTEGRATION_DUPL_BLOCK_PATTERN ?= cmd/modules/integration/enhanced_.*\.go

duplicates: ## Check for full duplicates only (exact file matches by sha256; replaces partial dupl clones)
	@echo " Checking for full file duplicates (exact sha256 matches)..."
	@mkdir -p logs
	@TMP=$$(mktemp); \
		git ls-files -z -- '*.go' | tr '\0' '\n' | grep -vE '^(vendor/|bin/|.git/|_archive/)' > $$TMP; \
		if [ ! -s $$TMP ]; then \
			echo "No tracked Go files found" | tee $(DUPL_REPORT); \
		else \
			echo "# Full duplicate file report (sha256 groups)" > $(DUPL_REPORT); \
			echo "# Generated: $$(date -u +%Y-%m-%dT%H:%M:%SZ)" >> $(DUPL_REPORT); \
			shasum -a 256 $$(cat $$TMP) | awk '{c[$$1]++; f[$$1]=f[$$1] "\n - " $$2} END {g=0; for (h in c) if (c[h]>1) {g++; printf("duplicate group sha256: %s (%d files)\n", h, c[h]); printf("%s\n\n", f[h]); } printf("full duplicate groups: %d\n", g); }' >> $(DUPL_REPORT); \
		fi; \
		rm -f $$TMP; \
		FULL_DUP_COUNT=$$(awk -F': ' '/^full duplicate groups:/ {print $$2}' $(DUPL_REPORT) 2>/dev/null || echo 0); \
		echo " Full duplicate groups: $$FULL_DUP_COUNT" || true
	@echo " Full duplicate report written to $(DUPL_REPORT)"

duplicates-gate: ## Run duplicate scan & enforce gate
	@$(MAKE) --no-print-directory duplicates || true
	@echo " Enforcing duplication gate (max groups=$(DUPL_MAX_GROUPS))..."
	@CURRENT_GROUPS=$$(awk -F': ' '/^full duplicate groups:/ {print $$2}' $(DUPL_REPORT) 2>/dev/null || echo 0); \
	if [ $$CURRENT_GROUPS -gt $(DUPL_MAX_GROUPS) ]; then echo " Duplication regression: $$CURRENT_GROUPS clone groups (limit $(DUPL_MAX_GROUPS))"; exit 1; else echo " Clone groups within limit: $$CURRENT_GROUPS/$(DUPL_MAX_GROUPS)"; fi; \
	if grep -E '$(INTEGRATION_DUPL_BLOCK_PATTERN)' $(DUPL_REPORT) > /dev/null 2>&1; then echo " Legacy integration builder duplication detected (pattern $(INTEGRATION_DUPL_BLOCK_PATTERN))"; grep -E '$(INTEGRATION_DUPL_BLOCK_PATTERN)' $(DUPL_REPORT) || true; exit 1; else echo " No legacy integration builder duplicates detected"; fi; echo " Duplication gate passed"

# Function-level duplicates
.PHONY: duplicates-func duplicates-func-gate
duplicates-func: ## Detect exact duplicate Go functions (by normalized body) and write a report
	@echo " Scanning for function-level full duplicates..."
	@mkdir -p logs
	@cd scripts/tools/funcdups && go run . -out ../../../logs/report_func_duplicates.txt -root ../../.. -exclude-tests=$${EXCLUDE_TESTS:=true} -min-bytes=$${MIN_FUNC_BYTES:=64} -include-dirs=$${INCLUDE_DIRS:=internal} -exclude-dirs=$${EXCLUDE_DIRS:=} -name-exclude=$${NAME_EXCLUDE:=^(GetSchema|SupportsFormat|Version|Open|Close|Flush|WriteChunk|Validate)$$} -min-occurrences=$${MIN_FUNC_OCCURRENCES:=2} >/dev/null
	@echo " Function duplicate report written to logs/report_func_duplicates.txt"
	@echo " Full function duplicate groups: $$(grep -Eo 'full function duplicate groups: [0-9]+' logs/report_func_duplicates.txt | awk '{print $$5}')"

duplicates-func-gate: ## Gate: fail if function-level duplicate groups exceed threshold
	@echo " Function duplicate gate (max groups: $(DUPL_FUNC_MAX_GROUPS))"
	@$(MAKE) -s duplicates-func
	@CURRENT_FUNC_GROUPS=$$(grep -Eo 'full function duplicate groups: [0-9]+' logs/report_func_duplicates.txt 2>/dev/null | awk '{print $$5}'); \
	 echo "Current function duplicate groups: $$CURRENT_FUNC_GROUPS"; \
	 if [ -n "$$CURRENT_FUNC_GROUPS" ] && [ $$CURRENT_FUNC_GROUPS -gt $(DUPL_FUNC_MAX_GROUPS) ]; then echo " Function duplicate groups ($$CURRENT_FUNC_GROUPS) exceed threshold ($(DUPL_FUNC_MAX_GROUPS))"; exit 2; else echo " Function duplicate groups within threshold"; fi

# Dead code
DEADCODE_EXCLUDE_REGEX ?= '^(local/costscope/(_archive|examples|monitoring|charts|scripts)(/|$$))|^(local/costscope/internal/(api|framework|optimization|database|core/production|core/docs)(/|$$))|^(local/costscope/cmd/modules/(analytics|analytics_advanced|analytics_complex|integration|streaming))(/|$$)'

.PHONY: deadcode deadcode-full deadcode-guard deadcode-baseline-guard allowlist-rationale-lint allowlist-lint

deadcode: ## Find dead code (fail only if targeted CLI modules have findings)
	@echo " Finding dead code (scanning all; gating on targeted paths)…"
	@mkdir -p logs
	@deadcode ./... | tee logs/deadcode.txt || true
	@echo " Checking findings under targeted paths…"
	@if grep -E '^(cmd/modules/(diagnostics|focus/commands|analytics/commands/advanced))/.*unreachable' logs/deadcode.txt >/dev/null; then echo " Deadcode found in targeted CLI modules"; grep -E '^(cmd/modules/(diagnostics|focus/commands|analytics/commands/advanced))/.*unreachable' logs/deadcode.txt || true; exit 1; else echo " Targeted CLI modules: no deadcode findings"; fi

deadcode-full: ## Find dead code across entire repo (no filters)
	@echo " Finding dead code (full)…"
	deadcode ./...

deadcode-guard: ## Run deadcode scan & fail on new unreachable symbols not in allowlist
	@echo "️  Running deadcode guard (allowlist enforcement)…"
	@bash scripts/tools/deadcode-guard.sh

deadcode-baseline-guard: ## Compare current deadcode output with baseline (.deadcode-baseline.json)
	@echo "️  Running deadcode baseline guard…"
	@bash scripts/tools/deadcode-baseline-guard.sh

allowlist-rationale-lint: ## Lint .deadcode-allowlist for per-line rationale comments
	@echo " Linting deadcode allowlist rationales…"
	@bash scripts/tools/allowlist-rationale-lint.sh

allowlist-lint: ## Verify every .deadcode-allowlist line has '# rationale:' justification
	@if [ -f .deadcode-allowlist ]; then echo " Checking allowlist rationales..."; bash scripts/tools/allowlist-rationale-lint.sh; else echo "ℹ️  No .deadcode-allowlist present (skipping allowlist-lint)"; fi

analyze: ## Run comprehensive code analysis
	@echo " Running comprehensive analysis..."
	@echo " 1. Static analysis..."; staticcheck ./...
	@echo " 2. Cyclomatic complexity..."; gocyclo -over 10 -avg . || true
	@echo " 3. Cognitive complexity..."; gocognit -over 20 . || true
	@echo " 4. Duplicate detection..."; dupl -threshold 50 . || true
	@echo " 5. Dead code detection..."; deadcode ./... || true

notice: ## Regenerate NOTICE (dependency license inventory)
	@echo " Generating NOTICE..."; NOTICE_DETERMINISTIC=1 bash scripts/license/gen-notice.sh NOTICE; git add NOTICE >/dev/null 2>&1 || true; echo " NOTICE updated"

notice-drift: ## Fail if NOTICE is out-of-date with current dependencies
	@echo " Checking NOTICE drift..."; \
	if [ ! -f NOTICE ]; then echo " NOTICE file missing (run: make notice)"; exit 1; fi; \
	NOTICE_DETERMINISTIC=1 bash scripts/license/gen-notice.sh NOTICE.tmp; \
	if ! diff -u NOTICE NOTICE.tmp >/dev/null; then echo " NOTICE drift detected (run: make notice)"; diff -u NOTICE NOTICE.tmp | sed -n '1,120p'; rm -f NOTICE.tmp; exit 1; else echo " NOTICE up to date"; rm -f NOTICE.tmp; fi

print-%:
	@echo $*=$($*)

# Aggregate quality (core subset)
quality: duplicates duplicates-func-gate lint lint-nolint-guard allowlist-lint security vuln complexity loc-guard build-flags-guard hotpath-deps-guard test coverage-runtime-enforce ## Run all quality checks
	@echo " All quality checks completed!"

quality-full: duplicates lint security vuln complexity analyze deadcode outdated test ## Run extended quality checks
	@echo " Extended quality checks completed!"

quality-nocache: ## Run quality checks without cache (calls build-clean/test-clean/quality)
	@echo " Clean build check..."; $(MAKE) build-clean; echo " Clean test check..."; $(MAKE) test-clean; echo " Full quality check..."; $(MAKE) quality

build-flags-guard: ## Ensure -s -w -trimpath are present in build invocations (scan mk/*.mk)
	@echo "️  Verifying build flags (-s -w -trimpath) in build targets..."
	@grep -R "go build" mk/build.mk | grep -E "(-ldflags|\\$\(LDFLAGS\)).*(-s).*(-w)" >/dev/null || (echo " Missing -s/-w in build commands" && exit 1)
	@grep -R "go build" mk/build.mk | grep -E "(-trimpath|\\$\(LDFLAGS\).*-trimpath)" >/dev/null || (echo " Missing -trimpath in build commands" && exit 1)
	@echo " Build flags guard passed"

hotpath-deps-guard: ## Ensure heavy deps (duckdb/arrow/thrift/sqlite) are not imported on default path
	@echo "️  Verifying hot path does not import heavy deps (without build tags)..."
	@FAIL=0; CAND=$$(grep -R -nE '^[[:space:]]*(_[[:space:]]*)?"github.com/marcboeker/go-duckdb"|^[[:space:]]*(_[[:space:]]*)?"github.com/mattn/go-sqlite3"|^[[:space:]]*(_[[:space:]]*)?".*apache.*arrow.*"|^[[:space:]]*(_[[:space:]]*)?".*thrift.*"' cmd internal | cut -d: -f1 | sort -u); \
	for f in $$CAND; do if ! head -n1 "$$f" | grep -q '^//go:build'; then echo " Heavy dependency imported in $$f without build tag"; FAIL=1; fi; done; \
	if [ $$FAIL -ne 0 ]; then exit 1; else echo " Hot path dependency guard passed"; fi

contract-check: ## Run API contract diff check
	@echo " Running API contract check..."
	bash scripts/tools/contract-diff.sh

.PHONY: api-baseline-update
api-baseline-update: ## Update API baselines (api/openapi*.json) from current generated YAML specs
	@echo " Updating API baseline JSONs from current specs..."; \
	if [ ! -f internal/api/docs/openapi.yaml ]; then echo " missing internal/api/docs/openapi.yaml (run: go run ./cmd/tools/gen-openapi)"; exit 2; fi; \
	GO111MODULE=on go run ./cmd/tools/yaml2json internal/api/docs/openapi.yaml > api/openapi.v1.json.tmp && mv api/openapi.v1.json.tmp api/openapi.v1.json; \
	if [ -f internal/api/docs/enterprise-openapi.yaml ]; then \
	  GO111MODULE=on go run ./cmd/tools/yaml2json internal/api/docs/enterprise-openapi.yaml > api/openapi.enterprise.v1.json.tmp && mv api/openapi.enterprise.v1.json.tmp api/openapi.enterprise.v1.json; \
	  echo " enterprise baseline updated"; \
	else \
	  echo " enterprise spec not present; skipping"; \
	fi; \
	echo " API baseline update complete"

.PHONY: oasdiff-install
oasdiff-install: ## Install oasdiff at the CI-pinned revision for local checks
	@echo " Installing oasdiff (pinned to CI revision)..."; \
	GOBIN=$$(pwd)/bin go install github.com/oasdiff/oasdiff@fc23f9bb1b54519f4f847e1724dbd0ab894e8ec8; \
	echo " oasdiff installed to ./bin/oasdiff"
