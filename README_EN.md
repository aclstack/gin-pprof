# gin-pprof

[![Go](https://github.com/aclstack/gin-pprof/workflows/Go/badge.svg)](https://github.com/aclstack/gin-pprof/actions)
[![GoDoc](https://godoc.org/github.com/aclstack/gin-pprof?status.svg)](https://godoc.org/github.com/aclstack/gin-pprof)
[![Go Report Card](https://goreportcard.com/badge/github.com/aclstack/gin-pprof)](https://goreportcard.com/report/github.com/aclstack/gin-pprof)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A dynamic, configuration-driven profiling middleware for Gin web framework that enables **production-safe** performance analysis without service restarts.

English|[中文](README) 
## ✨ Features

- 🔄 **Dynamic Configuration**: Enable/disable profiling for specific endpoints without restarting your service
- 🎯 **Route-Aware**: Intelligent matching of parameterized routes (e.g., `/users/:id`)
- 🏗️ **Multiple Profiling Types**: CPU, memory (heap), and goroutine profiling support
- 📊 **Smart Sampling**: Configurable sample rates to minimize performance impact
- 🚦 **Concurrency Control**: Built-in limits to prevent resource exhaustion
- 🔌 **Pluggable Architecture**: Support for multiple configuration sources and storage backends
- ☁️ **Cloud-Native**: Nacos configuration center integration
- 📈 **Monitoring**: Built-in statistics and health endpoints
- 🛡️ **Production-Safe**: Non-blocking design with automatic cleanup

## 🚀 Quick Start

### Installation

```bash
go get github.com/aclstack/gin-pprof
```

### Basic Usage

```go
package main

import (
    "github.com/gin-gonic/gin"
    ginpprof "github.com/aclstack/gin-pprof"
)

func main() {
    r := gin.Default()

    // Create profiler with file-based configuration
    // If profiling.yaml doesn't exist, an example config will be auto-created
    profiler := ginpprof.New().
        WithFileConfig("./profiling.yaml").
        WithFileStorage("./profiles").
        Build()
    defer profiler.Close()

    // Add profiling middleware
    r.Use(profiler.Middleware())

    // Add monitoring endpoints
    debug := r.Group("/debug/profiling")
    {
        debug.GET("/status", profiler.StatusHandler())
        debug.GET("/tasks", profiler.TasksHandler())
        debug.GET("/stats", profiler.StatsHandler())
    }

    // Your API endpoints
    r.GET("/api/users/:id", getUser)
    
    r.Run(":8080")
}
```

> **📝 Note**: If `profiling.yaml` doesn't exist, gin-pprof will automatically create an example configuration file with detailed comments. Simply uncomment and modify the examples to enable profiling for your endpoints.

### Configuration File (profiling.yaml)

```yaml
profiles:
  # Single HTTP method
  - path: "/api/users/:id"
    methods: ["GET"]      # Single method
    expires_at: "2025-12-31T23:59:59Z"
    duration: 10          # Profile for 10 seconds
    profile_type: "cpu"   # CPU profiling
  
  # Multiple HTTP methods
  - path: "/api/heavy"
    methods: ["POST", "PUT"]  # Multiple methods array
    expires_at: "2025-12-31T23:59:59Z"
    duration: 15
    profile_type: "heap"  # Memory profiling
    
  # Wildcard for common methods
  - path: "/api/data/*"
    methods: ["*"]        # Matches GET, POST, PUT, DELETE
    expires_at: "2025-12-31T23:59:59Z"
    duration: 20
    profile_type: "goroutine"
```

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────┐
│                   Application Layer                     │
├─────────────────────────────────────────────────────────┤
│                 Adapter Layer                           │
│  Gin│Echo│Fiber  Nacos│File│Etcd  Local│S3│OSS         │
├─────────────────────────────────────────────────────────┤
│                 Interface Layer                         │
│  HTTPContext│ConfigProvider│Storage│Logger              │
├─────────────────────────────────────────────────────────┤
│                  Core Layer                             │
│     Manager│Profiler│Scheduler│Sampler                 │
└─────────────────────────────────────────────────────────┘
```

## 📖 Usage Examples

### File-Based Configuration

```go
profiler := ginpprof.New().
    WithFileConfig("./profiling.yaml").
    WithFileStorage("./profiles").
    Build()
```

### Nacos Configuration

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

### Advanced Configuration

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

## 🔧 Configuration Reference

### Profile Configuration

| Field | Type | Description | Default |
|-------|------|-------------|---------|
| `path` | string | Route path pattern (e.g., `/users/:id`) | required |
| `methods` | array | HTTP methods array: `["GET"]`, `["POST", "PUT"]`, or `["*"]` for common methods | `["GET"]` |
| `expires_at` | string | Expiration time in RFC3339 format | required |
| `duration` | int | Profiling duration in seconds | 30 |
| `sample_rate` | int | Profile every N requests | 1 |
| `profile_type` | string | Profiling type: `cpu`, `heap`, `goroutine` | cpu |

### Options Configuration

| Field | Type | Description | Default |
|-------|------|-------------|---------|
| `max_concurrent` | int | Maximum concurrent profiling sessions | 3 |
| `default_duration` | duration | Default profiling duration | 30s |
| `cleanup_interval` | duration | Profile cleanup interval | 10m |
| `max_file_age` | duration | Maximum age of profile files | 24h |
| `enabled` | bool | Enable/disable profiling | true |
| `profile_dir` | string | Profile files directory | ./profiles |
| `default_sample_rate` | int | Default sample rate | 1 |

## 🔌 Adapters

### Configuration Providers

- **File**: Local YAML/JSON files with hot-reload support
- **Nacos**: Nacos configuration center with real-time updates
- **Environment**: Environment variables (coming soon)
- **Consul**: Consul KV store (coming soon)

### Storage Backends

- **File**: Local file system
- **Memory**: In-memory storage (for testing)
- **S3**: AWS S3 (coming soon)
- **OSS**: Alibaba Cloud OSS (coming soon)

### Logging

- **Standard**: Go standard library logger
- **Noop**: Disabled logging
- **Zap**: Uber Zap integration (coming soon)
- **Logrus**: Sirupsen Logrus integration (coming soon)

## 📊 Monitoring

### Status Endpoint

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

### Statistics Endpoint

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

## 🔥 Analyzing Profiles

### View CPU Profile

```bash
go tool pprof ./profiles/cpu/profile_api_users__id_20250809_103000_123456.pprof
```

### Generate Flame Graph

```bash
# Install pprof tool
go install github.com/google/pprof@latest

# Generate flame graph
pprof -http=:8081 ./profiles/cpu/profile_api_users__id_20250809_103000_123456.pprof
```

### Web Interface

Open `http://localhost:8081` in your browser to view the interactive flame graph.

## 🛡️ Production Best Practices

### 1. Use Appropriate Sample Rates

```yaml
profiles:
  # High-traffic endpoint - sample every 100th request
  - path: "/api/search"
    sample_rate: 100
    
  # Critical endpoint - profile every request
  - path: "/api/payment"
    sample_rate: 1
```

### 2. Set Reasonable Durations

```yaml
profiles:
  # Short duration for frequent endpoints
  - path: "/api/health"
    duration: 5
    
  # Longer duration for complex operations
  - path: "/api/analytics/report"
    duration: 30
```

### 3. Limit Concurrent Sessions

```go
profiler := ginpprof.New().
    WithOptions(core.Options{
        MaxConcurrent: 2,  // Max 2 concurrent profiling sessions
    }).
    Build()
```

### 4. Regular Cleanup

```go
profiler := ginpprof.New().
    WithOptions(core.Options{
        CleanupInterval: 5 * time.Minute,  // Clean every 5 minutes
        MaxFileAge:      1 * time.Hour,    // Keep profiles for 1 hour
    }).
    Build()
```

## 🧪 Testing

```bash
# Run unit tests
go test ./...

# Run integration tests
go test -tags=integration ./...

# Run with race detection
go test -race ./...
```

## 📚 Examples

Check out the [examples](./examples/) directory for complete working examples:

- [Basic Usage](./examples/basic/) - Simple file-based configuration
- [Nacos Integration](./examples/nacos/) - Nacos configuration center
- [Advanced Setup](./examples/advanced/) - Custom options and multiple endpoints

## 🤝 Contributing

Contributions are welcome! Please read our [Contributing Guide](./CONTRIBUTING.md) for details on our code of conduct and the process for submitting pull requests.

### Development Setup

```bash
git clone https://github.com/aclstack/gin-pprof.git
cd gin-pprof
go mod tidy
go test ./...
```

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Inspired by the built-in `net/http/pprof` package
- Built on top of the excellent [Gin](https://github.com/gin-gonic/gin) web framework
- Configuration management powered by [Nacos](https://github.com/alibaba/nacos)

## 📧 Support

- 📖 [Documentation](./docs/)
- 🐛 [Issue Tracker](https://github.com/aclstack/gin-pprof/issues)
- 💬 [Discussions](https://github.com/aclstack/gin-pprof/discussions)

---

⭐ **Star this project** if you find it useful!