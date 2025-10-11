package logging

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// LogLevel represents log levels
type LogLevel string

const (
	LevelDebug LogLevel = "debug"
	LevelInfo  LogLevel = "info"
	LevelWarn  LogLevel = "warn"
	LevelError LogLevel = "error"
	LevelFatal LogLevel = "fatal"
)

// Logger is a basic logger implementation
type Logger struct {
	level LogLevel
	base  map[string]interface{}
	// maxLogBytes limits a single log line size to avoid flooding; 0 means default (8192)
	maxLogBytes int
}

var (
	defaultLogger   *Logger
	defaultLoggerMu sync.RWMutex // protects defaultLogger initialization and replacement
)

// NewLogger creates a new logger with specified level
func NewLogger(level LogLevel) *Logger {
	return &Logger{
		level:       level,
		base:        map[string]interface{}{"service": "costscope"},
		maxLogBytes: getMaxLogBytesFromEnv(),
	}
}

// ctxKey is a typed key for storing correlation IDs in context.
type ctxKey string

// Exported typed context keys for correlation
// Prefer these over plain strings to avoid collisions.
var (
	CtxRequestID ctxKey = "request_id"
	CtxTraceID   ctxKey = "trace_id"
	CtxSpanID    ctxKey = "span_id"
)

// ContextWithIDs returns a derived context containing correlation IDs using typed keys.
func ContextWithIDs(ctx context.Context, reqID, traceID, spanID string) context.Context {
	if reqID != "" {
		ctx = context.WithValue(ctx, CtxRequestID, reqID)
	}
	if traceID != "" {
		ctx = context.WithValue(ctx, CtxTraceID, traceID)
	}
	if spanID != "" {
		ctx = context.WithValue(ctx, CtxSpanID, spanID)
	}
	return ctx
}

// getMaxLogBytesFromEnv returns max log bytes from env or default
func getMaxLogBytesFromEnv() int {
	v := os.Getenv("COSTSCOPE_LOG_MAX_BYTES")
	if v == "" {
		return 8192
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return 8192
	}
	return n
}

// IsEnabledFor checks if the logger is enabled for the given level
func (l *Logger) IsEnabledFor(level LogLevel) bool {
	levels := map[LogLevel]int{
		LevelDebug: 0,
		LevelInfo:  1,
		LevelWarn:  2,
		LevelError: 3,
		LevelFatal: 4,
	}
	return levels[level] >= levels[l.level]
}

// log is the internal logging method
func (l *Logger) log(level LogLevel, msg string) {
	if l.IsEnabledFor(level) {
		timestamp := time.Now().UTC().Format(time.RFC3339Nano)
		entry := map[string]interface{}{
			"ts":    timestamp,
			"level": string(level),
			"msg":   redactValue(msg),
		}
		for k, v := range l.base {
			entry[k] = v
		}

		// Marshal as JSON
		b, _ := json.Marshal(entry)
		out := string(b)
		if l.maxLogBytes > 0 && len(out) > l.maxLogBytes {
			// Truncate safely
			suffix := "... [truncated]"
			cut := l.maxLogBytes - len(suffix)
			if cut < 0 {
				cut = 0
			}
			out = out[:cut] + suffix
		}
		fmt.Fprintln(os.Stderr, out)
	}
}

// Debug logs a debug message
func (l *Logger) Debug(msg string) {
	l.log(LevelDebug, msg)
}

// Info logs an info message
func (l *Logger) Info(msg string) {
	l.log(LevelInfo, msg)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string) {
	l.log(LevelWarn, msg)
}

// Error logs an error message
func (l *Logger) Error(msg string) {
	l.log(LevelError, msg)
}

// Formatted logging helpers (implement config.Logger interface)
func (l *Logger) Infof(format string, args ...interface{}) {
	l.log(LevelInfo, fmt.Sprintf(format, args...))
}
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.log(LevelError, fmt.Sprintf(format, args...))
}
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.log(LevelDebug, fmt.Sprintf(format, args...))
}
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.log(LevelWarn, fmt.Sprintf(format, args...))
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(msg string) {
	l.log(LevelFatal, msg)
	os.Exit(1)
}

// GetLogger returns a global logger instance (for compatibility)
func GetLogger() *Logger {
	// Fast path with read lock
	defaultLoggerMu.RLock()
	dl := defaultLogger
	defaultLoggerMu.RUnlock()
	if dl != nil {
		return dl
	}
	// Initialize under write lock if still nil
	defaultLoggerMu.Lock()
	if defaultLogger == nil { // double-check
		defaultLogger = NewLogger(LevelInfo)
	}
	dl = defaultLogger
	defaultLoggerMu.Unlock()
	return dl
}

