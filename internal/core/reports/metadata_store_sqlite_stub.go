//go:build !sqlite

package reports

import (
	"context"
	"fmt"
	"time"

	"local/costscope/internal/core/logging"
)

// SQLiteMetadataStore stub when built without 'sqlite' tag.
type SQLiteMetadataStore struct{}

func NewSQLiteMetadataStore(path string, log *logging.Logger, maxRecords int, maxAge time.Duration) (*SQLiteMetadataStore, error) {
	return nil, fmt.Errorf("sqlite metadata store not enabled (missing 'sqlite' build tag)")
}

func (s *SQLiteMetadataStore) Save(ctx context.Context, md *ReportMetadata) error {
	return fmt.Errorf("sqlite metadata store disabled")
}
func (s *SQLiteMetadataStore) Get(ctx context.Context, id string) (*ReportMetadata, error) {
	return nil, fmt.Errorf("sqlite metadata store disabled")
}
func (s *SQLiteMetadataStore) List(ctx context.Context) ([]*ReportMetadata, error) {
	return []*ReportMetadata{}, nil
}
func (s *SQLiteMetadataStore) ListOptions(ctx context.Context, opts *MetadataListOptions) ([]*ReportMetadata, error) {
	return []*ReportMetadata{}, nil
}
