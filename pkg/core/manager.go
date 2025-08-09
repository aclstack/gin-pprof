package core

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Manager 管理性能分析会话和任务
type Manager struct {
	mu            sync.RWMutex
	tasks         map[string]ProfilingTask
	stats         ProfilingStats
	options       Options
	limiter       chan struct{}
	requestCount  map[string]int64
	configProvider ConfigProvider
	storage       Storage
	logger        Logger
	pathMatcher   PathMatcher
	profilers     map[string]Profiler
	cleanupStop   chan struct{}
	cleanupDone   chan struct{}
}

// NewManager 创建新的性能分析管理器
func NewManager(opts Options, configProvider ConfigProvider, storage Storage, logger Logger, pathMatcher PathMatcher) *Manager {
	m := &Manager{
		tasks:          make(map[string]ProfilingTask),
		options:        opts,
		limiter:        make(chan struct{}, opts.MaxConcurrent),
		requestCount:   make(map[string]int64),
		configProvider: configProvider,
		storage:        storage,
		logger:         logger,
		pathMatcher:    pathMatcher,
		profilers:      make(map[string]Profiler),
		cleanupStop:    make(chan struct{}),
		cleanupDone:    make(chan struct{}),
		stats: ProfilingStats{
			LastUpdate: time.Now(),
		},
	}

	// 注册默认分析器
	m.RegisterProfiler(NewCPUProfiler())
	m.RegisterProfiler(NewHeapProfiler())
	m.RegisterProfiler(NewGoroutineProfiler())

	// 启动后台任务
	go m.startConfigSync()
	go m.startCleanup()

	logger.Info("Profiling manager initialized", map[string]interface{}{
		"max_concurrent": opts.MaxConcurrent,
		"enabled":        opts.Enabled,
		"profile_dir":    opts.ProfileDir,
	})

	return m
}

// RegisterProfiler 为特定类型注册分析器
func (m *Manager) RegisterProfiler(profiler Profiler) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.profilers[profiler.GetProfileType()] = profiler
	
	m.logger.Info("Profiler registered", map[string]interface{}{
		"type": profiler.GetProfileType(),
	})
}

// ShouldProfile 检查是否应该对请求进行性能分析
func (m *Manager) ShouldProfile(path, method string) (ProfilingTask, bool) {
	if !m.options.Enabled {
		return ProfilingTask{}, false
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	// 查找匹配的任务
	var matchedTask ProfilingTask
	var found bool

	for _, task := range m.tasks {
		if m.pathMatcher.Match(task.Path, path) && task.ShouldMatchMethod(method) {
			matchedTask = task
			found = true
			break
		}
	}

	if !found {
		return ProfilingTask{}, false
	}

	// 检查是否已过期
	if time.Now().After(matchedTask.ExpiresAt) {
		return ProfilingTask{}, false
	}

	// 采样率控制
	if matchedTask.SampleRate > 1 {
		m.requestCount[matchedTask.Path]++
		count := m.requestCount[matchedTask.Path]

		if count%int64(matchedTask.SampleRate) != 0 {
			return ProfilingTask{}, false
		}
	}

	// 检查并发限制
	select {
	case m.limiter <- struct{}{}:
		return matchedTask, true
	default:
		m.stats.FailedCount++
		m.logger.Warn("Concurrent limit exceeded", map[string]interface{}{
			"path": path,
			"limit": m.options.MaxConcurrent,
		})
		return ProfilingTask{}, false
	}
}

// StartProfiling 开始性能分析会话
func (m *Manager) StartProfiling(ctx context.Context, path string, task ProfilingTask) (ProfileSession, error) {
	// 获取适当的分析器
	profiler, exists := m.profilers[task.ProfileType]
	if !exists {
		m.releaseLimiter()
		err := fmt.Errorf("profiler type %s not found", task.ProfileType)
		m.logger.Error("Profiler not found", map[string]interface{}{
			"type":  task.ProfileType,
			"path":  path,
			"error": err.Error(),
		})
		return nil, err
	}

	// 开始性能分析会话
	session, err := profiler.StartProfiling(ctx, task)
	if err != nil {
		m.releaseLimiter()
		m.mu.Lock()
		m.stats.FailedCount++
		m.mu.Unlock()
		
		m.logger.Error("Failed to start profiling", map[string]interface{}{
			"path":  path,
			"type":  task.ProfileType,
			"error": err.Error(),
		})
		return nil, err
	}

	m.mu.Lock()
	m.stats.ActiveProfiles++
	m.stats.ProfiledCount++
	m.mu.Unlock()

	m.logger.Info("Profiling started", map[string]interface{}{
		"path":     path,
		"type":     task.ProfileType,
		"duration": task.Duration,
	})

	return session, nil
}

// StopProfiling 停止性能分析会话并保存结果
func (m *Manager) StopProfiling(ctx context.Context, path, method string, task ProfilingTask, session ProfileSession) (*ProfilingResult, error) {
	defer m.releaseLimiter()

	startTime := session.GetStartTime()
	data, err := session.Stop()
	
	m.mu.Lock()
	m.stats.ActiveProfiles--
	m.mu.Unlock()

	result := &ProfilingResult{
		Path:        path,
		StartTime:   startTime,
		Duration:    time.Since(startTime),
		ProfileType: task.ProfileType,
		Success:     err == nil,
	}

	if err != nil {
		result.Error = err.Error()
		m.mu.Lock()
		m.stats.FailedCount++
		m.mu.Unlock()
		
		m.logger.Error("Failed to stop profiling", map[string]interface{}{
			"path":  path,
			"error": err.Error(),
		})
		return result, err
	}

	if len(data) == 0 {
		m.logger.Warn("Empty profiling data", map[string]interface{}{
			"path": path,
			"type": task.ProfileType,
		})
		return result, nil
	}

	// 生成文件名
	filename := m.generateFilename(path, method, task.ProfileType)
	result.Filename = filename
	result.FileSize = int64(len(data))

	// 保存到存储
	if err := m.storage.Save(ctx, filename, data); err != nil {
		result.Success = false
		result.Error = err.Error()
		m.mu.Lock()
		m.stats.FailedCount++
		m.mu.Unlock()
		
		m.logger.Error("Failed to save profile", map[string]interface{}{
			"path":     path,
			"filename": filename,
			"error":    err.Error(),
		})
		return result, err
	}

	m.logger.Info("Profiling completed", map[string]interface{}{
		"path":        path,
		"filename":    filename,
		"duration_ms": result.Duration.Milliseconds(),
		"file_size":   result.FileSize,
		"type":        task.ProfileType,
	})

	return result, nil
}

// GetStats 返回当前统计信息
func (m *Manager) GetStats() ProfilingStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 计算总请求数
	var totalRequests int64
	for _, count := range m.requestCount {
		totalRequests += count
	}
	
	stats := m.stats
	stats.TotalRequests = totalRequests
	return stats
}

