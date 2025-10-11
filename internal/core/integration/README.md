# Integration Module Documentation

## Overview

The Integration module provides comprehensive capabilities for connecting CostScope with external systems, managing cost alerts, automating workflows, and operating interactive dashboards. This module serves as the central hub for external integrations and automation within the CostScope ecosystem.

## Features

###  System Integrations
- **Billing Systems**: AWS, Azure, GCP, CloudHealth, CloudCheckr
- **ITSM Tools**: ServiceNow, JIRA, FreshService, Zendesk
- **BI Platforms**: Tableau, PowerBI, Looker, QlikSense, Grafana
- **Monitoring**: Datadog, New Relic, Splunk, Elasticsearch, Prometheus
- **Notifications**: Slack, Teams, Discord, PagerDuty, Twilio
- **Automation**: Jenkins, GitHub Actions, GitLab, Ansible, Terraform

###  Alert Management
- **Alert Types**: Budget, Threshold, Anomaly, Forecast, Usage, Compliance, Optimization
- **Severity Levels**: Low, Medium, High, Critical
- **Notification Channels**: Email, Slack, SMS, Webhook, Teams, Discord, Dashboard, PagerDuty
- **Flexible Conditions**: Metric-based triggers with duration and aggregation
- **Scheduling**: Configurable alert checking schedules

###  Workflow Automation
- **Automated Workflows**: Cost analysis, optimization, reporting, and alerting
- **Scheduling**: Cron-based, interval-based, or manual execution
- **Step Types**: Analysis, optimization, alert, report, action, condition, integration, notification
- **Dependencies**: Multi-step workflows with conditional execution
- **Cost Savings Tracking**: Monitor savings from automated optimizations

###  Interactive Dashboard
- **Web Interface**: Real-time cost visualization and monitoring
- **Themes**: Light, dark, and auto themes
- **Real-time Updates**: Configurable refresh rates
- **User Management**: Authentication and access control
- **Metrics Display**: Cost trends, top services, alerts, and system status

###  Webhook Support
- **Event-driven Integrations**: Cost changes, budget alerts, anomaly detection
- **Security**: Secret-based verification and custom headers
- **Retry Logic**: Configurable retry policies with exponential backoff
- **Event Types**: Cost change, budget exceeded, anomaly detected, report generated

## Architecture

### Service Layer
- **IntegrationService**: Main service interface with comprehensive operations
- **SystemConnection**: Manages individual system connections with health monitoring
- **AuditLogger**: Tracks all integration activities for compliance
- **MetricsCollector**: Collects performance metrics for monitoring

### Type System
- **Comprehensive Types**: 200+ lines of type definitions covering all integration scenarios
- **Integration Categories**: 8 categories with specific system types
- **Status Management**: Connection, alert, workflow, and webhook status tracking
- **Configuration**: Flexible configuration system for different integration types

### CLI Interface
- **Command Structure**: Hierarchical commands with comprehensive help
- **Parameter Validation**: Input validation with clear error messages
- **Output Formatting**: Tabular output with proper alignment and colors
- **Examples**: Comprehensive examples for all commands

## Installation & Setup

### Prerequisites
- Go 1.24.5 or later
- Access to external systems (optional for testing)
- Network connectivity for webhooks and API calls

### Build & Install
```bash
# Build the project
go build -o bin/costscope

# Run integration commands
./bin/costscope integration --help
```

## Usage Examples

### List Available Integrations
```bash
# List all available integrations
costscope integration list

# Filter by category
costscope integration list --category billing

# Filter by status
costscope integration list --status connected
```

### Connect to External Systems
```bash
# Connect to AWS
costscope integration connect aws \
  --config access_key=AKIA... \
  --config secret_key=wJa... \
  --config region=us-east-1

# Connect to Slack
costscope integration connect slack \
  --credential webhook_url=https://hooks.slack.com/...

# Test connection without persisting
costscope integration connect aws --test-mode
```

### Manage Cost Alerts
```bash
# Create budget alert
costscope integration alert create \
  --name "Monthly Budget Alert" \
  --type budget \
  --threshold 1000 \
  --severity high \
  --channels email,slack

# Create anomaly detection alert
costscope integration alert create \
  --name "Cost Anomaly Detection" \
  --type anomaly \
  --severity critical \
  --channels pagerduty,teams

# List all alerts
costscope integration alert list

# Update an alert
costscope integration alert update alert_123 \
  --threshold 1500 \
  --enabled true

# Test notification channels
costscope integration alert test
```

### Automate Workflows
```bash
# Create daily analysis workflow
costscope integration workflow create \
  --name "Daily Cost Analysis" \
  --description "Automated daily cost analysis and optimization" \
  --schedule daily \
  --steps "cost_analysis,optimization,reporting"

# Execute workflow immediately
costscope integration workflow execute workflow_123

# List all workflows
costscope integration workflow list

# Update workflow schedule
costscope integration workflow update workflow_123 \
  --schedule weekly \
  --enabled true
```

