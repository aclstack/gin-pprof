package core

import (
	"time"
)

// ProfilingTask 表示性能分析任务配置
type ProfilingTask struct {
	Path        string    `yaml:"path" json:"path"`                 // 路径
	Methods     []string  `yaml:"methods" json:"methods"`           // HTTP方法数组，支持多个方法或使用"*"表示常用方法
	ExpiresAt   time.Time `yaml:"expires_at" json:"expires_at"`     // 过期时间
	Duration    int       `yaml:"duration" json:"duration"`         // 最大分析持续时间(秒)
	SampleRate  int       `yaml:"sample_rate" json:"sample_rate"`   // 每N个请求进行采样
	ProfileType string    `yaml:"profile_type" json:"profile_type"` // cpu, heap, goroutine等
}

// ProfilingStats 表示性能分析统计信息
type ProfilingStats struct {
	TotalRequests  int64     `json:"total_requests"`  // 总请求数
	ProfiledCount  int64     `json:"profiled_count"`  // 已分析数量
	FailedCount    int64     `json:"failed_count"`    // 失败数量
	ActiveProfiles int64     `json:"active_profiles"` // 活跃分析数
	LastUpdate     time.Time `json:"last_update"`     // 最后更新时间
}

// ProfilingResult 表示性能分析会话的结果
type ProfilingResult struct {
	Path        string        `json:"path"`         // 路径
	StartTime   time.Time     `json:"start_time"`   // 开始时间
	Duration    time.Duration `json:"duration"`     // 持续时间
	Filename    string        `json:"filename"`     // 文件名
	FileSize    int64         `json:"file_size"`    // 文件大小
	ProfileType string        `json:"profile_type"` // 分析类型
	Success     bool          `json:"success"`      // 是否成功
	Error       string        `json:"error,omitempty"` // 错误信息
}