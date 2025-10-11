# mk/dev.mk - development convenience & environment targets

.PHONY: clean dev env fmt fmt-check hooks-install install-hooks zip-backup

dev: ## Run fast build + tests + lint (developer iteration)
	$(MAKE) build-slim
	$(MAKE) test-slim
	$(MAKE) lint
	@echo " Dev cycle complete"

clean: ## Clean build artifacts, coverage, temp files
	@echo " Cleaning workspace..."
	rm -rf bin/* logs/coverage.out logs/coverage.html coverage.out coverage.html coverage.min perf_metrics.prom test_run.log tmp/* *.prof
	find . -type f -name '*.prof' -delete
	@echo " Clean complete"

fmt: ## Format Go code (go fmt + goimports if present)
	@echo " Formatting code..."
	go fmt ./...
	@if command -v goimports >/dev/null 2>&1; then goimports -w $$(go list -f '{{.Dir}}' ./...); else echo 'goimports not installed (optional)'; fi

fmt-check: ## Check formatting (fail if unformatted)
	@echo " Checking formatting..."
	@unformatted=$$(gofmt -l . | grep -v vendor | grep -v '^\./bin' || true); \
	if [ -n "$$unformatted" ]; then echo "Files need formatting:"; echo "$$unformatted"; exit 1; else echo " Formatting OK"; fi

install-hooks hooks-install: ## Install git pre-commit hooks
	@echo " Installing git hooks..."
	@[ -d .git/hooks ] || mkdir -p .git/hooks
	@cp scripts/hooks/pre-commit .git/hooks/pre-commit 2>/dev/null || true
	@chmod +x .git/hooks/pre-commit 2>/dev/null || true
	@echo " Hooks installed"

env: ## Show key environment info for diagnostics
	@echo "Go version: $$(go version)"; \
	 echo "Git version: $$(git --version)"; \
	 echo "OS: $$(uname -a)"; \
	 echo "GOMODCACHE: $$GOMODCACHE"; \
	 echo "GOROOT: $$GOROOT"; \
	 echo "GOPATH: $$GOPATH"; \
	 echo "Build tags (inferred): $$(grep -h "//go:build" -R . | cut -d' ' -f2 | sort -u | tr '\n' ' ' )"; \
	 echo "Binary artifacts: $$(ls -1 bin 2>/dev/null | tr '\n' ' ')"

zip-backup: ## Create a timestamped zip backup of key artifacts
	@ts=$$(date +%Y%m%d-%H%M%S); zip -r backup-$$ts.zip mk/ go.mod go.sum configs/ docs/ internal/ cmd/ scripts/ Makefile || true; echo "Created backup-$$ts.zip"