### Interactive Dashboard
```bash
# Start dashboard on default port (8080)
costscope integration dashboard start

# Start with custom configuration
costscope integration dashboard start \
  --port 9090 \
  --theme dark \
  --auto-open \
  --refresh-rate 15

# Get dashboard status
costscope integration dashboard status

# Get current metrics
costscope integration dashboard metrics

# Stop dashboard
costscope integration dashboard stop
```

### Webhook Management
```bash
# Create webhook for Slack notifications
costscope integration webhook create \
  --name "Slack Cost Alerts" \
  --url "https://hooks.slack.com/services/..." \
  --events cost_change,budget_exceeded \
  --secret "webhook_secret_key" \
  --enabled

# Create webhook with custom headers
costscope integration webhook create \
  --name "Custom API Integration" \
  --url "https://api.example.com/webhook" \
  --events anomaly_detected,report_generated \
  --headers "Authorization=Bearer token123" \
  --headers "Content-Type=application/json"

# Test webhook
costscope integration webhook test webhook_123

# List all webhooks
costscope integration webhook list
```

### Connection Management
```bash
# Get connection status
costscope integration status aws

# Test connection
costscope integration test slack

# Disconnect from system
costscope integration disconnect aws
```

## Configuration

### System-Specific Configuration

#### AWS Configuration
```bash
costscope integration connect aws \
  --config access_key=AKIA... \
  --config secret_key=wJa... \
  --config region=us-east-1 \
  --config session_token=... # Optional for temporary credentials
```

#### Slack Configuration
```bash
costscope integration connect slack \
  --credential webhook_url=https://hooks.slack.com/services/... \
  --config channel="#cost-alerts" \
  --config username="CostScope Bot"
```

#### Tableau Configuration
```bash
costscope integration connect tableau \
  --config server_url=https://tableau.company.com \
  --credential username=user@company.com \
  --credential password=secure_password \
  --config site_id=default
```

### Alert Configuration
- **Budget Alerts**: Monitor spending against predefined budgets
- **Threshold Alerts**: Trigger when costs exceed specific values
- **Anomaly Alerts**: Detect unusual spending patterns using ML
- **Forecast Alerts**: Warn about projected future costs
- **Usage Alerts**: Monitor resource utilization patterns
- **Compliance Alerts**: Ensure adherence to cost policies

### Workflow Steps
- **Analysis**: Cost analysis, trend analysis, variance analysis
- **Optimization**: Resource right-sizing, waste elimination, recommendations
- **Alert**: Trigger alerts based on conditions
- **Report**: Generate and distribute cost reports
- **Action**: Execute cost optimization actions
- **Condition**: Conditional logic for workflow branching
- **Integration**: Sync data with external systems
- **Notification**: Send notifications to stakeholders

## API Reference

### IntegrationService Interface
```go
type IntegrationService interface {
    // Connection Management
    ListIntegrations(filter *IntegrationFilter) (*IntegrationListResult, error)
    ConnectToSystem(request *ConnectionRequest) (*ConnectionResult, error)
    DisconnectFromSystem(systemName string) (*DisconnectionResult, error)
    GetConnectionStatus(systemName string) (*ConnectionStatus, error)
    TestConnection(systemName string) (*ConnectionTestResult, error)

    // Alert Management
    CreateAlert(request *AlertCreateRequest) (*AlertCreateResult, error)
    ListAlerts(filter *AlertFilter) (*AlertListResult, error)
    UpdateAlert(alertID string, request *AlertUpdateRequest) (*AlertUpdateResult, error)
    DeleteAlert(alertID string) (*AlertDeleteResult, error)
    TestAlertChannels() (*AlertTestResult, error)

    // Workflow Management
    CreateWorkflow(request *WorkflowCreateRequest) (*WorkflowCreateResult, error)
    ListWorkflows(filter *WorkflowFilter) (*WorkflowListResult, error)
    ExecuteWorkflow(workflowID string) (*WorkflowExecutionResult, error)
    UpdateWorkflow(workflowID string, request *WorkflowUpdateRequest) (*WorkflowUpdateResult, error)
    DeleteWorkflow(workflowID string) (*WorkflowDeleteResult, error)

    // Dashboard Management
    StartDashboard(config *DashboardConfig) (*DashboardStartResult, error)
    StopDashboard() (*DashboardStopResult, error)
    GetDashboardStatus() (*DashboardStatusResult, error)
    GetDashboardMetrics() (*DashboardMetricsResult, error)

    // Webhook Management
    CreateWebhook(request *WebhookCreateRequest) (*WebhookCreateResult, error)
    ListWebhooks() (*WebhookListResult, error)
    TestWebhook(webhookID string) (*WebhookTestResult, error)
    DeleteWebhook(webhookID string) (*WebhookDeleteResult, error)
}
```

## Security Features

