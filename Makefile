## Root Makefile (delegator)
## Only includes fragment mk/*.mk and exposes meta (help) & lightweight audit targets.

# Ordered includes (variables/config first). -include = non-fatal missing.
-include mk/common.mk
-include mk/build.mk
-include mk/test.mk
-include mk/perf.mk
-include mk/gen.mk
-include mk/quality.mk
-include mk/security.mk
-include mk/release.mk
-include mk/docker.mk
-include mk/tools.mk
-include mk/dev.mk
-include mk/docs.mk
-include mk/internal.mk

.PHONY: help
help: ## Show aggregated help for all public targets
	@echo "CostScope Make Targets"; echo ""; \
	grep -hE '^[A-Za-z0-9_.-]+:.*## ' $(MAKEFILE_LIST) | \
	awk -v pfx="${INTERNAL_PREFIX:_}" 'BEGIN{FS=":.*## "} {n=$$1; if (index(n,pfx)==1) next; printf "\033[36m%-28s\033[0m %s\n", n, $$2}' | sort

.PHONY: help-json
help-json: ## Emit machine-readable JSON list of targets/descriptions
	@grep -hE '^[A-Za-z0-9_.-]+:.*## ' $(MAKEFILE_LIST) | \
	awk -v pfx="${INTERNAL_PREFIX:_}" 'BEGIN{print "["} {t=$$1; sub(/:.*/,"",t); if (index(t,pfx)==1) next; line=$$0; sub(/^[^#]*## /,"",line); gsub(/"/,"\\\""); if (c++) printf(",\n"); printf("{\"target\":\"%s\",\"description\":\"%s\"}", t, line)} END{print "\n]"}'

