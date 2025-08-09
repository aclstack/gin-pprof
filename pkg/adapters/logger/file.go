package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/aclstack/gin-pprof/pkg/core"
)

// FileLogger implements core.Logger interface for file-based JSON logging
type FileLogger struct {
	filePath string
	file     *os.File
	mu       sync.Mutex
}

// LogEntry represents a log entry in JSON format
type LogEntry struct {
	Timestamp string                 `json:"timestamp"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
}

// NewFileLogger creates a new file logger that writes JSON logs to the specified path
func NewFileLogger(filePath string) (core.Logger, error) {
	// Ensure directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %v", err)
	}

	// Open file for writing (append mode)
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %v", err)
	}

	return &FileLogger{
		filePath: filePath,
		file:     file,
	}, nil
}

// Debug logs debug level message
func (f *FileLogger) Debug(msg string, fields map[string]interface{}) {
	f.writeLog("DEBUG", msg, fields)
}

// Info logs info level message
func (f *FileLogger) Info(msg string, fields map[string]interface{}) {
	f.writeLog("INFO", msg, fields)
}

// Warn logs warning level message
func (f *FileLogger) Warn(msg string, fields map[string]interface{}) {
	f.writeLog("WARN", msg, fields)
}

// Error logs error level message
func (f *FileLogger) Error(msg string, fields map[string]interface{}) {
	f.writeLog("ERROR", msg, fields)
}

// writeLog writes a log entry to the file in JSON format
func (f *FileLogger) writeLog(level, message string, fields map[string]interface{}) {
	f.mu.Lock()
	defer f.mu.Unlock()

	entry := LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     level,
		Message:   message,
		Fields:    fields,
	}

	data, err := json.Marshal(entry)
	if err != nil {
		// Fallback to simple text log if JSON marshaling fails
		fmt.Fprintf(f.file, "[%s] %s %s: %+v\n", 
			entry.Timestamp, level, message, fields)
		return
	}

	fmt.Fprintf(f.file, "%s\n", string(data))
}

// Close closes the log file
func (f *FileLogger) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	
	if f.file != nil {
		return f.file.Close()
	}
	return nil
}