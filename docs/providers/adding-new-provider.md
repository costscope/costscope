---
title: Adding a New Provider
description: Scaffold, implement, test, and register a new cloud provider integration.
---

# Adding a New Cloud Provider

This guide explains how to add a new cloud provider integration to CostScope in a modular, repeatable way.

## Overview

Providers implement cost ingestion + FOCUS mapping through a narrow interface and are registered via the ProviderManager. Keep logic separated into: API client, raw data ingestion/streaming, field mapping (FOCUS), normalization, and tests (unit + sample fixtures).

## Quick Start Diagram

Below is a high‑level visual of the recommended fast path when introducing a new provider. Use the scaffold generator, fill stubs, wire self‑registration, add tests + fixtures, then run quality gates and documentation updates.

```mermaid
flowchart LR
    A[Scaffold<br/>make gen-provider name=<provider>] --> B[Fill Stubs<br/>provider.go / headers.go]
    B --> C[Implement Mapper & Converter<br/>mapper.go / converter.go]
    C --> D[Add Tests & Fixtures<br/>mapper/converter table tests]
    D --> E[Self-Register<br/>init_registry.go]
    E --> F[Docs Update<br/>README + focus_conversion + this guide]
    F --> G[Quality Gates<br/>make test perf-bench lint]
    G --> H[Parity & Invariants (opt)<br/>perf-parity / invariants]
    H --> I[Commit & PR]
```

ASCII fallback (if Mermaid rendering not available):

```
Scaffolder → Fill Stubs → Mapper/Converter → Tests/Fixtures → Self-Register → Docs Update → Quality Gates (test | perf-bench | lint) → Commit/PR
```

Key guardrails (always apply): unified logging, any SQL via QueryBuilder, discount/credit normalization consistent with existing providers, optional invariants hashing for regression safety.

## Directory Layout

```
internal/providers/<provider>/
  provider.go            # Implements types.CloudProvider
  ingest/                # (optional) Input readers / pagination / streaming
  focusmapper/           # Provider -> FOCUS mapping helpers
  normalizer/            # Classification, discount/credit normalization
  testdata/              # Small synthetic CSV/JSON fixtures (non‑sensitive)
```

If the provider needs complex mapping, mirror existing pattern under `internal/core/focus/conversion/<provider>/` and keep provider package thin (registration + health).

## Steps

You can either follow the manual steps below or bootstrap a ready-to-fill scaffold with the generator.

### Fast Path: Scaffold Generator (Recommended)

Run:

```bash
make gen-provider name=<provider>
# or directly
go run ./scripts/tools/gen-provider -name <provider>
```

Generated layout:

```
internal/providers/<provider>/
   README.md
   config.go
   provider.go
   register.go          # registry self-registration (minimal instance)
   provider_test.go     # basic sanity test
   doc.go
internal/core/focus/conversion/<provider>/
   README.md
   headers.go           # placeholder for expected raw fields
   mapper.go            # stub record mapper
   converter.go         # streaming converter stub
   converter_test.go
```

Follow-up after generation:

1. Replace stub logic in `provider.go` (credentials validation, metadata, data retrieval).
2. Populate `headers.go` with actual input headers (order matters for CSV readers).
3. Implement field mapping in `mapper.go` and conversion/common helpers.
4. Flesh out streaming parsing & invariants in `converter.go` (mirror aws/azure/gcp patterns).
5. Add table-driven tests for mapper & converter (discount/credit classification, negative usage, time parsing, normalization).
6. Provide small deterministic fixtures (place under a new `testdata/` folder if needed).
7. Update docs (README provider list, `docs/architecture/focus-conversion.md` capabilities, and this file with nuances).
8. Run the local quality gates (tests, lint, perf). Example:

```bash
make test
make lint
make perf-bench
```

To overwrite existing (unsafe) files use `-force-dry-run` to preview without writing.

### Manual Steps (Alternative)

1. Define credentials struct & config additions (extend `types.ProviderConfig` if needed with additive fields; avoid breaking changes).
2. Implement the provider interface and wiring under `internal/providers/<provider>`.
3. Add FOCUS mapping: create converter modules under `internal/core/focus/conversion/<provider>/` (reader, mapper, normalizer).
4. Register provider in `ProviderManager` (or dynamic auto-discovery later) inside an init function or explicit registration function.
5. Add unit tests:
   - Credential validation edge cases
   - Health check / connectivity (mock HTTP or SDK)
   - Mapping tests (cover discounts, credits, zero / negative usage, timestamp parsing)
