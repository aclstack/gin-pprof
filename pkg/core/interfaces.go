package core

import (
	"context"
	"time"
)

// HTTPContext 为不同框架抽象HTTP请求上下文
type HTTPContext interface {
	// GetPath 返回路由模板（例如："/users/:id"）
	GetPath() string
	// GetMethod 返回HTTP方法
	GetMethod() string
	// GetHeaders 返回请求头
	GetHeaders() map[string]string
	// SetContext 在上下文中设置键值对
	SetContext(key, value interface{})
	// GetContext 通过键从上下文获取值
	GetContext(key interface{}) interface{}
	// GetRequestPath 返回实际请求路径
	GetRequestPath() string
}

// ConfigProvider 抽象性能分析任务的配置源
type ConfigProvider interface {
	// GetTasks 返回当前性能分析任务
	GetTasks(ctx context.Context) ([]ProfilingTask, error)
	// Subscribe 订阅配置变更
	Subscribe(ctx context.Context, callback func([]ProfilingTask)) error
	// Close 关闭提供程序并释放资源
	Close() error
}

// Storage 抽象性能分析文件的存储后端
type Storage interface {
	// Save 将性能分析数据保存到存储
	Save(ctx context.Context, filename string, data []byte) error
	// List 列出匹配给定模式的文件
	List(ctx context.Context, pattern string) ([]string, error)
	// Delete 从存储中删除文件
	Delete(ctx context.Context, filename string) error
	// Clean 删除超过maxAge的文件
	Clean(ctx context.Context, maxAge time.Duration) error
}

// Logger 抽象日志功能
type Logger interface {
	// Info 记录信息消息
	Info(msg string, fields map[string]interface{})
	// Warn 记录警告消息
	Warn(msg string, fields map[string]interface{})
	// Error 记录错误消息
	Error(msg string, fields map[string]interface{})
	// Debug 记录调试消息
	Debug(msg string, fields map[string]interface{})
}

// PathMatcher 抽象路径匹配逻辑
type PathMatcher interface {
	// Match 检查实际路径是否匹配模板
	Match(template, actual string) bool
	// ExtractParams 使用模板从实际路径提取参数
	ExtractParams(template, actual string) map[string]string
}

// Profiler 抽象不同类型的性能分析
type Profiler interface {
	// StartProfiling 使用给定配置开始性能分析
	StartProfiling(ctx context.Context, task ProfilingTask) (ProfileSession, error)
	// GetProfileType 返回此分析器处理的性能分析类型
	GetProfileType() string
}

// ProfileSession 表示活跃的性能分析会话
type ProfileSession interface {
	// Stop 停止性能分析会话并返回结果
	Stop() ([]byte, error)
	// GetStartTime 返回会话开始时间
	GetStartTime() time.Time
	// IsRunning 如果会话仍处于活跃状态则返回true
	IsRunning() bool
}