package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// RateLimiter represents a rate limiter with different limits
type RateLimiter struct {
	visitors map[string]*rate.Limiter
	mu       sync.RWMutex
	rps      rate.Limit
	burst    int
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(rps int, burst int) *RateLimiter {
	return &RateLimiter{
		visitors: make(map[string]*rate.Limiter),
		rps:      rate.Limit(rps),
		burst:    burst,
	}
}

// GetLimiter returns the rate limiter for a visitor
func (rl *RateLimiter) GetLimiter(key string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.visitors[key]
	if !exists {
		limiter = rate.NewLimiter(rl.rps, rl.burst)
		rl.visitors[key] = limiter
	}

	return limiter
}

// CleanupVisitors removes old visitors periodically
func (rl *RateLimiter) CleanupVisitors() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		// Clear all visitors to prevent memory leak
		rl.visitors = make(map[string]*rate.Limiter)
		rl.mu.Unlock()
	}
}

// RateLimitMiddleware creates a rate limiting middleware
func RateLimitMiddleware(rps, burst int) gin.HandlerFunc {
	limiter := NewRateLimiter(rps, burst)
	
	// Start cleanup goroutine
	go limiter.CleanupVisitors()

	return func(c *gin.Context) {
		// Use IP address as the key
		key := c.ClientIP()

		// Check if authenticated user, use user ID instead
		if userID, exists := c.Get("user_id"); exists {
			if uid, ok := userID.(string); ok && uid != "" {
				key = "user:" + uid
			}
		}

		limiter := limiter.GetLimiter(key)
		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "rate_limit_exceeded",
				"message": "Too many requests. Please try again later.",
				"retry_after": "60",
			})
			c.Header("Retry-After", "60")
			c.Abort()
			return
		}

		c.Next()
	}
}

// RateLimitConfig holds rate limit configurations for different endpoints
type RateLimitConfig struct {
	Global     *RateLimiter // Global rate limit for all endpoints
	PerIP      *RateLimiter // Per-IP rate limit
	PerUser    *RateLimiter // Per-user rate limit
	Anonymous  *RateLimiter // Rate limit for anonymous users
}

// NewRateLimitConfig creates a new rate limit configuration
func NewRateLimitConfig() *RateLimitConfig {
	return &RateLimitConfig{
		Global:    NewRateLimiter(1000, 2000), // 1000 rps globally
		PerIP:     NewRateLimiter(100, 200),   // 100 rps per IP
		PerUser:   NewRateLimiter(1000, 2000), // 1000 rps per user
		Anonymous: NewRateLimiter(10, 20),     // 10 rps for anonymous
	}
}

// AdvancedRateLimitMiddleware creates an advanced rate limiting middleware
func AdvancedRateLimitMiddleware(config *RateLimitConfig) gin.HandlerFunc {
	// Start cleanup for all limiters
	go config.PerIP.CleanupVisitors()
	go config.PerUser.CleanupVisitors()
	go config.Anonymous.CleanupVisitors()

	return func(c *gin.Context) {
		ip := c.ClientIP()
		userID, authenticated := c.Get("user_id")

		var limiter *rate.Limiter

		if authenticated {
			// Use per-user rate limit for authenticated users
			uid, ok := userID.(string)
			if ok && uid != "" {
				limiter = config.PerUser.GetLimiter("user:" + uid)
			} else {
				limiter = config.PerIP.GetLimiter(ip)
			}
		} else {
			// Use stricter rate limit for anonymous users
			limiter = config.Anonymous.GetLimiter(ip)
		}

		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "rate_limit_exceeded",
				"message":     "Too many requests. Please try again later.",
				"retry_after": "60",
			})
			c.Header("Retry-After", "60")
			c.Abort()
			return
		}

		c.Next()
	}
}
