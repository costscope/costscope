//go:build sqlite
// +build sqlite

package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"

	streamingTypes "local/costscope/cmd/modules/streaming/types"
	"local/costscope/internal/core/logging"
	providerTypes "local/costscope/internal/providers/types"
)

// SQLiteRepository implements Repository interface using SQLite
type SQLiteRepository struct {
	db     *sql.DB
	logger *logging.Logger
}

// NewSQLiteRepository creates a new SQLite repository
func NewSQLiteRepository(config *DatabaseConfig) (*SQLiteRepository, error) {
	db, err := sql.Open("sqlite3", config.GetConnectionString())
	if err != nil {
		return nil, fmt.Errorf("failed to open SQLite database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetConnMaxLifetime(config.ConnMaxLifetime)
	db.SetConnMaxIdleTime(config.ConnMaxIdleTime)

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping SQLite database: %w", err)
	}

	repo := &SQLiteRepository{
		db:     db,
		logger: logging.NewLogger(logging.LevelInfo),
	}

	// Run migrations
	if err := repo.migrate(); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return repo, nil
}

// migrate runs database migrations
func (r *SQLiteRepository) migrate() error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS streaming_jobs (
			job_id TEXT PRIMARY KEY,
			provider TEXT NOT NULL,
			input_path TEXT NOT NULL,
			output_path TEXT NOT NULL,
			workers INTEGER NOT NULL DEFAULT 4,
			memory INTEGER NOT NULL DEFAULT 512,
			parameters TEXT NOT NULL DEFAULT '{}',
			status TEXT NOT NULL DEFAULT 'created',
			progress REAL NOT NULL DEFAULT 0.0,
			processed_rows INTEGER NOT NULL DEFAULT 0,
			total_rows INTEGER NOT NULL DEFAULT 0,
			processed_bytes INTEGER NOT NULL DEFAULT 0,
			total_bytes INTEGER NOT NULL DEFAULT 0,
			error_message TEXT,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			start_time DATETIME,
			last_update DATETIME,
			estimated_end DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS provider_configs (
			name TEXT PRIMARY KEY,
			type TEXT NOT NULL,
			credentials TEXT NOT NULL DEFAULT '{}',
			settings TEXT NOT NULL DEFAULT '{}',
			regions TEXT NOT NULL DEFAULT '[]',
			is_default BOOLEAN NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_jobs_status ON streaming_jobs(status)`,
		`CREATE INDEX IF NOT EXISTS idx_jobs_provider ON streaming_jobs(provider)`,
		`CREATE INDEX IF NOT EXISTS idx_jobs_created_at ON streaming_jobs(created_at)`,
		`CREATE INDEX IF NOT EXISTS idx_providers_type ON provider_configs(type)`,
	}

	for _, migration := range migrations {
		if _, err := r.db.Exec(migration); err != nil {
			return fmt.Errorf("failed to execute migration: %w", err)
		}
	}

	return nil
}

// SaveJob saves a streaming job to the database
func (r *SQLiteRepository) SaveJob(ctx context.Context, job *streamingTypes.StreamingJobInfo) error {
	parametersJSON, err := json.Marshal(job.Config.Parameters)
	if err != nil {
		return fmt.Errorf("failed to marshal job parameters: %w", err)
	}

	query := `INSERT OR REPLACE INTO streaming_jobs (
		job_id, provider, input_path, output_path, workers, memory, parameters,
		status, progress, processed_rows, total_rows, processed_bytes, total_bytes,
		error_message, created_at, updated_at, start_time, last_update, estimated_end
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err = r.db.ExecContext(ctx, query,
		job.Config.JobID,
		job.Config.Provider,
		job.Config.InputPath,
		job.Config.OutputPath,
		job.Config.Workers,
		job.Config.Memory,
		string(parametersJSON),
		job.Status.Status,
		job.Status.Progress,
		job.Status.ProcessedRows,
		job.Status.TotalRows,
		job.Status.ProcessedBytes,
		job.Status.TotalBytes,
		job.Status.ErrorMessage,
		job.Config.CreatedAt,
		job.Config.UpdatedAt,
		timeToNullTime(job.Status.StartTime),
		timeToNullTime(job.Status.LastUpdate),
		timeToNullTime(job.Status.EstimatedEnd),
	)

	if err != nil {
		return fmt.Errorf("failed to save job: %w", err)
	}

	r.logger.Info(fmt.Sprintf("Saved job: %s", job.Config.JobID))
	return nil
}

// GetJob retrieves a streaming job from the database
func (r *SQLiteRepository) GetJob(ctx context.Context, jobID string) (*streamingTypes.StreamingJobInfo, error) {
	query := `SELECT 
		job_id, provider, input_path, output_path, workers, memory, parameters,
		status, progress, processed_rows, total_rows, processed_bytes, total_bytes,
		error_message, created_at, updated_at, start_time, last_update, estimated_end
	FROM streaming_jobs WHERE job_id = ?`

	row := r.db.QueryRowContext(ctx, query, jobID)

	var job streamingTypes.StreamingJobInfo
	var parametersJSON string
	var startTime, lastUpdate, estimatedEnd sql.NullTime

	err := row.Scan(
		&job.Config.JobID,
		&job.Config.Provider,
		&job.Config.InputPath,
		&job.Config.OutputPath,
		&job.Config.Workers,
		&job.Config.Memory,
		&parametersJSON,
		&job.Status.Status,
		&job.Status.Progress,
		&job.Status.ProcessedRows,
		&job.Status.TotalRows,
		&job.Status.ProcessedBytes,
		&job.Status.TotalBytes,
		&job.Status.ErrorMessage,
		&job.Config.CreatedAt,
		&job.Config.UpdatedAt,
		&startTime,
		&lastUpdate,
		&estimatedEnd,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("job not found: %s", jobID)
		}
		return nil, fmt.Errorf("failed to get job: %w", err)
	}

	// Parse parameters JSON
	if err := json.Unmarshal([]byte(parametersJSON), &job.Config.Parameters); err != nil {
		return nil, fmt.Errorf("failed to unmarshal job parameters: %w", err)
	}

	// Handle nullable times
	job.Status.JobID = job.Config.JobID
	job.Status.StartTime = nullTimeToTime(startTime)
	job.Status.LastUpdate = nullTimeToTime(lastUpdate)
	job.Status.EstimatedEnd = nullTimeToTime(estimatedEnd)

	return &job, nil
}

// ListJobs retrieves streaming jobs with filters
func (r *SQLiteRepository) ListJobs(ctx context.Context, filters JobFilters) ([]*streamingTypes.StreamingJobInfo, error) {
	query := `SELECT 
		job_id, provider, input_path, output_path, workers, memory, parameters,
		status, progress, processed_rows, total_rows, processed_bytes, total_bytes,
		error_message, created_at, updated_at, start_time, last_update, estimated_end
	FROM streaming_jobs WHERE 1=1`

	args := []interface{}{}
	argCount := 0

	// Apply filters
	if len(filters.Status) > 0 {
		placeholders := ""
		for i, status := range filters.Status {
			if i > 0 {
				placeholders += ","
			}
			placeholders += "?"
			args = append(args, status)
			argCount++
		}
		query += fmt.Sprintf(" AND status IN (%s)", placeholders)
	}

	if len(filters.Provider) > 0 {
		placeholders := ""
		for i, provider := range filters.Provider {
			if i > 0 {
				placeholders += ","
			}
			placeholders += "?"
			args = append(args, provider)
			argCount++
		}
		query += fmt.Sprintf(" AND provider IN (%s)", placeholders)
	}

	if filters.CreatedAt.From != nil {
		query += " AND created_at >= ?"
		args = append(args, *filters.CreatedAt.From)
	}

	if filters.CreatedAt.To != nil {
		query += " AND created_at <= ?"
		args = append(args, *filters.CreatedAt.To)
	}

	// Add sorting
	if filters.SortBy != "" {
		query += fmt.Sprintf(" ORDER BY %s", filters.SortBy)
		if filters.SortOrder == "DESC" {
			query += " DESC"
		} else {
			query += " ASC"
		}
	} else {
		query += " ORDER BY created_at DESC"
	}

	// Add pagination
	if filters.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filters.Limit)
	}

	if filters.Offset > 0 {
		query += " OFFSET ?"
		args = append(args, filters.Offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list jobs: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			r.logger.Error(fmt.Sprintf("Failed to close rows: %v", closeErr))
		}
	}()

	var jobs []*streamingTypes.StreamingJobInfo

	for rows.Next() {
		var job streamingTypes.StreamingJobInfo
		var parametersJSON string
		var startTime, lastUpdate, estimatedEnd sql.NullTime

		err := rows.Scan(
			&job.Config.JobID,
			&job.Config.Provider,
			&job.Config.InputPath,
			&job.Config.OutputPath,
			&job.Config.Workers,
			&job.Config.Memory,
			&parametersJSON,
			&job.Status.Status,
			&job.Status.Progress,
			&job.Status.ProcessedRows,
			&job.Status.TotalRows,
			&job.Status.ProcessedBytes,
			&job.Status.TotalBytes,
			&job.Status.ErrorMessage,
			&job.Config.CreatedAt,
			&job.Config.UpdatedAt,
			&startTime,
			&lastUpdate,
			&estimatedEnd,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan job row: %w", err)
		}

		// Parse parameters JSON
		if err := json.Unmarshal([]byte(parametersJSON), &job.Config.Parameters); err != nil {
			return nil, fmt.Errorf("failed to unmarshal job parameters: %w", err)
		}

		// Handle nullable times
		job.Status.JobID = job.Config.JobID
		job.Status.StartTime = nullTimeToTime(startTime)
		job.Status.LastUpdate = nullTimeToTime(lastUpdate)
		job.Status.EstimatedEnd = nullTimeToTime(estimatedEnd)

		jobs = append(jobs, &job)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate job rows: %w", err)
	}

	return jobs, nil
}

