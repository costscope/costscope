# FOCUS conversion packages (refactor notes)

This directory is being split into cohesive subpackages to improve maintainability.

Planned structure:

- universal/ — orchestration (universal converter + job manager)
- common/ — shared utilities (normalizers, classifiers, constants)
- aws/ — AWS-specific converter, mapper, pipeline, IO
- azure/ — Azure-specific converter, mapper, pipeline, IO
- gcp/ — GCP-specific converter, mapper, pipeline, IO

Readers will move to `internal/core/focus/reader/` under provider-specific subfolders
(e.g., `reader/aws`, `reader/azure`, `reader/gcp`). During the transition we will keep
small forwarders in original locations to avoid breaking imports.

Goals:
- No behavior changes; parity and performance guards must pass.
- Clear boundaries and smaller packages for faster builds/tests.

The conversion packages are split by provider for clarity and to avoid import cycles, with a thin universal orchestrator.
