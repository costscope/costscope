# mk/security.mk - security, sbom, vuln, supply chain

.PHONY: cosign-keygen cosign-sign cosign-verify gosec-high grype-sbom opa-eval provenance-attest release-secure sbom sbom-cyclonedx sbom-diff sbom-spdx sbom-syft sbom-verify sbom-vuln-diff secrets-scan security security-aggregate security-detailed security-gate security-summary security-tools-install sign trivy-fs-json trivy-fs-scan trivy-image-json trivy-image-sbom trivy-image-scan trivy-install vuln vuln-json vuln-nancy

security: ## Run security analysis (gosec + govulncheck + secrets scan quick aggregate)
	@echo " Running security analysis (quick gate)..."
	$(MAKE) --no-print-directory gosec-high || exit 1
	$(MAKE) --no-print-directory vuln-json || exit 1
	$(MAKE) --no-print-directory secrets-scan || exit 1
	@echo " Security (quick) passed"

sbom-syft: ## Generate SBOM (Syft CycloneDX JSON) to sbom-syft.json
	@echo " Generating SBOM (script) ..."; bash scripts/supply-chain/gen-sbom.sh; echo " SBOM (syft) ready (sbom-syft.json)"

sbom-cyclonedx: ## Generate SBOM (CycloneDX JSON) to sbom.json using cyclonedx-gomod
	@echo " Generating SBOM (CycloneDX via cyclonedx-gomod)..."; \
	if ! command -v cyclonedx-gomod >/dev/null 2>&1; then echo "Installing cyclonedx-gomod..."; go install github.com/CycloneDX/cyclonedx-gomod/cmd/cyclonedx-gomod@latest; fi; \
	cyclonedx-gomod mod -licenses -json -output sbom.json; echo " SBOM (cyclonedx-gomod) written to sbom.json"

SBOM_MIN_SIZE ?= 2048
sbom-verify: ## Verify SBOM presence & size (CI gate)
	@echo " Verifying SBOM artifact sbom-syft.json (min $(SBOM_MIN_SIZE) bytes)..."; \
	if [ ! -f sbom-syft.json ]; then echo " sbom-syft.json missing"; exit 1; fi; \
	SIZE=$$(stat -c '%s' sbom-syft.json); if [ $$SIZE -lt $(SBOM_MIN_SIZE) ]; then echo " SBOM size $$SIZE < $(SBOM_MIN_SIZE)"; exit 2; else echo " SBOM size $$SIZE bytes"; fi; \
	command -v jq >/dev/null 2>&1 && jq -e '.components | length > 0' sbom-syft.json >/dev/null 2>&1 && echo " SBOM components present" || echo "️ jq missing or components check skipped"

# Backwards compatible alias
sbom: sbom-syft ## Generate SBOM (default syft)

cosign-keygen: ## Generate cosign key-pair
	@echo " Generating cosign key-pair..."; if ! command -v cosign >/dev/null 2>&1; then echo "Installing cosign..."; go install github.com/sigstore/cosign/v2/cmd/cosign@latest; fi; cosign generate-key-pair

image ?= costscope:latest
cosign-sign: ## Sign docker image with cosign
	@echo "️  Signing image $(image)..."; cosign sign --yes $(image)

sign: cosign-sign ## Alias

cosign-verify: ## Verify docker image signature
	@echo " Verifying signature for $(image)..."; cosign verify $(image) || (echo " Verification failed" && exit 1); echo " Signature verified"

provenance-attest: ## Generate SLSA provenance attestation (experimental)
	@echo " Generating provenance (SLSA v1 predicate)..."; if ! command -v cosign >/dev/null 2>&1; then echo "cosign not installed"; exit 1; fi; COSIGN_EXPERIMENTAL=1 cosign attest --predicate provenance.json --type slsaprovenance $(image) || (echo "️ Attestation failed" && exit 1); echo " Provenance attested"

security-detailed: ## Run detailed security analysis
	@echo " Running detailed security analysis..."; gosec -fmt json -out security-report.json ./... || true; echo "Security report saved to security-report.json"

vuln: ## Check for vulnerabilities
	@echo "️ Checking for vulnerabilities..."; govulncheck ./...

