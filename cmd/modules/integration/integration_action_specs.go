package integration

// integration_action_specs.go
// Canonical registrar & declarative specs for integration actions (TASK-INTEGRATION-REGISTRAR)
// Moved from action_specs.go (kept as thin shim) to satisfy naming requirement.

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/costscope/costscope/cmd/modules/integration/alerts"
	"github.com/costscope/costscope/cmd/modules/integration/connections"
	"github.com/costscope/costscope/cmd/modules/integration/dashboard"
	"github.com/costscope/costscope/cmd/modules/integration/webhooks"
	"github.com/costscope/costscope/internal/core/logging"
	"github.com/costscope/costscope/internal/core/monitoring/telemetry"
)

// FlagSpec describes a CLI flag for an action.
type FlagSpec struct {
	Name      string
	Type      string // string,bool,int,duration,stringSlice,stringToString,float64
	Usage     string
	Default   interface{}
	Required  bool
	Shorthand string // optional
	// Extended metadata (optional)
	Hidden     bool
	Deprecated string
	Env        string
	Choices    []string
}

// ActionSpec describes an integration action → Cobra command mapping.
type ActionSpec struct {
	ID             string
	Category       string
	Use            string
	Short          string
	Long           string
	Example        string
	Flags          []FlagSpec
	FeatureGate    string
	Parents        []string // optional chain of parent command names inside category (e.g. ["delivery"])
	Group          bool     // when true this spec only ensures a parent/group command exists (no handler)
	HandlerFactory func(ctx *RegistrationContext, spec ActionSpec) func(cmd *cobra.Command, args []string) error
	// Optional runtime constraints for flags
	MutuallyExclusive [][]string // groups of flag names where at most one may be set
	AtLeastOneOf      []string   // require at least one flag from this set
}

// Sentinel errors for typed classification (optional use by handlers)
var (
	ErrNotFound     = errors.New("integration:not_found")
	ErrTimeout      = errors.New("integration:timeout")
	ErrUnauthorized = errors.New("integration:unauthorized")
	ErrValidation   = errors.New("integration:validation")
	ErrConflict     = errors.New("integration:conflict")
)

// classifyIntegrationError classifies an error into a stable label. Prefers sentinel errors then falls back to message inspection.
func classifyIntegrationError(err error) string { // TASK-INTEGRATION-ERROR-METRICS (enhanced typed)
	if err == nil {
		return ""
	}
	switch {
	case errors.Is(err, ErrNotFound):
		return "not_found"
	case errors.Is(err, ErrTimeout):
		return "timeout"
	case errors.Is(err, ErrUnauthorized):
		return "unauthorized"
	case errors.Is(err, ErrValidation):
		return "validation"
	case errors.Is(err, ErrConflict):
		return "conflict"
	}
	m := strings.ToLower(err.Error())
	switch {
	case strings.Contains(m, "not found") || strings.Contains(m, "missing"):
		return "not_found"
	case strings.Contains(m, "timeout") || strings.Contains(m, "deadline"):
		return "timeout"
	case strings.Contains(m, "unauthorized") || strings.Contains(m, "forbidden") || strings.Contains(m, "permission"):
		return "unauthorized"
	case strings.Contains(m, "invalid") || strings.Contains(m, "must ") || strings.Contains(m, "required"):
		return "validation"
	case strings.Contains(m, "conflict") || strings.Contains(m, "already exists"):
		return "conflict"
	default:
		return "other"
	}
}

// RegistrationContext holds managers so factories stay thin.
type RegistrationContext struct {
	WebhookMgr   *webhooks.WebhookManager
	DashboardMgr *dashboard.DashboardManager
	ConnMgr      *connections.ConnectionManager
	AlertMgr     *alerts.AlertManager // reserved (not yet migrated in this task)
}

// RegisterIntegrationActions registers all provided specs under the integration root.
func RegisterIntegrationActions(root *cobra.Command, ctx *RegistrationContext, specs []ActionSpec) map[string]*cobra.Command {
	parents := ensureCategoryParents(root, []string{"webhook", "dashboard", "connections"})
	registered := make(map[string]*cobra.Command)
	log := logging.GetLogger()
	for _, spec := range specs {
		parent, ok := parents[spec.Category]
		if !ok {
			log.WarnWithFields("integration action category missing", map[string]interface{}{"action_id": spec.ID, "category": spec.Category})
			continue
		}
		current := ensureParentChain(parent, spec)
		// group spec creation (no handler/flags)
		if spec.Group {
			grp := ensureGroupNode(current, spec)
			registered[spec.ID] = grp
			log.InfoWithFields("integration action group registered", map[string]interface{}{"action_id": spec.ID, "category": spec.Category, "group": true})
			continue
		}
		cmd := &cobra.Command{Use: spec.Use, Short: spec.Short, Long: spec.Long, Example: spec.Example}
		if cmd.Annotations == nil {
			cmd.Annotations = map[string]string{}
		}
		cmd.Annotations["category"] = spec.Category
		wireFlags(cmd, spec)
		if spec.HandlerFactory != nil {
			baseHandler := spec.HandlerFactory(ctx, spec)
			cmd.RunE = wrapRunEWithValidationAndTelemetry(baseHandler, spec)
		}
		current.AddCommand(cmd)
		registered[spec.ID] = cmd
		log.InfoWithFields("integration action registered", map[string]interface{}{"action_id": spec.ID, "category": spec.Category, "registered": true})
	}
	return registered
}