6. Add sample fixture(s) (redacted, minimal) under `testdata/` and a conversion test invoking unified conversion path.
7. Update docs:
   - README: list new provider
   - `docs/architecture/focus-conversion.md`: add provider capabilities
   - This guide with any provider-specific nuances
8. Run quality gates: tests, lint and perf checks.

### Registry-Based Self-Registration (New Path)

Providers now self-register via a lightweight registry to eliminate central switch edits. The previous switch in `ProviderManager.CreateProvider` remains temporarily as a fallback; new providers SHOULD use the registry path only.

Implementation steps (augmenting above):

1. Ensure your provider struct implements Name() returning the canonical lowercase key (e.g. `"acme"`).
2. Add `init_registry.go` under `internal/providers/<provider>/` with a registration function calling `registry.Register("<provider>", factory)`.
3. Avoid heavy logic in `init()`; keep registration light.
4. Run `go test ./internal/providers/...` to confirm no duplicate key errors.

Guidelines:

- Registration key must stay stable; changing it is a breaking change for existing configs.
- Keep constructors pure (no network calls); defer I/O to explicit `ValidateCredentials` or later health probes.

## Testing Strategy

- Small fast unit tests only; large integration tests behind build tag or env guards.
- Use deterministic fixtures (no dates that vary daily unless frozen via constant).
- Validate aggregate parity against unified mapper if implemented.

## Checklist

Use this expanded checklist before opening a PR (order roughly matches the quick start diagram):

### Scaffold & Structure

- [ ] Scaffold generated (or manual layout created)

```bash
make gen-provider name=<provider>
# or
go run ./scripts/tools/gen-provider -name <provider>
```

### Implementation

- [ ] `provider.go` implements types.CloudProvider (pure constructor, no network side-effects)
- [ ] Credentials / config struct added (additive only; validated via `ValidateCredentials`)
- [ ] `headers.go` filled with real raw input headers (stable ordering for CSV streaming)
- [ ] `mapper.go` maps required FOCUS v1.2 fields (cost, usage, service, resource, timestamps)
- [ ] Normalization / classification (Discount vs Credit, negative usage rules) implemented & logged
- [ ] Optional: invariants aggregation integrated (if large streaming conversion implemented)

---

title: Adding a New Provider
description: Scaffold, implement, test, and register a new cloud provider integration.

---

{{/* Original content follows (lightly untouched) */}}

# Adding a New Cloud Provider

This guide explains how to add a new cloud provider integration to CostScope in a modular, repeatable way.

## Overview

Providers implement cost ingestion + FOCUS mapping through a narrow interface and are registered via the ProviderManager. Keep logic separated into: API client, raw data ingestion/streaming, field mapping (FOCUS), normalization, and tests (unit + sample fixtures).

## Quick Start Diagram

Below is a high‑level visual of the recommended fast path when introducing a new provider. Use the scaffold generator, fill stubs, wire self‑registration, add tests + fixtures, then run quality gates and documentation updates.

```mermaid
flowchart LR
    A[Scaffold<br/>make gen-provider name=&lt;provider&gt;] --> B[Fill Stubs<br/>provider.go / headers.go]
    B --> C[Implement Mapper & Converter<br/>mapper.go / converter.go]
    C --> D[Add Tests & Fixtures<br/>mapper/converter table tests]
    D --> E[Self-Register<br/>init_registry.go]
    E --> F[Docs Update<br/>README + focus_conversion + this guide]
    F --> G[Quality Gates<br/>make test perf-bench lint]
    G --> H[Parity & Invariants (opt)<br/>perf-parity / invariants]
    H --> I[Commit & PR]
```

ASCII fallback (if Mermaid rendering not available):

```
 Scaffolder → Fill Stubs → Mapper/Converter → Tests/Fixtures → Self-Register
         → Docs Update → Quality Gates (test | perf-bench | lint)
            → (Parity / Invariants optional) → Commit PR
```