.PHONY: help-lint
help-lint: ## Verify all public targets have help comments (##)
	@missing=$$(awk -v pfx="${INTERNAL_PREFIX:_}" -F: 'BEGIN{IGNORECASE=0} /^[A-Za-z0-9_.-]+:/{t=$$1; if(t ~ /%/ || t==".PHONY" || index(t,pfx)==1) next; if($$0 !~ /##/) miss[t]=FILENAME} END{for(m in miss){print m" ("miss[m]")"} exit(length(miss)>0)}' Makefile mk/*.mk); \
	if [ -n "$$missing" ]; then echo " Missing help comments for:"; echo "$$missing"; exit 1; else echo " All public targets documented"; fi

.PHONY: targets-duplicate-guard
targets-duplicate-guard: ## Fail if duplicate public targets are defined across fragments
	@awk -v pfx="${INTERNAL_PREFIX:_}" -F: '/^[A-Za-z0-9_.-]+:/{t=$$1; if(t ~ /%/ || t==".PHONY" || index(t,pfx)==1) next; seen[t]++; files[t]=files[t]" "FILENAME} END{e=0; for(k in seen) if(seen[k]>1){print "Duplicate target:" k " in" files[k]; e=1} exit e}' Makefile mk/*.mk

.PHONY: duplicates-gate
duplicates-gate: targets-duplicate-guard ## Alias for CI: duplicate targets guard

.PHONY: sanity
sanity: ## Fast pre-commit sanity (fmt-check, vet, staticcheck, build-slim)
	@$(MAKE) --no-print-directory fmt-check
	@echo " go vet"; go vet ./...
	@if ! command -v staticcheck >/dev/null 2>&1; then echo "Installing staticcheck..."; go install honnef.co/go/tools/cmd/staticcheck@latest; fi
	@staticcheck ./... || (echo "️ staticcheck issues (non-fatal)" && true)
	@$(MAKE) --no-print-directory build-slim
	@echo " Sanity OK"

.PHONY: ci-audit
ci-audit: ## Audit workflow files for unknown 'make <target>' invocations
	@command -v jq >/dev/null 2>&1 || { echo "jq required (install first)"; exit 2; }
	@known=$$( $(MAKE) --no-print-directory help-json | jq -r '.[].target'); \
	awk '/make /{for(i=1;i<=NF;i++) if($$i=="make"){t=$(i+1); if(t!="" && t!~/^-/) print t}}' .github/workflows/*.yml | sort -u > /tmp/ci_make_targets.txt; \
	missing=0; while read t; do echo "$$known" | grep -qx "$$t" || { echo " Unknown workflow target: $$t"; missing=1; }; done < /tmp/ci_make_targets.txt; \
	if [ $$missing -ne 0 ]; then echo "CI audit failed"; exit 1; else echo " CI audit OK"; fi

.PHONY: guardrails
guardrails: ## Lightweight guard target for CI pre-flight (no-op placeholder for local runner compatibility)
	@echo "Running guardrails (no-op)"

.PHONY: all
all: build ## Default build

.PHONY: tests
tests: test ## Alias for test

.PHONY: scan-packages
scan-packages: ## Run repository package scanner (scripts/scan_packages.sh)
	@./scripts/scan_packages.sh

.PHONY: includes-check
includes-check: ## Verify core include fragments exist
	@missing=0; for f in mk/common.mk mk/build.mk mk/test.mk mk/quality.mk; do [ -f $$f ] || { echo "Missing required include: $$f"; missing=1; }; done; [ $$missing -eq 0 ] || { echo " Required include files missing"; exit 1; }

.PHONY: coverage-guard
coverage-guard: ## Fail if mapping coverage drops by >2pp from baseline (lightweight drift gate)
	@tmp=$$(mktemp); echo "[coverage-guard] running mapping tests"; \
	go test -coverprofile=$$tmp ./internal/framework/mapping >/dev/null 2>&1 || { echo "[coverage-guard] test run failed"; rm -f $$tmp; exit 1; }; \
	cur=$$(go tool cover -func=$$tmp | awk '/total:/ {gsub("%","",$$3); print $$3}'); rm -f $$tmp; baseline=87.2; allowed=2.0; min=$$(awk -v b=$$baseline -v a=$$allowed 'BEGIN{printf "%.1f", b-a}'); \
	echo "[coverage-guard] current=$$cur baseline=$$baseline min=$$min"; awk -v c=$$cur -v m=$$min 'BEGIN{ if (c+0 < m+0) exit 1 }' || { echo "[coverage-guard] FAIL: $$cur < $$min"; exit 2; }; echo "[coverage-guard] OK";

.PHONY: coverage-guard-production
coverage-guard-production: ## Fail if production coverage drops by >2pp from baseline (lightweight drift gate)
	@set -euo pipefail; if [ -x scripts/ci/coverage_guard.sh ]; then \
	  scripts/ci/coverage_guard.sh --mode production; \
	else \
	  echo "[coverage-guard] script missing, using inline fallback" >&2; \
	  tmp=$$(mktemp); echo "[coverage-guard] running production tests"; \
	  if ! go test -count=1 -coverprofile=$$tmp ./internal/core/production 2>&1 | sed 's/^/[coverage-guard] test: /'; then \
	    echo "[coverage-guard] test run failed"; rm -f $$tmp; exit 1; \
	  fi; \
	  cur=$$(go tool cover -func=$$tmp | awk '/total:/ {gsub("%","",$$3); print $$3}'); rm -f $$tmp; baseline=92; allowed=2.0; min=$$(awk -v b=$$baseline -v a=$$allowed 'BEGIN{printf "%.1f", b-a}'); \
	  echo "[coverage-guard] current=$$cur baseline=$$baseline min=$$min"; awk -v c=$$cur -v m=$$min 'BEGIN{ if (c+0 < m+0) exit 1 }' || { echo "[coverage-guard] FAIL: $$cur < $$min"; exit 2; }; echo "[coverage-guard] OK"; \
	fi

.PHONY: config-validate
config-validate: ## Validate layered configuration files (YAML) via scripts/tools/config-validate
	@bash ./scripts/ci/run-config-validate.sh
