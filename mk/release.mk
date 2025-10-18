## Release helpers

.PHONY: release-dry-run
release-dry-run: ## Local dry-run: build, preview notes, checksums, SBOM (no tag push). Usage: make release-dry-run VERSION=vX.Y.Z[-rc]
	@[ -n "$(VERSION)" ] || { echo "VERSION is required (e.g. v0.1.0)"; exit 2; }
	bash scripts/release/local-dry-run.sh $(VERSION)

# mk/release.mk - release & promotion related targets

.PHONY: checksums release release-notes release-promo sign-checksums supply-chain-all verify-checksums sign-checksums-act verify-checksums-act

.PHONY: gen-version-json
gen-version-json: ## Generate version.json at repository root (for release pipelines)
	@echo " Generating version.json..."; \
	bash ./scripts/generate-version-json.sh || { echo " gen-version-json failed"; exit 1; }

release: clean quality build-release ## Build release version with quality checks
	@echo " Release build completed!"

# Compare current community OpenAPI spec against baseline (non-fatal by default)
API_SPEC ?= api/openapi.v1.json
API_BASELINE ?= api/baseline/openapi.v1.json
api-contract-diff: ## Diff current OpenAPI spec vs baseline (fails if breaking removal detected)
	@echo " Diffing OpenAPI spec ($(API_SPEC)) vs baseline ($(API_BASELINE))"; \
	if [ ! -f $(API_SPEC) ] || [ ! -f $(API_BASELINE) ]; then echo " Spec or baseline missing"; exit 1; fi; \
	TMP_CUR=$$(mktemp); TMP_BASE=$$(mktemp); \
	jq -r '.paths | keys[]' $(API_SPEC) 2>/dev/null | sort > $$TMP_CUR; \
	jq -r '.paths | keys[]' $(API_BASELINE) 2>/dev/null | sort > $$TMP_BASE; \
	CUR_COUNT=$$(wc -l < $$TMP_CUR | tr -d ' '); BASE_COUNT=$$(wc -l < $$TMP_BASE | tr -d ' '); \
	echo " Path counts: current=$$CUR_COUNT baseline=$$BASE_COUNT"; \
	REMOVED=$$(comm -23 $$TMP_BASE $$TMP_CUR || true); \
	if [ -n "$$REMOVED" ]; then echo " Breaking change: removed paths"; echo "$$REMOVED"; rm -f $$TMP_CUR $$TMP_BASE; exit 2; fi; \
	ADDED=$$(comm -13 $$TMP_BASE $$TMP_CUR || true); \
	echo " No removed paths (additive change)"; \
	if [ -n "$$ADDED" ]; then echo " Added paths:"; echo "$$ADDED"; fi; \
	rm -f $$TMP_CUR $$TMP_BASE; \
	echo " OpenAPI contract diff complete"

release-promo: ## Orchestrate secure release promotion pipeline (build→sign→sbom→smoke→stage→promote→prod)
	@if [ -z "$(RELEASE_VERSION)" ]; then echo " RELEASE_VERSION not set (e.g. make release-promo RELEASE_VERSION=1.2.3)"; exit 1; fi
	@echo " Starting release promotion pipeline for version $(RELEASE_VERSION)"; \
	echo " Step 1: Security gate"; $(MAKE) security-gate; \
	echo " Step 2: Build production binary"; $(MAKE) build-production; \
	echo " Step 2.5: Generate version.json"; $(MAKE) gen-version-json; \
	echo " Step 3: Build container image"; $(MAKE) docker-build; \
	echo "️  Step 4: Sign image (if cosign available)"; if command -v cosign >/dev/null 2>&1; then $(MAKE) cosign-sign image=costscope:latest || { echo "️  Image signing failed"; exit 1; }; else echo "ℹ️  cosign not installed, skipping signing"; fi; \
	echo " Step 5: Generate SBOM"; $(MAKE) sbom; \
	echo " Step 6: Smoke tests"; bash scripts/release/smoke.sh; \
	echo "  Step 7: Stage tagging"; docker tag costscope:latest costscope:stage-$(RELEASE_VERSION) || echo "(skipped docker tag)"; \
	echo " Step 8: Promotion tagging"; docker tag costscope:stage-$(RELEASE_VERSION) costscope:$(RELEASE_VERSION) || echo "(skipped docker tag)"; \
	echo "  Step 9: Git tag (v$(RELEASE_VERSION))"; if git rev-parse "v$(RELEASE_VERSION)" >/dev/null 2>&1; then echo "Tag v$(RELEASE_VERSION) already exists"; else git tag -a v$(RELEASE_VERSION) -m "Release v$(RELEASE_VERSION)"; fi; \
	echo " Step 10: Release notes"; RELEASE_VERSION=$(RELEASE_VERSION) bash scripts/release/generate-release-notes.sh; \
	echo " Step 11: Checksums"; shasum -a 256 bin/costscope-production > checksums.txt || echo "(checksum skipped)"; \
	echo " Release promotion pipeline complete"

