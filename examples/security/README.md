# Casbin PoC Example (build tag required)

This example shows how to run the Enterprise API with the optional Casbin RBAC PoC enabled via a build tag.

Prerequisites:
- Build the binary with the `casbinpoc` tag
- Provide a model and policy (examples are in `configs/`)

Build (macOS/Linux):

```bash
# Slim
go build -tags casbinpoc -o costscope ./

# DuckDB + Enterprise + PoC
go build -tags "duckdb enterprise casbinpoc" -o costscope ./
```

Run API with example policy:

```bash
export CASBIN_MODEL_PATH=configs/rbac_model.conf.example
export CASBIN_POLICY_PATH=configs/rbac_policy.csv.example
# optional: export CASBIN_DOMAIN=default

# JWT-based subject extraction uses roles claim; ensure COSTSCOPE_JWT_SECRET is set
export COSTSCOPE_JWT_SECRET="$(openssl rand -base64 48)"
./costscope api enterprise --port 8080
```

Quick smoke with JWT (preferred):

```bash
# 1) Build with casbin PoC
go build -tags casbinpoc -o costscope ./

# 2) Start API with model/policy and a strong JWT secret
export CASBIN_MODEL_PATH=$(pwd)/examples/security/model.conf
export CASBIN_POLICY_PATH=$(pwd)/examples/security/policy.csv
export COSTSCOPE_JWT_SECRET="$(openssl rand -base64 48)"
./costscope api enterprise --port 8080 &

# 3) Generate a JWT that grants role:admin and call a protected endpoint
TOKEN=$(COSTSCOPE_JWT_SECRET="$COSTSCOPE_JWT_SECRET" go run ./examples/security/jwt_quickstart.go -roles admin)
curl -i -H "Authorization: Bearer $TOKEN" -X POST http://localhost:8080/api/v1/focus/convert
```

Quick smoke (header subject extractor via X-Subject header, for PoC only):

```bash
# Path under /api/v1 should be allowed for role:admin by the example policy
curl -i -H 'X-Subject: role:admin' -X POST http://localhost:8080/api/v1/focus/convert
```

Notes:
- Without the `casbinpoc` build tag, the server runs without Casbin even if CASBIN_* variables are set.
- For JWT extraction, tokens must include a `roles` array claim; the middleware expects entries like `admin` or `role:admin`.
 - The helper `examples/security/jwt_quickstart.go` reads the HMAC secret from `COSTSCOPE_JWT_SECRET`. Ensure it matches the API's secret.