Key guardrails (always apply): unified logging (`logging.Loggercostscope-data/`, any SQL via QueryBuilder, discount/credit normalization consistent with existing providers, optional invariants hashing for regression safety.

```bash
), YAML + precedence resolvers (no ad-hoc env lookups), no hardcoded secrets, Parquet output under
```

## Directory Layout

```
internal/providers/<provider>/
  provider.go            # Implements types.CloudProvider
  ingest/                # (optional) Input readers / pagination / streaming
  focusmapper/           # Provider -> FOCUS mapping helpers
  normalizer/            # Classification, discount/credit normalization
  testdata/              # Small synthetic CSV/JSON fixtures (non‑sensitive)
```

If the provider needs complex mapping, mirror existing pattern under `internal/core/focus/conversion/<provider>/` and keep provider package thin (registration + health).

## Steps

You can either follow the manual steps below or bootstrap a ready-to-fill scaffold with the generator.

### Fast Path: Scaffold Generator (Recommended)

Run:

```bash
make gen-provider name=<provider>
# or directly
go run ./scripts/tools/gen-provider -name <provider>
```

Generated layout:

```
internal/providers/<provider>/
   README.md
   config.go
   provider.go
   register.go          # registry self-registration (minimal instance)
   provider_test.go     # basic sanity test
   doc.go
internal/core/focus/conversion/<provider>/
   README.md
   headers.go           # placeholder for expected raw fields
   mapper.go            # stub record mapper
   converter.go         # streaming converter stub
   converter_test.go
```

Follow-up after generation:

1. Replace stub logic in `provider.go` (credentials validation, metadata, data retrieval).
2. Populate `headers.go` with actual input headers (order matters for CSV readers).
3. Implement field mapping in `mapper.goconversion/common`).

```bash
(use helpers in
```

4. Flesh out streaming parsing & invariants in `converter.go` (mirror aws/azure/gcp patterns).
5. Add table-driven tests for mapper & converter (discount/credit classification, negative usage, time parsing, normalization).
6. Provide small deterministic fixtures (place under a new `testdata/` folder if needed).
7. Update docs (README provider list, `docs/architecture/focus-conversion.md` capabilities, and this file with nuances).
8. Run the local quality gates (tests, lint, perf). Example:

```bash
make test
make lint
make perf-bench
```

To overwrite existing (unsafe) files use `-force-dry-run` to preview without writing.

```bash
. Use
```

### Manual Steps (Alternative)

1. Define credentials struct & config additions (extend `types.ProviderConfig` if needed with additive fields; avoid breaking changes).
2. Implement the `types.CloudProvider<provider>/provider.go`.

```bash
interface in
```

3. Add FOCUS mapping: create converter modules under `internal/core/focus/conversion/<provider>/` (reader, mapper, normalizer) following aws/azure/gcp layout.
4. Register provider in `ProviderManager` (or dynamic auto-discovery later) inside an init function or explicit registration function.
5. Add unit tests:
   - Credential validation edge cases
   - Health check / connectivity (mock HTTP or SDK)
   - Mapping tests (cover discounts, credits, zero usage, timestamp parsing)
6. Add sample fixture(s) (redacted, minimal) under `testdata/` and a conversion test invoking unified conversion path.
7. Update docs:
   - README: list new provider
   - `docs/architecture/focus-conversion.md`: add provider capabilities
   - This guide with any provider-specific nuances
8. Run quality gates: tests, lint and perf checks (example below).

```bash
make test
make lint
make perf-bench
```

### Registry-Based Self-Registration (New Path)

Providers now self-register via a lightweight registry to eliminate central switch edits. The previous switch in `ProviderManager.CreateProvider` remains temporarily as a fallback; new providers SHOULD use the registry path only.

Implementation steps (augmenting above):

1. Ensure your provider struct implements returning the canonical lowercase key (e.g. `"acme"`).

```bash
Name() string
```

2. Add `init_registry.gointernal/providers/<provider>/`:

```bash
under
```

```go
package <provider>

import (
   "github.com/costscope/costscope/internal/providers/registry"
   "github.com/costscope/costscope/internal/providers/types"
)

func init() {
   registry.Register("<provider>", func(cfg *types.ProviderConfig) (registry.Provider, error) {
      return New<ProviderCamelCase>Provider(cfg)
   })
}
```

3. Avoid heavy logic in `init()registry.Register`.

```bash
; only call
```

4. Run `go test ./internal/providers/...` to confirm no duplicate key errors.

Fallback Removal Plan:

- Phase 1 (current): Registry preferred, switch fallback present.
- Phase 2: Telemetry/metrics (optional) to verify zero fallback usage.
- Phase 3: Remove switch; registry becomes mandatory.

Guidelines:

- Registration key must stay stable; changing it is a breaking change for existing configs.
- Keep constructors pure (no network calls); defer I/O to explicit `ValidateCredentials` or later health probes.
- If additional validation is needed at registration time, wrap the provider or perform lazy checks when first used.

