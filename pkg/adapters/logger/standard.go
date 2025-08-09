package logger

import (
	"encoding/json"
	"log"
	"time"

	"github.com/aclstack/gin-pprof/pkg/core"
)

// StandardLogger implements Logger interface using Go's standard log package
type StandardLogger struct {
	component string
}

// NewStandardLogger creates a new StandardLogger
func NewStandardLogger(component string) core.Logger {
	return &StandardLogger{
		component: component,
	}
}

// Info logs an info message
func (l *StandardLogger) Info(msg string, fields map[string]interface{}) {
	l.logJSON("info", msg, fields)
}

// Warn logs a warning message
func (l *StandardLogger) Warn(msg string, fields map[string]interface{}) {
	l.logJSON("warn", msg, fields)
}

// Error logs an error message
func (l *StandardLogger) Error(msg string, fields map[string]interface{}) {
	l.logJSON("error", msg, fields)
}

// Debug logs a debug message
func (l *StandardLogger) Debug(msg string, fields map[string]interface{}) {
	l.logJSON("debug", msg, fields)
}

// logJSON outputs JSON formatted log
func (l *StandardLogger) logJSON(level, msg string, fields map[string]interface{}) {
	logData := map[string]interface{}{
		"level":     level,
		"component": l.component,
		"message":   msg,
		"timestamp": time.Now().Format(time.RFC3339),
	}

	// Merge additional fields
	for k, v := range fields {
		logData[k] = v
	}

	jsonBytes, _ := json.Marshal(logData)
	log.Printf(string(jsonBytes))
}