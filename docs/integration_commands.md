# Integration Commands

Auto-generated summary of declarative integration ActionSpecs. Do not edit manually.

Schema Version: 1.1  \n+Checksum: `6523760403835bd3ca0ede2414ca43bebdfb8475a4941692557baf3fae01f6da`  \n+Commands: 32

## Summary Table

| Path | Category | Description | Flags |
|------|----------|-------------|-------|
| `connections connect` | connections | Connect to third-party systems with advanced management | 7 |
| `dashboard config` | dashboard | Manage dashboard configuration | 0 |
| `dashboard config reset` | dashboard | Reset configuration to defaults | 0 |
| `dashboard config set` | dashboard | Set configuration value | 0 |
| `dashboard config show` | dashboard | Show current configuration | 0 |
| `dashboard export` | dashboard | Export dashboard data | 0 |
| `dashboard plugin` | dashboard | Manage dashboard plugins | 0 |
| `dashboard plugin disable` | dashboard | Disable plugin | 0 |
| `dashboard plugin enable` | dashboard | Enable plugin | 0 |
| `dashboard plugin install` | dashboard | Install plugin | 0 |
| `dashboard plugin list` | dashboard | List plugins | 0 |
| `dashboard start` | dashboard | Start enhanced dashboard server | 7 |
| `dashboard status` | dashboard | Show dashboard status and metrics | 1 |
| `dashboard stop` | dashboard | Stop dashboard server | 0 |
| `dashboard theme` | dashboard | Manage dashboard themes | 0 |
| `dashboard widget` | dashboard | Manage dashboard widgets | 0 |
| `dashboard widget add` | dashboard | Add new widget | 0 |
| `dashboard widget configure` | dashboard | Configure widget | 0 |
| `dashboard widget list` | dashboard | List dashboard widgets | 0 |
| `dashboard widget remove` | dashboard | Remove widget | 0 |
| `webhook create` | webhook | Create a new webhook with advanced features | 7 |
| `webhook delete` | webhook | Delete a webhook | 2 |
| `webhook delivery` | webhook | Manage webhook deliveries | 0 |
| `webhook delivery list` | webhook | List webhook deliveries | 0 |
| `webhook delivery retry` | webhook | Retry failed delivery | 0 |
| `webhook delivery stats` | webhook | Show delivery statistics | 0 |
| `webhook event` | webhook | Manage webhook events | 0 |
| `webhook event list` | webhook | List available webhook events | 0 |
| `webhook event trigger` | webhook | Trigger a test event | 0 |
| `webhook list` | webhook | List all webhooks | 2 |
| `webhook security` | webhook | Manage webhook security | 0 |
| `webhook test` | webhook | Test webhook delivery | 3 |

## Command Tree

````
  connections connect
  dashboard config
    dashboard config reset
    dashboard config set
    dashboard config show
  dashboard export
  dashboard plugin
    dashboard plugin disable
    dashboard plugin enable
    dashboard plugin install
    dashboard plugin list
  dashboard start
  dashboard status
  dashboard stop
  dashboard theme
  dashboard widget
    dashboard widget add
    dashboard widget configure
    dashboard widget list
    dashboard widget remove
  webhook create
  webhook delete
  webhook delivery
    webhook delivery list
    webhook delivery retry
    webhook delivery stats
  webhook event
    webhook event list
    webhook event trigger
  webhook list
  webhook security
  webhook test
````

## Command Details

### `connections connect`

Connect to third-party systems with advanced management

Integrate with billing systems, ITSM tools, and BI platforms with health monitoring and retry.

| Flag | Type | Required | Default | Usage |
|------|------|----------|---------|-------|
| auto-retry | bool | no | true | enable automatic retry on connection failures |
| category | string | no | <nil> | filter by integration category |
| config | stringToString | no | [] | connection configuration (key=value pairs) |
| health-check | bool | no | true | enable health monitoring for the connection |
| list | bool | no | false | list available integrations |
| system | string | no | <nil> | system to connect to |
| timeout | duration | no | 30s | connection timeout |

### `dashboard config`

Manage dashboard configuration

View and modify dashboard configuration settings.

_Group command (no direct execution)._

No flags.

### `dashboard config reset`

Reset configuration to defaults

No flags.

### `dashboard config set`

Set configuration value

No flags.

### `dashboard config show`

Show current configuration

