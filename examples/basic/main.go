package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	ginpprof "github.com/aclstack/gin-pprof"
)

func main() {
	// Create Gin router
	r := gin.Default()

	// Create profiler with file-based configuration
	profiler := ginpprof.New().
		WithFileConfig("./profiling.yaml").
		WithFileStorage("./profiles").
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
	r.GET("/api/users", getUsers)
	r.GET("/api/users/:id", getUser)
	r.POST("/api/users", createUser)
	r.GET("/api/heavy", heavyComputation)

	log.Println("Server starting on :8080")
	log.Println("Profiling status: http://localhost:8080/debug/profiling/status")
	log.Println("Create profiling.yaml to enable profiling for specific endpoints")
	
	r.Run(":8080")
}

func getUsers(c *gin.Context) {
	// Simulate some work
	time.Sleep(10 * time.Millisecond)
	
	users := []map[string]interface{}{
		{"id": 1, "name": "Alice"},
		{"id": 2, "name": "Bob"},
	}
	c.JSON(http.StatusOK, gin.H{"users": users})
}

func getUser(c *gin.Context) {
	id := c.Param("id")
	// Simulate database query
	time.Sleep(5 * time.Millisecond)
	
	c.JSON(http.StatusOK, gin.H{
		"id":   id,
		"name": "User " + id,
	})
}

func createUser(c *gin.Context) {
	// Simulate user creation
	time.Sleep(20 * time.Millisecond)
	
	c.JSON(http.StatusCreated, gin.H{
		"id":      123,
		"message": "User created successfully",
	})
}

func heavyComputation(c *gin.Context) {
	// Simulate heavy computation
	result := 0
	for i := 0; i < 1000000; i++ {
		result += i * i
	}
	
	c.JSON(http.StatusOK, gin.H{
		"result": result,
		"message": "Heavy computation completed",
	})
}