package security

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"local/costscope/internal/core/logging"
	"local/costscope/internal/core/monitoring/telemetry"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Permission represents an action allowed on a resource
type Permission struct {
	Resource string `json:"resource"`
	Action   string `json:"action"`
}

// Role represents a named set of permissions
type Role struct {
	Name        string       `json:"name"`
	Description string       `json:"description,omitempty"`
	Permissions []Permission `json:"permissions"`
	CreatedAt   time.Time    `json:"created_at"`
}

// Assignment represents a user-to-role mapping
type Assignment struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
}

// RBACStore defines persistence for RBAC data
type RBACStore interface {
	Load() error
	Save() error
	AddRole(role Role) error
	GetRole(name string) (Role, bool)
	ListRoles() []Role
}

// fileRBACStore is a simple JSON-file backed RBAC store
type fileRBACStore struct {
	path  string
	roles map[string]Role
	mu    sync.RWMutex
}

// NewFileRBACStore creates a new file-backed RBAC store at data/security/roles.json
func NewFileRBACStore(baseDir string) *fileRBACStore {
	if baseDir == "" {
		baseDir = "data/security"
	}
	// Create directory with restrictive permissions
	_ = os.MkdirAll(baseDir, 0o750)
	return &fileRBACStore{
		path:  filepath.Join(baseDir, "roles.json"),
		roles: map[string]Role{},
	}
}

func (s *fileRBACStore) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	f, err := os.Open(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
	}()
	dec := json.NewDecoder(f)
	var roles []Role
	if err := dec.Decode(&roles); err != nil {
		return err
	}
	s.roles = map[string]Role{}
	for _, r := range roles {
		s.roles[r.Name] = r
	}
	return nil
}

func (s *fileRBACStore) Save() error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	tmp := make([]Role, 0, len(s.roles))
	for _, r := range s.roles {
		tmp = append(tmp, r)
	}
	f, err := os.Create(s.path)
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
	}()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(tmp)
}

func (s *fileRBACStore) AddRole(role Role) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.roles[role.Name]; exists {
		return fmt.Errorf("role %s already exists", role.Name)
	}
	s.roles[role.Name] = role
	return nil
}

func (s *fileRBACStore) GetRole(name string) (Role, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.roles[name]
	return r, ok
}

func (s *fileRBACStore) ListRoles() []Role {
	s.mu.RLock()
	defer s.mu.RUnlock()
	res := make([]Role, 0, len(s.roles))
	for _, r := range s.roles {
		res = append(res, r)
	}
	return res
}

// RBACService provides RBAC operations
type RBACService struct {
	store  RBACStore
	logger *logging.Logger
}

// NewRBACService creates a new RBAC service
func NewRBACService(store RBACStore, logger *logging.Logger) *RBACService {
	return &RBACService{store: store, logger: logger}
}

// CreateRole creates a new role with permissions
func (s *RBACService) CreateRole(name, description string, perms []Permission) (Role, error) {
	if name == "" {
		return Role{}, errors.New("role name is required")
	}
	role := Role{
		Name:        name,
		Description: description,
		Permissions: perms,
		CreatedAt:   time.Now().UTC(),
	}
	if err := s.store.AddRole(role); err != nil {
		return Role{}, err
	}
	if err := s.store.Save(); err != nil {
		return Role{}, err
	}
	if s.logger != nil {
		s.logger.InfoWithFields("RBAC: role created", map[string]interface{}{"role": name, "perms": len(perms)})
	}
	return role, nil
}

// HasPermission checks if a role has a specific permission
// Deprecated: Use CheckPermission for new code to ensure tracing & metrics.
func (s *RBACService) HasPermission(roleName, resource, action string) bool {
	role, ok := s.store.GetRole(roleName)
	if !ok {
		return false
	}
	for _, p := range role.Permissions {
		// Support wildcard semantics: a permission entry may specify Resource=="*" and/or Action=="*"
		// to allow any resource and/or action respectively. This keeps evaluation O(n) over permissions
		// while providing expressive baseline policy examples (e.g., admin *:*). Wildcards are limited
		// to a literal asterisk to avoid introducing glob parsing / regex overhead.
		resourceMatch := p.Resource == resource || p.Resource == "*"
		actionMatch := p.Action == action || p.Action == "*"
		if resourceMatch && actionMatch {
			return true
		}
	}
	return false
}

// CheckPermission is a tracing-enabled wrapper around HasPermission that records an OTel span.
// It does not introduce additional allocation heavy logic; span creation is skipped when
// no tracer provider is configured (noop provider).
// Span name: rbac.has_permission
// Attributes:
//
//	rbac.role      - evaluated role
//	rbac.resource  - target resource
//	rbac.action    - target action
//	rbac.allowed   - boolean result
//
// NOTE: We keep HasPermission for lightweight internal / test usage; middleware and
// centralized enforcement paths SHOULD prefer CheckPermission to gain observability.
func (s *RBACService) CheckPermission(ctx context.Context, roleName, resource, action string) bool {
	start := time.Now()
	tracer := otel.Tracer("costscope.rbac")
	_, span := tracer.Start(ctx, "rbac.has_permission", trace.WithSpanKind(trace.SpanKindInternal))
	span.SetAttributes(
		attribute.String("rbac.role", roleName),
		attribute.String("rbac.resource", resource),
		attribute.String("rbac.action", action),
	)
	allowed := s.HasPermission(roleName, resource, action)
	span.SetAttributes(attribute.Bool("rbac.allowed", allowed))
	span.End()
	// Metrics (labels kept low cardinality: resource/action must be bounded enums)
	outcome := "denied"
	if allowed {
		outcome = "allowed"
	}
	telemetry.RBACChecksTotal.WithLabelValues(resource, action, outcome).Inc()
	telemetry.RBACCheckLatency.WithLabelValues(resource, action).Observe(time.Since(start).Seconds())
	return allowed
}

// Predefined resource & action constants (grow conservatively to limit metric cardinality)
const (
	ResourceReports = "reports"
	ActionGenerate  = "generate"
	ActionExport    = "export"

	ResourceFocus  = "focus"
	ActionConvert  = "convert"
	ActionValidate = "validate"

	ResourceProviders = "providers"
	ActionConnect     = "connect"
	ActionList        = "list"

	ResourceAnalytics     = "analytics"
	ActionForecast        = "forecast"
	ActionDetectAnomalies = "anomalies"
	ActionRecommendations = "recommendations"
	ActionTrends          = "trends"
	ActionTrainModel      = "train_model"

	ResourceStreaming = "streaming"
	ActionCreateJob   = "create_job"
	ActionStartJob    = "start_job"
	ActionStopJob     = "stop_job"
	ActionDeleteJob   = "delete_job"
)

// PermissionMatrix documents intended resource/action pairs for validation / audit. Not enforced yet.
var PermissionMatrix = map[string][]string{
	ResourceReports:   {ActionGenerate, ActionExport},
	ResourceFocus:     {ActionConvert, ActionValidate},
	ResourceProviders: {ActionConnect, ActionList},
	ResourceAnalytics: {ActionForecast, ActionDetectAnomalies, ActionRecommendations, ActionTrends, ActionTrainModel},
	ResourceStreaming: {ActionCreateJob, ActionStartJob, ActionStopJob, ActionDeleteJob},
}