// ensureCategoryParents makes sure top-level category commands exist and returns a map.
func ensureCategoryParents(root *cobra.Command, cats []string) map[string]*cobra.Command {
	parents := map[string]*cobra.Command{}
	for _, cat := range cats {
		var existing *cobra.Command
		for _, c := range root.Commands() {
			if c.Use == cat || strings.HasPrefix(c.Use, cat+" ") {
				existing = c
				break
			}
		}
		if existing == nil {
			existing = &cobra.Command{Use: cat, Short: fmt.Sprintf("%s operations", cat)}
			root.AddCommand(existing)
		}
		parents[cat] = existing
	}
	return parents
}

// ensureParentChain walks/creates the parent chain for a spec and returns the final parent.
func ensureParentChain(parent *cobra.Command, spec ActionSpec) *cobra.Command {
	current := parent
	for _, p := range spec.Parents {
		var next *cobra.Command
		for _, c := range current.Commands() {
			if c.Use == p {
				next = c
				break
			}
		}
		if next == nil {
			next = &cobra.Command{Use: p, Short: fmt.Sprintf("%s %s", spec.Category, p)}
			current.AddCommand(next)
		}
		current = next
	}
	return current
}

// ensureGroupNode finds or creates a group node and annotates it.
func ensureGroupNode(current *cobra.Command, spec ActionSpec) *cobra.Command {
	var grp *cobra.Command
	for _, c := range current.Commands() {
		if c.Use == spec.Use {
			grp = c
			break
		}
	}
	if grp == nil {
		grp = &cobra.Command{Use: spec.Use}
		current.AddCommand(grp)
	}
	if spec.Short != "" {
		grp.Short = spec.Short
	}
	if spec.Long != "" {
		grp.Long = spec.Long
	}
	if grp.Annotations == nil {
		grp.Annotations = map[string]string{}
	}
	grp.Annotations["category"] = spec.Category
	return grp
}

// wireFlags defines flags on a command and applies metadata/annotations.
func wireFlags(cmd *cobra.Command, spec ActionSpec) {
	log := logging.GetLogger()
	for _, f := range spec.Flags {
		switch f.Type {
		case flagTypeString:
			defVal, _ := f.Default.(string)
			cmd.Flags().StringP(f.Name, f.Shorthand, defVal, f.Usage)
		case flagTypeBool:
			defVal, _ := f.Default.(bool)
			cmd.Flags().BoolP(f.Name, f.Shorthand, defVal, f.Usage)
		case flagTypeInt:
			defVal, _ := f.Default.(int)
			cmd.Flags().IntP(f.Name, f.Shorthand, defVal, f.Usage)
		case flagTypeFloat64:
			defVal, _ := f.Default.(float64)
			cmd.Flags().Float64P(f.Name, f.Shorthand, defVal, f.Usage)
		case flagTypeDuration:
			defVal, _ := f.Default.(time.Duration)
			cmd.Flags().DurationP(f.Name, f.Shorthand, defVal, f.Usage)
		case flagTypeStringSlice:
			defVal, _ := f.Default.([]string)
			cmd.Flags().StringSliceP(f.Name, f.Shorthand, defVal, f.Usage)
		case flagTypeStringToString:
			defVal, _ := f.Default.(map[string]string)
			cmd.Flags().StringToStringP(f.Name, f.Shorthand, defVal, f.Usage)
		default:
			log.WarnWithFields("unsupported flag type", map[string]interface{}{"action_id": spec.ID, "flag": f.Name, "type": f.Type})
		}
		if fl := cmd.Flags().Lookup(f.Name); fl != nil {
			if f.Hidden {
				fl.Hidden = true
			}
			if f.Deprecated != "" {
				fl.Deprecated = f.Deprecated
			}
			if f.Env != "" {
				if fl.Annotations == nil {
					fl.Annotations = map[string][]string{}
				}
				fl.Annotations["env"] = []string{f.Env}
			}
			if len(f.Choices) > 0 {
				if fl.Annotations == nil {
					fl.Annotations = map[string][]string{}
				}
				fl.Annotations["choices"] = append([]string{}, f.Choices...)
			}
		}
		if f.Required {
			_ = cmd.MarkFlagRequired(f.Name)
		}
	}
}

