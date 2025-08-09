# Getting Started with gin-pprof

This guide will help you get started with gin-pprof, a dynamic profiling middleware for Gin applications.

## Installation

```bash
go get github.com/aclstack/gin-pprof
```

## Basic Setup

### 1. Add Middleware to Your Gin Application

```go
package main

import (
    "github.com/gin-gonic/gin"
    ginpprof "github.com/aclstack/gin-pprof"
)

func main() {
    r := gin.Default()

    // Create and configure profiler
    profiler := ginpprof.New().
        WithFileConfig("./profiling.yaml").
        WithFileStorage("./profiles").
        Build()
    defer profiler.Close()

    // Add profiling middleware
    r.Use(profiler.Middleware())

    // Add your routes
    r.GET("/api/users/:id", getUserHandler)
    r.GET("/api/orders", getOrdersHandler)
    
    r.Run(":8080")
}
```

### 2. Create Configuration File

Create a `profiling.yaml` file in your project root:

```yaml
profiles:
  - path: "/api/users/:id"
    expires_at: "2025-12-31T23:59:59Z"
    duration: 10
    sample_rate: 1
    profile_type: "cpu"
  
  - path: "/api/orders"
    expires_at: "2025-12-31T23:59:59Z"
    duration: 15
    sample_rate: 5
    profile_type: "heap"
```

### 3. Run Your Application

```bash
go run main.go
```

### 4. Test Profiling

Make requests to your profiled endpoints:

```bash
# This will trigger CPU profiling for /api/users/:id
curl http://localhost:8080/api/users/123

# This will trigger heap profiling every 5th request
curl http://localhost:8080/api/orders
```

### 5. Check Profile Files

Profile files will be saved in the `./profiles` directory:

```
profiles/
├── cpu/
│   └── profile_api_users__id_20250809_143025_123456.pprof
└── heap/
    └── profile_api_orders_20250809_143030_789012.pprof
```

### 6. Monitor Profiling Status

Access the built-in monitoring endpoints:

```bash
# Check profiling status
curl http://localhost:8080/debug/profiling/status

# View active tasks
curl http://localhost:8080/debug/profiling/tasks

# Get statistics
curl http://localhost:8080/debug/profiling/stats
```

## Next Steps

- [Configuration Guide](./CONFIGURATION.md) - Learn about advanced configuration options
- [Nacos Integration](./NACOS.md) - Set up dynamic configuration with Nacos
- [Profile Analysis](./ANALYSIS.md) - How to analyze generated profiles
- [Production Deployment](./PRODUCTION.md) - Best practices for production use

## Common Issues

### Profile Files Not Generated

1. Check if the endpoint path matches your configuration
2. Ensure the task hasn't expired
3. Verify the sample rate settings
4. Check application logs for errors

### High Resource Usage

1. Reduce the number of concurrent profiling sessions
2. Increase sample rates for high-traffic endpoints
3. Decrease profiling durations
4. Implement proper cleanup intervals

### Path Matching Issues

Make sure your path patterns match exactly:

```yaml
# Correct
- path: "/api/users/:id"

# Incorrect - missing leading slash
- path: "api/users/:id"

# Incorrect - different parameter name
- path: "/api/users/:userId"
```