//go:build sqlite

package reports

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"local/costscope/internal/core/logging"

	_ "github.com/mattn/go-sqlite3" // sqlite driver
)

// SQLiteMetadataStore provides durable metadata persistence using SQLite (enabled with 'sqlite' build tag).
// Table schema (automatically created):
// CREATE TABLE IF NOT EXISTS report_metadata (
//   id TEXT PRIMARY KEY,
//   path TEXT NOT NULL,
//   format TEXT NOT NULL,
//   size_bytes INTEGER NOT NULL,
//   checksum_sha256 TEXT,
//   created_at TIMESTAMP NOT NULL,
//   output_dir_source TEXT
// );
// CREATE INDEX IF NOT EXISTS idx_report_metadata_created_at ON report_metadata(created_at);
// CREATE INDEX IF NOT EXISTS idx_report_metadata_format ON report_metadata(format);
// Retention: max records & max age enforced opportunistically on Save.

type SQLiteMetadataStore struct {
	db      *sql.DB
	log     *logging.Logger
	maxRecs int
	maxAge  time.Duration
}

func NewSQLiteMetadataStore(path string, log *logging.Logger, maxRecords int, maxAge time.Duration) (*SQLiteMetadataStore, error) {
	if path == "" {
		return nil, fmt.Errorf("sqlite metadata path empty")
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	dsn := fmt.Sprintf("file:%s?_busy_timeout=5000&_journal=WAL&_fk=1", abs)
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}
	stmts := []string{
		"CREATE TABLE IF NOT EXISTS report_metadata (id TEXT PRIMARY KEY, path TEXT NOT NULL, format TEXT NOT NULL, size_bytes INTEGER NOT NULL, checksum_sha256 TEXT, created_at TIMESTAMP NOT NULL, output_dir_source TEXT)",
		"CREATE INDEX IF NOT EXISTS idx_report_metadata_created_at ON report_metadata(created_at)",
		"CREATE INDEX IF NOT EXISTS idx_report_metadata_format ON report_metadata(format)",
	}
	for _, s := range stmts {
		if _, err := db.Exec(s); err != nil {
			_ = db.Close()
			return nil, err
		}
	}
	return &SQLiteMetadataStore{db: db, log: log, maxRecs: maxRecords, maxAge: maxAge}, nil
}

func (s *SQLiteMetadataStore) Save(ctx context.Context, md *ReportMetadata) error {
	if md == nil {
		return errors.New("nil metadata")
	}
	_, err := s.db.ExecContext(ctx, `INSERT OR REPLACE INTO report_metadata (id,path,format,size_bytes,checksum_sha256,created_at,output_dir_source) VALUES (?,?,?,?,?,?,?)`, md.ID, md.Path, md.Format, md.SizeBytes, md.ChecksumSHA256, md.CreatedAt.UTC(), md.OutputDirSource)
	if err != nil {
		return err
	}
	// Opportunistic retention pruning (best-effort, ignore errors)
	s.prune(ctx)
	return nil
}

func (s *SQLiteMetadataStore) Get(ctx context.Context, id string) (*ReportMetadata, error) {
	row := s.db.QueryRowContext(ctx, `SELECT id,path,format,size_bytes,checksum_sha256,created_at,output_dir_source FROM report_metadata WHERE id=?`, id)
	var md ReportMetadata
	var created time.Time
	if err := row.Scan(&md.ID, &md.Path, &md.Format, &md.SizeBytes, &md.ChecksumSHA256, &created, &md.OutputDirSource); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("metadata not found: %s", id)
		}
		return nil, err
	}
	md.CreatedAt = created.UTC()
	return &md, nil
}

func (s *SQLiteMetadataStore) List(ctx context.Context) ([]*ReportMetadata, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,path,format,size_bytes,checksum_sha256,created_at,output_dir_source FROM report_metadata ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*ReportMetadata
	for rows.Next() {
		var md ReportMetadata
		var created time.Time
		if err := rows.Scan(&md.ID, &md.Path, &md.Format, &md.SizeBytes, &md.ChecksumSHA256, &created, &md.OutputDirSource); err == nil {
			md.CreatedAt = created.UTC()
			out = append(out, &md)
		}
	}
	return out, nil
}

func (s *SQLiteMetadataStore) ListOptions(ctx context.Context, opts *MetadataListOptions) ([]*ReportMetadata, error) {
	if opts == nil { // delegate to List for ordering
		return s.List(ctx)
	}
	query := `SELECT id,path,format,size_bytes,checksum_sha256,created_at,output_dir_source FROM report_metadata WHERE 1=1`
	args := []interface{}{}
	if opts.Format != "" {
		query += " AND format=?"
		args = append(args, opts.Format)
	}
	if opts.CreatedAfter != nil {
		query += " AND created_at >= ?"
		args = append(args, opts.CreatedAfter.UTC())
	}
	if opts.CreatedBefore != nil {
		query += " AND created_at <= ?"
		args = append(args, opts.CreatedBefore.UTC())
	}
	query += " ORDER BY created_at DESC"
	if opts.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", opts.Limit)
	}
	if opts.Offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", opts.Offset)
	}
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*ReportMetadata
	for rows.Next() {
		var md ReportMetadata
		var created time.Time
		if err := rows.Scan(&md.ID, &md.Path, &md.Format, &md.SizeBytes, &md.ChecksumSHA256, &created, &md.OutputDirSource); err == nil {
			md.CreatedAt = created.UTC()
			out = append(out, &md)
		}
	}
	return out, nil
}

func (s *SQLiteMetadataStore) prune(ctx context.Context) {
	if s.maxRecs <= 0 && s.maxAge <= 0 {
		return
	}
	// Age-based pruning
	if s.maxAge > 0 {
		cutoff := time.Now().Add(-s.maxAge)
		if _, err := s.db.ExecContext(ctx, `DELETE FROM report_metadata WHERE created_at < ?`, cutoff.UTC()); err != nil {
			s.log.DebugWithFields("metadata retention prune age failed", map[string]interface{}{"error": err.Error()})
		}
	}
	// Count-based pruning: delete oldest beyond limit
	if s.maxRecs > 0 {
		// Use a subquery to delete ids beyond limit
		if _, err := s.db.ExecContext(ctx, `DELETE FROM report_metadata WHERE id IN (SELECT id FROM report_metadata ORDER BY created_at DESC LIMIT -1 OFFSET ?)`, s.maxRecs); err != nil {
			s.log.DebugWithFields("metadata retention prune count failed", map[string]interface{}{"error": err.Error()})
		}
	}
}

// Migration helper: load file-based JSONL store into sqlite (best-effort, idempotent).
func MigrateFileMetadataToSQLite(ctx context.Context, filePath string, sqliteStore *SQLiteMetadataStore) error {
	if sqliteStore == nil {
		return fmt.Errorf("nil sqlite store")
	}
	if filePath == "" {
		return fmt.Errorf("file path empty")
	}
	// Reuse existing FileMetadataStore reader.
	fm := NewFileMetadataStore(filePath, sqliteStore.log)
	records, err := fm.List(ctx)
	if err != nil {
		return err
	}
	for _, r := range records {
		_ = sqliteStore.Save(ctx, r)
	}
	return nil
}
