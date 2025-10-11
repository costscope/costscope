---
title: ADR Process
description: How to write, review and record Architectural Decision Records (ADRs) for CostScope.
last_reviewed: 2025-09-08
---

# ADRs — Architecture Decision Records

A short, versioned record of a significant architectural decision: the problem, the chosen solution, and its consequences.

This document describes the minimal workflow for creating, reviewing, and recording ADRs in this repository.

## Checklist — creating a new ADR
1. Create a file under `docs/adr/XXXX-short-title.mdXXXX0003-new-jobstore.md`).

```bash
named
```

```bash
where
```

```bash
is a zero-padded sequence number (e.g.
```
2. Add YAML front matter with `titledatestatusdescription` (see template below).

```bash
,
```

```bash
and a short
```
3. Document the Context, the Decision, and the Consequences (including rollback steps and CI requirements).
4. Open a PR, tag relevant teams (architecture, SRE, Security) and label the PR with `docsadr`.

```bash
and
```
5. After discussion, update `status:acceptedrejectedsupersededproposed` while in review.

```bash
to
```

```bash
or keep
```
6. Cross-link the ADR from related docs (architecture, runbooks, release notes) when applicable.

## ADR statuses
- proposed — under discussion, in a PR
- accepted — decision recorded and implemented (or scheduled)
- rejected — decision not accepted; record kept for historical context
- superseded — replaced by a newer ADR (include a link to the replacing ADR)

## PR & review expectations
- Assign 2+ reviewers: at least one architect or senior engineer and one SRE/security reviewer for infra/security-impacting ADRs.
- In the PR description include the motivation, key risks, and rollback notes.
- When accepted, update the `status` in the ADR file and merge the PR.

## ADR template
Copy the content below into a new file and fill in the sections.

```md
---
title: ADR XXXX - <Short Title>
date: 2025-09-08
status: proposed
description: Short one-line summary of the decision
---

# ADR XXXX: <Short Title>

## Context
Describe the problem, constraints, and why the current approach is unsatisfactory.

## Decision
Describe the chosen approach. Include configuration, interfaces, and implementation notes.

## Consequences
List what changes: operational steps, CI requirements, rollback plan, observability, and security implications.

## Alternatives considered
- Alternative A — reason not chosen
- Alternative B — reason not chosen

## Links
- Link to PR
- Related documents

```

## Minimal example

```md
---
title: ADR 0001 - Unified Mapper Adoption
date: 2025-09-08
status: accepted
description: Consolidate mapper implementations and enforce parity/perf checks in CI.
---

# ADR 0001: Unified Mapper Adoption

## Context
Two parallel mapper implementations increased maintenance burden and introduced divergence.

## Decision
Merge `fast` and `experimental` mappers into a single `unified` mapper. Require parity and perf benchmarks in CI for changes touching the conversion path.

## Consequences
- Maintain versioned baseline artifacts and update them through a controlled process.
- PRs touching conversion must run parity and perf checks; failure blocks merge.
- Rollback path must be available (revert to `fast` mapper) for production incidents.

```

## Tips
- Keep ADRs concise and focused.
- ADRs are historical records — they document rationale even when a decision changes later.

If you want, I can add `docs/adr/TEMPLATE.md` with this template or add a CI check that validates ADR front matter.
