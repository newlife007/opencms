package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/openwan/media-asset-management/internal/session"
)

var sessionStore session.Store

// SetSessionStore sets the session store for authentication middleware
func SetSessionStore(store session.Store) {
	sessionStore = store
}

// Auth middleware validates authentication (optional - doesn't block if not authenticated)
func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Try session cookie first
		sessionID, err := c.Cookie("openwan_session")
		if err == nil && sessionID != "" && sessionStore != nil {
			sess, err := sessionStore.Get(c.Request.Context(), sessionID)
			if err == nil && sess != nil {
				// Set user information in context
				c.Set("user_id", uint(sess.UserID))
				c.Set("username", sess.Username)
				c.Set("group_id", uint(sess.GroupID))
				c.Set("is_admin", sess.IsAdmin)
				c.Next()
				return
			}
		}
		
		// Try Bearer token
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" && len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			token := authHeader[7:]
			if sessionStore != nil {
				sess, err := sessionStore.Get(c.Request.Context(), token)
				if err == nil && sess != nil {
					// Set user information in context
					c.Set("user_id", uint(sess.UserID))
					c.Set("username", sess.Username)
					c.Set("group_id", uint(sess.GroupID))
					c.Set("is_admin", sess.IsAdmin)
				}
			}
		}
		
		c.Next()
	}
}

// RequireAuth middleware requires authentication
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Try session cookie first
		sessionID, err := c.Cookie("openwan_session")
		if err == nil && sessionID != "" && sessionStore != nil {
			sess, err := sessionStore.Get(c.Request.Context(), sessionID)
			if err == nil && sess != nil {
				// Set user information in context
				c.Set("user_id", uint(sess.UserID))
				c.Set("username", sess.Username)
				c.Set("group_id", uint(sess.GroupID))
				c.Set("is_admin", sess.IsAdmin)
				c.Next()
				return
			}
		}
		
		// Try Bearer token
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" && len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			token := authHeader[7:]
			// Validate token using session store (token is stored as sessionID)
			if sessionStore != nil {
				sess, err := sessionStore.Get(c.Request.Context(), token)
				if err == nil && sess != nil {
					// Set user information in context
					c.Set("user_id", uint(sess.UserID))
					c.Set("username", sess.Username)
					c.Set("group_id", uint(sess.GroupID))
					c.Set("is_admin", sess.IsAdmin)
					c.Next()
					return
				}
			}
		}
		
		// No valid authentication found
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "Authentication required",
			"error":   "No valid session cookie or Bearer token found",
		})
		c.Abort()
	}
}
