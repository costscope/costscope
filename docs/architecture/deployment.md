---
title: Deployment Architecture
status: draft-split
description: Deployment topology, modes, and sizing guidance for CostScope.
---

# Deployment Architecture

This document was split from `overview.md` to provide focused guidance on deployment topologies, sizing, and infrastructure requirements.

## Production Deployment Options

### 1. Standalone Deployment
```yaml
costscope:
  binary: costscope
  config: /etc/costscope/config.yaml
  data: /var/lib/costscope/
  logs: /var/log/costscope/
```

### 2. Container Deployment
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: costscope
spec:
  replicas: 3
  selector:
    matchLabels:
      app: costscope
  template:
    spec:
      containers:
      - name: costscope
        image: costscope:latest
        ports:
        - containerPort: 8080
        env:
        - name: COSTSCOPE_CONFIG
          value: /config/production.yaml
```

### 3. Microservices Deployment
```yaml
services:
  - costscope-api        # REST API server
  - costscope-workers    # Background job processors
  - costscope-stream     # Streaming data processor
  - costscope-scheduler  # Cron job scheduler
```

## Infrastructure Requirements

| Profile            | CPU        | Memory       | Storage                | Network |
|--------------------|-----------:|-------------:|------------------------|---------|
| Minimum Dev/Test   | 2 cores    | 4 GB         | 50 GB (local)          | 100 Mbps|
| Recommended Prod   | 8+ cores   | 16+ GB       | 500+ GB SSD            | 1 Gbps  |
| Enterprise Scale   | 16+ cores  | 64+ GB       | Multi-TB distributed   | 10 Gbps |

### Scaling Guidance
- Prefer horizontal scaling (additional replicas) before vertical scaling.
- Separate worker and API process pools under sustained ingestion loads.
- Use dedicated storage volume for DuckDB/Parquet to avoid I/O contention.

### Configuration Isolation
| Concern        | Strategy                          |
|----------------|------------------------------------|
| Secrets        | Mounted via runtime secret store   |
| Config         | Immutable baseline + overrides     |
| Certificates   | Managed by ingress / gateway       |
| Temp Storage   | Fast ephemeral (container disk)    |

## Deployment Checklist (Abbreviated)
- Health probes wired (liveness & readiness)
- Metrics endpoint scraped (Prometheus)
- Log sink structured (JSON)
- Config immutability verified (checksum)
- RBAC policies loaded
- Storage path writable & capacity monitored
- Graceful shutdown signals handled

## See also
- `../ops/production-deployment.md`
- `../release/checklist.md`
- `../dev/performance-benchmarks.md`
- `monitoring.md`

---
_Status: initial extracted split. Improve with topology diagrams if needed._
