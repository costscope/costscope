---
title: Production Readiness CLI
description: Command suite for assessing, validating, optimizing and reporting on production deployments.
---

# Production Readiness CLI Commands

Comprehensive production readiness assessment suite for CostScope.

## Overview
The `prod-readiness` subcommands (see `internal/core/production/`) provide real-time insights and actionable recommendations.

These commands provide comprehensive tools for assessing, monitoring, and optimizing production deployments. They integrate with probes, metrics exporters, and reporting outputs.

## Commands

### Assess

```bash
costscope prod-readiness assess [environment]
```

Comprehensive production readiness assessment including system health analysis, performance benchmarking, security compliance verification, and deployment prerequisites checking.

**Usage:**
```bash
# Basic assessment for production environment
costscope prod-readiness assess
```

... (trimmed quick section in index view; full deep-dive below) ...

---
## Deep-dive (Full Command Reference)

### Assess Command Features
- Readiness scoring (0-100) with weighting (performance, security, configuration, capacity)
- Failure classification (critical / warning / info)
- JSON structure stable for automation (`assessment_version` key)

### Metrics Command Output Types
- Health metrics: uptime, worker utilization, queue depth
- Performance metrics: CPU%, memory RSS, GC pause p95, conversion throughput
- Security metrics: pending vulnerabilities (count), audit mode status
- Integration metrics: provider connectivity, failing integrations count

### Optimization Command Categories
- Performance (CPU, memory, I/O recommendations)
- Cost (resource right-sizing hints)
- Security (hardening & policy reminders)
- Configuration (flag / ENV / file improvements)

### Deployment Planner
Produces structured plan sections: prerequisites, phased steps, verification gates, rollback steps, risk matrix (impact vs likelihood).

### Report Command Types
| Type | Audience | Emphasis |
|------|----------|----------|
| executive | Leadership | KPIs, summarized risk, trend deltas |
| technical | Engineering | Detailed metrics & recommendations |
| operational | SRE/Ops | Run readiness, probes, scaling & alerting |
| security | Security | Vulnerabilities, hardening, compliance |
| cost | FinOps | Resource utilization & optimization ROI |

### Output Format Notes
- HTML/PDF include embedded charts (latency, utilization, risk distribution)
- CSV includes flattened key metrics (timestamped) for spreadsheet import
- Prometheus format exposes ephemeral gauges (on demand)

### Performance Benchmarks
| Operation | Target | Notes |
|-----------|--------|-------|
| assess | <1s | Baseline on 8-core dev host |
| metrics (all) | <200ms | No remote calls blocking |
| optimize | <500ms | In-memory heuristic only |

### Security Guarantees
All commands avoid printing secrets, redact values with patterns `(secret|token|key)`.

### Examples (Extended)
```bash
# Production full pipeline
costscope prod-readiness assess production --detailed --output json > assessment.json
costscope prod-readiness metrics all --output json > metrics.json
costscope prod-readiness optimize --categories=performance,cost,security --output yaml > optimization.yaml
costscope prod-readiness deploy canary --dry-run --output json > canary-plan.json
costscope prod-readiness report executive --format pdf --output exec.pdf
```

### Roadmap
- Integrate probe latency trend analysis
- Add anomaly detection on readiness score slope
- Auto-open GitHub issues for critical recurring findings (toggle)

---
_Merged deep-dive content from an earlier root file; original removed._
---
title: Production Readiness CLI
description: Command suite for assessing, validating, optimizing and reporting on production deployments.
---

# Production Readiness CLI Commands

Comprehensive production readiness assessment suite for CostScope.

## Overview
The `prod-readinessinternal/core/production/` module to provide real-time insights and actionable recommendations.

```bash
commands provide comprehensive tools for assessing, monitoring, and optimizing production deployments. These commands integrate with the powerful
```

## Commands

###

```bash
costscope prod-readiness assess [environment]
```

Comprehensive production readiness assessment including system health analysis, performance benchmarking, security compliance verification, and deployment prerequisites checking.

**Usage:**
```bash
# Basic assessment for production environment
costscope prod-readiness assess
```

... (trimmed quick section in index view; full deep-dive below) ...

---
## Deep-dive (Full Command Reference)

### Assess Command Features
- Readiness scoring (0-100) with weighting (performance, security, configuration, capacity)
- Failure classification (critical / warning / info)
- JSON structure stable for automation (`assessment_version` key)

### Metrics Command Output Types
- Health metrics: uptime, worker utilization, queue depth
- Performance metrics: CPU%, memory RSS, GC pause p95, conversion throughput
- Security metrics: pending vulnerabilities (count), audit mode status
- Integration metrics: provider connectivity, failing integrations count

### Optimization Command Categories
- Performance (CPU, memory, I/O recommendations)
- Cost (resource right-sizing hints)
- Security (hardening & policy reminders)
- Configuration (flag / ENV / file improvements)

### Deployment Planner
Produces structured plan sections: prerequisites, phased steps, verification gates, rollback steps, risk matrix (impact vs likelihood).

### Report Command Types
| Type | Audience | Emphasis |
|------|----------|----------|
| executive | Leadership | KPIs, summarized risk, trend deltas |
| technical | Engineering | Detailed metrics & recommendations |
| operational | SRE/Ops | Run readiness, probes, scaling & alerting |
| security | Security | Vulnerabilities, hardening, compliance |
| cost | FinOps | Resource utilization & optimization ROI |

### Output Format Notes
- HTML/PDF include embedded charts (latency, utilization, risk distribution)
- CSV includes flattened key metrics (timestamped) for spreadsheet import
- Prometheus format exposes ephemeral gauges (on demand)

### Performance Benchmarks
| Operation | Target | Notes |
|-----------|--------|-------|
| assess | <1s | Baseline on 8-core dev host |
| metrics (all) | <200ms | No remote calls blocking |
| optimize | <500ms | In-memory heuristic only |

### Security Guarantees
All commands avoid printing secrets, redact values with patterns `(secret|token|key)`.

### Examples (Extended)
```bash
# Production full pipeline
costscope prod-readiness assess production --detailed --output json > assessment.json
costscope prod-readiness metrics all --output json > metrics.json
costscope prod-readiness optimize --categories=performance,cost,security --output yaml > optimization.yaml
costscope prod-readiness deploy canary --dry-run --output json > canary-plan.json
costscope prod-readiness report executive --format pdf --output exec.pdf
```

### Roadmap
- Integrate probe latency trend analysis
- Add anomaly detection on readiness score slope
- Auto-open GitHub issues for critical recurring findings (toggle)

---
_Merged deep-dive content from an earlier root file; original removed._
