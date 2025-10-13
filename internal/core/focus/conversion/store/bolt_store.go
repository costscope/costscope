package store

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"

	focustypes "github.com/costscope/costscope/internal/core/focus/types"
	"github.com/costscope/costscope/internal/core/logging"
	"github.com/costscope/costscope/internal/core/monitoring/telemetry"

	bbolt "go.etcd.io/bbolt"
)

const boltJobsBucket = "jobs"

// BoltJobStore is a durable bbolt-backed implementation of JobStore.
type BoltJobStore struct {
	db     *bbolt.DB
	bucket []byte
	path   string
}

// NewBoltJobStore opens (or creates) a bbolt database at path and ensures the jobs bucket exists.
func NewBoltJobStore(path string) (*BoltJobStore, error) {
	db, err := bbolt.Open(path, 0600, &bbolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, err
	}
	if err := db.Update(func(tx *bbolt.Tx) error {
		_, e := tx.CreateBucketIfNotExists([]byte(boltJobsBucket))
		return e
	}); err != nil {
		if cerr := db.Close(); cerr != nil {
			logging.GetLogger().WarnWithFields("closing bolt db after bucket create failed", map[string]interface{}{"path": path, "err": cerr.Error()})
		}
		return nil, err
	}
	return &BoltJobStore{db: db, bucket: []byte(boltJobsBucket), path: path}, nil
}

// SaveResult stores or replaces a ConversionResult keyed by ConversionId.
func (s *BoltJobStore) SaveResult(res *focustypes.ConversionResult) error {
	if res == nil || res.ConversionId == "" {
		logging.GetLogger().ErrorWithFields("invalid conversion result", map[string]interface{}{"conversion_id": ""})
		telemetry.ConversionPersistenceFailure.Inc()
		return errors.New("invalid conversion result")
	}
	data, err := json.Marshal(res)
	if err != nil {
		logging.GetLogger().ErrorWithFields("marshal conversion result failed", map[string]interface{}{"conversion_id": res.ConversionId, "err": err.Error()})
		telemetry.ConversionPersistenceFailure.Inc()
		return err
	}
	return s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(s.bucket)
		if err := b.Put([]byte(res.ConversionId), data); err != nil {
			logging.GetLogger().ErrorWithFields("bolt put failed", map[string]interface{}{"conversion_id": res.ConversionId, "err": err.Error()})
			telemetry.ConversionPersistenceFailure.Inc()
			return err
		}
		logging.GetLogger().InfoWithFields("saved conversion result", map[string]interface{}{"conversion_id": res.ConversionId})
		telemetry.ConversionPersistenceSuccess.Inc()
		return nil
	})
}

// ListResults returns up to `limit` results ordered by StartTime ascending.
// If limit <= 0, all results are returned.
func (s *BoltJobStore) ListResults(limit int) []*focustypes.ConversionResult {
	var out []*focustypes.ConversionResult
	_ = s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(s.bucket)
		if b == nil {
			return nil
		}
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var r focustypes.ConversionResult
			if err := json.Unmarshal(v, &r); err != nil {
				// skip malformed
				continue
			}
			out = append(out, &r)
		}
		return nil
	})
	// sort by StartTime ascending
	if len(out) > 1 {
		// simple insertion sort for small N
		for i := 1; i < len(out); i++ {
			for j := i; j > 0 && out[j-1].StartTime.After(out[j].StartTime); j-- {
				out[j], out[j-1] = out[j-1], out[j]
			}
		}
	}
	if limit <= 0 || limit >= len(out) {
		// return a copy to avoid external mutation
		res := make([]*focustypes.ConversionResult, len(out))
		copy(res, out)
		return res
	}
	res := make([]*focustypes.ConversionResult, limit)
	copy(res, out[len(out)-limit:])
	return res
}

// FinalizeResultTiming updates the EndTime and Duration fields for an existing result.
func (s *BoltJobStore) FinalizeResultTiming(conversionID string, end time.Time, duration time.Duration) error {
	err := s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(s.bucket)
		if b == nil {
			return nil
		}
		v := b.Get([]byte(conversionID))
		if v == nil {
			return nil
		}
		var r focustypes.ConversionResult
		if err := json.Unmarshal(v, &r); err != nil {
			return err
		}
		r.EndTime = end
		r.Duration = duration
		data, err := json.Marshal(&r)
		if err != nil {
			return err
		}
		return b.Put([]byte(conversionID), data)
	})
	if err != nil {
		telemetry.ConversionPersistenceFailure.Inc()
	} else {
		telemetry.ConversionPersistenceSuccess.Inc()
	}
	return err
}

// Close closes the underlying bbolt database.
func (s *BoltJobStore) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

