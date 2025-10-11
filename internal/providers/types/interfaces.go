package types

import (
	"context"
	"time"
)

// CloudProvider represents a unified interface for cloud cost management providers
type CloudProvider interface {
	// Authentication and validation
	ValidateCredentials(ctx context.Context, config map[string]string) error
	GetProviderInfo(ctx context.Context) (ProviderInfo, error)

	// Cost data operations
	GetCostData(ctx context.Context, params CostDataParams) ([]CostRecord, error)
	GetResourceData(ctx context.Context, params ResourceDataParams) ([]ResourceRecord, error)

	// Configuration
	GetName() string
	GetType() ProviderType
	GetSupportedRegions() []string
}

// ProviderType represents the type of cloud provider
type ProviderType string

const (
	ProviderTypeAWS   ProviderType = "aws"
	ProviderTypeAzure ProviderType = "azure"
	ProviderTypeGCP   ProviderType = "gcp"
)

// ProviderInfo contains basic information about a provider
type ProviderInfo struct {
	Name             string            `json:"name"`
	Type             ProviderType      `json:"type"`
	Version          string            `json:"version"`
	SupportedRegions []string          `json:"supported_regions"`
	Capabilities     []string          `json:"capabilities"`
	Metadata         map[string]string `json:"metadata"`
}

// ProviderConfig holds configuration for a cloud provider
type ProviderConfig struct {
	Name        string                 `json:"name"`
	Type        ProviderType           `json:"type"`
	Credentials map[string]string      `json:"credentials"`
	Settings    map[string]interface{} `json:"settings"`
	Regions     []string               `json:"regions"`
	IsDefault   bool                   `json:"is_default"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	// TenantID optional multi-tenant identifier (feature flag gated)
	TenantID string `json:"tenant_id,omitempty"`
}

// CostDataParams defines parameters for cost data queries
type CostDataParams struct {
	StartDate   time.Time              `json:"start_date"`
	EndDate     time.Time              `json:"end_date"`
	Granularity string                 `json:"granularity"` // DAILY, WEEKLY, MONTHLY
	GroupBy     []string               `json:"group_by"`
	Filters     map[string]interface{} `json:"filters"`
	Metrics     []string               `json:"metrics"`
	MaxResults  int                    `json:"max_results"`
}

// ResourceDataParams defines parameters for resource data queries
type ResourceDataParams struct {
	ResourceTypes []string               `json:"resource_types"`
	Regions       []string               `json:"regions"`
	Tags          map[string]string      `json:"tags"`
	Filters       map[string]interface{} `json:"filters"`
	MaxResults    int                    `json:"max_results"`
}

// CostRecord represents a single cost data point
type CostRecord struct {
	Date       time.Time              `json:"date"`
	Amount     float64                `json:"amount"`
	Currency   string                 `json:"currency"`
	Service    string                 `json:"service"`
	Region     string                 `json:"region"`
	ResourceID string                 `json:"resource_id"`
	Tags       map[string]string      `json:"tags"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// ResourceRecord represents a single cloud resource
type ResourceRecord struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Type       string                 `json:"type"`
	Region     string                 `json:"region"`
	Status     string                 `json:"status"`
	Cost       float64                `json:"cost"`
	Currency   string                 `json:"currency"`
	Tags       map[string]string      `json:"tags"`
	Properties map[string]interface{} `json:"properties"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
}

// ProviderStatus represents the operational status of a provider
type ProviderStatus struct {
	Name           string       `json:"name"`
	Type           ProviderType `json:"type"`
	IsConnected    bool         `json:"is_connected"`
	LastSyncTime   time.Time    `json:"last_sync_time"`
	ErrorMessage   string       `json:"error_message,omitempty"`
	HealthStatus   string       `json:"health_status"`
	MetricsCount   int          `json:"metrics_count"`
	ResourcesCount int          `json:"resources_count"`
}
