package exporters

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	gcs "cloud.google.com/go/storage"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"google.golang.org/api/option"
)

// ObjectStore abstracts destination storage for report artifacts.
type ObjectStore interface {
	Put(ctx context.Context, dest string, r io.Reader) error
}

// NewObjectStore constructs an ObjectStore based on URL scheme: s3://, gs://, or file path.
func NewObjectStore(ctx context.Context, dest string) (ObjectStore, string, error) {
	if strings.HasPrefix(dest, "s3://") {
		s, kind, err := newS3Store(ctx)
		return s, kind, err
	}
	if strings.HasPrefix(dest, "gs://") {
		s, kind, err := newGCSStore(ctx)
		return s, kind, err
	}
	if strings.HasPrefix(dest, "file://") {
		// Treat as filesystem path after stripping scheme
		return &fsStore{}, "fs", nil
	}
	// default to filesystem
	return &fsStore{}, "fs", nil
}

// ---------------- Filesystem ----------------
type fsStore struct{}

func (f *fsStore) Put(ctx context.Context, dest string, r io.Reader) error {
	// Normalize file:// URLs to local path
	if strings.HasPrefix(dest, "file://") {
		// url.Parse to properly handle file URLs
		if u, err := url.Parse(dest); err == nil {
			if u.Host == "" {
				dest = u.Path
			} else {
				dest = u.Host + u.Path
			}
		} else {
			// Fallback: strip prefix
			dest = strings.TrimPrefix(dest, "file://")
		}
	}
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(dest), 0o750); err != nil {
		return err
	}
	// #nosec G304 - path validated by caller
	fOut, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer func() { _ = fOut.Close() }()
	if _, err := io.Copy(fOut, r); err != nil {
		return err
	}
	return nil
}

// ---------------- S3 ----------------
type s3Store struct {
	client *s3.Client
}

func newS3Store(ctx context.Context) (ObjectStore, string, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("aws config: %w", err)
	}
	return &s3Store{client: s3.NewFromConfig(cfg)}, "s3", nil
}

func (s *s3Store) Put(ctx context.Context, dest string, r io.Reader) error {
	// dest format: s3://bucket/key
	u, err := url.Parse(dest)
	if err != nil {
		return err
	}
	if u.Scheme != "s3" || u.Host == "" {
		return errors.New("invalid s3 url")
	}
	bucket := u.Host
	key := strings.TrimPrefix(u.Path, "/")
	up := manager.NewUploader(s.client)
	_, err = up.Upload(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   r,
	})
	return err
}

// ---------------- GCS ----------------
type gcsStore struct {
	client *gcs.Client
}

func newGCSStore(ctx context.Context) (ObjectStore, string, error) {
	// Credentials picked from env (GOOGLE_APPLICATION_CREDENTIALS) or adc
	c, err := gcs.NewClient(ctx, option.WithScopes(gcs.ScopeReadWrite))
	if err != nil {
		return nil, "", fmt.Errorf("gcs client: %w", err)
	}
	return &gcsStore{client: c}, "gcs", nil
}

func (g *gcsStore) Put(ctx context.Context, dest string, r io.Reader) error {
	// dest format: gs://bucket/path
	u, err := url.Parse(dest)
	if err != nil {
		return err
	}
	if u.Scheme != "gs" || u.Host == "" {
		return errors.New("invalid gs url")
	}
	b := g.client.Bucket(u.Host)
	obj := strings.TrimPrefix(u.Path, "/")
	w := b.Object(obj).NewWriter(ctx)
	defer func() { _ = w.Close() }()
	if _, err := io.Copy(w, r); err != nil {
		return err
	}
	return nil
}

// ---------------- HTTP helper ----------------
// httpDo performs an HTTP request with context, timeout and returns status error if non-2xx
func httpDo(ctx context.Context, req *http.Request, client *http.Client) (*http.Response, error) {
	req = req.WithContext(ctx)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		_ = resp.Body.Close()
		return nil, fmt.Errorf("http status %d", resp.StatusCode)
	}
	return resp, nil
}
