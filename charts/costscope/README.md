# CostScope Helm Chart

This chart deploys the CostScope Enterprise API.

## Key Endpoints and Probes
- Readiness: GET `/health/ready` (200 only when subsystems are ready; 503 with diagnostics otherwise)
- Liveness: GET `/health/live` (200 while the process is running)

The Deployment template is wired to use these endpoints. Switch probe scheme to HTTPS when in-app TLS is enabled or if you terminate TLS at a sidecar/ingress using HTTPS-only probes.

## Required Secrets
A strong JWT secret is required (minimum 32 random bytes; recommend 48+). The chart refuses to install if it’s missing.

Values:
- `.Values.env.COSTSCOPE_JWT_SECRET` – required. Inject via `--set` or a values file. Never commit real secrets.

The rendered Secret has key `jwt-secret`. The Deployment consumes it as `COSTSCOPE_JWT_SECRET`.

## TLS and CORS
TLS (in-app) can be enabled via flags; CORS origins should be restricted in production.

Values (excerpt):
```yaml
env:
  # CORS
  CORS_ORIGINS: "https://app.example.com,https://admin.example.com"  # avoid "*" in prod

  # TLS (in-app)
  TLS_ENABLED: false
  TLS_CERT_FILE: "/tls/tls.crt"
  TLS_KEY_FILE: "/tls/tls.key"
  TLS_MIN_VERSION: "1.2"   # or "1.3"
  TLS_CIPHER_SUITES: ""    # optional CSV for TLS 1.2
  TLS_PREFER_SERVER_CIPHERS: true

probes:
  scheme: HTTP  # set to HTTPS when TLS is enabled
```

When `env.TLS_ENABLED` is true, the chart passes `--tls-enabled`, `--tls-cert`, `--tls-key`, and optional hardening flags to the container.

## Service and Ingress
- Service port defaults to 8080
- Optional ingress support via `.Values.ingress.*`

## Metrics
- Prometheus endpoint: GET `/metrics`
- Optional ServiceMonitor can be enabled under `.Values.serviceMonitor.enabled`

## Install Examples
Basic (dev):
```bash
helm upgrade --install costscope ./charts/costscope \
  --set env.COSTSCOPE_JWT_SECRET="$(openssl rand -base64 48)"
```

TLS + restricted CORS:
```bash
helm upgrade --install costscope ./charts/costscope \
  --set env.COSTSCOPE_JWT_SECRET="$(openssl rand -base64 48)" \
  --set env.CORS_ORIGINS="https://app.example.com,https://admin.example.com" \
  --set env.TLS_ENABLED=true \
  --set env.TLS_CERT_FILE=/tls/tls.crt \
  --set env.TLS_KEY_FILE=/tls/tls.key \
  --set probes.scheme=HTTPS
```

See `values.yaml` for all available options.
