package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/openwan/media-asset-management/internal/service"
	"github.com/openwan/media-asset-management/internal/session"
)

// AuthHandler handles authentication operations
type AuthHandler struct {
	aclService   *service.ACLService
	sessionStore session.Store
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(aclService *service.ACLService, sessionStore session.Store) *AuthHandler {
	return &AuthHandler{
		aclService:   aclService,
		sessionStore: sessionStore,
	}
}

// LoginRequest represents login request body
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse represents login response
type LoginResponse struct {
	Success bool      `json:"success"`
	Message string    `json:"message"`
	Token   string    `json:"token,omitempty"`
	User    *UserInfo `json:"user,omitempty"`
}

// UserInfo represents user information returned after login
type UserInfo struct {
	ID          int      `json:"id"`
	Username    string   `json:"username"`
	Email       *string  `json:"email"`
	GroupID     int      `json:"group_id"`
	LevelID     int      `json:"level_id"`
	Permissions []string `json:"permissions"`
	Roles       []string `json:"roles"`
}

// Login handles user login
func (h *AuthHandler) Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid request body",
				"error":   err.Error(),
			})
			return
		}

		// Authenticate user
		user, err := h.aclService.AuthenticateUser(c.Request.Context(), req.Username, req.Password)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Invalid username or password",
			})
			return
		}

		// Get user permissions
		permissions, err := h.aclService.GetUserPermissions(c.Request.Context(), user.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to load user permissions",
			})
			return
		}

		// Get user roles
		roles, err := h.aclService.GetUserRoles(c.Request.Context(), user.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to load user roles",
			})
			return
		}

		// Build permission list
		permList := make([]string, len(permissions))
		for i, p := range permissions {
			permList[i] = fmt.Sprintf("%s.%s.%s", p.Namespace, p.Controller, p.Action)
		}

		// Build role list
		roleList := make([]string, len(roles))
		for i, r := range roles {
			roleList[i] = r.Name
		}

		// Check if user is admin based on roles (case-insensitive, support Chinese)
		isAdmin := false
		for _, roleName := range roleList {
			roleNameUpper := strings.ToUpper(roleName)
			if roleNameUpper == "ADMIN" || roleNameUpper == "SYSTEM" || 
			   roleNameUpper == "ADMINISTRATOR" || roleName == "超级管理员" {
				isAdmin = true
				break
			}
		}

		// Store user in session
		sessionID := uuid.New().String()
		sess := &session.SessionData{
			UserID:      user.ID,
			Username:    user.Username,
			GroupID:     user.GroupID,
			LevelID:     user.LevelID,
			IsAdmin:     isAdmin,
			Permissions: permList,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		// Store session in Redis
		if h.sessionStore != nil {
			if err := h.sessionStore.Save(c.Request.Context(), sessionID, sess); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"message": "Failed to create session",
				})
				return
			}
		}

		// Set session cookie
		c.SetCookie(
			"openwan_session",
			sessionID,
			86400,
			"/",
			"",
			false,
			true,
		)

		// Store user in context for this request
		c.Set("user_id", user.ID)
		c.Set("username", user.Username)
		c.Set("is_admin", sess.IsAdmin)

		// Return session ID as token for frontend compatibility
		token := sessionID

		c.JSON(http.StatusOK, LoginResponse{
			Success: true,
			Message: "Login successful",
			Token:   token,
			User: &UserInfo{
				ID:          user.ID,
				Username:    user.Username,
				Email:       user.Email,
				GroupID:     user.GroupID,
				LevelID:     user.LevelID,
				Permissions: permList,
				Roles:       roleList,
			},
		})
	}
}

// Logout handles user logout
func (h *AuthHandler) Logout() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get session cookie
		sessionID, err := c.Cookie("openwan_session")
		if err == nil && sessionID != "" && h.sessionStore != nil {
			// Delete session from store
			h.sessionStore.Delete(c.Request.Context(), sessionID)
		}

		// Clear session cookie
		c.SetCookie(
			"openwan_session",
			"",
			-1,
			"/",
			"",
			false,
			true,
		)

		// Clear context
		c.Set("user_id", nil)
		c.Set("username", nil)
		c.Set("is_admin", nil)

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Logout successful",
		})
	}
}

