package logger

import "github.com/aclstack/gin-pprof/pkg/core"

// NoopLogger is a logger that does nothing (for testing or when logging is disabled)
type NoopLogger struct{}

// NewNoopLogger creates a new NoopLogger
func NewNoopLogger() core.Logger {
	return &NoopLogger{}
}

// Info does nothing
func (l *NoopLogger) Info(msg string, fields map[string]interface{}) {}

// Warn does nothing
func (l *NoopLogger) Warn(msg string, fields map[string]interface{}) {}

// Error does nothing
func (l *NoopLogger) Error(msg string, fields map[string]interface{}) {}

// Debug does nothing
func (l *NoopLogger) Debug(msg string, fields map[string]interface{}) {}