SYFT_VERSION ?= v1.18.0
COSIGN_VERSION ?= v2.2.4

.PHONY: docker-build-release docker-build

docker-build-release: ## Build release binary then build release image with metadata labels
	@echo " Building release binary..."; \
	$(MAKE) --no-print-directory build-release; \
	VERSION_FLAG=$$(git describe --tags --always --dirty 2>/dev/null || echo dev); \
	COMMIT_FLAG=$$(git rev-parse --short=12 HEAD 2>/dev/null || echo none); \
	# Compute BUILD_DATE in a portable way. Prefer GNU date -d, fall back to python when needed.
	BUILD_DATE_FLAG=$$( \
		if [ -n "$$SOURCE_DATE_EPOCH" ]; then \
			( date -u -d @$$SOURCE_DATE_EPOCH +%Y-%m-%dT%H:%M:%SZ 2>/dev/null || date -u --date="@$$SOURCE_DATE_EPOCH" +%Y-%m-%dT%H:%M:%SZ 2>/dev/null || python -c "from datetime import datetime,os; print(datetime.utcfromtimestamp(int(os.environ.get('SOURCE_DATE_EPOCH'))).strftime('%Y-%m-%dT%H:%M:%SZ'))" ); \
		else \
			date -u +%Y-%m-%dT%H:%M:%SZ; \
		fi; \
	); \
	GO_VERSION_FLAG=$$(go version | awk '{print $$3}'); \
	# Build image using a Dockerfile that copies the prebuilt binary (Dockerfile.release)
	# Pass metadata as build args and also add labels for OCI metadata
	docker build -f Dockerfile.release \
	  --build-arg VERSION=$$VERSION_FLAG \
	  --build-arg COMMIT=$$COMMIT_FLAG \
	  --build-arg BUILD_DATE=$$BUILD_DATE_FLAG \
	  --build-arg GOVERSION=$$GO_VERSION_FLAG \
	  --build-arg SOURCE_REPO=$$(git config --get remote.origin.url || echo "https://github.com/your/repo") \
	  -t costscope:$$VERSION_FLAG . || (echo " Docker build failed" && exit 1); \
	echo " Docker image built: costscope:$$VERSION_FLAG"

.PHONY: docker-build
docker-build: docker-build-release ## Alias to docker-build-release


checksums: ## Generate sha256 checksums for release binaries -> checksums.txt
	@echo " Generating checksums (checksums.txt)..."; \
	OUT=checksums.txt; rm -f $$OUT; FOUND=0; \
	if [ -d dist ]; then find dist -type f -maxdepth 3 -name 'costscope-*' -exec sha256sum {} \; >> $$OUT || true; fi; \
	for f in costscope-* bin/costscope-production; do if [ -f "$$f" ] && [ ! -d "$$f" ]; then sha256sum "$$f" >> $$OUT; FOUND=1; fi; done; \
	if [ ! -s $$OUT ]; then echo "No release binaries found (build or run release workflow first)"; rm -f $$OUT; exit 1; fi; \
	sort -k2 $$OUT -o $$OUT; echo " checksums.txt written";

