package integration

// dsl.go
// Declarative Integration Action DSL (Go struct + YAML ingestion) and converter to ActionSpec.

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// DSLFlag defines a flag in DSL form (YAML ingestable)
type DSLFlag struct {
	Name      string      `yaml:"name"`
	Type      string      `yaml:"type"` // string,bool,int,duration,stringSlice,stringToString,float64
	Usage     string      `yaml:"usage"`
	Default   interface{} `yaml:"default,omitempty"`
	Required  bool        `yaml:"required,omitempty"`
	Shorthand string      `yaml:"shorthand,omitempty"`
	// Extended metadata
	Hidden     bool     `yaml:"hidden,omitempty"`
	Deprecated string   `yaml:"deprecated,omitempty"`
	Env        string   `yaml:"env,omitempty"`
	Choices    []string `yaml:"choices,omitempty"`
}

// DSLAction defines a single action in the DSL
type DSLAction struct {
	ID          string    `yaml:"id"`
	Category    string    `yaml:"category"`
	Use         string    `yaml:"use"`
	Short       string    `yaml:"short"`
	Long        string    `yaml:"long,omitempty"`
	Example     string    `yaml:"example,omitempty"`
	Flags       []DSLFlag `yaml:"flags,omitempty"`
	FeatureGate string    `yaml:"featureGate,omitempty"`
	Parents     []string  `yaml:"parents,omitempty"`
	Group       bool      `yaml:"group,omitempty"`
	// Optional extras (not used in v1 generation but reserved)
	InputsSchema map[string]any `yaml:"inputsSchema,omitempty"`
	AuthScope    string         `yaml:"authScope,omitempty"`
	// Constraints (additive, optional)
	MutuallyExclusive [][]string `yaml:"mutuallyExclusive,omitempty"` // list of flag-name groups where only one may be set
	AtLeastOneOf      []string   `yaml:"atLeastOneOf,omitempty"`      // require at least one flag from this list
}

// ActionDSL is the top-level YAML document
type ActionDSL struct {
	Version string      `yaml:"version"`
	Actions []DSLAction `yaml:"actions"`
}

// LoadActionDSL loads the YAML file from path
func LoadActionDSL(path string) (*ActionDSL, error) {
	// Path comes from internal generator/config, not untrusted user input.
	// #nosec G304 -- controlled file inclusion by design for SSOT YAML
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var d ActionDSL
	if err := yaml.Unmarshal(b, &d); err != nil {
		return nil, err
	}
	return &d, nil
}

// Convert the DSL into registrar ActionSpecs (HandlerFactory is nil; wired by registrar)
func (d *ActionDSL) ToActionSpecs() ([]ActionSpec, error) {
	out := make([]ActionSpec, 0, len(d.Actions))
	for _, a := range d.Actions {
		as := ActionSpec{
			ID:                a.ID,
			Category:          a.Category,
			Use:               a.Use,
			Short:             a.Short,
			Long:              a.Long,
			Example:           a.Example,
			FeatureGate:       a.FeatureGate,
			Parents:           a.Parents,
			Group:             a.Group,
			MutuallyExclusive: a.MutuallyExclusive,
			AtLeastOneOf:      a.AtLeastOneOf,
		}
		// Convert flags with type-aware defaults
		for _, f := range a.Flags {
			fs := FlagSpec{Name: f.Name, Type: f.Type, Usage: f.Usage, Required: f.Required, Shorthand: f.Shorthand, Hidden: f.Hidden, Deprecated: f.Deprecated, Env: f.Env, Choices: f.Choices}
			switch f.Type {
			case flagTypeString:
				if v, ok := f.Default.(string); ok {
					fs.Default = v
				}
			case flagTypeBool:
				if v, ok := f.Default.(bool); ok {
					fs.Default = v
				}
			case flagTypeInt:
				// YAML may parse numbers as int or float64; handle both
				switch v := f.Default.(type) {
				case int:
					fs.Default = v
				case int64:
					fs.Default = int(v)
				case float64:
					fs.Default = int(v)
				}
			case flagTypeFloat64:
				switch v := f.Default.(type) {
				case float64:
					fs.Default = v
				case int:
					fs.Default = float64(v)
				case int64:
					fs.Default = float64(v)
				}
			case flagTypeDuration:
				// Allow seconds as number or Go duration string
				switch v := f.Default.(type) {
				case string:
					if dur, err := time.ParseDuration(v); err == nil {
						fs.Default = dur
					}
				case int:
					fs.Default = time.Duration(v) * time.Second
				case int64:
					fs.Default = time.Duration(v) * time.Second
				case float64:
					fs.Default = time.Duration(int(v)) * time.Second
				}
			case flagTypeStringSlice:
				// YAML can decode sequence into []any; normalize to []string
				switch vv := f.Default.(type) {
				case []any:
					arr := make([]string, 0, len(vv))
					for _, it := range vv {
						arr = append(arr, fmt.Sprint(it))
					}
					fs.Default = arr
				case []string:
					fs.Default = vv
				}
			case flagTypeStringToString:
				switch vv := f.Default.(type) {
				case map[string]string:
					fs.Default = vv
				case map[string]any:
					m := map[string]string{}
					for k, v := range vv {
						m[k] = fmt.Sprint(v)
					}
					fs.Default = m
				}
			default:
				// Keep type string as-is; validation catches unsupported types.
			}
			as.Flags = append(as.Flags, fs)
		}
		out = append(out, as)
	}
	return out, nil
}