// RemoveOlderThan deletes entries whose EndTime (falling back to StartTime)
// is strictly before the provided cutoff. It returns the number of deleted
// entries.
func (s *BoltJobStore) RemoveOlderThan(cutoff time.Time) (int, error) {
	if s == nil || s.db == nil {
		return 0, nil
	}
	removed := 0
	err := s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(s.bucket)
		if b == nil {
			return nil
		}
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var r focustypes.ConversionResult
			if err := json.Unmarshal(v, &r); err != nil {
				// skip malformed
				continue
			}
			t := r.EndTime
			if t.IsZero() {
				t = r.StartTime
			}
			if !t.IsZero() && t.Before(cutoff) {
				if err := c.Delete(); err != nil {
					return err
				}
				removed++
			}
		}
		return nil
	})
	if err != nil {
		logging.GetLogger().ErrorWithFields("remove older than failed", map[string]interface{}{"cutoff": cutoff.String(), "err": err.Error()})
		telemetry.ConversionPersistenceFailure.Inc()
	} else {
		logging.GetLogger().InfoWithFields("remove older than completed", map[string]interface{}{"cutoff": cutoff.String(), "removed": removed})
		// treat removals as persistence success events for observability
		if removed > 0 {
			telemetry.ConversionPersistenceSuccess.Add(float64(removed))
		}
	}
	return removed, err
}

// Compact rewrites the underlying bbolt file to reclaim free pages and
// produce a compacted database file. Compact performs the rewrite into a
// temporary file in the same directory and atomically replaces the original
// database file.
func (s *BoltJobStore) Compact() error {
	if s == nil || s.db == nil {
		return nil
	}
	// ensure we have the path to work with
	if s.path == "" {
		return errors.New("bolt store path is unknown")
	}
	dir := filepath.Dir(s.path)
	tmpFile, err := os.CreateTemp(dir, "bolt-compact-*.db")
	if err != nil {
		return err
	}
	tmpPath := tmpFile.Name()
	if err := tmpFile.Close(); err != nil {
		if rerr := os.Remove(tmpPath); rerr != nil {
			logging.GetLogger().WarnWithFields("failed to remove tmp compact file after tmp close failure", map[string]interface{}{"tmp": tmpPath, "err": rerr.Error()})
		}
		return err
	}

	newDB, err := bbolt.Open(tmpPath, 0600, &bbolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		if rerr := os.Remove(tmpPath); rerr != nil {
			logging.GetLogger().WarnWithFields("failed to remove tmp compact file after open failure", map[string]interface{}{"tmp": tmpPath, "err": rerr.Error()})
		}
		return err
	}
	if err := newDB.Update(func(tx *bbolt.Tx) error {
		_, e := tx.CreateBucketIfNotExists([]byte(boltJobsBucket))
		return e
	}); err != nil {
		if cerr := newDB.Close(); cerr != nil {
			logging.GetLogger().WarnWithFields("failed to close new bolt db during compact error path", map[string]interface{}{"tmp": tmpPath, "err": cerr.Error()})
		}
		if rerr := os.Remove(tmpPath); rerr != nil {
			logging.GetLogger().WarnWithFields("failed to remove tmp compact file", map[string]interface{}{"tmp": tmpPath, "err": rerr.Error()})
		}
		return err
	}

	// copy entries into new DB
	if err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(s.bucket)
		if b == nil {
			return nil
		}
		return b.ForEach(func(k, v []byte) error {
			return newDB.Update(func(ntx *bbolt.Tx) error {
				nb := ntx.Bucket([]byte(boltJobsBucket))
				if nb == nil {
					return errors.New("destination bucket missing")
				}
				return nb.Put(k, v)
			})
		})
	}); err != nil {
		if cerr := newDB.Close(); cerr != nil {
			logging.GetLogger().WarnWithFields("failed to close new bolt db after copy failure", map[string]interface{}{"tmp": tmpPath, "err": cerr.Error()})
		}
		if rerr := os.Remove(tmpPath); rerr != nil {
			logging.GetLogger().WarnWithFields("failed to remove tmp compact file after copy failure", map[string]interface{}{"tmp": tmpPath, "err": rerr.Error()})
		}
		return err
	}

	if err := newDB.Close(); err != nil {
		logging.GetLogger().WarnWithFields("failed to close new bolt db after copy", map[string]interface{}{"tmp": tmpPath, "err": err.Error()})
		if rerr := os.Remove(tmpPath); rerr != nil {
			logging.GetLogger().WarnWithFields("failed to remove tmp compact file after close failure", map[string]interface{}{"tmp": tmpPath, "err": rerr.Error()})
		}
		return err
	}

	// close original DB so we can replace the file
	if err := s.db.Close(); err != nil {
		logging.GetLogger().WarnWithFields("failed to close original bolt db before rename", map[string]interface{}{"path": s.path, "err": err.Error()})
		if rerr := os.Remove(tmpPath); rerr != nil {
			logging.GetLogger().WarnWithFields("failed to remove tmp compact file after original close failure", map[string]interface{}{"tmp": tmpPath, "err": rerr.Error()})
		}
		return err
	}

	// atomically replace original file with compacted file
	if err := os.Rename(tmpPath, s.path); err != nil {
		return err
	}

	// reopen and set db
	db, err := bbolt.Open(s.path, 0600, &bbolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return err
	}
	s.db = db
	return nil
}
