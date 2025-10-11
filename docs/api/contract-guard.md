# API Contract Guard & Override Process

Automated contract guard using baseline specs in `api/openapi.v1.jsonapi/openapi.enterprise.v1.json`.

```bash
and
```

Example invocation:

```bash
make contract-check
# to allow a documented breaking change (include CHANGELOG note):
ALLOW_API_DIFF=1 make contract-check
```
