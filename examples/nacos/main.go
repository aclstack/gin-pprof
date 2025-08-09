package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	ginpprof "github.com/aclstack/gin-pprof"
	"github.com/aclstack/gin-pprof/pkg/adapters/config"
	"github.com/aclstack/gin-pprof/pkg/core"
)

func main() {
	// Create Gin router
	r := gin.Default()

	// Create profiler with Nacos configuration
	profiler := ginpprof.New().
		WithNacosConfig(config.NacosOptions{
			ServerAddr: getEnv("NACOS_SERVER_ADDR", "127.0.0.1:8848"),
			Namespace:  getEnv("NACOS_NAMESPACE", "public"),
			Group:      getEnv("NACOS_GROUP", "DEFAULT_GROUP"),
			DataID:     getEnv("NACOS_DATA_ID", "gin-pprof.yaml"),
			Username:   getEnv("NACOS_USERNAME", ""),
			Password:   getEnv("NACOS_PASSWORD", ""),
		}).
		WithFileStorage("./profiles").
		WithOptions(core.Options{
			MaxConcurrent:     5,
			DefaultDuration:   30 * time.Second,
			CleanupInterval:   5 * time.Minute,
			MaxFileAge:        2 * time.Hour,
			Enabled:           true,
			ProfileDir:        "./profiles",
			DefaultSampleRate: 1,
		}).
		Build()
	defer profiler.Close()

	// Add profiling middleware
	r.Use(profiler.Middleware())

	// Add profiling status endpoints
	debug := r.Group("/debug/profiling")
	{
		debug.GET("/status", profiler.StatusHandler())
		debug.GET("/tasks", profiler.TasksHandler())
		debug.GET("/stats", profiler.StatsHandler())
	}

	// Sample API endpoints
	api := r.Group("/api/v1")
	{
		api.GET("/users", getUsers)
		api.GET("/users/:id", getUser)
		api.POST("/users", createUser)
		api.PUT("/users/:id", updateUser)
		api.DELETE("/users/:id", deleteUser)
		api.GET("/orders", getOrders)
		api.GET("/orders/:id", getOrder)
		api.GET("/heavy", heavyComputation)
		api.GET("/memory", memoryIntensiveTask)
	}

	log.Println("Server starting on :8080")
	log.Println("Profiling status: http://localhost:8080/debug/profiling/status")
	log.Println("Configure profiling tasks via Nacos:")
	log.Printf("- Server: %s", getEnv("NACOS_SERVER_ADDR", "127.0.0.1:8848"))
	log.Printf("- Namespace: %s", getEnv("NACOS_NAMESPACE", "public"))
	log.Printf("- Group: %s", getEnv("NACOS_GROUP", "DEFAULT_GROUP"))
	log.Printf("- DataID: %s", getEnv("NACOS_DATA_ID", "gin-pprof.yaml"))
	
	r.Run(":8080")
}

func getUsers(c *gin.Context) {
	// Simulate some work
	time.Sleep(10 * time.Millisecond)
	
	users := make([]map[string]interface{}, 100)
	for i := 0; i < 100; i++ {
		users[i] = map[string]interface{}{
			"id":    i + 1,
			"name":  "User " + string(rune(65+i%26)),
			"email": "user" + string(rune(48+i%10)) + "@example.com",
		}
	}
	c.JSON(http.StatusOK, gin.H{"users": users})
}

func getUser(c *gin.Context) {
	id := c.Param("id")
	// Simulate database query
	time.Sleep(5 * time.Millisecond)
	
	c.JSON(http.StatusOK, gin.H{
		"id":    id,
		"name":  "User " + id,
		"email": "user" + id + "@example.com",
	})
}

func createUser(c *gin.Context) {
	// Simulate user creation with validation
	time.Sleep(20 * time.Millisecond)
	
	// Simulate some CPU-intensive validation
	for i := 0; i < 10000; i++ {
		_ = i * i
	}
	
	c.JSON(http.StatusCreated, gin.H{
		"id":      123,
		"message": "User created successfully",
	})
}

func updateUser(c *gin.Context) {
	id := c.Param("id")
	// Simulate update logic
	time.Sleep(15 * time.Millisecond)
	
	c.JSON(http.StatusOK, gin.H{
		"id":      id,
		"message": "User updated successfully",
	})
}

func deleteUser(c *gin.Context) {
	id := c.Param("id")
	// Simulate deletion
	time.Sleep(8 * time.Millisecond)
	
	c.JSON(http.StatusOK, gin.H{
		"id":      id,
		"message": "User deleted successfully",
	})
}

func getOrders(c *gin.Context) {
	// Simulate database query with joins
	time.Sleep(25 * time.Millisecond)
	
	orders := make([]map[string]interface{}, 50)
	for i := 0; i < 50; i++ {
		orders[i] = map[string]interface{}{
			"id":     i + 1,
			"amount": float64(i*10 + 100),
			"status": "active",
		}
	}
	c.JSON(http.StatusOK, gin.H{"orders": orders})
}

func getOrder(c *gin.Context) {
	id := c.Param("id")
	// Simulate complex order query
	time.Sleep(12 * time.Millisecond)
	
	c.JSON(http.StatusOK, gin.H{
		"id":     id,
		"amount": 199.99,
		"status": "completed",
		"items":  []string{"item1", "item2", "item3"},
	})
}

func heavyComputation(c *gin.Context) {
	// CPU-intensive task
	result := 0
	for i := 0; i < 5000000; i++ {
		result += i * i
	}
	
	c.JSON(http.StatusOK, gin.H{
		"result":  result,
		"message": "Heavy computation completed",
	})
}

func memoryIntensiveTask(c *gin.Context) {
	// Memory-intensive task
	data := make([][]int, 1000)
	for i := range data {
		data[i] = make([]int, 1000)
		for j := range data[i] {
			data[i][j] = i * j
		}
	}
	
	// Use the data to prevent optimization
	sum := 0
	for i := range data {
		for j := range data[i] {
			sum += data[i][j]
		}
	}
	
	c.JSON(http.StatusOK, gin.H{
		"sum":     sum,
		"message": "Memory intensive task completed",
	})
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}