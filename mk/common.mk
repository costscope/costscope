# mk/common.mk - shared variables & lightweight utility targets

# Version & build flags (shared)
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)

# External linker flags (macOS-specific strip) are disabled by default.
# Rationale: build-slim uses internal linking (CGO_ENABLED=0); passing
# -extldflags with '-Wl,-dead_strip' causes the go internal linker to error
# (flag provided but not defined). For targets that force external linking
# (e.g., CGO_ENABLED=1), consider appending these flags locally in that target.
EXT_LD_ARG :=

LDFLAGS := -ldflags="-w -s -X main.version=$(VERSION) -X main.buildTime=$(shell date -u +%Y-%m-%dT%H:%M:%SZ) $(EXT_LD_ARG)" -trimpath -buildvcs=false
GCFLAGS := -gcflags="-l=4"

# Security thresholds
GOVULNCHECK_SCORE_MIN ?= 7
GOSEC_FAIL_LEVEL ?= HIGH

.PHONY: deps
deps: ## Download and tidy Go modules
	@echo " Managing Go dependencies..."
	go mod download
	go mod tidy
