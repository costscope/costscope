# mk/gen.mk - code generation & spec related targets

# Ensure Go invocations during generation/drift checks don't fail due to missing VCS metadata
# in ephemeral CI containers. Respect existing GOFLAGS and append -buildvcs=false.
GOFLAGS := $(GOFLAGS) -buildvcs=false
export GOFLAGS

.PHONY: gen-actions gen-actions-generate gen-commands gen-commands-drift gen-enterprise-stub gen-integration-cli-drift gen-integration-cli-docs gen-provider loc-guard

gen-enterprise-stub: ## Generate enterprise stub scaffolds
	@./scripts/generate-enterprise-stubs.sh "$(FEATURE)" "$(PKG)" "$(TYPE)" "$(IFACE)"
	@echo "Generated stubs for feature=$(FEATURE) type=$(TYPE)"

loc-guard: ## Fail if serve.go exceeds 450 lines (API LOC budget)
	@MAX=450; FILE=cmd/modules/api/serve.go; \
	LINES=$$(wc -l < $$FILE | tr -d ' '); \
	if [ $$LINES -gt $$MAX ]; then \
	  echo " LOC budget exceeded for $$FILE: $$LINES > $$MAX"; \
	  exit 1; \
	else \
	  echo " LOC budget OK for $$FILE: $$LINES/$$MAX"; \
	fi

gen-provider: ## Generate new provider scaffold (usage: make gen-provider name=foo [force=1])
	@if [ -z "$(name)" ]; then echo "name= (provider name) required"; exit 1; fi
	@echo " Generating provider scaffold for $(name)..." ; \
	CMD="go run ./scripts/tools/gen-provider -name $(name)"; \
	if [ "$(force)" = "1" ]; then CMD="$$CMD -force"; fi; \
	sh -c "$$CMD" || { echo " gen-provider failed"; exit 2; }

# Generated command files (kept in source control)
GENERATED_COMMAND_FILES := \
	cmd/modules/analytics/commands/zz_generated_command_builder.go \
	cmd/modules/multicloud/commands/zz_generated_command_builder.go

gen-commands: ## Generate Cobra command builders from specs (analytics, multicloud)
	@echo "️  Generating Cobra builders from specs..."
	go run -buildvcs=false ./scripts/tools/commandgen -receiver AnalyticsCommands -spec cmd/modules/analytics/commands/command_spec.yaml -out cmd/modules/analytics/commands/zz_generated_command_builder.go
	go run -buildvcs=false ./scripts/tools/commandgen -receiver MulticloudCommands -spec cmd/modules/multicloud/commands/command_spec.yaml -out cmd/modules/multicloud/commands/zz_generated_command_builder.go
	@if command -v gofmt >/dev/null 2>&1; then \
		gofmt -w cmd/modules/analytics/commands/zz_generated_command_builder.go cmd/modules/multicloud/commands/zz_generated_command_builder.go || true; \
	fi
	@chmod 0644 cmd/modules/analytics/commands/zz_generated_command_builder.go cmd/modules/multicloud/commands/zz_generated_command_builder.go || true
	@echo " Command builders generated"

gen-commands-drift: ## Guard: fail if regenerated command builders introduce diff (delegates to script)
	@echo " Running CLI drift check script..."
	bash scripts/ci/cli-drift-check.sh

gen-integration-cli-docs: ## Generate integration CLI command summary (prototype)
	@echo " Generating integration CLI command summary..."
	go run -buildvcs=false ./scripts/tools/gen-integration-cli-docs -out integration_commands.json -md-out docs/integration_commands.md
	@echo " Wrote integration_commands.json and docs/integration_commands.md"

gen-integration-cli-drift: ## Check for drift in integration CLI command summary
	@echo " Checking integration CLI command drift..."
	go run -buildvcs=false ./scripts/tools/gen-integration-cli-docs -out integration_commands.json -md-out docs/integration_commands.md -drift-check || (echo " Drift detected" && exit 1)
	@echo " No drift detected"

gen-actions: ## Validate Integration Action DSL
	@echo " Validating Integration Action DSL..."
	go run -buildvcs=false ./scripts/tools/gen-actions -lint-only -yaml cmd/modules/integration/actions.yaml || (echo " Integration Action DSL validation failed" && exit 1)
	@echo " Integration Action DSL valid"

gen-actions-generate: ## Generate BuildDefaultActionSpecs from DSL
	@echo "️  Generating ActionSpecs from DSL..."
	go run -buildvcs=false ./scripts/tools/gen-actions -yaml cmd/modules/integration/actions.yaml -gen-go cmd/modules/integration/generated_actions.go
	@echo " Generated cmd/modules/integration/generated_actions.go"

.PHONY: generate
generate: ## Run project-wide go generate (fallback used by CI/workflows)
	@echo "Running: go generate ./..."
	go generate ./...
