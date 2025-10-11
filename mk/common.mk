# mk/common.mk - shared variables & lightweight utility targets

# Version & build flags (shared)
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)

# Conditionally include platform-specific external linker flags.
# The '-dead_strip' flag is a Darwin/macOS linker option and is rejected
# by GNU ld on Linux (error: unable to disambiguate: -dead_strip). Only
# append the extldflags when building on Darwin hosts. This keeps local
# builds on Linux-compatible CI/act runners working while preserving the
# original optimization on macOS hosts.
UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Darwin)
	EXT_LD_ARG := -extldflags '\'-Wl,-dead_strip -Wl,-x\''
else
	EXT_LD_ARG :=
endif

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
