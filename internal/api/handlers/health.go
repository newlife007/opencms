package handlers

import (
	"context"
	"fmt"
	"net/http"
	"os/exec"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

// HealthCheckDependencies holds all dependencies needed for health checks
type HealthCheckDependencies struct {
	DB          *gorm.DB
	RedisClient *redis.Client
	FFmpegPath  string
	QueueHealthCheck func(context.Context) error
	StorageHealthCheck func(context.Context) error
}

var healthDeps *HealthCheckDependencies

// SetHealthCheckDependencies sets the dependencies for health checks
func SetHealthCheckDependencies(deps *HealthCheckDependencies) {
	healthDeps = deps
}

// HealthCheck handles /health endpoint with comprehensive dependency checks
func HealthCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
		defer cancel()
		
		startTime := time.Now()
		
		health := gin.H{
			"status":    "healthy",
			"service":   "openwan-api",
			"version":   "1.0.0",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"uptime":    getUptime(),
		}
		
		checks := make(map[string]interface{})
		allHealthy := true
		
		// Check database connectivity
		if healthDeps != nil && healthDeps.DB != nil {
			dbCheck := checkDatabase(ctx, healthDeps.DB)
			checks["database"] = dbCheck
			if dbCheck["status"] != "healthy" {
				allHealthy = false
			}
		} else {
			checks["database"] = gin.H{"status": "unknown", "message": "database not initialized"}
			allHealthy = false
		}
		
		// Check Redis connectivity
		if healthDeps != nil && healthDeps.RedisClient != nil {
			redisCheck := checkRedis(ctx, healthDeps.RedisClient)
			checks["redis"] = redisCheck
			if redisCheck["status"] != "healthy" {
				allHealthy = false
			}
		} else {
			checks["redis"] = gin.H{"status": "unknown", "message": "redis not initialized"}
		}
		
		// Check RabbitMQ/Queue connectivity
		if healthDeps != nil && healthDeps.QueueHealthCheck != nil {
			queueCheck := checkQueue(ctx, healthDeps.QueueHealthCheck)
			checks["queue"] = queueCheck
			if queueCheck["status"] != "healthy" {
				allHealthy = false
			}
		} else {
			checks["queue"] = gin.H{"status": "unknown", "message": "queue not initialized"}
		}
		
		// Check Storage (S3) connectivity
		if healthDeps != nil && healthDeps.StorageHealthCheck != nil {
			storageCheck := checkStorage(ctx, healthDeps.StorageHealthCheck)
			checks["storage"] = storageCheck
			if storageCheck["status"] != "healthy" {
				// Storage check is non-critical for read operations
				// allHealthy = false
			}
		} else {
			checks["storage"] = gin.H{"status": "unknown", "message": "storage not initialized"}
		}
		
		// Check FFmpeg availability
		if healthDeps != nil && healthDeps.FFmpegPath != "" {
			ffmpegCheck := checkFFmpeg(healthDeps.FFmpegPath)
			checks["ffmpeg"] = ffmpegCheck
			// FFmpeg is not critical for API health
		} else {
			checks["ffmpeg"] = gin.H{"status": "unknown", "message": "ffmpeg path not configured"}
		}
		
		health["checks"] = checks
		health["response_time_ms"] = time.Since(startTime).Milliseconds()
		
		if !allHealthy {
			health["status"] = "unhealthy"
			c.JSON(http.StatusServiceUnavailable, health)
			return
		}
		
		c.JSON(http.StatusOK, health)
	}
}

// ReadyCheck handles /ready endpoint for readiness probe
func ReadyCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()
		
		// Service is ready if critical dependencies are accessible
		ready := true
		checks := make(map[string]string)
		
		// Check database - critical for readiness
		if healthDeps != nil && healthDeps.DB != nil {
			if err := healthDeps.DB.WithContext(ctx).Exec("SELECT 1").Error; err != nil {
				ready = false
				checks["database"] = fmt.Sprintf("not ready: %v", err)
			} else {
				checks["database"] = "ready"
			}
		} else {
			ready = false
			checks["database"] = "not initialized"
		}
		
		// Check Redis - critical for session management
		if healthDeps != nil && healthDeps.RedisClient != nil {
			if err := healthDeps.RedisClient.Ping(ctx).Err(); err != nil {
				ready = false
				checks["redis"] = fmt.Sprintf("not ready: %v", err)
			} else {
				checks["redis"] = "ready"
			}
		} else {
			ready = false
			checks["redis"] = "not initialized"
		}
		
		if !ready {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "not ready",
				"checks": checks,
			})
			return
		}
		
		c.JSON(http.StatusOK, gin.H{
			"status": "ready",
			"checks": checks,
		})
	}
}

// AliveCheck handles /alive endpoint for liveness probe
func AliveCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Simple liveness check - process is running
		c.JSON(http.StatusOK, gin.H{
			"status":    "alive",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	}
}