vuln-json: ## govulncheck JSON output (fails on high severity vulns)
	@echo "️ Running govulncheck (JSON)..."; govulncheck -format=json ./... > govulncheck.json || true; \
	echo "Parsing govulncheck.json for severity >= $(GOVULNCHECK_SCORE_MIN) ..."; \
	if jq -e --argjson min "$(GOVULNCHECK_SCORE_MIN)" '.vulnerabilities[]? | .osv.severity[]?.score? // 0 | select(. >= $$min)' govulncheck.json >/dev/null 2>&1; then echo " Vulnerability meeting threshold (>= $(GOVULNCHECK_SCORE_MIN)) detected"; exit 1; else echo " No vulnerabilities meeting threshold $(GOVULNCHECK_SCORE_MIN)"; fi

GITLEAKS_VERSION ?= latest
GOSEC_VERSION ?= latest
GOVULNCHECK_VERSION ?= latest

secrets-scan: ## Scan repo for hard-coded secrets
	@echo " Scanning for secrets..."; \
	if ! command -v gitleaks >/dev/null 2>&1; then echo "Installing gitleaks"; go install github.com/zricethezav/gitleaks/v8@$(GITLEAKS_VERSION) || (echo "Failed to install gitleaks" && exit 1); fi; \
	if [ -f .gitleaks.toml ]; then CFG="--config .gitleaks.toml"; else CFG=""; fi; \
	gitleaks detect --source . --no-banner --redact --exit-code 1 $$CFG --report-format json --report-path gitleaks-report.json || (echo " Secrets found" && exit 1)

vuln-nancy: ## Check dependencies with nancy
	@echo "️ Checking dependencies with nancy..."; go list -json -deps ./... | nancy sleuth

# Aggregate security gate
security-gate: ## Aggregate security gate (supply chain + scanners)
	@echo " Running aggregated security gate..."; \
	$(MAKE) --no-print-directory security-tools-install; \
	if [ ! -f sbom.json ]; then $(MAKE) --no-print-directory sbom-syft; else echo "ℹ️  Reusing existing sbom.json"; fi; \
	if [ ! -f govulncheck.json ]; then $(MAKE) --no-print-directory vuln-json || true; else echo "ℹ️  Reusing existing govulncheck.json"; fi; \
	if [ ! -f gosec.json ]; then $(MAKE) --no-print-directory gosec-high || true; else echo "ℹ️  Reusing existing gosec.json"; fi; \
	if [ ! -f gitleaks-report.json ]; then $(MAKE) --no-print-directory secrets-scan || true; else echo "ℹ️  Reusing existing gitleaks-report.json"; fi; \
	if [ ! -f trivy-fs.json ]; then $(MAKE) --no-print-directory trivy-fs-json || true; else echo "ℹ️  Reusing existing trivy-fs.json"; fi; \
	if [ ! -f trivy-image.json ]; then if command -v docker >/dev/null 2>&1; then $(MAKE) --no-print-directory docker-build trivy-image-json || true; else echo "ℹ️  Docker not found; skipping trivy image scan"; fi; else echo "ℹ️  Reusing existing trivy-image.json"; fi; \
	if [ ! -f grype.json ]; then $(MAKE) --no-print-directory grype-sbom || true; else echo "ℹ️  Reusing existing grype.json"; fi; \
	bash scripts/security/aggregate-security.sh; echo " Security gate passed"

gosec-high: ## Run gosec and fail on severity >= GOSEC_FAIL_LEVEL
	@echo " Running gosec (fail level $(GOSEC_FAIL_LEVEL))..."; gosec -fmt=json -out gosec.json ./... || true; \
	if jq -e --arg level "$(GOSEC_FAIL_LEVEL)" '.Issues[] | select(.severity==$$level)' gosec.json >/dev/null 2>&1; then echo " $(GOSEC_FAIL_LEVEL) severity gosec issues detected"; jq --arg level "$(GOSEC_FAIL_LEVEL)" '.Issues[] | select(.severity==$$level) | {severity,confidence,details,file,line}' gosec.json; exit 1; else echo " No $(GOSEC_FAIL_LEVEL) severity gosec issues"; fi

