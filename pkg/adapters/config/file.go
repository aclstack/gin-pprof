package config

import (
	"context"
	"os"
	"time"

	"gopkg.in/yaml.v3"
	"github.com/aclstack/gin-pprof/pkg/core"
)

// FileConfig implements ConfigProvider interface using local YAML file
type FileConfig struct {
	filePath string
	logger   core.Logger
}

// FileConfigFormat represents the structure of the config file
type FileConfigFormat struct {
	Profiles []core.ProfilingTask `yaml:"profiles"`
}

// NewFileConfig creates a new FileConfig
func NewFileConfig(filePath string, logger core.Logger) core.ConfigProvider {
	f := &FileConfig{
		filePath: filePath,
		logger:   logger,
	}
	
	// Check if file exists, create example config if not
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		if err := f.createExampleConfig(); err != nil {
			logger.Warn("Failed to create example config file", map[string]interface{}{
				"file":  filePath,
				"error": err.Error(),
				"hint":  "Please create the config file manually or check file permissions",
			})
		} else {
			logger.Info("Created example config file", map[string]interface{}{
				"file": filePath,
				"hint": "Edit the config file to enable profiling for specific endpoints",
				"docs": "https://github.com/aclstack/gin-pprof/examples",
			})
		}
	}
	
	return f
}

// GetTasks returns current profiling tasks from file
func (f *FileConfig) GetTasks(ctx context.Context) ([]core.ProfilingTask, error) {
	data, err := os.ReadFile(f.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			f.logger.Info("Config file not found, profiling is disabled", map[string]interface{}{
				"file": f.filePath,
				"hint": "Create and configure the file to enable profiling",
				"example": "https://github.com/aclstack/gin-pprof/blob/main/examples/basic/profiling.yaml",
			})
			return []core.ProfilingTask{}, nil
		}
		f.logger.Error("Failed to read config file", map[string]interface{}{
			"file":  f.filePath,
			"error": err.Error(),
		})
		return nil, err
	}

	var config FileConfigFormat
	if err := yaml.Unmarshal(data, &config); err != nil {
		f.logger.Error("Failed to parse config file", map[string]interface{}{
			"file":  f.filePath,
			"error": err.Error(),
		})
		return nil, err
	}

	// Filter out expired tasks
	var validTasks []core.ProfilingTask
	var expiredCount int
	now := time.Now()
	for _, task := range config.Profiles {
		if now.Before(task.ExpiresAt) {
			// Set default values if not specified
			if task.Duration == 0 {
				task.Duration = 30
			}
			if task.SampleRate == 0 {
				task.SampleRate = 1
			}
			if task.ProfileType == "" {
				task.ProfileType = "cpu"
			}
			// 设置默认方法 - 如果没有指定方法，默认为GET
			if len(task.Methods) == 0 {
				task.Methods = []string{"GET"}
			}
			validTasks = append(validTasks, task)
		} else {
			expiredCount++
			f.logger.Warn("Task expired", map[string]interface{}{
				"path":       task.Path,
				"expires_at": task.ExpiresAt.Format(time.RFC3339),
			})
		}
	}

	// Add warning if all tasks are expired
	if len(config.Profiles) > 0 && len(validTasks) == 0 {
		f.logger.Error("All profiling tasks have expired!", map[string]interface{}{
			"total_tasks":   len(config.Profiles),
			"expired_tasks": expiredCount,
			"file":          f.filePath,
			"action":        "Please update the expires_at timestamps in your config file",
			"hint":          "All configured profiling tasks are past their expiration dates",
		})
	}

	f.logger.Info("Config loaded", map[string]interface{}{
		"file":         f.filePath,
		"total_tasks":  len(config.Profiles),
		"valid_tasks":  len(validTasks),
		"expired_tasks": expiredCount,
	})

	return validTasks, nil
}

// Subscribe is not implemented for file config (file watching could be added later)
func (f *FileConfig) Subscribe(ctx context.Context, callback func([]core.ProfilingTask)) error {
	// For file config, we don't support real-time subscription
	// This could be enhanced with file watching in the future
	f.logger.Info("File config subscription not implemented", map[string]interface{}{
		"file": f.filePath,
	})
	return nil
}

// Close closes the file config provider
func (f *FileConfig) Close() error {
	f.logger.Info("File config provider closed", map[string]interface{}{
		"file": f.filePath,
	})
	return nil
}

// createExampleConfig creates an example configuration file
func (f *FileConfig) createExampleConfig() error {
	exampleConfig := `# gin-pprof 配置文件
# 这是一个自动生成的示例配置文件
# 编辑此文件以为特定端点启用性能分析

profiles:
  # 示例：用户详情端点的CPU分析（单个方法）
  # - path: "/api/users/:id"
  #   methods: ["GET"]       # 单个HTTP方法
  #   expires_at: "2025-12-31T23:59:59Z"
  #   duration: 10          # 分析10秒
  #   sample_rate: 1        # 每个请求都分析
  #   profile_type: "cpu"   # CPU分析
  
  # 示例：多个方法的内存分析
  # - path: "/api/data/heavy"
  #   methods: ["POST", "PUT"]  # 多个HTTP方法数组
  #   expires_at: "2025-12-31T23:59:59Z"
  #   duration: 15          # 分析15秒
  #   sample_rate: 5        # 每5个请求分析1次
  #   profile_type: "heap"  # 内存分析
  
  # 示例：所有常用方法的协程分析
  # - path: "/api/concurrent/:operation"
  #   methods: ["*"]         # 通配符：匹配GET, POST, PUT, DELETE
  #   expires_at: "2025-12-31T23:59:59Z" 
  #   duration: 20
  #   sample_rate: 2
  #   profile_type: "goroutine"
  
  # 示例：默认行为（仅GET）
  # - path: "/api/health"
  #   # methods未指定 - 默认为GET
  #   expires_at: "2025-12-31T23:59:59Z"
  #   duration: 5
  #   profile_type: "cpu"

# 启用性能分析步骤：
# 1. 取消上述一个或多个配置的注释
# 2. 修改'path'以匹配你的API端点
# 3. 设置HTTP方法：
#    - methods: ["GET"] （单个方法）
#    - methods: ["POST", "PUT"] （多个方法）
#    - methods: ["*"] （常用方法：GET, POST, PUT, DELETE）
#    - 留空则默认为GET
# 4. 设置合适的'expires_at'时间
# 5. 根据需要调整'duration'和'sample_rate'
# 6. 选择'profile_type'：cpu, heap或goroutine
#
# 更多示例和文档：
# https://github.com/aclstack/gin-pprof/blob/main/README.md
`
	
	return os.WriteFile(f.filePath, []byte(exampleConfig), 0644)
}