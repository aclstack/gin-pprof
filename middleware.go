package ginpprof

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/aclstack/gin-pprof/pkg/adapters/http"
)

// Middleware 为动态性能分析创建一个Gin中间件
func (p *Profiler) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 如果性能分析已禁用或管理器不可用则跳过
		if p.manager == nil || !p.manager.IsEnabled() {
			c.Next()
			return
		}

		// 创建HTTP上下文适配器
		httpCtx := http.NewGinContext(c)
		path := httpCtx.GetPath()
		method := httpCtx.GetMethod()

		// 检查是否应该对此请求进行性能分析
		task, shouldProfile := p.manager.ShouldProfile(path, method)
		if !shouldProfile {
			c.Next()
			return
		}

		// 开始性能分析
		ctx := context.Background()
		session, err := p.manager.StartProfiling(ctx, path, task)
		if err != nil {
			// 如果性能分析失败不要让请求失败
			p.logger.Error("Failed to start profiling", map[string]interface{}{
				"path":  path,
				"error": err.Error(),
			})
			c.Next()
			return
		}

		// 设置超时和执行
		done := make(chan bool, 1)
		timeout := time.Duration(task.Duration) * time.Second
		if timeout == 0 {
			timeout = 30 * time.Second
		}

		// 在协程中执行业务逻辑以控制超时
		go func() {
			defer func() {
				if r := recover(); r != nil {
					p.logger.Error("Request panic during profiling", map[string]interface{}{
						"path":  path,
						"panic": r,
					})
				}
				done <- true
			}()
			
			c.Next()
		}()

		// 等待完成或超时
		select {
		case <-done:
			// 请求正常完成
		case <-time.After(timeout):
			// 请求超时
			p.logger.Warn("Request timed out during profiling", map[string]interface{}{
				"path":    path,
				"timeout": timeout.Seconds(),
			})
		}

		// 停止性能分析并保存结果
		result, err := p.manager.StopProfiling(ctx, path, method, task, session)
		if err != nil {
			p.logger.Error("Failed to stop profiling", map[string]interface{}{
				"path":  path,
				"error": err.Error(),
			})
		} else if result != nil && result.Success {
			p.logger.Info("Profiling completed successfully", map[string]interface{}{
				"path":        path,
				"filename":    result.Filename,
				"duration_ms": result.Duration.Milliseconds(),
				"file_size":   result.FileSize,
				"type":        result.ProfileType,
			})
		}
	}
}