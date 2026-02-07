package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Logging middleware adds structured logging to all requests
func Logging() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Add request ID
		requestID := uuid.New().String()
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method
		
		// Process request
		c.Next()
		
		// Log after request
		latency := time.Since(start)
		statusCode := c.Writer.Status()
		
		// Structured log (in production, use proper logger like zap or logrus)
		logData := map[string]interface{}{
			"timestamp":  time.Now().Format(time.RFC3339),
			"request_id": requestID,
			"method":     method,
			"path":       path,
			"status":     statusCode,
			"latency_ms": latency.Milliseconds(),
			"ip":         c.ClientIP(),
		}
		
		if userID, exists := c.Get("user_id"); exists {
			logData["user_id"] = userID
		}
		
		// Log based on status code
		if statusCode >= 500 {
			// Log error (in production use proper logger)
		}
		
		// In production, output structured JSON logs
		// For now, simple output
		fmt.Printf("[%s] %s %s - %d (%dms)\n", 
			time.Now().Format("2006-01-02 15:04:05"),
			method, path, statusCode, latency.Milliseconds())
	}
}