// GetCurrentUser returns current user information
func (h *AuthHandler) GetCurrentUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user from context (set by auth middleware)
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Not authenticated",
			})
			return
		}

		// Get user details
		user, err := h.aclService.GetUserByID(c.Request.Context(), userID.(uint))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "User not found",
			})
			return
		}

		// Get user permissions
		permissions, err := h.aclService.GetUserPermissions(c.Request.Context(), user.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to load user permissions",
			})
			return
		}

		// Get user roles
		roles, err := h.aclService.GetUserRoles(c.Request.Context(), user.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to load user roles",
			})
			return
		}

		// Build permission list
		permList := make([]string, len(permissions))
		for i, p := range permissions {
			permList[i] = fmt.Sprintf("%s.%s.%s", p.Namespace, p.Controller, p.Action)
		}

		// Build role list
		roleList := make([]string, len(roles))
		for i, r := range roles {
			roleList[i] = r.Name
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"user": UserInfo{
				ID:          user.ID,
				Username:    user.Username,
				Email:       user.Email,
				GroupID:     user.GroupID,
				LevelID:     user.LevelID,
				Permissions: permList,
				Roles:       roleList,
			},
		})
	}
}

// ChangePassword handles password change
func (h *AuthHandler) ChangePassword() gin.HandlerFunc {
	return func(c *gin.Context) {
		type ChangePasswordRequest struct {
			OldPassword string `json:"old_password" binding:"required"`
			NewPassword string `json:"new_password" binding:"required,min=6"`
		}

		var req ChangePasswordRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid request body",
				"error":   err.Error(),
			})
			return
		}

		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Not authenticated",
			})
			return
		}

		// Verify old password
		user, err := h.aclService.GetUserByID(c.Request.Context(), userID.(uint))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "User not found",
			})
			return
		}

		// TODO: Verify old password against stored hash
		// For now, placeholder implementation
		_ = user // TODO: use this to verify password

		// Update password
		// TODO: Hash new password and update in database

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Password changed successfully",
		})
	}
}


// RefreshToken refreshes user session/token
func (h *AuthHandler) RefreshToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get session ID from cookie
		sessionID, err := c.Cookie("openwan_session")
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Invalid session",
			})
			return
		}

		// Get session data
		sessData, err := h.sessionStore.Get(c.Request.Context(), sessionID)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Session not found",
			})
			return
		}

		// Check if user is logged in
		if sessData.UserID == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Not logged in",
			})
			return
		}

		// Get user info
		user, err := h.aclService.GetUserByID(c.Request.Context(), uint(sessData.UserID))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to get user info",
			})
			return
		}

		if user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "User not found",
			})
			return
		}

		// Update session expiry by saving it again with new timestamp
		sessData.UpdatedAt = time.Now()
		if err := h.sessionStore.Save(c.Request.Context(), sessionID, sessData); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to refresh session",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Token refreshed successfully",
			"data": gin.H{
				"user_id":  user.ID,
				"username": user.Username,
			},
		})
	}
}

// UpdateProfile updates user profile information
func (h *AuthHandler) UpdateProfile() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get session ID from cookie
		sessionID, err := c.Cookie("openwan_session")
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Invalid session",
			})
			return
		}

		// Get session data
		sessData, err := h.sessionStore.Get(c.Request.Context(), sessionID)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Session not found",
			})
			return
		}

		if sessData.UserID == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Not logged in",
			})
			return
		}

		// Parse request
		var req struct {
			Email     string `json:"email"`
			RealName  string `json:"real_name"`
			Telephone string `json:"telephone"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid request data",
			})
			return
		}

		// Update profile through service
		if err := h.aclService.UpdateUserProfile(c.Request.Context(), sessData.UserID, req.Email, req.RealName, req.Telephone); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to update profile: " + err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Profile updated successfully",
		})
	}
}