sign-checksums: ## Sign checksums.txt with cosign keyless
	@if [ ! -f checksums.txt ]; then echo "checksums.txt missing (run: make checksums)"; exit 1; fi
	@if ! command -v cosign >/dev/null 2>&1; then echo "Installing cosign $(COSIGN_VERSION)..."; curl -sSfL https://github.com/sigstore/cosign/releases/download/$(COSIGN_VERSION)/cosign-`uname -s | tr '[:upper:]' '[:lower:]'`-amd64 -o $(shell go env GOPATH)/bin/cosign && chmod +x $(shell go env GOPATH)/bin/cosign || go install github.com/sigstore/cosign/v2/cmd/cosign@$(COSIGN_VERSION); fi
	@echo "️  Signing checksums.txt (keyless)..."; COSIGN_EXPERIMENTAL=1 cosign sign-blob --yes --output-signature checksums.txt.sig checksums.txt; echo " Signature created: checksums.txt.sig"

verify-checksums: ## Verify signature for checksums.txt (keyless)
	@if [ ! -f checksums.txt ] || [ ! -f checksums.txt.sig ]; then echo "Missing checksums.txt or checksums.txt.sig"; exit 1; fi; COSIGN_EXPERIMENTAL=1 cosign verify-blob --signature checksums.txt.sig checksums.txt; echo " Signature verification passed"

supply-chain-all: sbom checksums sign-checksums verify-checksums ## Generate SBOM, checksums, sign & verify
	@echo "️  Supply chain artifact bundle complete"

release-notes: ## Generate release_notes.md from CHANGELOG diff (requires RELEASE_VERSION)
	@if [ -z "$(RELEASE_VERSION)" ]; then echo " RELEASE_VERSION not set"; exit 1; fi
	RELEASE_VERSION=$(RELEASE_VERSION) bash scripts/release/generate-release-notes.sh

# Act-friendly offline signing: produce a deterministic pseudo-signature and verify it
sign-checksums-act: ## Act mode: create a non-cryptographic pseudo-signature for checksums.txt (offline, no OIDC)
	@set -eu; \
	if [ ! -f checksums.txt ]; then echo "checksums.txt missing (run: make checksums)"; exit 1; fi; \
	echo "️  [act] Generating pseudo-signature for checksums.txt"; \
	if command -v sha256sum >/dev/null 2>&1; then \
		sha256sum checksums.txt | awk '{print $$1}' > checksums.txt.sig; \
	elif command -v shasum >/dev/null 2>&1; then \
		shasum -a 256 checksums.txt | awk '{print $$1}' > checksums.txt.sig; \
	else \
		echo "No sha256 tool available (sha256sum/shasum)"; exit 1; \
	fi; \
	echo " [act] Wrote checksums.txt.sig (sha256 of checksums.txt)";

verify-checksums-act: ## Act mode: verify pseudo-signature matches sha256 of checksums.txt
	@set -eu; \
	if [ ! -f checksums.txt ] || [ ! -f checksums.txt.sig ]; then echo "Missing checksums.txt or checksums.txt.sig"; exit 1; fi; \
	if command -v sha256sum >/dev/null 2>&1; then \
		expected=$$(sha256sum checksums.txt | awk '{print $$1}'); \
	elif command -v shasum >/dev/null 2>&1; then \
		expected=$$(shasum -a 256 checksums.txt | awk '{print $$1}'); \
	else \
		echo "No sha256 tool available (sha256sum/shasum)"; exit 1; \
	fi; \
	actual=$$(tr -d '\n\r' < checksums.txt.sig); \
	if [ "$$expected" != "$$actual" ]; then echo "[act] Pseudo-signature mismatch"; echo " expected: $$expected"; echo "   actual: $$actual"; exit 2; fi; \
	echo " [act] Pseudo-signature verification passed";
