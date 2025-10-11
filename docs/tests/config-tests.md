# Config module test matrix

| ID | Area | Case | Source(s) | Expected |
|----|------|------|-----------|----------|
| M1 | Merge | File-only | file | Loaded values used |
| M2 | Merge | Env-only overrides | file+env | Env overrides file |
| M3 | Merge | Flags override highest | file+env+flags | Flags override env and file |
| R1 | Required | Missing core.app_name | any | Validation error with key=app_name |
| F1 | Fallback | GetEnvironment before load | none | Development returned |
| F2 | Fallback | GetDataDir before load | none | ./data returned |
| D1 | Defaults | Zero env int not override | file+env(0) | File positive preserved |
| N1 | Negative | Bad YAML | file | Parse error |
| N2 | Negative | Unknown key | file | Parse error (strict) |
| N3 | Negative | Empty lines | file | No error |
| REG | Regression | Placeholder for future bugs | - | Tracked as added |

Coverage target: +15% for package internal/core/config; report exported as coverage.html.
