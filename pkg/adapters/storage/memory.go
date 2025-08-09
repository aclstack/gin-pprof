package storage

import (
	"context"
	"path/filepath"
	"sync"
	"time"

	"github.com/aclstack/gin-pprof/pkg/core"
)

// MemoryStorage implements Storage interface using in-memory storage (for testing)
type MemoryStorage struct {
	mu    sync.RWMutex
	files map[string]*memoryFile
	logger core.Logger
}

type memoryFile struct {
	data     []byte
	modTime  time.Time
}

// NewMemoryStorage creates a new MemoryStorage
func NewMemoryStorage(logger core.Logger) core.Storage {
	return &MemoryStorage{
		files:  make(map[string]*memoryFile),
		logger: logger,
	}
}

// Save saves profile data to memory
func (m *MemoryStorage) Save(ctx context.Context, filename string, data []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.files[filename] = &memoryFile{
		data:    make([]byte, len(data)),
		modTime: time.Now(),
	}
	copy(m.files[filename].data, data)

	m.logger.Info("Profile saved to memory", map[string]interface{}{
		"filename": filename,
		"size":     len(data),
	})

	return nil
}

// List lists files matching the given pattern
func (m *MemoryStorage) List(ctx context.Context, pattern string) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var matches []string
	for filename := range m.files {
		matched, err := filepath.Match(pattern, filename)
		if err != nil {
			return nil, err
		}
		if matched {
			matches = append(matches, filename)
		}
	}

	return matches, nil
}

// Delete deletes a file from memory
func (m *MemoryStorage) Delete(ctx context.Context, filename string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.files[filename]; !exists {
		return nil // File doesn't exist, consider it deleted
	}

	delete(m.files, filename)

	m.logger.Info("Profile deleted from memory", map[string]interface{}{
		"filename": filename,
	})

	return nil
}

// Clean removes files older than maxAge
func (m *MemoryStorage) Clean(ctx context.Context, maxAge time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	cleanedCount := 0

	for filename, file := range m.files {
		if now.Sub(file.modTime) > maxAge {
			delete(m.files, filename)
			cleanedCount++
		}
	}

	if cleanedCount > 0 {
		m.logger.Info("Memory profile cleanup completed", map[string]interface{}{
			"cleaned_files": cleanedCount,
			"max_age":       maxAge.String(),
		})
	}

	return nil
}