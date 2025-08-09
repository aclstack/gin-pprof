package storage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aclstack/gin-pprof/pkg/core"
)

// FileStorage implements Storage interface using local file system
type FileStorage struct {
	baseDir string
	logger  core.Logger
}

// NewFileStorage creates a new FileStorage
func NewFileStorage(baseDir string, logger core.Logger) (core.Storage, error) {
	// Create directory if it doesn't exist
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, err
	}

	return &FileStorage{
		baseDir: baseDir,
		logger:  logger,
	}, nil
}

// Save saves profile data to a file
func (f *FileStorage) Save(ctx context.Context, filename string, data []byte) error {
	filePath := filepath.Join(f.baseDir, filename)

	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		// 如果连创建目录都失败了，就没必要继续了
		f.logger.Error("Failed to create profile directory", map[string]interface{}{
			"directory": filepath.Dir(filePath),
			"error":     err.Error(),
		})
		return err
	}

	file, err := os.Create(filePath)
	if err != nil {
		f.logger.Error("Failed to create profile file", map[string]interface{}{
			"filename": filename,
			"error":    err.Error(),
		})
		return err
	}
	defer file.Close()

	_, err = file.Write(data)
	if err != nil {
		f.logger.Error("Failed to write profile data", map[string]interface{}{
			"filename": filename,
			"error":    err.Error(),
		})
		return err
	}

	f.logger.Info("Profile saved", map[string]interface{}{
		"filename": filename,
		"size":     len(data),
		"path":     filePath,
	})

	return nil
}

// List lists files matching the given pattern
func (f *FileStorage) List(ctx context.Context, pattern string) ([]string, error) {
	fullPattern := filepath.Join(f.baseDir, pattern)
	matches, err := filepath.Glob(fullPattern)
	if err != nil {
		return nil, err
	}

	// Remove base directory prefix
	var result []string
	for _, match := range matches {
		rel, err := filepath.Rel(f.baseDir, match)
		if err != nil {
			continue
		}
		result = append(result, rel)
	}

	return result, nil
}

// Delete deletes a file from storage
func (f *FileStorage) Delete(ctx context.Context, filename string) error {
	filePath := filepath.Join(f.baseDir, filename)

	err := os.Remove(filePath)
	if err != nil {
		f.logger.Error("Failed to delete profile file", map[string]interface{}{
			"filename": filename,
			"error":    err.Error(),
		})
		return err
	}

	f.logger.Info("Profile deleted", map[string]interface{}{
		"filename": filename,
	})

	return nil
}

// Clean removes files older than maxAge
func (f *FileStorage) Clean(ctx context.Context, maxAge time.Duration) error {
	files, err := f.List(ctx, "*.pprof")
	if err != nil {
		return err
	}

	now := time.Now()
	cleanedCount := 0

	for _, filename := range files {
		filePath := filepath.Join(f.baseDir, filename)

		stat, err := os.Stat(filePath)
		if err != nil {
			continue
		}

		if now.Sub(stat.ModTime()) > maxAge {
			if err := f.Delete(ctx, filename); err == nil {
				cleanedCount++
			}
		}
	}

	if cleanedCount > 0 {
		f.logger.Info("Profile cleanup completed", map[string]interface{}{
			"cleaned_files": cleanedCount,
			"max_age":       maxAge.String(),
		})
	}

	return nil
}

// sanitizePath cleans path for use in filenames
func sanitizePath(path string) string {
	sanitized := strings.ReplaceAll(path, "/", "_")
	sanitized = strings.ReplaceAll(sanitized, ":", "_")
	sanitized = strings.ReplaceAll(sanitized, "*", "_")
	sanitized = strings.ReplaceAll(sanitized, "?", "_")
	sanitized = strings.ReplaceAll(sanitized, "<", "_")
	sanitized = strings.ReplaceAll(sanitized, ">", "_")
	sanitized = strings.ReplaceAll(sanitized, "|", "_")
	return sanitized
}

// GenerateFilename generates a filename for a profiling task
func GenerateFilename(path, profileType string) string {
	sanitized := sanitizePath(path)
	timestamp := time.Now().Format("20060102_150405")
	nanos := time.Now().UnixNano() % 1000000

	return filepath.Join(profileType,
		fmt.Sprintf("profile_%s_%s_%d.pprof", sanitized, timestamp, nanos))
}