// wrapRunEWithValidationAndTelemetry returns a cobra RunE with validation, tracing, and metrics.
func wrapRunEWithValidationAndTelemetry(baseHandler func(cmd *cobra.Command, args []string) error, spec ActionSpec) func(*cobra.Command, []string) error {
	return func(c *cobra.Command, args []string) error {
		start := time.Now()
		tracer := otel.Tracer("costscope.integration")
		ctx := c.Context()
		ctx, span := tracer.Start(ctx, "integration.action", trace.WithSpanKind(trace.SpanKindInternal))
		span.SetAttributes(
			attribute.String("integration.action_id", spec.ID),
			attribute.String("integration.category", spec.Category),
		)
		valCtx, valSpan := tracer.Start(ctx, "integration.action.validation")
		valSpan.SetAttributes(attribute.Int("integration.flags.count", c.Flags().NFlag()))
		if err := validateFlagRuntimeConstraints(c, spec); err != nil {
			valSpan.End()
			span.End()
			return err
		}
		valSpan.End()
		ioCtx, ioSpan := tracer.Start(valCtx, "integration.action.io")
		statusLabel := "success"
		errType := ""
		execCtx, execSpan := tracer.Start(ioCtx, "integration.action.execute")
		_ = execCtx
		err := baseHandler(c, args)
		execSpan.End()
		if err != nil {
			statusLabel = "error"
			errType = classifyIntegrationError(err)
			span.RecordError(err)
			span.SetAttributes(attribute.String("integration.error_type", errType))
		}
		ioSpan.End()
		span.SetAttributes(attribute.String("integration.status", statusLabel))
		telemetry.IntegrationActionCalls.With(prometheus.Labels{"action_id": spec.ID, "category": spec.Category, "status": statusLabel}).Inc()
		telemetry.IntegrationActionDuration.With(prometheus.Labels{"action_id": spec.ID, "category": spec.Category, "status": statusLabel}).Observe(time.Since(start).Seconds())
		if errType != "" {
			telemetry.IntegrationActionErrors.With(prometheus.Labels{"action_id": spec.ID, "category": spec.Category, "error_type": errType}).Inc()
		}
		span.End()
		return err
	}
}

// validateFlagRuntimeConstraints checks mutuallyExclusive, atLeastOneOf, and choices against provided flags.
func validateFlagRuntimeConstraints(c *cobra.Command, spec ActionSpec) error {
	changed := map[string]bool{}
	c.Flags().Visit(func(f *pflag.Flag) { changed[f.Name] = true })
	// MutuallyExclusive: for each group ensure <=1 set
	for _, grp := range spec.MutuallyExclusive {
		count := 0
		var setNames []string
		for _, name := range grp {
			if changed[name] {
				count++
				setNames = append(setNames, name)
			}
		}
		if count > 1 {
			return fmt.Errorf("invalid flag combination: %v are mutually exclusive", setNames)
		}
	}
	// AtLeastOneOf: ensure at least one of the set is provided
	if len(spec.AtLeastOneOf) > 0 {
		ok := false
		for _, name := range spec.AtLeastOneOf {
			if changed[name] {
				ok = true
				break
			}
		}
		if !ok {
			return fmt.Errorf("at least one of flags must be set: %v", spec.AtLeastOneOf)
		}
	}
	// Choices validation for string and stringSlice (best-effort)
	choiceSet := map[string]map[string]struct{}{}
	for _, f := range spec.Flags {
		if len(f.Choices) == 0 {
			continue
		}
		set := map[string]struct{}{}
		for _, v := range f.Choices {
			set[strings.ToLower(v)] = struct{}{}
		}
		choiceSet[f.Name] = set
	}
	for name, allowed := range choiceSet {
		if !changed[name] {
			continue
		}
		if fv, err := c.Flags().GetString(name); err == nil {
			if _, ok := allowed[strings.ToLower(fv)]; !ok {
				return fmt.Errorf("invalid value for --%s: %s (allowed: %v)", name, fv, keys(allowed))
			}
			continue
		}
		if arr, err := c.Flags().GetStringSlice(name); err == nil {
			for _, it := range arr {
				if _, ok := allowed[strings.ToLower(it)]; !ok {
					return fmt.Errorf("invalid value for --%s: %s (allowed: %v)", name, it, keys(allowed))
				}
			}
		}
	}
	return nil
}

// BuildDefaultActionSpecs is generated by scripts/tools/gen-actions.
// See generated file cmd/modules/integration/generated_actions.go

// helper: keys of map[string]struct{}
func keys(m map[string]struct{}) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