// Validate performs structural validation of the DSL document.
// Returns nil when valid or an error aggregating all issues found.
func (d *ActionDSL) Validate() error {
	var errs []string
	// basic version presence (accept any non-empty for now)
	if strings.TrimSpace(d.Version) == "" {
		errs = append(errs, "version must be set (e.g., v1)")
	}
	// ID uniqueness and required fields
	seen := map[string]struct{}{}
	// path uniqueness: category + parents + use (groups included)
	seenPaths := map[string]string{} // path -> actionID
	for i, a := range d.Actions {
		// core action validations (id/category/use/group/auth/schema)
		errs = append(errs, validateActionCore(a, i)...)
		// id uniqueness
		if a.ID != "" {
			if _, ok := seen[a.ID]; ok {
				errs = append(errs, fmt.Sprintf("duplicate id: %s", a.ID))
			}
			seen[a.ID] = struct{}{}
		}
		// compute path uniqueness key
		pathKey := buildActionPathKey(a)
		if prior, ok := seenPaths[pathKey]; ok {
			errs = append(errs, fmt.Sprintf("action[%s]: path '%s' conflicts with action[%s]", a.ID, pathKey, prior))
		} else {
			seenPaths[pathKey] = a.ID
		}
		// flags + constraints validations
		flagErrs, seenFlagNames := validateFlags(a)
		errs = append(errs, flagErrs...)
		errs = append(errs, validateConstraints(a, seenFlagNames)...)
	}
	if len(errs) == 0 {
		return nil
	}
	return errors.New(strings.Join(errs, "; "))
}

// buildActionPathKey returns the unique path key for an action
func buildActionPathKey(a DSLAction) string {
	pathSegs := append([]string{a.Category}, a.Parents...)
	if a.Use != "" { // groups still have use as the group node name
		pathSegs = append(pathSegs, a.Use)
	}
	return strings.Join(pathSegs, "/")
}

