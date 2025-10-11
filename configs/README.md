# CostScope Configuration Examples

This directory contains example configuration files for different deployment scenarios.

## Configuration Files

- `development.yaml` - Development environment configuration
- `staging.yaml` - Staging environment configuration  
- `production.yaml` - Production environment configuration
- `docker.yaml` - Docker deployment configuration

## Environment Variables

Configuration can be overridden using environment variables with the prefix `COSTSCOPE_`.

### Core Settings
- `COSTSCOPE_ENVIRONMENT` - deployment environment (development/staging/production/testing)
- `COSTSCOPE_DATA_DIR` - data directory path
- `COSTSCOPE_CORE_APP_NAME` - application name
- `COSTSCOPE_CORE_VERSION` - application version
- `COSTSCOPE_CORE_LOG_LEVEL` - logging level (debug/info/warn/error)

### Database Settings
- `COSTSCOPE_DATABASE_TYPE` - database type (sqlite/postgres/mysql)
- `COSTSCOPE_DATABASE_CONNECTION_STRING` - database connection string
- `COSTSCOPE_DATABASE_MAX_CONNECTIONS` - maximum database connections

### Provider Settings
- `COSTSCOPE_PROVIDERS_AWS_ENABLED` - enable AWS provider
- `COSTSCOPE_PROVIDERS_AWS_REGION` - AWS region
- `COSTSCOPE_PROVIDERS_AWS_PROFILE` - AWS profile

### Security Settings
- `COSTSCOPE_SECURITY_ENCRYPTION_ENABLED` - enable encryption
- `COSTSCOPE_SECURITY_TLS_ENABLED` - enable TLS

For complete list of environment variables, see the source code in `internal/core/config/loader.go`.