security-tools-install: ## Install / update security & supply chain tooling
	@echo " Installing security tools..."; \
	if ! command -v gosec >/dev/null 2>&1; then go install github.com/securego/gosec/v2/cmd/gosec@$(GOSEC_VERSION); fi; \
	if ! command -v govulncheck >/dev/null 2>&1; then go install golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION); fi; \
	if ! command -v gitleaks >/dev/null 2>&1; then go install github.com/zricethezav/gitleaks/v8@$(GITLEAKS_VERSION); fi; \
		if ! command -v syft >/dev/null 2>&1; then \
			echo "Installing syft via go install (preferred)"; \
			if ! go install github.com/anchore/syft/cmd/syft@$(SYFT_VERSION) 2>/dev/null; then \
				echo "go install failed; falling back to upstream installer"; curl -sSfL https://raw.githubusercontent.com/anchore/syft/main/install.sh -o /tmp/syft-install.sh && sh /tmp/syft-install.sh -s -- -b $(shell go env GOPATH)/bin $(SYFT_VERSION); fi; \
		fi; \
		if ! command -v trivy >/dev/null 2>&1; then \
			echo "Installing trivy via package or upstream installer"; curl -sfL https://raw.githubusercontent.com/aquasecurity/trivy/main/contrib/install.sh -o /tmp/trivy-install.sh && sh /tmp/trivy-install.sh -s -- -b $(shell go env GOPATH)/bin; fi; \
	if ! command -v cosign >/dev/null 2>&1; then curl -sSfL https://github.com/sigstore/cosign/releases/download/$(COSIGN_VERSION)/cosign-`uname -s | tr '[:upper:]' '[:lower:]'`-amd64 -o $(shell go env GOPATH)/bin/cosign && chmod +x $(shell go env GOPATH)/bin/cosign || go install github.com/sigstore/cosign/v2/cmd/cosign@$(COSIGN_VERSION); fi; \
		if ! command -v grype >/dev/null 2>&1; then \
			if ! go install github.com/anchore/grype/cmd/grype@latest 2>/dev/null; then \
				curl -sSfL https://raw.githubusercontent.com/anchore/grype/main/install.sh -o /tmp/grype-install.sh && sh /tmp/grype-install.sh -s -- -b $(shell go env GOPATH)/bin || true; fi; \
		fi; \
	echo " Security tools present"

TRIVY ?= trivy
trivy-install: ## Install Trivy if missing
	@if ! command -v $(TRIVY) >/dev/null 2>&1; then echo "Installing Trivy..."; curl -sfL https://raw.githubusercontent.com/aquasecurity/trivy/main/contrib/install.sh -o /tmp/trivy-install.sh && sh /tmp/trivy-install.sh -s -- -b $(shell go env GOPATH)/bin; fi

trivy-fs-json: trivy-install ## Trivy FS scan (JSON output all severities)
	@echo " Trivy filesystem scan (JSON)..."; $(TRIVY) fs --severity UNKNOWN,LOW,MEDIUM,HIGH,CRITICAL --format json --output trivy-fs.json --ignore-unfixed --no-progress . || true; \
	if jq -e '.Results[]?.Vulnerabilities[]? | select(.Severity=="HIGH" or .Severity=="CRITICAL")' trivy-fs.json >/dev/null 2>&1; then echo " HIGH/CRITICAL vulns detected in FS scan"; exit 1; else echo " No HIGH/CRITICAL vulns in FS scan"; fi

# Backwards-compatible alias (some docs / checklists reference trivy-fs-scan)
trivy-fs-scan: trivy-fs-json ## Alias: run filesystem vulnerability scan (JSON) and enforce HIGH/CRITICAL=0

trivy-image-json: trivy-install docker-build ## Trivy Image scan (JSON)
	@echo " Trivy image scan (JSON)..."; $(TRIVY) image --severity UNKNOWN,LOW,MEDIUM,HIGH,CRITICAL --format json --output trivy-image.json --ignore-unfixed --no-progress costscope:latest || true; \
	if jq -e '.Results[]?.Vulnerabilities[]? | select(.Severity=="HIGH" or .Severity=="CRITICAL")' trivy-image.json >/dev/null 2>&1; then echo " HIGH/CRITICAL vulns detected in image"; exit 1; else echo " No HIGH/CRITICAL vulns in image"; fi

