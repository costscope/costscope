package security

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// AuditEvent represents a security-relevant event
type AuditEvent struct {
	Timestamp time.Time              `json:"timestamp"`
	Actor     string                 `json:"actor"`
	Action    string                 `json:"action"`
	Resource  string                 `json:"resource"`
	Result    string                 `json:"result"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
}

// Auditor writes audit events to an append-only JSONL file
type Auditor struct {
	path string
	mu   sync.Mutex
}

// NewAuditor creates a new auditor writing to data/security/audit.log
func NewAuditor(baseDir string) *Auditor {
	if baseDir == "" {
		baseDir = "data/security"
	}
	_ = os.MkdirAll(baseDir, 0o750)
	return &Auditor{path: filepath.Join(baseDir, "audit.log")}
}

// Write writes an audit event as a JSON line
func (a *Auditor) Write(evt AuditEvent) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	f, err := os.OpenFile(a.path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	b, err := json.Marshal(evt)
	if err != nil {
		return err
	}
	b = append(b, '\n')
	_, err = f.Write(b)
	return err
}