No flags.

### `dashboard export`

Export dashboard data

No flags.

### `dashboard plugin`

Manage dashboard plugins

Install, enable, and configure dashboard plugins for extended functionality.

_Group command (no direct execution)._

No flags.

### `dashboard plugin disable`

Disable plugin

No flags.

### `dashboard plugin enable`

Enable plugin

No flags.

### `dashboard plugin install`

Install plugin

No flags.

### `dashboard plugin list`

List plugins

No flags.

### `dashboard start`

Start enhanced dashboard server

Start the enhanced dashboard server with customizable features and security options.

| Flag | Type | Required | Default | Usage |
|------|------|----------|---------|-------|
| allowed-ips | stringSlice | no | [] | Allowed IP addresses/ranges (CIDR notation) |
| auth | bool | no | false | Enable authentication |
| auto-open | bool | no | true | Automatically open dashboard in browser |
| features | stringSlice | no | [real-time,interactive,mobile] | Enable specific features |
| layout | string | no | default | Dashboard layout (default, grid, compact, executive) |
| port | int | no | 8080 | Dashboard port |
| theme | string | no | modern | Dashboard theme (modern, dark, light, corporate) |

### `dashboard status`

Show dashboard status and metrics

Display dashboard status, performance metrics, and configuration details.

| Flag | Type | Required | Default | Usage |
|------|------|----------|---------|-------|
| verbose | bool | no | false | Show detailed status information |

### `dashboard stop`

Stop dashboard server

Gracefully stop the dashboard server and save current state.

No flags.

### `dashboard theme`

Manage dashboard themes

No flags.

### `dashboard widget`

Manage dashboard widgets

Add, remove, and configure dashboard widgets for custom layouts.

_Group command (no direct execution)._

No flags.

### `dashboard widget add`

Add new widget

No flags.

### `dashboard widget configure`

Configure widget

No flags.

### `dashboard widget list`

List dashboard widgets

No flags.

### `dashboard widget remove`

Remove widget

No flags.

### `webhook create`

Create a new webhook with advanced features

Create a new webhook with support for custom headers, security, retry policies, and event filtering.

**Example:**\n
````bash
costscope integration webhook create --name alerts --url https://example.com --events cost.threshold.exceeded
````

| Flag | Type | Required | Default | Usage |
|------|------|----------|---------|-------|
| events | stringSlice | no | [*] | Events to subscribe to |
| headers | stringToString | no | [] | Custom HTTP headers |
| max-retries | int | no | 3 | Maximum retry attempts |
| name | string | yes | <nil> | Webhook name |
| secret | string | no | <nil> | Secret for signature verification |
| timeout | duration | no | 30s | Request timeout |
| url | string | yes | <nil> | Webhook URL |

### `webhook delete`

Delete a webhook

Delete a webhook and its delivery history.

| Flag | Type | Required | Default | Usage |
|------|------|----------|---------|-------|
| confirm | bool | no | false | Skip confirmation prompt |
| webhook-id | string | yes | <nil> | Webhook ID to delete |

### `webhook delivery`

Manage webhook deliveries

View and manage webhook delivery history and statistics.

_Group command (no direct execution)._

No flags.

### `webhook delivery list`

List webhook deliveries

No flags.

### `webhook delivery retry`

Retry failed delivery

No flags.

### `webhook delivery stats`

Show delivery statistics

No flags.

### `webhook event`

Manage webhook events

View available events and trigger test events.

_Group command (no direct execution)._

No flags.

### `webhook event list`

List available webhook events

No flags.

### `webhook event trigger`

Trigger a test event

No flags.

### `webhook list`

List all webhooks

List all configured webhooks with their status and recent delivery statistics.

| Flag | Type | Required | Default | Usage |
|------|------|----------|---------|-------|
| status | string | no | <nil> | Filter by status (active, error, disabled) |
| verbose | bool | no | false | Show detailed information |

### `webhook security`

Manage webhook security

View or rotate webhook security credentials.

No flags.

### `webhook test`

Test webhook delivery

Send a test event to a webhook to verify configuration and connectivity.

| Flag | Type | Required | Default | Usage |
|------|------|----------|---------|-------|
| event | string | no | test.webhook | Event type to send |
| payload | string | no | <nil> | Custom JSON payload |
| webhook-id | string | yes | <nil> | Webhook ID to test |