// UpdateJobStatus updates only the status of a job
func (r *SQLiteRepository) UpdateJobStatus(ctx context.Context, jobID string, status *streamingTypes.StreamingJobStatus) error {
	query := `UPDATE streaming_jobs SET 
		status = ?, progress = ?, processed_rows = ?, total_rows = ?, 
		processed_bytes = ?, total_bytes = ?, error_message = ?,
		last_update = ?, estimated_end = ?, updated_at = ?
	WHERE job_id = ?`

	_, err := r.db.ExecContext(ctx, query,
		status.Status,
		status.Progress,
		status.ProcessedRows,
		status.TotalRows,
		status.ProcessedBytes,
		status.TotalBytes,
		status.ErrorMessage,
		timeToNullTime(status.LastUpdate),
		timeToNullTime(status.EstimatedEnd),
		time.Now(),
		jobID,
	)

	if err != nil {
		return fmt.Errorf("failed to update job status: %w", err)
	}

	return nil
}

// DeleteJob deletes a streaming job from the database
func (r *SQLiteRepository) DeleteJob(ctx context.Context, jobID string) error {
	query := `DELETE FROM streaming_jobs WHERE job_id = ?`

	result, err := r.db.ExecContext(ctx, query, jobID)
	if err != nil {
		return fmt.Errorf("failed to delete job: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("job not found: %s", jobID)
	}

	r.logger.Info(fmt.Sprintf("Deleted job: %s", jobID))
	return nil
}

// SaveProvider saves a provider configuration to the database
func (r *SQLiteRepository) SaveProvider(ctx context.Context, config *providerTypes.ProviderConfig) error {
	credentialsJSON, err := json.Marshal(config.Credentials)
	if err != nil {
		return fmt.Errorf("failed to marshal provider credentials: %w", err)
	}

	settingsJSON, err := json.Marshal(config.Settings)
	if err != nil {
		return fmt.Errorf("failed to marshal provider settings: %w", err)
	}

	regionsJSON, err := json.Marshal(config.Regions)
	if err != nil {
		return fmt.Errorf("failed to marshal provider regions: %w", err)
	}

	query := `INSERT OR REPLACE INTO provider_configs (
		name, type, credentials, settings, regions, is_default, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	_, err = r.db.ExecContext(ctx, query,
		config.Name,
		string(config.Type),
		string(credentialsJSON),
		string(settingsJSON),
		string(regionsJSON),
		config.IsDefault,
		config.CreatedAt,
		config.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to save provider: %w", err)
	}

	r.logger.Info(fmt.Sprintf("Saved provider: %s", config.Name))
	return nil
}

// GetProvider retrieves a provider configuration from the database
func (r *SQLiteRepository) GetProvider(ctx context.Context, name string) (*providerTypes.ProviderConfig, error) {
	query := `SELECT 
		name, type, credentials, settings, regions, is_default, created_at, updated_at
	FROM provider_configs WHERE name = ?`

	row := r.db.QueryRowContext(ctx, query, name)

	var config providerTypes.ProviderConfig
	var typeStr, credentialsJSON, settingsJSON, regionsJSON string

	err := row.Scan(
		&config.Name,
		&typeStr,
		&credentialsJSON,
		&settingsJSON,
		&regionsJSON,
		&config.IsDefault,
		&config.CreatedAt,
		&config.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("provider not found: %s", name)
		}
		return nil, fmt.Errorf("failed to get provider: %w", err)
	}

	// Parse JSON fields
	config.Type = providerTypes.ProviderType(typeStr)

	if err := json.Unmarshal([]byte(credentialsJSON), &config.Credentials); err != nil {
		return nil, fmt.Errorf("failed to unmarshal provider credentials: %w", err)
	}

	if err := json.Unmarshal([]byte(settingsJSON), &config.Settings); err != nil {
		return nil, fmt.Errorf("failed to unmarshal provider settings: %w", err)
	}

	if err := json.Unmarshal([]byte(regionsJSON), &config.Regions); err != nil {
		return nil, fmt.Errorf("failed to unmarshal provider regions: %w", err)
	}

	return &config, nil
}

// ListProviders retrieves all provider configurations
func (r *SQLiteRepository) ListProviders(ctx context.Context) ([]*providerTypes.ProviderConfig, error) {
	query := `SELECT 
		name, type, credentials, settings, regions, is_default, created_at, updated_at
	FROM provider_configs ORDER BY name`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list providers: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			r.logger.Error(fmt.Sprintf("Failed to close rows: %v", closeErr))
		}
	}()

	var configs []*providerTypes.ProviderConfig

	for rows.Next() {
		var config providerTypes.ProviderConfig
		var typeStr, credentialsJSON, settingsJSON, regionsJSON string

		err := rows.Scan(
			&config.Name,
			&typeStr,
			&credentialsJSON,
			&settingsJSON,
			&regionsJSON,
			&config.IsDefault,
			&config.CreatedAt,
			&config.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan provider row: %w", err)
		}

		// Parse JSON fields
		config.Type = providerTypes.ProviderType(typeStr)

		if err := json.Unmarshal([]byte(credentialsJSON), &config.Credentials); err != nil {
			return nil, fmt.Errorf("failed to unmarshal provider credentials: %w", err)
		}

		if err := json.Unmarshal([]byte(settingsJSON), &config.Settings); err != nil {
			return nil, fmt.Errorf("failed to unmarshal provider settings: %w", err)
		}

		if err := json.Unmarshal([]byte(regionsJSON), &config.Regions); err != nil {
			return nil, fmt.Errorf("failed to unmarshal provider regions: %w", err)
		}

		configs = append(configs, &config)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate provider rows: %w", err)
	}

	return configs, nil
}

// DeleteProvider deletes a provider configuration from the database
func (r *SQLiteRepository) DeleteProvider(ctx context.Context, name string) error {
	query := `DELETE FROM provider_configs WHERE name = ?`

	result, err := r.db.ExecContext(ctx, query, name)
	if err != nil {
		return fmt.Errorf("failed to delete provider: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("provider not found: %s", name)
	}

	r.logger.Info(fmt.Sprintf("Deleted provider: %s", name))
	return nil
}

// Health checks the database connection
func (r *SQLiteRepository) Health(ctx context.Context) error {
	return r.db.PingContext(ctx)
}

// Close closes the database connection
func (r *SQLiteRepository) Close() error {
	return r.db.Close()
}

// Helper functions for handling nullable times
func timeToNullTime(t time.Time) sql.NullTime {
	if t.IsZero() {
		return sql.NullTime{Valid: false}
	}
	return sql.NullTime{Time: t, Valid: true}
}

func nullTimeToTime(nt sql.NullTime) time.Time {
	if nt.Valid {
		return nt.Time
	}
	return time.Time{}
}
