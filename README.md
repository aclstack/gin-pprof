# gin-pprof

[![Go](https://github.com/aclstack/gin-pprof/workflows/Go/badge.svg)](https://github.com/aclstack/gin-pprof/actions)
[![GoDoc](https://godoc.org/github.com/aclstack/gin-pprof?status.svg)](https://godoc.org/github.com/aclstack/gin-pprof)
[![Go Report Card](https://goreportcard.com/badge/github.com/aclstack/gin-pprof)](https://goreportcard.com/report/github.com/aclstack/gin-pprof)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

一个为 Gin Web 框架设计的动态、配置驱动的性能分析中间件，支持**生产环境安全**的性能分析，无需重启服务。

[English](README_EN) | 中文

## ✨ 特性

- 🔄 **动态配置**：无需重启服务即可启用/禁用特定接口的性能分析
- 🎯 **路由感知**：智能匹配参数化路由（如 `/users/:id`）
- 🏗️ **多种分析类型**：支持 CPU、内存（堆）和 Goroutine 分析
- 📊 **智能采样**：可配置采样率，最小化性能影响
- 🚦 **并发控制**：内置限制防止资源耗尽
- 🔌 **插件化架构**：支持多种配置源和存储后端
- ☁️ **云原生**：Nacos 配置中心集成
- 📈 **监控**：内置统计和健康检查端点
- 🛡️ **生产安全**：非阻塞设计，自动清理

## 🚀 快速开始

### 安装

```bash
go get github.com/aclstack/gin-pprof
```

### 基本使用

```go
package main

import (
    "github.com/gin-gonic/gin"
    ginpprof "github.com/aclstack/gin-pprof"
)

func main() {
    r := gin.Default()

    // 创建基于文件配置的性能分析器
    // 如果 profiling.yaml 不存在，将自动创建示例配置文件
    profiler := ginpprof.New().
        WithFileConfig("./profiling.yaml").
        WithFileStorage("./profiles").
        Build()
    defer profiler.Close()

    // 添加性能分析中间件
    r.Use(profiler.Middleware())

    // 添加监控端点
    debug := r.Group("/debug/profiling")
    {
        debug.GET("/status", profiler.StatusHandler())
        debug.GET("/tasks", profiler.TasksHandler())
        debug.GET("/stats", profiler.StatsHandler())
    }

    // 你的 API 端点
    r.GET("/api/users/:id", getUser)
    
    r.Run(":8080")
}
```

> **📝 说明**：如果 `profiling.yaml` 不存在，gin-pprof 会自动创建一个包含详细注释的示例配置文件。只需取消注释并修改示例配置即可为你的接口启用性能分析。

### 配置文件 (profiling.yaml)

```yaml
profiles:
  # 单个 HTTP 方法
  - path: "/api/users/:id"
    methods: ["GET"]      # 单个方法
    expires_at: "2025-12-31T23:59:59Z"
    duration: 10          # 分析 10 秒
    profile_type: "cpu"   # CPU 分析
  
  # 多个 HTTP 方法
  - path: "/api/heavy"
    methods: ["POST", "PUT"]  # 多个方法数组
    expires_at: "2025-12-31T23:59:59Z"
    duration: 15
    profile_type: "heap"  # 内存分析
    
  # 通配符匹配常见方法
  - path: "/api/data/*"
    methods: ["*"]        # 匹配 GET, POST, PUT, DELETE
    expires_at: "2025-12-31T23:59:59Z"
    duration: 20
    profile_type: "goroutine"
```

## 🏗️ 架构

```
┌─────────────────────────────────────────────────────────┐
│                   应用层                                │
├─────────────────────────────────────────────────────────┤
│                 适配器层                                │
│  Gin│Echo│Fiber  Nacos│文件│Etcd  本地│S3│OSS           │
├─────────────────────────────────────────────────────────┤
│                 接口层                                  │
│  HTTPContext│ConfigProvider│Storage│Logger              │
├─────────────────────────────────────────────────────────┤
│                  核心层                                 │
│     Manager│Profiler│Scheduler│Sampler                 │
└─────────────────────────────────────────────────────────┘
```

## 📖 使用示例

### 文件配置

```go
profiler := ginpprof.New().
    WithFileConfig("./profiling.yaml").
    WithFileStorage("./profiles").
    Build()
```

### Nacos 配置

```go
profiler := ginpprof.New().
    WithNacosConfig(config.NacosOptions{
        ServerAddr: "127.0.0.1:8848",
        Namespace:  "production",
        Group:      "profiling",
        DataID:     "gin-pprof.yaml",
        Username:   "nacos",
        Password:   "nacos",
    }).
    WithFileStorage("./profiles").
    Build()
```

### 高级配置

```go
profiler := ginpprof.New().
    WithFileConfig("./profiling.yaml").
    WithFileStorage("./profiles").
    WithLogger(customLogger).
    WithOptions(core.Options{
        MaxConcurrent:     5,
        DefaultDuration:   30 * time.Second,
        CleanupInterval:   10 * time.Minute,
        MaxFileAge:        24 * time.Hour,
        Enabled:           true,
        ProfileDir:        "./profiles",
        DefaultSampleRate: 1,
    }).
    Build()
```

## 🔧 配置参考

### 性能分析配置

| 字段 | 类型 | 描述 | 默认值 |
|------|------|------|--------|
| `path` | string | 路由路径模式（如 `/users/:id`） | 必填 |
| `methods` | array | HTTP 方法数组：`["GET"]`、`["POST", "PUT"]` 或 `["*"]` 匹配常见方法 | `["GET"]` |
| `expires_at` | string | 过期时间，RFC3339 格式 | 必填 |
| `duration` | int | 分析持续时间（秒） | 30 |
| `sample_rate` | int | 每 N 个请求分析一次 | 1 |
| `profile_type` | string | 分析类型：`cpu`, `heap`, `goroutine` | cpu |

### 选项配置

| 字段 | 类型 | 描述 | 默认值 |
|------|------|------|--------|
| `max_concurrent` | int | 最大并发分析会话数 | 3 |
| `default_duration` | duration | 默认分析持续时间 | 30s |
| `cleanup_interval` | duration | 清理间隔 | 10m |
| `max_file_age` | duration | 文件最大保存时间 | 24h |
| `enabled` | bool | 启用/禁用分析 | true |
| `profile_dir` | string | 分析文件目录 | ./profiles |
| `default_sample_rate` | int | 默认采样率 | 1 |

## 🔌 适配器

### 配置提供器

- **文件**：本地 YAML/JSON 文件，支持热重载
- **Nacos**：Nacos 配置中心，支持实时更新
- **环境变量**：环境变量配置（即将支持）
- **Consul**：Consul KV 存储（即将支持）

### 存储后端

- **文件**：本地文件系统
- **内存**：内存存储（用于测试）
- **S3**：AWS S3（即将支持）
- **OSS**：阿里云对象存储（即将支持）

### 日志记录

- **标准**：Go 标准库日志记录器
- **空日志**：禁用日志记录
- **Zap**：Uber Zap 集成（即将支持）
- **Logrus**：Sirupsen Logrus 集成（即将支持）

## 📊 监控

### 状态端点

```bash
curl http://localhost:8080/debug/profiling/status
```

```json
{
  "enabled": true,
  "stats": {
    "total_requests": 1250,
    "profiled_count": 45,
    "failed_count": 2,
    "active_profiles": 1,
    "last_update": "2025-08-09T10:30:00Z"
  },
  "active_tasks": 3,
  "total_tasks": 5,
  "profile_dir": "./profiles"
}
```

### 统计端点

```bash
curl http://localhost:8080/debug/profiling/stats
```

```json
{
  "total_requests": 1250,
  "profiled_count": 45,
  "failed_count": 2,
  "active_profiles": 1,
  "success_rate": 95.6,
  "last_update": "2025-08-09T10:30:00Z"
}
```

## 🔥 分析性能文件

### 查看 CPU 分析

```bash
go tool pprof ./profiles/cpu/profile_api_users__id_20250809_103000_123456.pprof
```

### 生成火焰图

```bash
# 安装 pprof 工具
go install github.com/google/pprof@latest

# 生成火焰图
pprof -http=:8081 ./profiles/cpu/profile_api_users__id_20250809_103000_123456.pprof
```

### Web 界面

在浏览器中打开 `http://localhost:8081` 查看交互式火焰图。

## 🛡️ 生产环境最佳实践

### 1. 使用合适的采样率

```yaml
profiles:
  # 高流量端点 - 每 100 个请求采样一次
  - path: "/api/search"
    sample_rate: 100
    
  # 关键端点 - 每个请求都分析
  - path: "/api/payment"
    sample_rate: 1
```

### 2. 设置合理的持续时间

```yaml
profiles:
  # 频繁端点使用短持续时间
  - path: "/api/health"
    duration: 5
    
  # 复杂操作使用长持续时间
  - path: "/api/analytics/report"
    duration: 30
```

### 3. 限制并发会话

```go
profiler := ginpprof.New().
    WithOptions(core.Options{
        MaxConcurrent: 2,  // 最多 2 个并发分析会话
    }).
    Build()
```

### 4. 定期清理

```go
profiler := ginpprof.New().
    WithOptions(core.Options{
        CleanupInterval: 5 * time.Minute,  // 每 5 分钟清理一次
        MaxFileAge:      1 * time.Hour,    // 保存 1 小时
    }).
    Build()
```

## 🧪 测试

```bash
# 运行单元测试
go test ./...

# 运行集成测试
go test -tags=integration ./...

# 使用竞态检测运行
go test -race ./...
```

## 📚 示例

查看 [examples](./examples/) 目录获取完整的工作示例：

- [基本使用](./examples/basic/) - 简单的文件配置
- [Nacos 集成](./examples/nacos/) - Nacos 配置中心
- [高级设置](./examples/advanced/) - 自定义选项和多端点

## 🤝 贡献

欢迎贡献！请阅读我们的[贡献指南](./CONTRIBUTING.md)了解行为准则和提交拉取请求的流程。

### 开发环境设置

```bash
git clone https://github.com/aclstack/gin-pprof.git
cd gin-pprof
go mod tidy
go test ./...
```

## 🆚 对比其他方案

| 特性 | gin-pprof | net/http/pprof | 其他工具 |
|------|-----------|----------------|----------|
| 动态配置 | ✅ | ❌ | ❌ |
| 路由感知 | ✅ | ❌ | ❌ |
| 生产安全 | ✅ | ⚠️ | ⚠️ |
| 采样控制 | ✅ | ❌ | ⚠️ |
| 配置中心集成 | ✅ | ❌ | ❌ |
| 自动清理 | ✅ | ❌ | ❌ |

## 📄 许可证

本项目基于 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件。

## 🙏 致谢

- 受 Go 内置 `net/http/pprof` 包启发
- 基于优秀的 [Gin](https://github.com/gin-gonic/gin) Web 框架构建
- 配置管理由 [Nacos](https://github.com/alibaba/nacos) 提供支持

## 📧 支持

- 📖 [文档](./docs/)
- 🐛 [问题追踪](https://github.com/aclstack/gin-pprof/issues)
- 💬 [讨论区](https://github.com/aclstack/gin-pprof/discussions)
- 🔧 技术交流群：[添加微信群二维码]

## 🌟 用户案例

如果你在生产环境中使用了 gin-pprof，欢迎提交你的用户案例！

---

⭐ **如果这个项目对你有帮助，请给我们一个星标！**