// GetTasks 返回当前任务
func (m *Manager) GetTasks() map[string]ProfilingTask {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tasks := make(map[string]ProfilingTask)
	for k, v := range m.tasks {
		tasks[k] = v
	}
	return tasks
}

// IsEnabled 返回性能分析是否开启
func (m *Manager) IsEnabled() bool {
	return m.options.Enabled
}

// Close 关闭管理器并释放资源
func (m *Manager) Close() error {
	close(m.cleanupStop)
	<-m.cleanupDone

	if m.configProvider != nil {
		m.configProvider.Close()
	}

	m.logger.Info("Profiling manager closed", nil)
	return nil
}

// startConfigSync 启动配置同步
func (m *Manager) startConfigSync() {
	// 初始加载
	ctx := context.Background()
	if tasks, err := m.configProvider.GetTasks(ctx); err == nil {
		m.updateTasks(tasks)
	} else {
		m.logger.Error("Failed to load initial config", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// 订阅变更
	m.configProvider.Subscribe(ctx, func(tasks []ProfilingTask) {
		m.updateTasks(tasks)
	})
}

// startCleanup 启动清理例程
func (m *Manager) startCleanup() {
	defer close(m.cleanupDone)
	
	ticker := time.NewTicker(m.options.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ctx := context.Background()
			if err := m.storage.Clean(ctx, m.options.MaxFileAge); err != nil {
				m.logger.Error("Storage cleanup failed", map[string]interface{}{
					"error": err.Error(),
				})
			}
			m.cleanExpiredTasks()
		case <-m.cleanupStop:
			return
		}
	}
}

// updateTasks 安全地更新任务列表
func (m *Manager) updateTasks(newTasks []ProfilingTask) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 转换为map以便查找
	taskMap := make(map[string]ProfilingTask)
	for _, task := range newTasks {
		taskMap[task.Path] = task
	}

	m.tasks = taskMap
	m.stats.LastUpdate = time.Now()

	// 清理已删除任务的请求计数
	for path := range m.requestCount {
		if _, exists := taskMap[path]; !exists {
			delete(m.requestCount, path)
		}
	}

	m.logger.Info("Tasks updated", map[string]interface{}{
		"task_count": len(newTasks),
	})
}

// cleanExpiredTasks 删除已过期任务
func (m *Manager) cleanExpiredTasks() {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	for path, task := range m.tasks {
		if now.After(task.ExpiresAt) {
			delete(m.tasks, path)
			delete(m.requestCount, path)
		}
	}
}

// releaseLimiter 从并发限制器释放一个槽位
func (m *Manager) releaseLimiter() {
	select {
	case <-m.limiter:
	default:
	}
}

// generateFilename 为性能分析文件生成文件名
func (m *Manager) generateFilename(path, method, profileType string) string {
	sanitized := sanitizePath(path)
	timestamp := time.Now().Format("20060102_150405")
	nanos := time.Now().UnixNano() % 1000000
	
	return fmt.Sprintf("%s/profile_%s_%s_%s_%d.pprof", profileType, sanitized, method, timestamp, nanos)
}

// sanitizePath 清理路径以便在文件名中使用
func sanitizePath(path string) string {
	// 替换有问题的字符
	replacements := map[rune]rune{
		'/': '_',
		':': '_',
		'*': '_',
		'?': '_',
		'<': '_',
		'>': '_',
		'|': '_',
		'"': '_',
		'\\': '_',
	}

	result := make([]rune, 0, len(path))
	for _, r := range path {
		if replacement, exists := replacements[r]; exists {
			result = append(result, replacement)
		} else {
			result = append(result, r)
		}
	}

	return string(result)
}