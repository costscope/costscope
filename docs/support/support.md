# Support

Thank you for using CostScope! This document explains the available support channels and how to get help effectively.

## Self-Service Resources
- README: installation & quick start
- `docs/` directory: architecture, deployment, logging & metrics, performance
- `CONTRIBUTING.md`: development workflow & quality gates
- `SECURITY.md`: vulnerability reporting

Search open and closed issues before filing a new one.

## Asking Questions
Open a GitHub Issue using the "Question" or "Support" label (or choose a template if provided) with:
- Environment details (OS, Go version, CostScope version )

```bash
costscope version
```
- Steps taken & expected vs actual behavior
- Relevant logs (redact secrets) – use fenced code blocks

## Bug Reports
Use the bug report template. Provide:
- Version / commit hash
- Reproduction steps (minimal script / commands)
- Input sample (sanitized)
- Observed output / stack traces
- Expected outcome

## Feature Requests
Use the feature request template. Include motivation, use case, and any proposed interface ideas.

## SLA / Response Times
Community support is offered on a best‑effort basis. Maintainer triage target: initial response within 5 business days (security reports: see `SECURITY.md`).

## Commercial / Extended Support
If you require guaranteed SLAs, roadmap alignment, or deployment assistance, contact: support@costscope.io.

## Version Support Policy
We generally support the latest minor release and the previous minor release for critical bug & security fixes. Patch releases focus on regressions and security.

## Troubleshooting Tips
Run the local quality checks to surface issues.

```bash
make quality
```
- Use `--verbose` flags for more logging where available.
- Confirm environment variables are set as expected ((see example below)).

```bash
env | grep COSTSCOPE_
```
- Validate config files with YAML linting before running.

## Contributing to Support Docs
Improvements welcome—open a PR updating docs or templates when you discover gaps.

Thank you for helping improve the project!