### Authentication & Authorization
- **API Key Management**: Secure storage and rotation of API keys
- **Token-based Authentication**: JWT tokens with configurable expiration
- **Role-based Access**: Different permission levels for different users
- **Audit Logging**: Complete audit trail of all integration activities

### Data Protection
- **Encryption in Transit**: TLS 1.3 for all external communications
- **Encryption at Rest**: Sensitive data encrypted using AES-256
- **Secret Management**: Secure storage of credentials and API keys
- **Data Masking**: Sensitive information masked in logs and outputs

### Network Security
- **Webhook Verification**: Secret-based webhook verification
- **IP Whitelisting**: Restrict access to specific IP ranges
- **Rate Limiting**: Prevent abuse with configurable rate limits
- **Certificate Validation**: Strict certificate validation for HTTPS

## Monitoring & Observability

### Metrics Collection
- **Performance Metrics**: Response times, throughput, error rates
- **Business Metrics**: Cost savings, alert accuracy, workflow success rates
- **System Metrics**: Connection health, resource utilization, uptime
- **Custom Metrics**: User-defined metrics for specific use cases

### Audit Logging
- **Activity Tracking**: All integration activities logged with timestamps
- **User Attribution**: Track which user performed which actions
- **Data Changes**: Log all configuration and data changes
- **Compliance Reports**: Generate compliance reports from audit logs

### Health Monitoring
- **Connection Health**: Monitor health of all external connections
- **Service Health**: Internal service health checks and monitoring
- **Dependency Monitoring**: Track health of external dependencies
- **Alerting**: Proactive alerts for system issues

## Troubleshooting

### Common Issues

#### Connection Failures
```bash
# Test connection to diagnose issues
costscope integration test system_name

# Check connection status
costscope integration status system_name

# View recent audit logs (implementation dependent)
costscope integration logs --system system_name --recent
```

#### Alert Not Triggering
1. Verify alert configuration: `costscope integration alert list`
2. Test notification channels: `costscope integration alert test`
3. Check alert conditions and thresholds
4. Verify data is being collected from connected systems

#### Workflow Execution Failures
1. Check workflow configuration: `costscope integration workflow list`
2. Execute workflow manually: `costscope integration workflow execute workflow_id`
3. Review workflow logs and error messages
4. Verify all required systems are connected

#### Dashboard Access Issues
1. Check dashboard status: `costscope integration dashboard status`
2. Verify port accessibility and firewall rules
3. Check authentication configuration if enabled
4. Review dashboard logs for errors

### Debug Mode
Enable debug logging by setting environment variables:
```bash
export COSTSCOPE_LOG_LEVEL=debug
export COSTSCOPE_INTEGRATION_DEBUG=true
```

## Best Practices

### Security
1. **Rotate Credentials Regularly**: Change API keys and passwords regularly
2. **Use Least Privilege**: Grant minimum required permissions
3. **Monitor Access**: Regularly review audit logs for suspicious activity
4. **Secure Webhooks**: Always use secrets for webhook verification

### Performance
1. **Connection Pooling**: Reuse connections where possible
2. **Caching**: Cache frequently accessed data
3. **Batch Operations**: Batch API calls to reduce overhead
4. **Rate Limiting**: Respect external system rate limits

### Reliability
1. **Error Handling**: Implement comprehensive error handling
2. **Retry Logic**: Use exponential backoff for failed operations
3. **Health Checks**: Regularly test connection health
4. **Monitoring**: Set up proactive monitoring and alerting

### Cost Optimization
1. **Alert Tuning**: Fine-tune alert thresholds to reduce noise
2. **Workflow Automation**: Automate routine cost optimization tasks
3. **Regular Reviews**: Regularly review and optimize integration costs
4. **Data Retention**: Implement appropriate data retention policies

## Contributing

### Development Setup
1. Clone the repository
2. Install Go 1.24.5 or later
3. Run tests: `go test ./internal/core/integration/... -v`
4. Build: `go build -o bin/costscope`

### Adding New Integrations
1. Add system type to `types.go`
2. Implement system-specific logic in `service.go`
3. Add CLI commands for the new integration
4. Write comprehensive tests
5. Update documentation

### Testing
- **Unit Tests**: Test individual components in isolation
- **Integration Tests**: Test end-to-end integration flows
- **Performance Tests**: Test under load conditions
- **Security Tests**: Test security features and vulnerabilities

## Changelog

### Version 1.0.0 (Phase 6.4)
-  Initial implementation of Integration module
-  Connection management for external systems
-  Alert management with multi-channel notifications
-  Workflow automation with scheduling
-  Interactive dashboard with real-time updates
-  Webhook support for event-driven integrations
-  Comprehensive CLI interface with 40+ commands
-  Complete test suite with 100% coverage
-  Security features including audit logging
-  Performance monitoring and metrics collection

## Support

For issues, feature requests, or questions:
1. Check the troubleshooting section above
2. Review the command help: `costscope integration --help`
3. Consult the API documentation
4. File an issue in the project repository

## License

This module is part of the CostScope project and is subject to the project's license terms.
