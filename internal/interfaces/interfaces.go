// Package interfaces defines common interfaces for CostScope framework
package interfaces

import (
	"context"
	"time"
)

// Service represents a basic service interface
type Service interface {
	Name() string
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Health() string
}

// Configurable represents services that can be configured
type Configurable interface {
	Configure(config map[string]interface{}) error
	GetConfig() map[string]interface{}
}

// Observable represents services that can be observed/monitored
type Observable interface {
	GetMetrics() map[string]interface{}
	GetStats() map[string]interface{}
}

// Cacheable represents services that support caching
type Cacheable interface {
	SetCache(key string, value interface{}, ttl time.Duration) error
	GetCache(key string) (interface{}, bool)
	ClearCache() error
}

// Provider represents a data provider interface
type Provider interface {
	Service
	GetData(ctx context.Context, query interface{}) (interface{}, error)
	ValidateQuery(query interface{}) error
	GetCapabilities() []string
}

// Analyzer represents a data analyzer interface
type Analyzer interface {
	Service
	Analyze(ctx context.Context, data interface{}) (interface{}, error)
	GetAnalysisTypes() []string
}

// Reporter represents a report generator interface
type Reporter interface {
	Service
	GenerateReport(ctx context.Context, params ReportParams) (Report, error)
	GetSupportedFormats() []string
}

// ReportParams defines parameters for report generation
type ReportParams struct {
	Type      string                 `json:"type"`
	Format    string                 `json:"format"`
	TimeRange TimeRange              `json:"time_range"`
	Filters   map[string]interface{} `json:"filters"`
	Options   map[string]interface{} `json:"options"`
}

// Report represents a generated report
type Report struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Format      string                 `json:"format"`
	GeneratedAt time.Time              `json:"generated_at"`
	Data        interface{}            `json:"data"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// TimeRange represents a time range for queries
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// Plugin represents a plugin interface
type Plugin interface {
	Service
	GetInfo() PluginInfo
	Initialize(ctx context.Context, deps Dependencies) error
}

// PluginInfo contains plugin metadata
type PluginInfo struct {
	Name        string                 `json:"name"`
	Version     string                 `json:"version"`
	Description string                 `json:"description"`
	Author      string                 `json:"author"`
	License     string                 `json:"license"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// Dependencies represents plugin dependencies
type Dependencies struct {
	Container interface{} `json:"-"`
	EventBus  interface{} `json:"-"`
	Logger    interface{} `json:"-"`
	Config    interface{} `json:"-"`
}

// EventHandler represents an event handler interface
type EventHandler interface {
	Handle(event Event) error
	GetEventTypes() []string
}

// Event represents an event in the system
type Event interface {
	GetName() string
	GetData() map[string]interface{}
	GetTimestamp() time.Time
	GetSource() string
}

// Repository represents a data repository interface
type Repository interface {
	Create(ctx context.Context, entity interface{}) error
	Read(ctx context.Context, id interface{}) (interface{}, error)
	Update(ctx context.Context, entity interface{}) error
	Delete(ctx context.Context, id interface{}) error
	List(ctx context.Context, filter interface{}) ([]interface{}, error)
}

// QueryBuilder represents a query builder interface
type QueryBuilder interface {
	Select(fields ...string) QueryBuilder
	From(table string) QueryBuilder
	Where(condition string, args ...interface{}) QueryBuilder
	GroupBy(fields ...string) QueryBuilder
	OrderBy(field string, direction string) QueryBuilder
	Limit(limit int) QueryBuilder
	Offset(offset int) QueryBuilder
	Build() (string, []interface{}, error)
}

// Validator represents a validation interface
type Validator interface {
	Validate(data interface{}) error
	GetRules() []ValidationRule
}

// ValidationRule represents a validation rule
type ValidationRule struct {
	Field    string      `json:"field"`
	Type     string      `json:"type"`
	Required bool        `json:"required"`
	Options  interface{} `json:"options"`
}

// Transformer represents a data transformer interface
type Transformer interface {
	Transform(ctx context.Context, input interface{}) (interface{}, error)
	GetInputType() string
	GetOutputType() string
}

// Serializer represents a data serializer interface
type Serializer interface {
	Serialize(data interface{}) ([]byte, error)
	Deserialize(data []byte, target interface{}) error
	GetFormat() string
}

// Middleware represents middleware interface
type Middleware interface {
	Process(ctx context.Context, request interface{}, next func(context.Context, interface{}) (interface{}, error)) (interface{}, error)
	GetPriority() int
}

// Authenticator represents authentication interface
type Authenticator interface {
	Authenticate(ctx context.Context, credentials interface{}) (Principal, error)
	Validate(ctx context.Context, token interface{}) (Principal, error)
}

// Principal represents an authenticated user/service
type Principal interface {
	GetID() string
	GetName() string
	GetRoles() []string
	GetPermissions() []string
	IsAuthenticated() bool
}

// Authorizer represents authorization interface
type Authorizer interface {
	Authorize(ctx context.Context, principal Principal, resource string, action string) error
	GetPermissions(ctx context.Context, principal Principal) ([]Permission, error)
}

// Permission represents a permission
type Permission struct {
	Resource string `json:"resource"`
	Action   string `json:"action"`
	Effect   string `json:"effect"` // allow/deny
}

// RateLimiter represents rate limiting interface
type RateLimiter interface {
	Allow(ctx context.Context, key string) (bool, error)
	Reset(ctx context.Context, key string) error
	GetLimit(key string) (int, time.Duration, error)
}

// HealthChecker represents health check interface
type HealthChecker interface {
	CheckHealth(ctx context.Context) HealthStatus
	GetDependencies() []string
}

// HealthStatus represents health check status
type HealthStatus struct {
	Status    string            `json:"status"` // healthy, degraded, unhealthy
	Timestamp time.Time         `json:"timestamp"`
	Duration  time.Duration     `json:"duration"`
	Details   map[string]string `json:"details"`
	Errors    []string          `json:"errors,omitempty"`
}

// Notifier represents notification interface
type Notifier interface {
	Notify(ctx context.Context, notification Notification) error
	GetChannels() []string
}

// Notification represents a notification
type Notification struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Recipient string                 `json:"recipient"`
	Subject   string                 `json:"subject"`
	Body      string                 `json:"body"`
	Data      map[string]interface{} `json:"data"`
	Channel   string                 `json:"channel"`
	Priority  int                    `json:"priority"`
	CreatedAt time.Time              `json:"created_at"`
}
