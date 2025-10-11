package config

// Retained minimal shared interfaces/types still referenced by active code.

// Logger interface for minimal logging dependency (kept for env_resolver and precedence packages).
type Logger interface {
	Infof(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Debugf(format string, args ...interface{})
	Warnf(format string, args ...interface{})
}

// ConfigError preserved for any remaining validation or precedence error pathways.
type ConfigError struct {
	Section ConfigSection
	Key     string
	Message string
	Err     error
}

func (e *ConfigError) Error() string {
	if e.Section != "" && e.Key != "" {
		return "config error in section '" + string(e.Section) + "', key '" + e.Key + "': " + e.Message
	}
	if e.Section != "" {
		return "config error in section '" + string(e.Section) + "': " + e.Message
	}
	return "config error: " + e.Message
}

func (e *ConfigError) Unwrap() error { return e.Err }

func NewConfigError(section ConfigSection, key, message string, err error) *ConfigError {
	return &ConfigError{Section: section, Key: key, Message: message, Err: err}
}
