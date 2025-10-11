# mk/docs.mk - canonical, minimal docs helpers (single copy)

.PHONY: docs docs-lint docs-missing docs-link-check check-docs docs-placeholder-guard

docs: ## Placeholder docs aggregate (no-op)
	@echo "Docs aggregate (no-op)"

docs-lint: ## Lint markdown files for trailing spaces
	@bash scripts/docs-lint.sh

docs-missing: ## Check for a few high-level docs
	@bash scripts/docs-missing.sh

docs-link-check: ## Verify local markdown links resolve
	@bash scripts/docs-link-check.sh

check-docs: ## Run additional doc checks
	@if [ -f scripts/check-docs.sh ]; then bash scripts/check-docs.sh; else echo "No scripts/check-docs.sh found, skipping"; fi


docs-placeholder-guard: ## Fail if migration placeholder markers remain
	@bash scripts/docs-placeholder-guard.sh
