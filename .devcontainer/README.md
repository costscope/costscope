# DevContainer README

## CostScope Development Environment

This DevContainer provides a complete, isolated development environment for CostScope with all necessary tools pre-installed.

### What's Included

#### Development Tools
- **Go 1.24.6 (pinned)** - Exact patch pinned to avoid auto toolchain mismatch (GOTOOLCHAIN=local)
- **golangci-lint** - Comprehensive static analysis
- **gosec** - Security analysis
- **govulncheck** - Vulnerability scanning
- **Air** - Live reloading for development
- **Delve** - Go debugger

#### VS Code Extensions
- Go language support with IntelliSense
- GitHub Copilot integration
- GitLens for enhanced Git experience
- Test Explorer for running tests
- Docker support

#### Quality Assurance
- Pre-commit hooks
- Automated duplicate detection
- Comprehensive testing tools
- Code coverage reporting

### Quick Start

1. **Open in DevContainer**
   ```bash
   # In VS Code, open command palette (Cmd+Shift+P)
   # Select: "Dev Containers: Reopen in Container"
   ```

2. **Verify Setup**
   ```bash
   make env
   ```

3. **Run Quality Checks**
   ```bash
   make quality
   ```

4. **Start Development**
   ```bash
   make dev
   ```

### Available Commands

#### Quality Assurance
```bash
make quality      # Run all quality checks
make duplicates   # Check for duplicate code
make lint         # Static analysis
make security     # Security scan
make vuln         # Vulnerability check
make test         # Run tests
```

#### Development
```bash
make build        # Build application
make dev          # Start with live reload
make clean        # Clean build artifacts
make fmt          # Format code
```

#### Docker
```bash
make docker-build # Build Docker image
make docker-run   # Run in container
docker-compose up # Start full stack
```

### Pre-commit Hooks

Git hooks are automatically installed to run:
1. Duplicate detection
2. Static analysis
3. Tests

Bypass hooks (not recommended):
```bash
git commit --no-verify
```

### VS Code Configuration

The DevContainer includes optimized VS Code settings:
- Auto-format on save
- Go-specific linting
- Test integration
- Debug configuration

### Environment Variables

Create `.devcontainer/.env` for local customization:
```bash
# Example environment variables
COSTSCOPE_LOG_LEVEL=debug
COSTSCOPE_DB_URL=postgres://user:pass@localhost/costscope
```

### Troubleshooting

#### Container Issues
```bash
# Rebuild container
Cmd+Shift+P -> "Dev Containers: Rebuild Container"

# Reset completely
Cmd+Shift+P -> "Dev Containers: Rebuild Without Cache"
```

#### Go Version / Toolchain Issues
Pinned mode disables implicit downloads. To bump version:
1. Edit `COSTSCOPE_GO_VERSION` in `devcontainer.json` (containerEnv) or export before rebuild.
2. Update `go.mod` directive to the same patch.
3. Rebuild container (without cache) and verify: `go version`.

If you accidentally switch `go.mod` to a newer patch without updating the container:
```
Cmd+Shift+P -> Dev Containers: Rebuild Container
```
Or manually force reinstall inside container:
```
export GOTOOLCHAIN=local
sudo rm -rf /usr/local/go
curl -fsSL https://go.dev/dl/go1.24.6.linux-arm64.tar.gz -o /tmp/go.tgz
sudo tar -C /usr/local -xzf /tmp/go.tgz && rm /tmp/go.tgz
go clean -cache -testcache -modcache
go version
```

#### Go Module Issues
```bash
make deps
go clean -modcache
```

#### Tool Issues
```bash
# Reinstall tools
bash .devcontainer/setup.sh
```

### Performance Tips

1. **Use bind mounts** for better file system performance
2. **Exclude node_modules** from file watching
3. **Use Go build cache** (`GOCACHE`)
4. **Configure git properly** for container use

### Contributing

When adding new tools or configurations:
1. Update `.devcontainer/setup.sh`
2. Update `devcontainer.json` if needed
3. Update this README
4. Test in clean container

### Support

For DevContainer-specific issues:
- Check VS Code DevContainer documentation
- Verify Docker is running
- Check container logs