Errors:

- Duplicate key: `registry.Register` returns an error (panic avoided). Adjust your key and retry.
- Missing registry entry (after switch removal): `ProviderManager` will return an unsupported error—add an init file or import path.

After Phase 3, adding a new provider becomes: create package, implement interface + Name, add `init_registry.go`, write tests, update docs—no core manager edits required.

## Interface Contract (types.CloudProvider)

## Key methods (verify in `internal/providers/types`):

```bash
ValidateCredentials(ctx, creds map[string]string) error
```

-

```bash
GetMetadata(ctx) (*ProviderMetadata, error)
```

- `ListAccounts/Projects` (if available) – keep additive

Return domain errors (wrapped) with sentinel classifications for retry vs auth.

## FOCUS Mapping Guidelines

- Do not write Parquet directly from provider package—use conversion orchestrators.
- Keep classification (Discount vs Credit) consistent with existing provider rules.
- Emit structured logs for precedence and normalization decisions.
- Add Prometheus counters for any new normalization heuristics.

## Testing Strategy

- Small fast unit tests only; large integration tests behind build tag or env guards.
- Use deterministic fixtures (no dates that vary daily unless frozen via constant).
- Validate aggregate parity against unified mapper if implemented.

## Modular Future Enhancements

Planned improvement: plugin registry allowing out-of-tree providers compiled with Go plugins or module discovery. For now, keep code additive and isolated to ease future extraction.

## Checklist

Use this expanded checklist before opening a PR (order roughly matches the quick start diagram):

### Scaffold & Structure

- [ ] Scaffold generated (or manual layout created)

```bash
make gen-provider name=<provider>
# or
go run ./scripts/tools/gen-provider -name <provider>
```

- [ ] Package path: `internal/providers/<provider>` (lowercase, stable key)
- [ ] FOCUS converter stubs under `internal/core/focus/conversion/<provider>/`

### Implementation

- [ ] `provider.gotypes.CloudProvider` (pure constructor, no network side-effects)

```bash
implements
```

- [ ] Credentials / config struct added (additive only; validated via `ValidateCredentials`)
- [ ] `headers.go` filled with real raw input headers (stable ordering for CSV streaming)
- [ ] `mapper.go` maps required FOCUS v1.2 fields (cost, usage, service, resource, timestamps)
- [ ] Normalization / classification (Discount vs Credit, negative usage rules) implemented & logged
- [ ] Optional: invariants aggregation integrated (if large streaming conversion implemented)

### Self-Registration & Config

- [ ] `init_registry.goregistry.Register("<provider>", factory)` (no heavy logic)

```bash
calls
```

- [ ] `Name()` returns canonical key; matches registration key
- [ ] No direct env parsing—uses config precedence resolvers for any tunables
- [ ] All logging via unified `logging.Logger`

### Tests & Fixtures

- [ ] Table-driven mapper tests (cover: discount, credit, zero / negative usage, timestamp formats)
- [ ] Converter streaming test (chunking + rotation if applicable)
- [ ] Credentials validation tests (edge cases, missing fields)
- [ ] Deterministic small fixtures in `testdata/` (no secrets, stable numeric/time values)
- [ ] (Optional) Parity test vs unified mapper (if unified path added)

### Docs & Observability

- [ ] README provider list updated
- [ ] `docs/architecture/focus-conversion.md` capabilities section updated
- [ ] This guide (`docs/providers/adding-new-provider.md`) updated with provider-specific nuances
- [ ] Any new normalization emits Prometheus counters (namespaced, low cardinality)
- [ ] Tracing spans added for mapping / conversion phases (follow existing providers)

### Quality & Performance

-- [ ] passes (unit & integration without flakes)

```bash
make test
```

-- [ ] passes (golangci-lint, formatting)

```bash
make lint
```

-- [ ] within accepted ratios (if mapper performance-sensitive)

```bash
make perf-bench
```

```bash
make test
make lint
make perf-bench
make perf-parity # optional
```

- [ ] No hardcoded secrets; credentials sourced via config/env

### Finalization

- [ ] CHANGELOG entry (feat: provider <name>) prepared (optional pre-1.0)
- [ ] New files added to CODEOWNERS if ownership differs
- [ ] All CI contract guards pass (API spec unchanged unless intentionally extended)
- [ ] Commit hooks run without `--no-verify`

If any item is intentionally deferred (e.g., invariants), note justification in the PR description.

---

Questions? Open an issue with subject .

```bash
provider: <name>
```
