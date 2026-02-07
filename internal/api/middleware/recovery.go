package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

// Recovery middleware recovers from panics
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Log the panic (in production use proper logger)
				requestID, _ := c.Get("request_id")
				
				// Log detailed panic information
				fmt.Printf("[PANIC] Request ID: %v, Error: %v\n", requestID, err)
				fmt.Printf("[PANIC] Stack trace:\n%s\n", debug.Stack())
				
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":      "Internal server error",
					"request_id": requestID,
				})
				
				c.Abort()
			}
		}()
		
		c.Next()
	}
}
