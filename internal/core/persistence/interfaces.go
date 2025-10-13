package persistence

import (
	"context"
	"time"

	streamingTypes "github.com/costscope/costscope/cmd/modules/streaming/types"
	providerTypes "github.com/costscope/costscope/internal/providers/types"
)

// Repository defines the interface for data persistence
type Repository interface {
	// Job operations
	SaveJob(ctx context.Context, job *streamingTypes.StreamingJobInfo) error
	GetJob(ctx context.Context, jobID string) (*streamingTypes.StreamingJobInfo, error)
	ListJobs(ctx context.Context, filters JobFilters) ([]*streamingTypes.StreamingJobInfo, error)
	UpdateJobStatus(ctx context.Context, jobID string, status *streamingTypes.StreamingJobStatus) error
	DeleteJob(ctx context.Context, jobID string) error

	// Provider operations
	SaveProvider(ctx context.Context, config *providerTypes.ProviderConfig) error
	GetProvider(ctx context.Context, name string) (*providerTypes.ProviderConfig, error)
	ListProviders(ctx context.Context) ([]*providerTypes.ProviderConfig, error)
	DeleteProvider(ctx context.Context, name string) error

	// Health and maintenance
	Health(ctx context.Context) error
	Close() error
}

// JobFilters defines filters for job queries
type JobFilters struct {
	Status    []string  `json:"status"`
	Provider  []string  `json:"provider"`
	CreatedAt TimeRange `json:"created_at"`
	UpdatedAt TimeRange `json:"updated_at"`
	Limit     int       `json:"limit"`
	Offset    int       `json:"offset"`
	SortBy    string    `json:"sort_by"`
	SortOrder string    `json:"sort_order"`
}

// TimeRange represents a time range filter
type TimeRange struct {
	From *time.Time `json:"from"`
	To   *time.Time `json:"to"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Type     DatabaseType `json:"type"`
	Host     string       `json:"host"`
	Port     int          `json:"port"`
	Database string       `json:"database"`
	Username string       `json:"username"`
	Password string       `json:"password"`
	SSLMode  string       `json:"ssl_mode"`
	FilePath string       `json:"file_path"` // For SQLite

	// Connection pool settings
	MaxOpenConns    int           `json:"max_open_conns"`
	MaxIdleConns    int           `json:"max_idle_conns"`
	ConnMaxLifetime time.Duration `json:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `json:"conn_max_idle_time"`
}

// DatabaseType represents supported database types
type DatabaseType string

const (
	// Only SQLite is supported today. PostgreSQL stub removed to reduce dead surface; reintroduce when adapter implemented.
	DatabaseTypeSQLite DatabaseType = "sqlite"
)

// IsValid retained for forward compatibility but now trivial; returns true only for known types.
func (dt DatabaseType) IsValid() bool { // nolint:revive
	return dt == DatabaseTypeSQLite
}

// DefaultDatabaseConfig returns default database configuration
func DefaultDatabaseConfig() *DatabaseConfig {
	return &DatabaseConfig{
		Type:            DatabaseTypeSQLite,
		FilePath:        "./costscope.db",
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 5 * time.Minute,
	}
}

// PostgreSQLConfig returns default PostgreSQL configuration.
// Intentional stub (future adapter): retained to avoid churn in docs referencing Postgres while
// the production implementation is deferred. Not used at runtime.
// PostgreSQLConfig removed (was a stub). Reintroduce with real connection settings when PostgreSQL repository is added.

// GetConnectionString returns database connection string
func (c *DatabaseConfig) GetConnectionString() string {
	if c.Type == DatabaseTypeSQLite {
		return c.FilePath
	}
	return "" // only sqlite currently supported
}

// Migrator defines the interface for database migrations
type Migrator interface {
	Up(ctx context.Context) error
	Down(ctx context.Context) error
	Version(ctx context.Context) (int, error)
	SetVersion(ctx context.Context, version int) error
}

// Transaction defines the interface for database transactions
type Transaction interface {
	Repository
	Commit() error
	Rollback() error
}