// validateActionCore performs validations that don't need cross-action state.
func validateActionCore(a DSLAction, index int) []string {
	var errs []string
	if a.ID == "" {
		errs = append(errs, fmt.Sprintf("action[%d]: id is required", index))
	}
	if a.Category == "" {
		errs = append(errs, fmt.Sprintf("action[%s]: category is required", a.ID))
	}
	if a.Use == "" && !a.Group {
		errs = append(errs, fmt.Sprintf("action[%s]: use is required for non-group actions", a.ID))
	}
	if a.Group && len(a.Flags) > 0 {
		errs = append(errs, fmt.Sprintf("action[%s]: group actions must not define flags", a.ID))
	}
	if strings.TrimSpace(a.AuthScope) != "" && !isValidAuthScope(a.AuthScope) {
		errs = append(errs, fmt.Sprintf("action[%s]: invalid authScope '%s' (allowed: [public, authenticated, reader, operator, admin])", a.ID, a.AuthScope))
	}
	if len(a.Flags) > 0 {
		hasRequired := false
		for _, f := range a.Flags {
			if f.Required {
				hasRequired = true
				break
			}
		}
		if hasRequired && len(a.InputsSchema) == 0 {
			errs = append(errs, fmt.Sprintf("action[%s]: inputsSchema must be provided when required flags are defined", a.ID))
		}
	}
	return errs
}

// validateFlags checks flags definitions and returns errors and the set of seen flag names.
func validateFlags(a DSLAction) ([]string, map[string]struct{}) {
	var errs []string
	seenFlagNames := map[string]struct{}{}
	for _, f := range a.Flags {
		if f.Name == "" {
			errs = append(errs, fmt.Sprintf("action[%s]: flag with empty name", a.ID))
		}
		if !isAllowedFlagType(f.Type) {
			errs = append(errs, fmt.Sprintf("action[%s]: flag '%s' has unsupported type '%s'", a.ID, f.Name, f.Type))
		}
		if f.Shorthand != "" && len([]rune(f.Shorthand)) != 1 {
			errs = append(errs, fmt.Sprintf("action[%s]: flag '%s' shorthand must be a single character", a.ID, f.Name))
		}
		// choices only valid for string and stringSlice (apply De Morgan's for staticcheck suggestion)
		if len(f.Choices) > 0 && f.Type != flagTypeString && f.Type != flagTypeStringSlice {
			errs = append(errs, fmt.Sprintf("action[%s]: flag '%s' defines choices but type is '%s' (only string|stringSlice supported)", a.ID, f.Name, f.Type))
		}
		if _, dup := seenFlagNames[f.Name]; dup {
			errs = append(errs, fmt.Sprintf("action[%s]: duplicate flag name '%s'", a.ID, f.Name))
		} else {
			seenFlagNames[f.Name] = struct{}{}
		}
	}
	return errs, seenFlagNames
}

// validateConstraints ensures constraint references point to existing flags and have valid shapes.
func validateConstraints(a DSLAction, seenFlagNames map[string]struct{}) []string {
	var errs []string
	if len(a.MutuallyExclusive) > 0 {
		for gi, grp := range a.MutuallyExclusive {
			if len(grp) < 2 {
				errs = append(errs, fmt.Sprintf("action[%s]: mutuallyExclusive group[%d] must contain at least 2 flags", a.ID, gi))
				continue
			}
			for _, fname := range grp {
				if _, ok := seenFlagNames[fname]; !ok {
					errs = append(errs, fmt.Sprintf("action[%s]: mutuallyExclusive references unknown flag '%s'", a.ID, fname))
				}
			}
		}
	}
	if len(a.AtLeastOneOf) > 0 {
		for _, fname := range a.AtLeastOneOf {
			if _, ok := seenFlagNames[fname]; !ok {
				errs = append(errs, fmt.Sprintf("action[%s]: atLeastOneOf references unknown flag '%s'", a.ID, fname))
			}
		}
	}
	return errs
}

func isAllowedFlagType(t string) bool {
	switch t {
	case flagTypeString, flagTypeBool, flagTypeInt, flagTypeDuration, flagTypeStringSlice, flagTypeStringToString, flagTypeFloat64:
		return true
	default:
		return false
	}
}

// isValidAuthScope validates allowed auth scopes (lightweight; keep additive-friendly)
func isValidAuthScope(s string) bool {
	switch s {
	case "public", "authenticated", "reader", "operator", "admin":
		return true
	default:
		// allow dotted forms for future role namespaces (e.g., "role:admin" is validated by RBAC elsewhere)
		if strings.HasPrefix(s, "role:") && len(s) > len("role:") {
			return true
		}
		return false
	}
}
