package ginpprof

import (
	"net/http"
	"os"
	"time"

	"github.com/aclstack/gin-pprof/pkg/adapters/config"
	ginpprofhttp "github.com/aclstack/gin-pprof/pkg/adapters/http"
	"github.com/aclstack/gin-pprof/pkg/adapters/logger"
	"github.com/aclstack/gin-pprof/pkg/adapters/storage"
	"github.com/aclstack/gin-pprof/pkg/core"
	"github.com/gin-gonic/gin"
)

// Profiler is the main profiler instance
type Profiler struct {
	manager *core.Manager
	logger  core.Logger
	options core.Options
}

// Builder provides a fluent interface for creating a Profiler
type Builder struct {
	options        core.Options
	configProvider core.ConfigProvider
	storage        core.Storage
	logger         core.Logger
	pathMatcher    core.PathMatcher
}

// New creates a new profiler builder
func New() *Builder {
	return &Builder{
		options: core.DefaultOptions(),
	}
}

// WithOptions sets custom options
func (b *Builder) WithOptions(opts core.Options) *Builder {
	b.options = opts
	return b
}

// WithFileConfig configures file-based configuration
func (b *Builder) WithFileConfig(filePath string) *Builder {
	b.configProvider = config.NewFileConfig(filePath, b.getOrCreateFileLogger())
	return b
}

// WithNacosConfig configures Nacos-based configuration
func (b *Builder) WithNacosConfig(opts config.NacosOptions) *Builder {
	fileLogger := b.getOrCreateFileLogger()

	nacosConfig, err := config.NewNacosConfig(opts, fileLogger)
	if err != nil {
		fileLogger.Error("Failed to create Nacos config", map[string]interface{}{
			"error": err.Error(),
		})
		return b
	}

	b.configProvider = nacosConfig
	return b
}

// WithFileStorage configures file-based storage
func (b *Builder) WithFileStorage(baseDir string) *Builder {
	fileLogger := b.getOrCreateFileLogger()

	// 自动创建目录（忽略已存在的情况）
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		fileLogger.Error("Failed to create baseDir", map[string]interface{}{
			"error": err.Error(),
		})
		return b
	}

	fileStorage, err := storage.NewFileStorage(baseDir, fileLogger)
	if err != nil {
		fileLogger.Error("Failed to create file storage", map[string]interface{}{
			"error": err.Error(),
		})
		return b
	}

	b.storage = fileStorage
	return b
}

// WithMemoryStorage configures memory-based storage (for testing)
func (b *Builder) WithMemoryStorage() *Builder {
	fileLogger := b.getOrCreateFileLogger()

	b.storage = storage.NewMemoryStorage(fileLogger)
	return b
}

// WithLogger sets a custom logger
func (b *Builder) WithLogger(logger core.Logger) *Builder {
	b.logger = logger
	return b
}

// WithNoLogger disables logging
func (b *Builder) WithNoLogger() *Builder {
	b.logger = logger.NewNoopLogger()
	return b
}

// Build creates the profiler instance
func (b *Builder) Build() *Profiler {
	// Set defaults if not provided
	if b.logger == nil {
		b.logger = b.getOrCreateFileLogger()
	}

	if b.configProvider == nil {
		b.logger.Warn("No config provider specified, using file config with default path", nil)
		b.configProvider = config.NewFileConfig("gin-pprof.yaml", b.logger)
	}

	if b.storage == nil {
		b.logger.Info("No storage specified, using file storage with default path", nil)
		fileStorage, err := storage.NewFileStorage(b.options.ProfileDir, b.logger)
		if err != nil {
			b.logger.Error("Failed to create default file storage", map[string]interface{}{
				"error": err.Error(),
			})
			b.storage = storage.NewMemoryStorage(b.logger)
		} else {
			b.storage = fileStorage
		}
	}

	if b.pathMatcher == nil {
		b.pathMatcher = ginpprofhttp.NewGinPathMatcher()
	}

	// Create manager
	manager := core.NewManager(
		b.options,
		b.configProvider,
		b.storage,
		b.logger,
		b.pathMatcher,
	)

	return &Profiler{
		manager: manager,
		logger:  b.logger,
		options: b.options,
	}
}

