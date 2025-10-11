# Production Deployment Guide

---
title: Production Deployment
description: Practical deployment guidance, environment variables, readiness probes, observability, and CI/CD outline.
---

# Production Deployment Guide

## Minimum requirements
- CPU: 2 vCPU (recommend 4+)
- RAM: 4 GB (recommend 8+ GB)
- Storage: 50 GB SSD (logs + Prometheus TSDB may need more)

## Environment variables
- ENV: production
- LOG_LEVEL: info
- COSTSCOPE_JWT_SECRET: secret for JWT signing (required)
- COSTSCOPE_CORS_ORIGINS: allowed CORS origins, e.g. "https://dashboard.company.com"

## Health/Readiness
- Liveness: GET /health/live
- Readiness: GET /health/ready
- Health: GET /health

## Observability
- Metrics endpoint: GET /metrics (Prometheus format)
- Prometheus scrape example: monitoring/prometheus.yml
- Grafana dashboard: monitoring/grafana/provisioning/dashboards/costscope-overview.json

## CI/CD outline
1. Build and test (go test ./...)
2. Scan (gosec) and SBOM
3. Buildx push to registry
4. Sign with cosign
5. helm upgrade --install charts/costscope

## Try it locally
- docker-compose up -d
- Open Grafana http://localhost:3000 (admin/admin)
- Prometheus at http://localhost:9090

## See also
- `production-readiness-cli.md`
- `logging.md`
- `monitoring/monitoring-overview.md`
- `../security/` (policies & supply chain)
- `../architecture/overview.md`
