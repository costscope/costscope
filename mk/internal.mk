# mk/internal.mk - internal/private helper targets (prefixed with _)

INTERNAL_PREFIX ?= _

.PHONY: _echo-env _noop

_echo-env: ## (internal) print selected env vars
	@echo INTERNAL_PREFIX=$(INTERNAL_PREFIX)
	@echo GOOS=$$(go env GOOS) GOARCH=$$(go env GOARCH)

_noop:
	@true
