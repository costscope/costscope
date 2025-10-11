# VS Code Troubleshooting

## Issue: Stale Compilation Errors

If VS Code shows errors like "duplicate declarations" or references moved files, follow these steps:

### 1. Clean Go Cache
```bash
go clean -cache -testcache -modcache
```

### 2. Restart Go Language Server
- Open Command Palette (Cmd+Shift+P)
- Run:

```bash
Go: Restart Language Server
```

### 3. Reload VS Code Window
- Command Palette (Cmd+Shift+P)
- Run:

```bash
Developer: Reload Window
```

### 4. Full VS Code Restart
- Close VS Code completely
- Reopen the project

### 5. Project Status Checks
```bash
# Build check
go build ./

# Linter check
go vet ./...

# Static analysis check
staticcheck ./...
```

## Project Structure

After reorganization, development utilities live in:
- `scripts/tools/code-optimizer/`
- `scripts/tools/gc-test/`
- `scripts/tools/performance-profiler/`
- `scripts/tools/gc-benchmark/`

Each utility is a separate executable with its own .

```bash
package main
```

## Building Utilities

```bash
# Build all utilities
make tools-build

# Clean utility binaries
make tools-clean
```

## See also
- `support.md`
- `faq.md`
- `troubleshooting.md`