// getOrCreateFileLogger creates or returns the file logger
func (b *Builder) getOrCreateFileLogger() core.Logger {
	if b.logger != nil {
		return b.logger
	}

	// Try to create file logger, fallback to standard logger if failed
	fileLogger, err := logger.NewFileLogger("./log/gin-pprof.log")
	if err != nil {
		// Fallback to standard logger if file logger creation fails
		b.logger = logger.NewStandardLogger("gin-pprof")
		b.logger.Warn("Failed to create file logger, using standard logger", map[string]interface{}{
			"error": err.Error(),
		})
		return b.logger
	}

	b.logger = fileLogger
	return b.logger
}

// GetStats returns profiling statistics
func (p *Profiler) GetStats() core.ProfilingStats {
	if p.manager == nil {
		return core.ProfilingStats{}
	}
	return p.manager.GetStats()
}

// GetTasks returns current profiling tasks
func (p *Profiler) GetTasks() map[string]core.ProfilingTask {
	if p.manager == nil {
		return make(map[string]core.ProfilingTask)
	}
	return p.manager.GetTasks()
}

// IsEnabled returns whether profiling is enabled
func (p *Profiler) IsEnabled() bool {
	if p.manager == nil {
		return false
	}
	return p.manager.IsEnabled()
}

// Close closes the profiler and releases resources
func (p *Profiler) Close() error {
	if p.manager != nil {
		return p.manager.Close()
	}
	return nil
}

// StatusHandler returns a Gin handler for profiling status
func (p *Profiler) StatusHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		if p.manager == nil {
			c.JSON(http.StatusOK, gin.H{
				"enabled": false,
				"message": "Profiling manager not initialized",
			})
			return
		}

		if !p.manager.IsEnabled() {
			c.JSON(http.StatusOK, gin.H{
				"enabled": false,
				"message": "Profiling disabled",
			})
			return
		}

		stats := p.manager.GetStats()
		tasks := p.manager.GetTasks()

		// Calculate active tasks count
		now := time.Now()
		activeTasks := 0
		for _, task := range tasks {
			if now.Before(task.ExpiresAt) {
				activeTasks++
			}
		}

		response := gin.H{
			"enabled":      true,
			"stats":        stats,
			"active_tasks": activeTasks,
			"total_tasks":  len(tasks),
			"profile_dir":  p.options.ProfileDir,
		}

		// Include task details if requested
		if c.Query("detail") == "true" {
			response["tasks"] = tasks
		}

		c.JSON(http.StatusOK, response)
	}
}

// TasksHandler returns a Gin handler for profiling tasks
func (p *Profiler) TasksHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		if p.manager == nil || !p.manager.IsEnabled() {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Profiling not enabled",
			})
			return
		}

		tasks := p.manager.GetTasks()
		now := time.Now()

		// Categorize tasks
		activeTasks := make([]core.ProfilingTask, 0)
		expiredTasks := make([]core.ProfilingTask, 0)

		for _, task := range tasks {
			if now.Before(task.ExpiresAt) {
				activeTasks = append(activeTasks, task)
			} else {
				expiredTasks = append(expiredTasks, task)
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"active_tasks":  activeTasks,
			"expired_tasks": expiredTasks,
			"total":         len(tasks),
		})
	}
}

// StatsHandler returns a Gin handler for profiling statistics
func (p *Profiler) StatsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		if p.manager == nil || !p.manager.IsEnabled() {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Profiling not enabled",
			})
			return
		}

		stats := p.manager.GetStats()

		// Calculate success rate
		successRate := float64(0)
		if stats.TotalRequests > 0 {
			successRate = float64(stats.ProfiledCount) / float64(stats.TotalRequests) * 100
		}

		response := map[string]interface{}{
			"total_requests":  stats.TotalRequests,
			"profiled_count":  stats.ProfiledCount,
			"failed_count":    stats.FailedCount,
			"active_profiles": stats.ActiveProfiles,
			"success_rate":    successRate,
			"last_update":     stats.LastUpdate.Format(time.RFC3339),
		}

		c.JSON(http.StatusOK, response)
	}
}