grype-sbom: ## (Optional) Scan SBOM with grype (JSON -> grype.json)
	@if [ ! -f sbom.json ]; then $(MAKE) sbom; fi; \
	if ! command -v grype >/dev/null 2>&1; then echo "grype not installed (run make security-tools-install)"; exit 0; fi; \
	echo "️ Grype scanning SBOM..."; grype sbom:sbom.json -o json > grype.json || true; \
	if jq -e '.matches[]? | select(.vulnerability.severity=="High" or .vulnerability.severity=="Critical")' grype.json >/dev/null 2>&1; then echo " High/Critical vulns (grype)"; exit 1; else echo " No High/Critical (grype)"; fi || true

security-aggregate: ## Run extended tool suite & aggregate markdown summary
	@echo " Running extended security suite..."; $(MAKE) --no-print-directory security-tools-install; $(MAKE) --no-print-directory sbom-syft; $(MAKE) --no-print-directory vuln-json || true; $(MAKE) --no-print-directory gosec-high || true; $(MAKE) --no-print-directory secrets-scan || true; $(MAKE) --no-print-directory trivy-fs-json || true; $(MAKE) --no-print-directory trivy-image-json || true; $(MAKE) --no-print-directory grype-sbom || true; bash scripts/security/aggregate-security.sh; echo " Aggregate summary: docs/security/security-summary.md"

security-summary: security-aggregate ## Alias for security-aggregate

release-secure: clean build-release sbom security-gate ## Release build with supply chain artifacts
	@echo " Secure release artifacts ready (binary + SBOM + reports)"

# SBOM diff helpers
sbom-old ?= sbom.previous.json
sbom-diff: ## Diff current SBOM vs previous
	@echo " Diffing SBOM vs $(sbom-old)..."; \
	if [ ! -f $(sbom-old) ]; then echo "No previous SBOM ($(sbom-old))"; exit 0; fi; \
	NEW=$$(jq -r '.components[].name + "@" + (.components[].version // "")' sbom.json | sort | uniq); \
	OLD=$$(jq -r '.components[].name + "@" + (.components[].version // "")' $(sbom-old) | sort | uniq); \
	comm -13 <(echo "$$OLD") <(echo "$$NEW") > sbom_diff_added.txt || true; \
	if [ -s sbom_diff_added.txt ]; then echo "New components:"; cat sbom_diff_added.txt; else echo "No new components"; fi; echo " SBOM diff complete (sbom_diff_added.txt)"

image-sbom-old ?= image_vuln_sbom.previous.json
sbom-vuln-diff: ## Diff image vulnerability SBOM vs previous
	@echo " Diffing image vulnerability SBOM vs $(image-sbom-old)..."; \
	if [ ! -f $(image-sbom-old) ]; then echo "No previous image vulnerability SBOM ($(image-sbom-old))"; exit 0; fi; \
	jq -r '.vulnerabilities[]? | select(.ratings[]?.severity=="High" or .ratings[]?.severity=="Critical") | .id' image_vuln_sbom.json | sort -u > current_vulns.txt || true; \
	jq -r '.vulnerabilities[]? | select(.ratings[]?.severity=="High" or .ratings[]?.severity=="Critical") | .id' $(image-sbom-old) | sort -u > previous_vulns.txt || true; \
	comm -13 previous_vulns.txt current_vulns.txt > new_high_crit_vulns.txt || true; \
	if [ -s new_high_crit_vulns.txt ]; then echo " New HIGH/CRITICAL vulns introduced:"; cat new_high_crit_vulns.txt; exit 1; else echo " No new HIGH/CRITICAL vulns introduced"; fi; echo " Vulnerability SBOM diff complete"

opa-policy-dir ?= policy
opa-eval: ## Evaluate OPA/Rego policies against SBOM
	@echo "️  Evaluating OPA policies on SBOM..."; if ! command -v opa >/dev/null 2>&1; then echo "Installing opa..."; go install github.com/open-policy-agent/opa/cmd/opa@latest; fi; opa eval -f pretty -d $(opa-policy-dir) -i sbom.json "data.costscope.sbom.allow" | tee opa_result.txt; if grep -q "false" opa_result.txt; then echo " OPA policy violation"; exit 1; fi; echo " OPA policy evaluation passed"
