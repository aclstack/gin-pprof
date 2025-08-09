package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	ginpprof "github.com/aclstack/gin-pprof"
	"github.com/aclstack/gin-pprof/pkg/adapters/logger"
	"github.com/aclstack/gin-pprof/pkg/core"
)

func main() {
	// Create custom logger (could be zap, logrus, etc.)
	customLogger := logger.NewStandardLogger("my-app-profiler")

	// Create profiler with advanced configuration
	profiler := ginpprof.New().
		WithFileConfig("./profiling.yaml").
		WithFileStorage("./profiles").
		WithLogger(customLogger).
		WithOptions(core.Options{
			MaxConcurrent:     10,                // Allow up to 10 concurrent profiling sessions
			DefaultDuration:   45 * time.Second, // Default 45 second profiling duration
			CleanupInterval:   2 * time.Minute,  // Clean up every 2 minutes
			MaxFileAge:        30 * time.Minute, // Keep profiles for 30 minutes
			Enabled:           true,
			ProfileDir:        "./profiles",
			DefaultSampleRate: 1,
		}).
		Build()
	defer profiler.Close()

	// Create Gin router with custom middleware
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	
	// Add profiling middleware
	r.Use(profiler.Middleware())

	// Add custom profiling dashboard
	r.Static("/static", "./static")
	r.LoadHTMLGlob("templates/*")
	
	// Profiling dashboard
	r.GET("/profiling", func(c *gin.Context) {
		stats := profiler.GetStats()
		tasks := profiler.GetTasks()
		
		c.HTML(http.StatusOK, "dashboard.html", gin.H{
			"stats": stats,
			"tasks": tasks,
		})
	})

	// Standard profiling endpoints
	debug := r.Group("/debug/profiling")
	{
		debug.GET("/status", profiler.StatusHandler())
		debug.GET("/tasks", profiler.TasksHandler())
		debug.GET("/stats", profiler.StatsHandler())
	}

	// Sample application endpoints
	setupRoutes(r)

	log.Println("Server starting on :8080")
	log.Println("Profiling dashboard: http://localhost:8080/profiling")
	log.Println("Profiling status: http://localhost:8080/debug/profiling/status")
	
	r.Run(":8080")
}

func setupRoutes(r *gin.Engine) {
	api := r.Group("/api/v1")
	
	// User management
	users := api.Group("/users")
	{
		users.GET("", listUsers)
		users.POST("", createUser)
		users.GET("/:id", getUser)
		users.PUT("/:id", updateUser)
		users.DELETE("/:id", deleteUser)
		users.GET("/:id/profile", getUserProfile)
		users.PUT("/:id/profile", updateUserProfile)
	}
	
	// Product management
	products := api.Group("/products")
	{
		products.GET("", listProducts)
		products.POST("", createProduct)
		products.GET("/:id", getProduct)
		products.PUT("/:id", updateProduct)
		products.DELETE("/:id", deleteProduct)
		products.GET("/:id/reviews", getProductReviews)
	}
	
	// Order management
	orders := api.Group("/orders")
	{
		orders.GET("", listOrders)
		orders.POST("", createOrder)
		orders.GET("/:id", getOrder)
		orders.PUT("/:id/status", updateOrderStatus)
		orders.GET("/:id/items", getOrderItems)
	}
	
	// Analytics endpoints (CPU intensive)
	analytics := api.Group("/analytics")
	{
		analytics.GET("/sales", getSalesAnalytics)
		analytics.GET("/users", getUserAnalytics)
		analytics.GET("/products", getProductAnalytics)
		analytics.POST("/report", generateReport)
	}
	
	// Resource-intensive endpoints
	intensive := api.Group("/intensive")
	{
		intensive.GET("/cpu", cpuIntensiveTask)
		intensive.GET("/memory", memoryIntensiveTask)
		intensive.GET("/io", ioIntensiveTask)
		intensive.GET("/mixed", mixedWorkload)
	}
}

// User endpoints
func listUsers(c *gin.Context) {
	time.Sleep(5 * time.Millisecond)
	c.JSON(http.StatusOK, gin.H{"users": []string{"user1", "user2", "user3"}})
}

func createUser(c *gin.Context) {
	// Simulate validation and creation
	time.Sleep(25 * time.Millisecond)
	c.JSON(http.StatusCreated, gin.H{"id": 123, "status": "created"})
}

func getUser(c *gin.Context) {
	id := c.Param("id")
	time.Sleep(3 * time.Millisecond)
	c.JSON(http.StatusOK, gin.H{"id": id, "name": "User " + id})
}

func updateUser(c *gin.Context) {
	time.Sleep(15 * time.Millisecond)
	c.JSON(http.StatusOK, gin.H{"status": "updated"})
}