// Detailed health check endpoint for operations team
func DetailedHealthCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
		defer cancel()
		
		health := gin.H{
			"service":   "openwan-api",
			"version":   "1.0.0",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"uptime":    getUptime(),
			"system":    getSystemInfo(),
		}
		
		checks := make(map[string]interface{})
		
		// Detailed checks for all dependencies
		if healthDeps != nil {
			if healthDeps.DB != nil {
				checks["database"] = checkDatabaseDetailed(ctx, healthDeps.DB)
			}
			if healthDeps.RedisClient != nil {
				checks["redis"] = checkRedisDetailed(ctx, healthDeps.RedisClient)
			}
			if healthDeps.QueueHealthCheck != nil {
				checks["queue"] = checkQueue(ctx, healthDeps.QueueHealthCheck)
			}
			if healthDeps.StorageHealthCheck != nil {
				checks["storage"] = checkStorage(ctx, healthDeps.StorageHealthCheck)
			}
			if healthDeps.FFmpegPath != "" {
				checks["ffmpeg"] = checkFFmpegDetailed(healthDeps.FFmpegPath)
			}
		}
		
		health["checks"] = checks
		c.JSON(http.StatusOK, health)
	}
}

var startTime = time.Now()

func getUptime() string {
	uptime := time.Since(startTime)
	return fmt.Sprintf("%.0f seconds", uptime.Seconds())
}

func getSystemInfo() gin.H {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	return gin.H{
		"num_goroutines": runtime.NumGoroutine(),
		"num_cpu":        runtime.NumCPU(),
		"memory_alloc":   fmt.Sprintf("%.2f MB", float64(m.Alloc)/1024/1024),
		"memory_sys":     fmt.Sprintf("%.2f MB", float64(m.Sys)/1024/1024),
	}
}

// checkDatabase performs database health check
func checkDatabase(ctx context.Context, db *gorm.DB) gin.H {
	sqlDB, err := db.DB()
	if err != nil {
		return gin.H{
			"status": "unhealthy",
			"error":  fmt.Sprintf("failed to get sql.DB: %v", err),
		}
	}
	
	if err := sqlDB.PingContext(ctx); err != nil {
		return gin.H{
			"status": "unhealthy",
			"error":  fmt.Sprintf("ping failed: %v", err),
		}
	}
	
	stats := sqlDB.Stats()
	
	return gin.H{
		"status":           "healthy",
		"open_connections": stats.OpenConnections,
		"in_use":           stats.InUse,
		"idle":             stats.Idle,
	}
}

func checkDatabaseDetailed(ctx context.Context, db *gorm.DB) gin.H {
	result := checkDatabase(ctx, db)
	
	// Add more details
	sqlDB, _ := db.DB()
	if sqlDB != nil {
		stats := sqlDB.Stats()
		result["max_open_connections"] = stats.MaxOpenConnections
		result["wait_count"] = stats.WaitCount
		result["wait_duration"] = stats.WaitDuration.String()
	}
	
	return result
}

// checkRedis performs Redis health check
func checkRedis(ctx context.Context, client *redis.Client) gin.H {
	if err := client.Ping(ctx).Err(); err != nil {
		return gin.H{
			"status": "unhealthy",
			"error":  fmt.Sprintf("ping failed: %v", err),
		}
	}
	
	return gin.H{
		"status": "healthy",
	}
}

func checkRedisDetailed(ctx context.Context, client *redis.Client) gin.H {
	result := checkRedis(ctx, client)
	
	// Add Redis info
	if result["status"] == "healthy" {
		info, err := client.Info(ctx, "stats").Result()
		if err == nil {
			result["info"] = info
		}
	}
	
	return result
}

// checkQueue performs queue health check
func checkQueue(ctx context.Context, healthCheck func(context.Context) error) gin.H {
	if err := healthCheck(ctx); err != nil {
		return gin.H{
			"status": "unhealthy",
			"error":  fmt.Sprintf("queue check failed: %v", err),
		}
	}
	
	return gin.H{
		"status": "healthy",
	}
}

// checkStorage performs storage health check
func checkStorage(ctx context.Context, healthCheck func(context.Context) error) gin.H {
	if err := healthCheck(ctx); err != nil {
		return gin.H{
			"status": "unhealthy",
			"error":  fmt.Sprintf("storage check failed: %v", err),
		}
	}
	
	return gin.H{
		"status": "healthy",
	}
}

// checkFFmpeg verifies FFmpeg binary availability
func checkFFmpeg(path string) gin.H {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	cmd := exec.CommandContext(ctx, path, "-version")
	if err := cmd.Run(); err != nil {
		return gin.H{
			"status": "unhealthy",
			"error":  fmt.Sprintf("ffmpeg not available: %v", err),
		}
	}
	
	return gin.H{
		"status": "healthy",
		"path":   path,
	}
}

func checkFFmpegDetailed(path string) gin.H {
	result := checkFFmpeg(path)
	
	if result["status"] == "healthy" {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		
		cmd := exec.CommandContext(ctx, path, "-version")
		output, err := cmd.Output()
		if err == nil {
			result["version"] = string(output[:100]) // First 100 chars
		}
	}
	
	return result
}
