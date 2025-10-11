# Metrics catalog

This file documents the primary metrics collected by CostScope, their metric names, types, labels, and intended uses.

Example metric entry format:

- name: `costscope_system_cpu_seconds_total`
  - type: counter
  - description: Total CPU seconds used by the service
  - labels: `instanceregioncomponent`

```bash
,
```
  - collection_interval: 30s

TODO: split metrics into system, application, business, integration, and provider sections.