func deleteUser(c *gin.Context) {
	time.Sleep(8 * time.Millisecond)
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

func getUserProfile(c *gin.Context) {
	time.Sleep(12 * time.Millisecond)
	c.JSON(http.StatusOK, gin.H{"profile": "user profile data"})
}

func updateUserProfile(c *gin.Context) {
	time.Sleep(20 * time.Millisecond)
	c.JSON(http.StatusOK, gin.H{"status": "profile updated"})
}

// Product endpoints
func listProducts(c *gin.Context) {
	time.Sleep(8 * time.Millisecond)
	c.JSON(http.StatusOK, gin.H{"products": []string{"product1", "product2"}})
}

func createProduct(c *gin.Context) {
	time.Sleep(30 * time.Millisecond)
	c.JSON(http.StatusCreated, gin.H{"status": "product created"})
}

func getProduct(c *gin.Context) {
	time.Sleep(4 * time.Millisecond)
	c.JSON(http.StatusOK, gin.H{"product": "product data"})
}

func updateProduct(c *gin.Context) {
	time.Sleep(18 * time.Millisecond)
	c.JSON(http.StatusOK, gin.H{"status": "product updated"})
}

func deleteProduct(c *gin.Context) {
	time.Sleep(10 * time.Millisecond)
	c.JSON(http.StatusOK, gin.H{"status": "product deleted"})
}

func getProductReviews(c *gin.Context) {
	time.Sleep(15 * time.Millisecond)
	c.JSON(http.StatusOK, gin.H{"reviews": []string{"review1", "review2"}})
}

// Order endpoints
func listOrders(c *gin.Context) {
	time.Sleep(12 * time.Millisecond)
	c.JSON(http.StatusOK, gin.H{"orders": []string{"order1", "order2"}})
}

func createOrder(c *gin.Context) {
	time.Sleep(40 * time.Millisecond)
	c.JSON(http.StatusCreated, gin.H{"status": "order created"})
}

func getOrder(c *gin.Context) {
	time.Sleep(6 * time.Millisecond)
	c.JSON(http.StatusOK, gin.H{"order": "order data"})
}

func updateOrderStatus(c *gin.Context) {
	time.Sleep(22 * time.Millisecond)
	c.JSON(http.StatusOK, gin.H{"status": "order status updated"})
}

func getOrderItems(c *gin.Context) {
	time.Sleep(14 * time.Millisecond)
	c.JSON(http.StatusOK, gin.H{"items": []string{"item1", "item2"}})
}

// Analytics endpoints (CPU intensive)
func getSalesAnalytics(c *gin.Context) {
	// Simulate heavy computation
	sum := 0
	for i := 0; i < 1000000; i++ {
		sum += i
	}
	c.JSON(http.StatusOK, gin.H{"analytics": "sales data", "sum": sum})
}

func getUserAnalytics(c *gin.Context) {
	// Simulate complex calculation
	result := 1.0
	for i := 1; i < 100000; i++ {
		result *= float64(i) / float64(i+1)
	}
	c.JSON(http.StatusOK, gin.H{"analytics": "user data", "result": result})
}

func getProductAnalytics(c *gin.Context) {
	// Simulate data processing
	data := make([]int, 500000)
	for i := range data {
		data[i] = i * 2
	}
	sum := 0
	for _, v := range data {
		sum += v
	}
	c.JSON(http.StatusOK, gin.H{"analytics": "product data", "sum": sum})
}

func generateReport(c *gin.Context) {
	// Very CPU intensive task
	result := 0
	for i := 0; i < 10000000; i++ {
		result += i * i % 1000
	}
	c.JSON(http.StatusOK, gin.H{"report": "generated", "result": result})
}

// Resource-intensive endpoints
func cpuIntensiveTask(c *gin.Context) {
	// Pure CPU work
	result := 0
	for i := 0; i < 5000000; i++ {
		result += i * i
	}
	c.JSON(http.StatusOK, gin.H{"result": result})
}

func memoryIntensiveTask(c *gin.Context) {
	// Memory allocation
	data := make([][]byte, 1000)
	for i := range data {
		data[i] = make([]byte, 10000)
		for j := range data[i] {
			data[i][j] = byte(j % 256)
		}
	}
	
	// Use data to prevent optimization
	sum := 0
	for i := range data {
		for j := range data[i] {
			sum += int(data[i][j])
		}
	}
	c.JSON(http.StatusOK, gin.H{"sum": sum})
}

func ioIntensiveTask(c *gin.Context) {
	// Simulate I/O with sleeps
	for i := 0; i < 10; i++ {
		time.Sleep(5 * time.Millisecond)
		// Simulate some processing between I/O
		for j := 0; j < 10000; j++ {
			_ = j * j
		}
	}
	c.JSON(http.StatusOK, gin.H{"status": "I/O completed"})
}

func mixedWorkload(c *gin.Context) {
	// Mixed CPU, memory, and I/O
	// CPU work
	result := 0
	for i := 0; i < 1000000; i++ {
		result += i
	}
	
	// Memory work
	data := make([]int, 100000)
	for i := range data {
		data[i] = i * 2
	}
	
	// I/O simulation
	time.Sleep(10 * time.Millisecond)
	
	c.JSON(http.StatusOK, gin.H{
		"cpu_result": result,
		"memory_sum": len(data),
		"status":     "mixed workload completed",
	})
}