// SetDefaultLogger sets the process-wide default logger
func SetDefaultLogger(l *Logger) {
	defaultLoggerMu.Lock()
	defaultLogger = l
	defaultLoggerMu.Unlock()
}

// FromContext returns a child logger with correlation fields from context if present
// Recognized keys: request_id, trace_id, span_id
func FromContext(ctx interface{ Value(any) any }) *Logger {
	base := GetLogger()
	if base == nil {
		base = NewLogger(LevelInfo)
		SetDefaultLogger(base)
	}
	fields := map[string]interface{}{}
	// Prefer typed keys
	if v := ctx.Value(CtxRequestID); v != nil {
		fields["request_id"] = v
	}
	if v := ctx.Value(CtxTraceID); v != nil {
		fields["trace_id"] = v
	}
	if v := ctx.Value(CtxSpanID); v != nil {
		fields["span_id"] = v
	}
	// Fallback to legacy string keys
	if _, ok := fields["request_id"]; !ok {
		if v := ctx.Value("request_id"); v != nil {
			fields["request_id"] = v
		}
	}
	if _, ok := fields["trace_id"]; !ok {
		if v := ctx.Value("trace_id"); v != nil {
			fields["trace_id"] = v
		}
	}
	if _, ok := fields["span_id"]; !ok {
		if v := ctx.Value("span_id"); v != nil {
			fields["span_id"] = v
		}
	}
	if len(fields) == 0 {
		return base
	}
	return base.WithFields(fields)
}

// InfoWithFields logs an info message with structured fields
func (l *Logger) InfoWithFields(msg string, fields map[string]interface{}) {
	l.logWithFields(LevelInfo, msg, fields)
}

// ErrorWithFields logs an error message with structured fields
func (l *Logger) ErrorWithFields(msg string, fields map[string]interface{}) {
	l.logWithFields(LevelError, msg, fields)
}

// DebugWithFields logs debug with fields
func (l *Logger) DebugWithFields(msg string, fields map[string]interface{}) {
	l.logWithFields(LevelDebug, msg, fields)
}

// WarnWithFields logs warn with fields
func (l *Logger) WarnWithFields(msg string, fields map[string]interface{}) {
	l.logWithFields(LevelWarn, msg, fields)
}

// FatalWithFields logs fatal with fields then exits
func (l *Logger) FatalWithFields(msg string, fields map[string]interface{}) {
	l.logWithFields(LevelFatal, msg, fields)
	os.Exit(1)
}

// WithFields returns a child logger that always includes these fields
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	child := &Logger{level: l.level, base: map[string]interface{}{}, maxLogBytes: l.maxLogBytes}
	for k, v := range l.base {
		child.base[k] = v
	}
	for k, v := range fields {
		child.base[k] = redactField(k, v)
	}
	return child
}

// logWithFields renders a structured JSON log with redaction
func (l *Logger) logWithFields(level LogLevel, msg string, fields map[string]interface{}) {
	if !l.IsEnabledFor(level) {
		return
	}
	timestamp := time.Now().UTC().Format(time.RFC3339Nano)
	entry := map[string]interface{}{
		"ts":    timestamp,
		"level": string(level),
		"msg":   redactValue(msg),
	}
	for k, v := range l.base {
		entry[k] = v
	}
	for k, v := range fields {
		entry[k] = redactField(k, v)
	}
	b, _ := json.Marshal(entry)
	out := string(b)
	if l.maxLogBytes > 0 && len(out) > l.maxLogBytes {
		suffix := "... [truncated]"
		cut := l.maxLogBytes - len(suffix)
		if cut < 0 {
			cut = 0
		}
		out = out[:cut] + suffix
	}
	fmt.Fprintln(os.Stderr, out)
}

// redactField redacts sensitive values by key
func redactField(key string, val interface{}) interface{} {
	lower := strings.ToLower(key)
	sensitive := []string{"password", "secret", "token", "api_key", "apikey", "authorization", "auth", "set-cookie", "cookie", "ssn", "email"}
	for _, s := range sensitive {
		if strings.Contains(lower, s) {
			return "[REDACTED]"
		}
	}
	switch v := val.(type) {
	case string:
		return redactValue(v)
	default:
		return val
	}
}

// redactValue redacts obvious secrets in free text
func redactValue(s string) string {
	if s == "" {
		return s
	}
	// Basic patterns
	patterns := []string{"Bearer ", "AWS ", "AKIA", "ya29.", "ghp_"}
	out := s
	for _, p := range patterns {
		if strings.Contains(out, p) {
			out = strings.ReplaceAll(out, p, p+"[REDACTED]")
		}
	}
	return out
}
